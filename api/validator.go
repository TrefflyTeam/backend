package api

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

var validUsername validator.Func = func(fl validator.FieldLevel) bool {
	username, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	return regexp.MustCompile(`^\p{L}-]+$`).MatchString(username)
}
