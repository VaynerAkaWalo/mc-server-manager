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
    <div className="flex flex-col sm:flex-row bg-stone-800 items-center gap-4 w-full mc-border p-4 sm:p-8">
      <div className="flex h-24 sm:h-32 shrink-0 items-center justify-center">
        <img className="max-h-full" src={wordUrl} alt="placeholder"/>
      </div>

      <div className="flex flex-col flex-1 min-w-0 text-zinc-300">
          <span className="font-bold text-xl mb-4">{name}</span>
          <span>{description}</span>
      </div>

      <div className="flex shrink-0 justify-center text-zinc-300">
        <button className="bg-zinc-500 hover:bg-zinc-600 px-8 py-2" onClick={handleCreateServerFromTemplate}>Select</button>
      </div>
    </div>
  )
}
