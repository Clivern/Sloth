Feature: Workspace access keys

  An access key lets a machine act inside one workspace with a fixed list of
  permissions, without being a user. The secret is shown once. A key can never
  hold a permission its creator lacks, and it can never reach past its own
  workspace or its own list.

  Routes (shapes in openapi.yaml):
    POST   /api/v1/workspaces/{workspaceId}/keys
    GET    /api/v1/workspaces/{workspaceId}/keys
    GET    /api/v1/workspaces/{workspaceId}/keys/{keyId}
    DELETE /api/v1/workspaces/{workspaceId}/keys/{keyId}

  Permissions an access key may hold are the CAN_* names in module/permission.go
  (AccessKeyPermissions).

  Background:
    Given the run's workspace exists

  # ---------------------------------------------------------------------------
  # Creating
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: A new key is returned once with its secret and permissions
    When the user POSTs a key named "spec-<runId>-deploy" with permissions
      ["CAN_LIST_AGENTS", "CAN_REGISTER_INSTANCE"]
    Then the response is 201
    And the body has id, workspaceId, name, createdAt and a non-empty key
    And permissions is exactly ["CAN_LIST_AGENTS", "CAN_REGISTER_INSTANCE"]

  @readonly
  Scenario: A key needs at least one permission
    When the user POSTs a key named "spec-<runId>-empty" with permissions []
    Then the response is 400
    And no key named "spec-<runId>-empty" exists afterwards

  @readonly
  Scenario: An unknown permission is rejected
    When the user POSTs a key named "spec-<runId>-bad" with permissions ["CAN_EVERYTHING"]
    Then the response is 400
    And no key named "spec-<runId>-bad" exists afterwards

  @readonly
  Scenario: Inviting members is not a permission an access key can hold
    When the user POSTs a key named "spec-<runId>-inviter" with permissions ["CAN_INVITE_MEMBER"]
    Then the response is 400
    And no key named "spec-<runId>-inviter" exists afterwards

  # An invite records the user who sent it and emails in their name, so it needs a person.

  @readonly
  Scenario: A key needs a name of at most 60 characters
    When the user POSTs a key with permissions [] and no name
    Then the response is 400
    When the user POSTs a key with a 61-character name
    Then the response is 400

  @readonly
  Scenario: An expiry that is not RFC 3339 is rejected
    When the user POSTs a key named "spec-<runId>-bad" with expiresAt "next week"
    Then the response is 400

  # ---------------------------------------------------------------------------
  # No escalation
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: An access key can mint a key with a subset of its own permissions
    Given an access key with only CAN_CREATE_WORKSPACE_ACCESS_KEY and CAN_LIST_AGENTS
    When that key POSTs a key named "spec-<runId>-child" with permissions ["CAN_LIST_AGENTS"]
    Then the response is 201

  @mutates
  Scenario: An access key cannot mint a key with a permission it lacks
    Given an access key with only CAN_CREATE_WORKSPACE_ACCESS_KEY and CAN_LIST_AGENTS
    When that key POSTs a key named "spec-<runId>-child" with permissions ["CAN_DELETE_AGENT"]
    Then the response is 403
    And no key named "spec-<runId>-child" exists afterwards

  # ---------------------------------------------------------------------------
  # Reading and deleting
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: Listing and reading a key never shows its secret
    Given an access key named "spec-<runId>-deploy"
    When the user lists the workspace's keys
    Then the list contains "spec-<runId>-deploy"
    And no item in the list has a key field
    When the user GETs that key by id
    Then the response is 200
    And the body has no key field

  @mutates
  Scenario: A deleted key stops working at once
    Given an access key with only CAN_LIST_AGENTS
    When the user DELETEs that key
    Then the response is 204
    And that key listing the workspace's agents gets 401
    And reading the key by id returns 404

  @readonly
  Scenario: Reading an unknown key returns not found
    When the user GETs keys/<a random UUID>
    Then the response is 404

  @mutates @slow
  Scenario: An expired key stops working
    Given an access key with only CAN_LIST_AGENTS expiring 10 seconds from now
    When 15 seconds pass
    And that key lists the workspace's agents
    Then the response is 401
