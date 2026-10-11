Feature: Agent identities and client credentials

  An identity is a credential an agent uses to prove who it is: an api_key, a
  basic_auth username and secret, or a client_credentials pair that trades for
  short-lived tokens. The secret is shown once. A turned-off or deleted
  identity must stop getting tokens at once.

  Routes (shapes in openapi.yaml):
    GET    /api/v1/workspaces/{workspaceId}/agents/{agentName}/identities
    POST   /api/v1/workspaces/{workspaceId}/agents/{agentName}/identities
    PUT    /api/v1/workspaces/{workspaceId}/agents/{agentName}/identities/{identityId}
    DELETE /api/v1/workspaces/{workspaceId}/agents/{agentName}/identities/{identityId}
    POST   /api/v1/public/oauth/token   (form encoded, no API key)

  Note: the token route is rate limited to 10 requests a minute per client IP.
  Space the token scenarios out, or run fewer than 10 of them per minute.

  Background:
    Given the run's workspace exists
    And the agent "spec-<runId>-echo" is registered

  # ---------------------------------------------------------------------------
  # Issuing
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: An api_key identity is issued with its secret once
    When the user POSTs an identity named "spec-<runId>-key" with type "api_key"
    Then the response is 201
    And the body has id, name, type "api_key", isActive true and a non-empty secret

  @mutates
  Scenario: A basic_auth identity needs a username
    When the user POSTs an identity named "spec-<runId>-basic" with type "basic_auth" and no username
    Then the response is 400
    When the user POSTs it with username "svc"
    Then the response is 201
    And username is "svc"

  @readonly
  Scenario: Only basic_auth identities take a username
    When the user POSTs an identity named "spec-<runId>-key" with type "api_key" and username "svc"
    Then the response is 400

  @mutates
  Scenario: A client_credentials identity defaults to a 300 second token lifetime
    When the user POSTs an identity named "spec-<runId>-cc" with type "client_credentials"
    Then the response is 201
    And tokenTtl is 300

  @readonly
  Scenario: A token lifetime must be 60 to 86400 seconds
    When the user POSTs a client_credentials identity with tokenTtl 59
    Then the response is 400

  @mutates
  Scenario: Identity names are unique per agent
    Given an identity named "spec-<runId>-key" exists
    When the user POSTs another identity named "spec-<runId>-key"
    Then the response is 409

  @mutates
  Scenario: Listing identities never shows a secret
    Given an identity named "spec-<runId>-key" exists
    When the user lists the agent's identities
    Then it contains "spec-<runId>-key"
    And no item has a secret field

  # ---------------------------------------------------------------------------
  # Client credentials tokens
  # ---------------------------------------------------------------------------

  @mutates @smoke
  Scenario: A client_credentials identity trades its id and secret for a token
    Given a client_credentials identity "spec-<runId>-cc" with its id and secret
    When a request with no API key POSTs the token route with form
      grant_type=client_credentials, client_id=<id>, client_secret=<secret>
    Then the response is 200
    And token_type is "Bearer", expires_in is 300 and access_token is not empty
    And Cache-Control is "no-store"

  @mutates
  Scenario: The id and secret can also be sent as basic auth
    Given a client_credentials identity "spec-<runId>-cc" with its id and secret
    When a request POSTs the token route with grant_type=client_credentials
      and basic auth <id>:<secret>
    Then the response is 200

  @readonly
  Scenario: Only the client_credentials grant is supported
    When a request POSTs the token route with grant_type=password
    Then the response is 400
    And error is "unsupported_grant_type"

  @mutates
  Scenario: A wrong secret is an invalid client
    Given a client_credentials identity "spec-<runId>-cc"
    When a request POSTs the token route with its id and secret "wrong"
    Then the response is 401
    And error is "invalid_client"

  @mutates
  Scenario: An identity of another type cannot get a token
    Given an api_key identity "spec-<runId>-key" with its id and secret
    When a request POSTs the token route with that id and secret
    Then the response is 401
    And error is "invalid_client"

  @mutates
  Scenario: A turned-off identity stops getting tokens
    Given a client_credentials identity "spec-<runId>-cc" with its id and secret
    When the user PUTs it with name "spec-<runId>-cc" and isActive false
    Then the response is 200 and isActive is false
    And a token request with its id and secret returns 401

  @mutates
  Scenario: A deleted identity stops getting tokens
    Given a client_credentials identity "spec-<runId>-cc" with its id and secret
    When the user DELETEs it
    Then the response is 204
    And a token request with its id and secret returns 401

  @mutates @slow
  Scenario: The token route is rate limited
    When a request POSTs the token route 11 times in a row with a wrong grant_type
    Then the first 10 responses are 400
    And the 11th response is 429 with an errorMessage

  # ---------------------------------------------------------------------------
  # Editing and access
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: Renaming onto another identity's name is a conflict
    Given identities "spec-<runId>-a" and "spec-<runId>-b" exist
    When the user PUTs "spec-<runId>-b" with name "spec-<runId>-a"
    Then the response is 409

  @readonly
  Scenario: Updating an unknown identity returns not found
    When the user PUTs identities/<a random UUID> with name "x" and isActive true
    Then the response is 404

  @mutates
  Scenario: An identity cannot be reached through another agent
    Given an identity on "spec-<runId>-echo"
    And a second agent "spec-<runId>-other" is registered
    When the user DELETEs that identity's id under agents/spec-<runId>-other
    Then the response is 404
    And the identity still exists on "spec-<runId>-echo"

  @mutates
  Scenario: Reading identities needs the agent read grant; issuing needs upsert
    Given an access key with only CAN_GET_AGENT
    When that key lists the agent's identities
    Then the response is 200
    When that key POSTs an identity named "spec-<runId>-sneaky" with type "api_key"
    Then the response is 403

  # ---------------------------------------------------------------------------
  # More on tokens
  # ---------------------------------------------------------------------------

  @mutates
  Scenario: A custom token lifetime is used for the tokens
    Given a client_credentials identity "spec-<runId>-cc120" with tokenTtl 120
    When a request trades its id and secret for a token
    Then expires_in is 120

  @mutates
  Scenario: Each token request gets a new token
    Given a client_credentials identity "spec-<runId>-cc"
    When a request trades its id and secret for a token twice
    Then the two access_tokens differ

  @mutates
  Scenario: A turned-off identity works again once turned back on
    Given a client_credentials identity "spec-<runId>-cc" that has been turned off
    When the user PUTs it with isActive true
    Then a token request with its id and secret returns 200

  @readonly
  Scenario: A client id that is not a UUID is an invalid client, not a server error
    When a request POSTs the token route with grant_type=client_credentials,
      client_id=not-a-uuid and client_secret=x
    Then the response is 401
    And error is "invalid_client"

  @readonly
  Scenario: A token request without a grant type is rejected
    When a request POSTs the token route with client_id and client_secret but no grant_type
    Then the response is 400
    And error is "unsupported_grant_type"

  @mutates
  Scenario: Only client_credentials identities carry a token lifetime
    When the user POSTs an identity named "spec-<runId>-key-ttl" with type "api_key" and tokenTtl 120
    Then the response is 201
    And the body has no tokenTtl field
