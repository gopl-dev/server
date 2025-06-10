package response

var Success = Status{Success: true}

type Status struct {
	Success bool `json:"success"`
}
