Feature: User API keys

  A user makes API keys to act as themselves from scripts and agents. The key
  is shown once, when it is made; after that only its name and dates are
  visible. Each key belongs to one user, and nobody else can list, read or
  revoke it.

  Routes (shapes in openapi.yaml):
    POST   /api/v1/apiKeys
    GET    /api/v1/apiKeys
    GET    /api/v1/apiKeys/{apiKeyId}
    DELETE /api/v1/apiKeys/{apiKeyId}

  Note: never revoke the key the run is using. Only touch keys named
  spec-<runId>-*.

  # ---------------------------------------------------------------------------
  # Creating
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: A new key is returned once with its secret
    When the user POSTs an API key named "spec-<runId>-ci"
    Then the response is 201
    And the body has id, name "spec-<runId>-ci", createdAt and a non-empty key
    And expiresAt is null

  @mutates
  Scenario: The new key works as a credential
    Given the user made an API key named "spec-<runId>-ci"
    When a request GETs /api/v1/me with that key
    Then the response is 200
    And owner.id is the user's id

  @mutates
  Scenario: A key can be given an expiry
    When the user POSTs an API key named "spec-<runId>-exp" with expiresAt one day from now in RFC 3339
    Then the response is 201
    And expiresAt is that time in UTC

  @readonly
  Scenario: An expiry that is not RFC 3339 is rejected
    When the user POSTs an API key named "spec-<runId>-bad" with expiresAt "tomorrow"
    Then the response is 400
    And no key named "spec-<runId>-bad" exists afterwards

  @readonly
  Scenario: A key needs a name
    When the user POSTs an API key with body {}
    Then the response is 400

  @mutates @slow
  Scenario: An expired key stops working
    Given the user made an API key named "spec-<runId>-short" expiring 10 seconds from now
    When 15 seconds pass
    And a request GETs /api/v1/me with that key
    Then the response is 401

  # ---------------------------------------------------------------------------
  # Reading
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: Listing and reading a key never shows its secret
    Given the user made an API key named "spec-<runId>-ci"
    When the user lists their API keys
    Then the list contains "spec-<runId>-ci"
    And no item in the list has a key field
    When the user GETs that key by id
    Then the response is 200
    And the body has no key field

  @readonly
  Scenario: Reading an unknown key returns not found
    When the user GETs /api/v1/apiKeys/<a random UUID>
    Then the response is 404

  # ---------------------------------------------------------------------------
  # Revoking
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A revoked key stops working at once
    Given the user made an API key named "spec-<runId>-ci"
    When the user DELETEs that key
    Then the response is 204
    And a request to /api/v1/me with that key returns 401
    And reading that key by id returns 404

  @mutates
  Scenario: A workspace access key cannot manage user API keys
    Given the run's workspace exists
    And an access key with every permission an access key may hold
    When that key lists API keys
    Then the response is 401 or 403

  @mutates
  Scenario: A key with an expiry in the past never works
    When the user POSTs an API key named "spec-<runId>-past" with expiresAt one hour ago
    Then the response is 201 or 400
    And if 201, a request to /api/v1/me with that key returns 401

  @mutates
  Scenario: The key list pages
    Given the user made API keys "spec-<runId>-p1" and "spec-<runId>-p2"
    When the user lists API keys with limit 1
    Then exactly one key comes back
    And _meta.total is at least 2
