package errors

import "errors"

var (
	ErrStoryAlreadyExists = errors.New("story already exists in the database")
)
