package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// バージョン情報（ビルド時に埋め込む）
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

// versionCmd はバージョン情報を表示するコマンド
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print detailed version information about gotodoist.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("gotodoist version %s\n", version)
		if IsVerbose() || IsDebug() {
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
		}
	},
}
