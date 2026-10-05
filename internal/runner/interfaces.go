package runner

import yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"

type WorkflowService interface {
	PrepareExecutionPlan(uow yamlcontracts.UnitOfWork) [][]yamlcontracts.UnitOfWorkSequence
}
