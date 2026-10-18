# Demo agents

Two A2A agents with fake data, for trying Sloth registration, discovery and the catalog.

| Agent | Port | Skills |
| --- | --- | --- |
| `weather` | 9201 | `weather.current.get`, `weather.forecast.get`, `weather.alerts.list` |
| `stock` | 9202 | `stock.quote.get`, `stock.history.get`, `stock.profile.get` |

## Run standalone

```sh
uv run uvicorn weather:app --port 9201
uv run uvicorn stock:app --port 9202
```

```sh
curl -s localhost:9201/.well-known/agent-card.json

curl -s localhost:9201/a2a/v1/message:send -H 'Content-Type: application/json' \
  -d '{"skill":"weather.forecast.get","message":{"role":"ROLE_USER","parts":[{"data":{"city":"Cairo","days":3}}]}}'

curl -s localhost:9202/a2a/v1/message:send -H 'Content-Type: application/json' \
  -d '{"skill":"stock.quote.get","message":{"role":"ROLE_USER","parts":[{"data":{"symbol":"NVDA"}}]}}'
```

## Register with Sloth

On boot each agent registers itself, defines a `heartbeat` TTL check, registers an instance
(`weather-9201`, `stock-9202`) and reports the check every 10 seconds. On shutdown it deregisters
the instance.

```sh
export SLOTH_URL=http://localhost:8000
export SLOTH_WORKSPACE=<workspace-id>
export SLOTH_ACCESS_KEY=<access-key>

uv run uvicorn weather:app --port 9201
uv run uvicorn stock:app --port 9202
```

Set `AGENT_ADDRESS` if Sloth reaches the agents on an address other than `127.0.0.1`.
