import {HTTPClient} from "./Client.ts";
import {AxiosPromise, isAxiosError} from "axios";

export interface CreateServerRequest {
  name: string,
  opts: Opts,
  duration: number,
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

export enum Tier {
  wooden = "wooden",
  iron = "iron"
}

export interface ServerResponse {
  name: string,
  IP: string,
  remainingTime: string,
  status: string,
  version?: string,
  maxPlayers?: number
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
    return HTTPClient.post("/servers", request)
  }

  public listServers = async (): Promise<AxiosPromise<ServerResponse[]>> => {
    return HTTPClient.get("/servers")
  }

  public getStats = async (): Promise<AxiosPromise<StatsResponse>> => {
    return HTTPClient.get("/servers/stats")
  }
}

export const ManagerClient = new ServerManager()
