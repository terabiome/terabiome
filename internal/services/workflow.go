package services

import yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"

type WorkflowService struct {
	strategy BatchingStrategy
}

func NewWorkflowService(strategy BatchingStrategy) *WorkflowService {
	if strategy == nil {
		strategy = &DepthFirstStrategy{} // Default
	}
	return &WorkflowService{strategy: strategy}
}

// PrepareExecutionPlan takes a UOW and returns structured batches ready for the Executor.
func (s *WorkflowService) PrepareExecutionPlan(uow yamlcontracts.UnitOfWork) [][]yamlcontracts.UnitOfWorkSequence {
	return s.strategy.BuildBatches(uow)
}
