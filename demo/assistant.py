"""Assistant agent with a chat page. uv run uvicorn assistant:app --port 9200

It answers a question by calling other agents through the Sloth gateway:

1. list the skills it may call: GET /agents/catalog/skills
2. ask an LLM which skills to call and with what data
3. call each one through Sloth: POST /agents/skills/{skill}/message:send
4. ask the LLM to turn the results into one short answer

Open http://127.0.0.1:9200 for the chat page. Besides the registration variables
in util.py it needs SLOTH_AGENT_SECRET, the secret of an api_key identity issued
to the assistant agent, and OPENROUTER_API_KEY. ASSISTANT_MODEL picks the model.
"""

import json
import os
from pathlib import Path

import httpx
from fastapi import FastAPI
from fastapi.responses import HTMLResponse

from util import URL, WORKSPACE, lifespan

NAME = "assistant"
VERSION = "1.0.0"
PORT = int(os.environ.get("PORT", "9200"))
SECRET = os.environ.get("SLOTH_AGENT_SECRET", "")
OPENROUTER_API_KEY = os.environ.get("OPENROUTER_API_KEY", "")
MODEL = os.environ.get("ASSISTANT_MODEL", "deepseek/deepseek-v4-flash")
GATEWAY = f"{URL}/api/v1/workspaces/{WORKSPACE}/agents"
MAX_CALLS = 3
PAGE = (Path(__file__).parent / "index.html").read_text()

PLAN = """You route a user's question to agent skills.
You get the question and the skills you may call, each with an id, a name and a
description that ends with an example of its data. Reply with JSON only:
{"calls": [{"skill": "<skill id>", "data": {...}}]}
Use only the listed skill ids, at most 3 calls, and fill data from the question
using the same keys as the example. Reply {"calls": []} when no skill helps."""

ANSWER = """You answer the user's question in two or three friendly sentences.
Base live data such as weather, stock prices and user details only on the agent
results you are given, and never invent it. Each result names the skill it came
from and its HTTP status; say plainly when a call failed or was not allowed.
When there are no results because no skill fits, answer general knowledge
questions yourself and say which kinds of questions your agents can help with."""

app = FastAPI(lifespan=lifespan(NAME, VERSION, PORT, calls=["weather", "stock", "users"]))


async def chat(client, system, content, as_json=False):
    response = await client.post(
        "https://openrouter.ai/api/v1/chat/completions",
        headers={"Authorization": f"Bearer {OPENROUTER_API_KEY}"},
        json={
            "model": MODEL,
            "messages": [
                {"role": "system", "content": system},
                {"role": "user", "content": json.dumps(content)},
            ],
            **({"response_format": {"type": "json_object"}} if as_json else {}),
        },
    )
    response.raise_for_status()

    return response.json()["choices"][0]["message"]["content"]


async def ask(question):
    """Answer a question through Sloth and return the answer with every step taken."""
    steps = []
    sloth = {"X-Source-Agent": NAME, "X-API-Key": SECRET}

    async with httpx.AsyncClient(timeout=60) as client:
        # 1. Sloth only lists skills an intention lets the assistant call.
        response = await client.get(f"{GATEWAY}/catalog/skills", headers=sloth)
        skills = response.json().get("skills", []) if response.is_success else []
        steps.append({"step": "catalog", "status": response.status_code, "skills": [skill["id"] for skill in skills]})

        if response.status_code == 401:
            return {
                "answer": "Sloth doesn't recognize me yet. Issue an api_key identity to the assistant and set SLOTH_AGENT_SECRET.",
                "steps": steps,
            }

        if not skills:
            return {
                "answer": "I can't call any agent yet. Allow the assistant to call an agent under Intentions in Sloth.",
                "steps": steps,
            }

        # 2. The LLM picks skills and their data from the question.
        choices = [{"id": s["id"], "name": s["name"], "description": s["description"]} for s in skills]
        plan = json.loads(await chat(client, PLAN, {"question": question, "skills": choices}, as_json=True))
        calls = [call for call in plan.get("calls", []) if call.get("skill") in {s["id"] for s in skills}][:MAX_CALLS]
        steps.append({"step": "plan", "calls": calls})

        # 3. Each call goes through Sloth, which checks intentions and picks a live instance.
        results = []
        for call in calls:
            response = await client.post(
                f"{GATEWAY}/skills/{call['skill']}/message:send",
                headers=sloth,
                json={
                    "skill": call["skill"],
                    "message": {"role": "ROLE_USER", "parts": [{"data": call.get("data", {})}]},
                },
            )
            result = response.json()
            results.append({"skill": call["skill"], "status": response.status_code, "result": result})
            steps.append({"step": "call", "skill": call["skill"], "data": call.get("data", {}), "status": response.status_code, "result": result})

        # 4. The LLM turns the raw results into one answer.
        answer = await chat(client, ANSWER, {"question": question, "results": results})
        steps.append({"step": "answer", "model": MODEL})

    return {"answer": answer, "steps": steps}


@app.get("/", response_class=HTMLResponse)
def page():
    return PAGE


@app.get("/info")
async def info():
    """The skills Sloth lets the assistant call, for the page's setup card."""
    async with httpx.AsyncClient(timeout=10) as client:
        response = await client.get(f"{GATEWAY}/catalog/skills", headers={"X-Source-Agent": NAME, "X-API-Key": SECRET})

    skills = response.json().get("skills", []) if response.is_success else []

    return {"status": response.status_code, "skills": [skill["id"] for skill in skills]}


@app.post("/ask")
async def ask_action(body: dict):
    return await ask(body["question"])


@app.get("/healthz")
def healthz():
    return {"status": "ok"}


@app.get("/.well-known/agent-card.json")
def agent_card():
    return {
        "name": "Assistant",
        "description": "Answers questions by asking other agents through Sloth",
        "version": VERSION,
        "supportedInterfaces": [
            {
                "url": f"http://127.0.0.1:{PORT}/a2a/v1",
                "protocolBinding": "HTTP+JSON",
                "protocolVersion": "1.0",
            }
        ],
        "defaultInputModes": ["text/plain"],
        "defaultOutputModes": ["text/plain"],
        "skills": [
            {
                "id": "assistant.ask",
                "name": "Ask the assistant",
                "description": "Answer a question in plain language using the agents the assistant may call",
                "tags": ["assistant"],
                "examples": ["What's the weather in Cairo and how is NVDA doing?"],
            }
        ],
    }


@app.post("/a2a/v1/message:send")
async def message_send(body: dict):
    question = body["message"]["parts"][0].get("text", "")
    result = await ask(question)

    return {"message": {"role": "ROLE_AGENT", "parts": [{"text": result["answer"]}]}}
