package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/problem"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const sessionCookie = "manager_session"
const loginCookie = "manager_login"
const sessionLifetime = 24 * time.Hour
const loginLifetime = 10 * time.Minute

type Store interface {
	CreateLoginAttempt(context.Context, store.LoginAttempt) error
	ConsumeLoginAttempt(context.Context, string) (store.LoginAttempt, error)
	CreateSession(context.Context, string, store.Session) error
	GetSession(context.Context, string) (store.Session, error)
	DeleteSession(context.Context, string) error
	DeleteExpiredAuth(context.Context) error
}

type Handlers struct {
	store      Store
	oauth      oauth2.Config
	issuer     string
	origin     string
	secure     bool
	providerMu sync.Mutex
	verifier   *oidc.IDTokenVerifier
}

func NewHandlers(issuer, clientID, appURL string, sessions Store) (*Handlers, error) {
	origin, err := url.Parse(appURL)
	if err != nil {
		return nil, fmt.Errorf("parse APP_URL: %w", err)
	}
	if origin.Host == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || (origin.Path != "" && origin.Path != "/") || (origin.Scheme != "https" && origin.Scheme != "http") {
		return nil, fmt.Errorf("APP_URL must be an http(s) origin")
	}
	if issuer == "" || clientID == "" {
		return nil, fmt.Errorf("OIDC_ISSUER and OIDC_CLIENT_ID are required")
	}

	appOrigin := origin.Scheme + "://" + origin.Host

	return &Handlers{
		store:  sessions,
		oauth:  oauth2.Config{ClientID: clientID, RedirectURL: appOrigin + "/auth/callback", Scopes: []string{oidc.ScopeOpenID}},
		issuer: issuer,
		origin: appOrigin,
		secure: origin.Scheme == "https",
	}, nil
}

func (h *Handlers) configuration(ctx context.Context) (oauth2.Config, *oidc.IDTokenVerifier, error) {
	h.providerMu.Lock()
	defer h.providerMu.Unlock()

	if h.verifier == nil {
		provider, err := oidc.NewProvider(ctx, h.issuer)
		if err != nil {
			return oauth2.Config{}, nil, err
		}

		h.oauth.Endpoint = provider.Endpoint()
		h.oauth.Endpoint.AuthStyle = oauth2.AuthStyleInParams
		h.verifier = provider.Verifier(&oidc.Config{ClientID: h.oauth.ClientID})
	}

	return h.oauth, h.verifier, nil
}

func (h *Handlers) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /auth/login", h.login)
	router.RegisterHandler("GET /auth/callback", h.callback)
	router.RegisterHandler("GET /api/v1/sessions/current", h.current)
	router.RegisterHandler("DELETE /api/v1/sessions/current", h.logout)
}

