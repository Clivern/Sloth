Feature: Workspace members

  Members are the users who belong to a workspace, each with a role: owner,
  admin, regular or readonly. Owners and admins manage the team; everyone else
  can only read. A workspace must always keep its owner, or nobody is left who
  can delete or recover it.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/members
    PUT    /api/v1/workspaces/{workspaceId}/members/{memberUserId}
    DELETE /api/v1/workspaces/{workspaceId}/members/{memberUserId}

  Note: the run has a single user, so these scenarios cover what one user can
  check. Adding a second member means accepting an invite, which happens at
  login and is outside the API.

  Background:
    Given the run's workspace exists

  @mutates @smoke
  Scenario: The creator is listed as the owner
    When the user lists the run workspace's members
    Then the response is 200
    And members contains exactly one entry
    And its userId is the user's id, its role is "owner" and it has name and email
    And _meta.total is 1

  @mutates
  Scenario: A role must be one of admin, regular or readonly
    When the user PUTs their own membership with role "superuser"
    Then the response is 400
    And the user is still "owner"

  @mutates
  Scenario: Changing the role of someone who is not a member returns not found
    When the user PUTs members/<a random UUID> with role "readonly"
    Then the response is 404

  @mutates
  Scenario: Removing someone who is not a member returns not found
    When the user DELETEs members/<a random UUID>
    Then the response is 404

  @mutates
  Scenario: The owner cannot demote themselves
    Given the user created a separate workspace "spec-<runId>-owner-check"
    When the user PUTs their own membership in it with role "admin"
    Then the response is 400 or 409
    And the user is still "owner" of it

  # "admin" rather than "readonly": if the demotion wrongly succeeds, an admin
  # can still delete the workspace during cleanup, while readonly could not.

  @mutates
  Scenario: The owner cannot remove themselves
    Given the user created a separate workspace "spec-<runId>-owner-check"
    When the user DELETEs their own membership in it
    Then the response is 400 or 409
    And the workspace still lists the user as "owner"

  @mutates
  Scenario: An access key with the member grants can read the team but not see past it
    Given an access key with only CAN_LIST_WORKSPACE_MEMBERS
    When that key lists the run workspace's members
    Then the response is 200
    When that key PUTs the user's membership with role "admin"
    Then the response is 403

  @mutates
  Scenario: An access key with the role grant cannot change the owner either
    Given an access key with only CAN_UPDATE_MEMBER_ROLE
    When that key PUTs the owner's membership with role "admin"
    Then the response is 409
    And the user is still "owner"

  @mutates
  Scenario: An access key with the remove grant cannot remove the owner either
    Given an access key with only CAN_REMOVE_MEMBER
    When that key DELETEs the owner's membership
    Then the response is 409
    And the user is still "owner"

  @mutates
  Scenario: Refusing to change the owner explains why
    When the user PUTs their own membership with role "admin"
    Then the errorMessage is "The workspace owner cannot be changed or removed."
