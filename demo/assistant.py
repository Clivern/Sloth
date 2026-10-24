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

import httpx
from fastapi import FastAPI
from fastapi.responses import HTMLResponse

from util import URL, WORKSPACE, lifespan

NAME = "assistant"
VERSION = "1.0.0"
PORT = 9200
SECRET = os.environ.get("SLOTH_AGENT_SECRET", "")
OPENROUTER_API_KEY = os.environ.get("OPENROUTER_API_KEY", "")
MODEL = os.environ.get("ASSISTANT_MODEL", "openai/gpt-4o-mini")
GATEWAY = f"{URL}/api/v1/workspaces/{WORKSPACE}/agents"
MAX_CALLS = 3

PLAN = """You route a user's question to agent skills.
You get the question and the skills you may call, each with an id, a name and a
description that ends with an example of its data. Reply with JSON only:
{"calls": [{"skill": "<skill id>", "data": {...}}]}
Use only the listed skill ids, at most 3 calls, and fill data from the question
using the same keys as the example. Reply {"calls": []} when no skill helps."""

ANSWER = """You answer the user's question in two or three friendly sentences,
using only the agent results you are given. Each result names the skill it came
from and its HTTP status; say plainly when a call failed or was not allowed.
Do not invent data."""

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


PAGE = """<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Sloth Assistant</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500&family=IBM+Plex+Sans:wght@400;500;600&display=swap" rel="stylesheet">
<script src="https://cdn.jsdelivr.net/npm/marked@15.0.12/marked.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/dompurify@3.2.6/dist/purify.min.js"></script>
<style>
  :root {
    --bg: #f6f6f4;
    --surface: #ffffff;
    --ink: #111111;
    --muted: #8a8a86;
    --line: #ebebe8;
    --chip: #f0f0ed;
    --bad: #8a2f2f;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: "IBM Plex Sans", sans-serif; background: var(--bg); color: var(--ink); min-height: 100vh; padding: 48px 24px 80px; }
  .page { max-width: 1080px; margin: 0 auto; }
  .brand { display: flex; flex-direction: column; align-items: center; gap: 10px; margin-bottom: 28px; }
  .brand svg { width: 56px; height: 56px; }
  .brand h1 { font-size: 15px; font-weight: 600; letter-spacing: -0.01em; }
  .brand p { font-size: 13px; color: var(--muted); }
  .playground { display: flex; flex-direction: column; gap: 12px; }
  .play-card { background: var(--surface); border: 1px solid var(--line); border-radius: 14px; box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03); padding: 16px; }
  .play-field { display: flex; flex-direction: column; gap: 6px; min-width: 0; }
  .play-label { font-size: 11px; font-weight: 500; letter-spacing: 0.07em; text-transform: uppercase; color: var(--muted); }
  .chips { display: flex; flex-wrap: wrap; gap: 6px; min-height: 40px; align-items: center; }
  .chip { font-family: "IBM Plex Mono", monospace; font-size: 12px; background: var(--chip); border-radius: 6px; padding: 4px 8px; }
  .play-note { font-size: 12px; line-height: 1.45; color: var(--muted); }
  .play-chat { min-height: 360px; max-height: min(60vh, 560px); overflow-y: auto; display: flex; flex-direction: column; gap: 18px; padding: 4px 2px; }
  .play-empty { margin: auto; text-align: center; color: var(--muted); font-size: 14px; line-height: 1.5; max-width: 360px; padding: 40px 16px; }
  .play-empty .examples { display: flex; flex-direction: column; gap: 6px; margin-top: 14px; }
  .play-empty button { font: inherit; font-size: 13px; color: var(--ink); background: var(--bg); border: 1px solid var(--line); border-radius: 8px; padding: 8px 12px; cursor: pointer; }
  .play-empty button:hover { background: var(--surface); border-color: #c8c8c4; }
  .play-msg { display: flex; flex-direction: column; gap: 6px; max-width: 780px; }
  .play-msg-role { font-size: 11px; font-weight: 500; letter-spacing: 0.06em; text-transform: uppercase; color: var(--muted); }
  .play-msg-body { font-size: 14px; line-height: 1.55; word-break: break-word; }
  .play-msg.user .play-msg-body, .play-msg.error .play-msg-body { white-space: pre-wrap; }
  .play-msg.error .play-msg-body { color: var(--bad); }
  .play-msg.pending .play-msg-body { color: var(--muted); }
  .play-msg-body > :first-child { margin-top: 0; }
  .play-msg-body > :last-child { margin-bottom: 0; }
  .play-msg-body p, .play-msg-body ul, .play-msg-body ol { margin: 0 0 0.75em; }
  .play-msg-body ul, .play-msg-body ol { padding-left: 1.35em; }
  .play-msg-body strong { font-weight: 600; }
  .play-msg-body code { font-family: "IBM Plex Mono", monospace; font-size: 12.5px; background: var(--chip); border-radius: 4px; padding: 1px 5px; }
  details.trace { margin-top: 4px; font-size: 13px; }
  details.trace summary { cursor: pointer; color: var(--muted); font-size: 12px; list-style: none; }
  details.trace summary::before { content: "▸ "; }
  details.trace[open] summary::before { content: "▾ "; }
  details.trace ol { margin: 10px 0 0; padding-left: 1.35em; display: flex; flex-direction: column; gap: 10px; }
  .trace .status { font-family: "IBM Plex Mono", monospace; font-size: 12px; font-weight: 500; padding: 1px 6px; border-radius: 4px; background: var(--chip); }
  .trace .status.bad { color: var(--bad); }
  .trace pre { margin-top: 6px; background: #4a4a45; color: #f5f5f2; border-radius: 10px; padding: 10px 12px; overflow-x: auto; font-family: "IBM Plex Mono", monospace; font-size: 12px; }
  .play-composer { display: flex; flex-direction: column; gap: 10px; }
  .play-composer textarea { font: inherit; font-size: 14px; border: 1px solid var(--line); background: var(--bg); border-radius: 8px; padding: 10px 12px; color: var(--ink); outline: none; width: 100%; min-height: 88px; resize: vertical; }
  .play-composer textarea:focus { border-color: #c8c8c4; background: var(--surface); }
  .play-composer-bar { display: flex; justify-content: space-between; align-items: center; gap: 10px; flex-wrap: wrap; }
  .play-composer-actions { display: flex; gap: 8px; margin-left: auto; }
  .play-btn { font: inherit; font-size: 13px; font-weight: 500; border-radius: 8px; padding: 10px 14px; cursor: pointer; border: 1px solid var(--line); background: var(--surface); color: var(--ink); white-space: nowrap; }
  .play-btn:hover { background: #fafaf8; }
  .play-btn.primary { background: var(--ink); border-color: var(--ink); color: #fff; }
  .play-btn.primary:hover { background: #2a2a2a; }
  .play-btn:disabled { opacity: 0.45; cursor: not-allowed; }
  .copyright { margin-top: 40px; text-align: center; font-size: 12px; color: var(--muted); }
  @media (max-width: 700px) {
    body { padding: 28px 14px 60px; }
    .play-chat { min-height: 280px; }
    .play-composer textarea { font-size: 16px; }
  }
</style>
</head>
<body>
<div class="page">
  <div class="brand">
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M12 8V4H8"/><rect width="16" height="12" x="4" y="8" rx="2"/><path d="M2 14h2"/><path d="M20 14h2"/><path d="M15 13v2"/><path d="M9 13v2"/>
    </svg>
    <h1>Sloth Assistant</h1>
    <p>Asks other agents through the Sloth gateway, then writes one answer.</p>
  </div>

  <div class="playground">
    <div class="play-card">
      <div class="play-field">
        <span class="play-label">Skills Sloth lets the assistant call</span>
        <div class="chips" id="skills"><span class="play-note">Loading…</span></div>
      </div>
    </div>

    <div class="play-card">
      <div id="chat" class="play-chat" aria-live="polite"></div>
    </div>

    <div class="play-card play-composer">
      <textarea id="input" placeholder="Ask about the weather, a stock or a user…" rows="3"></textarea>
      <div class="play-composer-bar">
        <span class="play-note" id="status">Enter to send, Shift+Enter for a new line</span>
        <div class="play-composer-actions">
          <button type="button" class="play-btn" id="clear">Clear chat</button>
          <button type="button" class="play-btn primary" id="send">Send</button>
        </div>
      </div>
    </div>
  </div>

  <p class="copyright">Sloth demo agent</p>
</div>
<script>
  const EXAMPLES = [
    "What's the weather in Cairo right now?",
    "Will it rain in Berlin in the next 3 days?",
    "How is NVDA doing, and who makes it?",
    "What plan is user u_123 on, and what language do they use?",
  ];
  const chat = document.getElementById("chat");
  const input = document.getElementById("input");
  const send = document.getElementById("send");
  const status = document.getElementById("status");
  let messages = [];
  let sending = false;

  const escapeHtml = (text) => String(text).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]);
  const markdown = (text) => DOMPurify.sanitize(marked.parse(text, { breaks: true, gfm: true }));
  const json = (value) => `<pre>${escapeHtml(JSON.stringify(value, null, 2))}</pre>`;
  const code = (value) => `<span class="status ${value >= 300 ? "bad" : ""}">${value}</span>`;

  function trace(steps) {
    const items = steps.map((step) => {
      if (step.step === "catalog") return `<li>Listed callable skills from Sloth ${code(step.status)}: ${escapeHtml(step.skills.join(", ") || "none")}</li>`;
      if (step.step === "plan") return `<li>The model chose ${step.calls.length} call(s)${json(step.calls)}</li>`;
      if (step.step === "call") return `<li>Called <code>${escapeHtml(step.skill)}</code> through Sloth ${code(step.status)}${json(step.result)}</li>`;
      return `<li>Rephrased the results with <code>${escapeHtml(step.model)}</code></li>`;
    });

    return `<details class="trace"><summary>How it got there</summary><ol>${items.join("")}</ol></details>`;
  }

  function render() {
    if (!messages.length) {
      chat.innerHTML = `<div class="play-empty">Ask a question and the assistant will find the agents that can answer it.
        <div class="examples">${EXAMPLES.map((e) => `<button type="button">${escapeHtml(e)}</button>`).join("")}</div></div>`;
      chat.querySelectorAll(".examples button").forEach((b) => b.addEventListener("click", () => { input.value = b.textContent; ask(); }));
      return;
    }

    chat.innerHTML = messages.map((m) => `
      <div class="play-msg ${m.role}${m.error ? " error" : ""}${m.pending ? " pending" : ""}">
        <div class="play-msg-role">${m.error ? "Error" : m.role === "user" ? "You" : "Assistant"}</div>
        <div class="play-msg-body">${m.role === "user" || m.error || m.pending ? escapeHtml(m.content) : markdown(m.content)}</div>
        ${m.steps ? trace(m.steps) : ""}
      </div>`).join("");
    chat.scrollTop = chat.scrollHeight;
  }

  async function ask() {
    const question = input.value.trim();
    if (sending || !question) return;

    messages.push({ role: "user", content: question });
    const reply = { role: "assistant", content: "Asking agents through Sloth…", pending: true };
    messages.push(reply);
    input.value = "";
    sending = true;
    send.disabled = true;
    status.textContent = "Working…";
    render();

    try {
      const response = await fetch("/ask", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ question }) });
      if (!response.ok) throw new Error(`The assistant answered ${response.status}`);
      const data = await response.json();
      Object.assign(reply, { content: data.answer, steps: data.steps, pending: false });
    } catch (error) {
      Object.assign(reply, { content: error.message, error: true, pending: false });
    } finally {
      sending = false;
      send.disabled = false;
      status.textContent = "Enter to send, Shift+Enter for a new line";
      render();
    }
  }

  async function info() {
    const data = await (await fetch("/info")).json();
    document.getElementById("skills").innerHTML = data.skills.length
      ? data.skills.map((s) => `<span class="chip">${escapeHtml(s)}</span>`).join("")
      : data.status === 401
        ? `<span class="play-note">None yet (Sloth answered 401). Issue an api_key identity to the assistant under Agents and set SLOTH_AGENT_SECRET.</span>`
        : `<span class="play-note">None yet (Sloth answered ${data.status}). Allow the assistant to call agents under Intentions.</span>`;
  }

  send.addEventListener("click", ask);
  document.getElementById("clear").addEventListener("click", () => { messages = []; render(); });
  input.addEventListener("keydown", (event) => {
    if (event.key === "Enter" && !event.shiftKey) { event.preventDefault(); ask(); }
  });

  render();
  info();
</script>
</body>
</html>
"""
