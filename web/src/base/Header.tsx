import {NavLink} from "react-router-dom";

export function Header() {

  return (
    <ul className="flex flex-row flex-wrap bg-surface border-b border-edge-strong justify-around items-center w-full min-h-12">
      <li><NavLink className="mc-nav-link" to={"/"} end>Home</NavLink></li>
      <li><NavLink className="mc-nav-link" to={"/my-servers"}>Servers</NavLink></li>
      <li><NavLink className="mc-nav-link" to={"/server-creator"}>Creator</NavLink></li>
      <li><NavLink className="mc-nav-link" to={"/server-templates"}>Templates</NavLink></li>
    </ul>
  )
}
