package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/valentinesamuel/activelog/pkg/response"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func Validate(i interface{}) error {
	return validate.Struct(i)
}

func FormatValidationErrors(err error) []response.ValidationErrorItem {
	accumulator := make(map[string][]string)
	order := []string{}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			var msg string
			switch e.Tag() {
			case "required":
				msg = fmt.Sprintf("%s should not be empty", field)
			case "min":
				msg = fmt.Sprintf("%s must be at least %s", field, e.Param())
			case "max":
				msg = fmt.Sprintf("%s must be at most %s characters", field, e.Param())
			case "email":
				msg = fmt.Sprintf("%s must be a valid email", field)
			default:
				msg = fmt.Sprintf("%s is invalid", field)
			}
			if _, exists := accumulator[field]; !exists {
				order = append(order, field)
			}
			accumulator[field] = append(accumulator[field], msg)
		}
	}

	result := make([]response.ValidationErrorItem, 0, len(order))
	for _, field := range order {
		result = append(result, response.ValidationErrorItem{
			Field:  field,
			Errors: accumulator[field],
		})
	}
	return result
}
