package tools

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnvironmentStrings(t *testing.T) {
	t.Parallel()
	os.Clearenv()

	value, err := GetEnvironmentStrings("X")
	require.ErrorContains(t, err, "No configuration found")
	assert.Empty(t, value)

	t.Setenv("X_y", "abc")

	value, err = GetEnvironmentStrings("X")
	require.NoError(t, err)
	assert.NotEmpty(t, value)
	assert.Len(t, value, 1)

	assert.Equal(t, "abc", value["y"])
}

func TestGetEnvironmentURLS(t *testing.T) {
	t.Parallel()
	os.Clearenv()

	t.Setenv("X_y", "%$%@6^^")

	value, err := GetEnvironmentURLS("X")
	require.ErrorContains(t, err, "url.Parse")
	assert.Empty(t, value)

	os.Clearenv()
	t.Setenv("X_y", "http://localhost/abc")
	t.Setenv("X_z", "http://xxx.com/abc")

	value, err = GetEnvironmentURLS("X")
	require.NoError(t, err)
	assert.NotEmpty(t, value)
	assert.Len(t, value, 2)

	assert.Equal(t, "localhost", value["y"].Hostname())
	assert.Equal(t, "xxx.com", value["z"].Hostname())
}
