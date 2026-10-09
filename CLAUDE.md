# Cognit Coding Rules

## Go Inline Spacing

Separate logical steps with a blank line so code does not stick together. Match `api/knowledge.go`.

### Handlers and actions

Put a blank line between each phase:

1. URL/param extraction and early validation
2. Request parsing (`DecodeAndValidate`, form upload, etc.)
3. Dependency wiring (`module.New…`, storage clients)
4. Business call
5. Success response

```go
// ✅ GOOD
func UploadDocumentAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		util.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	form, err := util.ParseUploadForm(r)
	if err != nil {
		util.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": err.Error(),
		})
		return
	}

	store, err := storage.New()
	if err != nil {
		// ...
		return
	}

	dm := module.NewDocument(
		db.NewWorkspaceDocumentRepository(db.GetDB()),
		db.NewWorkspaceRepository(db.GetDB()),
		store,
	)

	doc, err := dm.UploadDocument(r.Context(), form, db.Id(wid))
	if err != nil {
		// ...
		return
	}

	util.WriteJSON(w, http.StatusCreated, doc)
}
```

### Module and service methods

- Blank line between setup/lookup and the mutating call
- Blank line before the final `return`
- Blank line between distinct error-handling branches when a branch does more than a single `return`

### Types and constructors

Keep one blank line between a struct definition and its constructor when both live in the same file.

### Do not

- Add blank lines inside short early-return blocks
- Add blank lines between every statement
- Break `gofmt` layout or split import groups beyond std / internal / third-party

## Go File Layout

In every Go file, keep declarations in this order:

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

```go
// bad
type KV struct { ... }
func NewKV(...) *KV { ... }
type PutKVRequest struct { ... }
```

## Happy Path Programming Only

You are an ultra-minimalist developer. Your primary goal is to write clean, concise, readable code that handles the intended use-case perfectly, assuming all inputs are valid.

### Core Directives

* **Write Happy Path Only**: Assume inputs are always valid, data exists, and APIs succeed.
* **No Defensive Safeguards**: Do not add null checks, undefined guards, or type validations unless explicitly requested.
* **No Exception Wrappers**: Avoid try-catch blocks completely. Let the runtime handle errors.
* **Trust the Environment**: Never write boilerplate code to protect against "what ifs."
* **When Challenged**: If the user points out an issue, accept it objectively. Change the code immediately without writing lengthy apologies or defending previous iterations.

### Code Style

* No nested if-statements for validation.
* No redundant types or type assertions.
