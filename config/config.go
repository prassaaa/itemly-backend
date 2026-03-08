package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort      string `mapstructure:"APP_PORT"`
	DBHost       string `mapstructure:"DB_HOST"`
	DBPort       string `mapstructure:"DB_PORT"`
	DBUser       string `mapstructure:"DB_USER"`
	DBPassword   string `mapstructure:"DB_PASSWORD"`
	DBName       string `mapstructure:"DB_NAME"`
	DBSSLMode    string `mapstructure:"DB_SSLMODE"`
	JWTSecret    string `mapstructure:"JWT_SECRET"`
	JWTExpiryHrs int    `mapstructure:"JWT_EXPIRY_HOURS"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBName, c.DBSSLMode)
	if c.DBPassword != "" {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
	}
	return dsn
}
