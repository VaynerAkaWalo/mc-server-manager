import {Link} from "react-router-dom";

export function Header() {

  return (
    <ul className="flex flex-row bg-surface border-b-4 border-edge justify-around items-center w-full h-12">
      <li><Link className="mc-nav-link" to={"/"}>Servers</Link></li>
      <li><Link className="mc-nav-link" to={"/server-creator"}>Creator</Link></li>
      <li><Link className="mc-nav-link" to={"/server-templates"}>Templates</Link></li>
    </ul>
  )
}
