# Development Setup

Guide on how to set up your local instance of [gopl-dev/server](https://github.com/gopl-dev/server).

## Prerequisites

- [Git](https://git-scm.com/)
- [Go 1.26+](https://golang.org/dl/)
- [PostgreSQL 18](https://www.postgresql.org/download/)

If you are working on the frontend, you will also need:
- [templ](https://templ.guide/)
- [TailwindCSS](https://tailwindcss.com/)

## Setup

1. **Clone the repo:**
   ```bash
   git clone https://github.com/gopl-dev/server.git
   cd server
   ```

2. **Run the setup wizard:**
   ```bash
   go run ./cmd/setup_wizard/main.go
   ```
   This tool checks your DB connection and creates the necessary configuration files.

   <details>
   <summary>Manual setup (alternative):</summary>

    1. Copy `config.sample.yaml` to `.config.yaml` and edit the values. At least a DB connection is required for startup.
       > **Tip:** It's recommended to use a `_local_dev` suffix for the DB name (e.g., `myapp_local_dev`). The reset tool uses this convention to prevent accidental data loss.
    2. Create test configurations in:
        - `test/api_test/.config.yaml`
        - `test/service_test/.config.yaml`
        - `test/worker_test/.config.yaml`
    3. For each test config, update the following:
        - Set a test DB connection (e.g., `myapp_local_dev_test`).
        - `email.driver: "test"`
        - `tracing.enabled: false`
        - `files.storage_driver: "in-memory-fs"`

   </details>

3. **Run tests:**
   ```bash
   go test ./...
   ```

4. **Start the server:**
   ```bash
   go run ./cmd/server/main.go
   ```

## Seeding
To populate the database with test data, use the CLI tool:
```bash
go run ./cmd/cli/main.go sd
```
By default, it seeds all available entities. You can specify a specific entity and count:
* **Example:** `go run ./cmd/cli/main.go sd users 1000` (creates 1000 users).
* **Help:** `go run ./cmd/cli/main.go ? sd` for detailed options.

## Environment Reset
If you need a clean state, run:
```bash
go run ./cmd/cli/main.go rde
```
This command recreates the database, applies migrations, and creates a default user. 

---

## Toolchain

### templ — Live Watch & Reload
```bash
go tool templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/server/main.go"
```

### TailwindCSS — Watch
```bash
tailwindcss -i ./frontend/assets/input.css -o ./frontend/assets/output.css --watch
```

### Linting
Requires [golangci-lint](https://golangci-lint.run/welcome/install/#local-installation)
```bash
golangci-lint run
```

### OpenAPI & Swagger
Requires [swag](https://github.com/swaggo/swag)
```bash
# Format Swagger directives
swag fmt --dir server/handler

# Generate specifications
swag init --parseDependency --parseDepth 1 --dir server/handler -g handler.go -o server/docs
```