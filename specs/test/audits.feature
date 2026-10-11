Feature: Workspace audit log

  The audit log records who did what in a workspace. It is for owners and
  admins: ordinary members and access keys without the audit grants cannot
  read it.

  Routes (shapes in openapi.yaml):
    GET /api/v1/workspaces/{workspaceId}/audits
    GET /api/v1/workspaces/{workspaceId}/audits/{auditId}

  Note: no part of the server writes audit events yet, so a new workspace's log
  is empty. When events start being recorded, add scenarios here for what each
  action records.

  Background:
    Given the run's workspace exists

  @readonly @smoke
  Scenario: The owner can list the audit log
    When the user lists the run workspace's audits
    Then the response is 200
    And audits is an array
    And _meta has limit, offset and total

  @readonly
  Scenario: Reading an unknown audit event returns not found
    When the user GETs audits/<a random UUID>
    Then the response is 404

  @mutates
  Scenario: An access key needs the audit grant
    Given an access key with only CAN_LIST_AGENTS
    When that key lists the run workspace's audits
    Then the response is 403
    Given an access key with only CAN_LIST_WORKSPACE_AUDITS
    When that key lists the run workspace's audits
    Then the response is 200
