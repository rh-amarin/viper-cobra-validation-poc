package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/cobra-viper-demo/config"
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
		expectedConfig config.Config
	}{
		{
			name:          "Config File Only",
			args:          []string{},
			envVars:       map[string]string{},
			configContent: baseConfig,
			expectedConfig: config.Config{
				App: config.AppConfig{
					Name:        "ConfigAppName",
					Version:     "1.0.0",
					Environment: "production",
				},
				Server: config.ServerConfig{
					Host:    "localhost",
					Port:    8080,
					Timeout: 30,
				},
				Database: config.DatabaseConfig{
					Host:     "db.local",
					Port:     5432,
					Username: "user",
					Password: "password",
					Name:     "dbname",
				},
				Logging: config.LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			},
		},
		{
			name: "Environment Variable Override",
			args: []string{},
			envVars: map[string]string{
				"MYAPP_APP_NAME":    "EnvAppName",
				"MYAPP_SERVER_PORT": "8081",
			},
			configContent: baseConfig,
			expectedConfig: config.Config{
				App: config.AppConfig{
					Name:        "EnvAppName", // Env overrides config
					Version:     "1.0.0",
					Environment: "production",
				},
				Server: config.ServerConfig{
					Host:    "localhost",
					Port:    8081, // Env overrides config
					Timeout: 30,
				},
				Database: config.DatabaseConfig{
					Host:     "db.local",
					Port:     5432,
					Username: "user",
					Password: "password",
					Name:     "dbname",
				},
				Logging: config.LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			},
		},
		{
			name: "Flag Override",
			args: []string{"--app-name=FlagAppName", "--server-port=8082"},
			envVars: map[string]string{
				"MYAPP_APP_NAME": "EnvAppName", // Flag should override this
			},
			configContent: baseConfig,
			expectedConfig: config.Config{
				App: config.AppConfig{
					Name:        "FlagAppName", // Flag overrides Env & Config
					Version:     "1.0.0",
					Environment: "production",
				},
				Server: config.ServerConfig{
					Host:    "localhost",
					Port:    8082, // Flag overrides Config
					Timeout: 30,
				},
				Database: config.DatabaseConfig{
					Host:     "db.local",
					Port:     5432,
					Username: "user",
					Password: "password",
					Name:     "dbname",
				},
				Logging: config.LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			},
		},
		{
			name: "Mixed Priorities",
			args: []string{"--server-host=flag-host"},
			envVars: map[string]string{
				"MYAPP_APP_NAME": "EnvAppName",
			},
			configContent: baseConfig,
			expectedConfig: config.Config{
				App: config.AppConfig{
					Name:        "EnvAppName", // Env > Config
					Version:     "1.0.0",
					Environment: "production",
				},
				Server: config.ServerConfig{
					Host:    "flag-host", // Flag > Config
					Port:    8080,        // Config (no flag/env)
					Timeout: 30,
				},
				Database: config.DatabaseConfig{
					Host:     "db.local",
					Port:     5432,
					Username: "user",
					Password: "password",
					Name:     "dbname",
				},
				Logging: config.LoggingConfig{
					Level:  "info",
					Format: "json",
				},
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

			// Extract JSON from output (skip any messages before the JSON)
			// Find the first '{' which marks the start of JSON
			jsonStart := strings.Index(output, "{")
			if jsonStart == -1 {
				t.Fatalf("No JSON found in output:\n%s", output)
			}
			jsonOutput := output[jsonStart:]

			// Parse JSON output
			var actualConfig config.Config
			if err := json.Unmarshal([]byte(jsonOutput), &actualConfig); err != nil {
				t.Fatalf("Failed to parse JSON output: %v\nJSON part:\n%s", err, jsonOutput)
			}

			// Assertions - verify key fields match expected values
			if actualConfig.App.Name != tt.expectedConfig.App.Name {
				t.Errorf("Expected App.Name=%s, got %s", tt.expectedConfig.App.Name, actualConfig.App.Name)
			}
			if actualConfig.Server.Port != tt.expectedConfig.Server.Port {
				t.Errorf("Expected Server.Port=%d, got %d", tt.expectedConfig.Server.Port, actualConfig.Server.Port)
			}
			if actualConfig.Server.Host != tt.expectedConfig.Server.Host {
				t.Errorf("Expected Server.Host=%s, got %s", tt.expectedConfig.Server.Host, actualConfig.Server.Host)
			}
			// Verify other fields for completeness
			if actualConfig.App.Version != tt.expectedConfig.App.Version {
				t.Errorf("Expected App.Version=%s, got %s", tt.expectedConfig.App.Version, actualConfig.App.Version)
			}
			if actualConfig.App.Environment != tt.expectedConfig.App.Environment {
				t.Errorf("Expected App.Environment=%s, got %s", tt.expectedConfig.App.Environment, actualConfig.App.Environment)
			}
		})
	}
}
