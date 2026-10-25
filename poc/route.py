"""Turn a customer sentence into one reply.

For each question, Jev scores every skill in skills.json. Skills at 0.2 or
above are kept, a chat model fills their arguments from the text, and the
local agent returns fake data for each one. A second chat call writes a short
answer from those results.

Start the agent first, from this directory:

    uv run uvicorn agent:app --port 9100

Then:

    export OPENROUTER_API_KEY=sk-or-...
    uv run python route.py
"""

import json
import os
from pathlib import Path

import httpx

SKILLS = json.loads(Path(__file__).with_name("skills.json").read_text())
AGENT = "http://127.0.0.1:9100/a2a/v1/message:send"
QUESTIONS = [
    "What's on my profile (u_123) and when do I renew?",
    "When does the plan for u_123 renew?",
    "Show me the invoices for u_123.",
    "Send u_123 a password reset.",
    "Is the product down right now?",
    "Please add user joe@gmail.com to team sre-02 and remove user jane@acme.io from the team sre-03",
]


def choose(client, question):
    """Each skill is a separate yes/no, so one sentence can match more than one."""
    response = client.post(
        "https://openrouter.ai/api/alpha/decisions",
        headers={"Authorization": f"Bearer {os.environ['OPENROUTER_API_KEY']}"},
        json={
            "model": "typesafe/jev-1.13",
            "state": {"question": question},
            "questions": {
                skill["id"]: {
                    "type": "noul",
                    "instructions": f"Does the customer ask for this? {skill['description']}",
                    "criteria": {
                        "true": "The customer asks for this action.",
                        "false": "The customer does not ask for this action.",
                    },
                }
                for skill in SKILLS
            },
        },
    )
    response.raise_for_status()
    answers = response.json()["answers"]
    return [
        (skill, answers[skill["id"]]["noul"])
        for skill in SKILLS
        if answers[skill["id"]]["noul"] >= 0.2
    ]


def arguments(client, question, skill):
    schema = skill["arguments"]
    if not schema["properties"]:
        return {}
    response = client.post(
        "https://openrouter.ai/api/v1/chat/completions",
        headers={"Authorization": f"Bearer {os.environ['OPENROUTER_API_KEY']}"},
        json={
            "model": "deepseek/deepseek-v4-flash",
            "response_format": {"type": "json_object"},
            "messages": [
                {
                    "role": "user",
                    "content": (
                        "The text may contain several actions. "
                        f"Fill arguments only for {skill['id']}: {skill['description']} "
                        "Use only values written in the text that belong to this action. "
                        "Return a JSON object.\n"
                        f"Schema: {json.dumps(schema)}\n"
                        f"Text: {question}"
                    ),
                }
            ],
        },
    )
    response.raise_for_status()
    return json.loads(response.json()["choices"][0]["message"]["content"])


def call(client, skill_id, args):
    response = client.post(
        AGENT,
        json={
            "skill": skill_id,
            "message": {"role": "ROLE_USER", "parts": [{"data": args}]},
        },
    )
    response.raise_for_status()
    return response.json()["message"]["parts"][0]["data"]


def reply(client, question, results):
    response = client.post(
        "https://openrouter.ai/api/v1/chat/completions",
        headers={"Authorization": f"Bearer {os.environ['OPENROUTER_API_KEY']}"},
        json={
            "model": "deepseek/deepseek-v4-flash",
            "messages": [
                {
                    "role": "user",
                    "content": (
                        "Write a short answer to the customer using only these results.\n"
                        f"Question: {question}\n"
                        f"Results: {json.dumps(results)}"
                    ),
                }
            ],
        },
    )
    response.raise_for_status()
    return response.json()["choices"][0]["message"]["content"]


def main():
    with httpx.Client() as client:
        for question in QUESTIONS:
            print(question)
            results = []
            for skill, score in choose(client, question):
                args = arguments(client, question, skill)
                data = call(client, skill["id"], args)
                results.append({"skill": skill["id"], "data": data})
                print(skill["id"], score)
                print(data)
            print(reply(client, question, results))
            print()


if __name__ == "__main__":
    main()
