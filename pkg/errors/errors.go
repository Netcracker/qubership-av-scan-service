package errors

import (
	"encoding/json"
	"fmt"
	"io"
)

// APIError is a type used to return errors to external users
type APIError struct {
	// Code is an error code specific for application
	Code string `json:"code"`
	// Status is an HTTP status code related to this error
	Status int `json:"status"`
	// Reason is a static error title/description
	Reason string `json:"reason"`
	// Details is a dynamic part of error containing error details, may be empty
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s %s:%s", e.Code, e.Reason, e.Details)
}

func RequestBodyReadError(err error) *APIError {
	return &APIError{
		"AV-1500",
		500,
		"failed to read request body",
		err.Error(),
	}
}

func UnexpectedError(err error) *APIError {
	return &APIError{
		"AV-1900",
		500,
		"unexpected error",
		err.Error(),
	}
}

func ContentTypeUnsupportedError(unsupported string) *APIError {
	if unsupported == "" {
		unsupported = "empty"
	}
	return &APIError{
		"AV-5000",
		415,
		"unsupported content type",
		fmt.Sprintf("%s content-type not supported", unsupported),
	}
}

func FilenameNotSpecifiedError() *APIError {
	return &APIError{
		"AV-5001",
		415,
		"filename not specified",
		"",
	}
}

func ClamdPingError(err error) *APIError {
	return &APIError{
		"AV-7100",
		500,
		"clamd ping error",
		err.Error(),
	}
}

func ClamdScanError(err error) *APIError {
	return &APIError{
		"AV-7101",
		500,
		"clamd scan error",
		err.Error(),
	}
}

// Parse is used to decode JSON input to APIError
func Parse(r io.Reader) (*APIError, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var apiErr *APIError
	err = json.Unmarshal(data, &apiErr)
	return apiErr, err
}
