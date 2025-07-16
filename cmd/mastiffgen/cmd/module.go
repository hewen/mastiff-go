// Package cmd contains the module command.
package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/hewen/mastiff-go/cmd/mastiffgen/internal/scaffold"
)

func init() {
	RootCmd.AddCommand(ModuleCmd())
}

// ModuleCmd returns the module command.
func ModuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module [name]",
		Short: "Create a new module",
		Args:  cobra.ExactArgs(1),
		RunE:  runModuleCmd,
	}
	cmd.Flags().String("package", "", "Go package name (required)")
	cmd.Flags().StringP("dir", "d", ".", "Target directory")
	cmd.Flags().Bool("http", false, "Generate HTTP method")
	cmd.Flags().Bool("grpc", false, "Generate gRPC method (default)")
	return cmd
}

func runModuleCmd(cmd *cobra.Command, args []string) error {
	moduleName := args[0]
	packageName, _ := cmd.Flags().GetString("package")
	targetDir, _ := cmd.Flags().GetString("dir")
	targetDir = scaffold.ExpandPath(targetDir)

	if empty, err := scaffold.IsEmptyDir(targetDir); err != nil {
		return fmt.Errorf("failed to check directory: %v", err)
	} else if empty {
		return fmt.Errorf("target directory is empty")
	}

	caser := cases.Title(language.English)
	lowerModuleName := strings.ToLower(moduleName)
	data := scaffold.TemplateData{
		PackageName:     packageName,
		ModuleName:      moduleName,
		TargetName:      moduleName,
		TitleTargetName: caser.String(moduleName),
		LowerModuleName: lowerModuleName,
	}

	coreGoPath := filepath.Join(targetDir, "core", "core.go")
	coreModuleDir := filepath.Join(targetDir, "core", strings.ToLower(moduleName))
	_ = scaffold.GenerateTemplates("templates/module", coreModuleDir, data)

	routeLine := fmt.Sprintf("c.%s.Register%sRoutes(api)", caser.String(moduleName), caser.String(moduleName))
	_ = scaffold.AppendToCoreGoRoutes(coreGoPath, routeLine)

	fieldLine := fmt.Sprintf("%s *%s.%sModule", caser.String(moduleName), lowerModuleName, caser.String(moduleName))
	initLine := fmt.Sprintf("c.%s = &%s.%sModule{}", caser.String(moduleName), lowerModuleName, caser.String(moduleName))
	_ = scaffold.AppendToCoreGo(coreGoPath, fieldLine, initLine)

	packageLine := fmt.Sprintf(`"%s/core/%s"`, packageName, lowerModuleName)
	_ = scaffold.AppendToCorePackage(coreGoPath, packageLine)

	return nil
}
