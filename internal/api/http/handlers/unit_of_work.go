package handlershttpapi

import (
	"fmt"
	"net/http"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
)

type UnitOfWork struct {
	uowService UnitOfWorkService
}

func NewUnitOfWork(uowService UnitOfWorkService) *UnitOfWork {
	return &UnitOfWork{
		uowService: uowService,
	}
}

func (h *UnitOfWork) Submit(writer http.ResponseWriter, request *http.Request) {
	var uowContract yamlcontracts.UnitOfWork
	cb, err := ParseBodyAndHandleError(writer, request, &uowContract, true)
	if err != nil {
		cb()
		return
	}

	ctx := request.Context()
	if err = h.uowService.Submit(ctx, uowContract); err != nil {
		WriteResult(writer, http.StatusInternalServerError, ContentTypeJSON, GenericResponse{
			Message: "submission failed",
			Error:   fmt.Errorf("failed to submit unit-of-work: %w", err),
		})
		return
	}

	WriteResult(writer, http.StatusCreated, ContentTypeJSON, GenericResponse{
		Message: "ok",
	})
}
