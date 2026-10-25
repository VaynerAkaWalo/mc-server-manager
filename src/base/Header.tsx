import {Link} from "react-router-dom";

export function Header() {

  return (
    <ul className="flex flex-row flex-wrap bg-surface border-b-4 border-edge justify-around items-center w-full min-h-12">
      <li><Link className="mc-nav-link" to={"/"}>Home</Link></li>
      <li><Link className="mc-nav-link" to={"/my-servers"}>Servers</Link></li>
      <li><Link className="mc-nav-link" to={"/server-creator"}>Creator</Link></li>
      <li><Link className="mc-nav-link" to={"/server-templates"}>Templates</Link></li>
    </ul>
  )
}
