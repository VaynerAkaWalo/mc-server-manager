import {Link} from "react-router-dom";

interface ServerProps {
  id: number
  active: boolean
  name?: string
  ip?: string
  remainingTime?: string
  status?: string
}

export function Server({active, name, ip, remainingTime, status}: ServerProps) {

  const expire = () => {
    if (!remainingTime || remainingTime === "0s") {
      return "soon"
    }

    if (remainingTime.startsWith("-")) {
      return "expiring"
    }

    return remainingTime.replace("0s", "")
  }

  const activeServer = () => {
    return (
      <ul className="flex flex-col h-full justify-center items-center">
        <li>Name: {name}</li>
        <li>Expire: {expire()}</li>
        <li>Server IP: {ip}</li>
        <li>Status: {status}</li>
      </ul>
    )
  }

  return (
    <div className="mc-panel w-full h-full my-3">
      {!active &&
        <div className="flex h-full justify-center items-center">
            <Link className="mc-button px-2 w-1/3" to={"/server-creator"}>Claim server</Link>
        </div>}
      {active && activeServer()}
    </div>
  )
}
