import {UserManager, WebStorageStateStore} from "oidc-client-ts";

export const callbackPath = "/auth/callback"
export const silentCallbackPath = "/auth/silent-callback"

export const userManager = new UserManager({
  authority: import.meta.env.VITE_OIDC_ISSUER,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: window.location.origin + callbackPath,
  silent_redirect_uri: window.location.origin + silentCallbackPath,
  scope: "openid email",
  userStore: new WebStorageStateStore({store: window.localStorage}),
})

// Barricade issues no refresh tokens, so an ended login session there ends ours too
userManager.events.addSilentRenewError(() => userManager.removeUser())
