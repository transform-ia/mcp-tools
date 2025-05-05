package tools

import (
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustGetenv(t *testing.T) {
	tests := []struct {
		name      string
		envKey    string
		envValue  string
		wantValue string
		wantErr   bool
	}{
		{
			name:    "missing env var",
			envKey:  "MISSING",
			wantErr: true,
		},
		{
			name:      "existing env var",
			envKey:    "EXISTING",
			envValue:  "test-value",
			wantValue: "test-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.envValue != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}

			got, err := MustGetenv(tt.envKey)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), "missing environment variable")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, *got)
		})
	}
}

func TestWithConfigurationOption(t *testing.T) {
	tests := []struct {
		name        string
		resources   map[string]*int
		wantEnum    []string
		wantDefault string
	}{
		{
			name: "single resource",
			resources: map[string]*int{
				"one": new(int),
			},
			wantEnum:    []string{"one"},
			wantDefault: "one",
		},
		{
			name: "multiple resources",
			resources: map[string]*int{
				"one": new(int),
				"two": new(int),
			},
			wantEnum: []string{"one", "two"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithConfigurationOption(tt.resources)
			assert.NotNil(t, opt)
		})
	}
}

func TestSelectFromConfiguration(t *testing.T) {
	resources := map[string]*int{
		"one": new(int),
		"two": new(int),
	}

	tests := []struct {
		name      string
		config    string
		wantValue *int
		wantErr   bool
	}{
		{
			name:      "valid config",
			config:    "one",
			wantValue: resources["one"],
		},
		{
			name:    "invalid config",
			config:  "three",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &mcp.CallToolRequest{
				Params: struct {
					Name      string         `json:"name"`
					Arguments map[string]any `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Arguments: map[string]any{
						argumentConfiguration: tt.config,
					},
				},
			}

			got, err := SelectFromConfiguration(resources, req)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, got)
		})
	}
}

func TestServe(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		wantHTTP bool
		wantErr  bool
	}{
		{
			name:     "stdio mode",
			wantHTTP: false,
		},
		{
			name:     "http mode",
			port:     "8080",
			wantHTTP: true,
		},
		{
			name:    "invalid port",
			port:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.port != "" {
				t.Setenv("PORT", tt.port)
			}

			// Create a mock server
			srv := &server.MCPServer{}

			// In real tests, we'd need to mock the server.Start() method
			// For now just test the basic behavior
			err := Serve(srv)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

// Existing tests for GetEnvironmentStrings and GetEnvironmentURLS...
