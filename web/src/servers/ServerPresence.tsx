import {useId} from "react";
import {ServerResponse} from "../clients/ManagerClient.ts";
import {statusDots, statusLabels} from "./status.ts";
import {worldState} from "./serverPresentation.ts";

export function ServerPresence({server, managed = false}: {server: ServerResponse, managed?: boolean}) {
  const tooltipId = useId()
  const state = managed ? worldState(server) : server.status
  const label = state === "saving" ? "Saving" : statusLabels[state] ?? state
  const dot = state === "saving" ? "bg-warning motion-safe:animate-pulse" : statusDots[state] ?? "bg-content-muted"
  const players = server.status === "running" && server.players

  const shutdownAt = !managed && server.status === "expiring" ? server.idleShutdownAt : null
  const minutes = shutdownAt ? Math.round((new Date(shutdownAt).getTime() - Date.now()) / 60_000) : null
  const hint = minutes === null ? null : minutes < 1 ? "Shuts down any moment" : `Shuts down in ${minutes}m`

  return <div className="group relative flex items-center gap-2 self-start pt-1 text-xs focus-ring" tabIndex={hint ? 0 : undefined} aria-describedby={hint ? tooltipId : undefined}>
    <span className={`size-2 shrink-0 ${dot}`} aria-hidden="true"/>
    <span>{players ? <><span className="sr-only">{label} </span>{players.online}/{players.max}</> : label}</span>
    {hint && <span id={tooltipId} role="tooltip" className="absolute right-0 top-full z-10 hidden pt-1 group-hover:block group-focus:block">
      <span className="block whitespace-nowrap px-2 py-1 bg-inset border border-edge">{hint}</span>
    </span>}
  </div>
}
