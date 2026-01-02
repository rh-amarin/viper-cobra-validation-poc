package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/example/cobra-viper-demo/config"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

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

func init() {
	cobra.OnInitialize(initConfig)

	// Define flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// Application flags
	rootCmd.Flags().StringP("app-name", "n", "", "Application name")
	rootCmd.Flags().StringP("app-version", "v", "", "Application version")
	rootCmd.Flags().StringP("app-environment", "e", "", "Application environment")

	// Server flags
	rootCmd.Flags().StringP("server-host", "", "", "Server host")
	rootCmd.Flags().IntP("server-port", "p", 0, "Server port")
	rootCmd.Flags().IntP("server-timeout", "t", 0, "Server timeout in seconds")

	// Database flags
	rootCmd.Flags().StringP("db-host", "", "", "Database host")
	rootCmd.Flags().IntP("db-port", "", 0, "Database port")
	rootCmd.Flags().StringP("db-username", "u", "", "Database username")
	rootCmd.Flags().StringP("db-password", "", "", "Database password")
	rootCmd.Flags().StringP("db-name", "d", "", "Database name")

	// Logging flags
	rootCmd.Flags().StringP("log-level", "l", "", "Logging level")
	rootCmd.Flags().StringP("log-format", "f", "", "Logging format")

	// Bind flags to viper
	viper.BindPFlag("app.name", rootCmd.Flags().Lookup("app-name"))
	viper.BindPFlag("app.version", rootCmd.Flags().Lookup("app-version"))
	viper.BindPFlag("app.environment", rootCmd.Flags().Lookup("app-environment"))
	viper.BindPFlag("server.host", rootCmd.Flags().Lookup("server-host"))
	viper.BindPFlag("server.port", rootCmd.Flags().Lookup("server-port"))
	viper.BindPFlag("server.timeout", rootCmd.Flags().Lookup("server-timeout"))
	viper.BindPFlag("database.host", rootCmd.Flags().Lookup("db-host"))
	viper.BindPFlag("database.port", rootCmd.Flags().Lookup("db-port"))
	viper.BindPFlag("database.username", rootCmd.Flags().Lookup("db-username"))
	viper.BindPFlag("database.password", rootCmd.Flags().Lookup("db-password"))
	viper.BindPFlag("database.name", rootCmd.Flags().Lookup("db-name"))
	viper.BindPFlag("logging.level", rootCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("logging.format", rootCmd.Flags().Lookup("log-format"))
}

func initConfig() {
	// Check for config file in order of precedence:
	// 1. --config flag (highest priority)
	// 2. MYAPP_CONFIG environment variable
	// 3. Default search paths
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else if envConfigFile := os.Getenv("MYAPP_CONFIG"); envConfigFile != "" {
		// Use config file from environment variable
		viper.SetConfigFile(envConfigFile)
	} else {
		// Search for config in the current directory with name "config" (without extension)
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Enable environment variable support
	viper.SetEnvPrefix("MYAPP") // will be uppercased automatically
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read the configuration file
	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %s\n\n", viper.ConfigFileUsed())
	} else {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No config file found, using flags and environment variables only\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n\n", err)
		}
	}
}

func displayConfiguration() {
	// Unmarshal the configuration into the struct
	var cfg config.Config
	if err := viper.UnmarshalExact(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshaling config: %v\n", err)
		return
	}

	// Validate the configuration
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Configuration validation failed:")
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
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
		os.Exit(1)
	}

	fmt.Println("=== Configuration Values ===")
	fmt.Println()

	fmt.Println("Application:")
	fmt.Printf("  Name:        %s\n", cfg.App.Name)
	fmt.Printf("  Version:     %s\n", cfg.App.Version)
	fmt.Printf("  Environment: %s\n", cfg.App.Environment)
	fmt.Println()

	fmt.Println("Server:")
	fmt.Printf("  Host:    %s\n", cfg.Server.Host)
	fmt.Printf("  Port:    %d\n", cfg.Server.Port)
	fmt.Printf("  Timeout: %d seconds\n", cfg.Server.Timeout)
	fmt.Println()

	fmt.Println("Database:")
	fmt.Printf("  Host:     %s\n", cfg.Database.Host)
	fmt.Printf("  Port:     %d\n", cfg.Database.Port)
	fmt.Printf("  Username: %s\n", cfg.Database.Username)
	fmt.Printf("  Password: %s\n", cfg.Database.Password)
	fmt.Printf("  Name:     %s\n", cfg.Database.Name)
	fmt.Println()

	fmt.Println("Logging:")
	fmt.Printf("  Level:  %s\n", cfg.Logging.Level)
	fmt.Printf("  Format: %s\n", cfg.Logging.Format)
	fmt.Println()

	// Display extensible fields

	fmt.Println("=== Configuration Source Priority ===")
	fmt.Println("1. Command-line flags (highest)")
	fmt.Println("2. Environment variables (MYAPP_* prefix)")
	fmt.Println("3. Configuration file (lowest)")
}
