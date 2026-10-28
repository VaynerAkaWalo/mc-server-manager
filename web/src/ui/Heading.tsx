import {HTMLAttributes} from "react";

type Level = 1 | 2 | 3

const sizes: Record<Level, string> = {
  1: "text-2xl sm:text-3xl",
  2: "text-2xl",
  3: "text-base",
}

type Props = HTMLAttributes<HTMLHeadingElement> & { level: Level }

export function Heading({level, className, ...props}: Props) {
  const Tag = `h${level}` as const
  return <Tag className={["font-pixel", sizes[level], className].filter(Boolean).join(" ")} {...props}/>
}
