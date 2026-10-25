import {Link} from "react-router-dom";

// TODO: replace with real numbers once the manager exposes stats
const stats = {
  online: 0,
  created: 0,
}

// TODO: placeholder entries until there is a real changelog source
const news = [
  {date: "2026-10-25", title: "New home page", text: "A place to start from, with a quick way into the server creator."},
  {date: "2026-10-20", title: "Better small screens", text: "The creator and navigation now stay usable on narrow viewports."},
]

function Stat({label, value}: { label: string, value: number }) {
  return (
    <div className="mc-panel flex-1 flex flex-col items-center py-4 px-6">
      <span className="text-4xl text-gray-200">{value}</span>
      <span>{label}</span>
    </div>
  )
}

export function Home() {
  return (
    <div className="flex flex-col items-center gap-6 w-full max-w-3xl p-4 my-auto">
      <div className="mc-panel w-full flex flex-col sm:flex-row items-center justify-between gap-4 py-4 px-6 text-center sm:text-left">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl text-gray-200">Host your own Minecraft server</h1>
          <p>Create a server in a minute and play with your friends.</p>
        </div>
        <Link className="mc-button whitespace-nowrap" to="/server-creator">Get started</Link>
      </div>

      <div className="flex flex-col sm:flex-row gap-4 w-full">
        <Stat label="Servers online" value={stats.online}/>
        <Stat label="Servers created" value={stats.created}/>
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
