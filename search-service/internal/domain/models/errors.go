package models

import (
	"errors"
)

var ErrNil = errors.New("received nil")

type ResourceValidationError error

var (
	ValidationErrorMissingName       ResourceValidationError = errors.New("name is missing")
	ValidationErrorMissingType       ResourceValidationError = errors.New("type is missing")
	ValidationErrorMissingRawContent ResourceValidationError = errors.New("raw_content is missing")
	ValidationErrorMissingID         ResourceValidationError = errors.New("id is missing")
	ValidationErrorMissingOwnerID    ResourceValidationError = errors.New("owner is missing")
)
