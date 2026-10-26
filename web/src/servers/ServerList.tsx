import {useEffect, useState} from "react";
import {Link} from "react-router-dom";
import {ManagerClient, ServerResponse, ServerStatus} from "../clients/ManagerClient.ts";
import {UpdateNotification} from "../App.tsx";

function shutdownHint(idleShutdownAt: string) {
  const minutes = Math.round((new Date(idleShutdownAt).getTime() - Date.now()) / 60_000)
  return minutes < 1 ? "Shuts down any moment" : `Shuts down in ${minutes}m`
}

const activeStatuses: ServerStatus[] = ["new", "running", "degraded", "expiring"]

const statusDots: Partial<Record<ServerStatus, string>> = {
  new: "bg-starting motion-safe:animate-pulse",
  running: "bg-online",
  degraded: "bg-warning",
  expiring: "bg-warning"
}

const statusLabels: Record<ServerStatus, string> = {
  new: "Starting",
  running: "Online",
  degraded: "Not responding",
  expiring: "Shutting down soon",
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
                <p className="text-content-muted">{server.options.VERSION ?? "Unknown version"}</p>
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
  const hint = server.status === "expiring" && server.idleShutdownAt ? shutdownHint(server.idleShutdownAt) : label

  return (
    <div className="group relative flex items-center gap-2 self-start pt-1 text-xs">
      <span className={`size-2 shrink-0 ${statusDots[server.status] ?? "bg-content-muted"}`} role="img" aria-label={hint}/>
      <span>{players ? `${players.online}/${players.max}` : label}</span>
      {hint !== label && (
        <span role="tooltip" className="pointer-events-none absolute right-0 top-full z-10 mt-1 hidden whitespace-nowrap px-2 py-1 bg-inset border border-edge group-hover:block">
          {hint}
        </span>
      )}
    </div>
  )
}
