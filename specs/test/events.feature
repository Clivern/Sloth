Feature: Workspace events

  Every event the server triggers in a workspace is stored, whether or not an
  integration is subscribed to it. The log is for owners and admins: ordinary
  members and access keys without the event grants cannot read it.

  Routes (shapes in openapi.yaml):
    GET /api/v1/workspaces/{workspaceId}/events
    GET /api/v1/workspaces/{workspaceId}/events/{eventId}

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
