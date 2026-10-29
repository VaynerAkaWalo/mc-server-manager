import {ReactNode} from "react";
import {useNavigate} from "react-router-dom";
import {Opts} from "../clients/ManagerClient.ts";
import {Button} from "../ui/Button.tsx";
import {Heading} from "../ui/Heading.tsx";

export interface template {
  name: string
  description?: string
  preset?: string
  opts?: Opts
  version?: string
  art: ReactNode
  secondary?: boolean
}

export function Chip({children}: { children: ReactNode }) {
  return <span className="border border-edge-strong bg-surface px-2 py-0.5 text-xs text-content-muted">{children}</span>
}

export function Template({name, description, preset, opts, version, art, secondary}: template) {
  const navigate = useNavigate()

  const use = () => {
    navigate("/server-creator", preset && opts ? {state: {templateName: name, templatePreset: preset, templateOps: opts}} : undefined)
  }

  return (
    <li className="group flex flex-col border border-edge-strong bg-inset hover:border-edge-control">
      <div className="h-44 overflow-hidden border-b border-edge-strong">{art}</div>
      <div className="flex flex-1 flex-col gap-3 p-4">
        <div className="flex items-center justify-between gap-2">
          <Heading level={2}>{name}</Heading>
          {version && <Chip>{version}</Chip>}
        </div>
        {description && <p className="line-clamp-3 text-sm text-content-muted">{description}</p>}
        <Button variant={secondary ? "secondary" : "primary"} className="mt-auto w-full" aria-label={`Use ${name}`} onClick={use}>Use</Button>
      </div>
    </li>
  )
}
