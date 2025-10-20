package responce

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Alias  string `json:"alias,omitempty"`
}

const (
	StatusOk    = "ok"
	StatusError = "error"
)

func OK() Response {
	return Response{Status: StatusOk}
}
func Error(msg string) Response {
	return Response{Status: StatusError, Error: msg}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string
	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("%s is required", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("Invalid URL: %s", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("Invalid: %s", err.Field()))
		}
	}
	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ","),
	}
}
