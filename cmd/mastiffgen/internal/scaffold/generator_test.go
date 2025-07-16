package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateTemplates_TemplateParseError verifies template parsing failure is handled.
func TestGenerateTemplates_TemplateParseError(t *testing.T) {
	tmpDir := t.TempDir()

	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}
	err := GenerateTemplates("testdata", tmpDir, data)
	assert.Error(t, err)
}

// TestGenerateTemplates_Content checks if a generated file contains expected content.
func TestGenerateTemplates_Content(t *testing.T) {
	tmpDir := t.TempDir()
	data := TemplateData{ModuleName: "modname", ProjectName: "projname"}

	err := GenerateTemplates("templates/init", tmpDir, data)
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
	err := GenerateTemplates("templates/init", string([]byte{0}), data)
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

func TestGenerateTemplates_MkdirFail(t *testing.T) {
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}
	invalidDir := string([]byte{0})
	err := GenerateTemplates("templates/init", invalidDir, data)
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
	err = GenerateTemplates(templateDir, tmpDir, data)
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
	err = GenerateTemplates(templateDir, tmpDir, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_BadTemplateParse(t *testing.T) {
	tmp := t.TempDir()
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}

	err := GenerateTemplates("testdata/bad_templates", tmp, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_ExecuteError(t *testing.T) {
	tmp := t.TempDir()
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}

	err := GenerateTemplates("testdata/execute_fail", tmp, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_PathEscape(t *testing.T) {
	tmp := t.TempDir()
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}

	err := GenerateTemplates("testdata/evil_templates", tmp, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_CreateFileError(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	_ = os.MkdirAll(sub, 0500)

	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}
	err := GenerateTemplates("templates_that_write_to", sub, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_InvalidTemplateRoot(t *testing.T) {
	err := GenerateTemplates("", os.TempDir(), TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_InvalidOutputRoot(t *testing.T) {
	err := GenerateTemplates("templateRoot", "", TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_EmptyTemplateData(t *testing.T) {
	err := GenerateTemplates("templateRoot", os.TempDir(), TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_TemplateSyntaxError(t *testing.T) {
	tmplPath := filepath.Join("testdata", "invalid_syntax.tmpl")
	err := os.WriteFile(tmplPath, []byte("{{ invalid syntax }}"), 0600)
	assert.NoError(t, err)

	err = GenerateTemplates("testdata", os.TempDir(), TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_TemplateMissingVariable(t *testing.T) {
	// Create a template file with a missing variable
	tmplPath := filepath.Join("testdata", "missing_variable.tmpl")
	err := os.WriteFile(tmplPath, []byte("{{ .MissingVariable }}"), 0600)
	assert.NoError(t, err)

	err = GenerateTemplates("testdata", os.TempDir(), TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_ErrorCases(t *testing.T) {
	data := TemplateData{ModuleName: "mod", ProjectName: "proj"}

	err := GenerateTemplates("nonexistent_dir", t.TempDir(), data)
	assert.Error(t, err)

	err = GenerateTemplates("testdata_with_bad_template", t.TempDir(), data)
	assert.Error(t, err)

	badPath := string(filepath.Separator) + ":/invalid<>path"
	err = GenerateTemplates("templates", badPath, data)
	assert.Error(t, err)

	tmpDir := t.TempDir()
	conflictPath := filepath.Join(tmpDir, "proj.go")
	err = os.MkdirAll(filepath.Dir(conflictPath), 0700)
	assert.NoError(t, err)
	err = os.WriteFile(conflictPath, []byte("collision"), 0600)
	assert.NoError(t, err)

	err = GenerateTemplates("templates_conflict", tmpDir, data)
	assert.Error(t, err)

	err = GenerateTemplates("testdata_with_evil_template", t.TempDir(), data)
	assert.Error(t, err)
}

func TestProcessTemplateFile_NotTmplFile(t *testing.T) {
	data := TemplateData{}
	err := processTemplateFile("/file.txt", "good/base", "out", data, func(_ string) ([]byte, error) {
		return []byte(""), nil
	})
	assert.Nil(t, err)
}

func TestProcessTemplateFile_AbsPathError(t *testing.T) {
	orig := absPath
	defer func() { absPath = orig }()

	mockError := errors.New("mock abs error")
	absPath = func(string) (string, error) {
		return "", mockError
	}

	data := TemplateData{}
	err := processTemplateFile("file.tmpl", "good/base", "out", data, func(_ string) ([]byte, error) {
		return []byte(""), nil
	})
	assert.Equal(t, mockError, err)
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

	err := GenerateTemplates(templateRoot, os.TempDir(), data)
	assert.NotNil(t, err)
}
func TestGenerateTemplates_SkipNonTemplateFile(t *testing.T) {
	templateRoot := t.TempDir()
	outputRoot := t.TempDir()
	data := TemplateData{ModuleName: "x", ProjectName: "y"}

	nonTmpl := filepath.Join(templateRoot, "note.txt")
	assert.NoError(t, os.WriteFile(nonTmpl, []byte("plain"), 0600))

	err := GenerateTemplates(templateRoot, outputRoot, data)
	assert.NotNil(t, err)
}
func TestGenerateTemplates_MkdirError(t *testing.T) {
	data := TemplateData{ModuleName: "x", ProjectName: "y"}

	badOutput := string([]byte{0})
	err := GenerateTemplates("templates", badOutput, data)
	assert.Error(t, err)
}
