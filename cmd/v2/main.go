package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/terabiome/terabiome/internal/singletons"
)

type cliInput struct {
}

func NewCLIInput() *cliInput {
	return &cliInput{}
}

func (i *cliInput) Parse() error {
	// TODO: set flag.*
	// parse and check
	flag.Parse()

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

	// injectable dependencies live here
	singletons.Init()

	// Ctrl + C or something to stop the command halfway
	_, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// TODO: add subcommands here using flag.NewFlagSet
	// https://gobyexample.com/command-line-subcommands
	// HTTP server
	// Node NATS consumer

	log.Println("completed")
}
