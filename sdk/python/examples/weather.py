"""Weather agent with fake data, built with the Sloth SDK. python weather.py"""

import hashlib

from sloth import Agent

CONDITIONS = ["sunny", "cloudy", "rain", "snow", "windy", "fog"]

agent = Agent(
    name="weather",
    version="1.1.0",
    description="Current weather and forecasts by city",
    port=9201,
    allow_calls_from=["concierge"],
)


def seed(text):
    return int(hashlib.sha256(text.lower().encode()).hexdigest(), 16)


@agent.skill("weather.current", tags=["weather"], examples=['{"city": "Cairo"}'])
def current(city="Berlin"):
    """Return the temperature, humidity, wind and condition for a city."""
    n = seed(city)
    return {"city": city, "temperature_c": n % 35 - 5, "humidity": n % 60 + 30, "condition": CONDITIONS[n % len(CONDITIONS)]}


@agent.skill("weather.forecast", tags=["weather"], examples=['{"city": "Cairo", "days": 3}'])
def forecast(city="Berlin", days=3):
    """Return a daily high, low and condition for the next few days."""
    return {
        "city": city,
        "days": [{"high_c": seed(f"{city}{i}") % 30, "condition": CONDITIONS[seed(f"{city}{i}") % len(CONDITIONS)]} for i in range(int(days))],
    }


if __name__ == "__main__":
    agent.run()
