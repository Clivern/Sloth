# Follow-ups

All five endpoints below are built. These are the known gaps to enhance next.

## Docs and specs

- Add the three message and task endpoints to `openapi.yaml` and `specs/test/catalog.feature`.
- Decide on `cardUrl` in `GET /agents/catalog/skills/{skill}/agents`: it points at `/agents/{name}/.well-known/agent-card.json`, a route that was dropped, so it returns 404. Restore the route (find it with `git log --diff-filter=D -1 -- api/card.go`) or drop the field.
- `docs/agent-catalog.md` and `docs/agent-endpoints.md` are drafts and are not committed.

## Traffic

- `ai_traffic.task_id` is never set now that tasks live in `ai_agents_tasks`. Set `call.TaskId` in `Agents.send` when the reply is a task and in `Agents.GetTask`.
- `destination_instance` mixes values: the instance id for forwarded calls, the agent name for denied and no-instance calls, empty for a denied skill call. Store the agent name always (the instance id is already in `meta`) or add a `destination_agent` column.
- `source_instance` holds the caller's agent name, not an instance; the column name is misleading.
- Not recorded: 401 (fails in `AgentAuth`), 404 for an unknown agent, skill or task, and 500 storage errors.

## A2A compatibility

- Add `GET /agents/{agentName}/tasks/{taskId}` as an alias: A2A clients call `{card url}/tasks/{id}`, which is that path.
- Cards stored before the agents endpoints have no `a2a_url` agent meta and still point at `/a2a/agents/{name}`; Sloth forwards to them without the agent's path until the card is fetched again (re-register with a new version).
- Not supported yet: `message:stream`, push notifications, `tasks/cancel`.

## Data

- Migrations `…036` (`ai_agent_skills`) and `…037` (`ai_agents_tasks`) were edited in place; databases that already ran them must be reset.
- Cards stored before `SyncSkills` existed have no `ai_agent_skills` rows until the card is written again.
- `SyncSkills` fails the whole card write when `skills` is not an array or a skill id or name is longer than 120 characters.
- `ai_agents_tasks` grows forever; add a cleanup job for old rows (indexed on `created_at`).
- The intention rule (a skill rule beats an agent rule) lives in `db/skill.go` SQL and in `IntentionRepository.Match`; keep them in sync.

## Cleanup

- Unused: `TrafficRepository.GetAllowedByTaskId`, the `agent_card_pending` locale key.
- The instance repository live methods take a `now` parameter; `ListCallable` and `ListCallableAgents` compute it inside.
- Integration tests in `db/skill_test.go`, `db/task_test.go` and the skill cases in `db/agent_test.go` have not been run against Postgres yet.

---

# Agent catalog

What an agent calls to find the skills and agents it may call, send them a message and fetch the result. All paths are under `/api/v1/workspaces/{workspace}`.

```
GET  /agents/catalog/skills                    list the skills the caller may call
GET  /agents/catalog/skills/{skill}/agents     list the agents the caller may call for a skill
POST /agents/{agentName}/message:send          send an A2A message to a named agent
POST /agents/skills/{skill}/message:send       send an A2A message to an agent that offers a skill
GET  /agents/tasks/{taskId}                    get the task started by an earlier message:send
```

Typical flow: list the skills, pick an agent (or let Sloth pick by skill), send the message, then poll the task if the reply is one.

---

## Auth

An agent authenticates with its agent identity, the same way it does on the gateway:

- `api_key`: `X-Source-Agent: <agent>` and `X-API-Key: <secret>`
- `basic_auth`: `X-Source-Agent: <agent>` and `Authorization: Basic …`
- `client_credentials`: `X-Source-Agent: <agent>` and `Authorization: Bearer <token>`

The catalog results are filtered by intentions, so an agent only sees what `message:send` would let it call. `message:send` and `tasks` accept only an agent identity.

A user session or workspace access key with `CanListAgents` sees every skill and agent in the workspace, with no intention filter. The UI and admins use this.

## Intention filter

A skill on a destination agent is callable when `Match(caller, destination, skill)` returns an `allow` intention:

