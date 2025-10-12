package data

import "errors"

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Placeholder struct for all database models
// will use this later
type Models struct {
	Movie MovieModel
}
