package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

const DefaultConfigPath = "configs/config.yaml"

type Config struct {
	App     AppConfig     `mapstructure:"app"`
	HTTP    HTTPConfig    `mapstructure:"http"`
	Logging LoggingConfig `mapstructure:"logging"`
	Metrics MetricsConfig `mapstructure:"metrics"`
	Jira    JiraConfig    `mapstructure:"jira"`
	Teams   TeamsConfig   `mapstructure:"teams"`
	OAuth2  OAuth2Config  `mapstructure:"oauth2"`
}

type AppConfig struct {
	Name            string        `mapstructure:"name"`
	Environment     string        `mapstructure:"environment"`
	StartupTimeout  time.Duration `mapstructure:"startup_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type HTTPConfig struct {
	Address         string        `mapstructure:"address"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type LoggingConfig struct {
	Level     string `mapstructure:"level"`
	Format    string `mapstructure:"format"`
	AddSource bool   `mapstructure:"add_source"`
}

type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Path      string `mapstructure:"path"`
	Namespace string `mapstructure:"namespace"`
}

type JiraConfig struct {
	BaseURL    string `mapstructure:"base_url"`
	ProjectKey string `mapstructure:"project_key"`
}

type TeamsConfig struct {
	TenantID  string `mapstructure:"tenant_id"`
	TeamID    string `mapstructure:"team_id"`
	ChannelID string `mapstructure:"channel_id"`
}

type OAuth2Config struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	AuthURL      string   `mapstructure:"auth_url"`
	TokenURL     string   `mapstructure:"token_url"`
	Scopes       []string `mapstructure:"scopes"`
}

func Load(configPath string) (Config, error) {
	v := viper.New()
	setDefaults(v)
	v.SetEnvPrefix("TEAMS2ISSUE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if configPath == "" {
		configPath = DefaultConfigPath
	}

	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if !errors.As(err, &notFound) && !errors.Is(err, os.ErrNotExist) {
				return Config{}, fmt.Errorf("read config: %w", err)
			}
			if configPath != DefaultConfigPath {
				return Config{}, fmt.Errorf("read config: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	)); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	defaults := map[string]any{
		"app.name":              "teams2issue",
		"app.environment":       "development",
		"app.startup_timeout":   "10s",
		"app.shutdown_timeout":  "10s",
		"http.address":          "127.0.0.1:8080",
		"http.read_timeout":     "5s",
		"http.write_timeout":    "10s",
		"http.idle_timeout":     "30s",
		"http.shutdown_timeout": "10s",
		"logging.level":         "info",
		"logging.format":        "json",
		"logging.add_source":    false,
		"metrics.enabled":       true,
		"metrics.path":          "/metrics",
		"metrics.namespace":     "teams2issue",
		"jira.base_url":         "",
		"jira.project_key":      "",
		"teams.tenant_id":       "",
		"teams.team_id":         "",
		"teams.channel_id":      "",
		"oauth2.client_id":      "",
		"oauth2.client_secret":  "",
		"oauth2.auth_url":       "",
		"oauth2.token_url":      "",
		"oauth2.scopes":         []string{},
	}

	for key, value := range defaults {
		v.SetDefault(key, value)
	}
}

func (c Config) Validate() error {
	var errs []error

	if c.App.Name == "" {
		errs = append(errs, errors.New("app.name must not be empty"))
	}
	if c.HTTP.Address == "" {
		errs = append(errs, errors.New("http.address must not be empty"))
	}
	switch c.Logging.Format {
	case "json", "text":
	default:
		errs = append(errs, fmt.Errorf("logging.format must be one of json or text, got %q", c.Logging.Format))
	}
	switch c.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		errs = append(errs, fmt.Errorf("logging.level must be one of debug, info, warn or error, got %q", c.Logging.Level))
	}
	if c.Metrics.Enabled && !strings.HasPrefix(c.Metrics.Path, "/") {
		errs = append(errs, errors.New("metrics.path must start with / when metrics are enabled"))
	}
	if c.App.StartupTimeout <= 0 {
		errs = append(errs, errors.New("app.startup_timeout must be greater than zero"))
	}
	if c.App.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("app.shutdown_timeout must be greater than zero"))
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("http.shutdown_timeout must be greater than zero"))
	}

	return errors.Join(errs...)
}
