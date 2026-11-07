import {CatalogResponse, ServerResponse} from "../clients/ManagerClient.ts";
import {UpdateNotification} from "../App.tsx";
import {ListRow, ListRowStrip} from "../ui/ListRow.tsx";
import {ServerPresence} from "./ServerPresence.tsx";
import {lastPlayed, serverLabels, worldState} from "./serverPresentation.ts";

export function ServerRow({server, managed = false, catalog = null}: {
  server: ServerResponse,
  managed?: boolean,
  catalog?: CatalogResponse | null,
}) {
  const {setNotificationMessage} = UpdateNotification()
  const labels = serverLabels(server, catalog)
  const state = worldState(server)

  const copyHostname = async () => {
    try {
      await navigator.clipboard.writeText(server.hostname)
      setNotificationMessage("Hostname copied to clipboard")
    } catch {
      setNotificationMessage("Could not copy hostname")
    }
  }

  const summary = state === "stopped" ? `World saved · ${server.lastPlayedAt ? `Last played ${lastPlayed(server)}` : "Not played yet"}`
    : state === "saving" ? "Saving your world…"
    : state === "expired" ? server.saveId ? "World expired · Cannot resume" : "Guest world deleted · Cannot resume"
    : null

  return <ListRow name={server.name} metadata={managed ? `${labels.version} · ${labels.preset} · ${labels.tier}` : labels.version}
    status={<ServerPresence server={server} managed={managed}/>}
    to={managed ? `/manage-servers/${server.id}` : undefined}
    strip={managed
      ? <ListRowStrip action="Manage ›">
        {summary ? <span>{summary}</span> : <span className={`font-mono wrap-anywhere ${state === "new" ? "opacity-60" : ""}`}>{server.hostname}</span>}
      </ListRowStrip>
      : <ListRowStrip action="Copy" label={`Copy hostname ${server.hostname}`} onClick={copyHostname}>
        <span className="font-mono wrap-anywhere">{server.hostname}</span>
      </ListRowStrip>}/>
}
