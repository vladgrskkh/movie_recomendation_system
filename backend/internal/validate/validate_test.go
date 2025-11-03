package validate

import (
	"errors"
	"testing"

	"github.com/invopop/validation"
)

func TestUnique(t *testing.T) {
	tests := []struct {
		name  string
		value []string
		want  error
	}{
		{
			name:  "Unique",
			value: []string{"1", "2", "3"},
			want:  nil,
		},
		{
			name:  "Non-unique",
			value: []string{"drama", "drama", "smth"},
			want:  errUniqueValues,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Validate(&tt.value, validation.By(Unique(tt.value)))
			if !errors.Is(err, tt.want) {
				t.Errorf("got %v; want %v", err, tt.want)
			}
		})
	}
}
