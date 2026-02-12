// Package diff provides word-level diff visualization with context-aware hunk merging.
//
// This package is a wrapper around github.com/sergi/go-diff/diffmatchpatch that enhances
// the diff output by:
//   - Tokenizing text into words, whitespace, and punctuation
//   - Adding configurable context around changes
//   - Merging nearby changes into unified hunks
//   - Rendering diffs as HTML (with <del>/<ins> tags) or colored terminal output
package diff

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/gopl-dev/server/app"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	// DelBgColor is the ANSI escape code for deleted text background.
	DelBgColor = "\033[48;2;255;200;200m"

	// InsBgColor is the ANSI escape code for inserted text background.
	InsBgColor = "\033[48;2;200;255;200m"

	// ResetBgColor is the ANSI escape code to reset all formatting to default.
	ResetBgColor = "\033[0m"
)

// ContextWordsDefault is the default number of words to show around each change.
var ContextWordsDefault = 20

// Diff represents a computed difference between two texts with rendering capabilities.
type Diff struct {
	Text1, Text2 string
	ContextWords int
	Diff         []diffmatchpatch.Diff
}

// Compute calculates the difference between two texts and returns a diff object.
// The optional ContextWords parameter controls how many words of context to show
// around each change. If not provided, ContextWordsDefault is used.
//
// Usage:
//
//	diffHTML := diff.Compute("old text", "new text").HTML()
//	diffString := diff.Compute("old text", "new text", 5).String() // custom context size
func Compute(text1, text2 string, contextWordsOpt ...int) *Diff {
	cw := ContextWordsDefault
	if len(contextWordsOpt) > 0 {
		cw = contextWordsOpt[0]
	}

	d := &Diff{
		Text1:        text1,
		Text2:        text2,
		ContextWords: cw,
		Diff:         make([]diffmatchpatch.Diff, 0),
	}

	if text1 == text2 {
		return d
	}

	dmp := diffmatchpatch.New()
	dmp.DiffTimeout = 0

	diff := dmp.DiffMain(text1, text2, true)
	diff = dmp.DiffCleanupSemanticLossless(diff)
	diff = dmp.DiffCleanupSemantic(diff)

	// If changes are too extensive, simplify to full replacement
	if shouldSimplifyDiff(diff) {
		diff = []diffmatchpatch.Diff{
			{Type: diffmatchpatch.DiffDelete, Text: text1},
			{Type: diffmatchpatch.DiffInsert, Text: text2},
		}
	}

	d.Diff = diff
	return d
}

// shouldSimplifyDiff determines if the diff is too complex and should be simplified
// to a full replacement. Returns true if the ratio of changed characters to total
// characters exceeds 50%.
func shouldSimplifyDiff(diff []diffmatchpatch.Diff) bool {
	var equalChars, totalChars int

	for _, d := range diff {
		totalChars += len(d.Text)
		if d.Type == diffmatchpatch.DiffEqual {
			equalChars += len(d.Text)
		}
	}

	if totalChars == 0 {
		return false
	}

	// If less than 50% of characters are unchanged, show full replacement
	changeRatio := float64(totalChars-equalChars) / float64(totalChars)
	return changeRatio > 0.5 //nolint:mnd
}

// ComputeFromPatch applies a patch to text1 and computes the difference.
func ComputeFromPatch(text1, patch string, contextWordsOpt ...int) (*Diff, error) {
	text2, err := app.ApplyPatch(text1, patch)
	if err != nil {
		err = fmt.Errorf("compute from patch: %w", err)
		return nil, err
	}

	return Compute(text1, text2, contextWordsOpt...), nil
}

// HTML returns an HTML representation of the diff with <del> and <ins> tags.
// Multiple hunks are separated by <hr/> tags. Returns empty string if texts are identical.
//
// Example output:
//
//	You w<del>o</del><ins>a</ins>nt to contribute
func (d *Diff) HTML() string {
	if len(d.Diff) == 0 {
		return ""
	}

	hunks := buildHunks(d.Diff, d.ContextWords)
	var b strings.Builder
	for i, h := range hunks {
		if i > 0 {
			b.WriteString("\n<hr/>\n")
		}

		for _, part := range h {
			text := renderTokens(part.tokens)
			switch part.typ {
			case diffmatchpatch.DiffEqual:
				b.WriteString(text)
			case diffmatchpatch.DiffDelete:
				b.WriteString("<del>" + text + "</del>")
			case diffmatchpatch.DiffInsert:
				b.WriteString("<ins>" + text + "</ins>")
			}
		}
	}

	return b.String()
}

