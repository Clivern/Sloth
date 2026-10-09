# Cognit API

Cognit is a control plane for AI agents that speak the [A2A protocol](https://a2a-protocol.org/v1.0.0/specification/). A2A standardizes how you talk to an agent (Agent Card, messages, tasks). It does not standardize a registry. Cognit fills that gap: workspace-scoped catalog, health, policy, and a gateway in front of A2A.

Isolation is a **workspace**. Catalog, health, intentions, KV, and A2A routing never cross that boundary. A2A v1.0 still names its routing key `tenant` on `AgentInterface` and as an operation path parameter — set that field to the workspace id.

Base URL:

```
https://cognit.example.com
```

| Plane | Role |
| --- | --- |
| Control | Register, discover, health, policy |
| Data | Speak A2A to a resolved agent |

Auth on both planes: `Authorization: Bearer <token>`. Workspace is in the path. Optional `A2A-Version: 1.0` on data-plane calls.

---

## 1. Workspaces and identity

```
POST   /v1/workspaces
GET    /v1/workspaces
GET    /v1/workspaces/{workspace}
PATCH  /v1/workspaces/{workspace}
DELETE /v1/workspaces/{workspace}

POST   /v1/workspaces/{workspace}/api-keys
GET    /v1/workspaces/{workspace}/api-keys
DELETE /v1/workspaces/{workspace}/api-keys/{key_id}

GET    /v1/workspaces/{workspace}/principals/me
```

A token is scoped to one workspace. Cross-workspace calls fail closed (`404`, not `403`) so workspace IDs stay private.

---

## 2. Control plane — catalog

Agents publish an A2A **Agent Card**. The registry stores the card plus runtime metadata (address, lease, health).

```
PUT    /v1/workspaces/{workspace}/catalog/agents/{agent}
GET    /v1/workspaces/{workspace}/catalog/agents/{agent}
DELETE /v1/workspaces/{workspace}/catalog/agents/{agent}
GET    /v1/workspaces/{workspace}/catalog/agents
       ?skill=invoice.extract
       &tag=finance
       &status=passing
       &capability=streaming

GET    /v1/workspaces/{workspace}/catalog/agents/{agent}/card
PUT    /v1/workspaces/{workspace}/catalog/agents/{agent}/card

GET    /v1/workspaces/{workspace}/catalog/skills
GET    /v1/workspaces/{workspace}/catalog/skills/{skill}
GET    /v1/workspaces/{workspace}/catalog/skills/{skill}/agents
```

`PUT .../agents/{agent}` registers the agent. Body is the Agent Card plus instance info:

```json
{
  "card": {
    "name": "Invoice Extractor",
    "description": "Extracts line items from invoices",
    "version": "1.4.0",
    "capabilities": {
      "streaming": true,
      "pushNotifications": true,
      "extendedAgentCard": true
    },
    "defaultInputModes": ["text/plain", "application/pdf"],
    "defaultOutputModes": ["application/json"],
    "skills": [
      {
        "id": "invoice.extract",
        "name": "Extract invoice",
        "description": "Parse a PDF invoice into structured line items",
        "tags": ["finance", "ocr"]
      }
    ],
    "supportedInterfaces": [
      {
        "url": "https://invoice-worker.internal:8080/a2a/v1",
        "protocolBinding": "HTTP+JSON",
        "protocolVersion": "1.0",
        "tenant": "acme"
      }
    ]
  },
  "instance": {
    "id": "invoice-extractor-3",
    "address": "10.0.12.41",
    "port": 8080,
    "datacenter": "eu-west-1",
    "meta": { "model": "gpt-5", "region": "eu" }
  },
  "lease_ttl": "30s"
}
```

`GET .../catalog/agents` returns healthy instances and cards.

---

## 3. Control plane — self-register and leases

Workers talk to a local or cluster agent endpoint, not the full catalog UI.

```
PUT    /v1/workspaces/{workspace}/agent/register
PUT    /v1/workspaces/{workspace}/agent/deregister/{agent}

PUT    /v1/workspaces/{workspace}/agent/instances/{instance}
DELETE /v1/workspaces/{workspace}/agent/instances/{instance}

PUT    /v1/workspaces/{workspace}/agent/instances/{instance}/renew
GET    /v1/workspaces/{workspace}/agent/instances/{instance}

POST   /v1/workspaces/{workspace}/sessions
GET    /v1/workspaces/{workspace}/sessions/{session}
PUT    /v1/workspaces/{workspace}/sessions/{session}/renew
DELETE /v1/workspaces/{workspace}/sessions/{session}
```

Registration is lease-based. Miss the TTL and the instance is marked critical, then dropped from discovery.

---

## 4. Control plane — health

Checks are defined once per agent as templates, and copied onto every instance of that agent, including instances that register later. Results stay per instance, so one failing instance is dropped from discovery without affecting the others.

```
GET    /v1/workspaces/{workspace}/agents/{agent}/checks
PUT    /v1/workspaces/{workspace}/agents/{agent}/checks/{check}
DELETE /v1/workspaces/{workspace}/agents/{agent}/checks/{check}

POST   /v1/workspaces/{workspace}/agents/{agent}/instances/{instance}/checks/{check}/pass
POST   /v1/workspaces/{workspace}/agents/{agent}/instances/{instance}/checks/{check}/warn
POST   /v1/workspaces/{workspace}/agents/{agent}/instances/{instance}/checks/{check}/fail
```

```json
PUT /v1/workspaces/{workspace}/agents/support-assistant/checks/http
{ "type": "http", "path": "/healthz", "interval": 10, "timeout": 2 }
```

An instance's status is the worst status across its checks: any critical check, or an expired lease or TTL check, makes it critical and drops it from discovery. Warning instances stay in discovery.

Check types:

- **Lease** — built in. Every instance has one; renew extends it.
- **TTL heartbeat** (`ttl`) — the instance reports pass, warn or fail before the TTL runs out.
- **HTTP** (`http`) — Cognit requests `path` on the instance; 2xx is passing, 429 is warning, anything else is critical.
- **TCP** (`tcp`) — Cognit opens a connection to the instance address and port.

Planned:

- **Card freshness** — `/.well-known/agent-card.json` still matches the registered card
- **A2A probe** — `GetTask` or a cheap `SendMessage` against a health skill
- **Skill probe** — named skill still accepts the declared input modes
- **State queries** — `GET /v1/workspaces/{workspace}/health/state/{passing|warning|critical}`

---

## 5. Control plane — discovery and saved queries

A2A’s own discovery story is a curated registry with an unspecified API. This is that API.

```
GET    /v1/workspaces/{workspace}/discover
       ?skill=invoice.extract
       &tags=finance
       &input_mode=application/pdf
       &capability=streaming
       &passing=true
       &limit=5

POST   /v1/workspaces/{workspace}/discover/resolve
```

Resolve body:

```json
{
  "skill": "invoice.extract",
  "strategy": "least_outstanding_tasks",
  "constraints": {
    "datacenter": "eu-west-1",
    "capabilities": ["streaming"]
  }
}
```

Response is the card the caller should use, with `supportedInterfaces[0].url` pointed at Cognit’s gateway (not the raw pod IP). A2A’s `tenant` field is the workspace id:

```json
{
  "agent": "invoice-extractor",
  "instance": "invoice-extractor-3",
  "card": {
    "supportedInterfaces": [
      {
        "url": "https://cognit.example.com/v1/workspaces/acme/a2a/agents/invoice-extractor",
        "protocolBinding": "HTTP+JSON",
        "protocolVersion": "1.0",
        "tenant": "acme"
      }
    ]
  }
}
```

Saved queries:

```
PUT    /v1/workspaces/{workspace}/query/{name}
GET    /v1/workspaces/{workspace}/query/{name}
DELETE /v1/workspaces/{workspace}/query/{name}
GET    /v1/workspaces/{workspace}/query/{name}/execute
```

---

## 6. Control plane — intentions

Who may call whom.

```
PUT    /v1/workspaces/{workspace}/intentions
GET    /v1/workspaces/{workspace}/intentions
GET    /v1/workspaces/{workspace}/intentions/{id}
DELETE /v1/workspaces/{workspace}/intentions/{id}

POST   /v1/workspaces/{workspace}/intentions/check
```

```json
{
  "source": { "agent": "cfo-assistant" },
  "destination": { "agent": "invoice-extractor", "skill": "invoice.extract" },
  "action": "allow"
}
```

`POST .../intentions/check` is what the gateway calls on every A2A request. Default deny across workspaces. Default deny across agents inside a workspace unless an intention exists.

---

## 7. Control plane — KV, watches, events

```
GET    /v1/workspaces/{workspace}/kv/{key}
PUT    /v1/workspaces/{workspace}/kv/{key}
DELETE /v1/workspaces/{workspace}/kv/{key}
GET    /v1/workspaces/{workspace}/kv?prefix=agents/invoice-extractor/

GET    /v1/workspaces/{workspace}/watch/catalog?index={index}
GET    /v1/workspaces/{workspace}/watch/health?index={index}
GET    /v1/workspaces/{workspace}/watch/kv/{key}?index={index}

PUT    /v1/workspaces/{workspace}/event/{name}
GET    /v1/workspaces/{workspace}/event/{name}
```

Use KV for model pins, prompt refs, and feature flags — not secrets. Blocking `index` watches let agents long-poll and hot-reload cards and config.

---

## 8. Data plane — A2A well-known cards

A2A requires a card at `/.well-known/agent-card.json`. In a workspace-scoped registry you host per-agent cards, and optionally a registry card.

```
GET  /v1/workspaces/{workspace}/agents/{agent}/.well-known/agent-card.json
GET  /v1/workspaces/{workspace}/agents/{agent}/extendedAgentCard

GET  /.well-known/agent-card.json
```

The last one is Cognit itself (directory / router agent). Per-agent cards are what callers cache. Serve `Content-Type: application/a2a+json` and `ETag` from `card.version`.

If you need a public hostname per agent:

```
GET  https://{agent}.{workspace}.cognit.example.com/.well-known/agent-card.json
```

That still resolves to the same card, with `supportedInterfaces` aimed at the gateway.

---

## 9. Data plane — A2A HTTP+JSON

Official A2A v1.0 REST bindings, namespaced by workspace and agent. When forwarding, set A2A’s `tenant` path/header value to the workspace id.

```
POST   /v1/workspaces/{workspace}/a2a/agents/{agent}/message:send
POST   /v1/workspaces/{workspace}/a2a/agents/{agent}/message:stream

GET    /v1/workspaces/{workspace}/a2a/agents/{agent}/tasks
GET    /v1/workspaces/{workspace}/a2a/agents/{agent}/tasks/{id}
POST   /v1/workspaces/{workspace}/a2a/agents/{agent}/tasks/{id}:cancel
POST   /v1/workspaces/{workspace}/a2a/agents/{agent}/tasks/{id}:subscribe

POST   /v1/workspaces/{workspace}/a2a/agents/{agent}/tasks/{id}/pushNotificationConfigs
GET    /v1/workspaces/{workspace}/a2a/agents/{agent}/tasks/{id}/pushNotificationConfigs
GET    /v1/workspaces/{workspace}/a2a/agents/{agent}/tasks/{id}/pushNotificationConfigs/{configId}
DELETE /v1/workspaces/{workspace}/a2a/agents/{agent}/tasks/{id}/pushNotificationConfigs/{configId}

GET    /v1/workspaces/{workspace}/a2a/agents/{agent}/extendedAgentCard
```

JSON-RPC binding, same prefix, one POST:

```
POST   /v1/workspaces/{workspace}/a2a/agents/{agent}
```

Methods: `SendMessage`, `SendStreamingMessage`, `GetTask`, `ListTasks`, `CancelTask`, `SubscribeToTask`, `CreateTaskPushNotificationConfig`, `GetTaskPushNotificationConfig`, `ListTaskPushNotificationConfigs`, `DeleteTaskPushNotificationConfig`, `GetExtendedAgentCard`.

The gateway:

1. Authenticates the caller
2. Checks intentions
3. Picks a passing instance (or the instance sticky to `task.id`)
4. Forwards the A2A call
5. Records health and outstanding-task load

---

## 10. Skill-addressed A2A

Callers often do not know the agent name. They know a skill.

```
POST   /v1/workspaces/{workspace}/a2a/skills/{skill}/message:send
POST   /v1/workspaces/{workspace}/a2a/skills/{skill}/message:stream
GET    /v1/workspaces/{workspace}/a2a/skills/{skill}/tasks/{id}
POST   /v1/workspaces/{workspace}/a2a/skills/{skill}/tasks/{id}:cancel
POST   /v1/workspaces/{workspace}/a2a/skills/{skill}/tasks/{id}:subscribe
```

`{skill}` is `invoice.extract`. Cognit resolves, checks intentions, picks a healthy instance, and forwards the A2A call. Task IDs stay sticky to the instance that accepted the first message.

---

## 11. Cluster and ops

```
GET    /v1/status/leader
GET    /v1/status/peers
GET    /v1/health

GET    /v1/workspaces/{workspace}/datacenters
GET    /v1/workspaces/{workspace}/nodes
GET    /v1/workspaces/{workspace}/metrics
```

`GET /v1/health` is process liveness and is not workspace-scoped.

---

## Request shape

Every workspace-scoped call:

```
Authorization: Bearer <workspace-scoped token>
A2A-Version: 1.0
Content-Type: application/json
```

A2A send through the gateway:

```http
POST /v1/workspaces/acme/a2a/agents/invoice-extractor/message:send HTTP/1.1
Authorization: Bearer ...
A2A-Version: 1.0
Content-Type: application/json

{
  "message": {
    "role": "ROLE_USER",
    "parts": [
      { "text": "Extract line items" }
    ]
  }
}
```

---

## Build order

Thin slice that is already a useful A2A core:

1. `PUT/GET/DELETE /v1/workspaces/{workspace}/catalog/agents/{agent}`
2. `PUT .../agent/instances/{instance}/renew` (TTL)
3. `GET .../discover?skill=`
4. `GET .../agents/{agent}/.well-known/agent-card.json`
5. `POST .../a2a/agents/{agent}/message:send` and `GET .../tasks/{id}`
6. Intentions `check` on the gateway

KV, watches, saved queries, and skill-addressed routing come after that.

A2A’s job is the card plus message and task verbs. Cognit’s job is who is registered, who is healthy, who is allowed to call, and which instance gets the RPC — with workspace as a hard path boundary.
