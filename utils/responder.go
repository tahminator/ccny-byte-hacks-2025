package utils

type ApiResponder[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Payload *T     `json:"payload,omitempty"`
}

// No payload will be returned.
func Failure(message string) *ApiResponder[any] {
	return &ApiResponder[any]{
		Success: false,
		Message: message,
		Payload: nil,
	}
}

// payload will be returned and typed as T
func Success[T any](message string, payload T) *ApiResponder[T] {
	return &ApiResponder[T]{
		Success: true,
		Message: message,
		Payload: &payload,
	}
}


