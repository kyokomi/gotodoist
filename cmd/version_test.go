package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const cmdVersion = "version"

func TestVersionCommandDefinition(t *testing.T) {
	assert.Equal(t, cmdVersion, versionCmd.Use, "versionコマンドのUseが期待値と異なります")
	assert.Equal(t, "Print version information", versionCmd.Short, "versionコマンドのShortが期待値と異なります")
	assert.NotNil(t, versionCmd.Run, "versionコマンドのRun関数が定義されていません")
	assert.Nil(t, versionCmd.Args, "versionコマンドは引数を受け取るべきではありません")
}

func TestVersionCommand_Integration(t *testing.T) {
	// versionコマンドがrootコマンドに追加されているかテスト
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == cmdVersion {
			found = true
			break
		}
	}

	assert.True(t, found, "versionコマンドがrootコマンドに追加されていません")
}
