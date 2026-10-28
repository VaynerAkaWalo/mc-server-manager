import {NavLink, useLocation} from "react-router-dom";
import {useUser} from "../auth/useUser.ts";

export function Header() {
  const {user, loading, error, logOut} = useUser()
  const location = useLocation()

  const logIn = () => {
    window.location.assign("/auth/login?returnTo=" + encodeURIComponent(location.pathname + location.search + location.hash))
  }

  return (
    <ul className="flex flex-row flex-wrap bg-surface border-b border-edge-strong justify-around items-center w-full min-h-12">
      <li><NavLink className="mc-nav-link" to={"/"} end>Home</NavLink></li>
      <li><NavLink className="mc-nav-link" to={"/my-servers"}>Servers</NavLink></li>
      <li><NavLink className="mc-nav-link" to={"/server-creator"}>Creator</NavLink></li>
      <li><NavLink className="mc-nav-link" to={"/server-templates"}>Templates</NavLink></li>
      <li className="max-w-full px-3">
        <div className="flex flex-wrap items-center justify-center gap-x-3">
          {user?.email && <span className="break-all text-sm text-content-muted">{user.email}</span>}
          {user
            ? <button className="mc-nav-link cursor-pointer" disabled={loading} onClick={logOut}>Log out</button>
            : <button className="mc-nav-link cursor-pointer" disabled={loading} onClick={logIn}>Log in</button>}
        </div>
        {error && <span role="status" className="text-sm text-content-muted"> Login status unavailable. Try again.</span>}
      </li>
    </ul>
  )
}
