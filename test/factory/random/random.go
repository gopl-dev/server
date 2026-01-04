// Package random provides utility functions for generating various types of random data.
package random

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
		s[i] = alphanum[rand.IntN(len(alphanum))] //nolint:gosec
	}

	return string(s)
}

// Title generates a random string formatted in title case.
// It accepts an optional numWordsOpt to specify the number of words;
// otherwise, it defaults to a random count between 3 and 5.
func Title(numWordsOpt ...int) string {
	var numWords int
	if len(numWordsOpt) == 1 {
		numWords = numWordsOpt[0]
	} else {
		numWords = rand.IntN(3) + 3 //nolint:gosec,mnd
	}

	words := make([]string, numWords)
	for i := range words {
		wordLen := rand.IntN(8) + 3 //nolint:gosec,mnd
		words[i] = strings.ToLower(String(wordLen))
	}

	title := strings.Join(words, " ")
	caser := cases.Title(language.English)

	return caser.String(title)
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
	d := digits[rand.IntN(len(digits))] //nolint:gosec

	return string(d)
}

// Letter returns a single random alphabetic character (a-z, A-Z) as a string.
func Letter() string {
	l := letters[rand.IntN(len(letters))] //nolint:gosec

	return string(l)
}

// Int generates a random integer within the specified inclusive range [min, max].
// If max is less than min, the arguments are swapped to ensure the range is valid.
func Int(from, to int) int {
	if to < from {
		from, to = to, from
	}

	return rand.IntN(to-from+1) + from //nolint:gosec
}

// Element returns a random element from the provided slice.
// If the slice is empty, it returns the zero value for type T.
func Element[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}

	return slice[rand.IntN(len(slice))] //nolint:gosec
}

// URL generates a random mocked HTTPS URL string.
// It constructs a URL in the format: https://{10 chars}.{2 chars}/{10 chars}/.
func URL() string {
	return "https://" + String(10) + "." + String(2) + "/" + String(10) + "/" //nolint:mnd
}