- An `allow` with no skill covers every skill on the destination card.
- A skill-specific intention wins over one with no skill, so `deny user-directory/user.delete` removes that one skill from a broader `allow user-directory`.
- No matching intention means deny.
- The caller never appears in its own results.

---

## `GET /agents/catalog/skills`

List the skills the caller may call.

```http
GET /agents/catalog/skills?tags=profile&passing=true&limit=20&offset=0
X-Source-Agent: support-assistant
X-API-Key: <secret>
```

- `tags`: comma-separated; keeps skills that have at least one of these tags
- `passing`: when `true`, only count agents that have a live instance
- `limit`, `offset`: pagination, applied after filtering and sorting

Sloth goes through each skill on each agent card in the workspace, keeps the callable ones, groups them by skill id and sorts them by id. When cards disagree on a skill's `name`, `description` or `tags`, the values come from the agent whose name sorts first.

Response `200`:

```json
{
  "skills": [
    {
      "id": "user.profile.get",
      "name": "Get user profile",
      "description": "Return a user's profile by id",
      "tags": ["profile", "users"],
      "inputModes": ["application/json"],
      "outputModes": ["application/json"],
      "agents": 2
    },
    {
      "id": "user.profile.update",
      "name": "Update user profile",
      "description": "Update a user's display name and avatar",
      "tags": ["profile", "users"],
      "inputModes": ["application/json"],
      "outputModes": ["application/json"],
      "agents": 1
    }
  ],
  "_meta": {
    "limit": 20,
    "offset": 0,
    "total": 2
  }
}
```

`agents` is the number of agents the caller may call for the skill. List them with `GET /agents/catalog/skills/{skill}/agents`.

A caller with no `allow` intentions gets `200` and an empty `skills` list.

---

## `GET /agents/catalog/skills/{skill}/agents`

List the agents the caller may call for one skill.

```http
GET /agents/catalog/skills/user.profile.get/agents?passing=true&limit=20&offset=0
X-Source-Agent: support-assistant
X-API-Key: <secret>
```

- `{skill}`: exact skill id on the card, URL-escaped
- `passing`: when `true`, only return agents that have a live instance
- `limit`, `offset`: pagination

Sloth finds the agents whose card offers the skill, keeps the ones the caller may call for it and sorts them by name.

Response `200`:

```json
{
  "skill": {
    "id": "user.profile.get",
    "name": "Get user profile",
    "description": "Return a user's profile by id",
    "tags": ["profile", "users"]
  },
  "agents": [
    {
      "name": "user-directory",
      "description": "Source of truth for user profiles",
      "version": "1.4.0",
      "status": "passing",
      "instances": {
        "total": 3,
        "passing": 2
      },
      "cardUrl": "https://sloth.example.com/api/v1/workspaces/acme/agents/user-directory/.well-known/agent-card.json",
      "sendUrl": "https://sloth.example.com/api/v1/workspaces/acme/agents/user-directory/message:send"
    },
    {
      "name": "user-directory-legacy",
      "description": "Read-only mirror of the old user store",
      "version": "0.9.2",
      "status": "critical",
      "instances": {
        "total": 1,
        "passing": 0
      },
      "cardUrl": "https://sloth.example.com/api/v1/workspaces/acme/agents/user-directory-legacy/.well-known/agent-card.json",
      "sendUrl": "https://sloth.example.com/api/v1/workspaces/acme/agents/user-directory-legacy/message:send"
    }
  ],
  "_meta": {
    "limit": 20,
    "offset": 0,
    "total": 2
  }
}
```

- `status`: `passing` when at least one instance is live, `critical` otherwise
- `cardUrl` and `sendUrl` point at the Sloth gateway, never at an instance
- The full card is not included; fetch it from `cardUrl`, which supports `ETag` and `If-None-Match`

When agents offer the skill but the caller may call none of them, the response is `200` with an empty `agents` list.

---

## `POST /agents/{agentName}/message:send`

Send an A2A message to a named agent.

```http
POST /agents/user-directory/message:send
X-Source-Agent: support-assistant
X-API-Key: <secret>
A2A-Version: 1.0
Content-Type: application/json

{
  "message": {
    "messageId": "9b1c7e2a-4f0d-4c1e-9a55-0d2f6b8e1a10",
    "role": "ROLE_USER",
    "parts": [{ "data": { "user_id": "u_123" } }]
  },
  "configuration": {
    "acceptedOutputModes": ["application/json"]
  }
}
```

