import {useEffect, useState} from "react";
import {Link, useLoaderData} from "react-router-dom";
import {LinkButton} from "../ui/Button.tsx";
import {Heading} from "../ui/Heading.tsx";
import {Panel} from "../ui/Panel.tsx";
import {ServerResponse, StatsResponse} from "../clients/ManagerClient.ts";
import {statusDots, statusLabels} from "../servers/status.ts";

export type HomeData = { stats: StatsResponse, servers: ServerResponse[] | null }

const news = [
  {date: "2026-10-28", title: "Accounts", text: "You can now log in with your blamedevs account. Logged-in players will get new features soon, and everything keeps working without an account."},
  {date: "2026-10-28", title: "Guest servers", text: "Servers now come as Guest servers: up to 3 players and a smaller view distance, and the world is deleted when the server shuts down. Accounts with more options are coming soon."},
  {date: "2026-10-26", title: "Smoother servers", text: "Servers now come with performance mods, so new worlds generate faster. You still join with the regular game. Snapshots are no longer available."},
  {date: "2026-10-26", title: "Find your server", text: "Search servers by name, hostname or version, and scroll the list to find more worlds."},
  {date: "2026-10-26", title: "One-click servers", text: "Creating a server now fits on one screen and comes with a random name, so you can start playing with a single click."},
  {date: "2026-10-25", title: "No time limit", text: "Servers stay up while you play and stop 30 minutes after the last player leaves."},
  {date: "2026-10-25", title: "Server status", text: "My servers now show whether a server is online and how many players are connected."},
  {date: "2026-10-25", title: "Alpha 0.0.1", text: "We are now in alpha 0.0.1 and available at mc-server.blamedevs.com."},
]

const SERVER_PREVIEW_LIMIT = 4

function Stat({label, value}: { label: string, value: number }) {
  return (
    <div className="flex flex-col gap-1 border-l border-edge pl-6 first:border-l-0 first:pl-0">
      <dt className="order-2 text-sm text-content-muted">{label}</dt>
      <dd className="order-1 font-pixel text-3xl">{value}</dd>
    </div>
  )
}

function formatDate(date: string) {
  return new Date(date + "T00:00:00").toLocaleDateString("en", {month: "short", day: "numeric"})
}

export function Home() {
  const {stats, servers} = useLoaderData<HomeData>()
  const preview = (servers ?? []).slice(0, SERVER_PREVIEW_LIMIT)

  return (
    <div className="flex w-full max-w-content flex-col gap-6 p-4 sm:p-6">
      <Panel as="section" className="flex flex-col gap-6 p-6 sm:flex-row sm:items-end sm:justify-between sm:p-8">
        <div className="flex flex-col gap-6">
          <Heading level={1} className="max-w-lg text-balance">Host a server in minutes and play with your friends</Heading>
          <div className="flex flex-wrap gap-2">
            <LinkButton className="whitespace-nowrap" to="/server-creator">Create server</LinkButton>
            <LinkButton variant="secondary" className="whitespace-nowrap" to="/server-templates">Templates</LinkButton>
          </div>
        </div>
        <dl className="flex gap-6 sm:pb-1">
          <Stat label="Online" value={stats.online}/>
          <Stat label="Created" value={stats.created}/>
        </dl>
      </Panel>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
        <Panel as="section" className="flex flex-col p-6" aria-labelledby="whats-new-heading">
          <Heading level={2} id="whats-new-heading" className="mb-2">What's new</Heading>
          <ol className="-mx-2 max-h-80 overflow-y-auto px-2 focus-ring" tabIndex={0} aria-labelledby="whats-new-heading">
            {news.map(item => (
              <li key={item.title} className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-4 border-t border-edge py-3 first:border-t-0">
                <time dateTime={item.date} className="w-14 pt-0.5 text-xs text-content-muted">{formatDate(item.date)}</time>
                <div>
                  <Heading level={3}>{item.title}</Heading>
                  <p className="text-sm text-content-muted">{item.text}</p>
                </div>
              </li>
            ))}
          </ol>
        </Panel>

        <Panel as="section" className="flex flex-col p-6" aria-labelledby="servers-heading">
          <div className="mb-2 flex items-baseline justify-between gap-4">
            <Heading level={2} id="servers-heading">Servers</Heading>
            {servers && servers.length > 0 &&
              <Link to="/my-servers" className="text-sm text-action-content hover:underline focus-ring">All {servers.length} →</Link>}
          </div>
          {preview.length === 0 ? (
            <div className="flex flex-1 flex-col items-start justify-center gap-4 py-6">
              <p className="text-content-muted">{servers ? "None running" : "Unavailable"}</p>
              <LinkButton variant="secondary" size="sm" to="/server-creator">Create server</LinkButton>
            </div>
          ) : (
            <ul className="flex flex-col gap-2">
              {preview.map(server => <ServerRow key={server.id} server={server}/>)}
            </ul>
          )}
        </Panel>
      </div>
    </div>
  )
}

const COPIED_MS = 1500

function ServerRow({server}: { server: ServerResponse }) {
  const [copied, setCopied] = useState(false)
  const players = server.status === "running" && server.players
  const label = statusLabels[server.status] ?? server.status

  useEffect(() => {
    if (!copied) return
    const timer = setTimeout(() => setCopied(false), COPIED_MS)
    return () => clearTimeout(timer)
  }, [copied])

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(server.hostname)
      setCopied(true)
    } catch {
      setCopied(false)
    }
  }

  return (
    <li>
      <button type="button" aria-label={`Copy ${server.hostname}`} onClick={copy}
              className="group flex w-full cursor-pointer items-center gap-3 border border-edge bg-inset px-3 py-2 text-left hover:border-edge-control hover:bg-surface-hover focus-ring">
        <span className={`size-2 shrink-0 ${statusDots[server.status] ?? "bg-content-muted"}`} aria-hidden="true"/>
        <span className="min-w-0 flex-1">
          <span className="block truncate font-pixel text-sm">{server.name}</span>
          <span className="block truncate text-xs text-content-muted">{server.options.VERSION ?? "Unknown version"}</span>
        </span>
        <span className="shrink-0 text-xs text-content-muted">{players ? `${players.online}/${players.max}` : label}</span>
        <span aria-hidden="true" className={`grid size-4 shrink-0 place-items-center ${copied ? "text-action-content" : "text-content-muted opacity-0 group-hover:opacity-100 group-focus-visible:opacity-100"}`}>
          {copied
            ? <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" className="size-4"><path d="m3 8 3.5 3.5L13 5"/></svg>
            : <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className="size-4"><rect x="5.5" y="5.5" width="8" height="8"/><path d="M10.5 3.5v-1h-8v8h1"/></svg>}
        </span>
      </button>
      <span role="status" className="sr-only">{copied ? "Copied" : ""}</span>
    </li>
  )
}
