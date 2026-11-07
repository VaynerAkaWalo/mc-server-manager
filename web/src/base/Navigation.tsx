import {NavLink} from "react-router-dom";

const destinations = [
  {to: "/", label: "Home", icon: <path d="M3 9.5 10 3l7 6.5V17h-5v-5H8v5H3z"/>},
  {to: "/my-servers", label: "Server list", icon: <><rect x="3" y="3" width="14" height="6"/><rect x="3" y="11" width="14" height="6"/><path d="M5 6h2m-2 8h2"/></>},
  {to: "/server-creator", label: "Creator", icon: <path d="M10 3v14M3 10h14"/>},
  {to: "/server-templates", label: "Templates", icon: <><rect x="3" y="3" width="6" height="6"/><rect x="11" y="3" width="6" height="6"/><rect x="3" y="11" width="6" height="6"/><rect x="11" y="11" width="6" height="6"/></>},
]

export function Navigation({onNavigate}: {onNavigate?: () => void}) {
  return <nav aria-label="Main navigation" className="flex flex-col gap-1 py-4 pr-3">
    {destinations.map(({to, label, icon}) => <NavLink key={to} to={to} end={to === "/"} onClick={onNavigate} className="mc-sidebar-link">
      <svg aria-hidden="true" viewBox="0 0 20 20" className="size-5 shrink-0" fill="none" stroke="currentColor" strokeWidth="1.5">{icon}</svg>
      {label}
    </NavLink>)}
    <div className="mt-4 ml-4 border-t border-edge pt-4 pb-1 text-xs uppercase tracking-wide text-content-muted">Account</div>
    <NavLink to="/manage-servers" onClick={onNavigate} className="mc-sidebar-link">
      <svg aria-hidden="true" viewBox="0 0 20 20" className="size-5 shrink-0" fill="none" stroke="currentColor" strokeWidth="1.5"><circle cx="10" cy="6" r="3"/><path d="M4 17v-2a6 6 0 0 1 12 0v2"/></svg>
      Manage servers
    </NavLink>
  </nav>
}
