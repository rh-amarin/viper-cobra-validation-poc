package config

type Config struct {
	App      AppConfig      `mapstructure:"app" json:"app" validate:"required"`
	Server   ServerConfig   `mapstructure:"server" json:"server"`
	Database DatabaseConfig `mapstructure:"database" json:"database"`
	Logging  LoggingConfig  `mapstructure:"logging" json:"logging"`
}

type AppConfig struct {
	Name        string `mapstructure:"name" json:"name" validate:"required"`
	Version     string `mapstructure:"version" json:"version"`
	Environment string `mapstructure:"environment" json:"environment" validate:"omitempty,oneof=development staging production"`
}

type ServerConfig struct {
	Host    string `mapstructure:"host" json:"host"`
	Port    int    `mapstructure:"port" json:"port" validate:"gte=1024,lte=9000"`
	Timeout int    `mapstructure:"timeout" json:"timeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Username string `mapstructure:"username" json:"username"`
	Password string `mapstructure:"password" json:"password"`
	Name     string `mapstructure:"name" json:"name"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level" json:"level"`
	Format string `mapstructure:"format" json:"format"`
}
