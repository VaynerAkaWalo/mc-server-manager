import {useLoaderData, useLocation, useNavigate} from "react-router-dom";
import Select, {SingleValue} from "react-select";
import * as React from "react";
import {useEffect, useReducer, useState} from "react";
import {Checkbox, FormControlLabel} from "@mui/material";
import {sanitizeString} from "../utils/StringSanitizer.ts";
import {UpdateNotification} from "../App.tsx";
import {CatalogResponse, CreateServerRequest, ManagerClient, Opts, problemMessage, Tier} from "../clients/ManagerClient.ts";
import {randomServerName} from "./randomServerName.ts";
import {Spinner} from "../base/Spinner.tsx";

interface Server {
  name: string
  opts: Opts
  tier: Tier,
}

const DEFAULT_SERVER_TYPE = "VANILLA"

const initialServer: Server = {
  name: "",
  opts: {
    ONLINE_MODE: 'true',
    MAX_PLAYERS: "20",
    DIFFICULTY: "normal"
  },
  tier: "",
}

type ServerUpdateAction =
  | { type: 'UPDATE_NAME'; payload: string }
  | { type: 'UPDATE_OPTS'; payload: Partial<Opts> }
  | { type: 'UPDATE_TIER'; payload: Tier }

function serverReducer(state: Server, action: ServerUpdateAction): Server {
  switch (action.type) {
    case "UPDATE_NAME":
      return {...state, name: action.payload}
    case "UPDATE_OPTS":
      return {...state, opts: {...state.opts, ...action.payload}}
    case "UPDATE_TIER":
      return {...state, tier: action.payload}
  }
}

