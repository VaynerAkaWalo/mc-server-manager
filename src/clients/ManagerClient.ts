import {HTTPClient} from "./Client.ts";
import {AxiosPromise} from "axios";

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

class ServerManager {
  public createServer = (request: CreateServerRequest): Promise<AxiosPromise<ServerResponse>> => {
    return HTTPClient.post("/servers", request)
  }

  public listServers = async (): Promise<AxiosPromise<ServerResponse[]>> => {
    return HTTPClient.get("/servers")
  }
}

export const ManagerClient = new ServerManager()
