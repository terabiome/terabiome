package services

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/terabiome/infrastructures/instructions/internal/executor"
	"github.com/terabiome/infrastructures/instructions/internal/models"
)

type osInstructionExecutioner struct {
	executor Executor
	cfg      models.OSConfig
}

func NewOSInstructionExecutioner(executor Executor, cfg models.OSConfig) *osInstructionExecutioner {
	return &osInstructionExecutioner{
		executor: executor,
		cfg:      cfg,
	}
}

func (e *osInstructionExecutioner) Execute(ctx context.Context) error {
	if err := e.processShellInstructions(ctx); err != nil {
		return fmt.Errorf("failed to execute shell instructions: %w", err)
	}

	if err := e.processNetworkInstructions(ctx); err != nil {
		return fmt.Errorf("failed to execute network instructions: %w", err)
	}

	return nil
}

func (e *osInstructionExecutioner) Validate() error {
	return errors.Join(
		e.cfg.Shell.SoftValidate(),
		e.cfg.Network.Tailscale.SoftValidate(),
		e.cfg.Network.Firewall.SoftValidate(),
	)
}

func (e *osInstructionExecutioner) processShellInstructions(ctx context.Context) error {
	// TODO: follow example and iterate through the whole config file

	// ------------
	// --- SSHD ---
	// ------------
	output := e.executor.Execute(ctx, &executor.Input{
		Mode:       executor.ModeSync,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Executable: "systemctl",
		Arguments: []string{
			"enable", "--now",
			"sshd",
		},
	})
	if output.Done() && output.Error != nil {
		return fmt.Errorf("failed to start sshd service: %w", output.Error)
	}

	return nil
}

func (e *osInstructionExecutioner) processNetworkInstructions(ctx context.Context) error {
	return errors.Join(
		e.processTailscaleInstructions(ctx),
		e.processFirewallInstructions(ctx),
	)
}

func (e *osInstructionExecutioner) processTailscaleInstructions(ctx context.Context) error {
	// -----------------
	// --- Tailscale ---
	// -----------------
	output := e.executor.Execute(ctx, &executor.Input{
		Mode:       executor.ModeSync,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Executable: "zypper",
		Arguments: []string{
			"install", "-y",
			"tailscale",
		},
	})
	if output.Done() && output.Error != nil {
		return fmt.Errorf("failed to install tailscale: %w", output.Error)
	}
	output = e.executor.Execute(ctx, &executor.Input{
		Mode:       executor.ModeSync,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Executable: "systemctl",
		Arguments: []string{
			"enable", "--now",
			"tailscaled",
		},
	})
	if output.Done() && output.Error != nil {
		return fmt.Errorf("failed to enable tailscaled: %w", output.Error)
	}

	return nil
}

func (e *osInstructionExecutioner) processFirewallInstructions(ctx context.Context) error {
	return nil
}
