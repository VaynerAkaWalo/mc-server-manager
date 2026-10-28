import {useLoaderData, useLocation, useNavigate} from "react-router-dom";
import Select, {SingleValue} from "react-select";
import * as React from "react";
import {useEffect, useReducer, useState} from "react";
import {Button} from "../ui/Button.tsx";
import {Checkbox, Field, TextInput} from "../ui/Field.tsx";
import {Heading} from "../ui/Heading.tsx";
import {sanitizeString} from "../utils/StringSanitizer.ts";
import {UpdateNotification} from "../App.tsx";
import {CatalogResponse, CreateServerRequest, ManagerClient, Opts, problemMessage, Tier} from "../clients/ManagerClient.ts";
import {randomServerName} from "./randomServerName.ts";

interface Server {
  name: string
  opts: Opts
  tier: Tier,
}

const SERVER_TYPE = "FABRIC"

const initialServer: Server = {
  name: "",
  opts: {
    ONLINE_MODE: 'true',
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

function Included({value}: { value: boolean }) {
  return value
    ? <span>Yes</span>
    : <span className="text-content-muted">No</span>
}

function Limit({value, base}: { value: number, base: number }) {
  return <span className={value > base ? "text-action-content" : ""}>{value}</span>
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
      ...(templateName && templateOps ? templateOps : {VERSION: catalog?.versions?.[SERVER_TYPE]?.latestRelease}),
    },
  }))
  const {setNotificationMessage} = UpdateNotification()
  const navigate = useNavigate()
  const [showAdvanced, setShowAdvanced] = useState(false)

  useEffect(() => {
    if (!catalog) {
      setNotificationMessage("Failed to load server options, try again later")
    }
  }, []);

  function isTemplated(): boolean {
    return templateName && templateOps
  }

  const basicSettings = () => {
    const updateName = (e: React.ChangeEvent<HTMLInputElement>) => {
      const rawName = e.target.value

      dispatch({type: "UPDATE_NAME", payload: sanitizeString(rawName)})
    }

    const serverVersions = catalog?.versions?.[SERVER_TYPE]
    const versionOptions = (serverVersions?.items ?? [])
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
        <Field label="Server name" htmlFor="server-name-input">
          <div className="flex flex-row gap-2">
            <TextInput className="flex-1 min-w-0" id="server-name-input" placeholder="My Minecraft Server"
                       value={state.name} onChange={updateName}/>
            <Button variant="secondary" size="icon" title="Generate another name"
                    aria-label="Generate another name"
                    onClick={() => dispatch({type: "UPDATE_NAME", payload: randomServerName()})}>
              ↻
            </Button>
          </div>
        </Field>

        <Field label="Server version" htmlFor="server-version">
          <Select inputId="server-version" classNamePrefix="mc-select" options={versionOptions}
                  placeholder={"Select a version"} isDisabled={isTemplated()}
                  onChange={updateVersion} value={selectedVersion}/>
        </Field>
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
        <Field label="Message of the day" htmlFor="message-of-the-day">
          <TextInput id="message-of-the-day" placeholder="Welcome to the server!"
                     value={state.opts.MOTD ?? ""} onChange={updateMessageOfDay}/>
        </Field>

        <Field label="Server difficulty" htmlFor="server-difficulty">
          <Select inputId="server-difficulty" classNamePrefix="mc-select" options={difficultOptions}
                  onChange={updateDifficulty}
                  value={difficultOptions.find(option => option.value === state.opts.DIFFICULTY)}/>
        </Field>

        <Field label="Server operators" htmlFor="operators">
          <TextInput id="operators" value={state.opts.OPS ?? ""} onChange={updateOperators}/>
        </Field>

        <Checkbox label="Online mode" checked={state.opts.ONLINE_MODE === 'true'} onChange={toggleOnlineMode}/>
      </>
    )
  }

  const tierSettings = () => (
    <fieldset className="flex flex-col">
      <legend className="mc-label">Server tier</legend>
      <div className="grid grid-cols-[repeat(auto-fit,minmax(8rem,1fr))] gap-2">
        {catalog?.tiers.map(tier => {
          const selected = state.tier === tier.id
          return (
            <label key={tier.id}
                   className={`flex h-10 cursor-pointer items-center gap-2 border bg-inset px-3 text-sm has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-focus ${selected ? "border-action-content" : "border-edge-control hover:border-content-muted"}`}>
              <input className="sr-only" type="radio" name="tier" value={tier.id} checked={selected}
                     onChange={() => dispatch({type: "UPDATE_TIER", payload: tier.id})}/>
              <span aria-hidden="true" className={`grid size-3 shrink-0 place-items-center border ${selected ? "border-action-content" : "border-edge-control"}`}>
                {selected && <span className="size-1 bg-action-content"/>}
              </span>
              <span className={tier.available ? "" : "text-content-muted"}>{tier.name}</span>
              {!tier.available && <span className="ml-auto border border-edge-strong px-1 text-xs text-content-muted">Coming soon</span>}
            </label>
          )
        })}
      </div>
    </fieldset>
  )

  const selectedTier = catalog?.tiers.find(candidate => candidate.id === state.tier)

  const summary = () => {
    const tier = selectedTier
    if (!tier) return null

    const baseline = catalog?.tiers.find(candidate => candidate.default) ?? tier

    const version = isTemplated() ? `${templateName}-${templateOps.VERSION}` : state.opts.VERSION
    const groups: [string, React.ReactNode][][] = [
      [["Version", version], ["Players", <Limit value={tier.maxPlayers} base={baseline.maxPlayers}/>]],
      [
        ["Render distance", <Limit value={tier.viewDistance} base={baseline.viewDistance}/>],
        ["Simulation distance", <Limit value={tier.simulationDistance} base={baseline.simulationDistance}/>],
      ],
      [["Idle shutdown", "30 min"], ["World saved", <Included value={tier.persistent}/>], ["Name reserved", <Included value={tier.nameReserved}/>]],
    ]

    return (
      <aside className="flex flex-col gap-3 border-t border-edge bg-inset p-6 sm:col-start-2 sm:row-span-2 sm:row-start-1 sm:w-60 sm:border-t-0 sm:border-l" aria-labelledby="server-summary-heading">
        <Heading level={3} id="server-summary-heading">{tier.name} tier</Heading>
        <dl className="flex flex-col text-sm">
          {groups.map(rows => (
            <div key={rows[0][0]} className="flex flex-col gap-2 border-t border-edge py-3 first:border-t-0 first:pt-0">
              {rows.map(([label, value]) => (
                <div key={label} className="flex justify-between gap-2">
                  <dt className="text-content-muted">{label}</dt>
                  <dd>{value}</dd>
                </div>
              ))}
            </div>
          ))}
        </dl>
      </aside>
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
    <div className="mc-modal sm:grid sm:grid-cols-[1fr_auto]">
      <div className="flex flex-col">
        <div className="flex flex-row items-center justify-between gap-4 px-6 pt-6">
          <Heading level={2} className="grid">
            <span className={`[grid-area:1/1] ${showAdvanced ? "invisible" : ""}`} aria-hidden={showAdvanced}>New server</span>
            <span className={`[grid-area:1/1] ${showAdvanced ? "" : "invisible"}`} aria-hidden={!showAdvanced}>Advanced</span>
          </Heading>
          <Button variant="secondary" size="sm" className="grid" onClick={() => setShowAdvanced(advanced => !advanced)}>
            <span className={`[grid-area:1/1] ${showAdvanced ? "invisible" : ""}`}>Advanced</span>
            <span className={`[grid-area:1/1] ${showAdvanced ? "" : "invisible"}`}>Back</span>
          </Button>
        </div>

        {/* Both views share one grid cell so the modal keeps the height of the taller one when switching. */}
        <div className="grid flex-1 p-6">
          <div className={`flex flex-col gap-4 [grid-area:1/1] ${showAdvanced ? "opacity-0" : ""}`}
               inert={showAdvanced} aria-hidden={showAdvanced}>
            {basicSettings()}
            {tierSettings()}
          </div>

          <div className={`flex flex-col gap-4 [grid-area:1/1] ${showAdvanced ? "" : "opacity-0"}`}
               inert={!showAdvanced} aria-hidden={!showAdvanced}>
            {advancedSettings()}
          </div>
        </div>
      </div>

      {summary()}

      <div className="flex items-center gap-2 w-full p-6 sm:pt-0">
        <Button variant="secondary" className="flex-1 whitespace-nowrap" onClick={() => navigate("/my-servers")}>Cancel</Button>
        <Button className="flex-1 whitespace-nowrap" disabled={!selectedTier?.available} onClick={createServer}>
          {selectedTier?.available === false ? "Coming soon" : "Create server"}
        </Button>
      </div>
    </div>
  )
}
