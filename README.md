# server
A humble server that is doing its best to handle API calls to gopl.dev

# templ live watch & reload
`templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/server/main.go"`

# tailwind watch
`npx @tailwindcss/cli -i ./web/assets/input.css -o ./web/assets/output.css --watch`