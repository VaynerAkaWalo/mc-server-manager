import {Link} from "react-router-dom";

type View = {label: string, to: string, active: boolean}

export function ViewTabs({views, label}: {views: View[], label: string}) {
  return <nav aria-label={label} className="flex gap-4 border-b border-edge mb-6">
    {views.map(view => <Link key={view.to} to={view.to} aria-current={view.active ? "page" : undefined}
      className={`border-b-2 px-2 py-3 font-pixel text-sm hover:bg-surface-hover focus-ring ${view.active ? "border-action-content text-action-content" : "border-transparent text-content-muted hover:text-content"}`}>
      {view.label}
    </Link>)}
  </nav>
}
