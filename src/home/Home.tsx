import {Link} from "react-router-dom";

// TODO: replace with real numbers once the manager exposes stats
const stats = {
  online: 0,
  created: 0,
}

const features = [
  {title: "Play with friends", text: "Spin up your own Minecraft server in a minute and share the IP with your friends."},
  {title: "Free to start", text: "Some server templates are free, so you can try it out without paying anything."},
  {title: "Your rules", text: "Pick a template, tune the settings in the creator and shut it down whenever you like."},
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
    <div className="flex flex-col items-center gap-8 w-full max-w-3xl p-4 my-auto">
      <div className="mc-panel w-full flex flex-col items-center gap-4 py-8 px-4 text-center">
        <h1 className="text-3xl sm:text-4xl text-gray-200">Host your own Minecraft server</h1>
        <p>Create a server to play with your friends, in some cases for free.</p>
        <Link className="mc-button" to="/server-creator">Get started</Link>
      </div>

      <div className="flex flex-col sm:flex-row gap-4 w-full">
        <Stat label="Servers online" value={stats.online}/>
        <Stat label="Servers created" value={stats.created}/>
      </div>

      <ul className="flex flex-col sm:flex-row gap-4 w-full">
        {features.map(feature => (
          <li key={feature.title} className="mc-panel flex-1 p-4">
            <h2 className="text-xl text-gray-200 mb-2">{feature.title}</h2>
            <p>{feature.text}</p>
          </li>
        ))}
      </ul>
    </div>
  )
}
