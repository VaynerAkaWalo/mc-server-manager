import {NavLink, useLocation} from "react-router-dom";
import {useUser} from "../auth/useUser.ts";
import {userManager} from "../auth/userManager.ts";

export function Header() {
  const user = useUser()
  const location = useLocation()

  const logIn = () => userManager.signinRedirect({state: location.pathname + location.search})
  const logOut = () => userManager.removeUser()

  return (
    <ul className="flex flex-row flex-wrap bg-surface border-b border-edge-strong justify-around items-center w-full min-h-12">
      <li><NavLink className="mc-nav-link" to={"/"} end>Home</NavLink></li>
      <li><NavLink className="mc-nav-link" to={"/my-servers"}>Servers</NavLink></li>
      <li><NavLink className="mc-nav-link" to={"/server-creator"}>Creator</NavLink></li>
      <li><NavLink className="mc-nav-link" to={"/server-templates"}>Templates</NavLink></li>
      <li>
        {user
          ? <button className="mc-nav-link cursor-pointer" onClick={logOut}>Log out</button>
          : <button className="mc-nav-link cursor-pointer" onClick={logIn}>Log in</button>}
      </li>
    </ul>
  )
}
