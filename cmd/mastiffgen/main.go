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
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("mastiffgen", flag.ContinueOnError)
	targetDir := flags.String("dir", ".", "target directory to generate project files")
	moduleName := flags.String("module", "", "go module name for import paths")
	projectName := flags.String("project", "", "project name (used in templates)")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *moduleName == "" || *projectName == "" {
		return fmt.Errorf("please specify -module and -project")
	}

	if err := os.MkdirAll(*targetDir, 0750); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	empty, err := isEmptyDir(*targetDir)
	if err != nil {
		return fmt.Errorf("failed to check directory: %v", err)
	}
	if !empty {
		return fmt.Errorf("target directory is not empty")
	}

	data := TemplateData{
		ModuleName:  *moduleName,
		ProjectName: *projectName,
	}

	if err := generateTemplates("templates", *targetDir, data); err != nil {
		return fmt.Errorf("error generating project files: %v", err)
	}

	return nil
}

// ReadFileFunc defines a function type for reading files.
type ReadFileFunc func(name string) ([]byte, error)

// processTemplateFile processes a single template file, rendering it with the provided data.
func processTemplateFile(path string, templateRoot string, outputRoot string, data TemplateData, readFile ReadFileFunc) error {
	if !strings.HasSuffix(path, ".tmpl") {
		return nil
	}

	relPath, err := filepath.Rel(templateRoot, path)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(outputRoot, strings.TrimSuffix(relPath, ".tmpl"))

	if err := os.MkdirAll(filepath.Dir(targetPath), 0750); err != nil {
		return err
	}

	content, err := readFile(fsPath(path))
	if err != nil {
		return err
	}

	tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
	if err != nil {
		return err
	}

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

	outFile, err := os.Create(absTargetPath) // nolint: gosec
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
}

// generateTemplates walks the template directory and processes each .tmpl file.
func generateTemplates(templateRoot, outputRoot string, data TemplateData) error {
	return fsWalk(templateRoot, func(path string, _ os.FileInfo) error {
		return processTemplateFile(path, templateRoot, outputRoot, data, templatesFS.ReadFile)
	})
}

func fsPath(path string) string {
	return strings.ReplaceAll(path, `\`, `/`)
}

// fsWalk is like filepath.Walk but for embed.FS.
func fsWalk(root string, fn func(path string, info os.FileInfo) error) error {
	entries, err := templatesFS.ReadDir(fsPath(root))
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
