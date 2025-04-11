// Package tools is a collection or utilities for MCP Server's tool
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pkg/errors"
)

const keyIsJSON = "json_output"

// Tool are interface to a MCP tool implementation
type Tool interface {
	Name() string
	New() (*mcp.Tool, error)
	Exec(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// TextContentError creates a CallToolResult with an error message formatted as text content.
func TextContentError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Error: %q", err.Error()),
			},
		},
		IsError: true,
	}
}

// GetParam return a value from a MCP Tool request parameter
func GetParam[T any](request *mcp.CallToolRequest, paramName string) (*T, error) {
	untypedValue, exists := request.Params.Arguments[paramName]
	if !exists {
		return nil, errors.Errorf("missing argument %q", paramName)
	}

	typedValue, ok := untypedValue.(T)
	if !ok {
		return nil, errors.Errorf(
			"invalid value %q type for argument %q",
			untypedValue,
			paramName,
		)
	}

	return &typedValue, nil
}

// GetOptionalParam return a value from a MCP Tool request parameter
// nil if no value in request
func GetOptionalParam[T any](request *mcp.CallToolRequest, paramName string) (*T, error) {
	untypedValue, exists := request.Params.Arguments[paramName]
	if !exists {
		//nolint:nilnil
		return nil, nil
	}

	typedValue, ok := untypedValue.(T)
	if !ok {
		return nil, errors.Errorf(
			"invalid value %q type for argument %q",
			untypedValue,
			paramName,
		)
	}

	return &typedValue, nil
}

// WithOptionalJSONOutput create a Tool property to optionally return the output as JSON
func WithOptionalJSONOutput() mcp.ToolOption {
	return mcp.WithBoolean(
		keyIsJSON,
		mcp.Title("Output as json"),
		mcp.DefaultBool(false),
		mcp.Description("Output a JSON object instead of markdown"),
	)
}

// TextRenderOrJSON render a text/template.Template or a JSON string
func TextRenderOrJSON(data any, tpl *template.Template, isJSON bool) *mcp.CallToolResult {
	var (
		err error
		buf = bytes.NewBuffer(nil)
	)

	if isJSON {
		err = json.NewEncoder(buf).Encode(data)
	} else {
		err = tpl.Execute(buf, data)
	}

	if err != nil {
		return TextContentError(err)
	}

	return mcp.NewToolResultText(buf.String())
}

// ServerAddTools add to a server initialized Tool
func ServerAddTools(server *server.MCPServer, tools []Tool) error {
	for index, tool := range tools {
		toolInstance, err := tool.New()
		if err != nil {
			return errors.Wrapf(err, "tools[%d:%s].New()", index, tool.Name())
		}

		server.AddTool(*toolInstance, tool.Exec)
	}

	return nil
}