export function ServerCreator() {
  const catalog = useLoaderData<CatalogResponse | null>()
  const {templateName, templatePreset, templateOps} = useLocation().state || {}
  const [state, dispatch] = useReducer(serverReducer, undefined, (): Server => ({
    ...initialServer,
    name: randomServerName(),
    tier: (catalog?.tiers.find(tier => tier.default) ?? catalog?.tiers[0])?.id ?? "",
    opts: {
      ...initialServer.opts,
      ...(templateName && templateOps ? templateOps : {VERSION: catalog?.versions?.[DEFAULT_SERVER_TYPE]?.latestRelease}),
    },
  }))
  const {setNotificationMessage} = UpdateNotification()
  const navigate = useNavigate()
  const [showSnapshots, setShowSnapshots] = useState(false)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [isSwitching, setIsSwitching] = useState(false)

  useEffect(() => {
    if (!catalog) {
      setNotificationMessage("Failed to load server options, try again later")
    }
  }, []);

  function isTemplated(): boolean {
    return templateName && templateOps
  }

  useEffect(() => {
    if (!isSwitching) {
      return
    }

    const timer = setTimeout(() => setIsSwitching(false), 500)
    return () => clearTimeout(timer)
  }, [isSwitching, showAdvanced]);

  const toggleAdvanced = () => {
    setShowAdvanced(!showAdvanced)
    setIsSwitching(true)
  }

  const basicSettings = () => {
    const updateName = (e: React.ChangeEvent<HTMLInputElement>) => {
      const rawName = e.target.value

      dispatch({type: "UPDATE_NAME", payload: sanitizeString(rawName)})
    }

    const serverVersions = catalog?.versions?.[state.opts.TYPE ?? DEFAULT_SERVER_TYPE]
    const versionOptions = (serverVersions?.items ?? [])
      .filter(version => showSnapshots || version.kind === "release")
      .map(version => ({
        value: version.id,
        label: version.id === serverVersions?.latestRelease ? `${version.id} (latest)` : version.id
      }))

    const selectedVersion = isTemplated()
      ? {value: templateOps.VERSION, label: templateName + "-" + templateOps.VERSION}
      : versionOptions.find(option => option.value === state.opts.VERSION) ?? null

    const updateVersion = (selected: SingleValue<{ value: string }>) => {
      dispatch({type: "UPDATE_OPTS", payload: {VERSION: selected?.value}})
    }

    return (
      <>
        <div className="flex flex-col">
          <label className="mc-label" htmlFor="server-name-input">Server name</label>
          <div className="flex flex-row gap-2">
            <input className="mc-input flex-1 min-w-0" id="server-name-input" type="text" placeholder="My Minecraft Server"
                   value={state.name} onChange={updateName}/>
            <button className="mc-button-secondary px-3" type="button" title="Generate another name"
                    aria-label="Generate another name"
                    onClick={() => dispatch({type: "UPDATE_NAME", payload: randomServerName()})}>
              ↻
            </button>
          </div>
        </div>

        <div className="flex flex-col">
          <label className="mc-label" htmlFor="server-version">Server version</label>
          <Select inputId="server-version" classNamePrefix="mc-select" options={versionOptions}
                  placeholder={"Select a version"} isDisabled={isTemplated()}
                  onChange={updateVersion} value={selectedVersion}/>
        </div>
      </>
    )
  }

  const toggleSnapshots = () => {
    const serverVersions = catalog?.versions?.[state.opts.TYPE ?? DEFAULT_SERVER_TYPE]
    const isSnapshot = serverVersions?.items.find(version => version.id === state.opts.VERSION)?.kind === "snapshot"
    if (showSnapshots && isSnapshot) {
      dispatch({type: "UPDATE_OPTS", payload: {VERSION: serverVersions?.latestRelease}})
    }

    setShowSnapshots(!showSnapshots)
  }

  const advancedSettings = () => {
    const difficultOptions = (catalog?.options.find(option => option.id === "DIFFICULTY")?.values ?? [])
      .map(value => ({label: value.charAt(0).toUpperCase() + value.slice(1), value}))

    const updateMessageOfDay = (e: React.ChangeEvent<HTMLInputElement>) => {
      const rawMessage = e.target.value

      dispatch({type: "UPDATE_OPTS", payload: {MOTD: sanitizeString(rawMessage)}})
    }

    const updateDifficulty = (selectedDifficulty: SingleValue<{ value: string }>) => {
      dispatch({type: "UPDATE_OPTS", payload: {DIFFICULTY: selectedDifficulty?.value}})
    }

    const updateMaxPlayers = (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value
      if (!isNaN(Number(value))) {
        dispatch({type: "UPDATE_OPTS", payload: {MAX_PLAYERS: value}})
      }
    }

    const updateOperators = (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value
      dispatch({type: "UPDATE_OPTS", payload: {OPS: value}})
    }

    const toggleOnlineMode = () => {
      const newValue = !(state.opts.ONLINE_MODE === 'true')
      dispatch({type: "UPDATE_OPTS", payload: {ONLINE_MODE: newValue.toString()}})
    }

    return (
      <>
        <div className="flex flex-col">
          <label className="mc-label" htmlFor="message-of-the-day">Message of the day</label>
          <input className={"mc-input"} id={"message-of-the-day"} type={"text"} placeholder={"Welcome to the server!"}
                 value={state.opts.MOTD} onChange={updateMessageOfDay}/>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="flex flex-col">
            <label className="mc-label" htmlFor={"server-difficulty"}>Server difficulty</label>
            <Select inputId={"server-difficulty"} classNamePrefix="mc-select" options={difficultOptions}
                    onChange={updateDifficulty}
                    value={difficultOptions.find(option => option.value === state.opts.DIFFICULTY)}/>
          </div>

          <div className="flex flex-col">
            <label className="mc-label" htmlFor="player-limit">Player limit</label>
            <input className="mc-input text-right" id="player-limit" type="text" value={state.opts.MAX_PLAYERS}
                   onChange={updateMaxPlayers}/>
          </div>
        </div>

        <div className="flex flex-col">
          <label className="mc-label" htmlFor="operators">Server operators</label>
          <input className="mc-input" id="operators" type="text" value={state.opts.OPS} onChange={updateOperators}/>
        </div>

        <div className="grid sm:grid-cols-2">
          <FormControlLabel sx={{margin: 0}}
                            control={<Checkbox checked={state.opts.ONLINE_MODE === 'true'} onClick={toggleOnlineMode}/>}
                            label="Online mode"/>
          <FormControlLabel sx={{margin: 0}}
                            control={<Checkbox checked={showSnapshots} disabled={isTemplated()} onChange={toggleSnapshots}/>}
                            label="Show snapshot versions"/>
        </div>
      </>
    )
  }

  const tierSettings = () => (
    <fieldset className="flex flex-col">
      <legend className="mc-label">Server tier</legend>
      <div className="grid grid-cols-[repeat(auto-fit,minmax(8rem,1fr))] gap-2">
        {catalog?.tiers.map(tier => (
          <label key={tier.id}
                 className={`cursor-pointer border bg-inset px-3 py-2 has-focus-visible:outline-2 has-focus-visible:outline-action-content ${state.tier === tier.id ? "border-action-content" : "border-edge hover:border-edge-strong"}`}>
            <input className="sr-only" type="radio" name="tier" value={tier.id} checked={state.tier === tier.id}
                   onChange={() => dispatch({type: "UPDATE_TIER", payload: tier.id})}/>
            <span className="text-sm">{tier.name}</span>
          </label>
        ))}
      </div>
    </fieldset>
  )

  const createServer = async () => {
    if (state.name === "" || state.name.length < 3 || state.name.length > 25) {
      setNotificationMessage("Server name must have between 3 and 25 characters")
      return
    }

    if (!isTemplated() && !state.opts.VERSION) {
      setNotificationMessage("Server version must be selected")
      return
    }

    const sanitizedOpts = Object.fromEntries(Object.entries(state.opts).filter(([, value]) => value !== ""));

    const request: CreateServerRequest = {
      name: state.name,
      options: sanitizedOpts,
      tier: state.tier,
      preset: templatePreset,
    }

    try {
      await ManagerClient.createServer(request)
    } catch (err) {
      setNotificationMessage("Failed to create server: " + problemMessage(err))
      return
    }

    setNotificationMessage("Your server will be started shortly")
    navigate("/my-servers")
  }

  return (
    <div className="mc-modal">
      <div className="flex flex-row items-center justify-between gap-4 px-6 pt-6">
        <h2 className="font-pixel text-2xl">{showAdvanced ? "Advanced" : "New server"}</h2>
        <button className="mc-button-secondary grid py-1 text-sm" onClick={toggleAdvanced}>
          <span className={`[grid-area:1/1] ${showAdvanced ? "invisible" : ""}`}>Advanced</span>
          <span className={`[grid-area:1/1] ${showAdvanced ? "" : "invisible"}`}>Back</span>
        </button>
      </div>

      {/* Both views share one grid cell so the modal keeps the height of the taller one when switching. */}
      <div className="grid flex-1 p-6">
        <div className={`flex flex-col gap-5 [grid-area:1/1] ${showAdvanced || isSwitching ? "invisible" : ""}`}
             inert={showAdvanced || isSwitching}>
          {basicSettings()}
          {tierSettings()}
          <p className="text-sm text-content-muted">The server shuts down after 30 minutes without players.</p>
        </div>

        <div className={`flex flex-col gap-4 [grid-area:1/1] ${showAdvanced && !isSwitching ? "" : "invisible"}`}
             inert={!showAdvanced || isSwitching}>
          {advancedSettings()}
        </div>

        {isSwitching && <div className="flex items-center justify-center [grid-area:1/1]"><Spinner/></div>}
      </div>

      <div className="flex items-center gap-2 w-full px-6 pb-6">
        <button className="mc-button-secondary flex-1 whitespace-nowrap" onClick={() => navigate("/my-servers")}>Cancel</button>
        <button className="mc-button flex-1 whitespace-nowrap" onClick={createServer}>Create server</button>
      </div>
    </div>
  )
}
