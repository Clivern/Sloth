# API Specs

Each `.feature` file states how one area of the Cognit API must behave. The specs are written for an agent to run against a live server: you give it a server URL and a user API key, and it works through the scenarios with real HTTP requests and reports what passed and what didn't.

## Files

| File | Covers |
|---|---|
| `conventions.feature` | Error shape, language, paging, bad ids and unknown routes, for every route |
| `public.feature` | Probes, setup, logout, metrics |
| `me.feature` | `/me` for each kind of credential |
| `profile.feature` | The user's own profile |
| `settings.feature` | Platform settings (platform admins only) |
| `api_keys.feature` | User API keys |
| `workspaces.feature` | Creating, reading, renaming and deleting workspaces |
| `members.feature` | Workspace members and roles |
| `invites.feature` | Workspace invites |
| `access_keys.feature` | Workspace access keys and their permissions |
| `billing.feature` | Billing (SaaS edition) |
| `audits.feature` | The workspace audit log |
| `agents.feature` | Registering agents and serving their cards |
| `instances.feature` | Agent instances and the passing filter |
| `checks.feature` | Health check templates and ttl reporting |
| `identities.feature` | Agent identities and the client credentials token route |
| `protection.feature` | The credential the gateway presents to an agent |
| `intentions.feature` | Allow and deny rules between agents |
| `kv.feature` | The workspace key-value store |
| `traffic.feature` | Recorded gateway calls |
| `isolation.feature` | Nothing crosses from one workspace to another, swept across every resource |
| `pending.feature` | Routes that exist but are not built yet |

## Running a spec

Give the agent:

| Input | Example | Notes |
|---|---|---|
| `COGNIT_URL` | `http://localhost:8000` | Base URL of the server under test |
| `COGNIT_API_KEY` | `ck_...` | A user API key, sent as `X-API-Key` |
| Files | `agents.feature` | One or more files, or "all" |
| Tags | `@smoke`, `not @mutates` | Optional filter |

For example: *"Run specs/test/agents.feature against http://localhost:8000 with API key ck_xxx."*

## Rules for the agent

1. **Make your own workspace.** Before the first scenario, create a workspace named `spec-<runId>`, where `runId` is a short random id. This is "the run's workspace". Every resource you create goes in it or in another workspace the scenario asks you to create, and its name starts with `spec-<runId>`. Never read or change anything that you didn't create in this run.
2. **Clean up.** After the last scenario, delete everything you created, then the workspaces, even if scenarios failed. Report anything you couldn't delete.
3. **Mint keys inside the run.** When a step says "an access key with only X", create a workspace access key in the run's workspace with exactly those permissions, and delete it during cleanup. "Every permission an access key may hold" means the full `AccessKeyPermissions` list in `module/permission.go`.
4. **Respect the tags.**
   - `@readonly`: makes no lasting change outside the run's workspace.
   - `@mutates`: creates and deletes data inside the run's workspaces.
   - `@smoke`: a small set of key checks that runs quickly.
   - `@slow`: waits on timers such as ttl expiry, or uses up a rate limit. Skip unless asked.
   - `@destructive`: changes server-wide state. Skip unless you're told to run it.
   - `@admin`: needs the API key's user to be a platform admin. Skip, and report as skipped, if `/me` says otherwise.
   - `@saas` / `@oss`: run only on that edition, as reported by `/api/v1/public/action/setup/status`.
5. **Take the routes from the file.** Each feature lists its routes in the description, and `openapi.yaml` has the request and response shapes. Don't guess routes that aren't there. A path written as `agents/x` is relative to `/api/v1/workspaces/{the run's workspace}/`.
6. **Follow Then steps exactly.** A Then step passes only if the response matches every part of it: status code, fields and values. "The body has no X field" means the key is absent, not empty. If a step is ambiguous, mark the scenario `unclear` and say why instead of picking a reading.
7. **Keep scenarios independent.** Each one sets up what it needs, and Background steps run before every scenario. Don't rely on another scenario having run first.
8. **Mind the rate limits.** The token route and setup are rate limited per client IP. Leave the `@slow` rate-limit scenario until last.
9. **Keep secrets out.** Never write the API key, access keys, client secrets or tokens into the report. Replace them with `<redacted>`.

## Report

Write the results to `specs/reports/<YYYY-MM-DD>-<runId>.md` (the folder is gitignored):

```markdown
# Spec run 2026-10-16 a1b2c3

Server: http://localhost:8000 · Edition: oss · Files: all · Tags: not @slow

| Result | Count |
|---|---|
| pass | 180 |
| fail | 4 |
| skipped | 12 |
| unclear | 1 |

## Failures

### members.feature › The owner cannot demote themselves
- Step: Then the response is 400 or 409
- Request: PUT /api/v1/workspaces/{ws}/members/{userId} {"role":"readonly"}
- Expected: 400 or 409
- Got: 200 with role "readonly"
```

For every failure, record the step, the request (method and path, plus the body when it matters), what was expected, and what came back. Then list the skipped scenarios with the reason.

## Writing a spec

- Each file covers one area and starts with a short description of what the file guarantees and which routes it uses.
- Describe behaviour, not implementation. Name the request in the When step, and put status codes and fields in the Then steps.
- Each scenario should check one thing. If a Then step needs "and also", split the scenario.
- Use `<runId>` in names so the agent can substitute the real id, and `<a random UUID>` for ids that must not exist.
- A spec states what the server **should** do. If the server does something else, the scenario fails, and that is the point: fix whichever of the two is wrong.
