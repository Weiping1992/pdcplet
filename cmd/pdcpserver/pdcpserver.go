package main

import (
	"fmt"
	"log/slog"
	"os"
	baseconfig "pdcplet/pkg/config"
	"pdcplet/pkg/log"
	"pdcplet/pkg/pdcpserver"
	"pdcplet/pkg/pdcpserver/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const TARGET_NAME = "pdcpserver"

var CONFIG_LOOKUP_PATH = []string{
	".",
	"/etc/" + TARGET_NAME + "/",
}

var configFilePath string

var configContent config.ConfigurationFile

var rootCmd = &cobra.Command{
	Use:   TARGET_NAME,
	Short: "The server component of ProtoDissectCloudPlat",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logging
		initLog()

		address := viper.GetString("listen.address")
		port := viper.GetUint32("listen.port")
		dbType := viper.GetString("db.type")
		var dbFilePath string
		switch dbType {
		case "sqlite3":
			dbFilePath = viper.GetString("db.sqlite3.database")
		default:
			slog.Error("Unsupported db type", "dbType", dbType)
		}
		s := pdcpserver.New(address, port, dbFilePath)

		slog.Info("Starting server", "address", address, "port", port)
		if err := s.Start(); err != nil {
			slog.Error("Failed to start server", "error", err)
			panic(fmt.Errorf("failed to start server: %w", err))
		}
		slog.Info("Server started successfully")
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "f", "", "Config file path")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	if configFilePath != "" {
		viper.SetConfigFile(configFilePath)
	} else {
		viper.SetConfigName(TARGET_NAME + ".yaml")
		for _, path := range CONFIG_LOOKUP_PATH {
			viper.AddConfigPath(path)
		}
	}
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No config file found, using defaults")
		} else {
			panic(fmt.Errorf("config file error: %w", err))
		}
	}

	// Unmarshal the config file into the configContent struct
	if err := viper.Unmarshal(&configContent); err != nil {
		panic(fmt.Errorf("failed to unmarshal config file: %w", err))
	}
	fmt.Printf("%v\n", configContent)
	viper.AutomaticEnv()
	viper.SetEnvPrefix(TARGET_NAME)
}

func initLog() {
	// Init log
	var logCfg baseconfig.LogConfig
	viper.UnmarshalKey("log", &logCfg)
	version := viper.GetString("version")
	log.InitLogger(logCfg, TARGET_NAME, version)
}

func main() {
	Execute()
}
