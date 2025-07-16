// Package cmd contains the init command.
package cmd

import (
	"fmt"
	"os"

	"github.com/hewen/mastiff-go/cmd/mastiffgen/internal/scaffold"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(InitCmd())
}

// InitCmd returns the init command.
func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project",
		RunE:  runInitCmd,
	}
	cmd.Flags().String("package", "", "Go package name (required)")
	cmd.Flags().String("project", "", "Project name (required)")
	cmd.Flags().StringP("dir", "d", ".", "Target directory")
	_ = cmd.MarkFlagRequired("package")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func runInitCmd(cmd *cobra.Command, _ []string) error {
	packageName, _ := cmd.Flags().GetString("package")
	projectName, _ := cmd.Flags().GetString("project")
	targetDir, _ := cmd.Flags().GetString("dir")
	targetDir = scaffold.ExpandPath(targetDir)

	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}
	if empty, err := scaffold.IsEmptyDir(targetDir); err != nil {
		return fmt.Errorf("failed to check directory: %v", err)
	} else if !empty {
		return fmt.Errorf("target directory is not empty")
	}

	data := scaffold.TemplateData{
		PackageName: packageName,
		ProjectName: projectName,
	}

	return scaffold.GenerateTemplates("templates/init", targetDir, data)
}
