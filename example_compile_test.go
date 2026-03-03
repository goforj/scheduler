package scheduler

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestExamplesBuild(t *testing.T) {
	examplesDir := "examples"

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("cannot read examples directory: %v", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		// CAPTURE LOOP VARS
		name := e.Name()
		path := filepath.Join(examplesDir, name)

		t.Run(name, func(t *testing.T) {
			t.Parallel() // 🔑 enable concurrency

			if err := buildExample(path); err != nil {
				t.Fatalf("example %q failed to build:\n%s", name, err)
			}
		})
	}
}

func buildExample(exampleDir string) error {
	cmd := exec.Command(
		"go", "build",
		"-o", os.DevNull,
		"./"+exampleDir,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(stderr.String())
	}

	return nil
}
