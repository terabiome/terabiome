package services

import yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"

// type Repository interface {
// 	CreateUnitOfWork(ctx context.Context, uow *models.UnitOfWork, tasks []models.LinearLog) error
// 	GetUnitOfWork(ctx context.Context, id int64) (*models.UnitOfWork, error)
// 	UpdateTaskStatus(ctx context.Context, taskID string, status string, errMsg *string) error
// 	AppendTaskLogs(ctx context.Context, taskID string, logs []models.TaskLog) error
// }

// BatchingStrategy defines how a UnitOfWork should be split into execution batches.
type BatchingStrategy interface {
	BuildBatches(uow yamlcontracts.UnitOfWork) [][]yamlcontracts.UnitOfWorkSequence
}
