import {Link} from "react-router-dom";

export function Wordmark() {
  return (
    <Link to="/" className="flex h-10 items-center gap-3 px-1 focus-ring">
      <img src="/grass-block.svg" alt="" className="size-6 shrink-0 sm:size-7"/>
      <span className="whitespace-nowrap font-pixel text-xs text-content sm:text-base">Minecraft server hosting</span>
    </Link>
  )
}
