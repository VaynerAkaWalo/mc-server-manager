import {ReactNode} from "react";
import {Link} from "react-router-dom";

type View = {label: string, to: string, icon: ReactNode, active: boolean}

export function SidebarGroup({label, to, active, icon, views, onNavigate}: {
  label: string,
  to: string,
  active: boolean,
  icon: ReactNode,
  views: View[],
  onNavigate?: () => void,
}) {
  return <nav aria-label={`${label} navigation`}>
    <Link to={to} onClick={onNavigate} aria-current={active ? "page" : undefined} className={`mc-sidebar-link ${active ? "active" : ""}`}>
      <svg aria-hidden="true" viewBox="0 0 20 20" className="size-5 shrink-0" fill="none" stroke="currentColor" strokeWidth="1.5">{icon}</svg>
      {label}
    </Link>
    <div className="ml-6 flex flex-col gap-1 border-l border-edge">
      {views.map(view => <Link key={view.to} to={view.to} onClick={onNavigate}
        aria-current={view.active ? "page" : undefined} className={`mc-sidebar-link ${view.active ? "active" : ""}`}>
        <svg aria-hidden="true" viewBox="0 0 20 20" className="size-5 shrink-0" fill="none" stroke="currentColor" strokeWidth="1.5">{view.icon}</svg>
        {view.label}
      </Link>)}
    </div>
  </nav>
}
