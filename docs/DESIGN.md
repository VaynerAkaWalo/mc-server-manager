# Design rules for the web app

Read this before changing anything in `web/src`. If a rule blocks the task, stop and ask instead of working around it.

## Source of truth

- All colors, the type scale, radii and the content width live in `@theme` in `web/src/App.css`. Tailwind's default palette, type scale and radii are switched off there, so only these tokens exist as utilities.
- Adding a token is its own, deliberate change in `@theme`. Never hardcode a value on a page instead.

## Never

- Raw colors (`#hex`, `rgb()`) in TSX, or in CSS outside `@theme`.
- Arbitrary Tailwind values (`w-[...]`, `p-[...]`, `text-[13px]`). The only exception is a grid template (`grid-cols-[...]`) or `[grid-area:...]`.
- Importing `@mui/*` outside `web/src/ui/mui.ts`.
- Gradients, shadows, rounded corners, glow, emoji in the UI, or decorative illustrations.
- Status colors for other meanings. `online`, `starting` and `warning` describe a server's status (and `warning` a real warning). "Not included" is plain `text-content-muted`, a better-than-default value is `text-action-content`.

## Components (`web/src/ui`)

- `Button` / `LinkButton`: `variant="primary" | "secondary" | "ghost"`, `size="md" | "sm" | "icon"`. Do not write `<button className="...">` with its own padding and border.
- `Field` (label + control), `TextInput`, `Checkbox`. Selects use `react-select` with `classNamePrefix="mc-select"` inside a `Field`.
- `Panel` for every bordered surface, `Heading level={1 | 2 | 3}` for every heading.
- Missing something? Add it to `web/src/ui` first, then use it on the page.

## Typography

- `font-pixel` only for headings, buttons, navigation and big numbers. Everything else is `font-sans`.
- One size per role: h1 `text-2xl sm:text-3xl`, h2 `text-2xl`, h3 `text-base` (all via `Heading`), body and fields `text-base`, labels and secondary text `text-sm`, dates and meta `text-xs`.

## Spacing

- Use the steps 1, 2, 3, 4, 6 and 8 (4 to 32 px). Fields in a form are `gap-4` apart, panels use `p-6` (`p-4` on small screens).
- Controls are `h-10` (inputs, selects, choice cards).

## States

- Every interactive element has a hover state, the shared focus ring (`focus-ring` utility: 2 px outline in `--color-focus`, offset 2 px) and, where it can be disabled, `disabled:opacity-60 disabled:cursor-not-allowed`. The primitives already do this.
- Borders of controls use `border-edge-control` (at least 3:1 against `surface` and `inset`); `edge` is only for decorative dividers.

## Before opening a PR

- `npm run lint` and `npm run build` pass. The lint enforces the "Never" list above.
- Screenshots before and after are attached (see `AGENTS.md`).
