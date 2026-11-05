import {FormEvent, useEffect, useRef, useState} from "react";
import {UpdateNotification} from "../App.tsx";
import {ManagerClient, problemMessage, WorldResponse} from "../clients/ManagerClient.ts";
import {Button, LinkButton} from "../ui/Button.tsx";
import {Field, TextInput} from "../ui/Field.tsx";
import {Heading} from "../ui/Heading.tsx";
import {Panel} from "../ui/Panel.tsx";
import {useWorlds} from "./useWorlds.ts";

export function WorldList() {
  const {session: {user, loading: sessionLoading, error: sessionError}} = UpdateNotification()
  const {worlds, setWorlds, loading, error, reload} = useWorlds(user?.subject)
  const heading = useRef<HTMLHeadingElement>(null)
  const [name, setName] = useState("")
  const [busy, setBusy] = useState(false)
  const [failure, setFailure] = useState("")

  const create = async (event: FormEvent) => {
    event.preventDefault()
    setBusy(true)
    setFailure("")

    try {
      const {data} = await ManagerClient.createWorld(name.trim())
      setWorlds(current => [...current, data])
      setName("")
    } catch (err) {
      setFailure(problemMessage(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Panel as="section" className="mx-4 my-6 flex w-full max-w-content flex-col gap-6 p-4 sm:p-6">
      <header className="flex flex-wrap items-baseline justify-between gap-2">
        <Heading level={1} ref={heading} tabIndex={-1} className="focus-ring">My worlds</Heading>
        {user && !loading && !error && <span className="text-sm text-content-muted">{worlds.length} {worlds.length === 1 ? "world" : "worlds"}</span>}
      </header>
      <p className="text-sm text-content-muted">Return to your world on a Free server. Worlds are deleted after 90 days without playing.</p>

      {sessionLoading ? <p role="status">Checking your account...</p>
        : sessionError && !user ? <p role="alert">Login status unavailable. Try refreshing the page.</p>
        : !user ? <div className="flex flex-col items-start gap-4">
          <p>Log in to keep your worlds between sessions.</p>
          <Button onClick={() => window.location.assign("/auth/login?returnTo=%2Fmy-worlds")}>Log in</Button>
        </div>
        : loading ? <p role="status">Loading worlds...</p>
        : error ? <div className="flex flex-col items-start gap-4">
          <p role="alert">Could not load worlds. {error}</p>
          <Button variant="secondary" onClick={reload}>Try again</Button>
        </div>
        : <>
          <form onSubmit={create} className="flex flex-col gap-4 sm:flex-row sm:items-end">
            <div className="flex-1">
              <Field label="New world name" htmlFor="world-name">
                <TextInput id="world-name" value={name} onChange={event => setName(event.target.value)} maxLength={40} required
                           disabled={busy} autoComplete="off"/>
              </Field>
            </div>
            <Button type="submit" disabled={busy || !name.trim()}>{busy ? "Creating..." : "Create world"}</Button>
          </form>
          {failure && <p role="alert">{failure}</p>}
          {worlds.length === 0 ? <p className="py-6 text-content-muted">No worlds yet. Create your first world above, then start a server.</p>
            : <ul>{worlds.map(world => <WorldRow key={world.id} world={world}
                onRename={updated => setWorlds(current => current.map(item => item.id === updated.id ? updated : item))}
                onDelete={() => {setWorlds(current => current.filter(item => item.id !== world.id)); heading.current?.focus()}}/>)}</ul>}
        </>}
    </Panel>
  )
}

function WorldRow({world, onRename, onDelete}: {world: WorldResponse, onRename: (world: WorldResponse) => void, onDelete: () => void}) {
  const [mode, setMode] = useState<"view" | "rename" | "delete">("view")
  const renameButton = useRef<HTMLButtonElement>(null)
  const deleteButton = useRef<HTMLButtonElement>(null)
  const returnFocus = useRef<"rename" | "delete" | null>(null)
  const [name, setName] = useState(world.name)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => {
    if (mode !== "view" || !returnFocus.current) return

    const action = returnFocus.current === "rename" ? renameButton : deleteButton
    action.current?.focus()
    returnFocus.current = null
  }, [mode])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setBusy(true)
    setError("")

    try {
      if (mode === "rename") {
        const {data} = await ManagerClient.renameWorld(world.id, name.trim())
        onRename(data)
        setMode("view")
      } else {
        await ManagerClient.deleteWorld(world.id)
        onDelete()
      }
    } catch (err) {
      setError(problemMessage(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <li className="flex flex-col gap-4 border-t border-edge py-4">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="min-w-0">
          <Heading level={3} className="wrap-anywhere">{world.name}</Heading>
          <p className="text-xs text-content-muted">Created {new Date(world.createdAt).toLocaleDateString()}</p>
        </div>
        {mode === "view" && <div className="flex flex-wrap gap-2">
          <LinkButton size="sm" to="/server-creator" state={{worldId: world.id}}>Start server</LinkButton>
          <Button ref={renameButton} size="sm" variant="secondary" onClick={() => {returnFocus.current = "rename"; setName(world.name); setMode("rename"); setError("")}}>Rename</Button>
          <Button ref={deleteButton} size="sm" variant="ghost" onClick={() => {returnFocus.current = "delete"; setMode("delete"); setError("")}}>Delete</Button>
        </div>}
      </div>
      {mode !== "view" && <form onSubmit={submit} className="flex flex-col items-start gap-4 bg-inset p-4">
        {mode === "rename" ? <Field label="World name" htmlFor={`rename-${world.id}`}>
          <TextInput id={`rename-${world.id}`} autoFocus required maxLength={40} value={name} disabled={busy} onChange={event => setName(event.target.value)}/>
        </Field> : <p>Delete “{world.name}” and all its backups? This cannot be undone. A world still in use cannot be deleted.</p>}
        {error && <p role="alert">{error}</p>}
        <div className="flex flex-wrap gap-2">
          <Button type="submit" disabled={busy || (mode === "rename" && !name.trim())}>
            {busy ? "Saving..." : mode === "rename" ? "Save name" : "Delete world"}
          </Button>
          <Button variant="secondary" autoFocus={mode === "delete"} disabled={busy} onClick={() => setMode("view")}>Cancel</Button>
        </div>
      </form>}
    </li>
  )
}
