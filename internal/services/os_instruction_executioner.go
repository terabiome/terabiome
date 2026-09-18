package services

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/terabiome/infrastructures/internal/executor"
	"github.com/terabiome/infrastructures/internal/models"
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

	if err := e.processDiskInstructions(ctx); err != nil {
		return fmt.Errorf("failed to execute disk instructions: %w", err)
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
	shellCfg := e.cfg.Shell

	// ------------
	// --- SSHD ---
	// ------------

	// mutually exclusive
	if shellCfg.DisableService || shellCfg.EnableService {
		input := executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: "systemctl",
			Arguments:  []string{},
		}
		if shellCfg.DisableService {
			input.Arguments = append(input.Arguments, "disable")
		} else {
			input.Arguments = append(input.Arguments, "enable")
		}
		input.Arguments = append(input.Arguments, "sshd")
		output := e.executor.Execute(ctx, &input)

		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to enable/disable sshd: %w", output.Error)
		}
	}

	if shellCfg.StartService {
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: "systemctl",
			Arguments:  []string{"start", "sshd"},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to start sshd: %w", output.Error)
		}
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

	tailscaleCfg := e.cfg.Network.Tailscale

	if tailscaleCfg.InstallPackage {
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: "zypper",
			Arguments:  []string{"install", "-y", "tailscale"},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to install tailscale: %w", output.Error)
		}
	}

	// mutually exclusive
	if tailscaleCfg.DisableService || tailscaleCfg.EnableService {
		input := executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: "systemctl",
			Arguments:  []string{},
		}
		if tailscaleCfg.DisableService {
			input.Arguments = append(input.Arguments, "disable")
		} else {
			input.Arguments = append(input.Arguments, "enable")
		}
		input.Arguments = append(input.Arguments, "tailscaled")
		output := e.executor.Execute(ctx, &input)

		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to enable/disable tailscaled: %w", output.Error)
		}
	}

	if tailscaleCfg.StartService {
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: "systemctl",
			Arguments:  []string{"start", "tailscaled"},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to start tailscaled: %w", output.Error)
		}
	}

	// missing auth-key -> do nothing
	if tailscaleCfg.AuthKey != "" {
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: "tailscale",
			Arguments:  []string{"up", fmt.Sprintf("--auth-key=%s", tailscaleCfg.AuthKey)},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to connect to tailscale network: %w", output.Error)
		}
	}

	return nil
}

func (e *osInstructionExecutioner) processFirewallInstructions(ctx context.Context) error {
	// ----------------
	// --- Firewall ---
	// ----------------

	firewallCfg := e.cfg.Network.Firewall

	// mutually exclusive
	if firewallCfg.DisableService || firewallCfg.EnableService {
		input := executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: "systemctl",
			Arguments:  []string{},
		}
		if firewallCfg.DisableService {
			input.Arguments = append(input.Arguments, "disable")
		} else {
			input.Arguments = append(input.Arguments, "enable")
		}
		input.Arguments = append(input.Arguments, "firewalld")
		output := e.executor.Execute(ctx, &input)

		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to enable/disable firewalld: %w", output.Error)
		}
	}

	for zone, cfg := range firewallCfg.Zones {
		input := executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: "firewall-cmd",
		}

		var (
			interfaceAction string
			serviceAction   string
		)

		// interface
		switch cfg.Interface.Action {
		case models.FirewalldZoneEntityActionAdd:
			interfaceAction = "--add-interface=%s"
		case models.FirewalldZoneEntityActionRemove:
			interfaceAction = "--remove-interface=%s"
		default:
			return fmt.Errorf("not valid action on interface: %s", cfg.Interface.Action)
		}
		interfaceAction = fmt.Sprintf(interfaceAction, cfg.Interface.Name)
		input.Arguments = []string{fmt.Sprintf("--zone=%s", zone), interfaceAction}

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
			case models.FirewalldZoneEntityActionAdd:
				serviceAction = "--add-service=%s"
			case models.FirewalldZoneEntityActionRemove:
				serviceAction = "--remove-service=%s"
			default:
				return fmt.Errorf("not valid action on service: %s", service.Action)
			}
			serviceAction = fmt.Sprintf(serviceAction, service.Name)
			input.Arguments = []string{fmt.Sprintf("--zone=%s", zone), serviceAction}

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
		case models.FirewalldTargetActionAccept, models.FirewalldTargetActionDrop, models.FirewalldTargetActionReject:
		default:
			return fmt.Errorf("not valid target to set: %s", cfg.SetTarget)
		}
		input.Arguments = []string{fmt.Sprintf("--zone=%s", zone), fmt.Sprintf("--set-target=%s", cfg.SetTarget)}

		output = e.executor.Execute(ctx, &input)
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to set target with firewall-cmd (zone=%s, target=%s): %w",
				zone, cfg.SetTarget,
				output.Error,
			)
		}
	}

	if firewallCfg.Immediate {
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: "firewall-cmd",
			Arguments:  []string{"--reload"},
		})

		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to apply firewall changes: %w", output.Error)
		}
	}

	return nil
}

func (e *osInstructionExecutioner) processDiskInstructions(ctx context.Context) error {
	return errors.Join(
		e.processSwapInstructions(ctx),
	)
}

func (e *osInstructionExecutioner) processSwapInstructions(ctx context.Context) error {
	swapCfg := e.cfg.Disk.Swap

	// disable swap
	if swapCfg.DisableSwap {
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: "swapoff",
			Arguments:  []string{"-a"},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to run swapoff: %w", output.Error)
		}
	}

	if swapCfg.EditFstab {
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

	if swapCfg.MaskSwapRelatedServices {
		// list the units
		stdoutBuf := bytes.NewBuffer([]byte{})
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     stdoutBuf,
			Stderr:     os.Stderr,
			Executable: "systemctl",
			Arguments:  []string{"list-units", "--type=swap", "--all", "--no-legend"},
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
			output = e.executor.Execute(ctx, &executor.Input{
				Mode:       executor.ModeSync,
				Stdout:     stdoutBuf,
				Stderr:     os.Stderr,
				Executable: "systemctl",
				Arguments:  []string{"mask", swapUnit},
			})
			if output.Done() && output.Error != nil {
				return fmt.Errorf("failed to mask systemd swap unit %s: %w", swapUnit, output.Error)
			}
		}
	}

	return nil
}
