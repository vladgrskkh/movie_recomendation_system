package validate

import (
	"errors"

	"github.com/invopop/validation"
)

var (
	// errUniqueValues indicates that a slice contains duplicate values.
	errUniqueValues = errors.New("values must be unique")
)

// Unique returns a validation rule that ensures the provided slice contains only unique values.
// It validates the given values, not the input value, and is intended to be used with validation.By().
func Unique[T any](values []T) validation.RuleFunc {
	return func(value interface{}) error {
		uniqueValues := make(map[any]bool)

		for _, v := range values {
			uniqueValues[v] = true
		}
		if len(values) != len(uniqueValues) {
			return errUniqueValues
		}
		return nil
	}
}
