import {ReactNode} from "react";
import {Link} from "react-router-dom";

type View = {label: string, to: string, icon: ReactNode, active: boolean}

export function SidebarGroup({label, icon, views, onNavigate}: {
  label: string,
  icon: ReactNode,
  views: View[],
  onNavigate?: () => void,
}) {
  return <div>
    <div className="flex h-11 items-center gap-3 border-l-3 border-transparent px-3 font-pixel text-sm text-content">
      <svg aria-hidden="true" viewBox="0 0 20 20" className="size-5 shrink-0" fill="none" stroke="currentColor" strokeWidth="1.5">{icon}</svg>
      {label}
    </div>
    <nav aria-label={`${label} navigation`} className="ml-6 flex flex-col gap-1 border-l border-edge">
      {views.map(view => <Link key={view.to} to={view.to} onClick={onNavigate}
        aria-current={view.active ? "page" : undefined} className={`mc-sidebar-link ${view.active ? "active" : ""}`}>
        <svg aria-hidden="true" viewBox="0 0 20 20" className="size-5 shrink-0" fill="none" stroke="currentColor" strokeWidth="1.5">{view.icon}</svg>
        {view.label}
      </Link>)}
    </nav>
  </div>
}
