// Package scaffold provides functions for generating project and module code.
package scaffold

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/**/*
var templatesFS embed.FS

// FS represents the embedded filesystem interface.
type FS interface {
	// ReadFile reads the named file and returns its contents.
	ReadFile(name string) ([]byte, error)
	// ReadDir reads the named directory and returns its contents.
	ReadDir(name string) ([]os.DirEntry, error)
}

// TemplateData holds the data used for rendering templates.
type TemplateData struct {
	PackageName     string
	ProjectName     string
	ModuleName      string
	TargetName      string // used for module creation like "user"
	TitleTargetName string
	LowerModuleName string
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

	if err = os.MkdirAll(filepath.Dir(targetPath), 0750); err != nil {
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

	return nil
}

// GenerateTemplates walks the template directory and processes each .tmpl file.
func GenerateTemplates(templateRoot, outputRoot string, data TemplateData) error {
	return fsWalk(templateRoot, func(path string, _ os.FileInfo) error {
		return processTemplateFile(path, templateRoot, outputRoot, data, templatesFS.ReadFile)
	})
}

// fsPath converts a path to a format suitable for embed.FS.
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
