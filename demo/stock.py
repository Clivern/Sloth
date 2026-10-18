"""Stock agent with fake data. uv run uvicorn stock:app --port 9202"""

import hashlib

from fastapi import FastAPI

from util import lifespan

NAME = "stock"
VERSION = "1.0.0"
PORT = 9202
COMPANIES = {
    "AAPL": "Apple Inc.",
    "MSFT": "Microsoft Corporation",
    "GOOGL": "Alphabet Inc.",
    "AMZN": "Amazon.com, Inc.",
    "NVDA": "NVIDIA Corporation",
}

app = FastAPI(lifespan=lifespan(NAME, VERSION, PORT))


def seed(text):
    return int(hashlib.sha256(text.upper().encode()).hexdigest(), 16)


def price(symbol, offset=0):
    base = 50 + seed(symbol) % 45000 / 100
    return round(base * (1 + (seed(f"{symbol}{offset}") % 600 - 300) / 10000), 2)


def quote(args):
    symbol = (args.get("symbol") or "AAPL").upper()
    last = price(symbol)
    close = price(symbol, 1)
    return {
        "symbol": symbol,
        "price": last,
        "previous_close": close,
        "change_percent": round((last - close) / close * 100, 2),
        "currency": "USD",
    }


def history(args):
    symbol = (args.get("symbol") or "AAPL").upper()
    days = int(args.get("days") or 5)
    return {
        "symbol": symbol,
        "closes": [{"day": -i, "close": price(symbol, i)} for i in range(1, days + 1)],
    }


def profile(args):
    symbol = (args.get("symbol") or "AAPL").upper()
    return {
        "symbol": symbol,
        "name": COMPANIES.get(symbol, f"{symbol} Holdings"),
        "exchange": "NASDAQ",
        "market_cap_billion": seed(symbol) % 3000 + 10,
    }


SKILLS = {
    "stock.quote.get": quote,
    "stock.history.get": history,
    "stock.profile.get": profile,
}


@app.get("/.well-known/agent-card.json")
def agent_card():
    return {
        "name": "Stock",
        "description": "Stock quotes, price history and company profiles by ticker",
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
                "id": "stock.quote.get",
                "name": "Stock quote",
                "description": "Return the latest price, previous close and daily change for a ticker",
                "tags": ["stock", "quote", "finance"],
                "inputModes": ["application/json"],
                "examples": ['{"symbol": "AAPL"}'],
            },
            {
                "id": "stock.history.get",
                "name": "Price history",
                "description": "Return daily closing prices for a ticker over the last N days",
                "tags": ["stock", "history", "finance"],
                "inputModes": ["application/json"],
                "examples": ['{"symbol": "NVDA", "days": 10}'],
            },
            {
                "id": "stock.profile.get",
                "name": "Company profile",
                "description": "Return company name, exchange and market cap for a ticker",
                "tags": ["stock", "company", "finance"],
                "inputModes": ["application/json"],
                "examples": ['{"symbol": "MSFT"}'],
            },
        ],
    }


@app.post("/a2a/v1/message:send")
def message_send(body: dict):
    skill = body.get("skill") or "stock.quote.get"
    data = body["message"]["parts"][0].get("data") or {}
    return {"message": {"role": "ROLE_AGENT", "parts": [{"data": SKILLS[skill](data)}]}}
