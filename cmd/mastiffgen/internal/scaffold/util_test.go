package scaffold

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsEmptyDir_OpenError simulates failure to open directory by passing a file path.
func TestIsEmptyDir_OpenError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file.txt")
	err := os.WriteFile(filePath, []byte("x"), 0600)
	assert.Nil(t, err)

	empty, err := IsEmptyDir(filePath)
	assert.Error(t, err)
	assert.False(t, empty)
}

// TestIsEmptyDir_EmptyAndNonEmpty tests empty and non-empty dirs.
func TestIsEmptyDir_EmptyAndNonEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	empty, err := IsEmptyDir(tmpDir)
	assert.Nil(t, err)
	assert.True(t, empty)

	err = os.WriteFile(filepath.Join(tmpDir, "f.txt"), []byte("x"), 0600)
	assert.Nil(t, err)

	empty, err = IsEmptyDir(tmpDir)
	assert.Nil(t, err)
	assert.False(t, empty)
}

func TestExpandPath_WithTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	input := "~/myproject"
	expected := filepath.Join(home, "myproject")

	result := ExpandPath(input)
	assert.Equal(t, expected, result)
}

func TestExpandPath_Relative(t *testing.T) {
	input := "./subdir"
	result := ExpandPath(input)
	assert.True(t, strings.HasSuffix(result, "subdir"))
}

func TestExpandPath_UserCurrentError(t *testing.T) {
	orig := getCurrentUser
	defer func() { getCurrentUser = orig }()

	getCurrentUser = func() (*user.User, error) {
		return nil, errors.New("mocked user.Current error")
	}

	result := ExpandPath("~/project")
	assert.True(t, strings.HasSuffix(result, "project"))
}

func TestAppendToCoreGo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tmpDir := t.TempDir()
		coreGoPath := filepath.Join(tmpDir, "core.go")
		err := os.WriteFile(coreGoPath, []byte(`
			package core

			// MODULE_FIELDS_START
			// MODULE_FIELDS_END

			// MODULE_INITS_START
			// MODULE_INITS_END
		`), 0600)
		assert.Nil(t, err)

		err = AppendToCoreGo(coreGoPath, "fieldLine", "initLine")
		assert.Nil(t, err)

		content, err := os.ReadFile(coreGoPath) // #nosec
		assert.Nil(t, err)
		assert.Contains(t, string(content), "fieldLine")
		assert.Contains(t, string(content), "initLine")
	})
}

func TestAppendToCoreGoRoutes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tmpDir := t.TempDir()
		coreGoPath := filepath.Join(tmpDir, "core.go")
		err := os.WriteFile(coreGoPath, []byte(`
			package core

			// MODULE_ROUTES_START
			// MODULE_ROUTES_END
		`), 0600)
		assert.Nil(t, err)
		err = AppendToCoreGoRoutes(coreGoPath, "routeLine")
		assert.Nil(t, err)

		content, err := os.ReadFile(coreGoPath) // #nosec
		assert.Nil(t, err)
		assert.Contains(t, string(content), "routeLine")
	})
}

func TestAppendToCorePackage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tmpDir := t.TempDir()
		coreGoPath := filepath.Join(tmpDir, "core.go")
		err := os.WriteFile(coreGoPath, []byte(`
			package core

			// MODULE_PACKAGE_START
			// MODULE_PACKAGE_END
		`), 0600)
		assert.Nil(t, err)
		err = AppendToCorePackage(coreGoPath, "packageLine")
		assert.Nil(t, err)

		content, err := os.ReadFile(coreGoPath) // #nosec
		assert.Nil(t, err)
		assert.Contains(t, string(content), "packageLine")
	})
}

func TestInsertBetweenMarkers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		content := `
		package core

		// MODULE_PACKAGE_START
		// MODULE_PACKAGE_END
		`
		result, err := insertBetweenMarkers(content, "// MODULE_PACKAGE_START", "// MODULE_PACKAGE_END", "packageLine")
		assert.Nil(t, err)
		assert.Contains(t, result, "packageLine")
	})
}
