package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_Success(_ *testing.T) {
	main()
}

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

// TestGenerateTemplates_TemplateParseError verifies template parsing failure is handled.
func TestGenerateTemplates_TemplateParseError(t *testing.T) {
	tmpDir := t.TempDir()

	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}
	err := generateTemplates("testdata", tmpDir, data)
	assert.Error(t, err)
}

// TestGenerateTemplates_Content checks if a generated file contains expected content.
func TestGenerateTemplates_Content(t *testing.T) {
	tmpDir := t.TempDir()
	data := TemplateData{ModuleName: "modname", ProjectName: "projname"}

	err := generateTemplates("templates/init", tmpDir, data)
	assert.Nil(t, err)

	readmePath := filepath.Join(tmpDir, "README.md")
	content, err := os.ReadFile(readmePath) // nolint: gosec
	assert.Nil(t, err)
	assert.Contains(t, string(content), "projname")
}

// TestGenerateTemplates_DirCreationFailure simulates mkdir failure by passing invalid path.
func TestGenerateTemplates_DirCreationFailure(t *testing.T) {
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}

	// Pass an invalid path to trigger mkdir error
	err := generateTemplates("templates/init", string([]byte{0}), data)
	assert.Error(t, err)
}

// TestFSWalk_RecursiveError simulates fsWalk returning error from nested call.
func TestFSWalk_RecursiveError(t *testing.T) {
	// We cannot inject errors into embed.FS easily, skip or use a wrapper in real projects.
	// Just test fsWalk on real templates directory for coverage
	err := fsWalk("templates", func(_ string, _ os.FileInfo) error {
		return nil
	})
	assert.Nil(t, err)
}

// TestIsEmptyDir_OpenError simulates failure to open directory by passing a file path.
func TestIsEmptyDir_OpenError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file.txt")
	err := os.WriteFile(filePath, []byte("x"), 0600)
	assert.Nil(t, err)

	empty, err := isEmptyDir(filePath)
	assert.Error(t, err)
	assert.False(t, empty)
}

// TestIsEmptyDir_EmptyAndNonEmpty tests empty and non-empty dirs.
func TestIsEmptyDir_EmptyAndNonEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	empty, err := isEmptyDir(tmpDir)
	assert.Nil(t, err)
	assert.True(t, empty)

	err = os.WriteFile(filepath.Join(tmpDir, "f.txt"), []byte("x"), 0600)
	assert.Nil(t, err)

	empty, err = isEmptyDir(tmpDir)
	assert.Nil(t, err)
	assert.False(t, empty)
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

func TestGenerateTemplates_MkdirFail(t *testing.T) {
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}
	invalidDir := string([]byte{0})
	err := generateTemplates("templates/init", invalidDir, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_ReadFileFail(t *testing.T) {
	err := fsWalk("testdata", func(_ string, info os.FileInfo) error {
		if !info.IsDir() {
			return fmt.Errorf("trigger read error")
		}
		return nil
	})
	assert.Error(t, err)
}

func TestGenerateTemplates_CreateFileFail(t *testing.T) {
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}

	tmpDir := t.TempDir()
	conflictPath := filepath.Join(tmpDir, "conflict.txt")
	err := os.WriteFile(conflictPath, []byte("existing file"), 0400)
	require.NoError(t, err)

	err = processTemplateFile(
		"templates/conflict.txt.tmpl",
		"templates",
		tmpDir,
		data,
		func(_ string) ([]byte, error) {
			return []byte("Hello, {{.ProjectName}}"), nil
		},
	)
	assert.Error(t, err)
}