// String returns a colored terminal representation of the diff using ANSI colors.
// Deletes are shown with red background, inserts with green background.
// Multiple hunks are separated by ~~~ markers. Returns empty string if texts are identical.
func (d *Diff) String() string {
	if len(d.Diff) == 0 {
		return ""
	}

	hunks := buildHunks(d.Diff, d.ContextWords)
	var b strings.Builder
	for i, h := range hunks {
		if i > 0 {
			b.WriteString("\n~~~\n")
		}

		for _, part := range h {
			text := renderTokens(part.tokens)
			switch part.typ {
			case diffmatchpatch.DiffEqual:
				b.WriteString(text)
			case diffmatchpatch.DiffDelete:
				b.WriteString(DelBgColor + text + ResetBgColor)
			case diffmatchpatch.DiffInsert:
				b.WriteString(InsBgColor + text + ResetBgColor)
			}
		}
	}

	return b.String()
}

var (
	// wordRe matches words, whitespace, and punctuation as separate tokens.
	wordRe = regexp.MustCompile(`\w+|\s+|[^\w\s]+`)
	// isWordRe checks if a token is a word (alphanumeric only).
	isWordRe = regexp.MustCompile(`^\w+$`)
)

// token represents a single word, whitespace, or punctuation mark.
type token struct {
	text   string
	isWord bool
}

// hunkPart represents a segment of a diff hunk (equal, delete, or insert).
type hunkPart struct {
	typ    diffmatchpatch.Operation
	tokens []token
}

// hunk represents a group of related changes with surrounding context.
type hunk []hunkPart

// tokenize splits a string into tokens (words, whitespace, punctuation).
func tokenize(s string) []token {
	raw := wordRe.FindAllString(s, -1)
	out := make([]token, 0, len(raw))

	for _, r := range raw {
		out = append(out, token{
			text:   r,
			isWord: isWordRe.MatchString(r),
		})
	}

	return out
}

// diffToTokens converts raw diffs into tokenized hunk parts.
func diffToTokens(diffs []diffmatchpatch.Diff) []hunkPart {
	out := make([]hunkPart, 0, len(diffs))
	for _, d := range diffs {
		toks := tokenize(d.Text)
		if len(toks) == 0 {
			continue
		}
		out = append(out, hunkPart{
			typ:    d.Type,
			tokens: toks,
		})
	}

	return out
}

// takeWordContext extracts up to maxWords from tokens, either from the start or end.
func takeWordContext(tokens []token, maxWords int, fromEnd bool) []token {
	if fromEnd {
		return takeWordsFromEnd(tokens, maxWords)
	}

	return takeWordsFromStart(tokens, maxWords)
}

// takeWordsFromEnd extracts up to maxWords from the end of tokens, preserving order.
func takeWordsFromEnd(tokens []token, maxWords int) []token {
	var out []token
	count := 0
	for i := len(tokens) - 1; i >= 0; i-- {
		out = append([]token{tokens[i]}, out...)
		if tokens[i].isWord {
			count++
			if count == maxWords {
				break
			}
		}
	}

	return out
}

// takeWordsFromStart extracts up to maxWords from the start of tokens.
func takeWordsFromStart(tokens []token, maxWords int) []token {
	out := make([]token, 0, len(tokens))
	count := 0
	for _, t := range tokens {
		out = append(out, t)
		if t.isWord {
			count++
			if count == maxWords {
				break
			}
		}
	}

	return out
}

// countWords returns the number of word tokens (excluding whitespace and punctuation).
func countWords(tokens []token) int {
	count := 0
	for _, t := range tokens {
		if t.isWord {
			count++
		}
	}
	return count
}

// renderTokens converts tokens to HTML-escaped text.
func renderTokens(ts []token) string {
	var b strings.Builder
	for _, t := range ts {
		b.WriteString(html.EscapeString(t.text))
	}
	return b.String()
}

