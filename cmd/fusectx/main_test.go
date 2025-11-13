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

	t.Run("clean command", func(t *testing.T) {
		// Test clean for single file
		content := "# Test Clean\nTest content"
		err := os.WriteFile("testclean.md", []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// Build the file first
		cmd := exec.Command(binaryPath, "build", "testclean.md", "-o", "testclean.ctx")
		_, err = cmd.Output()
		if err != nil {
			t.Fatalf("build command failed: %v", err)
		}

		// Verify .ctx file exists
		if _, err := os.Stat("testclean.ctx"); os.IsNotExist(err) {
			t.Fatal("testclean.ctx should exist after build")
		}

		// Test dry-run
		cmd = exec.Command(binaryPath, "clean", "testclean.md", "--dry-run")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("clean dry-run failed: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "Would remove: testclean.ctx") {
			t.Errorf("expected 'Would remove: testclean.ctx' in output, got: %s", string(output))
		}

		// Verify file still exists after dry-run
		if _, err := os.Stat("testclean.ctx"); os.IsNotExist(err) {
			t.Fatal("testclean.ctx should still exist after dry-run")
		}

		// Actually clean the file
		cmd = exec.Command(binaryPath, "clean", "testclean.md")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("clean command failed: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "Removed: testclean.ctx") {
			t.Errorf("expected 'Removed: testclean.ctx' in output, got: %s", string(output))
		}

		// Verify file is removed
		if _, err := os.Stat("testclean.ctx"); !os.IsNotExist(err) {
			t.Fatal("testclean.ctx should be removed after clean")
		}

		// Test clean with custom output path
		cmd = exec.Command(binaryPath, "build", "testclean.md", "-o", "custom.out")
		_, err = cmd.Output()
		if err != nil {
			t.Fatalf("build with custom output failed: %v", err)
		}

		cmd = exec.Command(binaryPath, "clean", "testclean.md", "-o", "custom.out")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("clean with custom output failed: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "Removed: custom.out") {
			t.Errorf("expected 'Removed: custom.out' in output, got: %s", string(output))
		}

		// Verify custom file is removed
		if _, err := os.Stat("custom.out"); !os.IsNotExist(err) {
			t.Fatal("custom.out should be removed after clean")
		}
	})

	t.Run("clean-all command", func(t *testing.T) {
		// Create a subdirectory for clean tests
		cleanDir := "clean-test"
		err := os.MkdirAll(cleanDir, 0755)
		if err != nil {
			t.Fatalf("failed to create clean test directory: %v", err)
		}

		// Create test files
		testFiles := map[string]string{
			filepath.Join(cleanDir, "fusectx.md"):  "# Test 1",
			filepath.Join(cleanDir, "fusectx.ctx"): "Generated content 1",
			filepath.Join(cleanDir, "other.md"):    "# Test 2",
			filepath.Join(cleanDir, "other.ctx"):   "Generated content 2",
			filepath.Join(cleanDir, "orphan.ctx"):  "Orphan ctx file",
		}

		for path, content := range testFiles {
			err := os.WriteFile(path, []byte(content), 0644)
			if err != nil {
				t.Fatalf("failed to write test file %s: %v", path, err)
			}
		}

		// Test dry-run first
		cmd := exec.Command(binaryPath, "clean-all", cleanDir, "--dry-run")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("clean dry-run command failed: %v\nOutput: %s", err, string(output))
		}

		// Check dry-run output
		outputStr := string(output)
		if !strings.Contains(outputStr, "Would remove:") {
			t.Error("expected dry-run message in output")
		}
		if !strings.Contains(outputStr, "Would remove: "+filepath.Join(cleanDir, "fusectx.ctx")) {
			t.Errorf("expected 'Would remove: fusectx.ctx' in dry-run output, got:\n%s", outputStr)
		}
		if !strings.Contains(outputStr, "Would remove: "+filepath.Join(cleanDir, "other.ctx")) {
			t.Errorf("expected 'Would remove: other.ctx' in dry-run output, got:\n%s", outputStr)
		}
		// orphan.ctx should be mentioned as "Skipping", not as "Would remove"
		if strings.Contains(outputStr, "Would remove: "+filepath.Join(cleanDir, "orphan.ctx")) {
			t.Errorf("orphan.ctx should not be in 'Would remove' output (no corresponding .md file)\nFull output:\n%s", outputStr)
		}
		if !strings.Contains(outputStr, "Skipping "+filepath.Join(cleanDir, "orphan.ctx")) {
			t.Errorf("expected 'Skipping orphan.ctx' message in output, got:\n%s", outputStr)
		}

		// Verify files still exist after dry-run
		for path := range testFiles {
			if strings.HasSuffix(path, ".ctx") {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("file %s should still exist after dry-run", path)
				}
			}
		}

		// Test actual clean-all
		cmd = exec.Command(binaryPath, "clean-all", cleanDir)
		output, err = cmd.Output()
		if err != nil {
			t.Fatalf("clean command failed: %v", err)
		}

		// Check clean output
		outputStr = string(output)
		if !strings.Contains(outputStr, "Removed:") {
			t.Error("expected removal message in output")
		}
		if !strings.Contains(outputStr, "Removed 2 file(s)") {
			t.Error("expected correct count of removed files")
		}

		// Verify .ctx files with corresponding .md files are removed
		if _, err := os.Stat(filepath.Join(cleanDir, "fusectx.ctx")); !os.IsNotExist(err) {
			t.Error("fusectx.ctx should be removed")
		}
		if _, err := os.Stat(filepath.Join(cleanDir, "other.ctx")); !os.IsNotExist(err) {
			t.Error("other.ctx should be removed")
		}

		// Verify orphan.ctx is not removed (no corresponding .md)
		if _, err := os.Stat(filepath.Join(cleanDir, "orphan.ctx")); os.IsNotExist(err) {
			t.Error("orphan.ctx should not be removed (no corresponding .md file)")
		}

		// Verify .md files are not removed
		if _, err := os.Stat(filepath.Join(cleanDir, "fusectx.md")); os.IsNotExist(err) {
			t.Error("fusectx.md should not be removed")
		}
		if _, err := os.Stat(filepath.Join(cleanDir, "other.md")); os.IsNotExist(err) {
			t.Error("other.md should not be removed")
		}

		// Test clean-all with --force flag
		cmd = exec.Command(binaryPath, "clean-all", cleanDir, "--force")
		output, err = cmd.Output()
		if err != nil {
			t.Fatalf("clean --force command failed: %v", err)
		}

		// Verify orphan.ctx is now removed with --force
		if _, err := os.Stat(filepath.Join(cleanDir, "orphan.ctx")); !os.IsNotExist(err) {
			t.Error("orphan.ctx should be removed with --force flag")
		}
	})
}
