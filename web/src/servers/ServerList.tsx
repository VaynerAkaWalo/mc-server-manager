import {useEffect, useState} from "react";
import {useLoaderData} from "react-router-dom";
import {ServerResponse, ServerStatus} from "../clients/ManagerClient.ts";
import {UpdateNotification} from "../App.tsx";
import {activeServers} from "./activeServers.ts";
import {LinkButton} from "../ui/Button.tsx";
import {Heading} from "../ui/Heading.tsx";
import {Panel} from "../ui/Panel.tsx";
import {TextInput} from "../ui/Field.tsx";

function shutdownHint(idleShutdownAt: string) {
  const minutes = Math.round((new Date(idleShutdownAt).getTime() - Date.now()) / 60_000)
  return minutes < 1 ? "Shuts down any moment" : `Shuts down in ${minutes}m`
}

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
  const [servers, setServers] = useState(useLoaderData<ServerResponse[] | null>())
  const [search, setSearch] = useState("")

  const query = search.trim().toLowerCase()
  const matchingServers = servers?.filter(server =>
    [server.name, server.hostname, server.options.VERSION ?? ""].some(value => value.toLowerCase().includes(query))
  ) ?? []

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
        setServers(await activeServers())
      } catch {
        // keep the last known list, the next poll retries
      }
    }

    const loader = setInterval(loadServers, 10000)
    return () => clearInterval(loader)
  }, [])

  return (
    <Panel as="section" className="mx-4 my-6 flex-1 max-w-content p-4 sm:p-8 font-sans text-sm">
      <header className="flex flex-wrap items-center justify-between gap-4 mb-6">
        <Heading level={2} id="servers-heading">Servers</Heading>
        <LinkButton to="/server-creator">Create server</LinkButton>
      </header>

      <label className="sr-only" htmlFor="server-search">Search servers</label>
      <TextInput id="server-search" type="search" className="w-full mb-4"
             placeholder="Search by name, hostname or version"
             value={search} onChange={event => setSearch(event.target.value)}/>

      <div className="max-h-136 overflow-y-auto focus-ring"
           role="region" aria-labelledby="servers-heading" tabIndex={0}>
        {servers === null ? (
          <p className="py-8 text-content-muted" role="status">Loading servers...</p>
        ) : servers.length === 0 ? (
          <p className="py-8 text-content-muted">No servers running right now. Create one to get started.</p>
        ) : matchingServers.length === 0 ? (
          <p className="py-8 text-content-muted" role="status">No servers match your search.</p>
        ) : (
          <ul>
            {matchingServers.map(server => (
              <li key={server.id} className="grid grid-cols-[minmax(0,1fr)_auto] gap-x-4 gap-y-2 py-4 px-1 border-t border-edge">
                <div className="min-w-0">
                  <Heading level={3} className="wrap-anywhere">{server.name}</Heading>
                  <p className="text-content-muted">{server.options.VERSION ?? "Unknown version"}</p>
                </div>

                <ServerPresence server={server}/>

                <button className="col-span-full flex items-center justify-between gap-4 px-3 py-2 bg-inset hover:bg-surface-hover text-left cursor-pointer focus-ring"
                        aria-label={`Copy hostname ${server.hostname}`}
                        onClick={() => copyHostname(server.hostname)}>
                  <span className="font-mono text-xs text-content-muted wrap-anywhere">{server.hostname}</span>
                  <span className="shrink-0 text-xs text-action-content">Copy</span>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </Panel>
  )
}

function ServerPresence({server}: { server: ServerResponse }) {
  const label = statusLabels[server.status] ?? server.status
  const players = server.status === "running" && server.players
  const hint = server.status === "expiring" && server.idleShutdownAt ? shutdownHint(server.idleShutdownAt) : label

  return (
    <div className="group relative flex items-center gap-2 self-start pt-1 text-xs" tabIndex={hint !== label ? 0 : undefined}>
      <span className={`size-2 shrink-0 ${statusDots[server.status] ?? "bg-content-muted"}`} role="img" aria-label={hint}/>
      <span>{players ? `${players.online}/${players.max}` : label}</span>
      {hint !== label && (
        <span role="tooltip" className="absolute right-0 top-full z-10 hidden pt-1 group-hover:block group-focus:block">
          <span className="block whitespace-nowrap px-2 py-1 bg-inset border border-edge">{hint}</span>
        </span>
      )}
    </div>
  )
}
