import {useState} from "react";
import {Drawer, ListSubheader, Menu, MenuItem} from "../ui/mui.ts";
import {Button} from "../ui/Button.tsx";
import {Link, useLocation} from "react-router-dom";
import {useUser} from "../auth/useUser.ts";
import {Navigation} from "./Navigation.tsx";

export function Header() {
  const {user, loading, error, logOut} = useUser()
  const location = useLocation()
  const [navigationOpen, setNavigationOpen] = useState(false)
  const [accountAnchor, setAccountAnchor] = useState<HTMLElement | null>(null)

  const logIn = () => {
    window.location.assign("/auth/login?returnTo=" + encodeURIComponent(location.pathname + location.search + location.hash))
  }

  const signOut = async () => {
    await logOut()
    setAccountAnchor(null)
  }

  return (
    <>
      <header className="flex min-h-16 shrink-0 items-center justify-between gap-3 border-b border-edge-strong bg-surface px-3 sm:px-4">
        <div className="flex items-center gap-2">
          <Button variant="ghost" className="md:hidden"
                  aria-label="Open navigation" aria-controls={navigationOpen ? "mobile-navigation" : undefined} aria-expanded={navigationOpen}
                  onClick={() => setNavigationOpen(true)}>
            <svg aria-hidden="true" viewBox="0 0 16 16" className="size-3" fill="none" stroke="currentColor"><path d="M2 4h12M2 8h12M2 12h12"/></svg>
            Menu
          </Button>
          <Link to="/" className="flex items-center gap-3 p-2 font-pixel text-sm focus-ring">
            <img src="/grass-block.svg" alt="" className="hidden size-6 sm:block"/>
            Home
          </Link>
        </div>
        {user
          ? <Button id="account-button" variant="ghost" size="sm"
                    aria-haspopup="menu" aria-expanded={Boolean(accountAnchor)} aria-controls={accountAnchor ? "account-menu" : undefined}
                    onClick={event => setAccountAnchor(event.currentTarget)}>
            <span aria-hidden="true" className="flex size-6 items-center justify-center bg-surface-hover font-pixel text-xs text-action-content">{user.email?.[0].toUpperCase() ?? "?"}</span>
            Account
            <svg aria-hidden="true" viewBox="0 0 12 12" className="size-3" fill="none" stroke="currentColor"><path d="m3 4 3 3 3-3"/></svg>
          </Button>
          : <Button variant="ghost" disabled={loading} onClick={logIn}>Log in</Button>}
      </header>
      {error && <p role="status" className="border-b border-edge bg-surface px-4 py-2 text-sm text-warning">Login status unavailable. Try again.</p>}

      <Drawer anchor="left" open={navigationOpen} onClose={() => setNavigationOpen(false)}
              slotProps={{paper: {id: "mobile-navigation", role: "dialog", "aria-label": "Main navigation", sx: {width: 208, bgcolor: "var(--color-surface)", backgroundImage: "none", borderRight: "1px solid var(--color-edge-strong)"}}}}>
        <div className="flex min-h-16 items-center border-b border-edge px-3">
          <Button variant="ghost" onClick={() => setNavigationOpen(false)}>Close menu</Button>
        </div>
        <Navigation onNavigate={() => setNavigationOpen(false)}/>
      </Drawer>

      <Menu id="account-menu" anchorEl={accountAnchor} open={Boolean(accountAnchor) && Boolean(user)} onClose={() => setAccountAnchor(null)}
            anchorOrigin={{vertical: "bottom", horizontal: "right"}} transformOrigin={{vertical: "top", horizontal: "right"}}
            slotProps={{paper: {sx: {mt: 1, minWidth: 240, maxWidth: "calc(100vw - 32px)", borderRadius: 0, bgcolor: "var(--color-surface)", backgroundImage: "none", border: "1px solid var(--color-edge-strong)"}}, list: {"aria-labelledby": "account-button", sx: {py: 0}}}}>
        <ListSubheader component="div" disableSticky sx={{bgcolor: "transparent", px: 2, py: 2, lineHeight: 1.5}}>
          <p className="text-xs text-content-muted">{user?.email ? "Signed in as" : "Signed in"}</p>
          {user?.email && <p className="mt-1 break-all text-sm text-content">{user.email}</p>}
        </ListSubheader>
        <MenuItem disabled={loading} onClick={signOut} sx={{borderTop: "1px solid var(--color-edge)", px: 2, py: 1.5, fontSize: "var(--text-sm)", "&:hover, &.Mui-focusVisible": {bgcolor: "var(--color-surface-hover)"}}}>Log out</MenuItem>
      </Menu>
    </>
  )
}
