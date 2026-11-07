import {CatalogResponse, ServerResponse} from "../clients/ManagerClient.ts";

export const isStopped = (server: ServerResponse) => server.status === "stopped" || server.status === "expired"

export function worldState(server: ServerResponse) {
  if (!isStopped(server)) return server.status
  if (server.saveLocked) return "saving"
  return server.saveId && server.saveAvailable ? "stopped" : "expired"
}

export function serverLabels(server: ServerResponse, catalog: CatalogResponse | null) {
  return {
    version: server.options.VERSION ?? "Unknown version",
    preset: server.preset ? catalog?.presets.find(preset => preset.id === server.preset)?.name ?? server.preset : "Custom",
    tier: catalog?.tiers.find(tier => tier.id === server.tier)?.name ?? server.tier,
  }
}

export function lastPlayed(server: ServerResponse) {
  if (!server.lastPlayedAt) return "Not played yet"
  return new Date(server.lastPlayedAt).toLocaleString(undefined, {dateStyle: "medium", timeStyle: "short"})
}
