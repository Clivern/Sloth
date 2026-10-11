# Agent identity and policy

Every gateway call answers two questions:

1. **Who is calling?** The credential (`agent_identity`).
2. **May they call this agent?** The `intentions` table.

Callers go through Cognit. They do not call another agent’s IP.

## Example

Only `agent1` and `agent2` may reach `agent3`:

| Source | Destination | Action |
| --- | --- | --- |
| `agent1` | `agent3` | allow |
| `agent2` | `agent3` | allow |

Each caller has **its own** key. `agent3` does not hand out a shared secret.

## Identity

`agent_identity` is the caller’s identity (API key or basic auth). Cognit issues the secret once and stores a hash.

Verify the secret → that row’s agent is the source. Fail → **401**.

Manage identities on the agent page in the dashboard (**Identities**), or through the API:

```
GET    /api/v1/workspaces/{workspace}/agents/{agent}/identities
POST   /api/v1/workspaces/{workspace}/agents/{agent}/identities
PUT    /api/v1/workspaces/{workspace}/agents/{agent}/identities/{identityId}
DELETE /api/v1/workspaces/{workspace}/agents/{agent}/identities/{identityId}
```

`POST` takes `name`, `type` (`api_key` or `basic_auth`) and, for basic auth, `username`. The response carries `secret` once. `PUT` renames an identity or sets `isActive`; a disabled identity stays listed but must not authenticate.

`X-Source-Agent` is only a lookup hint so Cognit does not scan every key: find that agent, then verify the secret against **its** hashes. If the header says `agent1` but the secret is `agent2`’s → **401**. The header never wins.

## Policy

`Match(workspace, source, destination, skill)` uses the **verified** source and the destination from the URL.

- allow → forward
- deny → **403**
- no row → workspace default (allow or deny)

Always authenticate. “Open” means any **authenticated** agent in the workspace, not anonymous traffic.

## Flow

```text
X-Source-Agent: agent1     →  which agent to look up
X-API-Key / basic          →  prove it
URL / skill                →  destination (agent3)

verify secret  →  401 if not
Match(...)     →  403 if not allowed
forward
```

## Protection

Identity proves who is calling. Protection keeps the destination reachable **only** through Cognit.

Each agent has at most one protection: the credential Cognit sends when it forwards a call. Configure the agent to reject calls without it.

```
GET    /api/v1/workspaces/{workspace}/agents/{agent}/protection
PUT    /api/v1/workspaces/{workspace}/agents/{agent}/protection
PUT    /api/v1/workspaces/{workspace}/agents/{agent}/protection/status
DELETE /api/v1/workspaces/{workspace}/agents/{agent}/protection
```

| Type | Gateway sends | Stored as |
| --- | --- | --- |
| `api_key` | secret in `header` (default `X-API-Key`) | encrypted |
| `basic_auth` | `username` + secret | encrypted |

`PUT` takes `type`, `username` (basic auth), `header` (API key) and an optional `secret`. Without `secret`, Cognit generates one and returns it once. Saving again replaces the credential. `PUT …/status` takes `isActive` and keeps the secret.
