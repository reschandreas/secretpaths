# secretpaths

Visualizes secrets and access paths managed by policies in HashiCorp Vault / OpenBao. Two components plus a Helm chart.

## Layout

- `server/` — Go (module `secretpaths`, Go 1.27) Gin API, listens `:8081`. Entrypoint `server/main.go`; Vault client in `server/backend/`; models/tests in `server/models/`.
- `visualizer/` — SvelteKit (Svelte 5, adapter-node) UI, dev server `:5173`. Routes under `src/routes/` (`/`, `/paths`, `/discover`).
- `helm/secretpaths/` — Helm chart; chart-testing config in root `ct.yaml`.

## Commands

Server:
- `cd server && go test ./...` — Go tests (models, backend, main).
- `cd server && go build -o /tmp/server .` — build. Docker build sets `CGO_ENABLED=0 GOOS=linux` and `rm -rf demo/` first.
- Run: `cd server && go run .` with env configured.

Visualizer:
- `cd visualizer && npm ci`
- `cd visualizer && npm run dev` — Vite dev server `:5173`. Vite proxies `/v1/*` to `:8081` (see `vite.config.ts`), so the backend must be running. (Legacy `dev-interceptor.ts` is now a no-op.)
- `cd visualizer && npm run check` — runs `svelte-kit sync` then `svelte-check`. Always sync first.
- `cd visualizer && npm run lint` — `prettier --check . && eslint .` (check only).
- `cd visualizer && npm test` — Vitest unit tests (`src/**/*.test.ts`).
- `cd visualizer && npm run test:e2e` — Playwright visual regression (`e2e/`); update baselines with `test:e2e:update`.
- `cd visualizer && npm run format` — `prettier --write .`
- `cd visualizer && npm run build` — production build.

Helm:
- `ct lint` / `ct install` against `helm/` (see `.github/workflows/lint-test-helm.yaml`).

## Environment (server)

Auth is auto-selected in `backend.AutoAuth`: `KUBERNETES_ROLE` (+ optional `KUBERNETES_PATH`) → `APPROLE_ROLE_ID`+`APPROLE_SECRET_ID` → `VAULT_TOKEN`. Also `VAULT_ADDR` (default `http://127.0.0.1:8200`), `VAULT_KV_ENGINE` (default `secret`). Vault needs the `secretpaths` policy (see `server/README.md`).

## Gotchas

- Lint, format, and `svelte-check` are all clean. Keep them green (`npm run lint`, `npm run check`).
- In the Docker build the visualizer empties `dev-interceptor.ts` and deletes `hooks.server.ts` (`RUN echo "" > ...`, `RUN rm ./src/hooks.server.ts`), so dev-only proxy code must stay in those files.
- Server caches responses in otter with a 5-minute TTL; `/update` (GET) refreshes caches. Scheduler also polls `localhost:8081/update` every 3 minutes.
- `/v1/paths` and `/v1/annotatedSecrets` accept `?q=`, `?limit=`, `?offset=` for server-side search/pagination; `/v1/level?l=N` and `/v1/graph` are cached by depth. Responses are gzip-compressed.
- Go module path is `secretpaths` (no domain prefix); import packages as `secretpaths/...`.
- Module `go 1.27.0` with toolchain `go1.27.4`.

## CI

- `.github/workflows/build-and-push.yaml` builds/pushes both Docker images on push to `main`/tags/PR.
- `.github/workflows/lint-test-helm.yaml` runs chart-testing on PRs touching `helm/`.
- Renovate auto-merges patch/minor updates.