package db

import "fmt"

// MissingEntryError occurs when a Read operation runs with a key that has no corresponding entry
type MissingEntryError struct {
	message string
}

// NewMissingEntryError creates a new MissingEntryError object
func NewMissingEntryError(key string) *MissingEntryError {
	return &MissingEntryError{
		message: fmt.Sprintf("No entry for key '%s'", key),
	}
}

func (e *MissingEntryError) Error() string {
	return e.message
}
