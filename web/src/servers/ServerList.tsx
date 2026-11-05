import {useEffect, useRef, useState} from "react";
import {Link, useLoaderData, useSearchParams} from "react-router-dom";
import {ManagerClient, problemMessage, ServerResponse} from "../clients/ManagerClient.ts";
import {statusDots, statusLabels} from "./status.ts";
import {UpdateNotification} from "../App.tsx";
import {activeServers} from "./activeServers.ts";
import {latestServers, personalServers} from "./personalServers.ts";
import {Button, LinkButton} from "../ui/Button.tsx";
import {Heading} from "../ui/Heading.tsx";
import {Panel} from "../ui/Panel.tsx";
import {TextInput} from "../ui/Field.tsx";
import {ViewTabs} from "../ui/ViewTabs.tsx";
import {usePolling} from "../utils/usePolling.ts";

const SERVERS_REFRESH_MS = 10_000
const stopped = (server: ServerResponse) => server.status === "stopped" || server.status === "expired"

function shutdownHint(idleShutdownAt: string) {
  const minutes = Math.round((new Date(idleShutdownAt).getTime() - Date.now()) / 60_000)
  return minutes < 1 ? "Shuts down any moment" : `Shuts down in ${minutes}m`
}

export function ServerList() {
  const [params] = useSearchParams()
  const mine = params.get("owner") === "me"
  const {session: {user, loading, error}} = UpdateNotification()
  const initial = useLoaderData<ServerResponse[] | null>()
  const heading = useRef<HTMLHeadingElement>(null)

  return (
    <Panel as="section" className="mx-4 my-6 flex-1 max-w-content p-4 sm:p-6 font-sans">
      <header className="flex flex-wrap items-center justify-between gap-4 mb-6">
        <Heading level={1} id="servers-heading" ref={heading} tabIndex={-1} className="focus-ring">Servers</Heading>
        <LinkButton to="/server-creator">Create server</LinkButton>
      </header>
      <ViewTabs label="Server views" views={[
        {label: "All servers", to: "/my-servers", active: !mine},
        {label: "My servers", to: "/my-servers?owner=me", active: mine},
      ]}/>
      {!mine ? <PublicServers initial={initial}/>
        : loading ? <p role="status">Checking your account...</p>
        : error && !user ? <p role="alert">Login status unavailable. Try refreshing the page.</p>
        : !user ? <div className="flex flex-col items-start gap-4">
          <p>Log in to see your running and previous servers.</p>
          <Link to="/auth/login?returnTo=%2Fmy-servers%3Fowner%3Dme" reloadDocument className="text-action-content hover:underline focus-ring">Log in</Link>
        </div>
        : <PersonalServers key={user.subject} onStarted={() => heading.current?.focus()}/>}
    </Panel>
  )
}

function PublicServers({initial}: {initial: ServerResponse[] | null}) {
  const servers = usePolling(initial, activeServers, SERVERS_REFRESH_MS)
  return <ServerRows servers={servers} mine={false}/>
}

function PersonalServers({onStarted}: {onStarted: () => void}) {
  const [servers, setServers] = useState<ServerResponse[] | null>(null)
  const [error, setError] = useState("")
  const [revision, setRevision] = useState(0)
  const [starting, setStarting] = useState<string | null>(null)
  const [startError, setStartError] = useState<{id: string, message: string} | null>(null)
  const version = useRef(0)

  useEffect(() => {
    let active = true
    let loading = false
    const load = async () => {
      if (loading || document.hidden) return
      loading = true
      const current = version.current
      try {
        const items = await personalServers()
        if (active && current === version.current) {
          setServers(items)
          setError("")
        }
      } catch (err) {
        if (active && current === version.current) setError(problemMessage(err))
      } finally {
        loading = false
      }
    }

    void load()
    const interval = window.setInterval(load, SERVERS_REFRESH_MS)
    document.addEventListener("visibilitychange", load)
    return () => {
      active = false
      window.clearInterval(interval)
      document.removeEventListener("visibilitychange", load)
    }
  }, [revision])

  const start = async (server: ServerResponse) => {
    version.current++
    setStarting(server.id)
    setStartError(null)
    try {
      const {data} = await ManagerClient.createServer({name: server.name, tier: server.tier, preset: server.preset, options: server.options, saveId: server.saveId})
      version.current++
      setServers(current => [data, ...(current ?? [])])
      onStarted()
    } catch (err) {
      setStartError({id: server.id, message: problemMessage(err)})
    } finally {
      setStarting(null)
    }
  }

  return <div className="flex flex-col gap-4">
    <p className="text-sm text-content-muted">Your running and previous servers. Free worlds are kept for 90 days after you last play.</p>
    {error && <div className="flex flex-col items-start gap-2">
      <p role="alert">Could not {servers ? "refresh" : "load"} your servers. {error}</p>
      <Button size="sm" variant="secondary" onClick={() => setRevision(value => value + 1)}>Try again</Button>
    </div>}
    {(!error || servers) && <ServerRows servers={servers ? latestServers(servers) : null} mine start={start} starting={starting} startError={startError}/>}
  </div>
}

