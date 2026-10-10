"""Cognit stores Agent Cards. The LLM only sees skills-as-tools."""

import json

# What Cognit stores per registered agent: A2A card + this instance.
# PUT /v1/workspaces/acme/catalog/agents/{agent}
store = {
    "user-directory": {
        "card": {
            "name": "User Directory",
            "description": "Looks up people and their account profile",
            "version": "1.2.0",
            "capabilities": {"streaming": False},
            "defaultInputModes": ["application/json"],
            "defaultOutputModes": ["application/json"],
            "skills": [
                {
                    "id": "user.profile.get",
                    "name": "Get user profile",
                    "description": "Get name, email, plan, renewal date for a user id",
                    "tags": ["profile", "identity"],
                    "examples": ['{"user_id": "u_123"}'],
                    "inputModes": ["application/json"],
                    "inputSchema": {
                        "type": "object",
                        "properties": {"user_id": {"type": "string"}},
                        "required": ["user_id"],
                    },
                }
            ],
            "supportedInterfaces": [
                {
                    "url": "http://10.0.4.11:8080/a2a/v1",
                    "protocolBinding": "HTTP+JSON",
                    "protocolVersion": "1.0",
                    "tenant": "acme",
                }
            ],
        },
        "instance": {
            "id": "user-directory-1",
            "address": "10.0.4.11",
            "port": 8080,
            "health": "passing",
        },
    },
    "billing": {
        "card": {
            "name": "Billing",
            "description": "Plans, renewals, and invoices",
            "version": "3.0.1",
            "capabilities": {"streaming": True},
            "defaultInputModes": ["application/json"],
            "defaultOutputModes": ["application/json"],
            "skills": [
                {
                    "id": "billing.renewal.get",
                    "name": "Get renewal",
                    "description": "Get plan renewal date and price for a user id",
                    "tags": ["billing", "renewal"],
                    "examples": ['{"user_id": "u_123"}'],
                    "inputModes": ["application/json"],
                    "inputSchema": {
                        "type": "object",
                        "properties": {"user_id": {"type": "string"}},
                        "required": ["user_id"],
                    },
                },
                {
                    "id": "billing.invoice.list",
                    "name": "List invoices",
                    "description": "List invoices for a user id",
                    "tags": ["billing", "invoices"],
                    "examples": ['{"user_id": "u_123"}'],
                    "inputModes": ["application/json"],
                    "inputSchema": {
                        "type": "object",
                        "properties": {"user_id": {"type": "string"}},
                        "required": ["user_id"],
                    },
                },
            ],
            "supportedInterfaces": [
                {
                    "url": "http://10.0.4.22:8080/a2a/v1",
                    "protocolBinding": "HTTP+JSON",
                    "protocolVersion": "1.0",
                    "tenant": "acme",
                }
            ],
        },
        "instance": {
            "id": "billing-2",
            "address": "10.0.4.22",
            "port": 8080,
            "health": "passing",
        },
    },
}

query = """
hey so i was talking to jane yesterday about our seats,
anyway can you pull up whatever we have on me (i think my id is u_123)
and also when do i renew?? we might add 3 more people next month
oh and don't bother with invoices
"""


def cards_to_tools(store):
    """Flatten every skill on every card. The model never sees agent names."""
    tools = []
    for record in store.values():
        if record["instance"]["health"] != "passing":
            continue
        for skill in record["card"]["skills"]:
            tools.append(
                {
                    "type": "function",
                    "function": {
                        "name": skill["id"],
                        "description": skill["description"],
                        "parameters": skill["inputSchema"],
                    },
                }
            )
    return tools


def find_agent(store, skill_id):
    for agent_id, record in store.items():
        for skill in record["card"]["skills"]:
            if skill["id"] == skill_id:
                return agent_id
    raise KeyError(skill_id)


def fake_llm(query, tools):
    return [
        {"name": "user.profile.get", "arguments": {"user_id": "u_123"}},
        {"name": "billing.renewal.get", "arguments": {"user_id": "u_123"}},
    ]


def execute(tool_call, store):
    skill_id = tool_call["name"]
    return {
        "tool": skill_id,
        "args": tool_call["arguments"],
        "send": {
            "method": "POST",
            "url": f"/v1/workspaces/acme/a2a/skills/{skill_id}/message:send",
            "body": {
                "message": {
                    "role": "ROLE_USER",
                    "parts": [{"data": tool_call["arguments"]}],
                }
            },
        },
        "cognit_routes_to": find_agent(store, skill_id),
    }


tools = cards_to_tools(store)
calls = fake_llm(query, tools)

print("1. Cognit store (agent cards)\n")
print(json.dumps(store, indent=2))
print("\n2. tools the LLM sees (skills only)\n")
print(json.dumps(tools, indent=2))
print("\n3. LLM tool calls\n")
print(json.dumps(calls, indent=2))
print("\n4. execute — Cognit maps skill → agent\n")
print(json.dumps([execute(c, store) for c in calls], indent=2))
