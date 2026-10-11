# Cognit Coding Rules

## UI Forms

- Never disable a submit button because required fields are empty. Disable it only while the request is in flight (e.g. `:disabled="saving"`).
- Mark required inputs with the `required` attribute (plus `minlength`, `type="email"`, etc.) inside a `<form @submit.prevent>` so the browser shows the validation error on submit.
- Rules the browser cannot check (e.g. "select at least one permission") are validated in the submit handler and surfaced through the form's error message, never by silently returning.
