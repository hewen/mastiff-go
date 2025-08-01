package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestRun_Success verifies that the CLI runs successfully with valid flags.
func TestRunInitCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := &cobra.Command{}
	cmd.Flags().String("dir", tmpDir, "")
	cmd.Flags().String("module", "example.com/test", "")
	cmd.Flags().String("project", "testproj", "")

	err := runInitCmd(cmd, []string{})
	assert.NoError(t, err)

	expectedFile := filepath.Join(tmpDir, "README.md")
	info, statErr := os.Stat(expectedFile)
	assert.NoError(t, statErr)
	assert.False(t, info.IsDir())
}

// TestRun_MissingFlags ensures runInitCmd() fails when required flags are missing.
func TestRunInitCmd_MissingFlags(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("module", "", "")
	cmd.Flags().String("project", "", "")
	cmd.Flags().String("dir", ".", "")

	err := runInitCmd(cmd, []string{})
	assert.Error(t, err)
}

// TestRun_NotEmptyDir checks that runInitCmd() fails when target directory is not empty.
func TestRunInitCmd_NotEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "dummy.txt"), []byte("x"), 0600)
	assert.Nil(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("dir", tmpDir, "")
	cmd.Flags().String("module", "mod", "")
	cmd.Flags().String("project", "proj", "")

	err = runInitCmd(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not empty")
}

func TestRunInitCmd_MkdirFail(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("dir", string([]byte{0x00}), "")
	cmd.Flags().String("module", "mod", "")
	cmd.Flags().String("project", "proj", "")

	err := runInitCmd(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create target directory")
}

func TestRunInitCmd_IsEmptyDirFail(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notadir.txt")
	err := os.WriteFile(filePath, []byte("test"), 0600)
	assert.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("dir", filePath, "")
	cmd.Flags().String("module", "mod", "")
	cmd.Flags().String("project", "proj", "")

	err = runInitCmd(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mkdir")
}
