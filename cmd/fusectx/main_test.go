package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLICommands(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fusectx-cli-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Get the project root (two levels up from cmd/fusectx)
	projectRoot := filepath.Join(originalDir, "..", "..")

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	binaryPath := filepath.Join(projectRoot, "fusectx-test")

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/fusectx")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to build binary: %v\nOutput: %s", err, string(output))
		}
	}

	t.Run("build single file", func(t *testing.T) {
		content := "# Test\nTest content"
		err := os.WriteFile("test.md", []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		cmd := exec.Command(binaryPath, "build", "test.md")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("build command failed: %v", err)
		}

		if strings.TrimSpace(string(output)) != content {
			t.Errorf("expected %q, got %q", content, string(output))
		}
	})

	t.Run("build with extends", func(t *testing.T) {
		baseContent := "# Base\nBase content"
		err := os.WriteFile("base.md", []byte(baseContent), 0644)
		if err != nil {
			t.Fatalf("failed to write base file: %v", err)
		}

		mainContent := `---
extends: base.md
---
# Main
Main content`
		err = os.WriteFile("main.md", []byte(mainContent), 0644)
		if err != nil {
			t.Fatalf("failed to write main file: %v", err)
		}

		cmd := exec.Command(binaryPath, "build", "main.md")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("build command failed: %v", err)
		}

		expected := "# Base\nBase content\n\n# Main\nMain content"
		if strings.TrimSpace(string(output)) != expected {
			t.Errorf("expected:\n%s\n\ngot:\n%s", expected, string(output))
		}
	})

	t.Run("init command", func(t *testing.T) {
		testDir := "init-test"
		cmd := exec.Command(binaryPath, "init", testDir)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("init command failed: %v", err)
		}

		if !strings.Contains(string(output), "Created") {
			t.Error("expected creation message in output")
		}

		expectedFile := filepath.Join(testDir, "fusectx.md")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Error("expected fusectx.md to be created")
		}
	})

	t.Run("validate command", func(t *testing.T) {
		validContent := "# Valid\nValid content"
		err := os.WriteFile("valid.md", []byte(validContent), 0644)
		if err != nil {
			t.Fatalf("failed to write valid file: %v", err)
		}

		cmd := exec.Command(binaryPath, "validate", "valid.md")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("validate command failed: %v", err)
		}

		if !strings.Contains(string(output), "Validation successful") {
			t.Error("expected validation success message")
		}
	})
}
