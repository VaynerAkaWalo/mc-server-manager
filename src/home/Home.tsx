import {Link} from "react-router-dom";
import {useEffect, useState} from "react";
import {ManagerClient, StatsResponse} from "../clients/ManagerClient.ts";

const news = [
  {date: "2026-10-25", title: "Alpha 0.0.1", text: "We are now in alpha 0.0.1 and available at mc-server.blamedevs.com."},
]

function Stat({label, value}: { label: string, value?: number }) {
  return (
    <div className="mc-panel flex-1 flex flex-col items-center py-4 px-6">
      {value === undefined
        ? <span className="text-4xl text-gray-200 motion-safe:animate-pulse" aria-label="Loading">...</span>
        : <span className="text-4xl text-gray-200">{value}</span>}
      <span>{label}</span>
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
        <p className="text-3xl">Host a server in minutes and play with your friends</p>
        <Link className="mc-button whitespace-nowrap" to="/server-creator">Get started</Link>
      </div>

      <div className="flex flex-col sm:flex-row gap-4 w-full">
        <Stat label="Servers online" value={stats?.online}/>
        <Stat label="Servers created" value={stats?.created}/>
      </div>

      <section className="mc-panel w-full p-4">
        <h2 className="text-xl text-gray-200 mb-3">What's new</h2>
        <ul className="flex flex-col gap-3">
          {news.map(item => (
            <li key={item.title}>
              <span className="text-sm">{item.date}</span>
              <h3 className="text-gray-200">{item.title}</h3>
              <p>{item.text}</p>
            </li>
          ))}
        </ul>
      </section>
    </div>
  )
}
