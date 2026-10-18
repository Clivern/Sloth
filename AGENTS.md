# Sloth Coding Rules

## General

- Write the happy path only: assume inputs are valid, data exists and calls succeed.
- Do not add defensive null checks, guards, type validations or try/catch wrappers unless explicitly requested.
- No nested if-statements for validation, and no redundant types or type assertions.
- When an issue is pointed out, fix the code directly without defending the previous version.

## Go File Layout

Keep declarations in every Go file in this order:

1. `package` and imports
2. `const` / `var`
3. All types (`type`, interfaces, structs)
4. Constructors (`New...`) and functions/methods

Do not place request/response structs or other types after constructors or methods.

```go
// good
type KV struct { ... }
type PutKVRequest struct { ... }
type KVResponse struct { ... }

func NewKV(...) *KV { ... }
func (k *KV) ListKV(...) { ... }
```

## Go Spacing

Separate logical steps with a blank line so code does not stick together. Match `api/knowledge.go`.

- Handlers: blank line between URL/param extraction and validation, request parsing, dependency wiring (`module.New…`, storage clients), the business call, and the success response.
- Module and service methods: blank line between setup/lookup and the mutating call, and before the final `return`.
- Blank line between distinct error-handling branches when a branch does more than a single `return`.
- One blank line between a struct definition and its constructor.
- Do not add blank lines inside short early-return blocks or between every statement.
- Do not break `gofmt` layout or split import groups beyond std / internal / third-party.

## UI Forms

- Never disable a submit button because required fields are empty. Disable it only while the request is in flight (e.g. `:disabled="saving"`).
- Mark required inputs with the `required` attribute (plus `minlength`, `type="email"`, etc.) inside a `<form @submit.prevent>` so the browser shows the validation error on submit.
- Rules the browser cannot check (e.g. "select at least one permission") are validated in the submit handler and surfaced through the form's error message, never by silently returning.
