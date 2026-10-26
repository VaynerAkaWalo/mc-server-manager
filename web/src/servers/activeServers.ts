import {ManagerClient, ServerResponse, ServerStatus} from "../clients/ManagerClient.ts";

const activeStatuses: ServerStatus[] = ["new", "running", "degraded", "expiring"]

export async function activeServers(): Promise<ServerResponse[]> {
  const items: ServerResponse[] = []
  let cursor: string | undefined
  do {
    const {data} = await ManagerClient.listServers(cursor)
    items.push(...data.items)
    cursor = data.nextCursor || undefined
  } while (cursor)

  return items.filter(server => activeStatuses.includes(server.status))
}
