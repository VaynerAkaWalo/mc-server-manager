import {Link} from "react-router-dom";
import {useEffect, useState} from "react";
import {ManagerClient, StatsResponse} from "../clients/ManagerClient.ts";

const news = [
  {date: "2026-10-26", title: "Smoother servers", text: "Servers now come with performance mods, so new worlds generate faster. You still join with the regular game. Snapshots are no longer available."},
  {date: "2026-10-26", title: "Find your server", text: "Search servers by name, hostname or version, and scroll the list to find more worlds."},
  {date: "2026-10-26", title: "One-click servers", text: "Creating a server now fits on one screen and comes with a random name, so you can start playing with a single click."},
  {date: "2026-10-25", title: "No time limit", text: "Servers stay up while you play and stop 30 minutes after the last player leaves."},
  {date: "2026-10-25", title: "Server status", text: "My servers now show whether a server is online and how many players are connected."},
  {date: "2026-10-25", title: "Alpha 0.0.1", text: "We are now in alpha 0.0.1 and available at mc-server.blamedevs.com."},
]

function Stat({label, value}: { label: string, value?: number }) {
  return (
    <div className="mc-panel flex-1 flex flex-col items-center gap-1 py-4 px-6">
      {value === undefined
        ? <span className="font-pixel text-3xl motion-safe:animate-pulse" aria-label="Loading">...</span>
        : <span className="font-pixel text-3xl">{value}</span>}
      <span className="text-sm text-content-muted">{label}</span>
    </div>
  )
}

export function Home() {
  const [stats, setStats] = useState<StatsResponse>()

  useEffect(() => {
    const loadStats = async () => {
      try {
        const {data} = await ManagerClient.getStats()
        setStats(data)
      } catch {
        setStats({online: 0, created: 0})
      }
    }

    loadStats()
  }, [])

  return (
    <div className="flex flex-col items-center gap-6 w-full max-w-3xl p-4 my-auto">
      <div className="mc-panel w-full flex flex-col items-center gap-6 py-8 px-6 text-center">
        <h1 className="font-pixel text-2xl sm:text-3xl text-balance">Host a server in minutes and play with your friends</h1>
        <Link className="mc-button whitespace-nowrap" to="/server-creator">Get started</Link>
      </div>

      <div className="flex flex-col sm:flex-row gap-4 w-full">
        <Stat label="Servers online" value={stats?.online}/>
        <Stat label="Servers created" value={stats?.created}/>
      </div>

      <section className="mc-panel w-full p-6">
        <h2 id="whats-new-heading" className="font-pixel text-xl mb-2">What's new</h2>
        <ul className="max-h-80 sm:max-h-64 overflow-y-auto focus-visible:outline focus-visible:outline-action-content" tabIndex={0} aria-labelledby="whats-new-heading">
          {news.map(item => (
            <li key={item.title} className="py-3 border-t border-edge first:border-t-0">
              <div className="flex flex-wrap items-baseline justify-between gap-x-4">
                <h3 className="font-pixel">{item.title}</h3>
                <span className="text-xs text-content-muted">{item.date}</span>
              </div>
              <p className="text-sm text-content-muted">{item.text}</p>
            </li>
          ))}
        </ul>
      </section>
    </div>
  )
}
