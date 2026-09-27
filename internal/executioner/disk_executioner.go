package executioner

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type diskExecutioner struct {
	cfg      yamlcontracts.DiskConfig
	executor CommandExecutor
	logger   Logger
}

func (e *diskExecutioner) Validate() error {
	return nil
}

func (e *diskExecutioner) Process(ctx context.Context) error {
	return errors.Join(
		e.processSwap(ctx),
	)
}

func (e *diskExecutioner) processSwap(ctx context.Context) error {
	// disable swap
	if e.cfg.Swap.DisableSwap {
		output := e.executor.Execute(ctx, &shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
				Executable: "swapoff",
				Arguments:  []string{"-a"},
			},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to run swapoff: %w", output.Error)
		}
	}

	if e.cfg.Swap.EditFstab {
		content, err := os.ReadFile("/etc/fstab")
		if err != nil {
			return fmt.Errorf("failed to read /etc/fstab: %w", err)
		}
		lines := strings.Split(string(content), "\n")
		var modified []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				modified = append(modified, line)
				continue
			}
			if strings.Contains(strings.ToLower(line), "swap") {
				log.Printf("Commenting out swap entry: '%s'\n", line)
				modified = append(modified, "# "+line)
			} else {
				modified = append(modified, line)
			}
		}

		if err := os.WriteFile("/etc/fstab", []byte(strings.Join(modified, "\n")), 0644); err != nil {
			return fmt.Errorf("failed to write /etc/fstab: %w", err)
		}
	}

	if e.cfg.Swap.MaskSwapRelatedServices {
		// list the units
		stdoutBuf := bytes.NewBuffer([]byte{})
		output := e.executor.Execute(ctx, &shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: stdoutBuf,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
				Executable: "systemctl",
				Arguments:  []string{"list-units", "--type=swap", "--all", "--no-legend"},
			},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to list swap units: %w", output.Error)
		}

		// mask them
		var swapUnits []string
		outputScanner := bufio.NewScanner(stdoutBuf)
		for outputScanner.Scan() {
			if outputScanner.Err() != nil {
				return fmt.Errorf("failed to scan swap unit output: %w", outputScanner.Err())
			}
			fields := strings.Fields(outputScanner.Text())
			if len(fields) == 0 {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(fields[0]), ".swap") {
				continue
			}
			swapUnits = append(swapUnits, fields[0])
		}
		for _, swapUnit := range swapUnits {
			log.Printf("Masking systemd unit %s\n", swapUnit)
			output = e.executor.Execute(ctx, &shellexecutors.CommandInput{
				Mode:   shellexecutors.ModeSync,
				Stdout: io.Discard, // idc
				Stderr: os.Stderr,
				Command: shellexecutors.Command{
					Executable: "systemctl",
					Arguments:  []string{"mask", swapUnit},
				},
			})
			if output.Done() && output.Error != nil {
				return fmt.Errorf("failed to mask systemd swap unit %s: %w", swapUnit, output.Error)
			}
		}
	}

	return nil
}
