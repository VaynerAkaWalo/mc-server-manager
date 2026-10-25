import wordUrl from '../assets/world.png'
import {useNavigate} from "react-router-dom";
import {Opts} from "../clients/ManagerClient.ts";

export interface template {
  name: string
  description: string
  preset: string
  opts: Opts
}

export function Template({name, description, preset, opts}: template) {
  const navigate = useNavigate()

  const handleCreateServerFromTemplate = () => {
    navigate("/server-creator", { state: { templateName: name, templatePreset: preset, templateOps: opts } } )
  }

  return (
    <li className="flex flex-col sm:flex-row items-center gap-4 w-full mc-panel p-4 sm:p-6">
      <div className="flex size-20 shrink-0 items-center justify-center">
        <img className="max-h-full" src={wordUrl} alt=""/>
      </div>

      <div className="flex flex-col flex-1 min-w-0 gap-1 text-center sm:text-left">
        <h2 className="font-pixel text-lg">{name}</h2>
        <p className="text-sm text-content-muted">{description}</p>
      </div>

      <div className="flex shrink-0 justify-center">
        <button className="mc-button" onClick={handleCreateServerFromTemplate}>Select</button>
      </div>
    </li>
  )
}
