Feature: Public routes

  The routes a load balancer, an installer or a logged-out browser can call
  without credentials. They must answer without a key, and they must never
  hand out more than the caller is owed: a probe says up or down, setup runs
  once, and the metrics page stays behind its own password.

  Routes (shapes in openapi.yaml):
    GET  /api/v1/public/_health
    GET  /api/v1/public/_ready
    GET  /api/v1/public/action/setup/status
    POST /api/v1/public/action/setup
    POST /api/v1/public/action/logout
    GET  /api/v1/public/action/oauth/{github|google}
    GET  /api/v1/public/action/oauth/{github|google}/callback
    GET  /api/v1/public/_metrics

  Note: do not follow redirects in this file; check the 302 and its Location.
  Setup is rate limited to 5 requests a minute per client IP.

  # ---------------------------------------------------------------------------
  # Probes
  # ---------------------------------------------------------------------------

  @readonly @smoke
  Scenario: The liveness probe answers without credentials
    When a request with no credentials GETs /api/v1/public/_health
    Then the response is 200
    And the body is {"status":"ok"}

  @readonly @smoke
  Scenario: The readiness probe reports the database is reachable
    When a request with no credentials GETs /api/v1/public/_ready
    Then the response is 200
    And the body is {"status":"ok"}

  # ---------------------------------------------------------------------------
  # Setup
  # ---------------------------------------------------------------------------

  @readonly
  Scenario: Setup status says whether the platform is installed and which edition runs
    When a request with no credentials GETs /api/v1/public/action/setup/status
    Then the response is 200
    And installed is true
    And edition is "oss" or "saas"

  @readonly
  Scenario: Setup cannot run a second time
    When a request with no credentials POSTs setup with platformEmail "spec-<runId>@example.com"
    Then the response is 400
    And the platform's settings are unchanged

  @readonly
  Scenario: Setup rejects an invalid email before checking anything else
    When a request with no credentials POSTs setup with platformEmail "not-an-email"
    Then the response is 400

  @readonly @slow
  Scenario: Setup is rate limited
    Given a minute has passed since the last setup request from this IP
    When a request with no credentials POSTs setup 6 times in a row with platformEmail "not-an-email"
    Then the first 5 responses are 400
    And the 6th response is 429 with an errorMessage and a Retry-After header

  # ---------------------------------------------------------------------------
  # Logout
  # ---------------------------------------------------------------------------

  @readonly
  Scenario: Logout succeeds even with no session
    When a request with no credentials POSTs /api/v1/public/action/logout
    Then the response is 200
    And the body has a successMessage
    And the response clears the _cognit_session cookie

  # ---------------------------------------------------------------------------
  # OAuth sign-in
  # ---------------------------------------------------------------------------

  @readonly
  Scenario Outline: Starting sign-in redirects to the provider with a state cookie
    When a request with no credentials GETs /api/v1/public/action/oauth/<provider>
    Then the response is 302
    And Location points at <host>
    And the response sets an HttpOnly state cookie that lives 10 minutes
    And the state in Location equals the cookie's value

    Examples:
      | provider | host                      |
      | github   | github.com                |
      | google   | accounts.google.com       |

  @readonly
  Scenario Outline: A callback with a forged state never signs anyone in
    When a request GETs /api/v1/public/action/oauth/<provider>/callback?code=x&state=spec-<runId>
      with no state cookie
    Then the response is 302
    And Location ends with "/login?oauth_error=<provider>"
    And the response sets no _cognit_session cookie

    Examples:
      | provider |
      | github   |
      | google   |

  # ---------------------------------------------------------------------------
  # Metrics
  # ---------------------------------------------------------------------------

  @readonly
  Scenario: Metrics need their own basic auth, not an API key
    When the user GETs /api/v1/public/_metrics with only X-API-Key
    Then the response is 401

  @readonly
  Scenario: Metrics reject wrong basic auth credentials
    When a request GETs /api/v1/public/_metrics with basic auth "spec":"wrong-<runId>"
    Then the response is 401
