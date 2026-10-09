# Code Complexity Report

Generated on 2026-10-08 with `gocyclo` v0.6.0 (cyclomatic) and `gocognit` v1.2.0 (cognitive). Test files are excluded.

Re-run manually with:

```sh
go run github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0 -ignore '_test.go' db/ module/
go run github.com/uudashr/gocognit/cmd/gocognit@v1.2.0 -ignore '_test.go' db/ module/
```

Guideline: a function above ~10 cyclomatic or ~15 cognitive is worth a look.

## `db/` package

### Files (high to low)

| # | File | Cyclomatic (total) | Worst func (cyclo) | Funcs | Cognitive (total) | Worst func (cognit) | Lines |
|---|---|---|---|---|---|---|---|
| 1 | `db/health_check.go` | 48 | 8 | 21 | 31 | 9 | 566 |
| 2 | `db/user.go` | 36 | 4 | 19 | 20 | 4 | 510 |
| 3 | `db/agent_instance.go` | 32 | 4 | 15 | 20 | 4 | 384 |
| 4 | `db/agent_protection.go` | 31 | 4 | 15 | 19 | 4 | 384 |
| 5 | `db/workspace.go` | 31 | 4 | 17 | 16 | 4 | 371 |
| 6 | `db/invite.go` | 29 | 5 | 14 | 18 | 5 | 354 |
| 7 | `db/async_task.go` | 28 | 4 | 14 | 17 | 4 | 314 |
| 8 | `db/agent.go` | 28 | 4 | 15 | 15 | 4 | 319 |
| 9 | `db/intention.go` | 28 | 4 | 15 | 15 | 4 | 334 |
| 10 | `db/integration.go` | 26 | 4 | 14 | 14 | 4 | 282 |
| 11 | `db/subscription.go` | 24 | 4 | 14 | 11 | 4 | 287 |
| 12 | `db/session.go` | 22 | 4 | 13 | 10 | 4 | 247 |
| 13 | `db/helpers.go` | 21 | 5 | 7 | 16 | 5 | 138 |
| 14 | `db/workspace_user.go` | 19 | 4 | 10 | 11 | 4 | 220 |
| 15 | `db/agent_identity.go` | 17 | 4 | 8 | 11 | 4 | 208 |
| 16 | `db/access.go` | 16 | 4 | 9 | 8 | 4 | 194 |
| 17 | `db/api_key.go` | 16 | 4 | 9 | 8 | 4 | 197 |
| 18 | `db/usage.go` | 15 | 4 | 7 | 9 | 4 | 199 |
| 19 | `db/workspace_kv.go` | 12 | 4 | 6 | 7 | 4 | 151 |
| 20 | `db/agent_check.go` | 11 | 4 | 6 | 6 | 4 | 144 |
| 21 | `db/config.go` | 11 | 4 | 6 | 6 | 4 | 115 |
| 22 | `db/audit.go` | 10 | 4 | 5 | 6 | 4 | 153 |
| 23 | `db/traffic.go` | 10 | 4 | 5 | 6 | 4 | 171 |
| 24 | `db/types.go` | 10 | 4 | 5 | 4 | 2 | 61 |
| 25 | `db/password_reset_token.go` | 9 | 2 | 6 | 3 | 1 | 114 |
| 26 | `db/database.go` | 8 | 6 | 3 | 5 | 5 | 87 |
| 27 | `db/kv.go` | 8 | 2 | 5 | 3 | 1 | 105 |
| 28 | `db/token_purchase.go` | 4 | 3 | 2 | 2 | 2 | 59 |
| 29 | `db/workspace_stats.go` | 2 | 1 | 2 | 0 | 0 | 35 |

### Top 20 functions (high to low)

