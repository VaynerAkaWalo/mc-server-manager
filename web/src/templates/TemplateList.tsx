import worldUrl from '../assets/world.png'
import {useLoaderData} from "react-router-dom";
import {Template} from "./Template.tsx";
import {CatalogResponse} from "../clients/ManagerClient.ts";
import {Heading} from "../ui/Heading.tsx";
import {Panel} from "../ui/Panel.tsx";

const SERVER_TYPE = "FABRIC"

export function TemplateList() {
  const catalog = useLoaderData<CatalogResponse | null>()
  const presets = catalog?.presets ?? []
  const latest = catalog?.versions?.[SERVER_TYPE]?.latestRelease

  return (
    <Panel as="section" className="mx-4 my-6 flex w-full max-w-content flex-col gap-6 p-4 sm:p-8" aria-labelledby="templates-heading">
      <Heading level={1} id="templates-heading">Templates</Heading>
      <ul className="grid gap-4 sm:grid-cols-2">
        {presets.map(preset => (
          <Template key={preset.id} name={preset.name} description={preset.description} preset={preset.id}
                    opts={preset.options} version={preset.options.VERSION}
                    art={<img className="size-full object-cover [image-rendering:pixelated]" src={worldUrl} alt=""/>}/>
        ))}
        <Template name="Vanilla" version={latest} secondary
                  art={<div className="grid size-full place-items-center bg-chrome">
                    <img className="size-16 [image-rendering:pixelated]" src="/grass-block.svg" alt=""/>
                  </div>}/>
      </ul>
    </Panel>
  )
}
