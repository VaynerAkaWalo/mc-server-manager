import {useCallback, useEffect, useState} from "react";
import {ManagerClient, problemMessage, WorldResponse} from "../clients/ManagerClient.ts";

export function useWorlds(subject: string | undefined) {
  const [worlds, setWorlds] = useState<WorldResponse[]>([])
  const [loading, setLoading] = useState(Boolean(subject))
  const [error, setError] = useState("")
  const [revision, setRevision] = useState(0)
  const reload = useCallback(() => setRevision(value => value + 1), [])

  useEffect(() => {
    let active = true
    setWorlds([])
    setError("")
    setLoading(Boolean(subject))
    if (!subject) return

    const load = async () => {
      const items: WorldResponse[] = []
      let cursor: string | undefined

      do {
        const {data} = await ManagerClient.listWorlds(cursor)
        items.push(...data.items)
        cursor = data.nextCursor || undefined
      } while (cursor && active)

      if (active) setWorlds(items)
    }

    load().catch(err => {
      if (active) setError(problemMessage(err))
    }).finally(() => {
      if (active) setLoading(false)
    })

    return () => { active = false }
  }, [subject, revision])

  return {worlds, setWorlds, loading, error, reload}
}
