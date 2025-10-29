package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type Config struct {
	Redis RedisConfig `mapstructure:"redis"`
	API   APIConfig   `mapstructure:"api"`
}

type RedisConfig struct {
	Host     string    `mapstructure:"host"`
	Port     int       `mapstructure:"port"`
	Password string    `mapstructure:"password"`
	DB       int       `mapstructure:"db"`
	Username string    `mapstructure:"username"`
	UseTLS   bool      `mapstructure:"use_tls"`
	TLS      TLSConfig `mapstructure:"tls"`
}

type TLSConfig struct {
	CertFile           string `mapstructure:"cert_file"`
	KeyFile            string `mapstructure:"key_file"`
	CAFile             string `mapstructure:"ca_file"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
}

type APIConfig struct {
	Port int `mapstructure:"port"`
}

// LoadConfig loads configuration from config file and environment variables
// Environment variables take precedence over config file values
func LoadConfig() (*Config, error) {
	// Set defaults
	setDefaults()

	// Set config file name and paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.taskboard")

	// Read config file if it exists (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; ignore error and use defaults/env vars
	}

	// Enable environment variable support
	viper.SetEnvPrefix("TASKBOARD")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.username", "")
	viper.SetDefault("redis.use_tls", false)
	viper.SetDefault("redis.tls.cert_file", "")
	viper.SetDefault("redis.tls.key_file", "")
	viper.SetDefault("redis.tls.ca_file", "")
	viper.SetDefault("redis.tls.insecure_skip_verify", false)

	// API defaults
	viper.SetDefault("api.port", 1337)
}

// ToRedisOptions converts RedisConfig to redis.Options
func (r *RedisConfig) ToRedisOptions() (*redis.Options, error) {
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", r.Host, r.Port),
		Password: r.Password,
		DB:       r.DB,
		Username: r.Username,
	}

	if r.UseTLS {
		tlsConfig, err := r.buildTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		opts.TLSConfig = tlsConfig
	}

	return opts, nil
}

// buildTLSConfig creates a tls.Config from TLSConfig
func (r *RedisConfig) buildTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: r.TLS.InsecureSkipVerify,
	}

	// Load client cert and key if provided
	if r.TLS.CertFile != "" && r.TLS.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(r.TLS.CertFile, r.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA cert if provided
	if r.TLS.CAFile != "" {
		caCert, err := os.ReadFile(r.TLS.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}
