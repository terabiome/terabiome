package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/terabiome/infrastructures/internal/executor"
	"github.com/terabiome/infrastructures/internal/models"
	"github.com/terabiome/infrastructures/internal/services"
	"go.yaml.in/yaml/v4"
)

func processFlags(ctx context.Context, cliInput *cliInput) error {
	var (
		cfgStruct *models.Config
		output    executor.Output
		err       error
	)

	if cliInput.bareCommand.Executable != "" {
		output = executor.NewLocalShell().Execute(ctx, &executor.Input{
			Mode:    executor.ModeSync,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
			Command: cliInput.bareCommand,
		})
		output.Wait()
		if output.Error != nil {
			log.Fatalf("failed to execute bare command: %v", output.Error)
		}
		return nil
	}

	cfgStruct, err = models.NewLoader().LoadConfigFromLocalPath(cliInput.yamlConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load YAML config: %w", err)
	}

	if cfgStruct.DebugPrintOnly {
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)

		if err = encoder.Encode(cfgStruct); err != nil {
			return fmt.Errorf("failed to encode to YAML buffer: %w", err)
		}
		fmt.Fprintf(os.Stdout, "%s\n", buf.String())
		return nil
	}

	localShell := executor.NewLocalShell()

	// ----------
	// --- OS ---
	// ----------
	osInsExecutioner := services.NewOSInstructionExecutioner(localShell, cfgStruct.OS)
	if err = osInsExecutioner.Validate(); err != nil {
		return fmt.Errorf("failed to validate config for OS instruction executioner: %w", err)
	}

	if err = osInsExecutioner.Execute(ctx); err != nil {
		return fmt.Errorf("failed to execute OS instructions: %w", err)
	}

	return nil
}

func main() {
	if os.Getuid() != 0 {
		log.Fatalln("this script must be run as root/sudo")
	}

	cliInput := NewCLIInput()
	if err := cliInput.Parse(); err != nil {
		log.Fatalf("failed to parse argument flags: %v\n", err)
	}

	// Ctrl + C or something to stop the command halfway
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err := processFlags(ctx, cliInput); err != nil {
		log.Fatalf("failed to process input: %v\n", err)
	}

	log.Println("completed")
}
