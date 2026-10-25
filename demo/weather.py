"""Weather agent with fake data. uv run uvicorn weather:app --port 9201"""

import hashlib

from fastapi import FastAPI

from util import lifespan

NAME = "weather"
VERSION = "1.1.0"
PORT = 9201
CONDITIONS = ["sunny", "cloudy", "rain", "snow", "windy", "fog"]
DAYS = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"]

app = FastAPI(lifespan=lifespan(NAME, VERSION, PORT))


def seed(text):
    return int(hashlib.sha256(text.lower().encode()).hexdigest(), 16)


def current(args):
    city = args.get("city") or "Berlin"
    n = seed(city)
    return {
        "city": city,
        "temperature_c": n % 35 - 5,
        "humidity": n % 60 + 30,
        "wind_kph": n % 40,
        "condition": CONDITIONS[n % len(CONDITIONS)],
    }


def forecast(args):
    city = args.get("city") or "Berlin"
    days = int(args.get("days") or 3)
    return {
        "city": city,
        "days": [
            {
                "day": DAYS[i % len(DAYS)],
                "high_c": seed(f"{city}{i}") % 30,
                "low_c": seed(f"{city}{i}") % 30 - 8,
                "condition": CONDITIONS[seed(f"{city}{i}") % len(CONDITIONS)],
            }
            for i in range(days)
        ],
    }


def alerts(args):
    city = args.get("city") or "Berlin"
    n = seed(city)
    return {
        "city": city,
        "alerts": [{"type": "storm", "severity": "moderate"}] if n % 4 == 0 else [],
    }


SKILLS = {
    "weather.current.get": current,
    "weather.forecast.get": forecast,
    "weather.alerts.list": alerts,
}


@app.get("/healthz")
def healthz():
    return {"status": "ok"}


@app.get("/.well-known/agent-card.json")
def agent_card():
    return {
        "name": "Weather",
        "description": "Current conditions, forecasts and alerts for a city",
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
                "id": "weather.current.get",
                "name": "Current weather",
                "description": "Return temperature, humidity, wind and condition for a city. Data: {\"city\": \"Berlin\"}",
                "tags": ["weather", "current"],
                "inputModes": ["application/json"],
                "examples": ['{"city": "Berlin"}'],
            },
            {
                "id": "weather.forecast.get",
                "name": "Weather forecast",
                "description": "Return a daily high, low and condition forecast for a city. Data: {\"city\": \"Berlin\", \"days\": 3}",
                "tags": ["weather", "forecast"],
                "inputModes": ["application/json"],
                "examples": ['{"city": "Cairo", "days": 5}'],
            },
            {
                "id": "weather.alerts.list",
                "name": "Weather alerts",
                "description": "List active severe weather alerts for a city. Data: {\"city\": \"Miami\"}",
                "tags": ["weather", "alerts"],
                "inputModes": ["application/json"],
                "examples": ['{"city": "Miami"}'],
            },
        ],
    }


@app.post("/a2a/v1/message:send")
def message_send(body: dict):
    skill = body.get("skill") or "weather.current.get"
    data = body["message"]["parts"][0].get("data") or {}
    return {"message": {"role": "ROLE_AGENT", "parts": [{"data": SKILLS[skill](data)}]}}
