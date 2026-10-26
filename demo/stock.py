"""Stock agent with live data from Nasdaq. uv run uvicorn stock:app --port 9202"""

import os
from datetime import date, timedelta

import httpx
from fastapi import FastAPI

from util import lifespan

NAME = "stock"
VERSION = "1.0.0"
PORT = int(os.environ.get("PORT", "9202"))

HEADERS = {"User-Agent": "Mozilla/5.0", "Accept": "application/json"}

app = FastAPI(lifespan=lifespan(NAME, VERSION, PORT))


def money(text):
    return float(text.replace("$", "").replace(",", "").replace("+", "").rstrip("%"))


def info(symbol):
    response = httpx.get(
        f"https://api.nasdaq.com/api/quote/{symbol}/info?assetclass=stocks",
        headers=HEADERS,
        timeout=10,
    )
    response.raise_for_status()

    return response.json()["data"]


def quote(args):
    symbol = (args.get("symbol") or "AAPL").upper()
    data = info(symbol)
    last = money(data["primaryData"]["lastSalePrice"])
    change = money(data["primaryData"]["netChange"])

    return {
        "symbol": symbol,
        "price": last,
        "previous_close": round(last - change, 2),
        "change_percent": money(data["primaryData"]["percentageChange"]),
        "currency": data["primaryData"].get("currency") or "USD",
    }


def history(args):
    symbol = (args.get("symbol") or "AAPL").upper()
    days = int(args.get("days") or 5)
    start = (date.today() - timedelta(days=days * 2 + 10)).isoformat()
    response = httpx.get(
        f"https://api.nasdaq.com/api/quote/{symbol}/historical",
        params={"assetclass": "stocks", "fromdate": start, "limit": days},
        headers=HEADERS,
        timeout=10,
    )
    response.raise_for_status()
    rows = response.json()["data"]["tradesTable"]["rows"]

    return {
        "symbol": symbol,
        "closes": [{"day": -n, "close": money(row["close"])} for n, row in enumerate(rows[:days], start=1)],
    }


def profile(args):
    symbol = (args.get("symbol") or "AAPL").upper()
    data = info(symbol)

    return {
        "symbol": symbol,
        "name": data["companyName"],
        "exchange": data["exchange"],
    }


SKILLS = {
    "stock.quote.get": quote,
    "stock.history.get": history,
    "stock.profile.get": profile,
}


@app.get("/healthz")
def healthz():
    return {"status": "ok"}


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
                "description": "Return the latest price, previous close and daily change for a ticker. Data: {\"symbol\": \"AAPL\"}",
                "tags": ["stock", "quote", "finance"],
                "inputModes": ["application/json"],
                "examples": ['{"symbol": "AAPL"}'],
            },
            {
                "id": "stock.history.get",
                "name": "Price history",
                "description": "Return daily closing prices for a ticker over the last N days. Data: {\"symbol\": \"NVDA\", \"days\": 10}",
                "tags": ["stock", "history", "finance"],
                "inputModes": ["application/json"],
                "examples": ['{"symbol": "NVDA", "days": 10}'],
            },
            {
                "id": "stock.profile.get",
                "name": "Company profile",
                "description": "Return company name and exchange for a ticker. Data: {\"symbol\": \"MSFT\"}",
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
