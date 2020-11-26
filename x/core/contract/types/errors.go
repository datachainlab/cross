package types

import "fmt"

// ErrContractCall is raised when the caller is failed to call the contract
type ErrContractCall struct {
	err error
}

// NewErrContractCall creates a new instance of ErrContractCall
func NewErrContractCall(err error) ErrContractCall {
	return ErrContractCall{err: err}
}

// Error implements error.Error
func (e ErrContractCall) Error() string {
	return fmt.Sprintf("failed to call the contract: %v", e.err)
}
