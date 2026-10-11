Feature: Who am I

  Every authenticated request is made by exactly one principal: a user through
  an API key or a session, or a workspace access key. /me tells a caller which
  one the server sees. A client uses it to check its credentials before doing
  anything else, so it must agree with what the rest of the API enforces.

  Routes (shapes in openapi.yaml):
    GET /api/v1/me

  # ---------------------------------------------------------------------------
  # Each kind of credential
  # ---------------------------------------------------------------------------

  @readonly @smoke
  Scenario: A user API key resolves to its owner
    When the user GETs /api/v1/me with X-API-Key
    Then the response is 200
    And principalType is "api_key"
    And owner has an id, email and role
    And apiKey has an id and name
    And the body does not contain the API key itself

  @mutates
  Scenario: A workspace access key resolves to that key and its permissions
    Given the run's workspace exists
    And an access key with only CAN_LIST_AGENTS
    When that key GETs /api/v1/me with X-Access-Key
    Then the response is 200
    And principalType is "access_key"
    And accessKey.workspaceId is the run's workspace id
    And accessKey.permissions is exactly ["CAN_LIST_AGENTS"]

  # ---------------------------------------------------------------------------
  # Bad or missing credentials
  # ---------------------------------------------------------------------------

  @readonly @smoke
  Scenario: No credentials is unauthenticated
    When a request with no credentials GETs /api/v1/me
    Then the response is 401

  @readonly
  Scenario: An unknown API key is refused
    When a request GETs /api/v1/me with X-API-Key "spec-<runId>-not-a-key"
    Then the response is 401
    And the errorMessage is "Invalid API key."

  @readonly
  Scenario: An unknown access key is refused
    When a request GETs /api/v1/me with X-Access-Key "spec-<runId>-not-a-key"
    Then the response is 401

  @readonly
  Scenario: A garbage session cookie is refused
    When a request GETs /api/v1/me with cookie _cognit_session "spec-<runId>-garbage"
    Then the response is 401
