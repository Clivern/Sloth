# TODO

## Gateway endpoints (stubbed, return 501)

Handlers are placeholders. `api/gateway_stubs_test.go` pins each one to `501 not_implemented`. When a handler is implemented, delete its 501 case there and remove the `t.Skip` in its sample test file.

- [ ] Catalog: `GET /catalog/skills`, `GET /catalog/skills/{skill}/agents`, `GET /catalog/agents/{agentName}/card` (`api/catalog.go`)
- [ ] Discover: `GET /discover`, `POST /discover/resolve` (`api/discover.go`). Sample cases are skipped in `api/discover_test.go`.
- [ ] A2A: `POST /a2a/agents/{agentName}/message:send`, `POST /a2a/skills/{skill}/message:send`, `GET /a2a/tasks/{taskId}` (`api/a2a.go`)
- [ ] Watch: `GET /watch/catalog` (`api/watch.go`). Sample cases are skipped in `api/watch_test.go`.
- [ ] Watch: `GET /watch/health` and `GET /watch/kv/{key}` are in `docs/api.md` but not routed in `core/server.go`.

## Permissions for gateway routes

`core/server.go` lines 153–162 still carry `// TODO: perm for ...` comments on these routes. Pick the right permission for discovery, catalog, A2A and watch, then update the middleware config.

## Tests

- [ ] `ReadyAction` (`GET /api/v1/public/_ready`, `api/ready.go`) has no test. It is the only API handler without one.
- [ ] `module/bus.go`: `StopBus`, `EnqueueTask` and `RepublishPendingTasks` have no tests. They need a live NATS server, and the last two also need Postgres. Install `nats-server` locally and cover them.
- [ ] Module coverage is 78.1% of statements. The gap is mostly `bus.go`.

## Agent instance health checks

`http` and `tcp` checks are stored and scheduled, but nothing runs them yet. The status rules are in `docs/api.md`.

Agreed design: the scheduler claims due checks in Postgres and publishes one job per check over NATS. Runners probe and publish results. The result writer is the only code that writes outcomes. Postgres is the source of truth for what is due. NATS only delivers the work.

- [ ] Add a scheduler loop to the worker process. Each tick calls `HealthCheckRepository.ClaimDue(now, batchSize, hold)`. `hold` must be longer than the probe timeout plus queue and result delay.
- [ ] Put the instance address, port, type, path, scheme and timeout in each job, so runners never query Postgres.
- [ ] Publish jobs to a NATS subject in a queue group (for example `health.check.run`), so each check runs on one replica.
- [ ] Runners call `pkg/probe.HTTP` or `pkg/probe.TCP` and publish the result to `health.check.result`.
- [ ] Result writer (queue group `health-recorders`) maps the result to passing, warning or critical, then calls `HealthCheckRepository.Record(id, status, output, now+interval)`. It ignores a result older than the row's `last_run_at`.
- [ ] Stale probe results: when a probe check's `last_run_at` is older than two intervals, treat it as critical when computing instance and agent health. This covers a dead runner or scheduler.
- [ ] SSRF guard: probes may only target the registered instance address and port. Allow only `http` and `https` schemes. Decide whether to block loopback and link-local ranges.
- [ ] Tests for `ClaimDue` under two concurrent callers, to cover the `SKIP LOCKED` guarantee.

## Planned checks (from `docs/api.md`, not started)

- [ ] Card freshness: `/.well-known/agent-card.json` still matches the registered card
- [ ] A2A probe: `GetTask` or a cheap `SendMessage` against a health skill
- [ ] Skill probe: named skill still accepts the declared input modes
- [ ] State queries: `GET /v1/workspaces/{workspace}/health/state/{passing|warning|critical}`

## Schema

- [ ] Existing databases that ran migration `20260101000028` still have `health_checks.agent_instance_id`. Rename it to `instance_id` by hand, or reset the database. Migration 028 was edited in place, so it will not apply the rename.
- [ ] Naming clash: `agent_instances` already has an `instance_id` column (the node name). `health_checks.instance_id` holds the `agent_instances.id` UUID. Consider a clearer name such as `agent_instance_pk` or `instance_pk`.

## Open questions

- [ ] `ValidAccessKeyPermissions(nil)` returns false, so an access key must have at least one permission. Confirm this is intended.

## Style

- [ ] Enforce the 100-character signature rule (`memory/go-multiline-signatures`) with a lint check, so long signatures are caught in CI rather than by hand.
