Feature: Workspace events

  Every event the server triggers in a workspace is stored, whether or not an
  integration is subscribed to it. The log is for owners and admins: ordinary
  members and access keys without the event grants cannot read it.

  Instances and checks trigger events as they change. An event names the user
  or access key behind the request as its actor, or "system" when no request
  caused it, such as a probe result.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/events
    GET    /api/v1/workspaces/{workspaceId}/events/{eventId}
    PUT    /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}
    DELETE /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}
    POST   /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}/checks/{checkId}/pass

  Background:
    Given the run's workspace exists

  @readonly @smoke
  Scenario: The owner can list the events
    When the user lists the run workspace's events
    Then the response is 200
    And events is an array
    And _meta has limit, offset and total

  @readonly
  Scenario: Reading an unknown event returns not found
    When the user GETs events/<a random UUID>
    Then the response is 404

  @mutates
  Scenario: An access key needs the event grant
    Given an access key with only CAN_LIST_AGENTS
    When that key lists the run workspace's events
    Then the response is 403
    Given an access key with only CAN_LIST_EVENTS
    When that key lists the run workspace's events
    Then the response is 200

  # ---------------------------------------------------------------------------
  # Instance changes
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: Registering an instance records an instance changed event by its access key
    Given the agent "spec-<runId>-echo" is registered
    And an access key named "spec-<runId>-reg" with only CAN_REGISTER_INSTANCE
    When that key PUTs instance "i-1" of "spec-<runId>-echo" with address "127.0.0.1" and port 9000
    Then the newest event has type "agent.instance.changed" and resourceType "agent_instance"
    And its actorType is "access_key", actorName is "spec-<runId>-reg" and actorId is the key's id
    And its data has action "registered", agentName "spec-<runId>-echo", instanceId "i-1",
      scheme "http", address "127.0.0.1", port "9000" and path ""
    And it has ipAddress and userAgent

  @mutates
  Scenario: Registering again at the same endpoint records nothing
    Given the agent "spec-<runId>-echo" is registered
    And instance "i-1" is registered with address "127.0.0.1" and port 9000
    When the user PUTs instance "i-1" with address "127.0.0.1" and port 9000 again
    Then the events total has not changed

  @mutates
  Scenario: Moving an instance to a new endpoint records an update
    Given the agent "spec-<runId>-echo" is registered
    And instance "i-1" is registered with address "127.0.0.1" and port 9000
    When the user PUTs instance "i-1" with address "127.0.0.1" and port 9001
    Then the newest event has type "agent.instance.changed"
    And its data has action "updated" and port "9001"
    And its actorType is "user"

  @mutates
  Scenario: Deregistering an instance records it
    Given the agent "spec-<runId>-echo" is registered
    And instance "i-1" is registered
    When the user DELETEs instance "i-1"
    Then the newest event has type "agent.instance.changed"
    And its data has action "deregistered", agentName "spec-<runId>-echo" and instanceId "i-1"
    And its data has no address

  # ---------------------------------------------------------------------------
  # Health changes
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A ttl report that changes the status records a health changed event
    Given the agent "spec-<runId>-echo" is registered
    And a ttl check template "heartbeat" with ttl 60 and instance "i-1" reporting pass
    When the user POSTs heartbeat/warn for "i-1" with output "queue is slow"
    Then the newest event has type "agent.health.changed" and resourceType "health_check"
    And its actorType is "user"
    And its data has agentName "spec-<runId>-echo", instanceId "i-1", checkId "heartbeat",
      from "passing", to "warning" and output "queue is slow"

  @mutates
  Scenario: A first passing report is not a change
    Given the agent "spec-<runId>-echo" is registered
    And a ttl check template "heartbeat" with ttl 60 and instance "i-1"
    When the user POSTs heartbeat/pass for "i-1"
    Then no event has type "agent.health.changed" for check "heartbeat"

  @mutates
  Scenario: Reporting the same status again records nothing
    Given the agent "spec-<runId>-echo" is registered
    And a ttl check template "heartbeat" with ttl 60 and instance "i-1" reporting pass
    When the user POSTs heartbeat/pass for "i-1" again
    Then the events total has not changed

  @mutates @slow
  Scenario: A ttl check reported after its ttl counts as critical before the report
    Given the agent "spec-<runId>-echo" is registered
    And a ttl check template "heartbeat" with ttl 5 and instance "i-1" reporting pass
    When 10 seconds pass
    And the user POSTs heartbeat/pass for "i-1"
    Then the newest event has type "agent.health.changed"
    And its data has from "critical" and to "passing"

  # Needs a running `sloth worker`.
  @mutates @worker @slow
  Scenario: A probe that fails records a system health changed event
    Given the agent "spec-<runId>-echo" is registered
    And a tcp check template "port" with port 1, interval 5 and timeout 2
    And instance "i-1" registered with address "127.0.0.1"
    When 15 seconds pass
    Then an event has type "agent.health.changed" with data checkId "port", from "unknown" and to "critical"
    And its actorType is "system" and it has no actorId
