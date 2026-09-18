package executioner

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

type osExecutioner struct {
	// sub-executioners
	shellExecutioner   shellExecutioner
	networkExecutioner networkExecutioner
	diskExecutioner    diskExecutioner
}

func NewOSExecutioner(executor Executor, cfg models.OSConfig) *osExecutioner {
	return &osExecutioner{
		shellExecutioner:   shellExecutioner{cfg.Shell, executor},
		networkExecutioner: networkExecutioner{cfg.Network, executor},
		diskExecutioner:    diskExecutioner{cfg.Disk, executor},
	}
}

func (e *osExecutioner) Execute(ctx context.Context) error {
	if err := e.shellExecutioner.Process(ctx); err != nil {
		return fmt.Errorf("failed to execute shell instructions: %w", err)
	}

	if err := e.networkExecutioner.Process(ctx); err != nil {
		return fmt.Errorf("failed to execute network instructions: %w", err)
	}

	if err := e.diskExecutioner.Process(ctx); err != nil {
		return fmt.Errorf("failed to execute disk instructions: %w", err)
	}

	return nil
}

func (e *osExecutioner) Validate() error {
	return errors.Join(
		e.shellExecutioner.Validate(),
		e.networkExecutioner.Validate(),
		e.diskExecutioner.Validate(),
	)
}

type shellExecutioner struct {
	cfg      models.ShellConfig
	executor Executor
}

func (e *shellExecutioner) Validate() error {
	return errors.Join(
		e.cfg.SoftValidate(),
	)
}

func (e *shellExecutioner) Process(ctx context.Context) error {
	// ------------
	// --- SSHD ---
	// ------------

	// mutually exclusive
	if e.cfg.DisableService || e.cfg.EnableService {
		input := executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
				Executable: "systemctl",
				Arguments:  []string{},
			},
		}
		if e.cfg.DisableService {
			input.Command.Arguments = append(input.Command.Arguments, "disable")
		} else {
			input.Command.Arguments = append(input.Command.Arguments, "enable")
		}
		input.Command.Arguments = append(input.Command.Arguments, "sshd")
		output := e.executor.Execute(ctx, &input)

		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to enable/disable sshd: %w", output.Error)
		}
	}

	if e.cfg.StartService {
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
				Executable: "systemctl",
				Arguments:  []string{"start", "sshd"},
			},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to start sshd: %w", output.Error)
		}
	}

	return nil
}

type networkExecutioner struct {
	cfg      models.NetworkConfig
	executor Executor
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
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
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
		input := executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
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
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
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
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
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
		input := executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
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
		input := executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
				Executable: "firewall-cmd",
			},
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
			case models.FirewalldZoneEntityActionAdd:
				serviceAction = "--add-service=%s"
			case models.FirewalldZoneEntityActionRemove:
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
		case models.FirewalldTargetActionAccept, models.FirewalldTargetActionDrop, models.FirewalldTargetActionReject:
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
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
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

type diskExecutioner struct {
	cfg      models.DiskConfig
	executor Executor
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
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
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
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:   executor.ModeSync,
			Stdout: stdoutBuf,
			Stderr: os.Stderr,
			Command: models.Command{
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
			output = e.executor.Execute(ctx, &executor.Input{
				Mode:   executor.ModeSync,
				Stdout: stdoutBuf,
				Stderr: os.Stderr,
				Command: models.Command{
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
