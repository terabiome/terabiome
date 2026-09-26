package contracts

import "github.com/go-playground/validator/v10"

type Validatable interface {
	Validate(validator *validator.Validate) error
}
