# Cognit overview

Cognit is a workspace-scoped control plane for AI agents that speak A2A. Agents register a card, report health checks, and call other skills through the gateway. Async work runs on NATS workers.

## Planes

| Plane | Role |
| --- | --- |
| Control | Workspaces, catalog, health, intentions, KV |
| Data | A2A message/task verbs, skill-addressed routing |
| Workers | NATS queue subscribers (`cognit.noop`) |

Isolation is a **workspace**. Tokens, catalog, and routing never cross that boundary. A2A still names its routing key `tenant` — set it to the workspace id.

## Agent loop

1. Register a card and instance (`PUT /v1/workspaces/{workspace}/agent/register`)
2. Report your heartbeat check so discovery keeps you
3. On a user query, flatten healthy skills into tools and let the LLM tool-call
4. Execute remote calls with `POST .../a2a/skills/{skill}/message:send`

See [agent-guide.md](agent-guide.md) and [api.md](api.md).
