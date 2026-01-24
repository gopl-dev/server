// Package ds (Data Structure)
// All data models belonging to the app are stored here.
// Global functions should not be created here.
// Methods on types are welcome.
package ds

import (
	"time"
)

type ctxKey string

const (
	// PerPageDefault ...
	PerPageDefault = 25
	// PerPageMax ...
	PerPageMax = 100
)

// DataProvider exposes entity data in a generic key-value form.
//
// It is used to extract mutable fields of an entity for diffing,
// change requests, and edit workflows.
type DataProvider interface {
	Data() map[string]any
}

// FilterDT ...
type FilterDT struct {
	DT   *time.Time
	From *time.Time
	To   *time.Time
}

// DtAt ...
func DtAt(t time.Time) *FilterDT {
	return &FilterDT{
		DT: &t,
	}
}

// DtBefore ...
func DtBefore(t time.Time) *FilterDT {
	return &FilterDT{
		To: &t,
	}
}

// DtAfter ...
func DtAfter(t time.Time) *FilterDT {
	return &FilterDT{
		From: &t,
	}
}

// DtBetween ...
func DtBetween(from, to time.Time) *FilterDT {
	return &FilterDT{
		From: &from,
		To:   &to,
	}
}
