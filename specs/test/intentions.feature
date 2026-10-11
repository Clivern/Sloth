Feature: Intentions

  Intentions decide which agent may call which. Each rule names a source agent,
  a destination agent, optionally one skill, and allow or deny. A call with no
  matching rule is denied, and a rule for the exact skill beats a rule for the
  whole agent. Getting this wrong opens or closes the gateway for every caller.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/intentions
    PUT    /api/v1/workspaces/{workspaceId}/intentions
    POST   /api/v1/workspaces/{workspaceId}/intentions/check
    GET    /api/v1/workspaces/{workspaceId}/intentions/{intentionId}
    PUT    /api/v1/workspaces/{workspaceId}/intentions/{intentionId}
    DELETE /api/v1/workspaces/{workspaceId}/intentions/{intentionId}

  A rule body is: {"source": {"agent": "<a>"}, "destination": {"agent": "<b>",
  "skill": "<optional>"}, "action": "allow" | "deny"}. Agents named here do
  not need to be registered.

  Background:
    Given the run's workspace exists

  # ---------------------------------------------------------------------------
  # Rules
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: Creating a rule returns it
    When the user PUTs intentions with source "spec-<runId>-a", destination "spec-<runId>-b" and action "allow"
    Then the response is 201
    And source is "spec-<runId>-a", destination is "spec-<runId>-b" and action is "allow"

  @readonly
  Scenario: The action must be allow or deny
    When the user PUTs intentions with source "a", destination "b" and action "maybe"
    Then the response is 400

  @readonly
  Scenario: A rule needs a source and a destination agent
    When the user PUTs intentions with destination "b", action "allow" and no source
    Then the response is 400
    When the user PUTs intentions with source "a", action "allow" and no destination
    Then the response is 400

  @mutates
  Scenario: A rule can be read, updated and deleted
    Given a rule allowing "spec-<runId>-a" to call "spec-<runId>-b"
    When the user GETs it by id
    Then the response is 200
    When the user PUTs it with action "deny"
    Then the response is 200 and action is "deny"
    When the user DELETEs it
    Then the response is 204
    And reading it returns 404

  @mutates
  Scenario: A rule from another workspace cannot be reached through this one
    Given a second workspace "spec-<runId>-other" with a rule in it
    When the user GETs that rule's id under the run's workspace
    Then the response is 404
    When the user DELETEs that rule's id under the run's workspace
    Then the response is 404

  # ---------------------------------------------------------------------------
  # Checking a call
  # ---------------------------------------------------------------------------

  @readonly @smoke
  Scenario: A call with no matching rule is denied
    When the user POSTs intentions/check with source "spec-<runId>-x" and destination "spec-<runId>-y"
    Then the response is 200
    And action is "deny"
    And there is no intention in the body

  @mutates
  Scenario: An agent-wide allow lets every skill through
    Given a rule allowing "spec-<runId>-a" to call "spec-<runId>-b"
    When the user checks a call from "spec-<runId>-a" to "spec-<runId>-b" skill "search"
    Then action is "allow"
    And intention is the agent-wide rule

  @mutates
  Scenario: A skill rule beats an agent-wide rule
    Given a rule allowing "spec-<runId>-a" to call "spec-<runId>-b"
    And a rule denying "spec-<runId>-a" to call "spec-<runId>-b" skill "delete"
    When the user checks a call from "spec-<runId>-a" to "spec-<runId>-b" skill "delete"
    Then action is "deny"
    And intention.skill is "delete"
    When the user checks a call from "spec-<runId>-a" to "spec-<runId>-b" skill "search"
    Then action is "allow"

  @mutates
  Scenario: Rules are one-way
    Given a rule allowing "spec-<runId>-a" to call "spec-<runId>-b"
    When the user checks a call from "spec-<runId>-b" to "spec-<runId>-a"
    Then action is "deny"

  @mutates
  Scenario: A skill rule does not cover other skills or the agent as a whole
    Given a rule allowing "spec-<runId>-a" to call "spec-<runId>-b" skill "search"
    When the user checks a call from "spec-<runId>-a" to "spec-<runId>-b" with no skill
    Then action is "deny"

  # ---------------------------------------------------------------------------
  # Access
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: An access key with only the check grant can check but not change rules
    Given an access key with only CAN_CHECK_INTENTION
    When that key POSTs intentions/check with source "a" and destination "b"
    Then the response is 200
    When that key PUTs intentions with source "a", destination "b" and action "allow"
    Then the response is 403

  # ---------------------------------------------------------------------------
  # More validation and listing
  # ---------------------------------------------------------------------------

  @readonly
  Scenario: A skill may be at most 120 characters
    When the user PUTs a rule with a 121-character destination skill
    Then the response is 400

  @mutates
  Scenario: Updating a rule follows the same rules as creating one
    Given a rule allowing "spec-<runId>-a" to call "spec-<runId>-b"
    When the user PUTs it with action "maybe"
    Then the response is 400
    And reading it still returns action "allow"

  @readonly
  Scenario: Updating an unknown rule returns not found
    When the user PUTs intentions/<a random UUID> with a valid rule
    Then the response is 404

  @mutates
  Scenario: The rule list pages
    Given three rules exist in the run's workspace
    When the user lists intentions with limit 2
    Then exactly 2 come back
    And _meta.total is 3
