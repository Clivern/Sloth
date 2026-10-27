"""Users agent with fake profiles, protected by an API key. uv run uvicorn users:app --port 9203

On boot it sets its protection in Sloth: Sloth keeps the key and adds the
X-Users-Key header to every call it forwards, and the agent refuses
message:send without it. Other agents can only reach it through Sloth.
"""

import hashlib
import os
import secrets

from fastapi import FastAPI, Header, HTTPException

from util import lifespan

NAME = "users"
VERSION = "1.0.0"
PORT = int(os.environ.get("PORT", "9203"))
API_KEY = "users-demo-key-2026"
PLANS = ["free", "pro", "team", "enterprise"]
LANGUAGES = ["en", "fr", "de", "es"]

app = FastAPI(
    lifespan=lifespan(
        NAME,
        VERSION,
        PORT,
        protection={"type": "api_key", "header": "X-Users-Key", "secret": API_KEY},
    )
)


def seed(text):
    return int(hashlib.sha256(text.lower().encode()).hexdigest(), 16)


def profile(args):
    email = args.get("email") or "ada@example.com"
    n = seed(email)
    first, _, last = email.split("@")[0].partition(".")

    return {
        "user_id": f"u_{n % 900 + 100}",
        "name": f"{first.title()} {last.title()}".strip(),
        "email": email.lower(),
        "company": ["Acme", "Globex", "Initech", "Umbrella", "Stark", "Wayne"][n // 4 % 6],
        "role": ["Engineer", "Designer", "Manager", "Analyst", "Founder", "Support"][n // 24 % 6],
        "country": ["Netherlands", "Germany", "France", "Spain", "Egypt", "USA"][n // 144 % 6],
        "plan": PLANS[n % len(PLANS)],
        "member_since": f"20{n % 10 + 15}-0{n % 9 + 1}-1{n % 9}",
    }


def preferences(args):
    email = args.get("email") or "ada@example.com"
    n = seed(email)

    return {
        "user_id": f"u_{n % 900 + 100}",
        "language": LANGUAGES[n % len(LANGUAGES)],
        "timezone": ["UTC", "Europe/Berlin", "Africa/Cairo", "America/New_York"][n % 4],
        "newsletter": n % 2 == 0,
    }


SKILLS = {
    "user.profile.get": profile,
    "user.preferences.get": preferences,
}


@app.get("/healthz")
def healthz():
    return {"status": "ok"}


@app.get("/.well-known/agent-card.json")
def agent_card():
    return {
        "name": "Users",
        "description": "User profiles and preferences by email",
        "version": VERSION,
        "supportedInterfaces": [
            {
                "url": f"http://127.0.0.1:{PORT}/a2a/v1",
                "protocolBinding": "HTTP+JSON",
                "protocolVersion": "1.0",
            }
        ],
        "defaultInputModes": ["application/json"],
        "defaultOutputModes": ["application/json"],
        "skills": [
            {
                "id": "user.profile.get",
                "name": "User profile",
                "description": 'Return a user\'s name, company, role, country, plan and member date by email. Data: {"email": "ada@example.com"}',
                "tags": ["users", "profile"],
                "inputModes": ["application/json"],
                "examples": ['{"email": "ada@example.com"}'],
            },
            {
                "id": "user.preferences.get",
                "name": "User preferences",
                "description": 'Return a user\'s language, timezone and newsletter choice by email. Data: {"email": "ada@example.com"}',
                "tags": ["users", "preferences"],
                "inputModes": ["application/json"],
                "examples": ['{"email": "ada@example.com"}'],
            },
        ],
    }


@app.post("/a2a/v1/message:send")
def message_send(body: dict, x_users_key: str = Header(default="")):
    # Only Sloth holds the key, so a call without it did not come through the gateway.
    if not secrets.compare_digest(x_users_key, API_KEY):
        raise HTTPException(status_code=401, detail="missing or wrong X-Users-Key")

    skill = body.get("skill") or "user.profile.get"
    data = body["message"]["parts"][0].get("data") or {}

    return {"message": {"role": "ROLE_AGENT", "parts": [{"data": SKILLS[skill](data)}]}}
