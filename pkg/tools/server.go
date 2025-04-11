package tools

import (
	"os"

	"github.com/pkg/errors"
)

// MustGetenvErrorFormat is a convenient error wrapper format
const MustGetenvErrorFormat = "tools.MustGetenv"

// MustGetenv get value from a environment variable
func MustGetenv(keyName string) (*string, error) {
	value := os.Getenv(keyName)
	if len(value) == 0 {
		return nil, errors.Errorf("missing environment variable %q", keyName)
	}
	return &value, nil
}
