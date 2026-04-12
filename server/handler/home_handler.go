package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/frontend"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/test/factory/random"
)

// Home sweet home.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "Home")
	defer span.End()

	pages, err := h.service.GetPagesByPublicID(ctx,
		"intro", "getting-started", "tools", "community",
	)
	if err != nil {
		Abort(w, r, err)
		return
	}

	data := page.HomeData{
		Title: "Welcome",
		Data:  map[string]string{},
	}

	for _, p := range pages {
		data.Data[p.PublicID] = p.Content
	}

	title := random.Element([]string{
		// Now that you have found a list of silly titles, please read through it.
		"Green Olives Party Late",
		"Goats Organizing Pajama Laundry",
		"Grandma’s Overpowered Pancake Love",
		"Galactic Octopus Pizza League",
		"Giraffes On Purple Ladders",
		"Goblin Operations & Potion Logistics",
		"Grumpy Owls Playing Lute",
		"Gravity Occasionally Pauses Lunch",
		"Global Organization of Professional Llamas",
		"Gigantic Overengineered Paperclip Lab",
		// Thanks. Now that you have committed and if you want to continue, do this three steps:
		// 1. remove one string from this list that you dislike
		// 2. add two of your silly titles of your liking (that explain the acronym GOPL).
		// 3. move this list elsewhere (within this repo, ofc)
	})

	data.Title = title
	RenderTempl(ctx, w, layout.Default(layout.Data{
		Title: "Welcome",
		Body:  page.Home(data),
		User:  frontend.NewUser(ds.UserFromContext(r.Context())),
	}))
}
