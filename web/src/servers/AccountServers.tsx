import {ReactNode} from "react";
import {Link, useLocation} from "react-router-dom";
import {UpdateNotification} from "../App.tsx";

export function AccountServers({children}: {children: ReactNode}) {
  const {session: {user, loading, error}} = UpdateNotification()
  const location = useLocation()
  const returnTo = encodeURIComponent(location.pathname)

  if (loading) return <p role="status">Checking your account...</p>
  if (error && !user) return <p role="alert">Login status unavailable. Try refreshing the page.</p>
  if (!user) return <div className="flex flex-col items-start gap-4">
    <p>Log in to manage your servers and saved worlds.</p>
    <Link to={`/auth/login?returnTo=${returnTo}`} reloadDocument className="text-action-content hover:underline focus-ring">Log in</Link>
  </div>

  return <div key={user.subject}>{children}</div>
}
