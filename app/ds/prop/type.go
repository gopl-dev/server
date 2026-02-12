// Package prop provides property type definitions and utilities for handling
// different kinds of properties .
package prop

import "slices"

// Type represents the data type of property.
type Type string

// Supported data types for properties.
const (
	Unknown  Type = "unknown"
	String   Type = "string"
	Text     Type = "text"
	Markdown Type = "markdown"
	URL      Type = "url"
	Image    Type = "image"
	List     Type = "list"
)

// Patchable returns true if the property type can be modified through patch operations.
func (t Type) Patchable() bool {
	switch t {
	case String, Text, Markdown, URL:
		return true
	}

	return false
}

// Is checks if the type matches any of the provided types.
func (t Type) Is(tt ...Type) bool {
	return slices.Contains(tt, t)
}
