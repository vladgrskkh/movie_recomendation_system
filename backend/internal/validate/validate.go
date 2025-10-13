package validate

import (
	"errors"

	"github.com/invopop/validation"
)

var (
	errUniqueValues = errors.New("values must be unique")
)

// Unique function is a custom rule checking if slice contains unique values
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
