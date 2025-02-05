package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string `json:"status"` //Error, Ok
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{Status: StatusOK}
}
func Error(msg string) Response {
	return Response{Status: StatusError, Error: msg}
}
func ValidationError(err validator.ValidationErrors) Response {
	var errorMsgs []string
	for _, err := range err {
		switch err.ActualTag() {
		case "required":
			errorMsgs = append(errorMsgs, fmt.Sprintf("Field %s is a required field", err.Field()))
		default:
			errorMsgs = append(errorMsgs, fmt.Sprintf("Field %s is not valid", err.Field()))

		}

	}
	return Response{
		Status: StatusError,
	}
}