function ServerRows({servers, mine, start, starting, startError}: {
  servers: ServerResponse[] | null,
  mine: boolean,
  start?: (server: ServerResponse) => Promise<void>,
  starting?: string | null,
  startError?: {id: string, message: string} | null,
}) {
  const {setNotificationMessage} = UpdateNotification()
  const [search, setSearch] = useState("")
  const startTrigger = useRef<HTMLButtonElement | null>(null)

  useEffect(() => {
    if (startError && !starting) startTrigger.current?.focus()
  }, [startError, starting])
  const query = search.trim().toLowerCase()
  const matching = servers?.filter(server => [server.name, server.hostname, server.options.VERSION ?? ""].some(value => value.toLowerCase().includes(query))) ?? []

  const copyHostname = async (hostname: string) => {
    try {
      await navigator.clipboard.writeText(hostname)
      setNotificationMessage("Hostname copied to clipboard")
    } catch {
      setNotificationMessage("Could not copy hostname")
    }
  }

  return <div>
    <label className="sr-only" htmlFor="server-search">Search servers</label>
    <TextInput id="server-search" type="search" className="w-full mb-4" placeholder="Search by name, hostname or version" value={search} onChange={event => setSearch(event.target.value)}/>
    <div className="max-h-136 overflow-y-auto focus-ring" role="region" aria-labelledby="servers-heading" tabIndex={0}>
      {servers === null ? <p className="py-8 text-content-muted" role="status">Loading servers...</p>
        : servers.length === 0 ? <p className="py-8 text-content-muted">{mine ? "No servers yet. Create a Free server to keep your world between sessions." : "No servers running right now. Create one to get started."}</p>
        : matching.length === 0 ? <p className="py-8 text-content-muted" role="status">No servers match your search.</p>
        : <ul>{matching.map(server => <li key={server.id} className="grid grid-cols-[minmax(0,1fr)_auto] gap-x-4 gap-y-2 py-4 px-1 border-t border-edge">
          <div className="min-w-0">
            <Heading level={3} className="wrap-anywhere">{server.name}</Heading>
            <p className="text-sm text-content-muted">{server.options.VERSION ?? "Unknown version"}</p>
          </div>
          <ServerPresence server={server}/>
          {mine && <p className="col-span-full text-xs text-content-muted">Last started {new Date(server.createdAt).toLocaleDateString()}</p>}
          {mine && stopped(server) ? <div className="col-span-full flex flex-wrap items-center gap-4">
            {server.saveId ? <>
              <Button size="sm" disabled={Boolean(starting) || server.saveLocked || !server.saveAvailable} onClick={async event => {
                startTrigger.current = event.currentTarget
                await start?.(server)
              }}>
                {starting === server.id ? "Starting..." : "Start"}
              </Button>
              <span className="text-sm text-content-muted">{server.saveLocked ? "Saving your world..." : server.saveAvailable ? "Continue with the same world and settings" : "World expired after 90 days without playing"}</span>
            </> : <p className="text-sm text-content-muted">Guest worlds are not kept after shutdown.</p>}
          </div> : <Button variant="secondary" size="sm" className="col-span-full flex items-center justify-between gap-4 text-left"
            aria-label={`Copy hostname ${server.hostname}`} onClick={() => copyHostname(server.hostname)}>
            <span className="font-mono text-xs text-content-muted wrap-anywhere">{server.hostname}</span>
            <span className="shrink-0 text-xs text-action-content">Copy</span>
          </Button>}
          {startError?.id === server.id && <p className="col-span-full text-sm" role="alert">Could not start server. {startError.message}</p>}
        </li>)}</ul>}
    </div>
  </div>
}

function ServerPresence({server}: {server: ServerResponse}) {
  const label = stopped(server) ? server.saveLocked ? "Saving" : "Stopped" : statusLabels[server.status] ?? server.status
  const players = server.status === "running" && server.players
  const hint = server.status === "expiring" && server.idleShutdownAt ? shutdownHint(server.idleShutdownAt) : label

  return <div className="group relative flex items-center gap-2 self-start pt-1 text-xs" tabIndex={hint !== label ? 0 : undefined}>
    <span className={`size-2 shrink-0 ${stopped(server) ? "bg-content-muted" : statusDots[server.status] ?? "bg-content-muted"}`} role="img" aria-label={hint}/>
    <span>{players ? `${players.online}/${players.max}` : label}</span>
    {hint !== label && <span role="tooltip" className="absolute right-0 top-full z-10 hidden pt-1 group-hover:block group-focus:block">
      <span className="block whitespace-nowrap px-2 py-1 bg-inset border border-edge">{hint}</span>
    </span>}
  </div>
}
