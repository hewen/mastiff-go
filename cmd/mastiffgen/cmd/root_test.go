// Package cmd contains the root command.
package cmd

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	assert.NotPanics(t, func() { Execute() })
}

func TestExecute_Error(t *testing.T) {
	if os.Getenv("MASTIFFGEN_TEST_EXECUTE") == "1" {
		Execute()
		return
	}

	cmd := exec.Command(os.Args[0], " erorr command") // #nosec
	cmd.Env = append(os.Environ(), "MASTIFFGEN_TEST_EXECUTE=1")

	_, err := cmd.CombinedOutput()
	assert.Error(t, err)
}
