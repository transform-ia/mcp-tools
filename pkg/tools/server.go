package tools

import (
	"net/url"
	"os"
	"strings"

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
