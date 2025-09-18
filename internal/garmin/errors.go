package garmin

import "fmt"

// AuthenticationError represents an authentication failure with Garmin Connect.
type AuthenticationError struct {
	Err error
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication failed: %v", e.Err)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Err
}
