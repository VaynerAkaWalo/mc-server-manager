import {ManagerClient, ServerResponse} from "../clients/ManagerClient.ts";

export async function personalServers(): Promise<ServerResponse[]> {
  const items: ServerResponse[] = []
  let cursor: string | undefined
  do {
    const {data} = await ManagerClient.listServers(cursor, true)
    items.push(...data.items)
    cursor = data.nextCursor || undefined
  } while (cursor)

  return items
}

export function latestServers(servers: ServerResponse[]) {
  const seen = new Set<string>()
  return servers.filter(server => {
    const key = server.saveId ?? server.id
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}
