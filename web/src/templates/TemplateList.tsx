import * as React from "react";
import {useEffect} from "react";
import {Template, template} from "./Template.tsx";
import {ManagerClient} from "../clients/ManagerClient.ts";

export function TemplateList() {
  const [templates, setTemplates] = React.useState<template[]>([])

  useEffect(() => {
    ManagerClient.getCatalog().then(({data}) => {
      setTemplates(data.presets.map(preset => ({
        name: preset.name,
        description: preset.description,
        preset: preset.id,
        opts: preset.options,
      })))
    }).catch(() => setTemplates([]))
  }, []);

  return (
    <ul className="flex flex-col w-full max-w-3xl gap-4 p-4">
      {templates.map((template) => (
        <Template key={template.preset} name={template.name} description={template.description} preset={template.preset} opts={template.opts}/>
      ))}
    </ul>
  )
}
