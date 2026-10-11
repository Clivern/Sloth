Feature: Agents

  An agent is registered in a workspace by name with its A2A card. The registry
  keeps the card, but it points every interface URL at Cognit's own gateway so
  callers always go through it. Clients fetch the card and cache it by ETag,
  so the checksum must change exactly when the card does.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/agents
    PUT    /api/v1/workspaces/{workspaceId}/agents/{agentName}
    GET    /api/v1/workspaces/{workspaceId}/agents/{agentName}
    DELETE /api/v1/workspaces/{workspaceId}/agents/{agentName}
    GET    /api/v1/workspaces/{workspaceId}/agents/{agentName}/.well-known/agent-card.json

  A minimal card is: {"name": "<agent name>", "version": "1.0.0",
  "supportedInterfaces": [{"url": "http://127.0.0.1:9000", "protocolBinding": "JSONRPC"}]}

  Background:
    Given the run's workspace exists
    And a minimal card for "spec-<runId>-echo" with version "1.0.0"

  # ---------------------------------------------------------------------------
  # Registering
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: Registering a new agent creates it
    When the user PUTs {"card": <the card>} to agents/spec-<runId>-echo
    Then the response is 201
    And name is "spec-<runId>-echo", version is "1.0.0" and cardChecksum is not empty

  @mutates
  Scenario: Registering the same name again replaces the card
    Given the agent "spec-<runId>-echo" is registered
    When the user PUTs the card again with version "1.1.0"
    Then the response is 200
    And reading the agent returns version "1.1.0"
    And the agents list holds "spec-<runId>-echo" once

  @mutates
  Scenario: The card's interface URLs are rewritten to the gateway
    When the user registers "spec-<runId>-echo"
    Then every card.supportedInterfaces[].url ends with
      "/api/v1/workspaces/<workspaceId>/a2a/agents/spec-<runId>-echo"
    And none of them is "http://127.0.0.1:9000"

  @mutates
  Scenario: A top-level version wins over the card's version
    When the user PUTs {"card": <the card>, "version": "2.0.0"}
    Then version is "2.0.0"

  @mutates
  Scenario: Without a top-level version, the card's version is used
    When the user PUTs {"card": <the card>} with no top-level version
    Then version is "1.0.0"

  @readonly
  Scenario: An agent needs a version from somewhere
    When the user PUTs a card with no version and no top-level version
    Then the response is 400

  @readonly
  Scenario: A request without a card is rejected
    When the user PUTs agents/spec-<runId>-echo with body {}
    Then the response is 400
    And reading agents/spec-<runId>-echo returns 404

  @readonly
  Scenario: Agent names must start with a letter or digit and use only letters, digits, dot, dash and underscore
    When the user PUTs the card to agents/-spec-<runId>
    Then the response is 400
    When the user PUTs the card to agents/spec-<runId>%20echo
    Then the response is 400

  # ---------------------------------------------------------------------------
  # Reading
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A registered agent appears in the list with its health
    Given the agent "spec-<runId>-echo" is registered
    When the user lists the workspace's agents
    Then the response is 200
    And the agents array contains "spec-<runId>-echo" with health and statusCounts
    And _meta.total is at least 1

  @readonly
  Scenario: Reading an agent that does not exist returns not found
    When the user GETs agents/spec-<runId>-missing
    Then the response is 404

  # ---------------------------------------------------------------------------
  # The public card
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: The well-known card is the stored card with an ETag
    Given the agent "spec-<runId>-echo" is registered
    When the user GETs agents/spec-<runId>-echo/.well-known/agent-card.json
    Then the response is 200
    And the body equals the card from reading the agent
    And the ETag header is the cardChecksum in double quotes
    And Cache-Control is "no-cache"

  @mutates
  Scenario: A client with the current ETag gets 304
    Given the agent "spec-<runId>-echo" is registered
    When the user GETs the well-known card with If-None-Match set to its ETag
    Then the response is 304 with no body

  @mutates
  Scenario: Changing the card changes the ETag
    Given the agent "spec-<runId>-echo" is registered and its ETag is remembered
    When the user registers it again with version "1.1.0"
    And GETs the well-known card with If-None-Match set to the old ETag
    Then the response is 200
    And the ETag differs from the remembered one

  # ---------------------------------------------------------------------------
  # Deleting and access
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: Deleting an agent removes it
    Given the agent "spec-<runId>-echo" is registered
    When the user DELETEs agents/spec-<runId>-echo
    Then the response is 204
    And reading it returns 404
    And deleting it again returns 404

  @readonly @smoke
  Scenario: A request with no credentials is refused
    When a request with no credentials lists the run workspace's agents
    Then the response is 401

  @mutates
  Scenario: An access key can read agents but needs the upsert grant to write
    Given an access key with only CAN_LIST_AGENTS and CAN_GET_AGENT
    When that key lists the workspace's agents
    Then the response is 200
    When that key PUTs the card to agents/spec-<runId>-echo
    Then the response is 403
