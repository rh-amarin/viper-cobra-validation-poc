# Configuration with Extensible Fields

This package demonstrates how to unmarshal Viper configuration into a strongly-typed struct while maintaining extensibility for dynamic fields.

## Key Pattern

The `Config` struct uses the `mapstructure:",remain"` tag on the `Extensions` field to capture any configuration fields that don't match the predefined struct fields:

```go
type Config struct {
    App      AppConfig      `mapstructure:"app"`
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Logging  LoggingConfig  `mapstructure:"logging"`

    // Captures any additional configuration not defined above
    Extensions map[string]interface{} `mapstructure:",remain"`
}
```

## Usage

### Unmarshaling Config

```go
var cfg config.Config
if err := viper.Unmarshal(&cfg); err != nil {
    // handle error
}
```

### Accessing Structured Fields

```go
fmt.Printf("App Name: %s\n", cfg.App.Name)
fmt.Printf("Server Port: %d\n", cfg.Server.Port)
```

### Accessing Dynamic/Extensible Fields

#### Option 1: Direct map access with type assertions

```go
if features, ok := cfg.Extensions["features"].(map[string]interface{}); ok {
    if enableMetrics, ok := features["enable_metrics"].(bool); ok {
        fmt.Printf("Metrics enabled: %v\n", enableMetrics)
    }
}
```

#### Option 2: Using helper methods (recommended)

```go
if team, ok := cfg.GetExtensionString("custom", "team"); ok {
    fmt.Printf("Team: %s\n", team)
}

if enableMetrics, ok := cfg.GetExtensionBool("features", "enable_metrics"); ok {
    fmt.Printf("Metrics enabled: %v\n", enableMetrics)
}

if cacheTTL, ok := cfg.GetExtensionInt("features", "cache_ttl"); ok {
    fmt.Printf("Cache TTL: %d seconds\n", cacheTTL)
}
```

## Example Config File

```yaml
# Structured fields
app:
  name: "myapp"
  version: "1.0.0"

server:
  host: "localhost"
  port: 8080

# Extensible fields - automatically captured in Extensions map
features:
  enable_metrics: true
  cache_ttl: 3600

custom:
  team: "platform"
  region: "us-west-2"
```

## Benefits

1. **Type Safety**: Strongly-typed fields for known configuration
2. **Extensibility**: Dynamic fields for plugin configs, feature flags, or custom data
3. **Validation**: Can validate structured fields while allowing arbitrary extensions
4. **Migration**: Easy to promote frequently-used extensions to first-class fields
