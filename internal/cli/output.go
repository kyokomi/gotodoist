// Package cli provides command-line interface utilities for gotodoist.
package cli

import (
	"fmt"
	"io"
	"os"
)

// Output はCLI出力を統一管理する構造体
type Output struct {
	stdout  io.Writer
	stderr  io.Writer
	verbose bool
}

// New は新しいOutput構造体を作成する
func New(verbose bool) *Output {
	return &Output{
		stdout:  os.Stdout,
		stderr:  os.Stderr,
		verbose: verbose,
	}
}

// NewWithWriters はテスト用にカスタムWriterを指定できるOutput構造体を作成する
func NewWithWriters(stdout, stderr io.Writer, verbose bool) *Output {
	return &Output{
		stdout:  stdout,
		stderr:  stderr,
		verbose: verbose,
	}
}

// Successf は成功メッセージを出力する（stdout）
func (o *Output) Successf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, "✅ "+format+"\n", args...)
}

// Infof は情報メッセージを出力する（stdout）
func (o *Output) Infof(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, format+"\n", args...)
}

// Warningf は警告メッセージを出力する（stderr）
func (o *Output) Warningf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stderr, "⚠️  Warning: "+format+"\n", args...)
}

// Errorf はエラーメッセージを出力する（stderr）
func (o *Output) Errorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stderr, "❌ Error: "+format+"\n", args...)
}

// Debugf はデバッグメッセージを出力する（verbose時のみ、stderr）
func (o *Output) Debugf(format string, args ...interface{}) {
	if o.verbose {
		_, _ = fmt.Fprintf(o.stderr, "🔍 Debug: "+format+"\n", args...)
	}
}

// Listf はリスト形式のメッセージを出力する（stdout）
func (o *Output) Listf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, "📝 "+format+"\n", args...)
}

// Projectf はプロジェクト関連のメッセージを出力する（stdout）
func (o *Output) Projectf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, "📁 "+format+"\n", args...)
}

// Taskf はタスク関連のメッセージを出力する（stdout）
func (o *Output) Taskf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, "📋 "+format+"\n", args...)
}

// Syncf は同期関連のメッセージを出力する（stdout）
func (o *Output) Syncf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, "🔄 "+format+"\n", args...)
}

// Plainf は装飾なしでメッセージを出力する（stdout）
func (o *Output) Plainf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, format+"\n", args...)
}

// PlainNoNewlinef は装飾なしで改行なしのメッセージを出力する（stdout）
func (o *Output) PlainNoNewlinef(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, format, args...)
}
