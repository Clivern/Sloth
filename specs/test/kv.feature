Feature: Workspace key-value store

  Each workspace has a small key-value store for agents to share configuration.
  Keys are path-like (a/b/c), values are strings, and a key can expire. An
  expired key must behave as if it was never written.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/kv?prefix=<p>
    GET    /api/v1/workspaces/{workspaceId}/kv/{key...}
    PUT    /api/v1/workspaces/{workspaceId}/kv/{key...}
    DELETE /api/v1/workspaces/{workspaceId}/kv/{key...}

  Keys use letters, digits, underscore, dot, dash and slash, up to 200
  characters, and no segment may be empty, "." or "..". Put every key under
  "spec/<runId>/".

  Background:
    Given the run's workspace exists

  # ---------------------------------------------------------------------------
  # Writing and reading
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: A written key reads back
    When the user PUTs kv/spec/<runId>/greeting with value "hello"
    Then the response is 200
    And key is "spec/<runId>/greeting" and value is "hello"
    And reading kv/spec/<runId>/greeting returns value "hello"

  @mutates
  Scenario: Writing a key again replaces its value
    Given kv/spec/<runId>/greeting is "hello"
    When the user PUTs it with value "bonjour"
    Then reading it returns "bonjour"
    And the list holds the key once

  @readonly
  Scenario: A value is required
    When the user PUTs kv/spec/<runId>/empty with body {}
    Then the response is 400

  @readonly
  Scenario: Keys with empty segments or illegal characters are rejected
    When the user PUTs kv/spec/<runId>//double with value "x"
    Then the response is 400
    When the user PUTs kv/spec/<runId>/has%20space with value "x"
    Then the response is 400

  @readonly
  Scenario: Reading a key that was never written returns not found
    When the user GETs kv/spec/<runId>/missing
    Then the response is 404

  # ---------------------------------------------------------------------------
  # Listing
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: The prefix filter returns only keys under it
    Given kv/spec/<runId>/app/a, kv/spec/<runId>/app/b and kv/spec/<runId>/other are written
    When the user lists kv with prefix "spec/<runId>/app/"
    Then the items are exactly "spec/<runId>/app/a" and "spec/<runId>/app/b"

  # ---------------------------------------------------------------------------
  # Expiry and deletion
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A key can carry an expiry
    When the user PUTs kv/spec/<runId>/temp with value "x" and expiresAt one hour from now in RFC 3339
    Then expiresAt is that time in UTC

  @readonly
  Scenario: An expiry that is not RFC 3339 is rejected
    When the user PUTs kv/spec/<runId>/temp with value "x" and expiresAt "soon"
    Then the response is 400

  @mutates @slow
  Scenario: An expired key is gone
    Given kv/spec/<runId>/temp expires 5 seconds from now
    When 10 seconds pass
    Then reading it returns 404
    And listing with prefix "spec/<runId>/temp" does not include it

  @mutates
  Scenario: A deleted key is gone
    Given kv/spec/<runId>/greeting is "hello"
    When the user DELETEs it
    Then the response is 204
    And reading it returns 404
    And deleting it again returns 404

  # ---------------------------------------------------------------------------
  # Isolation and access
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: Keys do not leak between workspaces
    Given kv/spec/<runId>/secret is "a" in the run's workspace
    And a second workspace "spec-<runId>-other"
    When the user GETs kv/spec/<runId>/secret in the second workspace
    Then the response is 404

  @mutates
  Scenario: An access key with read grants cannot write
    Given an access key with only CAN_LIST_KEY_VALUE and CAN_GET_KEY_VALUE
    And kv/spec/<runId>/greeting is "hello"
    When that key reads it
    Then the response is 200
    When that key PUTs it with value "hacked"
    Then the response is 403
    And the value is still "hello"

  # ---------------------------------------------------------------------------
  # More on keys and expiry
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A key may be up to 200 characters, not more
    When the user PUTs a key of exactly 200 characters under "spec/<runId>/" with value "x"
    Then the response is 200
    When the user PUTs a key of 201 characters
    Then the response is 400

  @readonly
  Scenario: "." and ".." are not key segments
    When the user PUTs kv/spec/<runId>/./x with value "x", sending the path as is
    Then the response is 400 or 404
    When the user PUTs kv/spec/<runId>/../x with value "x", sending the path as is
    Then the response is 400 or 404

  @mutates
  Scenario: A key whose expiry has already passed is gone at once
    When the user PUTs kv/spec/<runId>/stale with value "x" and expiresAt one hour ago
    Then the response is 200 or 400
    And reading it returns 404

  @mutates
  Scenario: Writing a key without an expiry clears an old one
    Given kv/spec/<runId>/temp expires one hour from now
    When the user PUTs it again with value "y" and no expiresAt
    Then reading it returns value "y" and no expiresAt

  @mutates
  Scenario: Values keep their whitespace and newlines
    When the user PUTs kv/spec/<runId>/multi with value "  line one\nline two  "
    Then reading it returns exactly "  line one\nline two  "
