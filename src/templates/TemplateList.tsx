import * as React from "react";
import {useEffect} from "react";
import {Template, template} from "./Template.tsx";

export function TemplateList() {
  const [templates, setTemplates] = React.useState<template[]>([])

  useEffect(() => {
    setTemplates([
      {
        name: "Lockout",
        description: "Compete in bingo challenges to claim victory! This gameplay focuses on team-based objective completion.",
        preset: "lockout",
        opts: {
          VERSION: "1.21.3",
        }
      }
    ])
  }, []);

  return (
    <ul className="flex flex-col min-w-1/4 max-w-5/12 h-2/3 justify-evenly">
      {templates.map((template) => (
        <Template name={template.name} description={template.description} preset={template.preset} opts={template.opts}/>
      ))}
    </ul>
  )
}
