// Package main contains the main function for generating project files.
package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed templates/**/*
var templatesFS embed.FS

var getCurrentUser = user.Current

// TemplateData holds the data used for rendering templates.
type TemplateData struct {
	PackageName     string
	ProjectName     string
	ModuleName      string
	TargetName      string // used for module creation like "user"
	TitleTargetName string
	LowerModuleName string
}

// main is the entry point for the project scaffolding tool.
func main() {
	rootCmd := &cobra.Command{
		Use:   "mastiffgen",
		Short: "Mastiff project code generator",
	}

	// init command
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project",
		RunE:  runInitCmd,
	}
	initCmd.Flags().String("package", "", "Go package name (required)")
	initCmd.Flags().String("project", "", "Project name (required)")
	initCmd.Flags().StringP("dir", "d", ".", "Target directory")
	_ = initCmd.MarkFlagRequired("package")
	_ = initCmd.MarkFlagRequired("project")

	// module command
	moduleCmd := &cobra.Command{
		Use:   "module [name]",
		Short: "Create a new module",
		Args:  cobra.ExactArgs(1),
		RunE:  runModuleCmd,
	}
	moduleCmd.Flags().String("package", "", "Go package name (required)")
	moduleCmd.Flags().StringP("dir", "d", ".", "Target directory")
	moduleCmd.Flags().Bool("http", false, "Generate HTTP method")
	moduleCmd.Flags().Bool("grpc", false, "Generate gRPC method(default)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(moduleCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runInitCmd(cmd *cobra.Command, _ []string) error {
	packageName, _ := cmd.Flags().GetString("package")
	projectName, _ := cmd.Flags().GetString("project")
	targetDir, _ := cmd.Flags().GetString("dir")
	targetDir = expandPath(targetDir)

	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}
	if empty, err := isEmptyDir(targetDir); err != nil {
		return fmt.Errorf("failed to check directory: %v", err)
	} else if !empty {
		return fmt.Errorf("target directory is not empty")
	}

	data := TemplateData{
		PackageName: packageName,
		ProjectName: projectName,
	}

	return generateTemplates("templates/init", targetDir, data)
}

func runModuleCmd(cmd *cobra.Command, args []string) error {
	moduleName := args[0]
	packageName, _ := cmd.Flags().GetString("package")
	targetDir, _ := cmd.Flags().GetString("dir")
	targetDir = expandPath(targetDir)

	if empty, err := isEmptyDir(targetDir); err != nil {
		return fmt.Errorf("failed to check directory: %v", err)
	} else if empty {
		return fmt.Errorf("target directory is empty")
	}

	caser := cases.Title(language.English)
	lowerModuleName := strings.ToLower(moduleName)
	data := TemplateData{
		PackageName:     packageName,
		ModuleName:      moduleName,
		TargetName:      moduleName,
		TitleTargetName: caser.String(moduleName),
		LowerModuleName: lowerModuleName,
	}

	coreGoPath := filepath.Join(targetDir, "core", "core.go")
	coreModuleDir := filepath.Join(targetDir, "core", strings.ToLower(moduleName))

	if err := generateTemplates("templates/module", coreModuleDir, data); err != nil {
		return err
	}
	routeLine := fmt.Sprintf("c.%s.Register%sRoutes(api)", caser.String(moduleName), caser.String(moduleName))
	if err := appendToCoreGoRoutes(coreGoPath, routeLine); err != nil {
		return err
	}

	fieldLine := fmt.Sprintf("%s *%s.%sModule", caser.String(moduleName), lowerModuleName, caser.String(moduleName))
	initLine := fmt.Sprintf("c.%s = &%s.%sModule{}", caser.String(moduleName), lowerModuleName, caser.String(moduleName))

	if err := appendToCoreGo(coreGoPath, fieldLine, initLine); err != nil {
		return err
	}

	packageLine := fmt.Sprintf(`"%s/core/%s"`, packageName, lowerModuleName)
	if err := appendToCorePackage(coreGoPath, packageLine); err != nil {
		return err
	}

	return nil
}

func expandPath(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}

	usr, err := getCurrentUser()
	if err == nil && usr.HomeDir != "" {
		return filepath.Join(usr.HomeDir, path[2:])
	}

	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, path[2:])
	}

	return filepath.Join(".", path[2:])
}

func appendToCoreGo(coreGoPath string, fieldLine, initLine string) error {
	data, err := os.ReadFile(coreGoPath) // #nosec
	if err != nil {
		return err
	}
	content := string(data)

	fieldStart := "// MODULE_FIELDS_START"
	fieldEnd := "// MODULE_FIELDS_END"
	content, err = insertBetweenMarkers(content, fieldStart, fieldEnd, fieldLine)
	if err != nil {
		return err
	}

	initStart := "// MODULE_INITS_START"
	initEnd := "// MODULE_INITS_END"
	content, err = insertBetweenMarkers(content, initStart, initEnd, initLine)
	if err != nil {
		return err
	}

	return os.WriteFile(coreGoPath, []byte(content), 0600)
}

func appendToCoreGoRoutes(coreGoPath string, routeLine string) error {
	data, err := os.ReadFile(coreGoPath) // #nosec
	if err != nil {
		return err
	}
	content := string(data)

	routeStart := "// MODULE_ROUTES_START"
	routeEnd := "// MODULE_ROUTES_END"
	content, err = insertBetweenMarkers(content, routeStart, routeEnd, routeLine)
	if err != nil {
		return err
	}
	return os.WriteFile(coreGoPath, []byte(content), 0600)
}

func appendToCorePackage(coreGoPath string, packageLine string) error {
	data, err := os.ReadFile(coreGoPath) // #nosec
	if err != nil {
		return err
	}
	content := string(data)

	packageStart := "// MODULE_PACKAGE_START"
	packageEnd := "// MODULE_PACKAGE_END"
	content, err = insertBetweenMarkers(content, packageStart, packageEnd, packageLine)
	if err != nil {
		return err
	}
	return os.WriteFile(coreGoPath, []byte(content), 0600)
}

func insertBetweenMarkers(content, start, end, line string) (string, error) {
	startIdx := strings.Index(content, start)
	endIdx := strings.Index(content, end)
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return "", fmt.Errorf("markers %q or %q not found or invalid", start, end)
	}

	before := content[:startIdx+len(start)]
	middle := content[startIdx+len(start) : endIdx]
	after := content[endIdx:]

	if strings.Contains(middle, line) {
		return content, nil
	}

	middle = strings.TrimRight(middle, "\n\t ")

	var newMiddle string
	if middle == "" {
		newMiddle = "\n\t" + line + "\n\t"
	} else {
		newMiddle = middle + "\n\t" + line + "\n\t"
	}
	return before + newMiddle + after, nil
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
