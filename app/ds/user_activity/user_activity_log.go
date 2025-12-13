// Package useractivity ...
package useractivity

// Type defines the type of user activity.
type Type int

const (
	// None is the zero value for an activity type.
	None Type = iota
	// UserRegistered represents the event of a new user completing the sign-up process.
	UserRegistered
)
