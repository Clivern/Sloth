Feature: Routes that are not built yet

  Discovery, the skill catalog, the A2A gateway and catalog watching are routed
  but not built. Until they are, each must say so plainly with 501, and must
  still be behind authentication — an unbuilt route must not be an open door.
  When one of these ships, move its scenarios into a feature file of its own.

  Routes:
    GET  /api/v1/workspaces/{workspaceId}/discover
    POST /api/v1/workspaces/{workspaceId}/discover/resolve
    GET  /api/v1/workspaces/{workspaceId}/catalog/skills
    GET  /api/v1/workspaces/{workspaceId}/catalog/skills/{skill}/agents
    GET  /api/v1/workspaces/{workspaceId}/catalog/agents/{agentName}/card
    POST /api/v1/workspaces/{workspaceId}/a2a/agents/{agentName}/message:send
    POST /api/v1/workspaces/{workspaceId}/a2a/skills/{skill}/message:send
    GET  /api/v1/workspaces/{workspaceId}/a2a/tasks/{taskId}
    GET  /api/v1/workspaces/{workspaceId}/watch/catalog

  Background:
    Given the run's workspace exists

  @readonly @smoke
  Scenario: Each unbuilt route answers 501 Not Implemented
    When the user calls each route above with its method
    Then every response is 501
    And every errorMessage is "Not implemented."

  @readonly
  Scenario: Unbuilt routes still require credentials
    When a request with no credentials calls each route above
    Then every response is 401

  @mutates
  Scenario: Unbuilt routes still check the workspace grant
    Given an access key with only CAN_LIST_KEY_VALUE
    When that key calls each route above
    Then every response is 403
