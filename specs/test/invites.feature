Feature: Workspace invites

  Owners and admins invite people to a workspace by email with a role. An
  invite stays pending until it is accepted or revoked, and an address can only
  have one pending invite per workspace at a time.

  Routes (shapes in openapi.yaml):
    POST   /api/v1/workspaces/{workspaceId}/invites
    GET    /api/v1/workspaces/{workspaceId}/invites
    GET    /api/v1/workspaces/{workspaceId}/invites/{inviteId}
    DELETE /api/v1/workspaces/{workspaceId}/invites/{inviteId}

  Note: creating an invite may send an email. Use addresses at example.com,
  which never receive mail: spec-<runId>-<n>@example.com.

  Background:
    Given the run's workspace exists

  # ---------------------------------------------------------------------------
  # Creating
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: Inviting an address creates a pending invite
    When the user POSTs an invite for "spec-<runId>-1@example.com" with role "regular"
    Then the response is 201
    And the body has id, email "spec-<runId>-1@example.com", role "regular", status,
      inviterUserId, workspaceId and expiresAt
    And inviterUserId is the user's id
    And expiresAt is in the future

  @mutates
  Scenario: An address can only have one pending invite
    Given the user invited "spec-<runId>-1@example.com"
    When the user POSTs another invite for "spec-<runId>-1@example.com" with role "admin"
    Then the response is 409

  @mutates
  Scenario: A user who is already a member cannot be invited
    When the user POSTs an invite for their own email with role "regular"
    Then the response is 409

  @readonly
  Scenario: An invite needs a valid email and a known role
    When the user POSTs an invite for "not-an-email" with role "regular"
    Then the response is 400
    When the user POSTs an invite for "spec-<runId>-2@example.com" with role "owner"
    Then the response is 400

  # ---------------------------------------------------------------------------
  # Reading and revoking
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A pending invite appears in the list and can be read by id
    Given the user invited "spec-<runId>-1@example.com"
    When the user lists the run workspace's invites
    Then the invites array contains "spec-<runId>-1@example.com"
    When the user GETs that invite by id
    Then the response is 200
    And the body matches what creation returned

  @mutates
  Scenario: A revoked invite is gone
    Given the user invited "spec-<runId>-1@example.com"
    When the user DELETEs that invite
    Then the response is 204
    And reading it by id returns 404
    And the address can be invited again

  @readonly
  Scenario: Reading an unknown invite returns not found
    When the user GETs invites/<a random UUID>
    Then the response is 404

  @mutates
  Scenario: An invite from another workspace cannot be read through this one
    Given the user created a second workspace "spec-<runId>-other"
    And invited "spec-<runId>-3@example.com" to it
    When the user GETs that invite's id under the run's workspace
    Then the response is 404

  # ---------------------------------------------------------------------------
  # Access keys
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: An access key needs the invite grant to invite
    Given an access key with only CAN_LIST_WORKSPACE_INVITES
    When that key POSTs an invite for "spec-<runId>-4@example.com" with role "regular"
    Then the response is 403
    When that key lists the invites
    Then the response is 200
