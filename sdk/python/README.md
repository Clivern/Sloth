# Sloth Python SDK

Build A2A agents as plain Python functions. With a workspace access key, `run()` sets the agent up in Sloth and keeps it registered:

- registers the agent and serves its card, built from your functions
- turns on protection, so the agent only accepts calls forwarded by Sloth
- issues the agent an identity, so it can call other agents through the gateway
- creates the intentions you declare with `can_call` and `allow_calls_from`
- adds a heartbeat check and an HTTP check on `/healthz`
- registers the instance, keeps the heartbeat passing and deregisters on shutdown

You host the agent wherever you like. Sloth only routes, secures and watches it.

## Install

```bash
pip install ./sdk/python
```

## A first agent

```python
from sloth import Agent

agent = Agent(
    name="users",
    version="1.0.0",
    description="User profiles by user id",
    port=9203,
    allow_calls_from=["assistant"],
)


@agent.skill("user.profile.get", tags=["users"], examples=['{"user_id": "u_123"}'])
def profile(user_id="u_123"):
    """Return a user's name and plan."""
    return {"user_id": user_id, "name": "Ada Lovelace", "plan": "pro"}


agent.run()
```

```bash
export SLOTH_URL=http://localhost:8000
export SLOTH_WORKSPACE=<workspace-id>
export SLOTH_ACCESS_KEY=<access-key>
python users.py
```

Without these variables the agent still serves its card and skills locally, which is handy while building.

## Skills

- The decorator's first argument is the skill id. The function's docstring becomes the description, and its name (or `name=`) the display name.
- A message with a data part calls the function with the data as keyword arguments. A text part calls it with `text=`.
- Return a `dict` or `list` to reply with data, or a `str` to reply with text.
- Functions can be `async`.
- A message that names no skill goes to the first one.

## Calling other agents

Calls go through the Sloth gateway as this agent, so intentions apply and every call shows up in Traffic.

```python
agent = Agent(name="assistant", can_call=["users", "weather"])


@agent.skill("assistant.ask")
async def ask(text):
    user = await agent.call("users", {"user_id": "u_123"}, skill="user.profile.get")
    weather = await agent.call_skill("weather.current", {"city": "Cairo"})
    return f"{user['name']}, it is {weather['condition']} in Cairo."
```

- `call(agent, data=None, text=None, skill="")` calls a named agent.
- `call_skill(skill, data=None, text=None)` lets Sloth pick an allowed, healthy agent that offers the skill.

## Settings

Pass them to `Agent(...)` or set them in the environment:

- `url` / `SLOTH_URL`: the Sloth server
- `workspace` / `SLOTH_WORKSPACE`: the workspace id
- `access_key` / `SLOTH_ACCESS_KEY`: a workspace access key that can upsert agents, register instances, report checks and create intentions
- `address` / `AGENT_ADDRESS`: the address Sloth uses to reach this agent (default `127.0.0.1`)
- `port`: the port to serve on (default `8000`)
- `instance`: the instance id (default `<name>-<hostname>-<port>`)
- `protect`: turn protection on (default `True`)
- `can_call`, `allow_calls_from`: agents to create allow intentions for
- `heartbeat_ttl`: seconds for the heartbeat check (default `30`)

## How secrets work

The protection secret and the identity secret are derived from the access key with HMAC, per workspace and agent. Every instance and every restart computes the same values, so nothing has to be stored or shared, and scaling out just means starting more processes. Rotating the access key rotates both secrets on the next start.

## Development

```bash
cd sdk/python
uv sync
uv run pytest
```
