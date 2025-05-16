/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"pdcplet/pkg/app"
	"strings"

	"github.com/spf13/cobra"
)

type WatchMode int

const (
	WatchModeUnsupported WatchMode = iota
	WatchModeListWatch
	WatchModeWebhook
)

var rootCmd = &cobra.Command{
	Use:   "pdcplet",
	Short: "The control-plane component of ProtoDissectCloudPlat",
	Run: func(cmd *cobra.Command, args []string) {
		switch parseWatchModeFlag(cmd) {
		case WatchModeWebhook:
			fmt.Fprintln(os.Stderr, "Webhook watchmode DO NOT IMPLEMENT") // TODO：[log]
		case WatchModeListWatch:
			app.NewApp().Run()
		case WatchModeUnsupported:
			fmt.Fprintln(os.Stderr, "Unknow WatchMode which must in ['listandwatch', 'webhook']") // TODO：[log]
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