| # | Function | Location | Cyclomatic | Cognitive |
|---|---|---|---|---|
| 1 | `(*HealthCheckRepositoryPostgres).ClaimDue` | `db/health_check.go:367:1` | 8 | 9 |
| 2 | `NewConnection` | `db/database.go:35:1` | 6 | 5 |
| 3 | `InitDB` | `db/helpers.go:32:1` | 5 | 5 |
| 4 | `(*UserInviteRepositoryPostgres).ListByWorkspaceId` | `db/invite.go:139:1` | 5 | 5 |
| 5 | `(*UserInviteRepositoryPostgres).ListByEmail` | `db/invite.go:183:1` | 5 | 5 |
| 6 | `(*AccessKeyRepositoryPostgres).ListByWorkspaceId` | `db/access.go:116:1` | 4 | 4 |
| 7 | `(*AgentRepositoryPostgres).ListByWorkspaceId` | `db/agent.go:160:1` | 4 | 4 |
| 8 | `(*AgentMetaRepositoryPostgres).ListByAgentId` | `db/agent.go:277:1` | 4 | 4 |
| 9 | `(*AgentCheckRepositoryPostgres).ListByAgentId` | `db/agent_check.go:112:1` | 4 | 4 |
| 10 | `(*AgentIdentityRepositoryPostgres).ListByAgentId` | `db/agent_identity.go:141:1` | 4 | 4 |
| 11 | `(*AgentIdentityRepositoryPostgres).ListActiveByAgentId` | `db/agent_identity.go:176:1` | 4 | 4 |
| 12 | `(*AgentInstanceRepositoryPostgres).ListByAgentId` | `db/agent_instance.go:181:1` | 4 | 4 |
| 13 | `(*AgentInstanceRepositoryPostgres).ListLiveByAgentId` | `db/agent_instance.go:218:1` | 4 | 4 |
| 14 | `(*AgentInstanceMetaRepositoryPostgres).ListByAgentInstanceId` | `db/agent_instance.go:342:1` | 4 | 4 |
| 15 | `(*AgentProtectionRepositoryPostgres).ListByAgentId` | `db/agent_protection.go:172:1` | 4 | 4 |
| 16 | `(*AgentProtectionRepositoryPostgres).ListActiveByAgentId` | `db/agent_protection.go:208:1` | 4 | 4 |
| 17 | `(*AgentProtectionTokenRepositoryPostgres).ListByProtectionId` | `db/agent_protection.go:324:1` | 4 | 4 |
| 18 | `(*APIKeyRepositoryPostgres).ListByUserId` | `db/api_key.go:118:1` | 4 | 4 |
| 19 | `(*AsyncTaskRepositoryPostgres).ListByStatus` | `db/async_task.go:155:1` | 4 | 4 |
| 20 | `(*AsyncTaskMetaRepositoryPostgres).ListByAsyncTaskId` | `db/async_task.go:271:1` | 4 | 4 |

## `module/` package

### Files (high to low)

