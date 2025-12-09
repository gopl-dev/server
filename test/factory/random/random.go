// Package random provides utility functions for generating various types of random data.
package random

import (
	"fmt"
	"math/rand"
	"strings"
)

var (
	letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	digits  = []byte("0123456789")
)

// String generates a random alphanumeric string.
// The length of the string defaults to 10 if no argument is provided.
func String(lengthOpt ...int) string {
	length := 10
	if len(lengthOpt) == 1 {
		length = lengthOpt[0]
	}

	alphanum := letters
	alphanum = append(alphanum, digits...)

	s := make([]byte, length)
	for i := range length {
		s[i] = alphanum[rand.Intn(len(alphanum))] //nolint:gosec
	}

	return string(s)
}

// Email generates randomly constructed email address.
func Email() string {
	return strings.ToLower(fmt.Sprintf("%s@%s.%s", Letters(6), Letters(6), Letters(3))) //nolint:mnd
}

// Letters generates a random string consisting only of alphabetic characters (a-z, A-Z).
func Letters(length int) string {
	var sb strings.Builder
	sb.Grow(length)

	for range length {
		sb.WriteString(Letter())
	}

	return sb.String()
}

// Digits generates a random string consisting only of numeric characters (0-9).
func Digits(length int) string {
	var sb strings.Builder
	sb.Grow(length)

	for range length {
		sb.WriteString(Digit())
	}

	return sb.String()
}

// Digit returns a single random numeric character (0-9) as a string.
func Digit() string {
	d := digits[rand.Intn(len(digits))] //nolint:gosec

	return string(d)
}

// Letter returns a single random alphabetic character (a-z, A-Z) as a string.
func Letter() string {
	l := letters[rand.Intn(len(letters))] //nolint:gosec

	return string(l)
}

// Int generates a random integer within the specified inclusive range [min, max].
// If max is less than min, the arguments are swapped to ensure the range is valid.
func Int(from, to int) int {
	if to < from {
		from, to = to, from
	}

	return rand.Intn(to-from+1) + from //nolint:gosec
}
