package config

import (
	"fmt"
	"strings"

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
	JWTSecret           string  `mapstructure:"JWT_SECRET"`
	JWTExpiryHrs        int     `mapstructure:"JWT_EXPIRY_HOURS"`
	JWTRefreshExpiryHrs int     `mapstructure:"JWT_REFRESH_EXPIRY_HOURS"`
	AppEnv              string  `mapstructure:"APP_ENV"`
	CORSAllowedOrigins  string  `mapstructure:"CORS_ALLOWED_ORIGINS"`
	RateLimitRPS        float64 `mapstructure:"RATE_LIMIT_RPS"`
	RateLimitBurst      int     `mapstructure:"RATE_LIMIT_BURST"`
	MaxBodySize         int64   `mapstructure:"MAX_BODY_SIZE"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("JWT_REFRESH_EXPIRY_HOURS", 168)
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	viper.SetDefault("RATE_LIMIT_RPS", 1)
	viper.SetDefault("RATE_LIMIT_BURST", 5)
	viper.SetDefault("MAX_BODY_SIZE", 1048576)

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

func (c *Config) CORSOriginsList() []string {
	origins := strings.Split(c.CORSAllowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return origins
}

func (c *Config) IsProduction() bool {
	return strings.ToLower(c.AppEnv) == "production"
}
