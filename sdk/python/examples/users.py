"""Users agent with fake profiles, built with the Sloth SDK. python users.py

Protection is on by default: Sloth keeps the secret and adds it to every call it
forwards, so other agents can only reach this one through the gateway.
"""

import hashlib

from sloth import Agent

PLANS = ["free", "pro", "team", "enterprise"]

agent = Agent(
    name="users",
    version="1.0.0",
    description="User profiles by user id",
    port=9203,
    allow_calls_from=["concierge"],
)


@agent.skill("user.profile.get", tags=["users", "profile"], examples=['{"user_id": "u_123"}'])
def profile(user_id="u_123"):
    """Return a user's name, email and plan."""
    n = int(hashlib.sha256(user_id.encode()).hexdigest(), 16)
    first = ["Ada", "Grace", "Alan", "Linus"][n % 4]

    return {"user_id": user_id, "name": first, "email": f"{first.lower()}@example.com", "plan": PLANS[n % len(PLANS)]}


if __name__ == "__main__":
    agent.run()