- `A2A-Version` is optional and passed through to the instance.
- The body is forwarded byte for byte; Sloth does not validate it.

The gateway:

1. Verifies the caller → `401`
2. Matches an `allow` intention for the caller and the agent; only intentions without a skill match on this route → `403`
3. Picks a random live instance → `503` when none passes its checks
4. Forwards to `<advertised scheme>://<address>:<port><advertised path>/message:send` with the destination's protection credential → `502` when the instance cannot be reached
5. Records the call in traffic, with the task id, agent and instance when the reply carries a task

The instance status, `Content-Type` and body are returned as is. Under A2A 1.0 the reply is either a direct message, with nothing to poll:

```json
{
  "message": {
    "messageId": "c4e0…",
    "role": "ROLE_AGENT",
    "parts": [{ "data": { "user_id": "u_123", "name": "Ada Lovelace" } }]
  }
}
```

or a task, whose `task.id` you keep to fetch the result:

```json
{
  "task": {
    "id": "task-7f3a",
    "contextId": "ctx-21b9",
    "status": {
      "state": "TASK_STATE_WORKING",
      "timestamp": "2026-10-23T10:15:02Z"
    }
  }
}
```

---

## `POST /agents/skills/{skill}/message:send`

Same as above, but addressed by skill.

```http
POST /agents/skills/user.profile.get/message:send
X-Source-Agent: support-assistant
X-API-Key: <secret>
A2A-Version: 1.0
Content-Type: application/json
```

Sloth keeps the agents whose card offers the skill and that the caller may call for that skill, then picks a random live instance across all of them. `404 skill_not_found` when no card offers the skill, `403 a2a_call_denied` when the caller may call none of them.

---

## `GET /agents/tasks/{taskId}`

Get a task started by an earlier `message:send`.

```http
GET /agents/tasks/task-7f3a
X-Source-Agent: support-assistant
X-API-Key: <secret>
A2A-Version: 1.0
```

The gateway:

1. Verifies the caller → `401`
2. Looks up the task id in traffic → `404` when unknown or started by another agent
3. Checks the intention again for the skill of the original call → `403` when it was revoked
4. Forwards `GET /tasks/{taskId}` to the instance that accepted the task, never another one → `503` when that instance is gone, `502` when it cannot be reached

The instance reply is returned as is. A finished task:

```json
{
  "id": "task-7f3a",
  "contextId": "ctx-21b9",
  "status": {
    "state": "TASK_STATE_COMPLETED",
    "timestamp": "2026-10-23T10:15:09Z"
  },
  "artifacts": [
    {
      "artifactId": "a-1",
      "parts": [{ "data": { "user_id": "u_123", "name": "Ada Lovelace" } }]
    }
  ]
}
```

- Poll while `state` is `TASK_STATE_SUBMITTED` or `TASK_STATE_WORKING`.
- `TASK_STATE_COMPLETED`, `TASK_STATE_FAILED`, `TASK_STATE_CANCELED` and `TASK_STATE_REJECTED` are final.
- On `TASK_STATE_INPUT_REQUIRED`, send another `message:send` to the same agent with the task's `taskId` and `contextId` on the message.

---

## Errors

- `400 invalid_workspace_id`: workspace id missing from the path
- `401 invalid_agent_identity`: missing or wrong credential, or disabled identity
- `403 a2a_call_denied`: no `allow` intention for the caller and destination (`message:send` and `tasks`)
- `404 agent_not_found`: unknown destination agent (`/agents/{agentName}/message:send`)
- `404 skill_not_found`: no agent card in the workspace offers the skill (`/agents/catalog/skills/{skill}/agents` and `/agents/skills/{skill}/message:send`)
- `404 task_not_found`: unknown task, or a task started by another agent
- `500 failed_catalog_request`: storage error on a catalog endpoint
- `502 agent_unreachable`: the instance could not be reached
- `503 no_healthy_instance`: no instance passes its checks

These errors come from Sloth as `{"errorMessage": "…"}`. An error returned by the instance itself is passed through with its own status and body.
