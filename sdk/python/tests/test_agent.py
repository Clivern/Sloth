import json

import httpx
from starlette.testclient import TestClient

from sloth import Agent, Client
from sloth.agent import PROTECTION_HEADER


def make_agent(**kwargs):
    agent = Agent("users", version="1.2.0", description="User profiles", port=9203, **kwargs)

    @agent.skill("user.profile.get", tags=["users"], examples=['{"user_id": "u_1"}'])
    def profile(user_id="u_123"):
        """Return a user's profile."""
        return {"user_id": user_id, "name": "Ada"}

    @agent.skill("user.greet")
    async def greet(text):
        return f"hello {text}"

    return agent


def send(skill, part):
    return {"skill": skill, "message": {"role": "ROLE_USER", "parts": [part]}}


def test_card_lists_skills_from_functions():
    card = make_agent().card()

    assert card["name"] == "users"
    assert card["version"] == "1.2.0"
    assert card["supportedInterfaces"][0]["url"] == "http://127.0.0.1:9203/a2a/v1"
    assert [skill["id"] for skill in card["skills"]] == ["user.profile.get", "user.greet"]
    assert card["skills"][0]["description"] == "Return a user's profile."
    assert card["skills"][0]["name"] == "Profile"
    assert card["skills"][0]["tags"] == ["users"]


def test_message_dispatches_data_and_text_parts():
    with TestClient(make_agent().app) as http:
        reply = http.post("/a2a/v1/message:send", json=send("user.profile.get", {"data": {"user_id": "u_9"}}))
        assert reply.json()["message"]["parts"][0]["data"] == {"user_id": "u_9", "name": "Ada"}

        reply = http.post("/a2a/v1/message:send", json=send("user.greet", {"text": "sloth"}))
        assert reply.json()["message"]["parts"][0]["text"] == "hello sloth"


def test_message_without_skill_uses_the_first_one():
    with TestClient(make_agent().app) as http:
        reply = http.post("/a2a/v1/message:send", json={"message": {"parts": [{"data": {}}]}})

    assert reply.json()["message"]["parts"][0]["data"]["user_id"] == "u_123"


def test_unknown_skill_is_not_found():
    with TestClient(make_agent().app) as http:
        reply = http.post("/a2a/v1/message:send", json=send("nope", {"data": {}}))

    assert reply.status_code == 404


def test_protection_rejects_calls_that_bypass_sloth():
    agent = make_agent(url="http://sloth", workspace="ws", access_key="key")
    agent.client = object()

    # No context manager, so the lifespan does not try to reach Sloth
    http = TestClient(agent.app)
    assert http.post("/a2a/v1/message:send", json=send("user.greet", {"text": "x"})).status_code == 401

    headers = {PROTECTION_HEADER: agent.secret("protection")}
    assert http.post("/a2a/v1/message:send", json=send("user.greet", {"text": "x"}), headers=headers).status_code == 200


def test_secrets_are_stable_and_distinct():
    a = make_agent(workspace="ws", access_key="key")
    b = make_agent(workspace="ws", access_key="key")

    assert a.secret("protection") == b.secret("protection")
    assert a.secret("protection") != a.secret("identity")
    assert a.secret("protection") != make_agent(workspace="ws", access_key="other").secret("protection")


async def test_setup_configures_sloth():
    calls = []

    def handler(request):
        body = json.loads(request.content) if request.content else None
        calls.append((request.method, request.url.path.removeprefix("/api/v1/workspaces/ws"), body))

        if request.url.path.endswith("/identities") and request.method == "GET":
            return httpx.Response(200, json={"identities": [{"id": "old", "name": "sloth-sdk"}]})
        if request.url.path.endswith("/intentions") and request.method == "GET":
            intentions = [{"source": "users", "destination": "billing", "skill": ""}]
            return httpx.Response(200, json={"intentions": intentions, "_meta": {"total": 1}})

        return httpx.Response(200, json={})

    agent = make_agent(url="http://sloth", workspace="ws", access_key="key", can_call=["billing", "orders"], allow_calls_from=["desk"])
    agent.client = Client("http://sloth", "ws", "key")
    agent.client.http = httpx.AsyncClient(headers={"X-Access-Key": "key"}, transport=httpx.MockTransport(handler))

    await agent.setup()

    paths = [(method, path) for method, path, _ in calls]
    assert ("PUT", "/agents/users") in paths
    assert ("PUT", "/agents/users/protection") in paths
    assert ("DELETE", "/agents/users/identities/old") in paths
    assert ("POST", "/agents/users/identities") in paths
    assert ("PUT", "/agents/users/checks/heartbeat") in paths

    protection = next(body for method, path, body in calls if path == "/agents/users/protection")
    assert protection == {"type": "api_key", "header": PROTECTION_HEADER, "secret": agent.secret("protection")}

    intentions = [body for method, path, body in calls if method == "PUT" and path == "/intentions"]
    assert [(i["source"]["agent"], i["destination"]["agent"]) for i in intentions] == [("users", "orders"), ("desk", "users")]


async def test_gateway_calls_never_send_the_access_key():
    seen = {}

    def handler(request):
        seen.update(request.headers)
        return httpx.Response(200, json={"message": {"parts": [{"data": {"ok": True}}]}})

    agent = make_agent(url="http://sloth", workspace="ws", access_key="key")
    agent.client = Client("http://sloth", "ws", "key")
    agent.client.agent_http = httpx.AsyncClient(transport=httpx.MockTransport(handler))

    assert await agent.call("orders", {"id": 1}) == {"ok": True}
    assert "x-access-key" not in seen
    assert seen["x-source-agent"] == "users"
    assert seen["x-api-key"] == agent.secret("identity")
