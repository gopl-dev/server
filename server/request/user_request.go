// Package request ...
package request

// UserSignUp ...
type UserSignUp struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ConfirmEmail ...
type ConfirmEmail struct {
	Code string `json:"code"`
}

// UserSignIn ...
type UserSignIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
