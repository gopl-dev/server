# server
A humble server that serves gopl.dev

# templ live watch & reload
`go tool templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/server/main.go"`

# tailwind watch
`npx @tailwindcss/cli -i ./frontend/assets/input.css -o ./frontend/assets/output.css --watch`

# linting
`golangci-lint run`

