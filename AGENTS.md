# Agent guidelines

## Pull requests that change the UI

Attach screenshots to the PR showing how the affected UI looked before and after the change.

## After the PR is ready

1. Wait for the PR to be merged.
2. Wait until the main pipeline of this repo has built the new container image.
3. Manually dispatch the workflow in `VaynerAkaWalo/mc-server-manager` to rebuild the manager image so it ships the latest UI (`gh workflow run "Docker Build and Push" -R VaynerAkaWalo/mc-server-manager`).
