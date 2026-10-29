import lockoutArt from '../assets/template-lockout.webp'
import vanillaArt from '../assets/template-vanilla.webp'
import {useLoaderData} from "react-router-dom";
import {Template} from "./Template.tsx";
import {CatalogResponse} from "../clients/ManagerClient.ts";
import {Heading} from "../ui/Heading.tsx";
import {Panel} from "../ui/Panel.tsx";

const SERVER_TYPE = "FABRIC"

const presetArt: Record<string, string> = {lockout: lockoutArt}

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
                    art={<img className="size-full object-cover" src={presetArt[preset.id] ?? vanillaArt} alt=""/>}/>
        ))}
        <Template name="Vanilla" version={latest} secondary
                  art={<img className="size-full object-cover" src={vanillaArt} alt=""/>}/>
      </ul>
    </Panel>
  )
}
