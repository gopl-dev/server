package response

// UserSignIn ...
type UserSignIn struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}
