# Agent guidelines

## API

Follow [docs/API_RULES.md](docs/API_RULES.md) when adding or changing endpoints.

## UI and design system

Follow [docs/DESIGN.md](docs/DESIGN.md) for every change in `web/src`. Build new UI from the components in `web/src/ui`; if one is missing, add it there instead of styling it on the page.

## Pull requests that change the UI

Attach screenshots to the PR showing how the affected UI looked before and after the change.

## What's new on the home page

Keep the "What's new" section in `web/src/home/Home.tsx` up to date with changes that matter to players, such as a new release or a new place to find the app. It is not a changelog: keep entries short, skip technical details and leave out changes users would not notice.
