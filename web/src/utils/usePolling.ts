import {useEffect, useRef, useState} from "react";

export function usePolling<T>(initial: T, load: (current: T) => Promise<T>, intervalMs: number): T {
  const [value, setValue] = useState(initial)
  const valueRef = useRef(value)
  const loadRef = useRef(load)
  const inFlight = useRef(false)
  valueRef.current = value
  loadRef.current = load

  useEffect(() => {
    let timer: number | undefined
    let active = true

    const refresh = () => {
      if (inFlight.current) return
      inFlight.current = true

      loadRef.current(valueRef.current)
        .then(next => {
          if (active) setValue(next)
        }, () => undefined)
        .finally(() => {
          inFlight.current = false
        })
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
