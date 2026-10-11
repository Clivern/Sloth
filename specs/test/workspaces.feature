Feature: Workspaces

  A workspace holds a team's agents, keys and rules. Whoever creates one owns
  it. A workspace is visible only to its members, and the id of a workspace
  you don't belong to must not tell you anything about it.

  Routes (shapes in openapi.yaml):
    POST   /api/v1/workspaces
    GET    /api/v1/workspaces
    GET    /api/v1/workspaces/{workspaceId}
    PUT    /api/v1/workspaces/{workspaceId}
    DELETE /api/v1/workspaces/{workspaceId}

  # ---------------------------------------------------------------------------
  # Creating
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: Creating a workspace makes the creator its owner
    When the user POSTs a workspace named "spec-<runId>-team"
    Then the response is 201
    And the body has id, name "spec-<runId>-team", handle, createdAt and updatedAt
    And role is "owner"
    And count is 1

  @readonly
  Scenario: A workspace name must be 3 to 60 characters
    When the user POSTs a workspace named "ab"
    Then the response is 400
    When the user POSTs a workspace with a 61-character name
    Then the response is 400

  @mutates
  Scenario: Two workspaces with the same name get different handles
    Given the user created a workspace named "spec-<runId>-twin"
    When the user POSTs another workspace named "spec-<runId>-twin"
    Then the response is 201
    And the two workspaces have different ids and different handles

  # ---------------------------------------------------------------------------
  # Reading
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A new workspace appears in the user's list
    Given the user created a workspace named "spec-<runId>-team"
    When the user lists workspaces with limit 100
    Then the workspaces array contains "spec-<runId>-team" with role "owner"

  @mutates
  Scenario: A member can read their workspace
    Given the user created a workspace named "spec-<runId>-team"
    When the user GETs that workspace
    Then the response is 200
    And the body matches what creation returned

  @readonly
  Scenario: A workspace that does not exist returns not found
    When the user GETs /api/v1/workspaces/<a random UUID>
    Then the response is 404

  # ---------------------------------------------------------------------------
  # Updating and deleting
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: The owner can rename a workspace
    Given the user created a workspace named "spec-<runId>-team"
    When the user PUTs that workspace with name "spec-<runId>-renamed"
    Then the response is 200
    And name is "spec-<runId>-renamed"
    And the handle is unchanged

  @mutates
  Scenario: Renaming follows the same name rules as creating
    Given the user created a workspace named "spec-<runId>-team"
    When the user PUTs that workspace with name "ab"
    Then the response is 400

  @mutates
  Scenario: Deleting a workspace removes it and everything in it
    Given the user created a workspace named "spec-<runId>-doomed"
    And it has an agent "spec-<runId>-echo" and a kv key "spec/<runId>"
    When the user DELETEs that workspace
    Then the response is 204
    And reading the workspace returns 404
    And it is gone from the user's list

  # ---------------------------------------------------------------------------
  # Access keys and workspaces
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: An access key only works in its own workspace
    Given the user created workspaces "spec-<runId>-a" and "spec-<runId>-b"
    And an access key in "spec-<runId>-a" with only CAN_GET_WORKSPACE
    When that key GETs workspace "spec-<runId>-a"
    Then the response is 200
    When that key GETs workspace "spec-<runId>-b"
    Then the response is 403

  @mutates
  Scenario: An access key cannot create or list workspaces
    Given the run's workspace exists
    And an access key with every permission an access key may hold
    When that key POSTs a workspace named "spec-<runId>-sneaky"
    Then the response is 401 or 403
    When that key lists workspaces
    Then the response is 401 or 403

  # ---------------------------------------------------------------------------
  # More on access keys and deletion
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: An access key reading its workspace gets no role
    Given an access key with only CAN_GET_WORKSPACE
    When that key GETs the run's workspace
    Then the response is 200
    And role is ""

  @mutates
  Scenario: Deleting a workspace revokes its access keys
    Given the user created a workspace named "spec-<runId>-doomed"
    And an access key in it with only CAN_LIST_AGENTS
    When the user DELETEs that workspace
    Then that key listing agents anywhere gets 401

  @mutates
  Scenario: An access key with the delete grant can delete its own workspace
    Given the user created a workspace named "spec-<runId>-self-delete"
    And an access key in it with only CAN_DELETE_WORKSPACE
    When that key DELETEs that workspace
    Then the response is 204
    And the user reading it gets 404

  @mutates
  Scenario: Paging through workspaces never repeats or skips one
    Given the user created workspaces "spec-<runId>-p1", "spec-<runId>-p2" and "spec-<runId>-p3"
    When the user pages through /api/v1/workspaces with limit 1
    Then every workspace appears exactly once across the pages
    And the number of pages equals _meta.total
