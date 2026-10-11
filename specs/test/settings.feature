Feature: Platform settings

  The platform email and maintenance mode apply to everyone on the server, so
  only a platform admin may read or change them. Changing them on a shared
  server affects every user, which is why the write scenario is destructive.

  Routes (shapes in openapi.yaml):
    GET /api/v1/action/settings
    PUT /api/v1/action/settings

  @readonly @admin
  Scenario: A platform admin can read the settings
    Given the user's role from /api/v1/me is "admin"
    When the user GETs the settings
    Then the response is 200
    And the body has a settings object

  @readonly
  Scenario: A user who is not a platform admin cannot read the settings
    Given the user's role from /api/v1/me is not "admin"
    When the user GETs the settings
    Then the response is 403

  @readonly @admin
  Scenario: The platform email must be an email
    Given the user's role from /api/v1/me is "admin"
    When the user PUTs the settings with platformEmail "not-an-email"
    Then the response is 400
    And the settings are unchanged

  @destructive @admin
  Scenario: Saving the settings stores them
    Given the user's role from /api/v1/me is "admin"
    And the current settings are remembered
    When the user PUTs the same platformEmail with maintenanceMode unchanged
    Then the response is 200
    And reading the settings returns the same values

  @mutates
  Scenario: A workspace access key cannot reach the platform settings
    Given the run's workspace exists
    And an access key with every permission an access key may hold
    When that key GETs the settings
    Then the response is 401 or 403
