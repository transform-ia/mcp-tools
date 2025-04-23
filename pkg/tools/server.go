package tools

import (
	"maps"
	"net/url"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
)

// MustGetenvErrorFormat is a convenient error wrapper format
const (
	MustGetenvErrorFormat = "tools.MustGetenv"
	argumentConfiguration = "configuration"
)

// MustGetenv get value from a environment variable
func MustGetenv(keyName string) (*string, error) {
	value := os.Getenv(keyName)
	if len(value) == 0 {
		return nil, errors.Errorf("missing environment variable %q", keyName)
	}
	return &value, nil
}

// GetEnvironmentURLS compile a list of url made from environment
// variables that have a common prefix
func GetEnvironmentURLS(prefix string) (map[string]*url.URL, error) {
	output := make(map[string]*url.URL)

	values, err := GetEnvironmentStrings(prefix)
	if err != nil {
		return nil, errors.Wrap(err, "GetEnvironmentStrings")
	}

	for name, value := range values {
		parsed, err := url.Parse(value)
		if err != nil {
			return nil, errors.Wrapf(err, "url.Parse(%q)", value)
		}

		output[name] = parsed
	}

	return output, nil
}

// WithConfigurationOption create a tool property to select a configuration key
func WithConfigurationOption[T any](m map[string]*T) mcp.ToolOption {
	const (
		title       = "Configuration name"
		description = "Which configuration use to perform MCP server operations"
	)

	var (
		lenMap = len(m)
		keys   = make([]string, lenMap)
		index  = 0
	)

	for key := range maps.Keys(m) {
		keys[index] = key
		index++
	}

	if lenMap == 1 {
		return mcp.WithString(
			argumentConfiguration,
			mcp.Required(),
			mcp.Title(title),
			mcp.Description(description),
			mcp.DefaultString(keys[0]),
			mcp.Enum(keys...),
		)
	}

	return mcp.WithString(
		argumentConfiguration,
		mcp.Required(),
		mcp.Title(title),
		mcp.Description(description),
		mcp.Enum(keys...),
	)
}

// SelectFromConfiguration get the value of something based on MCP server request
func SelectFromConfiguration[T any](m map[string]*T, request *mcp.CallToolRequest) (*T, error) {
	config, err := GetParam[string](request, argumentConfiguration)
	if err != nil {
		return nil, errors.Wrap(err, GetParamError)
	}

	value, ok := m[*config]
	if !ok {
		return nil, errors.Errorf("invalid configuration %q", *config)
	}

	return value, nil
}

// GetEnvironmentStrings compile a list of string made from environment
// variables that have a common prefix
func GetEnvironmentStrings(prefix string) (map[string]string, error) {
	var (
		output           = make(map[string]string)
		prefixUnderscore = prefix + "_"
	)

	for _, env := range os.Environ() {
		key := strings.Split(env, "=")[0]

		if strings.HasPrefix(key, prefixUnderscore) {
			name, _ := strings.CutPrefix(key, prefixUnderscore)

			output[name] = os.Getenv(key)
		}
	}

	if len(output) == 0 {
		return nil, errors.New("No configuration found")
	}

	return output, nil
}
