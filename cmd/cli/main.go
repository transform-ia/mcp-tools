// Package main implements a CLI for github.com/metoro-io/mcp-golang
// It reads a YAML config file and executes MCP tools through a server process
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	mcp "github.com/metoro-io/mcp-golang"
)

const minArgs = 2 // Minimum required command line arguments

// Config represents the YAML configuration file structure
type Config struct {
	Server struct {
		Exec string            `yaml:"exec"`
		Env  map[string]string `yaml:"env"`
		Args []string          `yaml:"args"`
	} `yaml:"server"`
	Tools []struct {
		Name string `yaml:"name"`
		Arg  any    `yaml:"arg"`
	} `yaml:"tools"`
}

func logic() error {
	fmt.Println("Starting MCP client...")

	configFile := os.Args[1]
	fmt.Printf("Loading config from: %s\n", configFile)

	// Validate config file path is clean
	cleanPath := filepath.Clean(configFile)
	if cleanPath != configFile {
		return errors.New("Invalid config file path")
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return errors.Wrap(err, "ReadFile")
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return errors.Wrap(err, "yaml.Unmarshal")
	}

	fmt.Println("Successfully parsed config file")

	if config.Server.Exec == "" {
		return errors.New("No server executable specified in config")
	}

	fmt.Printf("Preparing to execute server: %s\n", config.Server.Exec)
	fmt.Printf("With arguments: %v\n", config.Server.Args)
	fmt.Printf("Environment variables: %v\n", config.Server.Env)

	// Validate server executable path is clean and safe
	cleanExec := filepath.Clean(config.Server.Exec)
	if cleanExec != config.Server.Exec {
		return errors.New("Invalid server executable path")
	}

	// Validate command arguments
	for _, arg := range config.Server.Args {
		if arg == "" {
			return errors.New("Empty command argument not allowed")
		}
	}

	// Use absolute path for additional safety
	absExec, err := filepath.Abs(cleanExec)
	if err != nil {
		return errors.Wrap(err, "Could not get absolute path for executable")
	}

	//nolint:gosec
	cmd := exec.Command(absExec, config.Server.Args...)
	for k, v := range config.Server.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	fmt.Println("Configured command with environment variables")

	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.New("cmd.StdinPipe")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "cmd.StdinPipe")
	}

	fmt.Println("running: " + cmd.String())

	if err = cmd.Start(); err != nil {
		return errors.Wrap(err, "cmd.Start")
	}

	fmt.Println("Creating stdio transport...")

	transport := stdio.NewStdioServerTransportWithIO(stdout, stdin)

	fmt.Println("Creating MCP client...")

	client := mcp.NewClient(transport)

	go func() {
		fmt.Println("Starting server process monitoring...")

		if err = cmd.Wait(); err != nil {
			fmt.Printf("Server process exited with error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Server process completed successfully")
	}()

	fmt.Println("Initializing MCP client...")

	if _, err = client.Initialize(context.Background()); err != nil {
		return errors.Wrap(err, "client.Initialize")
	}

	fmt.Println("Successfully initialized MCP client")

	// Run tools from config
	fmt.Printf("Executing %d tools from config...\n", len(config.Tools))

	for _, tool := range config.Tools {
		fmt.Printf("Executing tool: %s with args: %v\n", tool.Name, tool.Arg)

		result, err := client.CallTool(context.Background(), tool.Name, tool.Arg)
		if err != nil {
			return errors.Wrapf(err, "failed to call tool %s", tool.Name)
		}

		if result != nil && len(result.Content) > 0 {
			fmt.Printf("Tool %s completed successfully. Result: %v\n", tool.Name, result.Content[0].TextContent.Text)
		} else {
			fmt.Printf("Tool %s completed with no output\n", tool.Name)
		}
	}

	return nil
}

func main() {
	if len(os.Args) < minArgs {
		fmt.Println("Usage: mcp-utils <config-file>")
		os.Exit(1)
	}

	if err := logic(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
