// Package main is a simple command runner that executes a specified command
// and manages an embedded MQTT broker.
// It sets up the broker and sets environment variable MQTT_BROKER_ADDRESS,
// and then starts the command.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/grafana/xk6-mqtt/internal/broker"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		os.Exit(1)
	}

	// Set up the embedded MQTT broker
	server := broker.Setup()

	// Extract command and arguments
	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	// Create the command
	cmd := exec.Command(cmdName, cmdArgs...) //#nosec G204

	// Connect stdin, stdout, stderr to preserve interactivity
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Handle signals gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start command '%s': %v\n", cmdName, err)
		os.Exit(1)
	}

	// Handle signal forwarding to child process
	go func() {
		sig := <-sigChan

		if cmd.Process != nil {
			_ = cmd.Process.Signal(sig)
		}
	}()

	// Wait for command to complete
	err := cmd.Wait()
	if err != nil {
		var exitError *exec.ExitError

		if errors.As(err, &exitError) {
			// Command exited with non-zero status
			os.Exit(exitError.ExitCode())
		}

		fmt.Fprintf(os.Stderr, "Command execution error: %v\n", err)
		os.Exit(1)
	}

	log.Print("Shutting down embedded MQTT broker...")

	if err := server.Close(); err != nil {
		log.Fatalf("Error shutting down embedded broker: %v", err)
	}
}
