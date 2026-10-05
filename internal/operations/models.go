package operations

import (
	"github.com/terabiome/terabiome/internal/models"
	"github.com/terabiome/terabiome/internal/writers"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type OperationEnvironment struct {
	NodeInfo  models.NodeInfo
	Journaler *writers.Journaler
	Executor  *shellexecutors.CommandExecutor
}

type OperationName string

const (
	OperationNameDemo OperationName = "demo"
)
