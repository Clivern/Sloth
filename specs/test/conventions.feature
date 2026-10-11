Feature: API conventions

  Rules every route under /api/v1 follows, whatever the resource: how errors
  are shaped, which language they are in, how lists page, and what a bad id or
  an unknown route returns. A client is written once against these rules, so
  one route breaking them breaks every client quietly.

  Routes: any route under /api/v1; the scenarios below name the ones they use.

  Background:
    Given the run's workspace exists

  # ---------------------------------------------------------------------------
  # Error shape and language
  # ---------------------------------------------------------------------------

  @readonly @smoke
  Scenario: Every error body carries an errorMessage
    When the user GETs agents/spec-<runId>-missing in the run's workspace
    Then the response is 404
    And the body is JSON with a non-empty string errorMessage

  @readonly
  Scenario: Errors are in English by default
    When the user GETs agents/spec-<runId>-missing with no Accept-Language
    Then the errorMessage is "Agent not found."

  @readonly
  Scenario: Errors follow Accept-Language when the language is supported
    When the user GETs agents/spec-<runId>-missing with Accept-Language "fr-FR,fr;q=0.9"
    Then the errorMessage is "Agent introuvable."

  @readonly
  Scenario: An unsupported language falls back to English
    When the user GETs agents/spec-<runId>-missing with Accept-Language "de-DE"
    Then the errorMessage is "Agent not found."

  @readonly
  Scenario: A malformed JSON body is a 400, not a 500
    When the user PUTs kv/spec-<runId>/broken with the raw body "{not json"
    Then the response is 400
    And the body has an errorMessage

  # ---------------------------------------------------------------------------
  # Lists
  # ---------------------------------------------------------------------------

  @readonly
  Scenario: Lists report the paging they applied
    When the user GETs /api/v1/workspaces?limit=5&offset=0
    Then the response is 200
    And _meta has limit 5, offset 0 and a numeric total

  @readonly
  Scenario: A missing limit defaults to 50
    When the user GETs /api/v1/workspaces
    Then _meta.limit is 50

  @readonly
  Scenario: An out-of-range limit falls back to the default
    When the user GETs /api/v1/workspaces?limit=1000
    Then _meta.limit is 50
    When the user GETs /api/v1/workspaces?limit=-1
    Then _meta.limit is 50

  @readonly
  Scenario: A negative offset falls back to zero
    When the user GETs /api/v1/workspaces?offset=-5
    Then _meta.offset is 0

  # ---------------------------------------------------------------------------
  # Ids and unknown routes
  # ---------------------------------------------------------------------------

  @readonly
  Scenario: A workspace id that is not a UUID is not a server error
    When the user GETs /api/v1/workspaces/not-a-uuid
    Then the response is 400 or 404
    And the response is not 500

  @readonly
  Scenario: A resource id that is not a UUID is not a server error
    When the user GETs keys/not-a-uuid, invites/not-a-uuid, audits/not-a-uuid,
      traffic/not-a-uuid and intentions/not-a-uuid in the run's workspace
    Then each response is 400 or 404
    And none of them is 500

  @readonly
  Scenario: An unknown API route returns a JSON 404, not the dashboard page
    When the user GETs /api/v1/spec-<runId>-no-such-route
    Then the response is 404
    And the Content-Type is application/json

  # ---------------------------------------------------------------------------
  # Headers, methods and bodies
  # ---------------------------------------------------------------------------

  @readonly
  Scenario: Every response carries a request id
    When the user GETs /api/v1/me
    Then the X-Request-ID header is a UUID

  @readonly
  Scenario: A request id sent by the client is echoed back
    When the user GETs /api/v1/me with X-Request-ID "spec-<runId>-trace"
    Then the X-Request-ID header is "spec-<runId>-trace"

  @readonly
  Scenario: API responses are JSON, errors included
    When the user GETs /api/v1/me
    Then the Content-Type is application/json
    When the user GETs agents/spec-<runId>-missing in the run's workspace
    Then the Content-Type is application/json

  @readonly
  Scenario: A known route called with the wrong method is 405, not the dashboard
    When the user sends PATCH to /api/v1/me
    Then the response is 405
    And the body is not the dashboard HTML

  @readonly
  Scenario: A body over 10 MB is refused without a server error
    When the user PUTs kv/spec/<runId>/huge with a JSON body of 11 MB
    Then the response is 400 or 413
    And reading kv/spec/<runId>/huge returns 404

  @readonly
  Scenario: An offset past the end returns an empty page with the real total
    When the user GETs /api/v1/workspaces?offset=10000
    Then the response is 200
    And workspaces is empty
    And _meta.total equals the total from /api/v1/workspaces

  @mutates
  Scenario: Text comes back exactly as it was sent
    When the user PUTs kv/spec/<runId>/text with value "héllo \"wörld\" <script>alert(1)</script> ✓"
    Then reading it returns exactly that value
