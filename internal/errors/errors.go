package errors

import (
	"errors"
)

var DbConnectionFailed = errors.New("db connection failed")
var QueryError = errors.New("query error")
var KeyNotFound = errors.New("key not found")
var InvalidDate = errors.New("date is not valid")
var NotAFloat = errors.New("not a float")
var CountIsEmpty = errors.New("count is empty")
var CountIsLessThanOne = errors.New("count is less than one")
var GroupError = errors.New("group error")
