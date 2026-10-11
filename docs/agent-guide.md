# Agent guide

You do two jobs: **announce yourself**, then **look up other cards when a user ask needs a skill you do not have**. You never call another agent’s IP. You ask Cognit, then send A2A through the gateway.

Assume workspace `acme`, you are `support-assistant`, instance `support-assistant-1`.

Endpoint reference: [api.md](api.md).

---

## 1. Register

Register the agent once, by name and version. The card is not sent here:

```http
PUT /api/v1/workspaces/acme/agents/support-assistant
Content-Type: application/json

{ "version": "1.0.0", "cardPath": "/.well-known/agent-card.json" }
```

`cardPath` is optional and defaults to `/.well-known/agent-card.json`. On boot, serve your A2A card there, then register your instance with a workspace access key:

```http
PUT /api/v1/workspaces/acme/agents/support-assistant/instances/support-assistant-1
X-Access-Key: <your-key>
Content-Type: application/json

{ "address": "10.0.12.20", "port": 8080 }
```

While the agent has no card, Cognit fetches it from the registering instance at `https://10.0.12.20:8080{cardPath}`, falling back to `http`. It replaces `supportedInterfaces[].url` (and a legacy top-level `url`, if present) with the gateway URL before storing it, so the URL your card advertises is overwritten. The first card stored wins; bumping the agent's version or card path clears it so the next instance to register supplies a fresh one. A failed fetch does not fail the instance registration.

Then define a heartbeat check on your agent, and report it before its TTL runs out:

```http
PUT /v1/workspaces/acme/agents/support-assistant/checks/heartbeat
{ "type": "ttl", "ttl": 30 }

POST /v1/workspaces/acme/agents/support-assistant/instances/support-assistant-1/checks/heartbeat/pass
```

If the TTL runs out, Cognit marks you critical and drops you from discovery. Other agents will not see you.

You are now listed. You are not allowed to call others until an intention says so. Discovery can still show you cards; the gateway blocks the actual `message:send` if policy denies it.

---

## 2. User query lands on you

Someone sends you A2A (usually via the gateway):

```http
POST /v1/workspaces/acme/a2a/agents/support-assistant/message:send
```

```json
{
  "message": {
    "role": "ROLE_USER",
    "parts": [{ "text": "What's on my profile and when does my plan renew?" }]
  }
}
```

You decide you need **user profile**. You do not have that skill. You explore the catalog.

---

## 3. Explore capabilities

You usually do **not** know the other agent’s name. You search by skill, tags, or browse the skill index.

**If you already have a skill id** (best path):

```http
GET /v1/workspaces/acme/discover?skill=user.profile.get&passing=true&limit=5
Authorization: Bearer <your-token>
```

**If you only have a human phrase** like “get user profile”:

```http
GET /v1/workspaces/acme/catalog/skills
Authorization: Bearer <your-token>
```

That returns skill ids, names, descriptions, tags. You match `user.profile.get` from text like “Get user profile”. Then list who offers it:

```http
GET /v1/workspaces/acme/catalog/skills/user.profile.get/agents
```

or:

```http
GET /v1/workspaces/acme/discover?tags=profile&passing=true
```

Fuzzy search over descriptions is your job (or a small local ranker). Cognit returns **structured cards**, not a chatbot answer.

**Read the card before you call.** Either the discover response already includes it, or:

```http
GET /v1/workspaces/acme/catalog/agents/user-directory/card
```

or the A2A well-known:

```http
GET /v1/workspaces/acme/agents/user-directory/.well-known/agent-card.json
```

You care about:

| Field | Why |
| --- | --- |
| `skills[].id` | Exact skill to request (`user.profile.get`) |
| `skills[].description` / `examples` | Confirm it is “get profile”, not “update profile” |
| `skills[].inputModes` | What to put in `parts` (`application/json` vs text) |
| `supportedInterfaces[0].url` | Gateway URL you POST to |
| `security` / card auth | How to authenticate the A2A call |

Example of what you learn:

```json
{
  "name": "User Directory",
  "skills": [
    {
      "id": "user.profile.get",
      "name": "Get user profile",
      "description": "Return name, email, plan, and renewal date for a user id",
      "tags": ["profile", "identity"],
      "inputModes": ["application/json"],
      "examples": ["{\"user_id\": \"u_123\"}"]
    }
  ],
  "supportedInterfaces": [
    {
      "url": "https://cognit.example.com/v1/workspaces/acme/a2a/agents/user-directory",
      "protocolBinding": "HTTP+JSON",
      "protocolVersion": "1.0",
      "tenant": "acme"
    }
  ]
}
```

If you want Cognit to pick a healthy instance for you:

```http
POST /v1/workspaces/acme/discover/resolve
```

```json
{
  "skill": "user.profile.get",
  "strategy": "least_outstanding_tasks"
}
```

---

## 4. Call that agent

Do not use the worker’s private address. Use the URL from the card (Cognit’s gateway), or skip the name and address the skill:

```http
POST /v1/workspaces/acme/a2a/skills/user.profile.get/message:send
Authorization: Bearer <your-token>
A2A-Version: 1.0
Content-Type: application/json
```

```json
{
  "message": {
    "role": "ROLE_USER",
    "parts": [
      { "data": { "user_id": "u_123" } }
    ]
  }
}
```

Or by agent name from the card:

```http
POST /v1/workspaces/acme/a2a/agents/user-directory/message:send
```

Then `GET .../tasks/{id}` or `:subscribe` until it completes. Fold the profile into your answer to the original user.

If this returns denied / not found, discovery worked but **intentions** block `support-assistant → user.profile.get`. That is policy, not a missing agent. Someone must allow:

```json
{
  "source": { "agent": "support-assistant" },
  "destination": { "agent": "user-directory", "skill": "user.profile.get" },
  "action": "allow"
}
```

---

## Loop you run on every user message

```text
user message
  → can I answer with my own skills?
      yes → do it, return A2A task/message
      no  → discover (skill id, tags, or catalog/skills)
          → read card (inputs, examples)
          → message:send through gateway
          → wait for task
          → answer the user
```

Cache cards by `agent` + `card.version` / `ETag`. Re-discover if the next call fails or the card expires. Watch the catalog if you want new profile agents to appear without a restart:

```http
GET /v1/workspaces/acme/watch/catalog?index=...
```

Short version: **register your card and keep your heartbeat check passing**. When you need “get user profile”, **search `catalog/skills` or `discover?skill=user.profile.get`**, **read that agent’s card**, then **`message:send` on the skill or agent gateway URL**.
