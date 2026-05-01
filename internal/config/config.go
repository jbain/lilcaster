package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Scenarios []Scenario `yaml:"scenarios"`
}

type Scenario struct {
	Name    string        `yaml:"name"`
	Sources []Endpoint    `yaml:"sources"`
	Sinks   []Endpoint    `yaml:"sinks"`
	Filters []FilterEntry `yaml:"filters"`
	Loop    int           `yaml:"loop"`
}

// Endpoint is the shared shape for sources and sinks.
type Endpoint struct {
	Path string   `yaml:"path"`
	Args []string `yaml:"args"`
}

// Load reads and parses the config file using the 4-location precedence:
// 1. cfgFile arg  2. $LILCASTER_CONFIG  3. ./lilcaster.yml  4. ~/.config/lilcaster/lilcaster.yml
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else if env := os.Getenv("LILCASTER_CONFIG"); env != "" {
		v.SetConfigFile(env)
	} else {
		v.SetConfigName("lilcaster")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath(filepath.Join(xdgConfigHome(), "lilcaster"))
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found (searched: $LILCASTER_CONFIG, ./lilcaster.yml, %s)",
				filepath.Join(xdgConfigHome(), "lilcaster", "lilcaster.yml"))
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	data, err := os.ReadFile(v.ConfigFileUsed())
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", v.ConfigFileUsed(), err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

func xdgConfigHome() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config")
}
