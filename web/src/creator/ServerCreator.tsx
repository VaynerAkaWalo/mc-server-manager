import {useLocation, useNavigate} from "react-router-dom";
import Select, {SingleValue} from "react-select";
import * as React from "react";
import {useEffect, useReducer, useState} from "react";
import {Checkbox, FormControlLabel, Tooltip} from "@mui/material";
import {sanitizeString} from "../utils/StringSanitizer.ts";
import {UpdateNotification} from "../App.tsx";
import {CatalogResponse, CreateServerRequest, ManagerClient, Opts, problemMessage, Tier} from "../clients/ManagerClient.ts";

enum CreatorTab {
  BASIC,
  ADVANCED,
  PLAYERS,
  TIER,
}

interface Server {
  name: string
  opts: Opts
  duration: number
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
  duration: 2_700_000,
  tier: "",
}

type ServerUpdateAction =
  | { type: 'UPDATE_NAME'; payload: string }
  | { type: 'UPDATE_EXPIRE_TIME'; payload: number }
  | { type: 'UPDATE_OPTS'; payload: Partial<Opts> }
  | { type: 'UPDATE_TIER'; payload: Tier }

function serverReducer(state: Server, action: ServerUpdateAction): Server {
  switch (action.type) {
    case "UPDATE_NAME":
      return {...state, name: action.payload}
    case "UPDATE_EXPIRE_TIME":
      return {...state, duration: action.payload}
    case "UPDATE_OPTS":
      return {...state, opts: {...state.opts, ...action.payload}}
    case "UPDATE_TIER":
      return {...state, tier: action.payload}
  }
}

