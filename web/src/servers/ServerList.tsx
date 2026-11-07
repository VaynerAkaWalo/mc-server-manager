import {useState} from "react";
import {useLoaderData} from "react-router-dom";
import {CatalogResponse, ServerResponse} from "../clients/ManagerClient.ts";
import {Button, LinkButton} from "../ui/Button.tsx";
import {Heading} from "../ui/Heading.tsx";
import {Panel} from "../ui/Panel.tsx";
import {TextInput} from "../ui/Field.tsx";
import {AccountServers} from "./AccountServers.tsx";
import {ServerRow} from "./ServerRow.tsx";
import {latestServers} from "./personalServers.ts";
import {useServers} from "./useServers.ts";

export function ServerList({managed = false}: {managed?: boolean}) {
  return <Panel as="section" className="mx-4 my-6 flex-1 max-w-content p-4 sm:p-6 font-sans">
    <header className="flex flex-wrap items-center justify-between gap-4 mb-6">
      <Heading level={1} id="servers-heading">{managed ? "Manage servers" : "Server list"}</Heading>
      <LinkButton to="/server-creator">Create server</LinkButton>
    </header>
    {managed ? <AccountServers><PersonalServers/></AccountServers> : <PublicServers/>}
  </Panel>
}

function PublicServers() {
  const state = useServers(false)
  return <ServerCollection {...state} managed={false}/>
}

function PersonalServers() {
  const catalog = useLoaderData<CatalogResponse | null>()
  const state = useServers(true)
  return <ServerCollection {...state} servers={state.servers ? latestServers(state.servers) : null} managed catalog={catalog}/>
}

function ServerCollection({servers, error, reload, managed, catalog = null}: {
  servers: ServerResponse[] | null,
  error: string,
  reload: () => void,
  managed: boolean,
  catalog?: CatalogResponse | null,
}) {
  const [search, setSearch] = useState("")
  const query = search.trim().toLowerCase()
  const matching = servers?.filter(server => [server.name, server.hostname, server.options.VERSION ?? ""].some(value => value.toLowerCase().includes(query))) ?? []

  return <div>
    {error && <div className="flex flex-col items-start gap-2 mb-4">
      <p role="alert">Could not {servers ? "refresh" : "load"} servers. {error}</p>
      <Button size="sm" variant="secondary" onClick={reload}>Try again</Button>
    </div>}
    <label className="sr-only" htmlFor="server-search">Search servers</label>
    <TextInput id="server-search" type="search" className="w-full mb-4" placeholder="Search by name, hostname or version" value={search} onChange={event => setSearch(event.target.value)}/>
    <div className="max-h-136 overflow-y-auto focus-ring" role="region" aria-labelledby="servers-heading" tabIndex={0}>
      {servers === null ? !error && <p className="py-8 text-content-muted" role="status">Loading servers...</p>
        : servers.length === 0 ? <p className="py-8 text-content-muted">{managed ? "No servers yet. Create a Free server to keep your world between sessions." : "No servers running right now. Create one to get started."}</p>
        : matching.length === 0 ? <p className="py-8 text-content-muted" role="status">No servers match your search.</p>
        : <ul>{matching.map(server => <ServerRow key={server.id} server={server} managed={managed} catalog={catalog}/>)}</ul>}
    </div>
  </div>
}
