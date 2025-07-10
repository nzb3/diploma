package validator

import (
	"fmt"
)

type Validator[T any] interface {
	Validate(validators ...ValidateFunc[T]) error
}

func Validate[T any](obj *T) error {
	if validator, ok := any(obj).(Validator[T]); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
	}

	return nil
}

type ValidateFunc[T any] func(r *T) error
