// Package cli provides a lightweight CLI framework with support for
// positional arguments, named parameters, and flags.
package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	aur "github.com/logrusorgru/aurora"
)

// Handler ...
type Handler interface {
	Handle(ctx context.Context) error
}

// Confirm asks for y/n confirmation.
func Confirm(questionOpt ...string) (ok bool) {
	question := "Confirm?"
	yes := "y"
	yesAlt := "yes"

	if len(questionOpt) > 0 {
		question = questionOpt[0]
	}
	question += " y/n..."

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\n> " + aur.Bold(aur.Green(question)).String() + "\n")
	scanner.Scan()
	input := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if input == yes || input == yesAlt {
		return true
	}

	return false
}

// splitArgs splits a command line into args, respecting single and double quotes.
func splitArgs(s string) []string {
	var out []string
	var b strings.Builder

	flush := func() {
		if b.Len() == 0 {
			return
		}
		out = append(out, b.String())
		b.Reset()
	}

	// findClosingQuote searches for an unescaped closing quote starting from `start`.
	// It treats backslash as an escape inside quotes.
	findClosingQuote := func(quote byte, start int) int {
		esc := false
		for i := start; i < len(s); i++ {
			ch := s[i]
			if esc {
				esc = false
				continue
			}
			if ch == '\\' {
				esc = true
				continue
			}
			if ch == quote {
				return i
			}
		}
		return -1
	}

	for i := 0; i < len(s); i++ {
		ch := s[i]

		// Token separators outside quotes
		if ch == ' ' || ch == '\t' {
			flush()
			continue
		}

		// If we see a quote, only treat it as a quote if it has a closing pair.
		if ch == '"' || ch == '\'' {
			closing := findClosingQuote(ch, i+1)
			if closing == -1 {
				// No closing quote -> treat quote as a literal character.
				b.WriteByte(ch)
				continue
			}

			// consume quoted content with escapes.
			for j := i + 1; j < closing; j++ {
				c := s[j]
				if c == '\\' && j+1 < closing {
					// accept escaping inside quotes.
					j++
					b.WriteByte(s[j])
					continue
				}
				b.WriteByte(c)
			}

			i = closing
			continue
		}

		// Regular character
		b.WriteByte(ch)
	}

	flush()

	// drop empty tokens
	clean := out[:0]
	for _, t := range out {
		if strings.TrimSpace(t) != "" {
			clean = append(clean, t)
		}
	}

	return clean
}

// ---
// Shorthand funcs for lazy me.
// And you ^^
// ---

// Info prints an informational message to stdout in blue color.
// Intended for neutral, non-error messages (status, progress, hints).
func Info(s string, args ...any) {
	if len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}

	fmt.Println(aur.Blue(s).String())
}

// Err prints an error message to stdout in red color.
// Used to highlight errors or critical problems for the user.
func Err(s string, args ...any) {
	if len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}

	fmt.Println(aur.Red(s).String())
}

// OK prints a success message prefixed with a checkmark.
// Used to indicate successful completion of an operation.
func OK(s string, args ...any) {
	if len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}

	fmt.Println("✅ " + s)
}

const grayLevel = 12 // from 0 to 24.

// Gray return string in gray color.
func Gray(s string, args ...any) string {
	if len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}

	return aur.Gray(grayLevel, s).String()
}
