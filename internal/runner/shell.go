package runner

import (
	"context"
	"fmt"
	"runtime"
	"time"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
	"github.com/terabiome/terabiome/internal/executor"
	"github.com/terabiome/terabiome/internal/writers"

	"github.com/terabiome/terabiome/internal/models"
	"github.com/terabiome/terabiome/internal/operations"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
	loggingpkg "github.com/terabiome/terabiome/pkg/logging"
)

type ShellRunner struct {
	logger          loggingpkg.Logger
	workflowService WorkflowService
}

func NewShellRunner(
	logger loggingpkg.Logger,
	workflowService WorkflowService,
) *ShellRunner {
	return &ShellRunner{
		logger:          logger,
		workflowService: workflowService,
	}
}

func (r *ShellRunner) Run(ctx context.Context, uow yamlcontracts.UnitOfWork) error {
	// 1. Setup Local Environment
	journaler, _ := writers.NewJournaler(writers.JournalerParameters{
		MaxRecords:    50,
		FlushInterval: 1 * time.Second,
		BatchFn: func(entries []writers.JournalEntry) {
			for _, e := range entries {
				r.logger.InfoContext(ctx, e.Message,
					"timestamp", e.Timestamp.Format(time.RFC3339),
					"source", e.Source,
				)
			}
		},
	})
	defer journaler.Close()

	env := &operations.OperationEnvironment{
		NodeInfo:  r.probeNode(),
		Journaler: journaler,
		Executor:  shellexecutors.NewCommandExecutor(r.logger),
	}

	// 2. Execute using Intrinsic
	batches := r.workflowService.PrepareExecutionPlan(uow)
	wfExec := executor.NewWorkflowExecutor(env)

	fmt.Println("Starting local execution...")
	return wfExec.Execute(ctx, batches)
}

func (r *ShellRunner) probeNode() models.NodeInfo {
	return models.NodeInfo{
		MachineID: "local-shell",
		Arch:      runtime.GOARCH,
		OS:        runtime.GOOS,
	}
}
