import {useEffect, useRef, useState} from "react";
import {ManagerClient, problemMessage, ServerResponse} from "../clients/ManagerClient.ts";
import {activeServers} from "./activeServers.ts";
import {personalServers} from "./personalServers.ts";

export function useServers(mine: boolean) {
  const [servers, setServers] = useState<ServerResponse[] | null>(null)
  const [error, setError] = useState("")
  const [revision, setRevision] = useState(0)
  const [starting, setStarting] = useState(false)
  const [startError, setStartError] = useState("")
  const version = useRef(0)

  useEffect(() => {
    let active = true
    let loading = false
    const load = async () => {
      if (loading || document.hidden) return
      loading = true
      const current = version.current
      try {
        const items = await (mine ? personalServers() : activeServers())
        if (active && current === version.current) {
          setServers(items)
          setError("")
        }
      } catch (err) {
        if (active && current === version.current) setError(problemMessage(err))
      } finally {
        loading = false
      }
    }

    void load()
    const interval = window.setInterval(load, 10_000)
    document.addEventListener("visibilitychange", load)
    return () => {
      active = false
      window.clearInterval(interval)
      document.removeEventListener("visibilitychange", load)
    }
  }, [mine, revision])

  const start = async (server: ServerResponse) => {
    version.current++
    setStarting(true)
    setStartError("")
    try {
      const {data} = await ManagerClient.createServer({name: server.name, tier: server.tier, preset: server.preset, options: server.options, saveId: server.saveId})
      version.current++
      setServers(current => [data, ...(current ?? [])])
      return true
    } catch (err) {
      setStartError(problemMessage(err))
      return false
    } finally {
      setStarting(false)
    }
  }

  return {servers, error, reload: () => setRevision(value => value + 1), start, starting, startError}
}
