"""Agent: declare skills as Python functions, then run() sets the agent up in Sloth and serves it over A2A."""

import asyncio
import hashlib
import hmac
import inspect
import os
import secrets
import socket
from contextlib import asynccontextmanager
from dataclasses import dataclass, field

import httpx
import uvicorn
from starlette.applications import Starlette
from starlette.requests import Request
from starlette.responses import JSONResponse
from starlette.routing import Route

from sloth.client import Client

A2A_PATH = "/a2a/v1"
CARD_PATH = "/.well-known/agent-card.json"
IDENTITY_NAME = "sloth-sdk"
PROTECTION_HEADER = "X-Sloth-Protection"


@dataclass
class Skill:
    id: str
    func: callable
    name: str
    description: str
    tags: list = field(default_factory=list)
    examples: list = field(default_factory=list)

    def card(self):
        return {
            "id": self.id,
            "name": self.name,
            "description": self.description,
            "tags": self.tags,
            "examples": self.examples,
            "inputModes": ["application/json", "text/plain"],
            "outputModes": ["application/json", "text/plain"],
        }


class Agent:
    """An A2A agent that registers itself, its protection, its identity and its intentions with Sloth.

    Settings fall back to SLOTH_URL, SLOTH_WORKSPACE, SLOTH_ACCESS_KEY and AGENT_ADDRESS.
    """

    def __init__(
        self,
        name,
        version="1.0.0",
        description="",
        provider="",
        url=None,
        workspace=None,
        access_key=None,
        address=None,
        port=8000,
        instance=None,
        protect=True,
        can_call=(),
        allow_calls_from=(),
        heartbeat_ttl=30,
    ):
        self.name = name
        self.version = version
        self.description = description
        self.provider = provider
        self.url = url or os.environ.get("SLOTH_URL", "")
        self.workspace = workspace or os.environ.get("SLOTH_WORKSPACE", "")
        self.access_key = access_key or os.environ.get("SLOTH_ACCESS_KEY", "")
        self.address = address or os.environ.get("AGENT_ADDRESS", "127.0.0.1")
        self.port = port
        self.instance = instance or f"{name}-{socket.gethostname()}-{port}"
        self.protect = protect
        self.can_call = list(can_call)
        self.allow_calls_from = list(allow_calls_from)
        self.heartbeat_ttl = heartbeat_ttl
        self.skills = {}
        self.client = None
        self.app = Starlette(
            routes=[
                Route(CARD_PATH, self.card_action),
                Route("/healthz", self.health_action),
                Route(f"{A2A_PATH}/message:send", self.message_action, methods=["POST"]),
            ],
            lifespan=self.lifespan,
        )

    def skill(self, id, name=None, tags=(), examples=()):
        """Register a function as a skill; its docstring becomes the description and its arguments the input."""

        def register(func):
            self.skills[id] = Skill(
                id=id,
                func=func,
                name=name or func.__name__.replace("_", " ").capitalize(),
                description=inspect.cleandoc(func.__doc__ or ""),
                tags=list(tags),
                examples=list(examples),
            )
            return func

        return register

    def secret(self, purpose):
        """Derive a stable secret from the access key, so every instance and restart agrees on it without storage."""
        message = f"{self.workspace}:{self.name}:{purpose}".encode()
        return hmac.new(self.access_key.encode(), message, hashlib.sha256).hexdigest()

    def card(self):
        return {
            "name": self.name,
            "description": self.description,
            "version": self.version,
            **({"provider": {"organization": self.provider}} if self.provider else {}),
            "supportedInterfaces": [
                {
                    "url": f"http://{self.address}:{self.port}{A2A_PATH}",
                    "protocolBinding": "HTTP+JSON",
                    "protocolVersion": "1.0",
                }
            ],
            "defaultInputModes": ["application/json", "text/plain"],
            "defaultOutputModes": ["application/json", "text/plain"],
            "skills": [skill.card() for skill in self.skills.values()],
        }

    # Calling other agents through the gateway

    async def call(self, agent, data=None, text=None, skill=""):
        """Send a message to a named agent through Sloth and return the reply's data or text."""
        reply = await self.client.send_message(self.name, self.secret("identity"), agent, message(data, text), skill)
        return content(reply)

    async def call_skill(self, skill, data=None, text=None):
        """Send a message to any agent offering a skill, letting Sloth pick an allowed, healthy one."""
        reply = await self.client.send_skill_message(self.name, self.secret("identity"), skill, message(data, text))
        return content(reply)

    # HTTP handlers

    async def card_action(self, request: Request):
        return JSONResponse(self.card())

    async def health_action(self, request: Request):
        return JSONResponse({"status": "ok"})

    async def message_action(self, request: Request):
        # With protection on, Sloth adds this header to every call it forwards, so a call without it bypassed the gateway.
        if self.protect and self.client:
            given = request.headers.get(PROTECTION_HEADER, "")
            if not secrets.compare_digest(given, self.secret("protection")):
                return JSONResponse({"error": "call did not come through Sloth"}, status_code=401)

        body = await request.json()

        skill = self.skills.get(body.get("skill") or next(iter(self.skills)))
        if skill is None:
            return JSONResponse({"error": f"unknown skill {body.get('skill')}"}, status_code=404)

        part = body["message"]["parts"][0]
        args = part["data"] if "data" in part else {"text": part.get("text", "")}

        result = skill.func(**args)
        if inspect.isawaitable(result):
            result = await result

        return JSONResponse({"message": reply_message(result)})

    # Lifecycle

    async def setup(self):
        """Create or update everything the agent needs in Sloth. Safe to run on every start."""
        await self.client.upsert_agent(self.name, self.version, CARD_PATH)

        if self.protect:
            await self.client.save_protection(self.name, "api_key", header=PROTECTION_HEADER, secret=self.secret("protection"))

        # The identity secret is derived, so recreating it keeps every instance's copy valid.
        for identity in await self.client.list_identities(self.name):
            if identity["name"] == IDENTITY_NAME:
                await self.client.delete_identity(self.name, identity["id"])
        await self.client.create_identity(self.name, IDENTITY_NAME, secret=self.secret("identity"))

        existing = {(i["source"], i["destination"]) for i in await self.client.list_intentions() if not i.get("skill")}
        wanted = [(self.name, other) for other in self.can_call] + [(other, self.name) for other in self.allow_calls_from]
        for source, destination in wanted:
            if (source, destination) not in existing:
                await self.client.create_intention(source, destination)

        await self.client.upsert_check(self.name, "heartbeat", type="ttl", ttl=self.heartbeat_ttl)
        await self.client.upsert_check(self.name, "health", type="http", path="/healthz", interval=10)

    async def heartbeat(self):
        # Sloth fetches the card when the instance registers, so wait until this server is answering.
        async with httpx.AsyncClient() as local:
            while True:
                try:
                    await local.get(f"http://127.0.0.1:{self.port}{CARD_PATH}")
                    break
                except httpx.TransportError:
                    await asyncio.sleep(0.2)

        await self.client.register_instance(self.name, self.instance, self.address, self.port, A2A_PATH)

        while True:
            await self.client.pass_check(self.name, self.instance, "heartbeat")
            await asyncio.sleep(self.heartbeat_ttl / 3)

    @asynccontextmanager
    async def lifespan(self, app):
        # Without Sloth settings the agent still serves locally, which is handy while building skills.
        if not (self.url and self.workspace and self.access_key):
            yield
            return

        self.client = Client(self.url, self.workspace, self.access_key)
        await self.setup()
        task = asyncio.create_task(self.heartbeat())

        yield

        task.cancel()
        await self.client.deregister_instance(self.name, self.instance)
        await self.client.close()

    def run(self, host="0.0.0.0", port=None):
        """Serve the agent; on start it sets itself up in Sloth, on stop it deregisters its instance."""
        self.port = port or self.port
        uvicorn.run(self.app, host=host, port=self.port)


def message(data=None, text=None):
    part = {"data": data} if data is not None else {"text": text or ""}
    return {"role": "ROLE_USER", "parts": [part]}


def reply_message(result):
    part = {"text": result} if isinstance(result, str) else {"data": result}
    return {"role": "ROLE_AGENT", "parts": [part]}


def content(reply):
    part = reply["message"]["parts"][0]
    return part["data"] if "data" in part else part.get("text", "")
