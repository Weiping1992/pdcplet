/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"pdcplet/pkg/config"
	"pdcplet/pkg/framework"
	"pdcplet/pkg/log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const TARGET_NAME = "pdcplet"

var CONFIG_LOOKUP_PATH = []string{
	".",
	"/etc/" + TARGET_NAME + "/",
}

var configFilePath string

var configContent config.ConfigurationFile

var rootCmd = &cobra.Command{
	Use:   TARGET_NAME,
	Short: "The control-plane component of ProtoDissectCloudPlat",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logging
		initLog()

		f := framework.NewFramework()
		modulesNeedToStart := configContent.Modules
		if len(modulesNeedToStart) == 0 {
			slog.Error("No modules specified in config")
			panic("No modules specified in config")
		}

		slog.Info("Modules to start", "modules", listModules(modulesNeedToStart))
		for _, m := range modulesNeedToStart {
			moduleName := m.Name
			moduleParams := m.Config.Params
			moduleConnections := make([]config.Connection, 0, len(m.Config.Connections))
			for _, connName := range m.Config.Connections {
				alreadyFound := false
				for _, c := range configContent.Connections {
					if c.Name == connName {
						moduleConnections = append(moduleConnections, c)
						alreadyFound = true
						break
					}
				}
				if !alreadyFound {
					slog.Error("Connection not found in config", "connection", connName)
					panic(fmt.Errorf("connection %s not found in config", connName))
				}
			}
			err := f.AddModule(moduleName, moduleParams, moduleConnections)
			if err != nil {
				slog.Error("Failed to create module", "module", moduleName, "error", err)
				panic(fmt.Errorf("failed to create module %s: %w", moduleName, err))
			}
		}

		// Start the framework
		slog.Info("Starting framework")
		f.Start()
		slog.Info("Framework started successfully")
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
	var logCfg config.LogConfig
	viper.UnmarshalKey("log", &logCfg)
	version := viper.GetString("version")
	log.InitLogger(logCfg, TARGET_NAME, version)
}

func listModules(list []config.Module) string {
	if len(list) == 0 {
		return ""
	}

	names := make([]string, 0, len(list))
	for _, module := range list {
		names = append(names, module.Name)
	}
	return strings.Join(names, ", ")
}
