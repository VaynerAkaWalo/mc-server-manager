import {useEffect, useState} from "react";
import {Link} from "react-router-dom";
import {ManagerClient, ServerResponse, ServerStatus} from "../clients/ManagerClient.ts";
import {UpdateNotification} from "../App.tsx";

function formatRemainingTime(expiresAt: string) {
  const remainingMs = new Date(expiresAt).getTime() - Date.now()
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
    <div className="mc-panel w-4/5 max-w-4xl max-h-full overflow-y-auto p-4">
      <div className="flex flex-col gap-3 mb-4 sm:flex-row sm:justify-between sm:items-center">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
          <h2 className="text-xl">Servers</h2>
          <Link className="mc-button shrink-0" to="/server-creator">Create server</Link>
        </div>
      </div>

      <div className="h-[33rem] overflow-auto">
        {servers === null ? null : servers.length === 0 ? (
          <p className="text-center py-8">No servers running right now. Create one to get started.</p>
        ) : (
          <table className="w-full text-left">
            <thead>
              <tr>
                <th className="p-2">Name</th>
                <th className="p-2">Version</th>
                <th className="p-2">Status</th>
                <th className="p-2">Players</th>
                <th className="p-2">Remaining time</th>
                <th className="p-2">Hostname</th>
              </tr>
            </thead>
            <tbody>
              {servers.map(server => (
                <tr key={server.id}>
                  <td className="p-2 max-w-48 truncate" title={server.name}>{server.name}</td>
                  <td className="p-2">{server.options.VERSION ?? "-"}</td>
                  <td className="p-2">{statusLabels[server.status] ?? server.status}</td>
                  <td className="p-2">{server.players ? `${server.players.online}/${server.players.max}` : "-"}</td>
                  <td className="p-2">{formatRemainingTime(server.expiresAt)}</td>
                  <td className="p-2">
                    <button className="max-w-64 truncate cursor-pointer" title={`Copy ${server.hostname}`}
                            onClick={() => copyHostname(server.hostname)}>
                      {server.hostname}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
