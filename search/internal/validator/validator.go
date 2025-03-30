package validator

import (
	"fmt"
)

type Validator interface {
	Validate() error
}

func Validate[T any](obj *T) error {
	if validator, ok := any(obj).(Validator); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
	}

	return nil
}
