Feature: Workspace billing

  In the SaaS edition each workspace has a balance of AI tokens it buys through
  Stripe. Members can see the balance; only owners and admins can spend money.
  In the open-source edition these routes do not exist at all.

  Routes (shapes in openapi.yaml), SaaS edition only:
    GET  /api/v1/workspaces/{workspaceId}/billing
    GET  /api/v1/workspaces/{workspaceId}/billing/usage
    POST /api/v1/workspaces/{workspaceId}/billing/checkout
    POST /api/v1/workspaces/{workspaceId}/billing/portal
    POST /api/v1/public/action/stripe/webhook   (Stripe calls this, no API key)

  Note: read the edition from /api/v1/public/action/setup/status first. Run the
  @saas scenarios only when it is "saas" and the @oss ones only when it is "oss".
  Never complete a checkout: only check that a session URL comes back.

  Background:
    Given the run's workspace exists

  @readonly @saas @smoke
  Scenario: A member can read the workspace's billing status
    When the user GETs billing
    Then the response is 200 or 404
    And if 200, the body has workspaceId, aiTokensBalance and tokensPerUsd

  @readonly @saas
  Scenario: A member can read the workspace's usage
    When the user GETs billing/usage
    Then the response is 200 or 404
    And if 200, the body has periodReset, used and limits

  @mutates @saas
  Scenario: A checkout needs at least 100 cents
    When the user POSTs billing/checkout with amountCents 99
    Then the response is 400

  @mutates @saas
  Scenario: A valid checkout returns a Stripe session URL
    When the user POSTs billing/checkout with amountCents 500
    Then the response is 200
    And url is an https URL

  @mutates @saas
  Scenario: An access key without the billing write grant cannot start a checkout
    Given an access key with only CAN_GET_WORKSPACE_BILLING
    When that key POSTs billing/checkout with amountCents 500
    Then the response is 403
    When that key GETs billing
    Then the response is not 403

  @readonly @oss
  Scenario: The open-source edition has no billing API
    When the user GETs billing
    Then the response is 404

  @readonly @saas
  Scenario: The Stripe webhook rejects an unsigned event
    When a request with no credentials POSTs the webhook with body {"type": "checkout.session.completed"}
      and no Stripe-Signature header
    Then the response is 400 or 503
    And no workspace balance changes

  @readonly @saas
  Scenario: The Stripe webhook rejects a forged signature
    When a request POSTs the webhook with Stripe-Signature "t=1,v1=spec<runId>"
    Then the response is 400 or 503

  @readonly @oss
  Scenario: The open-source edition has no Stripe webhook
    When a request with no credentials POSTs the webhook
    Then the response is 404
