package request

// FilterEventLogs defines filtering options specific to event logs.
type FilterEventLogs struct {
	Page    int `json:"page" url:"page,omitempty"`
	PerPage int `json:"per_page" url:"per_page,omitempty"`
}
