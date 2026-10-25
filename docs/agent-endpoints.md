# Agent endpoints

What an agent calls to find other agents and talk to them. All paths are under `/api/v1/workspaces/{workspace}`.

Endpoint reference: [api.md](api.md). Caller auth and policy: [agent-identity.md](agent-identity.md).

```
GET  /agents/discover                                  search agents by skill, tags, input mode and capability
POST /agents/discover/resolve                          pick one agent and instance to call
GET  /agents/catalog/skills                            list every skill in the workspace
GET  /agents/catalog/skills/{skill}/agents             list agents that offer a skill
GET  /agents/catalog/{agentName}/card                  get an agent card from the catalog
GET  /agents/{agentName}/.well-known/agent-card.json   get an agent card at the A2A well-known path
POST /agents/{agentName}/message:send                  send an A2A message to a named agent
POST /agents/skills/{skill}/message:send               send an A2A message to an agent that offers a skill
GET  /agents/tasks/{taskId}                            get an A2A task
GET  /agents/watch/catalog                             long-poll for catalog changes
```

Typical flow: find the skill (`catalog/skills` or `discover`), read the card, send the message, then poll the task if the reply is one.

---

## Discovery

### `GET /agents/discover`

Search the workspace for agents.

```http
GET /agents/discover?skill=user.profile.get&tags=profile&input_mode=application/json&capability=streaming&passing=true&limit=5
```

- `skill`: exact skill id on the card
- `tags`: skill or agent tags
- `input_mode`: MIME type the skill accepts
- `capability`: card capability, such as `streaming`
- `passing`: only agents with a healthy instance
- `limit`: maximum number of agents

Returns the matching agents with their cards. Use it when you know roughly what you need and want to see the options.

### `POST /agents/discover/resolve`

Let Sloth pick one agent and instance for you.

```json
{
  "skill": "user.profile.get",
  "strategy": "least_outstanding_tasks",
  "constraints": { "capabilities": ["streaming"] }
}
```

`strategy` is `random`, `round_robin` or `least_outstanding_tasks`. Returns one agent, one instance and its card, with the card URL pointing at the Sloth gateway.

---

## Catalog

### `GET /agents/catalog/skills`

List the skills the caller may call, based on its intentions: id, name, description and tags. Use it to turn a phrase like "get user profile" into an exact skill id such as `user.profile.get`. See [agent-catalog.md](agent-catalog.md).

### `GET /agents/catalog/skills/{skill}/agents`

List the agents the caller may call for a skill. Use it when you have a skill id and want to know who provides it. See [agent-catalog.md](agent-catalog.md).

### `GET /agents/catalog/{agentName}/card`

Return the stored card of one agent as a regular Sloth API resource.

### `GET /agents/{agentName}/.well-known/agent-card.json`

Return the same card at the standard A2A location, so off-the-shelf A2A clients can find it.

- Served with a strong `ETag` (SHA-256 of the card) and `Cache-Control: no-cache`; `If-None-Match` returns `304` until the card changes.
- Returns `404` until an instance has supplied the card.
- `supportedInterfaces[].url` points at the Sloth gateway, not the agent instance.

### `GET /agents/watch/catalog`

Long-poll for catalog changes.

```http
GET /agents/watch/catalog?index=42
```

Blocks until the catalog changes after `index`, then returns the changes and a new index. Use it to hot-reload cached cards without polling constantly.

---

## Gateway

These endpoints authenticate the calling agent with an agent identity, not a user session or workspace access key. See [agent-identity.md](agent-identity.md#gateway).

- `api_key`: `X-Source-Agent: <agent>` and `X-API-Key: <secret>`
- `basic_auth`: `X-Source-Agent: <agent>` and `Authorization: Basic …`
- `client_credentials`: `X-Source-Agent: <agent>` and `Authorization: Bearer <token>`

### `POST /agents/{agentName}/message:send`

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

### `POST /agents/skills/{skill}/message:send`

Same as above, but addressed by skill.

```http
POST /agents/skills/user.profile.get/message:send
X-Source-Agent: support-assistant
X-API-Key: <secret>
A2A-Version: 1.0
Content-Type: application/json
```

Sloth keeps the agents whose card offers the skill and that the caller may call for that skill, then picks a random live instance across all of them. `404 skill_not_found` when no card offers the skill, `403 a2a_call_denied` when the caller may call none of them.

### `GET /agents/tasks/{taskId}`

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

### Errors

- `401 invalid_agent_identity`: missing or wrong credential, or disabled identity
- `403 a2a_call_denied`: no `allow` intention for the caller and destination
- `404 agent_not_found`: unknown destination agent
- `404 skill_not_found`: no agent offers the skill
- `404 task_not_found`: unknown task, or a task started by another agent
- `502 agent_unreachable`: the instance could not be reached
- `503 no_healthy_instance`: no instance passes its checks

These errors come from Sloth as `{"errorMessage": "…"}`. An error returned by the instance itself is passed through with its own status and body.

---

## Reserved agent names

These endpoints share the `/agents` prefix with `/agents/{agentName}`, so an agent cannot be named `discover`, `catalog`, `skills`, `tasks` or `watch`.