func (h *Handlers) login(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	returnTo := r.URL.Query().Get("returnTo")
	target, err := url.Parse(returnTo)
	if err != nil || target.IsAbs() || target.Host != "" || strings.HasPrefix(target.Path, "//") || strings.Contains(returnTo, `\`) || len(returnTo) == 0 || returnTo[0] != '/' || len(returnTo) > 2048 || target.Path == "/auth/login" || target.Path == "/auth/callback" {
		return problem.Write(w, http.StatusBadRequest, "Invalid return path")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	config, _, err := h.configuration(ctx)
	if err != nil {
		return h.unavailable(w, err)
	}

	state, nonce := rand.Text(), rand.Text()
	verifier := oauth2.GenerateVerifier()
	attempt := store.LoginAttempt{StateHash: hashToken(state), Verifier: verifier, Nonce: nonce, ReturnTo: returnTo, ExpiresAt: time.Now().Add(loginLifetime)}
	if err := h.store.CreateLoginAttempt(ctx, attempt); err != nil {
		return h.failure(w, "save login attempt", err)
	}

	h.setCookie(w, loginCookie, state, int(loginLifetime.Seconds()))
	http.Redirect(w, r, config.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)), http.StatusFound)
	return nil
}

func (h *Handlers) callback(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie(h.cookieName(loginCookie))
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		return problem.Write(w, http.StatusBadRequest, "Invalid login state")
	}
	h.setCookie(w, loginCookie, "", -1)

	attempt, err := h.store.ConsumeLoginAttempt(r.Context(), hashToken(state))
	if errors.Is(err, store.ErrNotFound) {
		return problem.Write(w, http.StatusBadRequest, "Login attempt expired or already used")
	}
	if err != nil {
		return h.failure(w, "consume login attempt", err)
	}
	if r.URL.Query().Get("error") != "" || r.URL.Query().Get("code") == "" {
		return problem.Write(w, http.StatusUnauthorized, "Login was not completed")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	config, verifier, err := h.configuration(ctx)
	if err != nil {
		return h.unavailable(w, err)
	}

	tokens, err := config.Exchange(ctx, r.URL.Query().Get("code"), oauth2.VerifierOption(attempt.Verifier))
	if err != nil {
		return h.failedLogin(w, "exchange authorization code", err)
	}
	rawIDToken, ok := tokens.Extra("id_token").(string)
	if !ok {
		return h.failedLogin(w, "read ID token", errors.New("missing ID token"))
	}
	identity, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return h.failedLogin(w, "verify ID token", err)
	}
	if identity.Nonce != attempt.Nonce || identity.Subject == "" {
		return h.failedLogin(w, "verify identity", errors.New("invalid nonce or subject"))
	}

	session := store.Session{Subject: identity.Subject, ExpiresAt: time.Now().UTC().Add(sessionLifetime)}
	token := rand.Text()
	if err := h.store.CreateSession(ctx, hashToken(token), session); err != nil {
		return h.failure(w, "save session", err)
	}
	if previous, err := r.Cookie(h.cookieName(sessionCookie)); err == nil {
		if err := h.store.DeleteSession(ctx, hashToken(previous.Value)); err != nil {
			return h.failure(w, "replace session", err)
		}
	}

	h.setCookie(w, sessionCookie, token, int(sessionLifetime.Seconds()))
	http.Redirect(w, r, attempt.ReturnTo, http.StatusSeeOther)
	return nil
}

func (h *Handlers) current(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	cookie, err := r.Cookie(h.cookieName(sessionCookie))
	if err != nil {
		return problem.Write(w, http.StatusUnauthorized, "No active session")
	}
	session, err := h.store.GetSession(r.Context(), hashToken(cookie.Value))
	if errors.Is(err, store.ErrNotFound) {
		h.setCookie(w, sessionCookie, "", -1)
		return problem.Write(w, http.StatusUnauthorized, "No active session")
	}
	if err != nil {
		return h.failure(w, "read session", err)
	}
	return xhttp.WriteResponse(w, http.StatusOK, session)
}

func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	if r.Header.Get("Origin") != h.origin {
		return problem.Write(w, http.StatusForbidden, "Invalid request origin")
	}
	if cookie, err := r.Cookie(h.cookieName(sessionCookie)); err == nil {
		if err := h.store.DeleteSession(r.Context(), hashToken(cookie.Value)); err != nil {
			return h.failure(w, "delete session", err)
		}
	}
	h.setCookie(w, sessionCookie, "", -1)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handlers) setCookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: h.cookieName(name), Value: value, Path: "/", HttpOnly: true, Secure: h.secure, SameSite: http.SameSiteLaxMode, MaxAge: maxAge})
}

func (h *Handlers) cookieName(name string) string {
	if h.secure {
		return "__Host-" + name
	}
	return name
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (h *Handlers) failure(w http.ResponseWriter, operation string, err error) error {
	slog.Error(operation, "error", err)
	return problem.Write(w, http.StatusInternalServerError, "Authentication is unavailable")
}

func (h *Handlers) unavailable(w http.ResponseWriter, err error) error {
	slog.Warn("discover OIDC provider", "error", err)
	return problem.Write(w, http.StatusServiceUnavailable, "Login is temporarily unavailable")
}

func (h *Handlers) failedLogin(w http.ResponseWriter, operation string, err error) error {
	slog.Warn(operation, "error", err)
	return problem.Write(w, http.StatusUnauthorized, "Login could not be verified")
}

func (h *Handlers) RunCleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		if err := h.store.DeleteExpiredAuth(ctx); err != nil {
			slog.Error("clean up authentication records", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
