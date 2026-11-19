package common

import "fmt"

var (
	ErrInvalidValues     = fmt.Errorf("invalid values")
	ErrNoRowsAffected    = fmt.Errorf("no rows affected")
	ErrOperationNotFound = fmt.Errorf("operation not found")
)
