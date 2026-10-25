import {useEffect, useState} from "react";
import {Link} from "react-router-dom";
import {ManagerClient, ServerResponse, ServerStatus} from "../clients/ManagerClient.ts";
import {UpdateNotification} from "../App.tsx";

function shutdownTime(server: ServerResponse) {
  const expiresAt = new Date(server.expiresAt).getTime()
  if (!server.idleShutdownAt || !server.players || server.players.online > 0) {
    return expiresAt
  }

  return Math.min(expiresAt, new Date(server.idleShutdownAt).getTime())
}

function formatRemainingTime(shutdownAt: number) {
  const remainingMs = shutdownAt - Date.now()
  if (remainingMs <= 0) {
    return "expiring"
  }

  const minutes = Math.round(remainingMs / 60_000)
  if (minutes < 1) {
    return "soon"
  }

  const days = Math.floor(minutes / 1440)
  const hours = Math.floor((minutes % 1440) / 60)
  const parts = [days && `${days}d`, hours && `${hours}h`, minutes % 60 && `${minutes % 60}m`]
  return parts.filter(Boolean).join(" ")
}

const activeStatuses: ServerStatus[] = ["new", "running", "degraded"]

const statusDots: Partial<Record<ServerStatus, string>> = {
  new: "bg-starting motion-safe:animate-pulse",
  running: "bg-online",
  degraded: "bg-warning"
}

const statusLabels: Record<ServerStatus, string> = {
  new: "Starting",
  running: "Online",
  degraded: "Not responding",
  stopped: "Stopped",
  expired: "Expired"
}

export function ServerList() {
  const {setNotificationMessage} = UpdateNotification()
  const [servers, setServers] = useState<ServerResponse[] | null>(null)

  const copyHostname = async (hostname: string) => {
    try {
      await navigator.clipboard.writeText(hostname)
      setNotificationMessage("Hostname copied to clipboard")
    } catch {
      setNotificationMessage("Could not copy hostname")
    }
  }

  useEffect(() => {
    const loadServers = async () => {
      try {
        const items: ServerResponse[] = []
        let cursor: string | undefined
        do {
          const {data} = await ManagerClient.listServers(cursor)
          items.push(...data.items)
          cursor = data.nextCursor || undefined
        } while (cursor)

        setServers(items.filter(server => activeStatuses.includes(server.status)))
      } catch {
        // keep the last known list, the next poll retries
      }
    }

    loadServers()
    const loader = setInterval(loadServers, 10000)
    return () => clearInterval(loader)
  }, [])

  return (
    <section className="mc-panel w-[min(60rem,calc(100%-2rem))] my-6 p-[clamp(1rem,3vw,2rem)] font-sans text-sm">
      <header className="flex flex-wrap items-center justify-between gap-4 mb-6">
        <h2 className="font-pixel text-2xl">Servers</h2>
        <Link className="mc-button" to="/server-creator">Create server</Link>
      </header>

      {servers === null ? null : servers.length === 0 ? (
        <p className="py-8 text-content-muted">No servers running right now. Create one to get started.</p>
      ) : (
        <ul>
          {servers.map(server => (
            <li key={server.id} className="grid grid-cols-[minmax(0,1fr)_auto] gap-x-4 gap-y-2 py-4 px-1 border-t border-edge">
              <div className="min-w-0">
                <h3 className="font-pixel text-base wrap-anywhere">{server.name}</h3>
                <div className="flex flex-wrap gap-x-4 text-content-muted">
                  <span>{server.options.VERSION ?? "Unknown version"}</span>
                  <span title="Shuts down after 30 minutes without players">{formatRemainingTime(shutdownTime(server))}</span>
                </div>
              </div>

              <ServerPresence server={server}/>

              <button className="col-span-full flex items-center justify-between gap-4 px-3 py-2 bg-inset hover:bg-surface-hover text-left cursor-pointer"
                      aria-label={`Copy hostname ${server.hostname}`}
                      onClick={() => copyHostname(server.hostname)}>
                <span className="font-mono text-xs text-content-muted wrap-anywhere">{server.hostname}</span>
                <span className="shrink-0 text-xs text-action-content">Copy</span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}

function ServerPresence({server}: { server: ServerResponse }) {
  const label = statusLabels[server.status] ?? server.status
  const players = server.status === "running" && server.players

  return (
    <div className="flex items-center gap-2 self-start pt-1 text-xs" title={label}>
      <span className={`size-2 shrink-0 ${statusDots[server.status] ?? "bg-content-muted"}`} role="img" aria-label={label}/>
      <span>{players ? `${players.online}/${players.max}` : label}</span>
    </div>
  )
}
