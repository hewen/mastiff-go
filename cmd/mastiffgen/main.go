// Package main contains the main function for generating project files.
package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*
var templatesFS embed.FS

// TemplateData holds the data used for rendering templates.
type TemplateData struct {
	ModuleName  string
	ProjectName string
}

// main is the entry point for the project scaffolding tool.
func main() {
	targetDir := flag.String("dir", ".", "target directory to generate project files")
	moduleName := flag.String("module", "", "go module name for import paths")
	projectName := flag.String("project", "", "project name (used in templates)")
	flag.Parse()

	if *moduleName == "" || *projectName == "" {
		fmt.Fprintln(os.Stderr, "please specify -module and -project")
		os.Exit(1)
	}

	// Create target directory if not exists
	if err := os.MkdirAll(*targetDir, 0750); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create target directory: %v\n", err)
		os.Exit(1)
	}

	// Check if the target dir is empty (required for init)
	empty, err := isEmptyDir(*targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to check directory: %v\n", err)
		os.Exit(1)
	}
	if !empty {
		fmt.Fprintln(os.Stderr, "target directory is not empty")
		os.Exit(1)
	}

	data := TemplateData{
		ModuleName:  *moduleName,
		ProjectName: *projectName,
	}

	err = generateTemplates("templates", *targetDir, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating project files: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Project scaffolding completed.")
}

// generateTemplates recursively walks template dir and renders files.
func generateTemplates(templateRoot, outputRoot string, data TemplateData) error {
	return fsWalk(templateRoot, func(path string, info os.FileInfo) error {
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		// Relative path in templates
		relPath, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return err
		}

		// Output file path: strip ".tmpl" and prefix with outputRoot
		targetPath := filepath.Join(outputRoot, strings.TrimSuffix(relPath, ".tmpl"))

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(targetPath), 0750); err != nil {
			return err
		}

		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
		if err != nil {
			return err
		}

		// Validate and ensure targetPath is inside outputRoot
		absTargetPath, err := filepath.Abs(targetPath)
		if err != nil {
			return err
		}
		absOutputRoot, err := filepath.Abs(outputRoot)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(absTargetPath, absOutputRoot) {
			return fmt.Errorf("invalid target path: %s is outside of %s", absTargetPath, absOutputRoot)
		}

		// #nosec G304 -- absTargetPath is safely validated above
		outFile, err := os.Create(absTargetPath)
		if err != nil {
			return err
		}
		defer func() {
			_ = outFile.Close()
		}()

		if err := tmpl.Execute(outFile, data); err != nil {
			return err
		}

		fmt.Printf("Generated %s\n", targetPath)
		return nil
	})
}

// fsWalk is like filepath.Walk but for embed.FS.
func fsWalk(root string, fn func(path string, info os.FileInfo) error) error {
	entries, err := templatesFS.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fullPath := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			err := fsWalk(fullPath, fn)
			if err != nil {
				return err
			}
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if err := fn(fullPath, info); err != nil {
			return err
		}
	}
	return nil
}

// isEmptyDir checks if directory is empty.
func isEmptyDir(dir string) (bool, error) {
	// Clean the path to remove any path traversal elements
	cleanDir := filepath.Clean(dir)
	// Ensure the path is absolute
	absDir, err := filepath.Abs(cleanDir)
	if err != nil {
		return false, err
	}

	// #nosec G304 -- absDir is cleaned and made absolute above
	f, err := os.Open(absDir)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.Readdir(1)
	if err == os.ErrNotExist || err == io.EOF {
		return true, nil
	}
	return false, err
}
