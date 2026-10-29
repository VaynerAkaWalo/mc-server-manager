import {ServerStatus} from "../clients/ManagerClient.ts";

export const statusDots: Partial<Record<ServerStatus, string>> = {
  new: "bg-starting motion-safe:animate-pulse",
  running: "bg-online",
  degraded: "bg-warning",
  expiring: "bg-warning"
}

export const statusLabels: Record<ServerStatus, string> = {
  new: "Starting",
  running: "Online",
  degraded: "Not responding",
  expiring: "Shutting down soon",
  stopped: "Stopped",
  expired: "Expired"
}
