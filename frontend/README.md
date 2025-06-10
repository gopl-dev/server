# Frontend

To compose our frontend, we use [templ](https://templ.guide/).\
For styling, we use [Tailwind CSS](https://tailwindcss.com/) and [daisyUI](https://daisyui.com/docs/install/) \
For interactivity, we use [Alpine.js](https://alpine.dev/).

To install and run these tools, please refer to their corresponding docs, as this is subject to change.\
As of now, it's as simple as that:

**templ**:\
`templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/server/main.go"`

**Tailwind CSS**:\
`tailwindcss -i ./frontend/assets/input.css -o ./frontend/assets/output.css --watch`

**Alpine.js**:\
It is just included in the head tag using `<script src="//unpkg.com/alpinejs" defer></script>` (see `head.templ` here).