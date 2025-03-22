package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type Response struct {
	Status  int               `json:"status"` //Message, Ok
	Message string            `json:"error,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func OK() Response {
	return Response{Status: 200}
}
func Error(msg string, status int) Response {
	return Response{Status: status, Message: msg}
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
		Status:  400,
		Message: strings.Join(errorMsgs, ", "),
	}
}
