import {HTTPClient} from "./Client.ts";
import {AxiosPromise, AxiosResponse, isAxiosError} from "axios";

export interface CreateServerRequest {
  name: string,
  options: Opts,
  durationMs: number,
  tier: Tier,
  preset?: string
}

export interface Opts {
  ONLINE_MODE?: string
  DIFFICULTY?: string
  MAX_PLAYERS?: string
  VERSION?: string
  TYPE?: string
  MOTD?: string
  OPS?: string
}

export type Tier = string

export interface GameVersion {
  id: string,
  kind: "release" | "snapshot"
}

export interface CatalogResponse {
  tiers: { id: Tier, name: string, description: string, default: boolean }[],
  presets: { id: string, name: string, description: string, options: Opts }[],
  options: { id: string, values?: string[] }[],
  durationMs: { min: number, max: number },
  versions: Record<string, { latestRelease: string, items: GameVersion[] }>
}

export type ServerStatus = "new" | "running" | "degraded" | "stopped" | "expired"

export interface ServerResponse {
  id: string,
  name: string,
  hostname: string,
  status: ServerStatus,
  players: { online: number, max: number } | null,
  tier: Tier,
  preset?: string,
  options: Opts,
  createdAt: string,
  expiresAt: string
}

export interface ServerCollection {
  items: ServerResponse[],
  nextCursor: string
}

export interface StatsResponse {
  online: number,
  created: number
}

export interface Problem {
  title: string,
  status: number,
  detail?: string
}

export const problemMessage = (err: unknown): string => {
  if (isAxiosError<Problem>(err) && err.response?.data?.detail) {
    return err.response.data.detail
  }
  return "Something went wrong, try again later"
}

class ServerManager {
  public createServer = (request: CreateServerRequest): Promise<AxiosPromise<ServerResponse>> => {
    return HTTPClient.post("/api/v1/servers", request)
  }

  public listServers = async (cursor?: string): Promise<AxiosPromise<ServerCollection>> => {
    const query = new URLSearchParams(cursor ? {cursor} : {})
    return HTTPClient.get(`/api/v1/servers?${query}`)
  }

  public getCatalog = async (): Promise<AxiosResponse<CatalogResponse>> => {
    return HTTPClient.get("/api/v1/catalog")
  }

  public getStats = async (): Promise<AxiosPromise<StatsResponse>> => {
    return HTTPClient.get("/api/v1/servers/stats?status=all")
  }
}

export const ManagerClient = new ServerManager()
