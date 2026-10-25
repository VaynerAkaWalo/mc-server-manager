import {useEffect, useState} from "react";
import {Link} from "react-router-dom";
import {ManagerClient, ServerResponse} from "../clients/ManagerClient.ts";
import {UpdateNotification} from "../App.tsx";

function formatRemainingTime(remainingTime: string) {
  if (!remainingTime || remainingTime === "0s") {
    return "soon"
  }

  if (remainingTime.startsWith("-")) {
    return "expiring"
  }

  return remainingTime.replace("0s", "")
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
        const {data} = await ManagerClient.listServers()
        setServers(data)
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
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl">Servers</h2>
        <Link className="mc-button" to="/server-creator">Create server</Link>
      </div>

      {servers === null ? null : servers.length === 0 ? (
        <p className="text-center py-8">No servers running right now. Create one to get started.</p>
      ) : (
        <table className="w-full text-left">
          <thead>
            <tr>
              <th className="p-2">Name</th>
              <th className="p-2">Version</th>
              <th className="p-2">Max players</th>
              <th className="p-2">Remaining time</th>
              <th className="p-2">Hostname</th>
            </tr>
          </thead>
          <tbody>
            {servers.map(server => (
              <tr key={server.name}>
                <td className="p-2 max-w-48 truncate" title={server.name}>{server.name}</td>
                <td className="p-2">{server.version ?? "-"}</td>
                <td className="p-2">{server.maxPlayers ?? "-"}</td>
                <td className="p-2">{formatRemainingTime(server.remainingTime)}</td>
                <td className="p-2">
                  <button className="max-w-64 truncate cursor-pointer" title="Copy hostname"
                          onClick={() => copyHostname(server.IP)}>
                    {server.IP}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
