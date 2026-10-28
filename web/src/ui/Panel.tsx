import {HTMLAttributes} from "react";

type Props = HTMLAttributes<HTMLElement> & { as?: "div" | "section" | "li" | "aside" }

export function Panel({as: Tag = "div", className, ...props}: Props) {
  return <Tag className={["mc-panel", className].filter(Boolean).join(" ")} {...props}/>
}
