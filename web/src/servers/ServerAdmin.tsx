import {useEffect, useRef} from "react";
import {Link, useLoaderData, useParams} from "react-router-dom";
import {CatalogResponse, ServerResponse} from "../clients/ManagerClient.ts";
import {Button} from "../ui/Button.tsx";
import {Details} from "../ui/Details.tsx";
import {Heading} from "../ui/Heading.tsx";
import {Panel} from "../ui/Panel.tsx";
import {AccountServers} from "./AccountServers.tsx";
import {ServerPresence} from "./ServerPresence.tsx";
import {lastPlayed, serverLabels, worldState} from "./serverPresentation.ts";
import {useServers} from "./useServers.ts";

export function ServerAdmin() {
  return <Panel as="section" className="mx-4 my-6 flex-1 max-w-content p-4 sm:p-6 font-sans">
    <Link to="/manage-servers" className="mb-6 inline-block text-sm text-action-content hover:underline focus-ring">‹ Manage servers</Link>
    <AccountServers><OwnedServer/></AccountServers>
  </Panel>
}

function OwnedServer() {
  const {serverId} = useParams()
  const catalog = useLoaderData<CatalogResponse | null>()
  const {servers, error, reload, start, starting, startError} = useServers(true)
  const heading = useRef<HTMLHeadingElement>(null)
  const startButton = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (startError && !starting) startButton.current?.focus()
  }, [startError, starting])

  const requested = servers?.find(server => server.id === serverId)
  const server = requested?.saveId ? servers?.find(item => item.saveId === requested.saveId) : requested

  return <div className="flex flex-col gap-6">
    {error && <div className="flex flex-col items-start gap-2">
      <p role="alert">Could not {servers ? "refresh" : "load"} your server. {error}</p>
      <Button variant="secondary" size="sm" onClick={reload}>Try again</Button>
    </div>}
    {servers === null ? !error && <p role="status">Loading server...</p>
      : !server ? <><Heading level={1}>Server not found</Heading><p>This server is not in your account.</p></>
      : <>
        <header className="flex flex-wrap items-start justify-between gap-4">
          <Heading level={1} ref={heading} tabIndex={-1} className="focus-ring wrap-anywhere">{server.name}</Heading>
          <ServerPresence server={server} managed/>
        </header>
        {worldState(server) === "stopped" && <div className="flex flex-col items-start gap-3">
          <Button ref={startButton} disabled={starting} onClick={async () => {if (await start(server)) heading.current?.focus()}}>{starting ? "Starting..." : "Start"}</Button>
          <p className="text-sm text-content-muted">Continue with the same world and settings.</p>
        </div>}
        {startError && worldState(server) === "stopped" && <p role="alert">Could not start server. {startError}</p>}
        <WorldReadiness server={server}/>
        <section aria-labelledby="server-config-heading">
          <Heading level={2} id="server-config-heading" className="mb-4">Configuration</Heading>
          <Details items={[
            {label: "Version", value: serverLabels(server, catalog).version},
            {label: "Preset", value: serverLabels(server, catalog).preset},
            {label: "Tier", value: serverLabels(server, catalog).tier},
            ...Object.entries(server.options).filter(([key]) => key !== "VERSION").map(([key, value]) => ({label: key.replace(/_/g, " "), value})),
          ]}/>
        </section>
      </>}
  </div>
}

function WorldReadiness({server}: {server: ServerResponse}) {
  const state = worldState(server)
  const message = state === "saving" ? "Saving your world. You can start the server once saving finishes."
    : state === "expired" ? server.saveId ? "This world has expired and cannot be resumed." : "Guest worlds are deleted when the server stops and cannot be resumed."
    : state === "stopped" ? "Your world is saved and ready to start."
    : server.saveId ? "Your world is saved automatically while you play and when the server stops."
    : "Guest worlds are deleted when the server stops."

  return <section aria-labelledby="world-readiness-heading">
    <Heading level={2} id="world-readiness-heading" className="mb-4">Saved world</Heading>
    <div className="bg-inset p-4">
      <p className="mb-4 text-sm">{message}</p>
      <Details items={[
        {label: "Last played", value: lastPlayed(server)},
        {label: "Retention", value: server.saveId ? "90 days without playing" : "Not retained"},
      ]}/>
    </div>
  </section>
}
