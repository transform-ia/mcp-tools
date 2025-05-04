// Package main implements a CLI for github.com/mark3labs/mcp-go
// It reads a YAML config file and executes MCP tools through a server process
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
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

	// Validate server executable path is clean and safe
	cleanExec := filepath.Clean(config.Server.Exec)
	if cleanExec != config.Server.Exec {
		return errors.New("Invalid server executable path")
	}

	absExec, err := filepath.Abs(cleanExec)
	if err != nil {
		return errors.Wrap(err, "Could not get absolute path for executable")
	}

	// Convert env map to slice for NewStdioMCPClient
	var (
		index    int
		envSlice = make([]string, len(config.Server.Env))
	)

	for k, v := range config.Server.Env {
		envSlice[index] = fmt.Sprintf("%s=%s", k, v)
		index++
	}

	fmt.Printf("Creating MCP client via stdio for command: %s\n", absExec)
	fmt.Printf("With arguments: %v\n", config.Server.Args)
	fmt.Printf("Environment variables: %v\n", envSlice)

	// Use NewStdioMCPClient - this handles process launch and communication
	cli, err := client.NewStdioMCPClient(absExec, envSlice, config.Server.Args...)
	if err != nil {
		return errors.Wrap(err, "NewStdioMCPClient failed")
	}
	// No need for manual cmd, pipes, start, wait, or transport creation

	fmt.Println("Initializing MCP client...")

	// Use the new client variable 'cli' and qualify InitializeRequest with mcp package
	if _, err = cli.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		return errors.Wrap(err, "client.Initialize")
	}

	fmt.Println("Successfully initialized MCP client")

	// Run tools from config
	fmt.Printf("Executing %d tools from config...\n", len(config.Tools))

	for _, tool := range config.Tools {
		fmt.Printf("Executing tool: %s with args: %v\n", tool.Name, tool.Arg)

		// Use the new client variable 'cli' and construct CallToolRequest correctly
		// with nested Params struct and type assertion for Arguments.
		var argsMap map[string]any

		if tool.Arg != nil {
			var ok bool

			argsMap, ok = tool.Arg.(map[string]any)
			if !ok {
				return errors.Errorf("tool %s arguments are not a map[string]interface{}", tool.Name)
			}
		}

		req := mcp.CallToolRequest{}
		req.Params.Name = tool.Name
		req.Params.Arguments = argsMap

		result, err := cli.CallTool(context.Background(), req)
		if err != nil {
			return errors.Wrapf(err, "failed to call tool %s", tool.Name)
		}

		// Use type assertion via AsTextContent to get text result
		var resultText string

		if result != nil && len(result.Content) > 0 {
			contentItem := result.Content[0]
			if textContent, ok := mcp.AsTextContent(contentItem); ok {
				resultText = textContent.Text
			}
		}

		if resultText != "" {
			fmt.Printf("Tool %s completed successfully. Result: %s\n", tool.Name, resultText)
		} else {
			fmt.Printf("Tool %s completed with no text output\n", tool.Name)
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
