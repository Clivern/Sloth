Feature: The user's profile

  A user reads and edits their own name, language and theme. Only the person
  holding the key can see or change them, and an edit must come back exactly
  as it was saved — the dashboard renders straight from this response.

  Routes (shapes in openapi.yaml):
    GET /api/v1/action/profile
    PUT /api/v1/action/profile

  Note: scenarios that change the profile restore the original values at the
  end, so the run leaves the user as it found them.

  @readonly @smoke
  Scenario: Reading the profile returns the user's own record
    When the user GETs the profile
    Then the response is 200
    And user has id, email, role, isActive, name, avatar, language, theme,
      lastLoginAt, createdAt and updatedAt
    And user.id matches owner.id from /api/v1/me

  @mutates
  Scenario: Updating the profile saves and echoes the new values
    Given the user's current profile is remembered
    When the user PUTs the profile with name "Spec <runId>", language "fr" and theme "dark"
    Then the response is 200
    And user.name is "Spec <runId>", user.language is "fr" and user.theme is "dark"
    And the body has a successMessage
    And reading the profile again returns the same values
    And the remembered profile is restored

  @readonly
  Scenario: An unsupported language is rejected
    When the user PUTs the profile with language "de" and theme "default"
    Then the response is 400
    And the profile is unchanged

  @readonly
  Scenario: An unknown theme is rejected
    When the user PUTs the profile with language "en" and theme "neon"
    Then the response is 400

  @readonly
  Scenario: Language and theme are both required
    When the user PUTs the profile with only a name
    Then the response is 400

  @mutates
  Scenario: A workspace access key cannot reach a user's profile
    Given the run's workspace exists
    And an access key with every permission an access key may hold
    When that key GETs the profile
    Then the response is 401 or 403
