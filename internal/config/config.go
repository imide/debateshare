package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig      `mapstructure:"app"`
	Server ServerConfig   `mapstructure:"server"`
	DB     DatabaseConfig `mapstructure:"database"`
	S3     S3Config       `mapstructure:"s3"`
	Log    LogConfig      `mapstructure:"logger"`
}

type AppConfig struct {
	Name        string        `mapstructure:"name"`
	Version     string        `mapstructure:"version"`
	Debug       bool          `mapstructure:"debug"`
	RoomTTL     time.Duration `mapstructure:"room_ttl"`
	MaxFileSize int64         `mapstructure:"max_file_size"`
}

type ServerConfig struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	BaseURL string `mapstructure:"base_url"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type S3Config struct {
	Endpoint           string `mapstructure:"endpoint"`
	Bucket             string `mapstructure:"bucket"`
	AWSAccessKeyID     string `mapstructure:"aws_access_key_id"`
	AWSSecretAccessKey string `mapstructure:"aws_secret_access_key"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

func expandEnvVars(v *viper.Viper) {
	re := regexp.MustCompile(`\$\{([^}]+)}`)
	for _, key := range v.AllKeys() {
		val := v.GetString(key)
		if re.MatchString(val) {
			v.Set(key, re.ReplaceAllStringFunc(val, func(match string) string {
				varName := match[2 : len(match)-1]
				envVal := os.Getenv(varName)
				if envVal == "" {
					return match
				}
				return envVal
			}))
		}
	}
}

func Load(rootDir string) (*Config, error) {
	FilePaths := []string{
		".",
		"./config",
		"./configs",
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	for _, filePath := range FilePaths {
		v.AddConfigPath(filepath.Join(rootDir, filePath))
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read base config: %w", err)
	}

	envConfigName := fmt.Sprintf("config.%s", env)
	v.SetConfigName(envConfigName)
	if err := v.MergeInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, fmt.Errorf("failed to read env config: %w", err)
		}
	}

	expandEnvVars(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// manually set max_size to be MB
	cfg.App.MaxFileSize = cfg.App.MaxFileSize << 20

	return &cfg, nil
}
