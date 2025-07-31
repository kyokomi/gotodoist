package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// フラグ用の変数
	verbose bool
	debug   bool
	lang    string
)

// rootCmdはアプリケーションのベースコマンド
var rootCmd = &cobra.Command{
	Use:   "gotodoist",
	Short: "A CLI tool for Todoist",
	Long: `gotodoist is a command-line interface tool for managing your Todoist tasks.
	
This tool allows you to manage tasks, projects, and more directly from your terminal.`,
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

func init() {
	// グローバルフラグの設定
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVar(&lang, "lang", "", "language preference (en/ja)")

	// 設定の初期化（後で実装）
	cobra.OnInitialize(initConfig)
}

// initConfig は設定を読み込む（後で実装）
func initConfig() {
	// TODO: viperを使った設定管理を実装
	// 環境変数、設定ファイル、フラグの優先順位で設定を読み込む
}
