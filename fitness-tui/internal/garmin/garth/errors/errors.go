package errors

type GarthError struct {
	Message string
	Cause   error
}

func (e *GarthError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *GarthError) Unwrap() error {
	return e.Cause
}

type GarthHTTPError struct {
	GarthError
	StatusCode int
	Response   string
}

type APIError struct {
	GarthHTTPError
}

type IOError struct {
	GarthError
}

// AuthenticationError represents an authentication failure (e.g., invalid credentials)
type AuthenticationError struct {
	GarthError
}

// OAuthError represents an OAuth authentication error
type OAuthError struct {
	GarthError
}

// ValidationError represents a data validation error
type ValidationError struct {
	GarthError
}
