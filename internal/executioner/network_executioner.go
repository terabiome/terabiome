package executioner

import (
	"context"
	"errors"
	"fmt"
	"os"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
	"github.com/terabiome/terabiome/pkg/logging"
)

type networkExecutioner struct {
	cfg      yamlcontracts.NetworkConfig
	executor CommandExecutor
	logger   logging.Logger
}

func (e *networkExecutioner) Validate() error {
	return errors.Join(
		e.cfg.Tailscale.SoftValidate(),
		e.cfg.Firewall.SoftValidate(),
	)
}

func (e *networkExecutioner) Process(ctx context.Context) error {
	return errors.Join(
		e.processTailscale(ctx),
		e.processFirewall(ctx),
	)
}

func (e *networkExecutioner) processTailscale(ctx context.Context) error {
	// -----------------
	// --- Tailscale ---
	// -----------------

	if e.cfg.Tailscale.InstallPackage {
		output := e.executor.Execute(ctx, &shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
				Executable: "zypper",
				Arguments:  []string{"install", "-y", "tailscale"},
			},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to install tailscale: %w", output.Error)
		}
	}

	// mutually exclusive
	if e.cfg.Tailscale.DisableService || e.cfg.Tailscale.EnableService {
		input := shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
				Executable: "systemctl",
				Arguments:  []string{},
			},
		}
		if e.cfg.Tailscale.DisableService {
			input.Command.Arguments = append(input.Command.Arguments, "disable")
		} else {
			input.Command.Arguments = append(input.Command.Arguments, "enable")
		}
		input.Command.Arguments = append(input.Command.Arguments, "tailscaled")
		output := e.executor.Execute(ctx, &input)

		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to enable/disable tailscaled: %w", output.Error)
		}
	}

	if e.cfg.Tailscale.StartService {
		output := e.executor.Execute(ctx, &shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
				Executable: "systemctl",
				Arguments:  []string{"start", "tailscaled"},
			},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to start tailscaled: %w", output.Error)
		}
	}

	// missing auth-key -> do nothing
	if e.cfg.Tailscale.AuthKey != "" {
		output := e.executor.Execute(ctx, &shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
				Executable: "tailscale",
				Arguments:  []string{"up", fmt.Sprintf("--auth-key=%s", e.cfg.Tailscale.AuthKey)},
			},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to connect to tailscale network: %w", output.Error)
		}
	}

	return nil
}

func (e *networkExecutioner) processFirewall(ctx context.Context) error {
	// ----------------
	// --- Firewall ---
	// ----------------

	// mutually exclusive
	if e.cfg.Firewall.DisableService || e.cfg.Firewall.EnableService {
		input := shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
				Executable: "systemctl",
				Arguments:  []string{},
			},
		}
		if e.cfg.Firewall.DisableService {
			input.Command.Arguments = append(input.Command.Arguments, "disable")
		} else {
			input.Command.Arguments = append(input.Command.Arguments, "enable")
		}
		input.Command.Arguments = append(input.Command.Arguments, "firewalld")
		output := e.executor.Execute(ctx, &input)

		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to enable/disable firewalld: %w", output.Error)
		}
	}

	for zone, cfg := range e.cfg.Firewall.Zones {
		input := shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
				Executable: "firewall-cmd",
			},
		}

		var (
			interfaceAction string
			serviceAction   string
		)

		// interface
		switch cfg.Interface.Action {
		case yamlcontracts.FirewalldZoneEntityActionAdd:
			interfaceAction = "--add-interface=%s"
		case yamlcontracts.FirewalldZoneEntityActionRemove:
			interfaceAction = "--remove-interface=%s"
		default:
			return fmt.Errorf("not valid action on interface: %s", cfg.Interface.Action)
		}
		interfaceAction = fmt.Sprintf(interfaceAction, cfg.Interface.Name)
		input.Command.Arguments = []string{fmt.Sprintf("--zone=%s", zone), interfaceAction}

		output := e.executor.Execute(ctx, &input)
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to take action with firewall-cmd (zone=%s, interface=%s, action=%s): %w",
				zone, cfg.Interface.Name, cfg.Interface.Action,
				output.Error,
			)
		}

		// services
		for _, service := range cfg.Services {
			switch service.Action {
			case yamlcontracts.FirewalldZoneEntityActionAdd:
				serviceAction = "--add-service=%s"
			case yamlcontracts.FirewalldZoneEntityActionRemove:
				serviceAction = "--remove-service=%s"
			default:
				return fmt.Errorf("not valid action on service: %s", service.Action)
			}
			serviceAction = fmt.Sprintf(serviceAction, service.Name)
			input.Command.Arguments = []string{fmt.Sprintf("--zone=%s", zone), serviceAction}

			output := e.executor.Execute(ctx, &input)
			if output.Done() && output.Error != nil {
				return fmt.Errorf("failed to set with firewall-cmd (zone=%s, service=%s, action=%s): %w",
					zone, service.Name, service.Action,
					output.Error,
				)
			}
		}

		// target
		switch cfg.SetTarget {
		case yamlcontracts.FirewalldTargetActionAccept, yamlcontracts.FirewalldTargetActionDrop, yamlcontracts.FirewalldTargetActionReject:
		default:
			return fmt.Errorf("not valid target to set: %s", cfg.SetTarget)
		}
		input.Command.Arguments = []string{fmt.Sprintf("--zone=%s", zone), fmt.Sprintf("--set-target=%s", cfg.SetTarget)}

		output = e.executor.Execute(ctx, &input)
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to set target with firewall-cmd (zone=%s, target=%s): %w",
				zone, cfg.SetTarget,
				output.Error,
			)
		}
	}

	if e.cfg.Firewall.Immediate {
		output := e.executor.Execute(ctx, &shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
				Executable: "firewall-cmd",
				Arguments:  []string{"--reload"},
			},
		})

		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to apply firewall changes: %w", output.Error)
		}
	}

	return nil
}
