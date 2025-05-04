package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type Response struct {
	StatusCode int         `json:"statusCode" validate:"required"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

func Error(msg string, status int) Response {
	return Response{StatusCode: status, Message: msg}
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
		StatusCode: 400,
		Message:    "Validation failed",
		Data: map[string]string{
			"errors": strings.Join(errorMsgs, ", "),
		},
	}
}
