# Project Context: gopl-server

## Overview
This is a Go-based web server application using a layered architecture. It serves both a JSON API and server-side rendered HTML pages using [templ](https://templ.guide/).

## Interaction Rules
- Be extremely concise and professional.
- Avoid conversational filler, politeness, or introductory phrases (e.g., "Certainly!", "I can help with that", "Here is the code").
- Provide direct answers or code immediately without unnecessary explanations.
- If a question can be answered with "Yes" or "No", do so.
- Do not explain obvious logic unless explicitly asked for a breakdown.
- Focus on "Code first, text second."

## Tech Stack
- **Language:** Go (Golang) 1.23+
- **Database:** PostgreSQL 18+
- **DB Driver:** `pgx/v5` with `georgysavva/scany/v2` for scanning.
- **Routing:** Standard `net/http` `ServeMux` with a custom wrapper (`server/endpoint/router.go`).
- **Frontend:** `templ` (Go templating), `tailwindcss`, `daisyui`.
- **Validation:** `github.com/Oudwins/zog`.
- **Config:** YAML based.
- **Tooling:** `templ`, `tailwindcss`, `golangci-lint`

## Architecture Layers

### 1. Domain Structs (`app/ds`)
- Contains pure data structures (models) matching database tables.
- JSON tags should be `snake_case`.

### 2. Repository (`app/repo`)
- **Responsibility:** Direct database access.
- **Rules:**
  - Use **Raw SQL** queries. No ORM.
  - Use `pgxscan` to map rows to structs.
  - Always pass `ctx` and start a tracer span.
  - Return sentinel errors defined in `repo.go` or specific files (e.g., `ErrUserNotFound`, `ErrEntityNotFound`) instead of raw pgx errors.
  - Use the `noRows(err)` helper to check for `pgx.ErrNoRows`.

### 3. Service (`app/service`)
- **Responsibility:** Business logic.
- **Rules:**
  - Always pass `ctx` and start a tracer span.
  - Performs validation.

### 4. Handler (`server/handler`)
- **Responsibility:** HTTP layer (parsing request, formatting response).
- **Rules:**
  - **Request DTOs:** Define structs in `server/request/`.
  - **Validation:** Use `handleJSON` helper which automatically validates using `zog`.
  - Call Service methods.
  - Return responses using helpers (`res.jsonSuccess()`, `renderTempl()`).

### 5. Frontend (`frontend/`)
- **Pages:** Located in `frontend/page/`.
- **Components:** Reusable components in `frontend/component/`.
- **Assets:** Served from `frontend/assets/` (disk in dev, embed in prod).
- **Styling:** Use TailwindCSS classes.

## Development Workflow
- **Run Command:** `go tool templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/server/main.go"`.
- **Assets:** In `dev` environment, assets are served from disk. In `prod`, they are embedded.

## Testing (`test/`)
- **Type:** Integration/API tests.
- **Location:** `test/api_test/`.
- **Helpers:**
  - Use `tt.Factory` to create test data (Users, Logs, etc.).
  - Use `POST(t, Request{...})` helper to make requests.
  - Use `test.AssertInDB` / `test.AssertNotInDB` to verify side effects.
  - Use `test.LoadEmailVars` to extract tokens from "sent" emails.

## Common Patterns & Gotchas
- **Tracing:** Every method in Repo, Service, and Handler should start with `ctx, span := r.tracer.Start(ctx, "MethodName")` and `defer span.End()`.
- **Validation:** Each public service method must validate it's input with `err = Normalize(in)`
- **SQL:** Use `$1`, `$2` placeholders.
- **Error Handling:** Do not use inline error checks (e.g., `if err := foo(); err != nil`). Always assign the error to a variable on a separate line before checking it.
