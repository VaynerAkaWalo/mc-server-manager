package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/frontend"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/go-jose/go-jose/v4"
	"golang.org/x/oauth2"
)

type memoryStore struct {
	attempts map[string]store.LoginAttempt
	sessions map[string]store.Session
}

func (s *memoryStore) CreateLoginAttempt(_ context.Context, a store.LoginAttempt) error {
	s.attempts[a.StateHash] = a
	return nil
}
func (s *memoryStore) ConsumeLoginAttempt(_ context.Context, hash string) (store.LoginAttempt, error) {
	a, ok := s.attempts[hash]
	if !ok || !a.ExpiresAt.After(time.Now()) {
		return store.LoginAttempt{}, store.ErrNotFound
	}
	delete(s.attempts, hash)
	return a, nil
}
func (s *memoryStore) CreateSession(_ context.Context, hash string, session store.Session) error {
	s.sessions[hash] = session
	return nil
}
func (s *memoryStore) GetSession(_ context.Context, hash string) (store.Session, error) {
	s1, ok := s.sessions[hash]
	if !ok || !s1.ExpiresAt.After(time.Now()) {
		return store.Session{}, store.ErrNotFound
	}
	return s1, nil
}
func (s *memoryStore) DeleteSession(_ context.Context, hash string) error {
	delete(s.sessions, hash)
	return nil
}
func (s *memoryStore) DeleteExpiredAuth(context.Context) error { return nil }

type oidcFixture struct {
	discoveryUnavailable atomic.Bool
	discoveryCalls       atomic.Int32
	handlers             *Handlers
	database             *memoryStore
	claims               map[string]any
	challenge            string
	tokenCalls           int
}

func newFixture(t *testing.T) *oidcFixture {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithHeader("kid", "test-key"))
	if err != nil {
		t.Fatal(err)
	}

	f := &oidcFixture{database: &memoryStore{attempts: map[string]store.LoginAttempt{}, sessions: map[string]store.Session{}}}
	var issuer string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			f.discoveryCalls.Add(1)
			if f.discoveryUnavailable.Load() {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"issuer": issuer, "authorization_endpoint": issuer + "/authorize", "token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys", "id_token_signing_alg_values_supported": []string{"RS256"}})
		case "/keys":
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "test-key", Algorithm: "RS256", Use: "sig"}}})
		case "/token":
			f.tokenCalls++
			_ = r.ParseForm()
			if r.Form.Get("client_id") != "manager" || r.Form.Get("redirect_uri") != "https://manager.test/auth/callback" || r.Form.Get("client_secret") != "" || r.Header.Get("Authorization") != "" || oauth2.S256ChallengeFromVerifier(r.Form.Get("code_verifier")) != f.challenge {
				t.Error("invalid code exchange or PKCE")
				http.Error(w, "invalid_grant", 400)
				return
			}
			payload, _ := json.Marshal(f.claims)
			signed, err := signer.Sign(payload)
			if err != nil {
				t.Error(err)
				return
			}
			token, _ := signed.CompactSerialize()
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "backend-only", "token_type": "Bearer", "expires_in": 120, "id_token": token})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(provider.Close)
	issuer = provider.URL
	f.claims = map[string]any{"iss": issuer, "aud": "manager", "sub": "player", "exp": time.Now().Add(time.Minute).Unix(), "iat": time.Now().Unix()}

	f.handlers, err = NewHandlers(issuer, "manager", "https://manager.test", f.database)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *oidcFixture) start(t *testing.T) (*http.Cookie, string) {
	t.Helper()
	w := httptest.NewRecorder()
	if err := f.handlers.login(w, httptest.NewRequest("GET", "/auth/login?returnTo=%2Fmy-servers%3Fpage%3D2%23list", nil)); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusFound {
		t.Fatalf("login status = %d", w.Code)
	}
	target, _ := url.Parse(w.Header().Get("Location"))
	q := target.Query()
	if q.Get("code_challenge_method") != "S256" || q.Get("scope") != "openid" {
		t.Fatalf("invalid authorization request: %s", target)
	}
	f.challenge = q.Get("code_challenge")
	f.claims["nonce"] = q.Get("nonce")
	cookie := w.Result().Cookies()[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.Name != "__Host-manager_login" || cookie.Domain != "" {
		t.Fatalf("unsafe login cookie: %+v", cookie)
	}
	return cookie, q.Get("state")
}

func (f *oidcFixture) finish(t *testing.T, cookie *http.Cookie, state string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest("GET", "/auth/callback?code=code&state="+url.QueryEscape(state), nil)
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	if err := f.handlers.callback(w, r); err != nil {
		t.Fatal(err)
	}
	return w
}

func TestLoginSessionAndLogout(t *testing.T) {
	f := newFixture(t)
	cookie, state := f.start(t)
	w := f.finish(t, cookie, state)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/my-servers?page=2#list" {
		t.Fatalf("callback: %d %s", w.Code, w.Body)
	}
	sessionCookie := w.Result().Cookies()[1]
	if !sessionCookie.HttpOnly || !sessionCookie.Secure || sessionCookie.Domain != "" || sessionCookie.MaxAge != 86400 || sessionCookie.Name != "__Host-manager_session" {
		t.Fatalf("unsafe session cookie: %+v", sessionCookie)
	}
	if _, ok := f.database.sessions[sessionCookie.Value]; ok {
		t.Fatal("raw session token stored")
	}

	r := httptest.NewRequest("GET", "/api/v1/sessions/current", nil)
	r.AddCookie(sessionCookie)
	current := httptest.NewRecorder()
	_ = f.handlers.current(current, r)
	if current.Code != 200 || !strings.Contains(current.Body.String(), `"subject":"player"`) || strings.Contains(current.Body.String(), "token") || current.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("current: %d %s", current.Code, current.Body)
	}

	if replay := f.finish(t, cookie, state); replay.Code != 400 || f.tokenCalls != 1 {
		t.Fatal("callback replay accepted")
	}

	for _, origin := range []string{"", "https://evil.test", "https://other.manager.test", "https://manager.test"} {
		r := httptest.NewRequest("DELETE", "/api/v1/sessions/current", nil)
		r.AddCookie(sessionCookie)
		r.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		_ = f.handlers.logout(w, r)
		if origin != "https://manager.test" {
			if w.Code != 403 || len(f.database.sessions) != 1 {
				t.Fatal("cross-origin logout accepted")
			}
		} else if w.Code != 204 || len(f.database.sessions) != 0 {
			t.Fatal("session not revoked")
		}
	}
	revoked := httptest.NewRecorder()
	_ = f.handlers.current(revoked, r)
	if revoked.Code != 401 {
		t.Fatal("revoked cookie accepted")
	}
}

