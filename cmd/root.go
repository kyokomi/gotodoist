// Package cmd provides command-line interface functionality for gotodoist.
// This package contains all CLI commands and related implementations using cobra framework.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/config"
)

const (
	notSetToken    = "(not set)"
	maskedToken    = "***"
	minTokenLength = 8
)

// GlobalFlags はコマンドラインフラグをまとめた構造体
type GlobalFlags struct {
	Verbose       bool
	Debug         bool
	ShowBenchmark bool
}

var (
	// グローバルフラグ
	globalFlags GlobalFlags
	// アプリケーション設定
	appConfig *config.Config
)

func init() {
	// グローバルフラグの設定
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.Verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&globalFlags.Debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().BoolVar(&globalFlags.ShowBenchmark, "benchmark", false, "show detailed performance timing")

	// 設定の初期化
	cobra.OnInitialize(initConfig)
}

// rootCmdはアプリケーションのベースコマンド
var rootCmd = &cobra.Command{
	Use:   "gotodoist",
	Short: "A CLI tool for Todoist",
	Long: `gotodoist is a command-line interface tool for managing your Todoist tasks.
	
This tool allows you to manage tasks, projects, and more directly from your terminal.

Configuration:
  Configuration file: ~/.config/gotodoist/config.yaml
  Environment variable: TODOIST_API_TOKEN

Get started:
  1. Set your API token: export TODOIST_API_TOKEN="your-token"
  2. Or run: gotodoist config init
  3. Get your API token from: https://todoist.com/prefs/integrations

For more configuration options, run: gotodoist config --help`,
	// ここではルートコマンド自体は何も実行しない
	// サブコマンドが指定されていない場合はヘルプを表示
}

// Execute はコマンドのエントリーポイント
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// initConfig は設定を読み込む
func initConfig() {
	var err error
	appConfig, err = config.LoadConfig()
	if err != nil {
		// 設定読み込みエラーは致命的エラーとして扱う
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// デバッグモードの場合、設定情報を表示
	if globalFlags.Debug {
		fmt.Fprintf(os.Stderr, "Configuration loaded:\n")
		fmt.Fprintf(os.Stderr, "  API Token: %s\n", maskToken(appConfig.APIToken))
		fmt.Fprintf(os.Stderr, "  Base URL:  %s\n", appConfig.BaseURL)
	}
}

// GetAppConfig はアプリケーション設定を返す
func GetAppConfig() *config.Config {
	return appConfig
}
