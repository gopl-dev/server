# 🚀 Serving gopl.dev

# templ live watch & reload
`go tool templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/server/main.go"`

# tailwind watch
`tailwindcss -i ./frontend/assets/input.css -o ./frontend/assets/output.css --watch`

# linting
https://golangci-lint.run/docs/welcome/install/local/

`golangci-lint run`
`go run ./cmd/service_call_guard ./app/service`


# openapi & swagger
`swag fmt --dir server/handler`
`swag init --parseDependency  --parseDepth 1 --dir server/handler -g handler.go -o server/docs`

# devtools
Reset dev environment (recreate DB, apply migrations & create default user). Useful during active development when you messed with the DB or need a clean state.

`go run ./cmd/cli/main.go rde`  
(run `go run ./cmd/cli/main.go ? rde` for details)

