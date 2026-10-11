Feature: Gateway traffic

  Every call the gateway forwards between agents is recorded so a workspace can
  see who called whom. Members can read it; it cannot be written through the
  API.

  Routes (shapes in openapi.yaml):
    GET /api/v1/workspaces/{workspaceId}/traffic
    GET /api/v1/workspaces/{workspaceId}/traffic/{trafficId}

  Note: the A2A gateway is not implemented yet (see pending.feature), so
  nothing records traffic and a new workspace's list is empty. When the gateway
  ships, add scenarios that make a call and find it here.

  Background:
    Given the run's workspace exists

  @readonly @smoke
  Scenario: A member can list the workspace's traffic
    When the user lists the run workspace's traffic
    Then the response is 200
    And calls is an array
    And _meta has limit, offset and total

  @readonly
  Scenario: Reading an unknown call returns not found
    When the user GETs traffic/<a random UUID>
    Then the response is 404

  @mutates
  Scenario: An access key needs the traffic grant
    Given an access key with only CAN_LIST_AGENTS
    When that key lists the run workspace's traffic
    Then the response is 403
    Given an access key with only CAN_LIST_TRAFFIC
    When that key lists the run workspace's traffic
    Then the response is 200
