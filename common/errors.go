package common

import "fmt"

var (
	ErrInvalidValues     = fmt.Errorf("invalid values")
	ErrNoRowsAffected    = fmt.Errorf("no rows affected")
	ErrOperationNotFound = fmt.Errorf("operation not found")
	// ErrOperationReverted: the latest operation for this email pair was already reverted.
	ErrOperationReverted = fmt.Errorf("operation already reverted")
	// ErrOperationNoMatch: no participant has the email and keycloak id the
	// rename (or its revert) starts from, so nothing was changed.
	ErrOperationNoMatch = fmt.Errorf("no participant matches this email and keycloak id")
)
