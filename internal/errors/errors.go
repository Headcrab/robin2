package errors

import (
	"errors"
)

var (
	DbConnectionFailed = errors.New("db connection failed")
	QueryError         = errors.New("query error")
	StoreError         = errors.New("store creation error")
	KeyNotFound        = errors.New("key not found")
	InvalidDate        = errors.New("date is not valid")
	NotAFloat          = errors.New("not a float")
	CountIsEmpty       = errors.New("count is empty")
	CountIsLessThanOne = errors.New("count is less than one")
	GroupError         = errors.New("group error")
)
