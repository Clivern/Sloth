"""Register a demo agent with Sloth and keep its heartbeat check passing.

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


async def heartbeat(client, base, instance):
    while True:
        await client.post(f"{base}/instances/{instance}/checks/heartbeat/pass")
        await asyncio.sleep(TTL / 3)


def lifespan(name, version, port):
    @asynccontextmanager
    async def run(app):
        if not URL:
            yield
            return

        base = f"{URL}/api/v1/workspaces/{WORKSPACE}/agents/{name}"
        instance = f"{name}-{port}"

        async with httpx.AsyncClient(headers={"X-Access-Key": ACCESS_KEY}) as client:
            await client.put(base, json={"version": version, "cardPath": "/.well-known/agent-card.json"})
            await client.put(f"{base}/checks/heartbeat", json={"type": "ttl", "ttl": TTL})
            await client.put(f"{base}/instances/{instance}", json={"address": ADDRESS, "port": port})

            task = asyncio.create_task(heartbeat(client, base, instance))
            yield
            task.cancel()

            await client.delete(f"{base}/instances/{instance}")

    return run
