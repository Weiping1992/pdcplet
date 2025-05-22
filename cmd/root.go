/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
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

type WatchMode int

const (
	WatchModeUnsupported WatchMode = iota
	WatchModeListWatch
	WatchModeWebhook
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   TARGET_NAME,
	Short: "The control-plane component of ProtoDissectCloudPlat",
	Run: func(cmd *cobra.Command, args []string) {
		// Init log
		var logCfg log.LogConfig
		viper.UnmarshalKey("log", &logCfg)
		version := viper.GetString("version")
		log.InitLogger(logCfg, TARGET_NAME, version)

		switch parseWatchModeFlag(cmd) {
		case WatchModeWebhook:
			//fmt.Fprintln(os.Stderr, "Webhook watchmode DO NOT IMPLEMENT") // TODO：[log]
			slog.Error("Webhook watchmode DO NOT IMPLEMENT")
		case WatchModeListWatch:
			f := framework.NewFramework([]string{"vmiproxy"})
			if f == nil {
				return
			}
			f.Init()
			f.Start()
		case WatchModeUnsupported:
			slog.Error("Unknow WatchMode which must in ['listandwatch', 'webhook']")
			//fmt.Fprintln(os.Stderr, "Unknow WatchMode which must in ['listandwatch', 'webhook']") // TODO：[log]
		}
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

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "f", "", "Config file path")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringP("watchmode", "m", "ListAndWatch", "监控VM模式，选择 listandwatch 或 webhook")
}

func parseWatchModeFlag(cmd *cobra.Command) (m WatchMode) {
	watchModeStr, _ := cmd.Flags().GetString("watchmode")
	if strings.ToLower(watchModeStr) == "webhook" {
		m = WatchModeWebhook
	} else if strings.ToLower(watchModeStr) == "listandwatch" {
		m = WatchModeListWatch
	} else {
		m = WatchModeUnsupported
	}
	return
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
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
			panic(fmt.Errorf("Config file error: %w", err))
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix(TARGET_NAME)
}
