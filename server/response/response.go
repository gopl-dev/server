package response

import "time"

var Success = Status{Success: true}

type Status struct {
	Success bool `json:"success"`
}

type ServerStatus struct {
	Env     string    `json:"env"`
	Version string    `json:"version"`
	Time    time.Time `json:"time"`
}
