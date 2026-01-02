package config

import "fmt"

// Example: Accessing extensible fields from the Extensions map
func ExampleUsage(cfg *Config) {
	// Access nested extensible fields safely
	if features, ok := cfg.Extensions["features"].(map[string]interface{}); ok {
		if enableMetrics, ok := features["enable_metrics"].(bool); ok {
			fmt.Printf("Metrics enabled: %v\n", enableMetrics)
		}

		if cacheTTL, ok := features["cache_ttl"].(int); ok {
			fmt.Printf("Cache TTL: %d seconds\n", cacheTTL)
		}
	}

	// Access custom fields
	if custom, ok := cfg.Extensions["custom"].(map[string]interface{}); ok {
		if team, ok := custom["team"].(string); ok {
			fmt.Printf("Team: %s\n", team)
		}

		if tags, ok := custom["tags"].([]interface{}); ok {
			fmt.Printf("Tags: %v\n", tags)
		}
	}
}

// Helper function to safely get a string from Extensions
func (c *Config) GetExtensionString(keys ...string) (string, bool) {
	current := interface{}(c.Extensions)

	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key - try to get the value
			if m, ok := current.(map[string]interface{}); ok {
				if val, ok := m[key].(string); ok {
					return val, true
				}
			}
			return "", false
		}

		// Navigate deeper
		if m, ok := current.(map[string]interface{}); ok {
			current = m[key]
		} else {
			return "", false
		}
	}

	return "", false
}

// Helper function to safely get a bool from Extensions
func (c *Config) GetExtensionBool(keys ...string) (bool, bool) {
	current := interface{}(c.Extensions)

	for i, key := range keys {
		if i == len(keys)-1 {
			if m, ok := current.(map[string]interface{}); ok {
				if val, ok := m[key].(bool); ok {
					return val, true
				}
			}
			return false, false
		}

		if m, ok := current.(map[string]interface{}); ok {
			current = m[key]
		} else {
			return false, false
		}
	}

	return false, false
}

// Helper function to safely get an int from Extensions
func (c *Config) GetExtensionInt(keys ...string) (int, bool) {
	current := interface{}(c.Extensions)

	for i, key := range keys {
		if i == len(keys)-1 {
			if m, ok := current.(map[string]interface{}); ok {
				if val, ok := m[key].(int); ok {
					return val, true
				}
			}
			return 0, false
		}

		if m, ok := current.(map[string]interface{}); ok {
			current = m[key]
		} else {
			return 0, false
		}
	}

	return 0, false
}
