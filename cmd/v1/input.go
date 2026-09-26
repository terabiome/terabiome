package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/terabiome/infrastructures/internal/models"
)

type cliInput struct {
	yamlConfigPath string
	bareCommand    models.Command
}

func NewCLIInput() *cliInput {
	return &cliInput{}
}

func (i *cliInput) Parse() error {
	var tempAnchor struct {
		yamlConfigPath *string
		bareCommand    *string
	}
	tempAnchor.bareCommand = flag.String("bare-command", "", "Only execute command and exit")
	tempAnchor.yamlConfigPath = flag.String("yaml-config-path", "", "Path to YAML config path")

	// parse and check
	flag.Parse()

	if tempAnchor.bareCommand != nil {
		if flagVal := *tempAnchor.bareCommand; flagVal != "" {
			command, parseErr := models.NewCommandFromString(flagVal)
			if parseErr != nil {
				return fmt.Errorf("failed to parse bare-command: %w", parseErr)
			}
			i.bareCommand = *command
			// short-circuit immediately
			return nil
		}
	}

	if tempAnchor.yamlConfigPath == nil {
		return errors.New("missing yaml-config-path")
	} else if flagVal := *tempAnchor.yamlConfigPath; flagVal == "" {
		return errors.New("yaml-config-path is present but empty")
	} else {
		i.yamlConfigPath = flagVal
	}

	return nil
}
