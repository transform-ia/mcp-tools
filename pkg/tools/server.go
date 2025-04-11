// Package tools is a collection or utilities for MCP Server's tool
package tools

import (
	"os"

	"github.com/pkg/errors"
)

// MustGetenv get value from a environment variable
func MustGetenv(keyName string) (*string, error) {
	value := os.Getenv(keyName)
	if len(value) == 0 {
		return nil, errors.Errorf("missing environment variable %q", keyName)
	}
	return &value, nil
}