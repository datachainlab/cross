package types

import "fmt"

// ErrIDNotFound is raised when the id not found in store
type ErrIDNotFound struct {
	id []byte
}

// NewErrIDNotFound creates a new instance of ErrIDNotFound
func NewErrIDNotFound(id []byte) ErrIDNotFound {
	return ErrIDNotFound{id: id}
}

// Error implements error.Error
func (e ErrIDNotFound) Error() string {
	return fmt.Sprintf("id '%x' not found", e.id)
}
