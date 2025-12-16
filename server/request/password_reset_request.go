package request

// PasswordRequestReset represents the request body for initiating a password reset.
type PasswordRequestReset struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordReset represents the request body for resetting a password with a token.
type PasswordReset struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}
