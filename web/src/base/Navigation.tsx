import {NavLink} from "react-router-dom";

const destinations = [
  {
    to: "/my-servers",
    label: "Servers",
    icon: <><rect x="3" y="3" width="14" height="6"/><rect x="3" y="11" width="14" height="6"/><path d="M5 6h2m-2 8h2"/></>,
  },
  {
    to: "/server-creator",
    label: "Creator",
    icon: <path d="M10 3v14M3 10h14"/>,
  },
  {
    to: "/server-templates",
    label: "Templates",
    icon: <><rect x="3" y="3" width="6" height="6"/><rect x="11" y="3" width="6" height="6"/><rect x="3" y="11" width="6" height="6"/><rect x="11" y="11" width="6" height="6"/></>,
  },
]

export function Navigation({onNavigate}: {onNavigate?: () => void}) {
  return (
    <nav aria-label="Main navigation" className="flex flex-col gap-1 p-3">
      {destinations.map(({to, label, icon}) => (
        <NavLink key={to} to={to} onClick={onNavigate} className="mc-sidebar-link">
          <svg aria-hidden="true" viewBox="0 0 20 20" className="size-4 shrink-0" fill="none" stroke="currentColor" strokeWidth="1.5">{icon}</svg>
          {label}
        </NavLink>
      ))}
    </nav>
  )
}
