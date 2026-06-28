// Package config manages GitSum's user-level configuration, such as the
// Anthropic API key. It's stored outside the project directory so it
// applies no matter which repo gitsum is run from.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const envFileName = ".env"

// Dir returns the directory GitSum stores its config in
// (e.g. %AppData%\gitsum on Windows, ~/.config/gitsum on Linux/macOS).
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving user config dir: %w", err)
	}
	return filepath.Join(base, "gitsum"), nil
}

// Path returns the full path to GitSum's stored env file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, envFileName), nil
}

// SetKey writes ANTHROPIC_API_KEY=<key> to the config file, creating the
// config directory if necessary, and returns the path written to. The file
// is restricted to the current user since it holds a secret.
func SetKey(key string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("creating config dir: %w", err)
	}

	path := filepath.Join(dir, envFileName)
	content := fmt.Sprintf("ANTHROPIC_API_KEY=%s\n", key)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("writing config file: %w", err)
	}

	return path, nil
}

// Load reads the stored config file, if any, and applies its values to the
// current process environment — but only for variables that aren't already
// set, so an explicit `export ANTHROPIC_API_KEY=...` always wins.
func Load() error {
	path, err := Path()
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading config file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
