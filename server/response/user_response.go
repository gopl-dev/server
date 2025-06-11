package response

type LoginUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}