func TestRejectInvalidIdentity(t *testing.T) {
	for _, claim := range []string{"nonce", "iss", "aud", "exp", "sub"} {
		t.Run(claim, func(t *testing.T) {
			f := newFixture(t)
			cookie, state := f.start(t)
			if claim == "exp" {
				f.claims[claim] = time.Now().Add(-time.Hour).Unix()
			} else {
				f.claims[claim] = ""
			}
			if w := f.finish(t, cookie, state); w.Code != 401 || len(f.database.sessions) != 0 {
				t.Fatalf("invalid %s accepted: %d", claim, w.Code)
			}
		})
	}
}

func TestRejectInvalidLoginState(t *testing.T) {
	f := newFixture(t)
	cookie, state := f.start(t)
	if w := f.finish(t, cookie, "wrong"); w.Code != 400 || f.tokenCalls != 0 {
		t.Fatal("mismatched state accepted")
	}
	if w := f.finish(t, &http.Cookie{Name: cookie.Name, Value: "wrong"}, state); w.Code != 400 || f.tokenCalls != 0 {
		t.Fatal("mismatched cookie accepted")
	}
	attempt := f.database.attempts[hashToken(state)]
	attempt.ExpiresAt = time.Now().Add(-time.Minute)
	f.database.attempts[hashToken(state)] = attempt
	if w := f.finish(t, cookie, state); w.Code != 400 || f.tokenCalls != 0 {
		t.Fatal("expired login accepted")
	}
}

func TestRejectExternalReturnPaths(t *testing.T) {
	h := &Handlers{}
	for _, target := range []string{"", "https://evil.test", "//evil.test", `/\evil.test`, "/%2Fevil.test", "/auth/login", "/auth/callback", "/\nLocation:evil"} {
		r := httptest.NewRequest("GET", "/auth/login?returnTo="+url.QueryEscape(target), nil)
		w := httptest.NewRecorder()
		_ = h.login(w, r)
		if w.Code != 400 {
			t.Errorf("accepted redirect %q", target)
		}
	}
}

func TestExpiredSession(t *testing.T) {
	f := newFixture(t)
	f.database.sessions[hashToken("expired")] = store.Session{Subject: "player", ExpiresAt: time.Now().Add(-time.Minute)}
	r := httptest.NewRequest("GET", "/api/v1/sessions/current", nil)
	r.AddCookie(&http.Cookie{Name: "__Host-manager_session", Value: "expired"})
	w := httptest.NewRecorder()
	_ = f.handlers.current(w, r)
	if w.Code != 401 || w.Result().Cookies()[0].MaxAge != -1 {
		t.Fatal("expired session accepted")
	}
}

func TestDiscoveryOutageDoesNotBlockAnonymousRoutesOrSessions(t *testing.T) {
	f := newFixture(t)
	f.discoveryUnavailable.Store(true)
	if f.discoveryCalls.Load() != 0 {
		t.Fatal("constructor contacted OIDC provider")
	}

	mux := http.NewServeMux()
	router := &xhttp.Router{ServeMux: mux}
	f.handlers.RegisterRoutes(router)
	frontend.NewFrontendHandlers(fstest.MapFS{"index.html": {Data: []byte("app")}}).RegisterRoutes(router)
	f.database.sessions[hashToken("existing")] = store.Session{Subject: "player", ExpiresAt: time.Now().Add(time.Hour)}

	for _, path := range []string{"/", "/api/v1/sessions/current", "/auth/login?returnTo=/"} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.AddCookie(&http.Cookie{Name: "__Host-manager_session", Value: "existing"})
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		want := http.StatusOK
		if path == "/auth/login?returnTo=/" {
			want = http.StatusServiceUnavailable
		}
		if w.Code != want {
			t.Fatalf("GET %s = %d, want %d", path, w.Code, want)
		}
	}
	if f.discoveryCalls.Load() != 1 {
		t.Fatal("anonymous or session routes contacted OIDC provider")
	}

	f.discoveryUnavailable.Store(false)
	f.start(t)
	f.start(t)
	if f.discoveryCalls.Load() != 2 {
		t.Fatal("discovery did not recover or was not cached")
	}
}

func TestCallbackDiscoversProviderAfterRestart(t *testing.T) {
	f := newFixture(t)
	cookie, state := f.start(t)
	h, err := NewHandlers(f.handlers.issuer, "manager", "https://manager.test", f.database)
	if err != nil {
		t.Fatal(err)
	}
	f.handlers = h

	if w := f.finish(t, cookie, state); w.Code != http.StatusSeeOther {
		t.Fatalf("callback after restart: %d %s", w.Code, w.Body)
	}
}
