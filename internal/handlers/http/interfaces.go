package httphandlers

import (
	"context"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
)

type UnitOfWorkService interface {
	Submit(ctx context.Context, request yamlcontracts.UnitOfWork) error
}
