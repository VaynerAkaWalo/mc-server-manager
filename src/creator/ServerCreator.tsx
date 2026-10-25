import {useLocation, useNavigate} from "react-router-dom";
import Select, {SingleValue} from "react-select";
import * as React from "react";
import {useEffect, useReducer, useState} from "react";
import {Checkbox, FormControlLabel, Tooltip} from "@mui/material";
import {sanitizeString} from "../utils/StringSanitizer.ts";
import {UpdateNotification} from "../App.tsx";
import {CreateServerRequest, ManagerClient, Opts, Tier} from "../clients/ManagerClient.ts";

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

const initialServer: Server = {
  name: "",
  opts: {
    ONLINE_MODE: 'true',
    MAX_PLAYERS: "20",
    DIFFICULTY: "normal"
  },
  duration: 2_700_000,
  tier: Tier.iron,
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

    function getVersion(): string {
      if (isTemplated()) {
        return templateName + "-" + templateOps.VERSION
      }

      return state.opts.VERSION || ""
    }

    const updateVersion = (e: React.ChangeEvent<HTMLInputElement>) => {
      if (isTemplated()) {
        setNotificationMessage("Version is locked for this template")
        return
      }
      const version = e.target.value
      dispatch({type: "UPDATE_OPTS", payload: {VERSION: version}})
    }

    const updateReservationTime = (selectedReservationTime: SingleValue<{ value: number }>) => {
      dispatch({type: "UPDATE_EXPIRE_TIME", payload: selectedReservationTime?.value || 0})
    }

    return (
      <>
        <div className="flex flex-col w-3/4">
          <label htmlFor="server-name-input">Server name</label>
          <input className="mc-input" id="server-name-input" type="text" placeholder="My Minecraft Server"
                 value={state.name} onChange={updateName}/>
        </div>

        <div className="flex flex-col w-3/4">
          <label htmlFor="server-version">Server version</label>
          <input className="mc-input" id="server-version" type="text" placeholder="Latest Version"
                 value={getVersion()} onChange={updateVersion}/>
        </div>

        <div className={"w-3/4"}>
          <label htmlFor={"server-reservation-time-select"}>Select reservation time</label>
          <Select id={"server-reservation-time-select"} className={"text-zinc-800"} options={reservationTimeOptions}
                  onChange={updateReservationTime}
                  value={reservationTimeOptions.find(option => option.value === state.duration)}/>
        </div>
      </>
    )
  }

  const advancedSettings = () => {
    const difficultOptions = [
      {
        label: "Peaceful",
        value: "peaceful"
      },
      {
        label: "Easy",
        value: "easy"
      },
      {
        label: "Normal",
        value: "normal"
      },
      {
        label: "Hard",
        value: "hard"
      },
    ]

    const updateMessageOfDay = (e: React.ChangeEvent<HTMLInputElement>) => {
      const rawMessage = e.target.value

      dispatch({type: "UPDATE_OPTS", payload: {MOTD: sanitizeString(rawMessage)}})
    }

    const updateDifficulty = (selectedDifficulty: SingleValue<{ value: string }>) => {
      dispatch({type: "UPDATE_OPTS", payload: {DIFFICULTY: selectedDifficulty?.value}})
    }

    return (
      <>
        <div className="flex flex-col w-3/4">
          <label htmlFor="message-of-the-day">Message of the day</label>
          <input className={"mc-input"} id={"message-of-the-day"} type={"text"} placeholder={"Welcome to the server!"}
                 value={state.opts.MOTD} onChange={updateMessageOfDay}/>
        </div>

        <div className={"w-3/4"}>
          <label htmlFor={"server-difficulty"}>Server difficulty</label>
          <Select id={"server-difficulty"} className={"text-zinc-800"} options={difficultOptions}
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
        <div className="flex flex-col w-3/4">
          <label htmlFor="player-limit">Player limit</label>
          <input className="mc-input text-right" id="player-limit" type="text" value={state.opts.MAX_PLAYERS}
                 onChange={updateMaxPlayers}/>
        </div>

        <div className="flex flex-col w-3/4">
          <label htmlFor="operators">Server operators</label>
          <input className="mc-input" id="operators" type="text" value={state.opts.OPS} onChange={updateOperators}/>
        </div>

        <div className={"flex flex-row items-center justify-evenly w-3/4"}>
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
        <div>
          <label className={"text-center"} htmlFor={"tier-settings"}>Server tier</label>
          <div id={"tier-settings"} className={"flex flex-row"}>
            <Tooltip title={"When chunks are already generated"}>
              <FormControlLabel
                control={<Checkbox value={Tier.wooden} checked={state.tier === Tier.wooden} onChange={updateTier}/>}
                label="Wooden"/>
            </Tooltip>
            <Tooltip title={"Default option, should handle up to 10 players"}>
              <FormControlLabel
                control={<Checkbox value={Tier.iron} checked={state.tier === Tier.iron} onChange={updateTier}/>}
                label="Iron"/>
            </Tooltip>
          </div>
        </div>
        <div>
          <label className={"text-center"} htmlFor={"optimization-settings"}>Server type</label>
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

    if (state.duration === 0) {
      setNotificationMessage("Reservation time must be selected")
      return
    }

    const sanitizedOpts = Object.fromEntries(Object.entries(state.opts).filter(([_, v]) => v !== ""));

    const request: CreateServerRequest = {
      name: state.name,
      duration: state.duration,
      opts: sanitizedOpts,
      tier: state.tier,
      preset: templatePreset,
    }

    await ManagerClient.createServer(request).catch(err => {
      setNotificationMessage("Failed to create server with unknown error " + err)
    }).then(() => {
      setNotificationMessage("Your server will be started shortly")
      navigate("/")
    })
  }

  return (
    <div className="mc-modal">
      <ul className="flex flex-row justify-evenly border-b-2 border-edge">
        <button className="mc-tab" value={CreatorTab.BASIC} onClick={() => setActiveTab(CreatorTab.BASIC)}>Basic
        </button>
        <button className="mc-tab" value={CreatorTab.TIER} onClick={() => setActiveTab(CreatorTab.TIER)}>Tier
        </button>
        <button className="mc-tab" value={CreatorTab.ADVANCED}
                onClick={() => setActiveTab(CreatorTab.ADVANCED)}>Advanced
        </button>
        <button className="mc-tab" value={CreatorTab.PLAYERS} onClick={() => setActiveTab(CreatorTab.PLAYERS)}>Players
        </button>
      </ul>
      <div className="flex flex-1 flex-col justify-evenly items-center">
        {activeTab === CreatorTab.BASIC && basicSettings()}
        {activeTab === CreatorTab.ADVANCED && advancedSettings()}
        {activeTab === CreatorTab.PLAYERS && playerSettings()}
        {activeTab === CreatorTab.TIER && tierSettings()}
      </div>
      <div className={"flex items-center justify-evenly gap-2 w-full px-4 pb-6"}>
        <button className={"mc-button flex-1 px-2 whitespace-nowrap"} onClick={() => navigate("/")}>Cancel</button>
        <button className={"mc-button flex-1 px-2 whitespace-nowrap"} onClick={createServer}>Create server</button>
      </div>
    </div>
  )
}