// mergeCloseHunks combines hunks that are separated by contextWordsDefault or fewer words.
// This prevents splitting closely related changes into separate hunks.
func mergeCloseHunks(hunks []hunk, contextWords int) []hunk {
	if len(hunks) <= 1 {
		return hunks
	}

	var merged []hunk
	current := hunks[0]

	for i := 1; i < len(hunks); i++ {
		next := hunks[i]

		// Extract context information from current and next hunks.
		var currentAfterWords, nextBeforeWords int
		var currentAfterIdx, nextBeforeIdx = -1, -1

		if len(current) > 0 {
			lastIdx := len(current) - 1
			lastPart := current[lastIdx]
			if lastPart.typ == diffmatchpatch.DiffEqual {
				currentAfterWords = countWords(lastPart.tokens)
				currentAfterIdx = lastIdx
			}
		}
		if len(next) > 0 {
			firstPart := next[0]
			if firstPart.typ == diffmatchpatch.DiffEqual {
				nextBeforeWords = countWords(firstPart.tokens)
				nextBeforeIdx = 0
			}
		}

		// Check if the context regions are identical (from same equal block).
		afterText := ""
		beforeText := ""
		if currentAfterIdx >= 0 {
			afterText = renderTokens(current[currentAfterIdx].tokens)
		}
		if nextBeforeIdx >= 0 {
			beforeText = renderTokens(next[nextBeforeIdx].tokens)
		}

		// Calculate distance: if contexts are identical, count only once.
		distanceWords := currentAfterWords + nextBeforeWords
		if afterText == beforeText && afterText != "" {
			distanceWords = currentAfterWords
		}

		// Merge hunks if they're close enough.
		if distanceWords <= contextWords {
			if afterText == beforeText && afterText != "" {
				// Identical contexts: skip the duplicate and merge remaining parts.
				current = append(current, next[1:]...)
			} else {
				// Different contexts: combine both into a middle section.
				var middle []token
				if currentAfterIdx >= 0 {
					middle = append(middle, current[currentAfterIdx].tokens...)
					current = current[:currentAfterIdx]
				}
				if nextBeforeIdx >= 0 {
					middle = append(middle, next[nextBeforeIdx].tokens...)
				}

				if len(middle) > 0 {
					current = append(current, hunkPart{
						typ:    diffmatchpatch.DiffEqual,
						tokens: middle,
					})
				}

				startIdx := 0
				if nextBeforeIdx >= 0 {
					startIdx = 1
				}
				current = append(current, next[startIdx:]...)
			}
		} else {
			// Too far apart: finalize current hunk and start new one.
			merged = append(merged, current)
			current = next
		}
	}

	merged = append(merged, current)
	return merged
}

// buildHunks constructs hunks from tokenized diffs, adding context around changes
// and merging nearby hunks based on contextWordsDefault distance.
func buildHunks(diffs []diffmatchpatch.Diff, contextWords int) []hunk {
	tokens := diffToTokens(diffs)
	var hunks []hunk
	i := 0

	for i < len(tokens) {
		// Skip equal sections between hunks
		if tokens[i].typ == diffmatchpatch.DiffEqual {
			i++
			continue
		}

		h := hunk{}

		// Add left context from previous equal section
		if i > 0 && tokens[i-1].typ == diffmatchpatch.DiffEqual {
			beforeToks := takeWordContext(tokens[i-1].tokens, contextWords, true)
			h = append(h, hunkPart{
				typ:    diffmatchpatch.DiffEqual,
				tokens: beforeToks,
			})
		}

		// Collect all consecutive changes (deletes and inserts)
		for i < len(tokens) && tokens[i].typ != diffmatchpatch.DiffEqual {
			h = append(h, hunkPart{
				typ:    tokens[i].typ,
				tokens: tokens[i].tokens,
			})
			i++
		}

		// Add right context from next equal section
		if i < len(tokens) && tokens[i].typ == diffmatchpatch.DiffEqual {
			afterToks := takeWordContext(tokens[i].tokens, contextWords, false)
			h = append(h, hunkPart{
				typ:    diffmatchpatch.DiffEqual,
				tokens: afterToks,
			})
		}

		hunks = append(hunks, h)
	}

	return mergeCloseHunks(hunks, contextWords)
}
