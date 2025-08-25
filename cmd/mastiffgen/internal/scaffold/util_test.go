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

// TestIsEmptyDir_EmptyAndNonEmpty tests empty and non-empty dirs.
func TestIsEmptyDir_EmptyAndNonEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	empty, err := IsEmptyDir(tmpDir)
	assert.Nil(t, err)
	assert.True(t, empty)

	_ = os.MkdirAll(tmpDir+"/.git", 0750)
	empty, err = IsEmptyDir(tmpDir)
	assert.Nil(t, err)
	assert.True(t, empty)

	err = os.WriteFile(filepath.Join(tmpDir, "f.txt"), []byte("x"), 0600)
	assert.Nil(t, err)

	empty, err = IsEmptyDir(tmpDir)
	assert.Nil(t, err)
	assert.False(t, empty)
}

func TestIsEmptyDir_AbsPathError(t *testing.T) {
	orig := absPath
	defer func() { absPath = orig }()

	absPath = func(string) (string, error) {
		return "", errors.New("mock abs error")
	}

	empty, err := IsEmptyDir("/some/path")
	assert.Error(t, err)
	assert.False(t, empty)
}

func TestIsEmptyDir_OpenError(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "non-existent-dir")

	empty, err := IsEmptyDir(nonExistentPath)
	assert.Error(t, err)
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

func TestInsertBetweenMarkers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		content := `
		package core

		// MODULE_PACKAGE_START
		// MODULE_PACKAGE_END`
		result, err := insertBetweenMarkers(content, "// MODULE_PACKAGE_START", "// MODULE_PACKAGE_END", "packageLine")
		assert.Nil(t, err)
		assert.Contains(t, result, "packageLine")
	})
}

func TestInsertBetweenMarkers_StartMarkerMissing(t *testing.T) {
	content := "// END\nline\n"
	_, err := insertBetweenMarkers(content, "// START", "// END", "new line")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "markers")
}

func TestInsertBetweenMarkers_EndMarkerMissing(t *testing.T) {
	content := "// START\nline\n"
	_, err := insertBetweenMarkers(content, "// START", "// END", "new line")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "markers")
}

func TestInsertBetweenMarkers_EndBeforeStart(t *testing.T) {
	content := "// END\n...\n// START\n"
	_, err := insertBetweenMarkers(content, "// START", "// END", "new line")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "markers")
}

func TestInsertBetweenMarkers_Success(t *testing.T) {
	content := `
	// START
	// END`
	expected := `
	// START
	new line
	// END`

	result, err := insertBetweenMarkers(content, "// START", "// END", "new line")
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestInsertBetweenMarkers_HasLine(t *testing.T) {
	content := `
	// START
	has line
	// END`
	expected := `
	// START
	has line
	// END`

	result, err := insertBetweenMarkers(content, "// START", "// END", "has line")
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestInsertBetweenMarkers_AddNewLine(t *testing.T) {
	content := `
	// START
	old line
	// END`
	expected := `
	// START
	old line
	new line
	// END`

	result, err := insertBetweenMarkers(content, "// START", "// END", "new line")
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
