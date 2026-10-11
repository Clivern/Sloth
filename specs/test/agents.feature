Feature: Agents

  An agent is registered in a workspace by name and version. Its A2A card is
  not sent on registration: the first instance to register serves it at the
  agent's card path, and Cognit fetches and stores it, pointing every
  interface URL at its own gateway so callers always go through it. Clients
  fetch the card and cache it by ETag, so the checksum must change exactly
  when the card does.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/agents
    PUT    /api/v1/workspaces/{workspaceId}/agents/{agentName}
    GET    /api/v1/workspaces/{workspaceId}/agents/{agentName}
    DELETE /api/v1/workspaces/{workspaceId}/agents/{agentName}
    GET    /api/v1/workspaces/{workspaceId}/agents/{agentName}/.well-known/agent-card.json
    PUT    /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}

  A card server is a local HTTP server the runner starts on 127.0.0.1 that
  answers GET <path> with a card. A minimal card is: {"name": "<agent name>",
  "version": "1.0.0", "supportedInterfaces": [{"url": "http://127.0.0.1:9000",
  "protocolBinding": "JSONRPC"}]}. "The agent has its card" means a card
  server serves the minimal card at /.well-known/agent-card.json and instance
  "i-1" registered with its address and port.

  Background:
    Given the run's workspace exists

  # ---------------------------------------------------------------------------
  # Registering
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: Registering a new agent creates it without a card
    When the user PUTs {"version": "1.0.0"} to agents/spec-<runId>-echo
    Then the response is 201
    And name is "spec-<runId>-echo" and version is "1.0.0"
    And card is null and cardChecksum is ""
    And cardPath is "/.well-known/agent-card.json"

  @mutates
  Scenario: A custom card path is kept
    When the user PUTs {"version": "1.0.0", "cardPath": "/a2a/agent.json"} to agents/spec-<runId>-echo
    Then the response is 201
    And cardPath is "/a2a/agent.json"

  @readonly
  Scenario: A card path must start with a slash
    When the user PUTs {"version": "1.0.0", "cardPath": "agent.json"} to agents/spec-<runId>-echo
    Then the response is 400
    And reading agents/spec-<runId>-echo returns 404

  @mutates
  Scenario: Registering the same name again updates the version
    Given the agent "spec-<runId>-echo" is registered with version "1.0.0"
    When the user PUTs {"version": "1.1.0"} to agents/spec-<runId>-echo
    Then the response is 200
    And reading the agent returns version "1.1.0"
    And the agents list holds "spec-<runId>-echo" once

  @readonly
  Scenario: A request without a version is rejected
    When the user PUTs agents/spec-<runId>-echo with body {}
    Then the response is 400
    And reading agents/spec-<runId>-echo returns 404

  @readonly
  Scenario: Agent names must start with a letter or digit and use only letters, digits, dot, dash and underscore
    When the user PUTs {"version": "1.0.0"} to agents/-spec-<runId>
    Then the response is 400
    When the user PUTs {"version": "1.0.0"} to agents/spec-<runId>%20echo
    Then the response is 400

  # ---------------------------------------------------------------------------
  # Fetching the card
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: The first instance to register supplies the card
    Given the agent "spec-<runId>-echo" is registered
    And a card server serving the minimal card at /.well-known/agent-card.json
    When instance "i-1" registers with the card server's address and port
    Then the response is 201
    And reading the agent returns card.name "spec-<runId>-echo" and a non-empty cardChecksum

  @mutates
  Scenario: The card is fetched from the agent's card path
    Given the agent "spec-<runId>-echo" is registered with cardPath "/a2a/agent.json"
    And a card server serving the minimal card at /a2a/agent.json and 404 everywhere else
    When instance "i-1" registers with the card server's address and port
    Then reading the agent returns a card

  @mutates
  Scenario: The fetched card's interface URLs are rewritten to the gateway
    Given the agent "spec-<runId>-echo" has its card
    When the user GETs the agent
    Then every card.supportedInterfaces[].url ends with
      "/api/v1/workspaces/<workspaceId>/a2a/agents/spec-<runId>-echo"
    And none of them is "http://127.0.0.1:9000"

  @mutates
  Scenario: A fetched card's own url field is rewritten to the gateway too
    Given the agent "spec-<runId>-echo" is registered
    And a card server serving a card that also has "url": "http://127.0.0.1:9000"
    When instance "i-1" registers with the card server's address and port
    Then card.url ends with "/api/v1/workspaces/<workspaceId>/a2a/agents/spec-<runId>-echo"

  @mutates
  Scenario: An instance still registers when its card cannot be fetched
    Given the agent "spec-<runId>-echo" is registered
    When instance "i-1" registers with address "127.0.0.1" and a port nothing listens on
    Then the response is 201
    And reading the agent returns card null

  @mutates
  Scenario: A later instance supplies the card when the first one could not
    Given the agent "spec-<runId>-echo" is registered
    And instance "i-1" registered with a port nothing listens on
    And a card server serving the minimal card at /.well-known/agent-card.json
    When instance "i-2" registers with the card server's address and port
    Then reading the agent returns a card

  @mutates
  Scenario: Once stored, the card is not replaced by later instances
    Given the agent "spec-<runId>-echo" has its card and its cardChecksum is remembered
    And a second card server serving a card with "description": "changed"
    When instance "i-2" registers with the second card server's address and port
    Then reading the agent returns the remembered cardChecksum

  @mutates
  Scenario: Changing the version clears the card until the next instance registers
    Given the agent "spec-<runId>-echo" has its card
    When the user PUTs {"version": "1.1.0"} to agents/spec-<runId>-echo
    Then card is null
    When instance "i-1" registers again with the card server's address and port
    Then reading the agent returns a card

  @mutates
  Scenario: Re-registering with the same version and card path keeps the card
    Given the agent "spec-<runId>-echo" has its card and its cardChecksum is remembered
    When the user PUTs {"version": "1.0.0"} to agents/spec-<runId>-echo
    Then cardChecksum is the remembered one

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
  Scenario: The well-known card is not found until it has been fetched
    Given the agent "spec-<runId>-echo" is registered
    When the user GETs agents/spec-<runId>-echo/.well-known/agent-card.json
    Then the response is 404
    And there is no ETag header

  @mutates
  Scenario: The well-known card is the stored card with an ETag
    Given the agent "spec-<runId>-echo" has its card
    When the user GETs agents/spec-<runId>-echo/.well-known/agent-card.json
    Then the response is 200
    And the body equals the card from reading the agent
    And the ETag header is the cardChecksum in double quotes
    And Cache-Control is "no-cache"

  @mutates
  Scenario: A client with the current ETag gets 304
    Given the agent "spec-<runId>-echo" has its card
    When the user GETs the well-known card with If-None-Match set to its ETag
    Then the response is 304 with no body

  @mutates
  Scenario: Changing the card changes the ETag
    Given the agent "spec-<runId>-echo" has its card and its ETag is remembered
    When the user PUTs {"version": "1.1.0"} to agents/spec-<runId>-echo
    And a card server serves a card with "description": "changed" and instance "i-1" registers with it
    And the user GETs the well-known card with If-None-Match set to the old ETag
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
    When that key PUTs {"version": "1.0.0"} to agents/spec-<runId>-echo
    Then the response is 403

  # ---------------------------------------------------------------------------
  # Health and lifecycle
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: An agent with no instances has unknown health
    Given the agent "spec-<runId>-echo" is registered with no instances
    When the user GETs the agent
    Then health is "unknown"
    And statusCounts is {"passing": 0, "warning": 0, "unknown": 0, "critical": 0}

  @mutates
  Scenario: An agent is as healthy as its worst instance
    Given a ttl check template "heartbeat" with ttl 60 on "spec-<runId>-echo"
    And instances "i-1" and "i-2" are registered
    And "i-1" reported "heartbeat" as passing and "i-2" as warning
    When the user GETs the agent
    Then health is "warning"
    And statusCounts.passing is 1 and statusCounts.warning is 1
    When "i-2" reports "heartbeat" as fail
    Then the agent's health is "critical"

  @readonly
  Scenario: A version may be at most 40 characters
    When the user PUTs a version of 41 characters to agents/spec-<runId>-echo
    Then the response is 400

  @mutates
  Scenario: Deleting an agent takes its checks, identities and protection with it
    Given "spec-<runId>-echo" has a check template, an identity and a protection
    When the user DELETEs the agent and registers it again
    Then it has no check templates, no identities and no protection

  @mutates
  Scenario: The same agent name in two workspaces is two agents
    Given a second workspace "spec-<runId>-other"
    And "spec-<runId>-echo" is registered in both with versions "1.0.0" and "2.0.0"
    When the user GETs the agent in each workspace
    Then the versions are "1.0.0" and "2.0.0"
    And deleting it in one leaves the other

  @mutates
  Scenario: The agent list pages
    Given agents "spec-<runId>-p1", "spec-<runId>-p2" and "spec-<runId>-p3" are registered
    When the user lists agents with limit 2 and offset 0, then offset 2
    Then the pages hold 2 and at least 1 agents with no name in both
    And _meta.total is at least 3