export function ServerCreator() {
  const [activeTab, setActiveTab] = useState<CreatorTab>(CreatorTab.BASIC)
  const [state, dispatch] = useReducer(serverReducer, initialServer)
  const {setNotificationMessage} = UpdateNotification()
  const {templateName, templatePreset, templateOps} = useLocation().state || {}
  const navigate = useNavigate()
  const [catalog, setCatalog] = useState<CatalogResponse | null>(null)
  const [showSnapshots, setShowSnapshots] = useState(false)

  useEffect(() => {
    ManagerClient.getCatalog().then(({data}) => {
      setCatalog(data)
      dispatch({type: "UPDATE_TIER", payload: (data.tiers.find(tier => tier.default) ?? data.tiers[0])?.id ?? ""})
      if (!isTemplated()) {
        dispatch({type: "UPDATE_OPTS", payload: {VERSION: data.versions?.[DEFAULT_SERVER_TYPE]?.latestRelease}})
      }
    }).catch(err => setNotificationMessage("Failed to load server options: " + problemMessage(err)))
  }, []);

  function isTemplated(): boolean {
    return templateName && templateOps
  }

  useEffect(() => {
    if (isTemplated()) {
      dispatch({type: "UPDATE_OPTS", payload: templateOps})
    }
  }, []);

  const basicSettings = () => {
    const reservationTimeOptions = [
      {
        value: 1_200_000,
        label: "20min"
      },
      {
        value: 2_700_000,
        label: "45min"
      },
      {
        value: 14_400_000,
        label: "4h"
      },
      {
        value: 43_200_000,
        label: "12h"
      }
    ]

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

    const updateShowSnapshots = () => {
      const isSnapshot = serverVersions?.items.find(version => version.id === state.opts.VERSION)?.kind === "snapshot"
      if (showSnapshots && isSnapshot) {
        dispatch({type: "UPDATE_OPTS", payload: {VERSION: serverVersions?.latestRelease}})
      }
      setShowSnapshots(!showSnapshots)
    }

    const updateReservationTime = (selectedReservationTime: SingleValue<{ value: number }>) => {
      dispatch({type: "UPDATE_EXPIRE_TIME", payload: selectedReservationTime?.value || 0})
    }

    return (
      <>
        <div className="flex flex-col">
          <label className="mc-label" htmlFor="server-name-input">Server name</label>
          <input className="mc-input" id="server-name-input" type="text" placeholder="My Minecraft Server"
                 value={state.name} onChange={updateName}/>
        </div>

        <div className="flex flex-col">
          <div className="flex flex-row items-center justify-between">
            <label className="mc-label mb-0" htmlFor="server-version">Server version</label>
            <FormControlLabel sx={{margin: 0, gap: 0.5}}
                              control={<Checkbox size="small" sx={{padding: 0}} checked={showSnapshots}
                                                 disabled={isTemplated()} onChange={updateShowSnapshots}/>}
                              label="Snapshots"/>
          </div>
          <Select inputId="server-version" classNamePrefix="mc-select" options={versionOptions}
                  placeholder={"Select a version"} isDisabled={isTemplated()}
                  onChange={updateVersion} value={selectedVersion}/>
        </div>

        <div className="flex flex-col">
          <label className="mc-label" htmlFor={"server-reservation-time-select"}>Reservation time</label>
          <Select inputId={"server-reservation-time-select"} classNamePrefix="mc-select" options={reservationTimeOptions}
                  onChange={updateReservationTime}
                  value={reservationTimeOptions.find(option => option.value === state.duration)}/>
        </div>
      </>
    )
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

    return (
      <>
        <div className="flex flex-col">
          <label className="mc-label" htmlFor="message-of-the-day">Message of the day</label>
          <input className={"mc-input"} id={"message-of-the-day"} type={"text"} placeholder={"Welcome to the server!"}
                 value={state.opts.MOTD} onChange={updateMessageOfDay}/>
        </div>

        <div className="flex flex-col">
          <label className="mc-label" htmlFor={"server-difficulty"}>Server difficulty</label>
          <Select inputId={"server-difficulty"} classNamePrefix="mc-select" options={difficultOptions}
                  onChange={updateDifficulty}
                  value={difficultOptions.find(option => option.value === state.opts.DIFFICULTY)}/>
        </div>
      </>
    )
  }

  const playerSettings = () => {
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
          <label className="mc-label" htmlFor="player-limit">Player limit</label>
          <input className="mc-input text-right" id="player-limit" type="text" value={state.opts.MAX_PLAYERS}
                 onChange={updateMaxPlayers}/>
        </div>

        <div className="flex flex-col">
          <label className="mc-label" htmlFor="operators">Server operators</label>
          <input className="mc-input" id="operators" type="text" value={state.opts.OPS} onChange={updateOperators}/>
        </div>

        <div className="flex flex-row items-center">
          <FormControlLabel control={<Checkbox checked={state.opts.ONLINE_MODE === 'true'} onClick={toggleOnlineMode}/>}
                            label="Online mode"/>
        </div>
      </>
    )
  }

  const tierSettings = () => {

    const updateTier = (e: React.ChangeEvent<HTMLInputElement>) => {
      const newValue = e.target.value as Tier
      dispatch({type: "UPDATE_TIER", payload: newValue})
    }

    return (
      <>
        <div className="flex flex-col">
          <label className="mc-label" htmlFor={"tier-settings"}>Server tier</label>
          <div id={"tier-settings"} className={"flex flex-row"}>
            {catalog?.tiers.map(tier => (
              <Tooltip key={tier.id} title={tier.description}>
                <FormControlLabel
                  control={<Checkbox value={tier.id} checked={state.tier === tier.id} onChange={updateTier}/>}
                  label={tier.name}/>
              </Tooltip>
            ))}
          </div>
        </div>
        <div className="flex flex-col">
          <label className="mc-label" htmlFor={"optimization-settings"}>Server type</label>
          <div id={"optimization-settings"} className={"flex flex-row"}>
            <Tooltip title={"Default option, uses vanilla server"}>
              <FormControlLabel control={<Checkbox checked={true} value={"vanilla"}/>} label="Vanilla"/>
            </Tooltip>
            <Tooltip title={"Highly optimized fabric server, gameplay should not be affected"}>
              <FormControlLabel control={<Checkbox checked={false} value={"fabric"}/>} label="Fabric"/>
            </Tooltip>
          </div>
        </div>
      </>
    )
  }

  const createServer = async () => {
    if (state.name === "" || state.name.length < 3 || state.name.length > 25) {
      setNotificationMessage("Server name must have between 3 and 25 characters")
      return
    }

    if (!isTemplated() && !state.opts.VERSION) {
      setNotificationMessage("Server version must be selected")
      return
    }

    if (state.duration === 0) {
      setNotificationMessage("Reservation time must be selected")
      return
    }

    const sanitizedOpts = Object.fromEntries(Object.entries(state.opts).filter(([, value]) => value !== ""));

    const request: CreateServerRequest = {
      name: state.name,
      durationMs: state.duration,
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

  const tabs = [
    {tab: CreatorTab.BASIC, label: "Basic"},
    {tab: CreatorTab.TIER, label: "Tier"},
    {tab: CreatorTab.ADVANCED, label: "Advanced"},
    {tab: CreatorTab.PLAYERS, label: "Players"},
  ]

  const selectTabWithArrows = (e: React.KeyboardEvent<HTMLDivElement>) => {
    const step = {ArrowRight: 1, ArrowLeft: -1}[e.key]
    if (!step) return

    const index = tabs.findIndex(({tab}) => tab === activeTab)
    const next = tabs[(index + step + tabs.length) % tabs.length].tab
    setActiveTab(next)
    document.getElementById(`creator-tab-${next}`)?.focus()
  }

  return (
    <div className="mc-modal">
      <div className="flex flex-row flex-wrap justify-evenly border-b border-edge" role="tablist"
           onKeyDown={selectTabWithArrows}>
        {tabs.map(({tab, label}) => (
          <button key={tab} id={`creator-tab-${tab}`} className={`mc-tab ${activeTab === tab ? "active" : ""}`}
                  role="tab" aria-selected={activeTab === tab} aria-controls="creator-tab-panel"
                  tabIndex={activeTab === tab ? 0 : -1} onClick={() => setActiveTab(tab)}>
            {label}
          </button>
        ))}
      </div>

      <div className="flex flex-1 flex-col gap-5 p-6" id="creator-tab-panel" role="tabpanel"
           aria-labelledby={`creator-tab-${activeTab}`}>
        {activeTab === CreatorTab.BASIC && basicSettings()}
        {activeTab === CreatorTab.ADVANCED && advancedSettings()}
        {activeTab === CreatorTab.PLAYERS && playerSettings()}
        {activeTab === CreatorTab.TIER && tierSettings()}
      </div>

      <div className="flex items-center gap-2 w-full px-6 pb-6">
        <button className="mc-button-secondary flex-1 whitespace-nowrap" onClick={() => navigate("/my-servers")}>Cancel</button>
        <button className="mc-button flex-1 whitespace-nowrap" onClick={createServer}>Create server</button>
      </div>
    </div>
  )
}
