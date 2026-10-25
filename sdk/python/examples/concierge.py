"""Concierge agent that answers by calling the users and weather agents through Sloth. python concierge.py

Once all three are running, call concierge.greet through the Sloth gateway like any
other agent, and watch the calls it makes appear in Traffic.
"""

from sloth import Agent

agent = Agent(
    name="concierge",
    version="1.0.0",
    description="Greets a user with the weather in their city",
    port=9200,
    can_call=["users", "weather"],
)


@agent.skill("concierge.greet", examples=['{"user_id": "u_7", "city": "Cairo"}'])
async def greet(user_id="u_123", city="Berlin"):
    """Greet a user by name with the current weather in a city."""
    user = await agent.call("users", {"user_id": user_id}, skill="user.profile.get")
    weather = await agent.call_skill("weather.current", {"city": city})

    return f"Hi {user['name']}, it is {weather['condition']} and {weather['temperature_c']}°C in {city}."


if __name__ == "__main__":
    agent.run()
