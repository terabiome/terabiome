package main

import (
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
)

type Flags struct {
	YAMLConfigPath string
	bareCommand    struct {
		executable string
		arguments  []string
	}
}

func parse() (*Flags, error) {
	var tempAnchor struct {
		yamlConfigPath *string
		bareCommand    *string
	}
	tempAnchor.bareCommand = flag.String("bare-command", "", "Only execute command and exit")
	tempAnchor.yamlConfigPath = flag.String("yaml-config-path", "", "Path to YAML config path")

	// parse and check
	flag.Parse()
	flagStruct := Flags{}

	if tempAnchor.bareCommand != nil {
		if flagVal := *tempAnchor.bareCommand; flagVal != "" {
			segments := strings.Split(flagVal, " ")
			if len(segments) == 0 {
				return nil, errors.New("no segment in -bare-command")
			}

			flagStruct.bareCommand = struct {
				executable string
				arguments  []string
			}{
				executable: segments[0],
				arguments:  segments[1:],
			}
			// short-circuit immediately
			return &flagStruct, nil
		}
	}

	if tempAnchor.yamlConfigPath == nil {
		return nil, errors.New("missing yaml-config-path")
	} else if flagVal := *tempAnchor.yamlConfigPath; flagVal == "" {
		return nil, errors.New("yaml-config-path is present but empty")
	} else {
		flagStruct.YAMLConfigPath = flagVal
	}

	return &flagStruct, nil
}

func processFlags(ctx context.Context, parsedFlags *Flags) error {
	var (
		output executor.Output
		err    error
	)

	if parsedFlags.bareCommand.executable != "" {
		output = executor.NewLocalShell().Execute(ctx, &executor.Input{
			Mode:       executor.ModeSync,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Executable: parsedFlags.bareCommand.executable,
			Arguments:  parsedFlags.bareCommand.arguments,
		})
		output.Wait()
		if output.Error != nil {
			log.Fatalf("failed to execute bare command: %v", output.Error)
		}
		return nil
	}

	_, err = models.NewLoader().LoadConfigFromLocalPath(parsedFlags.YAMLConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load YAML config: %w", err)
	}

	// TODO: add logic here (e.g. enable sshd systemd service, setup firewalld, ...)
	// for now, club everything in main.go, and split later once having the right structure

	// ------------
	// --- SSHD ---
	// ------------
	input := executor.Input{
		Mode:       executor.ModeSync,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Executable: "systemctl",
		Arguments: []string{
			"enable", "--now",
			"sshd",
		},
	}
	output = executor.NewLocalShell().Execute(ctx, &input)
	if output.Done() && output.Error != nil {
		return fmt.Errorf("failed to start sshd service: %w", output.Error)
	}

	return nil
}

func main() {
	if os.Getuid() != 0 {
		log.Fatalln("this script must be run as root/sudo")
	}

	parsedFlags, err := parse()
	if err != nil {
		log.Fatalf("failed to parse argument flags: %v\n", err)
	}

	// Ctrl + C or something to stop the command halfway
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err = processFlags(ctx, parsedFlags); err != nil {
		log.Fatalf("failed to process input: %v\n", err)
	}

	log.Println("completed")
}
