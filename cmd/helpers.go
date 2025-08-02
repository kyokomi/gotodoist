package cmd

import (
	"context"
)

// createBaseContext は統一されたベースコンテキストを作成する
func createBaseContext() context.Context {
	return context.Background()
}

// IsVerbose はverboseモードかどうかを返す
func IsVerbose() bool {
	return globalFlags.Verbose
}

// IsDebug はデバッグモードかどうかを返す
func IsDebug() bool {
	return globalFlags.Debug
}

// GetLanguage は設定された言語を返す
func GetLanguage() string {
	return globalFlags.Lang
}

// IsShowBenchmark はベンチマーク表示モードかどうかを返す
func IsShowBenchmark() bool {
	return globalFlags.ShowBenchmark
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
