package resourcemodel

import (
	"errors"
)

var ErrNil = errors.New("received nil")

type ResourceValidationError error

var (
	ErrorMissingName       ResourceValidationError = errors.New("name is missing")
	ErrorMissingType       ResourceValidationError = errors.New("type is missing")
	ErrorMissingRawContent ResourceValidationError = errors.New("raw_content is missing")
	ErrorMissingID         ResourceValidationError = errors.New("id is missing")
	ErrorMissingOwnerID    ResourceValidationError = errors.New("owner is missing")
	ErrorWrongType         ResourceValidationError = errors.New("type is wrong")
)
