package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/terabiome/infrastructures/instructions/internal/models"
)

type Flags struct {
	YAMLConfigPath string
}

func parse() (*Flags, error) {
	flagStruct := Flags{}

	// yaml-config-path
	if flagVal := flag.String("yaml-config-path", "", "Path to YAML config path"); flagVal == nil {
		return nil, errors.New("missing yaml-config-path")
	} else if *flagVal == "" {
		return nil, errors.New("yaml-config-path is present but empty")
	} else {
		flagVal = &flagStruct.YAMLConfigPath
	}

	return &flagStruct, nil
}

func processFlags(parsedFlags *Flags) error {
	cfgStruct := models.Config{}
	if err := cfgStruct.LoadFromLocalPath(parsedFlags.YAMLConfigPath); err != nil {
		return fmt.Errorf("failed to load YAML config: %w", err)
	}

	// TODO: add logic here (e.g. enable sshd systemd service, setup firewalld, ...)

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

	if err = processFlags(parsedFlags); err != nil {
		log.Fatalf("failed to process input: %v\n", err)
	}

	log.Println("completed")
}
