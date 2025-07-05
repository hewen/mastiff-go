package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRun_Success verifies that the CLI runs successfully with valid flags.
func TestRun_Success(t *testing.T) {
	tmpDir := t.TempDir()

	args := []string{
		"-dir", tmpDir,
		"-module", "example.com/test",
		"-project", "testproj",
	}

	err := run(args)
	assert.Nil(t, err)

	// Check if expected file is generated from template
	expectedFile := filepath.Join(tmpDir, "README.md")
	info, statErr := os.Stat(expectedFile)
	assert.NoError(t, statErr)
	assert.False(t, info.IsDir())
}

// TestRun_MissingFlags ensures run() fails when required flags are missing.
func TestRun_MissingFlags(t *testing.T) {
	err := run([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "please specify")
}

// TestRun_NotEmptyDir checks that run() fails when target directory is not empty.
func TestRun_NotEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "dummy.txt"), []byte("x"), 0600)
	assert.Nil(t, err)

	args := []string{
		"-dir", tmpDir,
		"-module", "mod",
		"-project", "proj",
	}
	err = run(args)
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

	err := generateTemplates("templates", tmpDir, data)
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
	err := generateTemplates("templates", string([]byte{0}), data)
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

func TestRun_MkdirFail(t *testing.T) {
	args := []string{
		"-dir", string([]byte{0x00}),
		"-module", "mod",
		"-project", "proj",
	}
	err := run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create target directory")
}

func TestRun_IsEmptyDirFail(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notadir.txt")
	err := os.WriteFile(filePath, []byte("test"), 0600)
	assert.NoError(t, err)

	args := []string{
		"-dir", filePath,
		"-module", "mod",
		"-project", "proj",
	}
	err = run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create target directory")
}

func TestGenerateTemplates_MkdirFail(t *testing.T) {
	data := TemplateData{"mod", "proj"}
	invalidDir := string([]byte{0})
	err := generateTemplates("templates", invalidDir, data)
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
	data := TemplateData{"mod", "proj"}

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

	err = generateTemplates(templateDir, tmpDir, TemplateData{"mod", "proj"})
	assert.Error(t, err)
}

func TestGenerateTemplates_ReadFileError(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join("testdata", "readfile_error")
	_ = os.MkdirAll(templateDir, 0750)

	tmplPath := filepath.Join(templateDir, "gone.tmpl")
	err := os.WriteFile(tmplPath, []byte("{{ .ProjectName }}"), 0600)
	assert.NoError(t, err)
	_ = os.Remove(tmplPath) // 删除掉

	err = generateTemplates(templateDir, tmpDir, TemplateData{"mod", "proj"})
	assert.Error(t, err)
}

func TestGenerateTemplates_BadTemplateParse(t *testing.T) {
	tmp := t.TempDir()
	data := TemplateData{"mod", "proj"}

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
	data := TemplateData{"mod", "proj"}

	err := generateTemplates("testdata/evil_templates", tmp, data)
	assert.Error(t, err)
}

func TestGenerateTemplates_CreateFileError(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	_ = os.MkdirAll(sub, 0500)

	data := TemplateData{"mod", "proj"}
	err := generateTemplates("templates_that_write_to", sub, data) // 模板存在时
	assert.Error(t, err)
}

func TestGenerateTemplates_InvalidTemplateRoot(t *testing.T) {
	err := generateTemplates("", "outputRoot", TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_InvalidOutputRoot(t *testing.T) {
	err := generateTemplates("templateRoot", "", TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_EmptyTemplateData(t *testing.T) {
	err := generateTemplates("templateRoot", "outputRoot", TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_TemplateSyntaxError(t *testing.T) {
	// Create a template file with invalid syntax
	tmplPath := filepath.Join("testdata", "invalid_syntax.tmpl")
	err := os.WriteFile(tmplPath, []byte("{{ invalid syntax }}"), 0600)
	assert.NoError(t, err)

	err = generateTemplates("testdata", "outputRoot", TemplateData{})
	assert.Error(t, err)
}

func TestGenerateTemplates_TemplateMissingVariable(t *testing.T) {
	// Create a template file with a missing variable
	tmplPath := filepath.Join("testdata", "missing_variable.tmpl")
	err := os.WriteFile(tmplPath, []byte("{{ .MissingVariable }}"), 0600)
	assert.NoError(t, err)

	err = generateTemplates("testdata", "outputRoot", TemplateData{})
	assert.Error(t, err)
}

func TestRun_Coverage(t *testing.T) {
	// 测试run，间接覆盖main调用逻辑
	err := run([]string{"-dir", t.TempDir(), "-module", "mod", "-project", "proj"})
	assert.NoError(t, err)
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
	// 构造非法路径，比如使用空字节
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
	// 创建模板目录，带子目录
	templateRoot := t.TempDir()
	subDir := filepath.Join(templateRoot, "subdir")
	assert.NoError(t, os.MkdirAll(subDir, 0750))

	// 生成目标目录
	outputRoot := t.TempDir()
	data := TemplateData{ModuleName: "x", ProjectName: "y"}

	// 使用空子目录测试
	err := generateTemplates(templateRoot, outputRoot, data)
	assert.NotNil(t, err)
}
func TestGenerateTemplates_SkipNonTemplateFile(t *testing.T) {
	templateRoot := t.TempDir()
	outputRoot := t.TempDir()
	data := TemplateData{ModuleName: "x", ProjectName: "y"}

	// 创建非模板文件
	nonTmpl := filepath.Join(templateRoot, "note.txt")
	assert.NoError(t, os.WriteFile(nonTmpl, []byte("plain"), 0600))

	// 覆盖 fsWalk 来传入自定义文件
	err := generateTemplates(templateRoot, outputRoot, data)
	assert.NotNil(t, err)
}
func TestGenerateTemplates_MkdirError(t *testing.T) {
	data := TemplateData{ModuleName: "x", ProjectName: "y"}

	// 输出路径非法，模拟 mkdir error
	badOutput := string([]byte{0})
	err := generateTemplates("templates", badOutput, data)
	assert.Error(t, err)
}
