package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/terabiome/infrastructures/instructions/internal/executor"
	"github.com/terabiome/infrastructures/instructions/internal/models"
	"go.yaml.in/yaml/v4"
)

type CLIInput struct {
	YAMLConfigPath string
	bareCommand    struct {
		executable string
		arguments  []string
	}
}

func parse() (*CLIInput, error) {
	var tempAnchor struct {
		yamlConfigPath *string
		bareCommand    *string
	}
	tempAnchor.bareCommand = flag.String("bare-command", "", "Only execute command and exit")
	tempAnchor.yamlConfigPath = flag.String("yaml-config-path", "", "Path to YAML config path")

	// parse and check
	flag.Parse()
	cliInput := CLIInput{}

	if tempAnchor.bareCommand != nil {
		if flagVal := *tempAnchor.bareCommand; flagVal != "" {
			segments := strings.Split(flagVal, " ")
			if len(segments) == 0 {
				return nil, errors.New("no segment in -bare-command")
			}

			cliInput.bareCommand = struct {
				executable string
				arguments  []string
			}{
				executable: segments[0],
				arguments:  segments[1:],
			}
			// short-circuit immediately
			return &cliInput, nil
		}
	}

	if tempAnchor.yamlConfigPath == nil {
		return nil, errors.New("missing yaml-config-path")
	} else if flagVal := *tempAnchor.yamlConfigPath; flagVal == "" {
		return nil, errors.New("yaml-config-path is present but empty")
	} else {
		cliInput.YAMLConfigPath = flagVal
	}

	return &cliInput, nil
}

func processFlags(ctx context.Context, cliInput *CLIInput) error {
	var (
		cfgStruct *models.Config
		output    executor.Output
		err       error
	)

	if cliInput.bareCommand.executable != "" {
		output = executor.NewLocalShell().Execute(ctx, &executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: cliInput.bareCommand.executable,
			Arguments:  cliInput.bareCommand.arguments,
		})
		output.Wait()
		if output.Error != nil {
			log.Fatalf("failed to execute bare command: %v", output.Error)
		}
		return nil
	}

	cfgStruct, err = models.NewLoader().LoadConfigFromLocalPath(cliInput.YAMLConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load YAML config: %w", err)
	}

	if cfgStruct.DebugPrintOnly {
		var buf bytes.Buffer
		if err = yaml.NewEncoder(&buf).Encode(cfgStruct); err != nil {
			return fmt.Errorf("failed to encode to YAML buffer: %w", err)
		}
		fmt.Fprintf(os.Stdout, "%s\n", buf.String())
		return nil
	}

	// Barricade: read from Config and start figuring out next steps
	return fmt.Errorf("config execution logic not supported as of now")

	// TODO: add logic here (e.g. enable sshd systemd service, setup firewalld, ...)
	// for now, club everything in main.go, and split later once having the right structure

	// ------------
	// --- SSHD ---
	// ------------
	output = executor.NewLocalShell().Execute(ctx, &executor.Input{
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
	// -----------------
	// --- Tailscale ---
	// -----------------
	output = executor.NewLocalShell().Execute(ctx, &executor.Input{
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
	output = executor.NewLocalShell().Execute(ctx, &executor.Input{
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

func main() {
	if os.Getuid() != 0 {
		log.Fatalln("this script must be run as root/sudo")
	}

	cliInput, err := parse()
	if err != nil {
		log.Fatalf("failed to parse argument flags: %v\n", err)
	}

	// Ctrl + C or something to stop the command halfway
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err = processFlags(ctx, cliInput); err != nil {
		log.Fatalf("failed to process input: %v\n", err)
	}

	log.Println("completed")
}
