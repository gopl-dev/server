package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app/api/request"
	"github.com/gopl-dev/server/app/service"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req request.RegisterUser
	h := handleJSON(w, r, &req)
	if h.Aborted() {
		return
	}

	user, err := service.RegisterUser(r.Context(), req.ToParams())
	if err != nil {
		h.Abort(err)
		return
	}

	h.jsonOK(user)
}
