Feature: Agent catalog

  An agent asks the catalog which skills it may call and which agents serve
  them. The answer must match what the gateway would allow: a skill shows up
  only when the best matching intention for the caller, the agent and the
  skill is allow. A rule for the exact skill beats a rule for the whole agent,
  and no rule means deny. Showing a skill the caller cannot call, or hiding
  one it can, sends agents to the wrong place.

  Routes (shapes in openapi.yaml):
    GET /api/v1/workspaces/{workspaceId}/agents/catalog/skills
    GET /api/v1/workspaces/{workspaceId}/agents/catalog/skills/{skill}/agents

  These routes authenticate the calling agent, not a user. "The caller sends"
  means X-Source-Agent: spec-<runId>-caller with X-API-Key set to the caller's
  api_key identity secret, and no user API key.

  A skill card is the minimal card from agents.feature plus a skills array,
  for example {"skills": [{"id": "spec-<runId>.profile.get", "name": "Get
  profile", "tags": ["profile"]}]}. "Agent X serves skills A and B" means X is
  registered, a card server serves a skill card listing A and B, and instance
  "i-1" registered with it.

  Background:
    Given the run's workspace exists
    And the agent "spec-<runId>-caller" is registered with an api_key identity
    And the agent "spec-<runId>-directory" serves skills "spec-<runId>.profile.get" tagged "profile" and "spec-<runId>.user.delete" tagged "admin"
    And the agent "spec-<runId>-mirror" serves skill "spec-<runId>.profile.get" tagged "profile"
    And the agent "spec-<runId>-billing" serves skill "spec-<runId>.invoice.get"

  # ---------------------------------------------------------------------------
  # Skills
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: The caller sees only the skills an allow intention covers
    Given an intention allows "spec-<runId>-caller" to call "spec-<runId>-directory"
    When the caller sends GET agents/catalog/skills
    Then the response is 200
    And skills has ids "spec-<runId>.profile.get" and "spec-<runId>.user.delete" in that order
    And no skill has id "spec-<runId>.invoice.get"
    And _meta.total is 2

  @mutates
  Scenario: A skill rule beats an agent rule
    Given an intention allows "spec-<runId>-caller" to call "spec-<runId>-directory"
    And an intention denies "spec-<runId>-caller" calling "spec-<runId>-directory" for skill "spec-<runId>.user.delete"
    When the caller sends GET agents/catalog/skills
    Then the response is 200
    And skills has only id "spec-<runId>.profile.get"

  @mutates
  Scenario: A skill offered by several callable agents is listed once
    Given an intention allows "spec-<runId>-caller" to call "spec-<runId>-directory"
    And an intention allows "spec-<runId>-caller" to call "spec-<runId>-mirror"
    When the caller sends GET agents/catalog/skills
    Then the skill "spec-<runId>.profile.get" appears once with agents 2

  @mutates
  Scenario: A caller with no intentions gets an empty list
    When the caller sends GET agents/catalog/skills
    Then the response is 200
    And skills is an empty list
    And _meta.total is 0

  @mutates
  Scenario: The caller never sees its own skills
    Given the agent "spec-<runId>-caller" also serves skill "spec-<runId>.ticket.open"
    And an intention allows "spec-<runId>-caller" to call "spec-<runId>-caller"
    When the caller sends GET agents/catalog/skills
    Then no skill has id "spec-<runId>.ticket.open"

  @mutates
  Scenario: Tags keep only skills that have one of them
    Given an intention allows "spec-<runId>-caller" to call "spec-<runId>-directory"
    When the caller sends GET agents/catalog/skills?tags=profile,ops
    Then skills has only id "spec-<runId>.profile.get"

  @mutates
  Scenario: Passing ignores agents without a live instance
    Given an intention allows "spec-<runId>-caller" to call "spec-<runId>-directory"
    And instance "i-1" of "spec-<runId>-directory" is deregistered
    When the caller sends GET agents/catalog/skills?passing=true
    Then skills is an empty list

  @mutates
  Scenario: The total counts every page
    Given an intention allows "spec-<runId>-caller" to call "spec-<runId>-directory"
    When the caller sends GET agents/catalog/skills?limit=1&offset=1
    Then skills has only id "spec-<runId>.user.delete"
    And _meta.total is 2

  # ---------------------------------------------------------------------------
  # Agents for a skill
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: The caller sees the agents it may call for a skill
    Given an intention allows "spec-<runId>-caller" to call "spec-<runId>-directory"
    When the caller sends GET agents/catalog/skills/spec-<runId>.profile.get/agents
    Then the response is 200
    And skill.id is "spec-<runId>.profile.get"
    And agents has only name "spec-<runId>-directory"
    And its status is "passing" and instances is {"total": 1, "passing": 1}
    And its sendUrl ends with "/api/v1/workspaces/<workspaceId>/agents/spec-<runId>-directory/message:send"

  @mutates
  Scenario: Agents the caller may not call are left out
    Given an intention allows "spec-<runId>-caller" to call "spec-<runId>-directory"
    When the caller sends GET agents/catalog/skills/spec-<runId>.profile.get/agents
    Then no agent has name "spec-<runId>-mirror"

  @mutates
  Scenario: A skill the caller may not call anywhere gives an empty list
    When the caller sends GET agents/catalog/skills/spec-<runId>.invoice.get/agents
    Then the response is 200
    And agents is an empty list

  @mutates
  Scenario: A skill no agent offers is not found
    When the caller sends GET agents/catalog/skills/spec-<runId>.missing/agents
    Then the response is 404
    And errorMessage is "No agent in the workspace advertises this skill."

  @mutates
  Scenario: Passing leaves out agents without a live instance
    Given an intention allows "spec-<runId>-caller" to call "spec-<runId>-directory"
    And an intention allows "spec-<runId>-caller" to call "spec-<runId>-mirror"
    And instance "i-1" of "spec-<runId>-mirror" is deregistered
    When the caller sends GET agents/catalog/skills/spec-<runId>.profile.get/agents?passing=true
    Then agents has only name "spec-<runId>-directory"

  # ---------------------------------------------------------------------------
  # Authentication
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A request without agent credentials is rejected
    When a request with no credentials calls GET agents/catalog/skills
    Then the response is 401

  @mutates
  Scenario: A wrong secret is rejected
    When the caller sends GET agents/catalog/skills with X-API-Key "wrong"
    Then the response is 401
    And errorMessage is "Invalid agent identity."

  @mutates
  Scenario: A user API key is not an agent credential
    When the user calls GET agents/catalog/skills with their own API key
    Then the response is 401

  @mutates
  Scenario: A turned-off identity is rejected
    Given the caller's api_key identity is turned off
    When the caller sends GET agents/catalog/skills
    Then the response is 401

  @mutates
  Scenario: A client credentials token needs the agent name
    Given the agent "spec-<runId>-caller" has a client_credentials identity and a token from it
    When a request sends GET agents/catalog/skills with only Authorization: Bearer <token>
    Then the response is 401
    When it also sends X-Source-Agent: spec-<runId>-caller
    Then the response is 200
