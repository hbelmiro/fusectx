package resolver

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Frontmatter struct {
	Extends  string   `yaml:"extends"`
	Includes []string `yaml:"includes"`
}

const frontmatterSeparator = "---"

func ParseFrontmatter(reader io.Reader) (*Frontmatter, string, error) {
	scanner := bufio.NewScanner(reader)
	var lines []string
	var inFrontmatter bool
	var frontmatterLines []string
	var contentLines []string

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)

		if len(lines) == 1 && strings.TrimSpace(line) == frontmatterSeparator {
			inFrontmatter = true
			continue
		}

		if inFrontmatter {
			if strings.TrimSpace(line) == frontmatterSeparator {
				inFrontmatter = false
				continue
			}
			frontmatterLines = append(frontmatterLines, line)
		} else {
			if len(lines) > 1 || strings.TrimSpace(line) != frontmatterSeparator {
				contentLines = append(contentLines, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("error reading file: %w", err)
	}

	var frontmatter Frontmatter
	if len(frontmatterLines) > 0 {
		frontmatterContent := strings.Join(frontmatterLines, "\n")
		if err := yaml.Unmarshal([]byte(frontmatterContent), &frontmatter); err != nil {
			return nil, "", fmt.Errorf("error parsing frontmatter: %w", err)
		}
	}

	content := strings.Join(contentLines, "\n")
	return &frontmatter, content, nil
}

func Resolve(filePath string, visited map[string]bool) (string, error) {
	if visited == nil {
		visited = make(map[string]bool)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("error resolving absolute path for %s: %w", filePath, err)
	}

	if visited[absPath] {
		return "", fmt.Errorf("circular dependency detected: %s", absPath)
	}

	visited[absPath] = true
	defer func() { delete(visited, absPath) }()

	file, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("error opening file %s: %w", absPath, err)
	}
	defer file.Close()

	frontmatter, content, err := ParseFrontmatter(file)
	if err != nil {
		return "", fmt.Errorf("error parsing file %s: %w", absPath, err)
	}

	var result strings.Builder

	if frontmatter.Extends != "" {
		extendsPath := resolvePath(frontmatter.Extends, filepath.Dir(absPath))
		extendsContent, err := Resolve(extendsPath, visited)
		if err != nil {
			return "", fmt.Errorf("error resolving extends file %s: %w", extendsPath, err)
		}
		if extendsContent != "" {
			result.WriteString(extendsContent)
			result.WriteString("\n\n")
		}
	}

	for _, includePath := range frontmatter.Includes {
		includeFullPath := resolvePath(includePath, filepath.Dir(absPath))
		includeContent, err := Resolve(includeFullPath, visited)
		if err != nil {
			return "", fmt.Errorf("error resolving include file %s: %w", includeFullPath, err)
		}
		if includeContent != "" {
			result.WriteString(includeContent)
			result.WriteString("\n\n")
		}
	}

	if content != "" {
		result.WriteString(content)
	}

	return strings.TrimSpace(result.String()), nil
}

func resolvePath(path, basePath string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(basePath, path)
}

func ValidateChain(filePath string) error {
	_, err := Resolve(filePath, nil)
	return err
}

func GetDependencyChain(filePath string, visited map[string]bool) ([]string, error) {
	if visited == nil {
		visited = make(map[string]bool)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error resolving absolute path for %s: %w", filePath, err)
	}

	if visited[absPath] {
		return nil, fmt.Errorf("circular dependency detected: %s", absPath)
	}

	visited[absPath] = true
	defer func() { delete(visited, absPath) }()

	var chain []string
	chain = append(chain, absPath)

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", absPath, err)
	}
	defer file.Close()

	frontmatter, _, err := ParseFrontmatter(file)
	if err != nil {
		return nil, fmt.Errorf("error parsing file %s: %w", absPath, err)
	}

	if frontmatter.Extends != "" {
		extendsPath := resolvePath(frontmatter.Extends, filepath.Dir(absPath))
		extendsChain, err := GetDependencyChain(extendsPath, visited)
		if err != nil {
			return nil, err
		}
		chain = append(extendsChain, chain...)
	}

	for _, includePath := range frontmatter.Includes {
		includeFullPath := resolvePath(includePath, filepath.Dir(absPath))
		includeChain, err := GetDependencyChain(includeFullPath, visited)
		if err != nil {
			return nil, err
		}
		chain = append(chain, includeChain...)
	}

	return chain, nil
}