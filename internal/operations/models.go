package operations

import (
	"github.com/terabiome/terabiome/internal/logging"
	"github.com/terabiome/terabiome/internal/models"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type OperationEnvironment struct {
	NodeInfo models.NodeInfo
	Jounr    *logging.TaskJournaler
	Executor *shellexecutors.CommandExecutor
}
