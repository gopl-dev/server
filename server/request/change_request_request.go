package request

import "github.com/gopl-dev/server/app/ds"

// FilterChangeRequests defines filtering options specific to change requests.
type FilterChangeRequests struct {
	Page    int                   `json:"page" url:"page,omitempty"`
	PerPage int                   `json:"per_page" url:"per_page,omitempty"`
	Status  ds.EntityChangeStatus `json:"status" url:"status,omitempty"`
}

// RejectChangeRequest represents a request payload for rejecting a change request.
type RejectChangeRequest struct {
	Note string `json:"note"`
}
