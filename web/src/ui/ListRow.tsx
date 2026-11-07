import {ReactNode} from "react";
import {Link} from "react-router-dom";
import {Heading} from "./Heading.tsx";

export function ListRow({name, metadata, status, strip, to}: {
  name: string,
  metadata: string,
  status: ReactNode,
  strip: ReactNode,
  to?: string,
}) {
  const content = <>
    <div className="min-w-0">
      <Heading level={3} className="wrap-anywhere group-hover:text-action-content">{name}</Heading>
      <p className="text-sm text-content-muted">{metadata}</p>
    </div>
    {status}
    {strip}
  </>
  const layout = "grid grid-cols-[minmax(0,1fr)_auto] gap-x-4 gap-y-2 py-4 px-1 border-t border-edge"

  return <li>{to
    ? <Link to={to} className={`group ${layout} focus-ring`}>{content}</Link>
    : <div className={layout}>{content}</div>}
  </li>
}

export function ListRowStrip({children, action, onClick, label}: {
  children: ReactNode,
  action: string,
  onClick?: () => void,
  label?: string,
}) {
  const content = <>{children}<span className="shrink-0 text-action-content">{action}</span></>
  const layout = "col-span-full flex w-full items-center justify-between gap-4 bg-inset px-3 py-2 text-left text-xs text-content-muted"

  return onClick
    ? <button type="button" aria-label={label} onClick={onClick} className={`${layout} hover:bg-surface-hover focus-ring`}>{content}</button>
    : <span className={`${layout} group-hover:bg-surface-hover`}>{content}</span>
}
