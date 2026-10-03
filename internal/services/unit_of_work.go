package services

import (
	"context"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
)

type UnitOfWork struct {
}

func NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{}
}

func (s *UnitOfWork) Submit(ctx context.Context, request yamlcontracts.UnitOfWork) error {
	// TODO: YAML is validated to be correct, now check its content
	return nil
}
