package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hbelmiro/fusectx/internal/resolver"
	"github.com/spf13/cobra"
)

const version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "fusectx",
	Short: "A CLI tool for resolving and concatenating hierarchical text files",
	Long: `fusectx recursively resolves a dependency chain of text files and concatenates them into a single output.
It supports both inheritance (extends) and composition (includes) through YAML frontmatter.`,
	Version: version,
}

var buildCmd = &cobra.Command{
	Use:   "build <source_file>",
	Short: "Resolves the full dependency chain and generates the final context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceFile := args[0]
		output, _ := cmd.Flags().GetString("output")
		silent, _ := cmd.Flags().GetBool("silent")

		content, err := resolver.Resolve(sourceFile, nil)
		if err != nil {
			return fmt.Errorf("failed to resolve %s: %w", sourceFile, err)
		}

		if output != "" {
			err = os.WriteFile(output, []byte(content), 0644)
			if err != nil {
				return fmt.Errorf("failed to write to %s: %w", output, err)
			}
			if !silent {
				fmt.Printf("Output written to %s\n", output)
			}
		} else {
			fmt.Print(content)
		}

		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Creates a boilerplate fusectx.md file to initialize a project",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetDir string
		if len(args) > 0 {
			targetDir = args[0]
		} else {
			targetDir = "."
		}

		extends, _ := cmd.Flags().GetString("extends")
		includes, _ := cmd.Flags().GetStringSlice("includes")
		force, _ := cmd.Flags().GetBool("force")

		filePath := filepath.Join(targetDir, "fusectx.md")

		if !force {
			if _, err := os.Stat(filePath); err == nil {
				return fmt.Errorf("file %s already exists. Use --force to overwrite", filePath)
			}
		}

		var content strings.Builder
		content.WriteString("---\n")

		if extends != "" {
			content.WriteString(fmt.Sprintf("extends: %s\n", extends))
		}

		if len(includes) > 0 {
			content.WriteString("includes:\n")
			for _, include := range includes {
				content.WriteString(fmt.Sprintf("  - %s\n", include))
			}
		}

		content.WriteString("---\n\n")
		content.WriteString("# Project Context\n\n")
		content.WriteString("This is a fusectx configuration file.\n")

		err := os.MkdirAll(targetDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		err = os.WriteFile(filePath, []byte(content.String()), 0644)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", filePath, err)
		}

		fmt.Printf("Created %s\n", filePath)
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate <source_file>",
	Short: "Checks the entire dependency chain for errors without generating an output",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceFile := args[0]
		showChain, _ := cmd.Flags().GetBool("show-chain")
		quiet, _ := cmd.Flags().GetBool("quiet")

		err := resolver.ValidateChain(sourceFile)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
			}
			os.Exit(1)
		}

		if showChain {
			chain, err := resolver.GetDependencyChain(sourceFile, nil)
			if err != nil {
				return fmt.Errorf("failed to get dependency chain: %w", err)
			}
			fmt.Println("Dependency chain:")
			for i, file := range chain {
				fmt.Printf("%d. %s\n", i+1, file)
			}
		}

		if !quiet {
			fmt.Println("Validation successful")
		}
		return nil
	},
}

var buildAllCmd = &cobra.Command{
	Use:   "build-all [directory]",
	Short: "Scans a directory to find and build all leaf project configurations",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetDir string
		if len(args) > 0 {
			targetDir = args[0]
		} else {
			targetDir = "."
		}

		silent, _ := cmd.Flags().GetBool("silent")

		fusectxFiles, err := findFusectxFiles(targetDir)
		if err != nil {
			return fmt.Errorf("failed to find fusectx files: %w", err)
		}

		if len(fusectxFiles) == 0 {
			if !silent {
				fmt.Println("No fusectx.md files found")
			}
			return nil
		}

		for _, file := range fusectxFiles {
			if !silent {
				fmt.Printf("Building %s...\n", file)
			}

			content, err := resolver.Resolve(file, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to build %s: %v\n", file, err)
				continue
			}

			outputFile := strings.TrimSuffix(file, ".md") + ".ctx"
			err = os.WriteFile(outputFile, []byte(content), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write output for %s: %v\n", file, err)
				continue
			}

			if !silent {
				fmt.Printf("Output written to %s\n", outputFile)
			}
		}

		return nil
	},
}

func findFusectxFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == "fusectx.md" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func init() {
	buildCmd.Flags().StringP("output", "o", "", "Output file path")
	buildCmd.Flags().BoolP("silent", "s", false, "Suppress output messages")

	initCmd.Flags().StringP("extends", "e", "", "Set extends path")
	initCmd.Flags().StringSliceP("includes", "i", nil, "Set includes paths")
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing file")

	validateCmd.Flags().Bool("show-chain", false, "Show the dependency chain")
	validateCmd.Flags().BoolP("quiet", "q", false, "Suppress output messages")

	buildAllCmd.Flags().BoolP("silent", "s", false, "Suppress output messages")

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(buildAllCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
