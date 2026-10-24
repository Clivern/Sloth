"""Register a demo agent with Sloth, keep its heartbeat check passing and let Sloth probe /healthz.

Set SLOTH_URL, SLOTH_WORKSPACE and SLOTH_ACCESS_KEY to register. Without
SLOTH_URL the agent only serves its card and message:send locally.
"""

import asyncio
import os
from contextlib import asynccontextmanager

import httpx

URL = os.environ.get("SLOTH_URL", "")
WORKSPACE = os.environ.get("SLOTH_WORKSPACE", "")
ACCESS_KEY = os.environ.get("SLOTH_ACCESS_KEY", "")
ADDRESS = os.environ.get("AGENT_ADDRESS", "127.0.0.1")
TTL = 30


async def heartbeat(client, base, instance, port):
    # Sloth fetches the card when the instance registers, so wait until the agent is serving it.
    while True:
        try:
            await client.get(f"http://127.0.0.1:{port}/.well-known/agent-card.json")
            break
        except httpx.TransportError:
            await asyncio.sleep(0.2)

    await client.put(f"{base}/instances/{instance}", json={"address": ADDRESS, "port": port, "path": "/a2a/v1"})

    while True:
        await client.post(f"{base}/instances/{instance}/checks/heartbeat/pass")
        await asyncio.sleep(TTL / 3)


async def allow(client, name, calls):
    """Allow name to call each agent in calls, skipping pairs that already have an intention."""
    response = await client.get(f"{URL}/api/v1/workspaces/{WORKSPACE}/intentions", params={"limit": 100})
    existing = {item["destination"] for item in response.json()["intentions"] if item["source"] == name}

    for destination in calls:
        if destination in existing:
            continue

        await client.put(
            f"{URL}/api/v1/workspaces/{WORKSPACE}/intentions",
            json={"source": {"agent": name}, "destination": {"agent": destination}, "action": "allow"},
        )


def lifespan(name, version, port, protection=None, calls=None):
    """Register the agent on boot; protection, when set, is the credential Sloth must send the agent,
    and calls lists the agents it is allowed to call."""

    @asynccontextmanager
    async def run(app):
        if not URL:
            yield
            return

        base = f"{URL}/api/v1/workspaces/{WORKSPACE}/agents/{name}"
        instance = f"{name}-{port}"

        async with httpx.AsyncClient(headers={"X-Access-Key": ACCESS_KEY}) as client:
            await client.put(base, json={"version": version, "cardPath": "/.well-known/agent-card.json"})

            if protection:
                await client.put(f"{base}/protection", json=protection)

            await client.put(f"{base}/checks/heartbeat", json={"type": "ttl", "ttl": TTL})
            await client.put(f"{base}/checks/health", json={"type": "http", "path": "/healthz", "interval": 10})

            if calls:
                await allow(client, name, calls)

            task = asyncio.create_task(heartbeat(client, base, instance, port))
            yield
            task.cancel()

            await client.delete(f"{base}/instances/{instance}")

    return run
