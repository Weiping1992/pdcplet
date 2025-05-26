package config

const (
	INPPLAT_CONNECTION_NAME    = "inpplat"
	PDCPSERVER_CONNECTION_NAME = "pdcpserver"
)

// ConfigurationFile represents the structure of the configuration file.
type ConfigurationFile struct {
	Version     string       `mapstructure:"version"`
	Log         LogConfig    `mapstructure:"log"`
	Connections []Connection `mapstructure:"connections"`
	Modules     []Module     `mapstructure:"modules"`
}

// Module represents a module configuration with its name and parameters
type Module struct {
	Name   string       `mapstructure:"name"`
	Config ModuleConfig `mapstructure:"config"`
}

// ModuleConfig represents the configuration for a specific module
type ModuleConfig struct {
	Connections []string               `mapstructure:"connections"`
	Params      map[string]interface{} `mapstructure:"params"`
}

// Connection represents a connection configuration with its name, type, and specific settings
type Connection struct {
	Name       string        `mapstructure:"name"`
	Category   string        `mapstructure:"type"`
	HttpConfig HttpOverTcpIp `mapstructure:"httpOverTcpIp"`
	UnixConfig UnixSocket    `mapstructure:"unixSocket"`
}

func (c *Connection) ConvertToMap() map[string]interface{} {
	result := make(map[string]interface{})
	result["name"] = c.Name
	result["type"] = c.Category

	switch c.Category {
	case "httpOverTcpIp":
		result["httpOverTcpIp"] = map[string]interface{}{
			"host":      c.HttpConfig.Host,
			"port":      c.HttpConfig.Port,
			"urlPrefix": c.HttpConfig.UrlPrefix,
			"authToken": c.HttpConfig.AuthToken,
			"timeout":   c.HttpConfig.Timeout,
		}
	case "unixSocket":
		result["unixSocket"] = map[string]interface{}{
			"path": c.UnixConfig.Path,
		}
	}
	return result
}

// HttpOverTcpIp represents an HTTP over TCP/IP connection configuration
type HttpOverTcpIp struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	UrlPrefix string `mapstructure:"urlPrefix"`
	AuthToken string `mapstructure:"authToken"`
	Timeout   string `mapstructure:"timeout"`
}

// UnixSocket represents a Unix socket connection configuration
type UnixSocket struct {
	Path string `mapstructure:"path"`
}

// LogConfig represents the logging configuration
type LogConfig struct {
	Level   string   `mapstructure:"level"`
	Format  string   `mapstructure:"format"`
	Outputs []string `mapstructure:"outputs"`
	File    struct {
		Path       string `mapstructure:"path"`
		MaxSize    int    `mapstructure:"maxSize"`
		Compress   bool   `mapstructure:"compress"`
		MaxBackups int    `mapstructure:"maxBackups"`
	} `mapstructure:"file"`
}
