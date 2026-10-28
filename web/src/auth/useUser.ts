import {useEffect, useRef, useState} from "react";

type Session = {
  subject: string
  expiresAt: string
}

export function useUser() {
  const [user, setUser] = useState<Session | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const requestVersion = useRef(0)
  const loggingOut = useRef(false)

  useEffect(() => {
    const controller = new AbortController()

    const refresh = async () => {
      if (loggingOut.current) return
      const version = ++requestVersion.current

      try {
        const response = await fetch("/api/v1/sessions/current", {signal: controller.signal})
        if (!response.ok && response.status !== 401) throw new Error("Session unavailable")

        const session: Session | null = response.status === 401 ? null : await response.json()
        if (controller.signal.aborted || version !== requestVersion.current) return

        setUser(session)
        setError(false)
      } catch {
        if (!controller.signal.aborted && version === requestVersion.current) setError(true)
      } finally {
        if (!controller.signal.aborted && version === requestVersion.current) setLoading(false)
      }
    }

    void refresh()
    window.addEventListener("focus", refresh)
    const interval = window.setInterval(refresh, 60_000)

    return () => {
      controller.abort()
      window.removeEventListener("focus", refresh)
      window.clearInterval(interval)
    }
  }, [])

  useEffect(() => {
    if (!user) return

    const timeout = window.setTimeout(() => setUser(null), Math.max(0, Date.parse(user.expiresAt) - Date.now()))
    return () => window.clearTimeout(timeout)
  }, [user])

  const logOut = async () => {
    loggingOut.current = true
    requestVersion.current++
    setLoading(true)


    try {
      const response = await fetch("/api/v1/sessions/current", {method: "DELETE"})
      if (!response.ok) throw new Error("Logout failed")

      setUser(null)
      setError(false)
    } catch {
      setError(true)
    } finally {
      loggingOut.current = false
      setLoading(false)
    }
  }

  return {user, loading, error, logOut}
}
