package config

import "pdcplet/pkg/config"

// ConfigurationFile represents the structure of the configuration file.
type ConfigurationFile struct {
	Version string           `mapstructure:"version"`
	Log     config.LogConfig `mapstructure:"log"`
	Listen  ListenConfig     `mapstructure:"listen"`
	Db      DBConfig         `mapstructure:"db"`
}

// Module represents a module configuration with its name and parameters
type ListenConfig struct {
	Address string `mapstructure:"address"`
	Port    uint32 `mapstructure:"port"`
}

type DBConfig struct {
	TypeStr string `mapstructure:"type"`
	Sqlite3 struct {
		Database string `mapstructure:"database"`
	} `mapstructure:"sqlite3"`
}
