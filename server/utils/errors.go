package utils

import "fmt"

// Error is an implementation of the golang error.
// It provides storage for extra fields.
type Error struct {
	Code    int                    `json:"code,omitempty"`
	Message string                 `json:"message,omitempty"`
	Args    map[string]interface{} `json:"args,omitempty"`
	Err     error                  `json:"err,omitempty"`
}

// ErrorParams are the arguments for creating an error
type ErrorParams struct {
	// Code is the HTTP status code
	Code int
	// Message is the error message
	Message string
	// Args holds any metadata around the arguments
	// To easily support generic arguments, args is an array
	// where even indexes are keys and odd indexes are values
	Args []interface{}
	// Err is the raw source message
	Err error
}

// NewError returns a new error
func NewError(params ErrorParams) error {
	// Convert params into map
	// This will properly panic if there's an odd number of args
	argsMap := make(map[string]interface{})
	for i := 0; i < len(params.Args); i += 2 {
		argsMap[params.Args[i].(string)] = params.Args[i+1]
	}

	return &Error{
		Code:    params.Code,
		Message: params.Message,
		Args:    argsMap,
		Err:     params.Err,
	}
}

// Error pretty prints the error string
func (e *Error) Error() string {
	return fmt.Sprintf("%v: code: %v, args: %+v, err: %+v", e.Message, e.Code, e.Args, e.Err)
}
