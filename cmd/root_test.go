package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestConfigurationOverrides(t *testing.T) {
	// Base configuration content
	baseConfig := `
app:
  name: "ConfigAppName"
  version: "1.0.0"
  environment: "production"
server:
  host: "localhost"
  port: 8080
  timeout: 30
database:
  host: "db.local"
  port: 5432
  username: "user"
  password: "password"
  name: "dbname"
logging:
  level: "info"
  format: "json"
`

	tests := []struct {
		name           string
		args           []string
		envVars        map[string]string
		configContent  string
		expectedValues map[string]string
	}{
		{
			name:          "Config File Only",
			args:          []string{},
			envVars:       map[string]string{},
			configContent: baseConfig,
			expectedValues: map[string]string{
				"Name:":    "ConfigAppName",
				"Port:":    "8080",
				"Host:":    "localhost",
			},
		},
		{
			name:          "Environment Variable Override",
			args:          []string{},
			envVars:       map[string]string{
				"MYAPP_APP_NAME":    "EnvAppName",
				"MYAPP_SERVER_PORT": "8081",
			},
			configContent: baseConfig,
			expectedValues: map[string]string{
				"Name:":    "EnvAppName",  // Env overrides config
				"Port:":    "8081",        // Env overrides config
				"Host:":    "localhost",   // From config (no env)
			},
		},
		{
			name:          "Flag Override",
			args:          []string{"--app-name=FlagAppName", "--server-port=8082"},
			envVars:       map[string]string{
				"MYAPP_APP_NAME": "EnvAppName", // Flag should override this
			},
			configContent: baseConfig,
			expectedValues: map[string]string{
				"Name:":    "FlagAppName", // Flag overrides Env & Config
				"Port:":    "8082",        // Flag overrides Config
				"Host:":    "localhost",   // From config
			},
		},
		{
			name:          "Mixed Priorities",
			args:          []string{"--server-host=flag-host"},
			envVars:       map[string]string{
				"MYAPP_APP_NAME": "EnvAppName",
			},
			configContent: baseConfig,
			expectedValues: map[string]string{
				"Name:":    "EnvAppName",  // Env > Config
				"Host:":    "flag-host",   // Flag > Config
				"Port:":    "8080",        // Config (no flag/env)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Config File
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}

			// Setup Environment
			os.Clearenv() // Be careful with this if running in parallel or needed system envs
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer os.Clearenv()

			// Reset Flags
			rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Changed {
					f.Value.Set(f.DefValue)
					f.Changed = false
				}
			})
			// Also reset the cfgFile variable
			cfgFile = ""

			// Setup Args
			// Prepend --config to point to our temp file
			args := append([]string{"--config", configPath}, tt.args...)
			rootCmd.SetArgs(args)

			// Capture Stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run Command
			// We cannot call rootCmd.Execute() directly easily because it might call os.Exit on error.
			// But our tests are happy paths.
			// However, Execute() captures errors and prints to stderr usually.
			// We can call rootCmd.ExecuteC() or just Execute()
			
			// Note: init() has already run. initConfig is registered.
			
			err = rootCmd.Execute()
			
			// Close and Restore Stdout
			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			// Read Output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Assertions
			for key, val := range tt.expectedValues {
				if !containsConfigValue(output, key, val) {
					t.Errorf("Expected configuration to contain '%s %s', but it didn't.\nOutput:\n%s", key, val, output)
				}
			}
		})
	}
}

// containsConfigValue checks if the output contains a line roughly matching "Key: Value"
// This is a simple check tailored to the displayConfiguration output format
func containsConfigValue(output, key, value string) bool {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key) {
			if strings.Contains(trimmed, value) {
				return true
			}
		}
	}
	return false
}
