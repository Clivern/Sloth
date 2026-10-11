Feature: Workspace isolation

  Everything in Cognit lives in a workspace, and nothing may cross from one
  workspace to another. A key minted for one team must not read, change or
  even confirm the existence of another team's agents, keys, rules or data.
  Each other file checks its own resource; this one sweeps all of them with
  the same two keys, so a route that forgets its workspace check shows up here.

  Routes: every workspace route; the scenarios name the ones they use.

  Background:
    Given the run's workspace "A" exists
    And a second workspace "spec-<runId>-b", called "B", exists
    And B has an agent "spec-<runId>-echo" with instance "i-1", a ttl check "heartbeat",
      an identity, a protection, a kv key "spec/<runId>/secret", a rule and an access key
    And an access key in A with every permission an access key may hold

  @mutates @smoke
  Scenario: A's key cannot read anything in B
    When A's key GETs, in B: the workspace, members, invites, keys, audits, traffic,
      agents, agents/spec-<runId>-echo, its instances, instance i-1, its checks,
      its identities, its protection, its well-known card, intentions,
      the rule by id, kv and kv/spec/<runId>/secret
    Then every response is 403

  @mutates
  Scenario: A's key cannot change anything in B
    When A's key, in B: PUTs the workspace name, PUTs agents/spec-<runId>-echo,
      PUTs instance i-2, POSTs heartbeat/pass for i-1, PUTs checks/heartbeat,
      POSTs an identity, PUTs protection, PUTs a rule, PUTs kv/spec/<runId>/secret,
      POSTs a key and DELETEs the agent
    Then every response is 403
    And everything in B is unchanged when the user reads it

  @mutates
  Scenario: A resource id from B is not found through A
    When the user reads B's rule id, access key id and invite id under A's routes
    Then every response is 404

  @readonly
  Scenario: A workspace that does not exist looks the same to every caller
    When the user and A's key each GET agents in a random workspace id
    Then the user gets 404
    And A's key gets 403 or 404
    And neither body names the workspace
