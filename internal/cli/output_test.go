package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutput_Successf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Successf("Task created successfully")

	expected := "✅ Task created successfully\n"
	assert.Equal(t, expected, stdout.String(), "Successf()のstdout出力が期待値と異なります")
	assert.Empty(t, stderr.String(), "Successf()のstderr出力が空ではありません")
}

func TestOutput_Warningf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Warningf("failed to close repository")

	expected := "⚠️  Warning: failed to close repository\n"
	assert.Equal(t, expected, stderr.String(), "Warningf()のstderr出力が期待値と異なります")
	assert.Empty(t, stdout.String(), "Warningf()のstdout出力が空ではありません")
}

func TestOutput_Errorf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Errorf("loading configuration failed")

	expected := "❌ Error: loading configuration failed\n"
	assert.Equal(t, expected, stderr.String(), "Errorf()のstderr出力が期待値と異なります")
	assert.Empty(t, stdout.String(), "Errorf()のstdout出力が空ではありません")
}

func TestOutput_Debugf_Verbose(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, true) // verbose = true

	output.Debugf("configuration loaded successfully")

	expected := "🔍 Debug: configuration loaded successfully\n"
	assert.Equal(t, expected, stderr.String(), "Debugf()のstderr出力が期待値と異なります")
	assert.Empty(t, stdout.String(), "Debugf()のstdout出力が空ではありません")
}

func TestOutput_Debugf_NotVerbose(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false) // verbose = false

	output.Debugf("configuration loaded successfully")

	assert.Empty(t, stderr.String(), "Debugf()のstderr出力が空ではありません")
	assert.Empty(t, stdout.String(), "Debugf()のstdout出力が空ではありません")
}

func TestOutput_Listf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Listf("Found %d task(s)", 3)

	expected := "📝 Found 3 task(s)\n"
	assert.Equal(t, expected, stdout.String(), "Listf()のstdout出力が期待値と異なります")
	assert.Empty(t, stderr.String(), "Listf()のstderr出力が空ではありません")
}

func TestOutput_Projectf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Projectf("Created project: %s", "New Project")

	expected := "📁 Created project: New Project\n"
	assert.Equal(t, expected, stdout.String(), "Projectf()のstdout出力が期待値と異なります")
}

func TestOutput_Taskf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Taskf("Task completed: %s", "Buy groceries")

	expected := "📋 Task completed: Buy groceries\n"
	assert.Equal(t, expected, stdout.String(), "Taskf()のstdout出力が期待値と異なります")
}

func TestOutput_Syncf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Syncf("Synchronization completed")

	expected := "🔄 Synchronization completed\n"
	assert.Equal(t, expected, stdout.String(), "Syncf()のstdout出力が期待値と異なります")
}

func TestOutput_Plainf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Plainf("   ID: %s", "12345")

	expected := "   ID: 12345\n"
	assert.Equal(t, expected, stdout.String(), "Plainf()のstdout出力が期待値と異なります")
}

func TestOutput_PlainNoNewlinef(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.PlainNoNewlinef("Enter your choice: ")

	expected := "Enter your choice: "
	assert.Equal(t, expected, stdout.String(), "PlainNoNewlinef()のstdout出力が期待値と異なります")
}

func TestOutput_Infof(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Infof("Configuration file: %s", "/path/to/config.yaml")

	expected := "Configuration file: /path/to/config.yaml\n"
	assert.Equal(t, expected, stdout.String(), "Infof()のstdout出力が期待値と異なります")
}

func TestNew(t *testing.T) {
	output := New(true)

	assert.Equal(t, true, output.verbose, "New()のverbose設定が期待値と異なります")
	assert.NotNil(t, output.stdout, "New()のstdoutがnilです")
	assert.NotNil(t, output.stderr, "New()のstderrがnilです")
}

func TestNewWithWriters(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	assert.Equal(t, false, output.verbose, "NewWithWriters()のverbose設定が期待値と異なります")
	assert.Equal(t, &stdout, output.stdout, "NewWithWriters()のstdoutが指定したwriterと一致しません")
	assert.Equal(t, &stderr, output.stderr, "NewWithWriters()のstderrが指定したwriterと一致しません")
}

func TestOutput_AllMethodsHaveCorrectOutputStreams(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, true)

	tests := []struct {
		name           string
		action         func()
		expectStdout   bool
		expectStderr   bool
		expectedPrefix string
	}{
		{
			name:           "Success",
			action:         func() { output.Successf("test") },
			expectStdout:   true,
			expectStderr:   false,
			expectedPrefix: "✅",
		},
		{
			name:           "Info",
			action:         func() { output.Infof("test") },
			expectStdout:   true,
			expectStderr:   false,
			expectedPrefix: "test",
		},
		{
			name:           "Warning",
			action:         func() { output.Warningf("test") },
			expectStdout:   false,
			expectStderr:   true,
			expectedPrefix: "⚠️",
		},
		{
			name:           "Error",
			action:         func() { output.Errorf("test") },
			expectStdout:   false,
			expectStderr:   true,
			expectedPrefix: "❌",
		},
		{
			name:           "Debug",
			action:         func() { output.Debugf("test") },
			expectStdout:   false,
			expectStderr:   true,
			expectedPrefix: "🔍",
		},
		{
			name:           "List",
			action:         func() { output.Listf("test") },
			expectStdout:   true,
			expectStderr:   false,
			expectedPrefix: "📝",
		},
		{
			name:           "Project",
			action:         func() { output.Projectf("test") },
			expectStdout:   true,
			expectStderr:   false,
			expectedPrefix: "📁",
		},
		{
			name:           "Task",
			action:         func() { output.Taskf("test") },
			expectStdout:   true,
			expectStderr:   false,
			expectedPrefix: "📋",
		},
		{
			name:           "Sync",
			action:         func() { output.Syncf("test") },
			expectStdout:   true,
			expectStderr:   false,
			expectedPrefix: "🔄",
		},
		{
			name:           "Plain",
			action:         func() { output.Plainf("test") },
			expectStdout:   true,
			expectStderr:   false,
			expectedPrefix: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()

			tt.action()

			hasStdout := stdout.Len() > 0
			hasStderr := stderr.Len() > 0

			assert.Equal(t, tt.expectStdout, hasStdout, "%s: stdout出力の有無が期待値と異なります (content: %q)", tt.name, stdout.String())
			assert.Equal(t, tt.expectStderr, hasStderr, "%s: stderr出力の有無が期待値と異なります (content: %q)", tt.name, stderr.String())

			if tt.expectStdout {
				assert.True(t, strings.HasPrefix(stdout.String(), tt.expectedPrefix), "%s: stdoutが期待されるプレフィックスで始まっていません (expected: %q, got: %q)", tt.name, tt.expectedPrefix, stdout.String())
			}
			if tt.expectStderr {
				assert.True(t, strings.HasPrefix(stderr.String(), tt.expectedPrefix), "%s: stderrが期待されるプレフィックスで始まっていません (expected: %q, got: %q)", tt.name, tt.expectedPrefix, stderr.String())
			}
		})
	}
}
