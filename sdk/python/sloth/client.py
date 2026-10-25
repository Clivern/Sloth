"""Async client for the Sloth REST API, covering what an agent needs to set itself up and call others."""

import httpx


class Client:
    """Talks to one Sloth workspace with an access key (setup) or an agent identity (gateway calls)."""

    def __init__(self, url, workspace, access_key, timeout=30):
        self.base = f"{url.rstrip('/')}/api/v1/workspaces/{workspace}"
        self.http = httpx.AsyncClient(headers={"X-Access-Key": access_key}, timeout=timeout)
        # The gateway forwards headers it does not know to the destination, so the access key must never ride along.
        self.agent_http = httpx.AsyncClient(timeout=timeout)

    async def close(self):
        await self.http.aclose()
        await self.agent_http.aclose()

    async def request(self, method, path, **kwargs):
        response = await self.http.request(method, f"{self.base}{path}", **kwargs)
        response.raise_for_status()

        return response.json() if response.content else None

    # Agents

    async def upsert_agent(self, name, version, card_path="/.well-known/agent-card.json"):
        return await self.request("PUT", f"/agents/{name}", json={"version": version, "cardPath": card_path})

    async def get_agent(self, name):
        return await self.request("GET", f"/agents/{name}")

    # Instances and checks

    async def register_instance(self, agent, instance, address, port, path="", scheme="http"):
        body = {"address": address, "port": port, "path": path, "scheme": scheme}
        return await self.request("PUT", f"/agents/{agent}/instances/{instance}", json=body)

    async def deregister_instance(self, agent, instance):
        return await self.request("DELETE", f"/agents/{agent}/instances/{instance}")

    async def upsert_check(self, agent, check_id, **check):
        return await self.request("PUT", f"/agents/{agent}/checks/{check_id}", json=check)

    async def pass_check(self, agent, instance, check_id, output=""):
        return await self.request("POST", f"/agents/{agent}/instances/{instance}/checks/{check_id}/pass", json={"output": output})

    # Protection and identities

    async def save_protection(self, agent, type="api_key", header="", username="", secret=""):
        body = {"type": type, "header": header, "username": username, "secret": secret}
        return await self.request("PUT", f"/agents/{agent}/protection", json={k: v for k, v in body.items() if v})

    async def list_identities(self, agent):
        return (await self.request("GET", f"/agents/{agent}/identities"))["identities"]

    async def create_identity(self, agent, name, type="api_key", secret=""):
        body = {"name": name, "type": type, "secret": secret}
        return await self.request("POST", f"/agents/{agent}/identities", json={k: v for k, v in body.items() if v})

    async def delete_identity(self, agent, identity_id):
        return await self.request("DELETE", f"/agents/{agent}/identities/{identity_id}")

    # Intentions

    async def list_intentions(self):
        intentions, offset = [], 0

        while True:
            page = await self.request("GET", "/intentions", params={"limit": 100, "offset": offset})
            intentions += page["intentions"]
            offset += len(page["intentions"])

            if not page["intentions"] or offset >= page["_meta"]["total"]:
                return intentions

    async def create_intention(self, source, destination, action="allow", skill=""):
        target = {"agent": destination, **({"skill": skill} if skill else {})}
        body = {"source": {"agent": source}, "destination": target, "action": action}
        return await self.request("PUT", "/intentions", json=body)

    # Gateway: these authenticate as an agent identity, not with the access key

    async def send_message(self, source, secret, agent, message, skill=""):
        return await self.gateway(source, secret, f"/agents/{agent}/message:send", message, skill)

    async def send_skill_message(self, source, secret, skill, message):
        return await self.gateway(source, secret, f"/agents/skills/{skill}/message:send", message, skill)

    async def gateway(self, source, secret, path, message, skill):
        headers = {"X-Source-Agent": source, "X-API-Key": secret}
        body = {"message": message, **({"skill": skill} if skill else {})}

        response = await self.agent_http.post(f"{self.base}{path}", json=body, headers=headers)
        response.raise_for_status()

        return response.json()