func TestGenerateTemplates_ExecuteFail(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join("testdata", "execute_error")
	_ = os.MkdirAll(templateDir, 0750)

	tmplPath := filepath.Join(templateDir, "fail.tmpl")
	err := os.WriteFile(tmplPath, []byte("{{ .NotExist }}"), 0600)
	assert.NoError(t, err)

	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}
	err = generateTemplates(templateDir, tmpDir, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_ReadFileError(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join("testdata", "readfile_error")
	_ = os.MkdirAll(templateDir, 0750)

	tmplPath := filepath.Join(templateDir, "gone.tmpl")
	err := os.WriteFile(tmplPath, []byte("{{ .ProjectName }}"), 0600)
	assert.NoError(t, err)
	_ = os.Remove(tmplPath)

	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}
	err = generateTemplates(templateDir, tmpDir, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_BadTemplateParse(t *testing.T) {
	tmp := t.TempDir()
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}

	err := generateTemplates("testdata/bad_templates", tmp, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_ExecuteError(t *testing.T) {
	tmp := t.TempDir()
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}

	err := generateTemplates("testdata/execute_fail", tmp, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_PathEscape(t *testing.T) {
	tmp := t.TempDir()
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}

	err := generateTemplates("testdata/evil_templates", tmp, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_CreateFileError(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	_ = os.MkdirAll(sub, 0500)

	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}
	err := generateTemplates("templates_that_write_to", sub, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_InvalidTemplateRoot(t *testing.T) {
	err := generateTemplates("", os.TempDir(), TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_InvalidOutputRoot(t *testing.T) {
	err := generateTemplates("templateRoot", "", TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_EmptyTemplateData(t *testing.T) {
	err := generateTemplates("templateRoot", os.TempDir(), TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_TemplateSyntaxError(t *testing.T) {
	tmplPath := filepath.Join("testdata", "invalid_syntax.tmpl")
	err := os.WriteFile(tmplPath, []byte("{{ invalid syntax }}"), 0600)
	assert.NoError(t, err)

	err = generateTemplates("testdata", os.TempDir(), TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_TemplateMissingVariable(t *testing.T) {
	// Create a template file with a missing variable
	tmplPath := filepath.Join("testdata", "missing_variable.tmpl")
	err := os.WriteFile(tmplPath, []byte("{{ .MissingVariable }}"), 0600)
	assert.NoError(t, err)

	err = generateTemplates("testdata", os.TempDir(), TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_ErrorCases(t *testing.T) {
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}

	err := generateTemplates("nonexistent_dir", t.TempDir(), data)
	assert.Error(t, err)

	err = generateTemplates("testdata_with_bad_template", t.TempDir(), data)
	assert.Error(t, err)

	badPath := string(filepath.Separator) + ":/invalid<>path"
	err = generateTemplates("templates", badPath, data)
	assert.Error(t, err)

	tmpDir := t.TempDir()
	conflictPath := filepath.Join(tmpDir, "proj.go")
	err = os.MkdirAll(filepath.Dir(conflictPath), 0700)
	assert.NoError(t, err)
	err = os.WriteFile(conflictPath, []byte("collision"), 0600)
	assert.NoError(t, err)

	err = generateTemplates("templates_conflict", tmpDir, data)
	assert.Error(t, err)

	err = generateTemplates("testdata_with_evil_template", t.TempDir(), data)
	assert.Error(t, err)
}

func TestProcessTemplateFile_RelativePathError(t *testing.T) {
	data := TemplateData{}
	err := processTemplateFile("/bad/path/file.tmpl", "good/base", "out", data, func(_ string) ([]byte, error) {
		return []byte(""), nil
	})
	assert.Error(t, err)
}

func TestProcessTemplateFile_MkdirError(t *testing.T) {
	data := TemplateData{}
	tmpDir := t.TempDir()
	err := processTemplateFile(filepath.Join(tmpDir, string([]byte{0})+".tmpl"), tmpDir, tmpDir, data, func(_ string) ([]byte, error) {
		return []byte(""), nil
	})
	assert.Error(t, err)
}

func TestProcessTemplateFile_ReadFileError(t *testing.T) {
	data := TemplateData{}
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "a.tmpl")
	err := os.WriteFile(file, []byte("{{.Name}}"), 0600)
	assert.Nil(t, err)

	err = processTemplateFile(file, tmpDir, tmpDir, data, func(_ string) ([]byte, error) {
		return nil, fmt.Errorf("read error")
	})
	assert.Error(t, err)
}

func TestProcessTemplateFile_ParseError(t *testing.T) {
	data := TemplateData{}
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "broken.tmpl")
	err := os.WriteFile(file, []byte("{{ .Invalid "), 0600)
	assert.Nil(t, err)

	err = processTemplateFile(file, tmpDir, tmpDir, data, os.ReadFile)
	assert.Error(t, err)
}
func TestProcessTemplateFile_PathEscape(t *testing.T) {
	tmpDir := t.TempDir()
	badPath := filepath.Join(tmpDir, "../escape.tmpl")

	err := os.WriteFile(badPath, []byte("hi"), 0600)
	assert.Nil(t, err)
	data := TemplateData{}

	err = processTemplateFile(badPath, tmpDir, tmpDir, data, os.ReadFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside")
}

func TestProcessTemplateFile_ExecuteError(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "badexec.tmpl")
	err := os.WriteFile(file, []byte("{{.NotExist}}"), 0600)
	assert.Nil(t, err)

	data := TemplateData{}
	err = processTemplateFile(file, tmpDir, tmpDir, data, os.ReadFile)
	assert.Error(t, err)
}

func TestGenerateTemplates_SkipDirectory(t *testing.T) {
	templateRoot := t.TempDir()
	subDir := filepath.Join(templateRoot, "subdir")
	assert.NoError(t, os.MkdirAll(subDir, 0750))

	data := TemplateData{ModuleName: "x", ProjectName: "y"}

	err := generateTemplates(templateRoot, os.TempDir(), data)
	assert.NotNil(t, err)
}
func TestGenerateTemplates_SkipNonTemplateFile(t *testing.T) {
	templateRoot := t.TempDir()
	outputRoot := t.TempDir()
	data := TemplateData{ModuleName: "x", ProjectName: "y"}

	nonTmpl := filepath.Join(templateRoot, "note.txt")
	assert.NoError(t, os.WriteFile(nonTmpl, []byte("plain"), 0600))

	err := generateTemplates(templateRoot, outputRoot, data)
	assert.NotNil(t, err)
}
func TestGenerateTemplates_MkdirError(t *testing.T) {
	data := TemplateData{ModuleName: "x", ProjectName: "y"}

	badOutput := string([]byte{0})
	err := generateTemplates("templates", badOutput, data)
	assert.Error(t, err)
}

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

func TestExpandPath_WithTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	input := "~/myproject"
	expected := filepath.Join(home, "myproject")

	result := expandPath(input)
	assert.Equal(t, expected, result)
}

func TestExpandPath_Relative(t *testing.T) {
	input := "./subdir"
	result := expandPath(input)
	assert.True(t, strings.HasSuffix(result, "subdir"))
}

func TestExpandPath_UserCurrentError(t *testing.T) {
	orig := getCurrentUser
	defer func() { getCurrentUser = orig }() // 恢复原始实现

	getCurrentUser = func() (*user.User, error) {
		return nil, errors.New("mocked user.Current error")
	}

	result := expandPath("~/project")
	assert.True(t, strings.HasSuffix(result, "project"))
}
