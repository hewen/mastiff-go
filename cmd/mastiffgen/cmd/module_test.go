package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunModuleCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()

	coreDir := filepath.Join(tmpDir, "core")
	err := os.MkdirAll(coreDir, 0750)
	assert.NoError(t, err)

	coreGoPath := filepath.Join(coreDir, "core.go")
	coreGoContent := `
package core

// MODULE_PACKAGE_START
// MODULE_PACKAGE_END

// MODULE_FIELDS_START
// MODULE_FIELDS_END

// MODULE_INITS_START
// MODULE_INITS_END

// MODULE_ROUTES_START
// MODULE_ROUTES_END
`
	err = os.WriteFile(coreGoPath, []byte(coreGoContent), 0600)
	assert.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("package", "github.com/example/project", "")
	cmd.Flags().String("dir", tmpDir, "")
	_ = cmd.ParseFlags([]string{"--package=github.com/example/project", "--dir=" + tmpDir})

	err = runModuleCmd(cmd, []string{"User"})
	assert.NoError(t, err)

	updatedContent, err := os.ReadFile(coreGoPath) // #nosec
	assert.NoError(t, err)

	result := string(updatedContent)
	assert.Contains(t, result, "c.User.RegisterUserRoutes(api)")
	assert.Contains(t, result, `User *user.UserModule`)
	assert.Contains(t, result, `c.User = &user.UserModule{}`)
	assert.Contains(t, result, `"github.com/example/project/core/user"`)

	expectedModuleFile := filepath.Join(tmpDir, "core", "user", "module.go")
	content, err := os.ReadFile(expectedModuleFile) // #nosec
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(content), "package user"))
}

func TestRunModuleCmd_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := &cobra.Command{}
	cmd.Flags().String("package", "github.com/example/project", "")
	cmd.Flags().String("dir", tmpDir, "")
	_ = cmd.ParseFlags([]string{"--package=github.com/example/project", "--dir=" + tmpDir})

	err := runModuleCmd(cmd, []string{"Test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target directory is empty")
}
