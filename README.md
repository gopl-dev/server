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
https://github.com/swaggo/swag  
`swag fmt --dir server/handler`  
`swag init --parseDependency  --parseDepth 1 --dir server/handler -g handler.go -o server/docs`

# devtools
`go run ./cmd/cli/main.go rde`  
Reset dev environment (recreate DB, apply migrations & create default user). Useful during active development when you messed with the DB or need a clean state.

`go run ./cmd/cli/main.go sd`  
Will seed data to the database. By default, it seeds all available data. You can specify an entity and a count, for example:
`go run ./cmd/cli/main.go sd users 1000`.
Run `go run ./cmd/cli/main.go ? sd` to see available options and a detailed description.


