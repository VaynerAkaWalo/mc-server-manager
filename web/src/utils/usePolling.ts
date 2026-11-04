import {useEffect, useRef, useState} from "react";

export function usePolling<T>(initial: T, load: () => Promise<T>, intervalMs: number): T {
  const [value, setValue] = useState(initial)
  const loadRef = useRef(load)
  loadRef.current = load

  useEffect(() => {
    let timer: number | undefined

    const refresh = () => loadRef.current().then(setValue, () => undefined)
    const stop = () => window.clearInterval(timer)
    const start = () => {
      stop()
      timer = window.setInterval(refresh, intervalMs)
    }
    const onVisibilityChange = () => {
      if (document.hidden) return stop()
      refresh()
      start()
    }

    if (!document.hidden) start()
    document.addEventListener("visibilitychange", onVisibilityChange)

    return () => {
      stop()
      document.removeEventListener("visibilitychange", onVisibilityChange)
    }
  }, [intervalMs])

  return value
}
