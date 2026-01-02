# Cobra + Viper Configuration Demo

This Go application demonstrates how to use [Cobra](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper) libraries to load configuration from multiple sources with proper precedence.

## Features

- **YAML Configuration File**: Load default values from `config.yaml`
- **Environment Variables**: Override config values using environment variables with `MYAPP_` prefix
- **Command-Line Flags**: Override any value using CLI flags (highest priority)
- **Configuration Precedence**: Flags > Environment Variables > Config File

## Prerequisites

- Go 1.25 or higher

## Installation

```bash
# Clone or navigate to the project directory
cd cobra-viper-demo

# Install dependencies (already done if you have go.mod)
go mod download

# Build the application
go build -o cobra-viper-demo
```

## Usage Examples

### 1. Using Config File Only

Run the application with default configuration from `config.yaml`:

```bash
go run main.go
```

Expected output shows values from `config.yaml`:
```
Using config file: ./config.yaml

=== Configuration Values ===

Application:
  Name:        MyApp
  Version:     1.0.0
  Environment: development
...
```

### 2. Using Environment Variables

Override specific values using environment variables (prefix: `MYAPP_`):

```bash
# Set environment variables (use underscore instead of dots)
export MYAPP_APP_NAME="EnvApp"
export MYAPP_SERVER_PORT=9000
export MYAPP_DATABASE_HOST="env-db-host"

go run main.go
```

The output will show:
- `app.name` = "EnvApp" (from env var)
- `server.port` = 9000 (from env var)
- `database.host` = "env-db-host" (from env var)
- Other values from config.yaml

### 3. Using Command-Line Flags

Override values using flags (highest priority):

```bash
go run main.go --app-name "FlagApp" --server-port 7000 -e production
```

### 4. Combining All Three Sources

Demonstrate the precedence order:

```bash
# Set environment variable
export MYAPP_SERVER_PORT=9000

# Run with flag (flag wins over env var and config file)
go run main.go --server-port 7000 --app-name "FlagApp"
```

Result:
- `server.port` = 7000 (from flag - highest priority)
- `app.name` = "FlagApp" (from flag)
- Other values from env vars or config file

### 5. Using a Custom Config File

```bash
go run main.go --config /path/to/custom-config.yaml
```

## Available Flags

### Application Flags
- `--app-name`, `-n`: Application name
- `--app-version`, `-v`: Application version
- `--app-environment`, `-e`: Application environment

### Server Flags
- `--server-host`: Server host
- `--server-port`, `-p`: Server port
- `--server-timeout`, `-t`: Server timeout in seconds

### Database Flags
- `--db-host`: Database host
- `--db-port`: Database port
- `--db-username`, `-u`: Database username
- `--db-password`: Database password
- `--db-name`, `-d`: Database name

### Logging Flags
- `--log-level`, `-l`: Logging level
- `--log-format`, `-f`: Logging format

## Environment Variable Mapping

Environment variables use the `MYAPP_` prefix and replace dots with underscores:

| Config Key | Environment Variable |
|------------|---------------------|
| app.name | MYAPP_APP_NAME |
| app.version | MYAPP_APP_VERSION |
| server.port | MYAPP_SERVER_PORT |
| database.host | MYAPP_DATABASE_HOST |
| logging.level | MYAPP_LOGGING_LEVEL |

## Configuration Precedence

1. **Command-line flags** (highest priority)
2. **Environment variables** (medium priority)
3. **Configuration file** (lowest priority)

## Project Structure

```
.
├── cmd/
│   └── root.go          # Cobra command and Viper configuration
├── config.yaml          # Default configuration file
├── main.go              # Application entry point
├── go.mod               # Go module file
└── README.md            # This file
```

## Testing the Precedence

Run this comprehensive test:

```bash
# Clean environment
unset MYAPP_APP_NAME
unset MYAPP_SERVER_PORT

# Test 1: Config file only
echo "Test 1: Config file only"
go run main.go | grep "Name:"

# Test 2: Environment variable overrides config
export MYAPP_APP_NAME="EnvOverride"
echo -e "\nTest 2: Environment variable overrides config"
go run main.go | grep "Name:"

# Test 3: Flag overrides both
echo -e "\nTest 3: Flag overrides environment and config"
go run main.go --app-name "FlagOverride" | grep "Name:"
```

## License

MIT
