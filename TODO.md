# TODO

## Gateway endpoints (stubbed, return 501)

Handlers are placeholders. `api/gateway_stubs_test.go` pins each one to `501 not_implemented`. Replace each stub test with real behavior tests when its handler is implemented.

- [ ] Catalog: `GET /catalog/skills`, `GET /catalog/skills/{skill}/agents`, `GET /catalog/agents/{agentName}/card` (`api/catalog.go`)
- [ ] Discover: `GET /discover`, `POST /discover/resolve` (`api/discover.go`)
- [ ] A2A: `POST /a2a/agents/{agentName}/message:send`, `POST /a2a/skills/{skill}/message:send`, `GET /a2a/tasks/{taskId}` (`api/a2a.go`)
- [ ] Watch: `GET /watch/catalog` (`api/watch.go`)

## Permissions for gateway routes

`core/server.go` lines 153–162 still carry `// TODO: perm for ...` comments on these routes. Pick the right permission for discovery, catalog, A2A and watch, then update the middleware config.

## Tests

- [ ] `ReadyAction` (`GET /api/v1/public/_ready`, `api/ready.go`) has no test. It is the only API handler without one.

## Agent instance health checks

`http` and `tcp` checks are stored and scheduled, but nothing runs them. The status rules are in `docs/api.md`.

- [ ] Add a scheduler loop to the worker process that calls `HealthCheckRepository.ClaimDue(now, batchSize, hold)` every few seconds. `hold` must be longer than the probe timeout.
- [ ] For each claimed check, look up the agent instance address (`ClaimDue` returns checks only).
- [ ] Run `pkg/probe.HTTP` or `pkg/probe.TCP` with the check's timeout.
- [ ] Map the result to passing, warning or critical, then call `HealthCheckRepository.Record(id, status, output, now+interval)`.
- [ ] Confirm how `pkg/probe` reports results and whether the instance lookup exists in the db layer before implementing.

## Planned checks (from `docs/api.md`, not started)

- [ ] Card freshness: `/.well-known/agent-card.json` still matches the registered card
- [ ] A2A probe: `GetTask` or a cheap `SendMessage` against a health skill
- [ ] Skill probe: named skill still accepts the declared input modes
- [ ] State queries: `GET /v1/workspaces/{workspace}/health/state/{passing|warning|critical}`
