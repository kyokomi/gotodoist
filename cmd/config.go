// Package cmd provides command-line interface functionality for gotodoist.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/config"
)

// configCmd は設定管理のコマンド
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gotodoist configuration",
	Long: `Configuration management for gotodoist CLI.

This command helps you manage your gotodoist configuration including:
- API token setup
- Configuration file location
- Language preferences
- Current configuration values

Configuration Priority (highest to lowest):
1. Environment variables (TODOIST_API_TOKEN, etc.)
2. Configuration file (~/.config/gotodoist/config.yaml)
3. Default values

The configuration file is automatically generated on first use.`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfig(cmd, args)
	},
}

// configShowCmd は現在の設定を表示するコマンド
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration values and their sources.`,
	Run:   showConfig,
}

// configPathCmd は設定ファイルのパスを表示するコマンド
var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Long:  `Display the path to the configuration file.`,
	Run: func(_ *cobra.Command, _ []string) {
		configDir, err := config.GetConfigDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config directory: %v\n", err)
			os.Exit(1)
		}
		configPath := fmt.Sprintf("%s/config.yaml", configDir)
		fmt.Println(configPath)
	},
}

// configInitCmd は設定ファイルを初期化するコマンド
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long: `Initialize a new configuration file with default values.

This command creates the configuration directory and file if they don't exist.
If the configuration file already exists, it will not be overwritten.`,
	Run: func(_ *cobra.Command, _ []string) {
		configDir, err := config.GetConfigDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config directory: %v\n", err)
			os.Exit(1)
		}

		configPath := fmt.Sprintf("%s/config.yaml", configDir)

		// ファイルがすでに存在するかチェック
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("Configuration file already exists: %s\n", configPath)
			return
		}

		// 設定ファイルを生成（LoadConfigが内部で生成する）
		_, err = config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Configuration file created: %s\n", configPath)
		fmt.Println("\nNext steps:")
		fmt.Println("1. Edit the configuration file to add your Todoist API token")
		fmt.Println("2. Or set the TODOIST_API_TOKEN environment variable")
		fmt.Println("3. Get your API token from: https://todoist.com/prefs/integrations")
	},
}

func showConfig(_ *cobra.Command, _ []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Current Configuration:")
	fmt.Printf("  API Token: %s\n", maskToken(cfg.APIToken))
	fmt.Printf("  Base URL:  %s\n", cfg.BaseURL)
	fmt.Printf("  Language:  %s\n", cfg.Language)

	configDir, err := config.GetConfigDir()
	if err == nil {
		configPath := fmt.Sprintf("%s/config.yaml", configDir)
		fmt.Printf("\nConfiguration file: %s\n", configPath)
	}

	fmt.Println("\nEnvironment Variables:")
	if token := os.Getenv("TODOIST_API_TOKEN"); token != "" {
		fmt.Printf("  TODOIST_API_TOKEN: %s\n", maskToken(token))
	} else {
		fmt.Println("  TODOIST_API_TOKEN: (not set)")
	}

	if lang := os.Getenv("GOTODOIST_LANG"); lang != "" {
		fmt.Printf("  GOTODOIST_LANG: %s\n", lang)
	} else {
		fmt.Println("  GOTODOIST_LANG: (not set)")
	}
}

// maskToken はトークンの一部を隠す
func maskToken(token string) string {
	if token == "" {
		return notSetToken
	}
	if len(token) < minTokenLength {
		return maskedToken
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func init() {
	// サブコマンドの追加
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configInitCmd)

	// ルートコマンドにconfigコマンドを追加
	rootCmd.AddCommand(configCmd)
}
