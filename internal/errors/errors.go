package errors

import (
	"errors"
)

var (
	ErrDbConnectionFailed = errors.New("db connection failed")
	ErrQueryError         = errors.New("query error")
	ErrStoreError         = errors.New("store creation error")
	ErrKeyNotFound        = errors.New("key not found")
	ErrInvalidDate        = errors.New("date is not valid")
	ErrNotAFloat          = errors.New("not a float")
	ErrCountIsEmpty       = errors.New("count is empty")
	ErrCountIsLessThanOne = errors.New("count is less than one")
	ErrGroupError         = errors.New("group error")
	ErrCurrDBNotFound     = errors.New("curr database name not found")
	ErrCurrCacheNotFound  = errors.New("curr cache name not found")
)
