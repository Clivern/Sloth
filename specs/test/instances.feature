Feature: Agent instances

  Each running copy of an agent registers itself as an instance with an
  address and port, usually with a narrow access key. An instance's status is
  the worst status among its checks, and the passing filter is what callers
  use to pick a live instance, so it must never return a sick one.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances
    GET    /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances?passing=true
    PUT    /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}
    GET    /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}
    DELETE /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}

  Background:
    Given the run's workspace exists
    And the agent "spec-<runId>-echo" is registered

  # ---------------------------------------------------------------------------
  # Registering
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: An access key with the register grant can register an instance
    Given an access key with only CAN_REGISTER_INSTANCE
    When that key PUTs instance "i-1" with address "127.0.0.1", port 9000 and meta {"zone": "a"}
    Then the response is 201
    And instanceId is "i-1", address is "127.0.0.1", port is 9000 and meta is {"zone": "a"}

  @mutates
  Scenario: Registering the same instance again updates it
    Given instance "i-1" is registered with port 9000
    When the user PUTs instance "i-1" with address "127.0.0.1" and port 9001
    Then the response is 200
    And reading the instance returns port 9001
    And the agent has one instance

  @mutates
  Scenario: An instance with no checks is passing
    When the user registers instance "i-1"
    Then status is "passing"
    And checks is an empty array

  @readonly
  Scenario: An instance cannot be registered before its agent
    When the user PUTs instance "i-1" under agents/spec-<runId>-ghost with address "127.0.0.1" and port 9000
    Then the response is 404

  @readonly
  Scenario: Address and port are required and the port must be 1 to 65535
    When the user PUTs instance "i-1" with port 9000 and no address
    Then the response is 400
    When the user PUTs instance "i-1" with address "127.0.0.1" and port 0
    Then the response is 400
    When the user PUTs instance "i-1" with address "127.0.0.1" and port 70000
    Then the response is 400

  @readonly
  Scenario: Instance ids follow the same rules as agent names
    When the user PUTs instance "-bad" with address "127.0.0.1" and port 9000
    Then the response is 400

  # ---------------------------------------------------------------------------
  # Reading and the passing filter
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: The list shows every instance of the agent
    Given instances "i-1" and "i-2" are registered
    When the user lists the agent's instances
    Then the response is 200
    And instances contains "i-1" and "i-2"

  @mutates
  Scenario: The passing filter leaves out an instance whose TTL check never reported
    Given a ttl check template "heartbeat" with ttl 60 on the agent
    And instances "i-1" and "i-2" are registered
    And "i-1" reported "heartbeat" as passing
    When the user lists the agent's instances with passing=true
    Then instances contains "i-1" and not "i-2"

  @readonly
  Scenario: Reading an unknown instance returns not found
    When the user GETs instance "spec-<runId>-missing"
    Then the response is 404

  # ---------------------------------------------------------------------------
  # Deregistering
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: Deregistering an instance removes it
    Given instance "i-1" is registered
    When the user DELETEs instance "i-1"
    Then the response is 204
    And reading it returns 404

  @mutates
  Scenario: Deleting the agent takes its instances with it
    Given instance "i-1" is registered
    When the user DELETEs agents/spec-<runId>-echo
    And registers "spec-<runId>-echo" again
    Then the agent has no instances

  @mutates
  Scenario: A read-only access key can see instances but not register them
    Given an access key with only CAN_LIST_INSTANCES and CAN_GET_INSTANCE
    When that key lists the agent's instances
    Then the response is 200
    When that key PUTs instance "i-9" with address "127.0.0.1" and port 9000
    Then the response is 403

  # ---------------------------------------------------------------------------
  # More on the passing filter and ownership
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: The passing filter keeps instances that are only warning
    Given a ttl check template "heartbeat" with ttl 60
    And instance "i-1" reported "heartbeat" as warning
    When the user lists the agent's instances with passing=true
    Then instances contains "i-1"

  @readonly
  Scenario: The passing filter on an unknown agent returns not found
    When the user lists agents/spec-<runId>-ghost/instances?passing=true
    Then the response is 404

  @mutates
  Scenario: An instance cannot be reached through another agent
    Given instance "i-1" is registered on "spec-<runId>-echo"
    And a second agent "spec-<runId>-other" is registered
    When the user GETs agents/spec-<runId>-other/instances/i-1
    Then the response is 404

  @mutates
  Scenario: Deregistering an unknown instance returns not found
    When the user DELETEs instance "spec-<runId>-missing"
    Then the response is 404

  @readonly
  Scenario: An address may be at most 255 characters
    When the user PUTs instance "i-1" with a 256-character address and port 9000
    Then the response is 400
