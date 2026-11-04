import {useEffect, useRef, useState} from "react";

export function usePolling<T>(initial: T, load: (current: T) => Promise<T>, intervalMs: number): T {
  const [value, setValue] = useState(initial)
  const valueRef = useRef(value)
  const loadRef = useRef(load)
  const requestVersion = useRef(0)
  valueRef.current = value
  loadRef.current = load

  useEffect(() => {
    let timer: number | undefined
    let active = true

    const refresh = () => {
      const version = ++requestVersion.current
      loadRef.current(valueRef.current).then(next => {
        if (active && version === requestVersion.current) setValue(next)
      }, () => undefined)
    }
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
      active = false
      stop()
      document.removeEventListener("visibilitychange", onVisibilityChange)
    }
  }, [intervalMs])

  return value
}
