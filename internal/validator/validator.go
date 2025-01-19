package validator

import (
	"regexp"
	"slices"
)

// checking the email format using regex expression
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Define a new Validator type that will contain the validator errors

type Validator struct {
	Errors map[string]string
}

// New is a helper which creates a new Validator instance with a empty Errors map
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// valid return true if the errors map doesn't conatain any entries
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddErrors() adds an error message to the map ( as long as no entries doesn't contain same keys)
func (v *Validator) AddErrors(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message to the map only if the validation is not ok
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddErrors(key, message)
	}
}

// Generic function which return true if a specific value is in a list of permitted values
func PermittedValue[T comparable](value T, PermittedValue ...T) bool {
	return slices.Contains(PermittedValue, value)
}

// Matches return true if a string value matches a specific regexp pattern.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Generic function which returns true if all values in a slice are unique
func Unique[T comparable](values []T) bool {
	uniqueValue := make(map[T]bool)

	for _, value := range values {
		uniqueValue[value] = true
	}
	return len(values) == len(uniqueValue)
}
