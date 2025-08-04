package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedFM      Frontmatter
		expectedContent string
		shouldError     bool
	}{
		{
			name: "no frontmatter",
			input: `# Header
Content here`,
			expectedFM:      Frontmatter{},
			expectedContent: "# Header\nContent here",
			shouldError:     false,
		},
		{
			name: "frontmatter with extends",
			input: `---
extends: base.md
---
# Header
Content here`,
			expectedFM:      Frontmatter{Extends: "base.md"},
			expectedContent: "# Header\nContent here",
			shouldError:     false,
		},
		{
			name: "frontmatter with includes",
			input: `---
includes:
  - file1.md
  - file2.md
---
Content`,
			expectedFM:      Frontmatter{Includes: []string{"file1.md", "file2.md"}},
			expectedContent: "Content",
			shouldError:     false,
		},
		{
			name: "frontmatter with extends and includes",
			input: `---
extends: base.md
includes:
  - file1.md
  - file2.md
---
Content`,
			expectedFM: Frontmatter{
				Extends:  "base.md",
				Includes: []string{"file1.md", "file2.md"},
			},
			expectedContent: "Content",
			shouldError:     false,
		},
		{
			name: "empty frontmatter",
			input: `---
---
Content`,
			expectedFM:      Frontmatter{},
			expectedContent: "Content",
			shouldError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			fm, content, err := ParseFrontmatter(reader)

			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if fm.Extends != tt.expectedFM.Extends {
				t.Errorf("expected extends %q, got %q", tt.expectedFM.Extends, fm.Extends)
			}

			if len(fm.Includes) != len(tt.expectedFM.Includes) {
				t.Errorf("expected %d includes, got %d", len(tt.expectedFM.Includes), len(fm.Includes))
			}

			for i, include := range tt.expectedFM.Includes {
				if i >= len(fm.Includes) || fm.Includes[i] != include {
					t.Errorf("expected include %d to be %q, got %q", i, include, fm.Includes[i])
				}
			}

			if strings.TrimSpace(content) != strings.TrimSpace(tt.expectedContent) {
				t.Errorf("expected content %q, got %q", tt.expectedContent, content)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fusectx-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		files    map[string]string
		target   string
		expected string
		hasError bool
	}{
		{
			name: "single file no dependencies",
			files: map[string]string{
				"simple.md": "# Simple File\nContent",
			},
			target:   "simple.md",
			expected: "# Simple File\nContent",
			hasError: false,
		},
		{
			name: "simple extends chain",
			files: map[string]string{
				"base.md": "# Base\nBase content",
				"child.md": `---
extends: base.md
---
# Child
Child content`,
			},
			target:   "child.md",
			expected: "# Base\nBase content\n\n# Child\nChild content",
			hasError: false,
		},
		{
			name: "multi-level extends chain",
			files: map[string]string{
				"root.md": "# Root\nRoot content",
				"middle.md": `---
extends: root.md
---
# Middle
Middle content`,
				"leaf.md": `---
extends: middle.md
---
# Leaf
Leaf content`,
			},
			target:   "leaf.md",
			expected: "# Root\nRoot content\n\n# Middle\nMiddle content\n\n# Leaf\nLeaf content",
			hasError: false,
		},
		{
			name: "includes only",
			files: map[string]string{
				"inc1.md": "# Include 1\nContent 1",
				"inc2.md": "# Include 2\nContent 2",
				"main.md": `---
includes:
  - inc1.md
  - inc2.md
---
# Main
Main content`,
			},
			target:   "main.md",
			expected: "# Include 1\nContent 1\n\n# Include 2\nContent 2\n\n# Main\nMain content",
			hasError: false,
		},
		{
			name: "extends and includes",
			files: map[string]string{
				"base.md": "# Base\nBase content",
				"inc1.md": "# Include 1\nInclude content",
				"main.md": `---
extends: base.md
includes:
  - inc1.md
---
# Main
Main content`,
			},
			target:   "main.md",
			expected: "# Base\nBase content\n\n# Include 1\nInclude content\n\n# Main\nMain content",
			hasError: false,
		},
		{
			name: "circular dependency",
			files: map[string]string{
				"a.md": `---
extends: b.md
---
Content A`,
				"b.md": `---
extends: a.md
---
Content B`,
			},
			target:   "a.md",
			expected: "",
			hasError: true,
		},
		{
			name: "missing file",
			files: map[string]string{
				"main.md": `---
extends: missing.md
---
Content`,
			},
			target:   "main.md",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			if err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}

			for filename, content := range tt.files {
				filePath := filepath.Join(testDir, filename)
				err := os.WriteFile(filePath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("failed to write test file %s: %v", filename, err)
				}
			}

			targetPath := filepath.Join(testDir, tt.target)
			result, err := Resolve(targetPath, nil)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.hasError && strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestValidateChain(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fusectx-validate-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	validFile := filepath.Join(tmpDir, "valid.md")
	err = os.WriteFile(validFile, []byte("# Valid\nContent"), 0644)
	if err != nil {
		t.Fatalf("failed to write valid file: %v", err)
	}

	invalidFile := filepath.Join(tmpDir, "invalid.md")
	err = os.WriteFile(invalidFile, []byte(`---
extends: nonexistent.md
---
Content`), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}

	err = ValidateChain(validFile)
	if err != nil {
		t.Errorf("expected valid file to pass validation, got error: %v", err)
	}

	err = ValidateChain(invalidFile)
	if err == nil {
		t.Error("expected invalid file to fail validation")
	}
}

func TestGetDependencyChain(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fusectx-chain-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	files := map[string]string{
		"root.md": "# Root\nRoot content",
		"middle.md": `---
extends: root.md
---
# Middle
Middle content`,
		"leaf.md": `---
extends: middle.md
includes:
  - inc.md
---
# Leaf
Leaf content`,
		"inc.md": "# Include\nInclude content",
	}

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write test file %s: %v", filename, err)
		}
	}

	leafPath := filepath.Join(tmpDir, "leaf.md")
	chain, err := GetDependencyChain(leafPath, nil)
	if err != nil {
		t.Fatalf("unexpected error getting dependency chain: %v", err)
	}

	if len(chain) < 3 {
		t.Errorf("expected at least 3 files in chain, got %d", len(chain))
	}

	chainFiles := make([]string, len(chain))
	for i, path := range chain {
		chainFiles[i] = filepath.Base(path)
	}

	expectedFiles := []string{"root.md", "middle.md", "leaf.md", "inc.md"}
	for _, expected := range expectedFiles {
		found := false
		for _, actual := range chainFiles {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected file %s to be in chain, but it wasn't found", expected)
		}
	}
}