package errors

import (
	"errors"
)

var ErrDbConnectionFailed = errors.New("db connection failed")
var ErrQueryError = errors.New("query error")
var ErrKeyNotFound = errors.New("key not found")
var ErrInvalidDate = errors.New("date is not valid")
var ErrNotAFloat = errors.New("not a float")
var ErrCountIsEmpty = errors.New("count is empty")
var ErrCountIsLessThanOne = errors.New("count is less than one")
