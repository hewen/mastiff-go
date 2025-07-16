package scaffold

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// getCurrentUser returns the current user.
var getCurrentUser = user.Current

// ExpandPath expands a path that starts with ~/ to the full path.
func ExpandPath(path string) string {
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

// AppendToCoreGo appends lines to core.go.
func AppendToCoreGo(coreGoPath string, fieldLine, initLine string) error {
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

// AppendToCoreGoRoutes appends lines to core.go.
func AppendToCoreGoRoutes(coreGoPath string, routeLine string) error {
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

// AppendToCorePackage appends lines to core.go.
func AppendToCorePackage(coreGoPath string, packageLine string) error {
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

// insertBetweenMarkers inserts a line between markers.
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

// IsEmptyDir checks if directory is empty.
func IsEmptyDir(dir string) (bool, error) {
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
