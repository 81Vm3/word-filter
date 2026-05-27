# Wails UI

This directory contains the desktop UI entrypoint and the Vue 3 frontend.

## Stack

- Wails v2 backend bridge (`cmd/wails`)
- Vue 3 + TypeScript frontend (`cmd/wails/frontend`)
- Vue components use **Option API only**

## Run (dev)

1. Install Wails CLI and platform dependencies.
2. Install frontend deps:
   ```bash
   cd cmd/wails/frontend
   pnpm install
   ```
3. From repo root run:
   ```bash
   cd cmd/wails
   wails dev -tags wails,desktop,dev
   ```

## Build

```bash
cd cmd/wails/frontend
pnpm build
cd /home/p3g4s/system/cmd/wails
wails build -tags wails,desktop,production
```

The UI binds to methods on `cmd/wails/app.go`, which forwards calls to `pkg/gui.Service`.