| # | File | Cyclomatic (total) | Worst func (cyclo) | Funcs | Cognitive (total) | Worst func (cognit) | Lines |
|---|---|---|---|---|---|---|---|
| 1 | `module/instance.go` | 116 | 45 | 6 | 151 | 65 | 669 |
| 2 | `module/agent.go` | 87 | 35 | 5 | 130 | 47 | 463 |
| 3 | `module/agent_check.go` | 79 | 38 | 5 | 96 | 60 | 451 |
| 4 | `module/kv.go` | 52 | 14 | 5 | 57 | 16 | 258 |
| 5 | `module/intention.go` | 49 | 13 | 7 | 43 | 12 | 330 |
| 6 | `module/billing.go` | 48 | 11 | 8 | 44 | 14 | 323 |
| 7 | `module/invite.go` | 46 | 14 | 6 | 49 | 17 | 360 |
| 8 | `module/oauth.go` | 34 | 34 | 1 | 41 | 41 | 178 |
| 9 | `module/workspace.go` | 33 | 9 | 6 | 33 | 11 | 255 |
| 10 | `module/access.go` | 33 | 9 | 5 | 30 | 9 | 258 |
| 11 | `module/member.go` | 22 | 8 | 4 | 20 | 9 | 152 |
| 12 | `module/api_key.go` | 22 | 6 | 5 | 18 | 5 | 173 |
| 13 | `module/session.go` | 21 | 9 | 6 | 17 | 10 | 115 |
| 14 | `module/permission.go` | 21 | 6 | 9 | 10 | 4 | 215 |
| 15 | `module/context.go` | 20 | 4 | 10 | 8 | 2 | 100 |
| 16 | `module/usage.go` | 17 | 8 | 6 | 14 | 10 | 118 |
| 17 | `module/traffic.go` | 15 | 7 | 3 | 13 | 7 | 135 |
| 18 | `module/bus.go` | 15 | 5 | 5 | 11 | 5 | 112 |
| 19 | `module/audit.go` | 13 | 6 | 3 | 10 | 5 | 125 |
| 20 | `module/setup.go` | 11 | 8 | 3 | 8 | 7 | 100 |
| 21 | `module/me.go` | 11 | 5 | 4 | 7 | 4 | 152 |
| 22 | `module/settings.go` | 8 | 4 | 3 | 9 | 6 | 64 |
| 23 | `module/profile.go` | 8 | 4 | 3 | 5 | 3 | 72 |
| 24 | `module/cache.go` | 6 | 3 | 4 | 2 | 2 | 47 |
| 25 | `module/auth.go` | 5 | 4 | 2 | 3 | 3 | 75 |
| 26 | `module/stats.go` | 5 | 4 | 2 | 3 | 3 | 43 |

### Top 20 functions (high to low)

| # | Function | Location | Cyclomatic | Cognitive |
|---|---|---|---|---|
| 1 | `(*Instance).RegisterInstance` | `module/instance.go:286:1` | 45 | 65 |
| 2 | `(*Check).UpsertAgentCheck` | `module/agent_check.go:153:1` | 38 | 60 |
| 3 | `(*Agent).UpsertAgent` | `module/agent.go:292:1` | 35 | 47 |
| 4 | `(*Auth).LoginWithOAuth` | `module/oauth.go:30:1` | 34 | 41 |
| 5 | `(*Agent).ListAgents` | `module/agent.go:91:1` | 22 | 44 |
| 6 | `(*Agent).GetAgent` | `module/agent.go:193:1` | 22 | 33 |
| 7 | `(*Instance).RenewInstance` | `module/instance.go:519:1` | 22 | 25 |
| 8 | `(*Check).ReportCheck` | `module/agent_check.go:372:1` | 21 | 19 |
| 9 | `(*Instance).ListInstances` | `module/instance.go:95:1` | 19 | 30 |
| 10 | `(*Instance).GetInstance` | `module/instance.go:192:1` | 19 | 22 |
| 11 | `(*Invite).CreateInvite` | `module/invite.go:102:1` | 14 | 17 |
| 12 | `(*KeyValue).ListKeyValue` | `module/kv.go:66:1` | 14 | 16 |
| 13 | `(*KeyValue).PutKeyValue` | `module/kv.go:162:1` | 13 | 15 |
| 14 | `(*Intention).UpdateIntention` | `module/intention.go:209:1` | 13 | 12 |
| 15 | `(*KeyValue).GetKeyValue` | `module/kv.go:115:1` | 12 | 13 |
| 16 | `(*KeyValue).DeleteKeyValue` | `module/kv.go:222:1` | 12 | 13 |
| 17 | `(*Billing).CreditTokenPurchase` | `module/billing.go:259:1` | 11 | 14 |
| 18 | `(*Check).DeleteAgentCheck` | `module/agent_check.go:327:1` | 11 | 10 |
| 19 | `(*Instance).DeregisterInstance` | `module/instance.go:631:1` | 10 | 9 |
| 20 | `(*Workspace).CreateWorkspace` | `module/workspace.go:67:1` | 9 | 11 |
