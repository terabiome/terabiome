package registry

import (
	"fmt"

	"github.com/terabiome/terabiome/internal/operations"
)

var registry = map[string]func() operations.Operation{}

func Get(name string) (operations.Operation, error) {
	factory, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("task '%s' not found in registry", name)
	}
	return factory(), nil
}

func ValidateAll(uow any) error {
	return nil
}
