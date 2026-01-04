package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/example/cobra-viper-demo/config"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	v       *viper.Viper
)

var rootCmd = &cobra.Command{
	Use:   "cobra-viper-demo",
	Short: "A demo application showcasing Cobra and Viper configuration",
	Long: `This application demonstrates how to load configuration values from multiple sources:
1. Command-line flags (highest priority)
2. Environment variables (medium priority)
3. Configuration file (lowest priority)`,
	Run: func(cmd *cobra.Command, args []string) {
		displayConfiguration()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// bindStringFlag defines a string flag and binds it to viper in one call
func bindStringFlag(cmd *cobra.Command, viperKey, flagName, shorthand, defaultVal, usage string) {
	cmd.Flags().StringP(flagName, shorthand, defaultVal, usage)
	if err := v.BindPFlag(viperKey, cmd.Flags().Lookup(flagName)); err != nil {
		panic(fmt.Sprintf("failed to bind flag %s to %s: %v", flagName, viperKey, err))
	}
}

// bindIntFlag defines an int flag and binds it to viper in one call
func bindIntFlag(cmd *cobra.Command, viperKey, flagName, shorthand string, defaultVal int, usage string) {
	cmd.Flags().IntP(flagName, shorthand, defaultVal, usage)
	if err := v.BindPFlag(viperKey, cmd.Flags().Lookup(flagName)); err != nil {
		panic(fmt.Sprintf("failed to bind flag %s to %s: %v", flagName, viperKey, err))
	}
}

// bindBoolFlag defines a bool flag and binds it to viper in one call
func bindBoolFlag(cmd *cobra.Command, viperKey, flagName, shorthand string, defaultVal bool, usage string) {
	cmd.Flags().BoolP(flagName, shorthand, defaultVal, usage)
	if err := v.BindPFlag(viperKey, cmd.Flags().Lookup(flagName)); err != nil {
		panic(fmt.Sprintf("failed to bind flag %s to %s: %v", flagName, viperKey, err))
	}
}

func init() {
	v = viper.New()
	cobra.OnInitialize(initConfig)

	// Config file flag (not bound to viper, handled separately)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// Application flags
	bindStringFlag(rootCmd, "app.name", "app-name", "n", "", "Application name")
	bindStringFlag(rootCmd, "app.version", "app-version", "v", "", "Application version")
	bindStringFlag(rootCmd, "app.environment", "app-environment", "e", "", "Application environment")

	// Server flags
	bindStringFlag(rootCmd, "server.host", "server-host", "", "", "Server host")
	bindIntFlag(rootCmd, "server.port", "server-port", "p", 0, "Server port")
	bindIntFlag(rootCmd, "server.timeout", "server-timeout", "t", 0, "Server timeout in seconds")

	// Database flags
	bindStringFlag(rootCmd, "database.host", "db-host", "", "", "Database host")
	bindIntFlag(rootCmd, "database.port", "db-port", "", 0, "Database port")
	bindStringFlag(rootCmd, "database.username", "db-username", "u", "", "Database username")
	bindStringFlag(rootCmd, "database.password", "db-password", "", "", "Database password")
	bindStringFlag(rootCmd, "database.name", "db-name", "d", "", "Database name")

	// Logging flags
	bindStringFlag(rootCmd, "logging.level", "log-level", "l", "", "Logging level")
	bindStringFlag(rootCmd, "logging.format", "log-format", "f", "", "Logging format")
}

func initConfig() {
	// Check for config file in order of precedence:
	// 1. --config flag (highest priority)
	// 2. MYAPP_CONFIG environment variable
	// 3. Default search paths
	if cfgFile != "" {
		// Use config file from the flag
		v.SetConfigFile(cfgFile)
	} else if envConfigFile := os.Getenv("MYAPP_CONFIG"); envConfigFile != "" {
		// Use config file from environment variable
		v.SetConfigFile(envConfigFile)
	} else {
		// Search for config in the current directory with name "config" (without extension)
		v.AddConfigPath(".")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Enable environment variable support
	v.SetEnvPrefix("MYAPP") // will be uppercased automatically
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read the configuration file
	if err := v.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %s\n\n", v.ConfigFileUsed())
	} else {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No config file found, using flags and environment variables only")
		} else {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n\n", err)
		}
	}
}

// loadAndValidateConfig loads configuration from viper and validates it
func loadAndValidateConfig() (*config.Config, error) {
	// Unmarshal the configuration into the struct
	var cfg config.Config
	if err := v.UnmarshalExact(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate the configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validateConfig validates the configuration struct and returns detailed error messages
func validateConfig(cfg *config.Config) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fmt.Fprintln(os.Stderr, "Configuration validation failed:")
			for _, fieldErr := range validationErrors {
				// Use Namespace to show the full path (e.g., "Config.Server.Port" instead of just "Port")
				fieldPath := fieldErr.Namespace()
				tag := fieldErr.Tag()
				currentValue := fieldErr.Value()
				param := fieldErr.Param()

				fmt.Fprintf(os.Stderr, "  - Field '%s' validation failed\n", fieldPath)
				fmt.Fprintf(os.Stderr, "    Current value: %v (type: %T)\n", currentValue, currentValue)

				// Provide detailed error messages based on validation tag
				switch tag {
				case "required":
					fmt.Fprintln(os.Stderr, "    Expected: non-empty value")
					if fieldErr.Field() == "Name" {
						fmt.Fprintln(os.Stderr, "    Hint: Application name is mandatory. Provide it via:")
						fmt.Fprintln(os.Stderr, "      • Flag: --app-name or -n")
						fmt.Fprintln(os.Stderr, "      • Environment variable: MYAPP_APP_NAME")
						fmt.Fprintln(os.Stderr, "      • Config file: app.name")
					}

				case "min":
					fmt.Fprintf(os.Stderr, "    Expected: minimum value of %s\n", param)

				case "max":
					fmt.Fprintf(os.Stderr, "    Expected: maximum value of %s\n", param)

				case "lte":
					fmt.Fprintf(os.Stderr, "    Expected: value less than or equal to %s\n", param)

				case "gte":
					fmt.Fprintf(os.Stderr, "    Expected: value greater than or equal to %s\n", param)

				case "lt":
					fmt.Fprintf(os.Stderr, "    Expected: value less than %s\n", param)

				case "gt":
					fmt.Fprintf(os.Stderr, "    Expected: value greater than %s\n", param)

				case "oneof":
					fmt.Fprintf(os.Stderr, "    Expected: one of [%s]\n", param)

				case "email":
					fmt.Fprintln(os.Stderr, "    Expected: valid email address format")

				case "url":
					fmt.Fprintln(os.Stderr, "    Expected: valid URL format")

				case "len":
					fmt.Fprintf(os.Stderr, "    Expected: length of %s\n", param)

				case "eq":
					fmt.Fprintf(os.Stderr, "    Expected: value equal to %s\n", param)

				case "ne":
					fmt.Fprintf(os.Stderr, "    Expected: value not equal to %s\n", param)

				default:
					fmt.Fprintf(os.Stderr, "    Validation rule: %s", tag)
					if param != "" {
						fmt.Fprintf(os.Stderr, " (parameter: %s)", param)
					}
					fmt.Fprintln(os.Stderr)
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
		}
		return err
	}

	return nil
}

// displayConfiguration loads, validates, and displays the configuration as JSON
func displayConfiguration() {
	cfg, err := loadAndValidateConfig()
	if err != nil {
		os.Exit(1)
	}

	// Marshal configuration to JSON with indentation
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling config to JSON: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}
