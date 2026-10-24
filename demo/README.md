# Demo agents

Four A2A agents for trying Sloth end to end: registration, the catalog, intentions, the gateway and traffic.

- `weather` on port 9201: `weather.current.get`, `weather.forecast.get`, `weather.alerts.list`
- `stock` on port 9202: `stock.quote.get`, `stock.history.get`, `stock.profile.get`
- `users` on port 9203: `user.profile.get`, `user.preferences.get`, protected by an API key
- `assistant` on port 9200: `assistant.ask`, plus a chat page that answers questions by calling the other agents through Sloth

`weather`, `stock` and `users` return fake data. `users` only answers calls that carry its `X-Users-Key` header, which only Sloth holds. `assistant` asks an LLM which skills to call, calls them through the Sloth gateway, then asks the LLM to rephrase the results into one answer.

## Run standalone

```sh
uv run uvicorn weather:app --port 9201
uv run uvicorn stock:app --port 9202
```

```sh
curl -s localhost:9201/.well-known/agent-card.json

curl -s localhost:9201/a2a/v1/message:send -H 'Content-Type: application/json' \
  -d '{"skill":"weather.forecast.get","message":{"role":"ROLE_USER","parts":[{"data":{"city":"Cairo","days":3}}]}}'
```

## Agent to agent through Sloth

### 1. Start Sloth with a fresh database

Recent migrations were edited in place, so start from an empty database, then create a workspace and an access key in the dashboard with `CAN_UPSERT_AGENT`, `CAN_REGISTER_INSTANCE`, `CAN_REPORT_CHECK`, `CAN_LIST_INTENTIONS` and `CAN_CREATE_INTENTION`.

```sh
export SLOTH_URL=http://localhost:8000
export SLOTH_WORKSPACE=<workspace-id>
export SLOTH_ACCESS_KEY=<access-key>
```

### 2. Start the agents

On boot each agent registers itself, defines a `heartbeat` TTL check and a `health` HTTP check that Sloth probes at `/healthz` every 10 seconds, registers an instance (`weather-9201`, `stock-9202`, `users-9203`, `assistant-9200`) at `/a2a/v1` and reports the check every 10 seconds. On shutdown it deregisters the instance.

```sh
uv run uvicorn weather:app --port 9201
uv run uvicorn stock:app --port 9202
uv run uvicorn users:app --port 9203
```

Or put your values in `run.local.sh` (git ignores it; `run.sh` sources it over its placeholder exports) and start all four agents at once. Each log line is prefixed with the agent's name, and Ctrl+C stops them all and deregisters their instances:

```sh
./run.sh
```

`users` also sets its protection in Sloth on boot: type `api_key`, header `X-Users-Key`, and the demo secret `users-demo-key-2026`. Sloth adds that header to every call it forwards to `users`, so a direct call fails:

```sh
curl -s -o /dev/null -w '%{http_code}\n' localhost:9203/a2a/v1/message:send -H 'Content-Type: application/json' \
  -d '{"skill":"user.profile.get","message":{"role":"ROLE_USER","parts":[{"data":{"user_id":"u_123"}}]}}'
# 401
```

The agent's Protection card in the dashboard shows it as active.

Set `AGENT_ADDRESS` if Sloth reaches the agents on an address other than `127.0.0.1`.

### 3. Give the assistant an identity and permission

In the dashboard:

- **Agents → assistant → Identities**: issue an `api_key` identity and copy its secret. The assistant agent appears once it has started once; start it, stop it, issue the identity, then start it again with the secret.
- **Intentions**: on boot the assistant creates `allow` intentions to `weather`, `stock` and `users`, skipping any destination that already has one, so a rule you turn off or change stays as you left it. Add a `deny` for a skill to see Sloth hide it.

### 4. Start the assistant and ask

```sh
export SLOTH_AGENT_SECRET=<identity secret>
export OPENROUTER_API_KEY=sk-or-...
uv run uvicorn assistant:app --port 9200
```

Open http://127.0.0.1:9200 and ask, for example, "Will it rain in Berlin in the next 3 days?" or "What plan is user u_123 on?". The second goes to the protected `users` agent; the assistant never sees its key. Under the answer, "How it got there" shows each step: the skills Sloth lets the assistant call, the calls the LLM chose, each gateway response with its status, and the model that wrote the answer. `ASSISTANT_MODEL` picks the model (default `openai/gpt-4o-mini`).

The assistant is an agent too, so other agents can ask it through Sloth:

```sh
curl -s $SLOTH_URL/api/v1/workspaces/$SLOTH_WORKSPACE/agents/assistant/message:send \
  -H 'Content-Type: application/json' \
  -H 'X-Source-Agent: <caller agent>' -H 'X-API-Key: <caller identity secret>' \
  -d '{"message":{"role":"ROLE_USER","parts":[{"text":"How is NVDA doing?"}]}}'
```

### 5. Watch it in the dashboard

- **Traffic**: one row per gateway call, with the caller, the instance that answered, the skill, the result and the latency.
- **Intentions**: turn the `assistant → weather` rule off and ask about the weather again. The skill disappears from the catalog, and a direct call is denied with 403 and recorded in Traffic as a deny.
- **Agents**: stop `weather` and its `health` check fails on the next probe while its heartbeat lapses within 30 seconds; Sloth stops routing to it and the assistant reports that no instance is available.
