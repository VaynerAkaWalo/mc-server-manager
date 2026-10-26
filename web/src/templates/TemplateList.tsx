import {useLoaderData} from "react-router-dom";
import {Template} from "./Template.tsx";
import {CatalogResponse} from "../clients/ManagerClient.ts";

export function TemplateList() {
  const presets = useLoaderData<CatalogResponse | null>()?.presets ?? []

  return (
    <ul className="flex flex-col w-full max-w-3xl gap-4 p-4">
      {presets.map(preset => (
        <Template key={preset.id} name={preset.name} description={preset.description} preset={preset.id} opts={preset.options}/>
      ))}
    </ul>
  )
}
