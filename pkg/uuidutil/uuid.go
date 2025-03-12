// Package uuidutil provides utility functions for working with UUIDs
package uuidutil

import (
	"github.com/google/uuid"
)

// New generates a new UUID v4 (random)
func New() uuid.UUID {
	return uuid.New()
}

// Parse parses a UUID string into a UUID object
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParse parses a UUID string into a UUID object, panicking on error
func MustParse(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// IsNil checks if a UUID is nil/zero value
func IsNil(id uuid.UUID) bool {
	return id == uuid.Nil
}

// String returns the string representation of a UUID
func String(id uuid.UUID) string {
	return id.String()
}
