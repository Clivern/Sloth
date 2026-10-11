Feature: Health checks

  An agent carries check templates, and every instance of it gets a copy of
  each one. A ttl check is kept alive by the instance reporting pass, warn or
  fail before its ttl runs out; an http or tcp check is probed by Cognit's
  worker. A ttl check that has not reported in time counts as critical — a
  silent instance must look dead, not healthy.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/agents/{agentName}/checks
    PUT    /api/v1/workspaces/{workspaceId}/agents/{agentName}/checks/{checkId}
    DELETE /api/v1/workspaces/{workspaceId}/agents/{agentName}/checks/{checkId}
    POST   /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}/checks/{checkId}/pass
    POST   /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}/checks/{checkId}/warn
    POST   /api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}/checks/{checkId}/fail

  Background:
    Given the run's workspace exists
    And the agent "spec-<runId>-echo" is registered

  # ---------------------------------------------------------------------------
  # Templates
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: A ttl check template is created
    When the user PUTs checks/heartbeat with type "ttl" and ttl 60
    Then the response is 201
    And the agent's checks list contains "heartbeat" with type "ttl"

  @mutates
  Scenario: Saving the same template again replaces it
    Given a ttl check template "heartbeat" with ttl 60
    When the user PUTs checks/heartbeat with type "ttl" and ttl 120
    Then the response is 200

  @mutates
  Scenario: A new template reaches instances that already exist
    Given instance "i-1" is registered
    When the user PUTs checks/heartbeat with type "ttl" and ttl 60
    Then instance "i-1" has a check "heartbeat" with source "agent"

  @mutates
  Scenario: A new instance gets every existing template
    Given a ttl check template "heartbeat" with ttl 60
    When the user registers instance "i-2"
    Then instance "i-2" has a check "heartbeat"

  @mutates
  Scenario: Deleting a template removes it from every instance
    Given a ttl check template "heartbeat" with ttl 60
    And instance "i-1" is registered
    When the user DELETEs checks/heartbeat
    Then the response is 204
    And instance "i-1" has no check "heartbeat"

  @readonly
  Scenario: A check type must be ttl, http or tcp
    When the user PUTs checks/bad with type "grpc"
    Then the response is 400

  @readonly
  Scenario: A ttl check needs a ttl of 5 to 86400 seconds
    When the user PUTs checks/bad with type "ttl" and no ttl
    Then the response is 400
    When the user PUTs checks/bad with type "ttl" and ttl 4
    Then the response is 400

  @readonly
  Scenario: An http check needs a path starting with a slash
    When the user PUTs checks/web with type "http", port 9000 and path "health"
    Then the response is 400
    When the user PUTs checks/web with type "http", port 9000 and path "/he alth"
    Then the response is 400

  @readonly
  Scenario: A probe's timeout must be shorter than its interval
    When the user PUTs checks/web with type "tcp", port 9000, interval 10 and timeout 10
    Then the response is 400

  @mutates
  Scenario: An http check defaults to the http scheme
    When the user PUTs checks/web with type "http", port 9000 and path "/health"
    Then the response is 201 or 200

  @readonly
  Scenario: A template for an unknown agent returns not found
    When the user PUTs checks/heartbeat under agents/spec-<runId>-ghost with type "ttl" and ttl 60
    Then the response is 404

  # ---------------------------------------------------------------------------
  # Reporting
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A ttl check that has not reported yet is unknown, not passing
    Given a ttl check template "heartbeat" with ttl 60
    When the user registers instance "i-1"
    Then reading "i-1" shows check "heartbeat" as "unknown" with output "Waiting for the first report"
    And its ttlExpiresAt is about 60 seconds from now
    And the instance status is "unknown"
    And listing instances with passing=true leaves out "i-1"

  @mutates @smoke
  Scenario: Reporting pass makes the check and the instance passing
    Given a ttl check template "heartbeat" with ttl 60
    And instance "i-1" is registered
    When the user POSTs heartbeat/pass for "i-1" with output "ok"
    Then the response is 200
    And status is "passing" and ttlExpiresAt is about 60 seconds from now
    And reading "i-1" shows status "passing"

  @mutates
  Scenario: Reporting warn makes the instance warning
    Given a ttl check template "heartbeat" with ttl 60 and instance "i-1"
    When the user POSTs heartbeat/warn for "i-1"
    Then status is "warning"
    And reading "i-1" shows status "warning"

  @mutates
  Scenario: Reporting fail makes the instance critical and drops it from passing
    Given a ttl check template "heartbeat" with ttl 60 and instance "i-1" reporting pass
    When the user POSTs heartbeat/fail for "i-1"
    Then status is "critical"
    And listing instances with passing=true leaves out "i-1"

  @mutates
  Scenario: Only ttl checks can be reported
    Given a tcp check template "port" on port 9000 and instance "i-1"
    When the user POSTs port/pass for "i-1"
    Then the response is 400

  @mutates
  Scenario: Reporting a check the instance does not have returns not found
    Given instance "i-1" is registered
    When the user POSTs spec-<runId>-nope/pass for "i-1"
    Then the response is 404

  @mutates @slow
  Scenario: A ttl check that stops reporting turns critical
    Given a ttl check template "heartbeat" with ttl 5 and instance "i-1" reporting pass
    When 10 seconds pass
    Then reading "i-1" shows check "heartbeat" as "critical"
    And the instance status is "critical"

  @mutates
  Scenario: An instance's own access key can report but not edit templates
    Given an access key with only CAN_REPORT_CHECK
    And a ttl check template "heartbeat" with ttl 60 and instance "i-1"
    When that key POSTs heartbeat/pass for "i-1"
    Then the response is 200
    When that key PUTs checks/heartbeat with type "ttl" and ttl 30
    Then the response is 403
