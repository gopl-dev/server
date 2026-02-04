package response

import (
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

// FilterEventLogs represents a paginated collection of event logs returned by a filter operation.
type FilterEventLogs struct {
	Data  []EventLog `json:"data"`
	Count int        `json:"count"`
}

// EventLog represents a serialized event log entry for API responses.
type EventLog struct {
	ID      ds.ID  `json:"id"`
	Message string `json:"message"`
	Date    string `json:"date"`
}

// NewFilterEventLog converts domain event logs into a response model.
func NewFilterEventLog(data []ds.EventLog, count int) FilterEventLogs {
	r := FilterEventLogs{
		Count: count,
		Data:  make([]EventLog, len(data)),
	}

	for i, d := range data {
		r.Data[i] = EventLog{
			ID:      d.ID,
			Message: d.RenderMessage(),
			Date:    app.HumanTime(d.CreatedAt),
		}
	}

	return r
}
