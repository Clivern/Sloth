Feature: Agent protection

  Protection is the credential Cognit's gateway presents to an agent when it
  forwards a call: an API key in a header, or basic auth. An agent has at most
  one. If Cognit generates the secret it is shown once; a secret the user
  supplied is never echoed back.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/agents/{agentName}/protection
    PUT    /api/v1/workspaces/{workspaceId}/agents/{agentName}/protection
    PUT    /api/v1/workspaces/{workspaceId}/agents/{agentName}/protection/status
    DELETE /api/v1/workspaces/{workspaceId}/agents/{agentName}/protection

  Background:
    Given the run's workspace exists
    And the agent "spec-<runId>-echo" is registered

  # ---------------------------------------------------------------------------
  # Setting
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: An api_key protection with no secret gets a generated one
    When the user PUTs protection with type "api_key"
    Then the response is 200
    And type is "api_key", header is "X-API-Key", isActive is true and secret is not empty

  @mutates
  Scenario: A supplied secret is stored but never echoed
    When the user PUTs protection with type "api_key", header "X-Agent-Key" and secret "spec-<runId>-s3cret"
    Then the response is 200
    And header is "X-Agent-Key"
    And the body has no secret field

  @mutates
  Scenario: Basic auth protection needs a username
    When the user PUTs protection with type "basic_auth" and no username
    Then the response is 400
    When the user PUTs protection with type "basic_auth" and username "gateway"
    Then the response is 200
    And username is "gateway"

  @readonly
  Scenario: Only basic_auth protection takes a username
    When the user PUTs protection with type "api_key" and username "gateway"
    Then the response is 400

  @readonly
  Scenario: A supplied secret must be at least 8 characters
    When the user PUTs protection with type "api_key" and secret "short"
    Then the response is 400

  @mutates
  Scenario: Saving again replaces the protection and turns it on
    Given an api_key protection that has been turned off
    When the user PUTs protection with type "basic_auth" and username "gateway"
    Then type is "basic_auth" and isActive is true
    And reading the protection returns the same

  # ---------------------------------------------------------------------------
  # Reading, toggling and removing
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: Reading the protection never shows a secret
    Given an api_key protection
    When the user GETs protection
    Then the response is 200
    And the body has no secret field

  @mutates
  Scenario: An agent without protection returns not found
    When the user GETs protection
    Then the response is 404

  @mutates
  Scenario: Protection can be turned off and on
    Given an api_key protection
    When the user PUTs protection/status with isActive false
    Then the response is 200 and isActive is false
    When the user PUTs protection/status with isActive true
    Then isActive is true

  @mutates
  Scenario: Turning off protection that does not exist returns not found
    When the user PUTs protection/status with isActive false
    Then the response is 404

  @mutates
  Scenario: Removing protection deletes it
    Given an api_key protection
    When the user DELETEs protection
    Then the response is 204
    And reading it returns 404

  @mutates
  Scenario: Protection on an unknown agent returns not found
    When the user PUTs protection with type "api_key" under agents/spec-<runId>-ghost
    Then the response is 404

  @mutates
  Scenario: Reading protection needs the agent read grant; changing it needs upsert
    Given an api_key protection
    And an access key with only CAN_GET_AGENT
    When that key GETs protection
    Then the response is 200
    When that key DELETEs protection
    Then the response is 403
