package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestOutput_Successf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Successf("Task created successfully")

	expected := "✅ Task created successfully\n"
	if stdout.String() != expected {
		t.Errorf("Successf() stdout = %q, want %q", stdout.String(), expected)
	}
	if stderr.String() != "" {
		t.Errorf("Successf() stderr = %q, want empty", stderr.String())
	}
}

func TestOutput_Warningf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Warningf("failed to close repository")

	expected := "⚠️  Warning: failed to close repository\n"
	if stderr.String() != expected {
		t.Errorf("Warningf() stderr = %q, want %q", stderr.String(), expected)
	}
	if stdout.String() != "" {
		t.Errorf("Warningf() stdout = %q, want empty", stdout.String())
	}
}

func TestOutput_Errorf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Errorf("loading configuration failed")

	expected := "❌ Error: loading configuration failed\n"
	if stderr.String() != expected {
		t.Errorf("Errorf() stderr = %q, want %q", stderr.String(), expected)
	}
	if stdout.String() != "" {
		t.Errorf("Errorf() stdout = %q, want empty", stdout.String())
	}
}

func TestOutput_Debugf_Verbose(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, true) // verbose = true

	output.Debugf("configuration loaded successfully")

	expected := "🔍 Debug: configuration loaded successfully\n"
	if stderr.String() != expected {
		t.Errorf("Debugf() stderr = %q, want %q", stderr.String(), expected)
	}
	if stdout.String() != "" {
		t.Errorf("Debugf() stdout = %q, want empty", stdout.String())
	}
}

func TestOutput_Debugf_NotVerbose(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false) // verbose = false

	output.Debugf("configuration loaded successfully")

	if stderr.String() != "" {
		t.Errorf("Debugf() stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "" {
		t.Errorf("Debugf() stdout = %q, want empty", stdout.String())
	}
}

func TestOutput_Listf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Listf("Found %d task(s)", 3)

	expected := "📝 Found 3 task(s)\n"
	if stdout.String() != expected {
		t.Errorf("Listf() stdout = %q, want %q", stdout.String(), expected)
	}
	if stderr.String() != "" {
		t.Errorf("Listf() stderr = %q, want empty", stderr.String())
	}
}

func TestOutput_Projectf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Projectf("Created project: %s", "New Project")

	expected := "📁 Created project: New Project\n"
	if stdout.String() != expected {
		t.Errorf("Projectf() stdout = %q, want %q", stdout.String(), expected)
	}
}

func TestOutput_Taskf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Taskf("Task completed: %s", "Buy groceries")

	expected := "📋 Task completed: Buy groceries\n"
	if stdout.String() != expected {
		t.Errorf("Taskf() stdout = %q, want %q", stdout.String(), expected)
	}
}

func TestOutput_Syncf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Syncf("Synchronization completed")

	expected := "🔄 Synchronization completed\n"
	if stdout.String() != expected {
		t.Errorf("Syncf() stdout = %q, want %q", stdout.String(), expected)
	}
}

func TestOutput_Plainf(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Plainf("   ID: %s", "12345")

	expected := "   ID: 12345\n"
	if stdout.String() != expected {
		t.Errorf("Plainf() stdout = %q, want %q", stdout.String(), expected)
	}
}

func TestOutput_PlainNoNewlinef(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.PlainNoNewlinef("Enter your choice: ")

	expected := "Enter your choice: "
	if stdout.String() != expected {
		t.Errorf("PlainNoNewlinef() stdout = %q, want %q", stdout.String(), expected)
	}
}

func TestOutput_Infof(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	output.Infof("Configuration file: %s", "/path/to/config.yaml")

	expected := "Configuration file: /path/to/config.yaml\n"
	if stdout.String() != expected {
		t.Errorf("Infof() stdout = %q, want %q", stdout.String(), expected)
	}
}

func TestNew(t *testing.T) {
	output := New(true)

	if output.verbose != true {
		t.Errorf("New() verbose = %v, want %v", output.verbose, true)
	}
	if output.stdout == nil {
		t.Error("New() stdout should not be nil")
	}
	if output.stderr == nil {
		t.Error("New() stderr should not be nil")
	}
}

func TestNewWithWriters(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewWithWriters(&stdout, &stderr, false)

	if output.verbose != false {
		t.Errorf("NewWithWriters() verbose = %v, want %v", output.verbose, false)
	}
	if output.stdout != &stdout {
		t.Error("NewWithWriters() stdout should match provided writer")
	}
	if output.stderr != &stderr {
		t.Error("NewWithWriters() stderr should match provided writer")
	}
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

			if hasStdout != tt.expectStdout {
				t.Errorf("%s: stdout output = %v, want %v (content: %q)",
					tt.name, hasStdout, tt.expectStdout, stdout.String())
			}
			if hasStderr != tt.expectStderr {
				t.Errorf("%s: stderr output = %v, want %v (content: %q)",
					tt.name, hasStderr, tt.expectStderr, stderr.String())
			}

			if tt.expectStdout && !strings.HasPrefix(stdout.String(), tt.expectedPrefix) {
				t.Errorf("%s: stdout should start with %q, got %q",
					tt.name, tt.expectedPrefix, stdout.String())
			}
			if tt.expectStderr && !strings.HasPrefix(stderr.String(), tt.expectedPrefix) {
				t.Errorf("%s: stderr should start with %q, got %q",
					tt.name, tt.expectedPrefix, stderr.String())
			}
		})
	}
}
