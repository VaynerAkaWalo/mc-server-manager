import {ReactNode} from "react";

export function Details({items}: {items: {label: string, value: ReactNode}[]}) {
  return <dl>{items.map(item => <div key={item.label} className="grid grid-cols-[minmax(0,1fr)_minmax(0,2fr)] gap-4 border-t border-edge py-3 text-sm">
    <dt className="text-content-muted">{item.label}</dt>
    <dd className="min-w-0 whitespace-pre-wrap wrap-anywhere">{item.value}</dd>
  </div>)}</dl>
}
