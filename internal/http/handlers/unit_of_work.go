package handlershttp

import (
	"net/http"

	yamlcontracts "github.com/terabiome/infrastructures/internal/contracts/yaml"
)

type UnitOfWork struct {
}

func NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{}
}

func (h *UnitOfWork) Submit(writer http.ResponseWriter, request *http.Request) {
	var uowContract yamlcontracts.UnitOfWork
	cb, err := ParseBodyAndHandleError(writer, request, &uowContract, true)
	if err != nil {
		cb()
		return
	}

	// TODO: add more logic here
	WriteResult(writer, http.StatusCreated, ContentTypeJSON, GenericResponse{
		Body:    uowContract,
		Message: "ok",
	})
}
