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

function Limit({value, base}: { value: number, base: number }) {
  return <span className={value > base ? "text-action-content" : ""}>{value}</span>
}

const DISTANCE_SCALE = 16

function Row({label, children}: { label: string, children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-2">
      <dt className="text-content-muted">{label}</dt>
      <dd>{children}</dd>
    </div>
  )
}

function Distance({label, value, base}: { label: string, value: number, base: number }) {
  const filled = Math.round(value / DISTANCE_SCALE * 8)
  return (
    <div className="grid grid-cols-[1fr_auto] gap-x-2 gap-y-2">
      <dt className="text-content-muted">{label}</dt>
      <dd><Limit value={value} base={base}/></dd>
      <dd aria-hidden="true" className="col-span-2 grid h-1.5 grid-cols-8 gap-0.5">
        {Array.from({length: 8}, (_, i) => <span key={i} className={i < filled ? "bg-action-content" : "bg-edge"}/>)}
      </dd>
    </div>
  )
}

const HOST_SUFFIX = ".mc-server.blamedevs.com"

/* Mirrors the backend: internal/definition/name_customizer.go (lowercase, spaces to dashes)
   and internal/server/utils.go serverAddress (name + ".mc-server.blamedevs.com"). */
function hostLabel(name: string) {
  return name.toLowerCase().replace(/ /g, "-")
}

function difficultyLabel(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1)
}

function DiceIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 20 20" className="size-5" fill="none" stroke="currentColor" strokeWidth="1.5">
      <rect x="3" y="3" width="14" height="14"/>
      <g fill="currentColor" stroke="none">
        <rect x="6" y="6" width="2" height="2"/><rect x="12" y="6" width="2" height="2"/>
        <rect x="9" y="9" width="2" height="2"/>
        <rect x="6" y="12" width="2" height="2"/><rect x="12" y="12" width="2" height="2"/>
      </g>
    </svg>
  )
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
        <Field label="Name" htmlFor="server-name-input">
          <div className="relative">
            <TextInput className="w-full pr-12" id="server-name-input" aria-describedby="server-host"
                       value={state.name} onChange={updateName}/>
            <button type="button" title="Random name" aria-label="Random name"
                    className="absolute top-1 right-1 grid size-8 place-items-center text-content-muted hover:bg-surface-hover hover:text-action-content active:translate-y-px focus-ring"
                    onClick={() => dispatch({type: "UPDATE_NAME", payload: randomServerName()})}>
              <DiceIcon/>
            </button>
          </div>
          <p id="server-host" className="mt-2 truncate font-mono text-xs text-content-muted" title={hostLabel(state.name) + HOST_SUFFIX}>
            <span className="text-action-content">{hostLabel(state.name) || "…"}</span>{HOST_SUFFIX}
          </p>
        </Field>

        <Field label="Version" htmlFor="server-version">
          <Select inputId="server-version" classNamePrefix="mc-select" options={versionOptions}
                  placeholder={"Select a version"} isDisabled={isTemplated()}
                  onChange={updateVersion} value={selectedVersion}/>
        </Field>
      </>
    )
  }

  const advancedSettings = () => {
    const updateMessageOfDay = (e: React.ChangeEvent<HTMLInputElement>) => {
      const rawMessage = e.target.value

      dispatch({type: "UPDATE_OPTS", payload: {MOTD: sanitizeString(rawMessage)}})
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
        <Field label="MOTD" htmlFor="message-of-the-day">
          <TextInput id="message-of-the-day"
                     value={state.opts.MOTD ?? ""} onChange={updateMessageOfDay}/>
        </Field>

        <Field label="Operators" htmlFor="operators">
          <TextInput id="operators" value={state.opts.OPS ?? ""} onChange={updateOperators}/>
        </Field>

        <Checkbox label="Online mode" checked={state.opts.ONLINE_MODE === 'true'} onChange={toggleOnlineMode}/>
      </>
    )
  }

  const difficultySettings = () => {
    const values = catalog?.options.find(option => option.id === "DIFFICULTY")?.values ?? []
    return (
      <fieldset className="flex flex-col">
        <legend className="mc-label">Difficulty</legend>
        <div className="grid grid-cols-4">
          {values.map(value => {
            const selected = state.opts.DIFFICULTY === value
            return (
              <label key={value}
                     className={`-ml-px flex h-10 cursor-pointer items-center justify-center border text-sm first:ml-0 has-focus-visible:z-10 has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-focus ${selected ? "z-10 border-action-content bg-surface-hover text-content" : "border-edge-control bg-inset text-content-muted hover:text-content"}`}>
                <input className="sr-only" type="radio" name="difficulty" value={value} checked={selected}
                       onChange={() => dispatch({type: "UPDATE_OPTS", payload: {DIFFICULTY: value}})}/>
                {difficultyLabel(value)}
              </label>
            )
          })}
        </div>
      </fieldset>
    )
  }

  const tierSettings = () => (
    <fieldset className="flex flex-col">
      <legend className="mc-label">Tier</legend>
      <div className="grid grid-cols-2 gap-2">
        {catalog?.tiers.map(tier => {
          const selected = state.tier === tier.id
          return (
            <label key={tier.id}
                   className={`flex h-12 cursor-pointer items-center gap-3 border px-4 has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-focus ${selected ? "border-action-content bg-surface-hover shadow-raised" : "border-edge-control bg-inset hover:border-content-muted"}`}>
              <input className="sr-only" type="radio" name="tier" value={tier.id} checked={selected}
                     onChange={() => dispatch({type: "UPDATE_TIER", payload: tier.id})}/>
              <span className={`font-pixel text-sm ${tier.available ? "" : "text-content-muted"}`}>{tier.name}</span>
              {!tier.available && <span className="ml-auto border border-edge-strong px-1 text-xs text-content-muted">Soon</span>}
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
    const perks = [tier.persistent && "World saved", tier.nameReserved && "Name reserved"].filter(Boolean) as string[]

    return (
      <aside className="flex flex-col gap-4 border-t border-edge bg-inset p-6 sm:col-start-2 sm:row-span-2 sm:row-start-1 sm:w-60 sm:border-t-0 sm:border-l" aria-labelledby="server-summary-heading">
        <Heading level={3} id="server-summary-heading">{tier.name}</Heading>
        {/* Three groups without headings: who plays, how the world runs, tier rules. */}
        <div className="flex flex-col text-sm">
          <dl className="flex flex-col gap-3 pb-4">
            <Row label="Players"><Limit value={tier.maxPlayers} base={baseline.maxPlayers}/></Row>
          </dl>
          <dl className="flex flex-col gap-3 border-t border-edge py-4">
            <Distance label="Render distance" value={tier.viewDistance} base={baseline.viewDistance}/>
            <Distance label="Simulation distance" value={tier.simulationDistance} base={baseline.simulationDistance}/>
            <Row label="Difficulty">{difficultyLabel(state.opts.DIFFICULTY ?? "")}</Row>
          </dl>
          <dl className="flex flex-col gap-3 border-t border-edge pt-4">
            <Row label="Idle shutdown">30 min</Row>
            {perks.map(perk => (
              <Row key={perk} label={perk}>
                <svg aria-hidden="true" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" className="size-4 text-action-content"><path d="m4 8 3 3 5-6"/></svg>
                <span className="sr-only">Yes</span>
              </Row>
            ))}
          </dl>
        </div>
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
      <div className="flex flex-col gap-4 p-6">
        <div className="flex items-center justify-between gap-4">
          <Heading level={2} className="grid">
            <span className={`[grid-area:1/1] ${showAdvanced ? "invisible" : ""}`} aria-hidden={showAdvanced}>New server</span>
            <span className={`[grid-area:1/1] ${showAdvanced ? "" : "invisible"}`} aria-hidden={!showAdvanced}>Advanced</span>
          </Heading>
          <Button variant="secondary" size="sm" className="grid"
                  aria-pressed={showAdvanced} aria-controls="basic-settings advanced-settings"
                  onClick={() => setShowAdvanced(advanced => !advanced)}>
            <span className={`[grid-area:1/1] ${showAdvanced ? "invisible" : ""}`}>Advanced</span>
            <span className={`[grid-area:1/1] ${showAdvanced ? "" : "invisible"}`}>← Basic</span>
          </Button>
        </div>

        {/* Both panes share one grid cell, so the modal keeps the height of the taller one when switching. */}
        <div className="grid flex-1">
          <div id="basic-settings" className={`flex flex-col gap-4 [grid-area:1/1] ${showAdvanced ? "invisible" : ""}`}
               inert={showAdvanced}>
            {basicSettings()}
            {tierSettings()}
          </div>
          <div id="advanced-settings" className={`flex flex-col gap-4 [grid-area:1/1] ${showAdvanced ? "" : "invisible"}`}
               inert={!showAdvanced}>
            {difficultySettings()}
            {advancedSettings()}
          </div>
        </div>
      </div>

      {summary()}

      <div className="flex items-center gap-2 w-full p-6 pt-0">
        <Button variant="secondary" className="flex-1 whitespace-nowrap" onClick={() => navigate("/my-servers")}>Cancel</Button>
        <Button className="flex-1 whitespace-nowrap" disabled={!selectedTier?.available} onClick={createServer}>
          {selectedTier?.available === false ? "Soon" : "Create"}
        </Button>
      </div>
    </div>
  )
}
