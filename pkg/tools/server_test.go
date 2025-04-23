package tools

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvironmentStrings(t *testing.T) {
	os.Clearenv()

	m, err := GetEnvironmentStrings("X")
	assert.NoError(t, err)
	assert.Empty(t, m)

	os.Setenv("X_y", "abc")
	m, err = GetEnvironmentStrings("X")
	assert.NoError(t, err)
	assert.NotEmpty(t, m)
	assert.Len(t, m, 1)

	assert.Equal(t, "abc", m["y"])
}

func TestGetEnvironmentURLS(t *testing.T) {
	os.Clearenv()

	m, err := GetEnvironmentURLS("X")
	assert.NoError(t, err)
	assert.Empty(t, m)

	os.Setenv("X_y", "%$%@6^^")
	m, err = GetEnvironmentURLS("X")
	assert.ErrorContains(t, err, "url.Parse")
	assert.Empty(t, m)

	os.Clearenv()
	os.Setenv("X_y", "http://localhost/abc")
	os.Setenv("X_z", "http://xxx.com/abc")
	m, err = GetEnvironmentURLS("X")
	assert.NoError(t, err)
	assert.NotEmpty(t, m)
	assert.Len(t, m, 2)

	assert.Equal(t, "localhost", m["y"].Hostname())
	assert.Equal(t, "xxx.com", m["z"].Hostname())
}
