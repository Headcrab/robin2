package errors

import (
	"errors"
)

var ErrInvalidDate = errors.New("date is not valid")
var ErrEmptyDate = errors.New("date is empty")
var ErrNotAFloat = errors.New("not a float")
