package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const manifestPath = "practices/01-from-evidence-to-action/practice.json"

type commandSpec struct {
	Run            string `json:"run"`
	ExpectedExit   int    `json:"expectedExit"`
	ExpectedOutput string `json:"expectedOutput"`
}

type manifest struct {
	Type     string                 `json:"type"`
	Commands map[string]commandSpec `json:"commands"`
}

func main() {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fatalf("read manifest: %v", err)
	}

	var candidate manifest
	if err := json.Unmarshal(data, &candidate); err != nil {
		fatalf("parse manifest: %v", err)
	}
	if candidate.Type != "practice" {
		fatalf("manifest type = %q, want practice", candidate.Type)
	}

	for _, name := range []string{"failure", "solution", "boundary", "verifyReader", "check"} {
		if _, ok := candidate.Commands[name]; !ok {
			fatalf("manifest command %q is missing", name)
		}
	}

	for _, name := range []string{"failure", "solution", "boundary"} {
		if err := runCommand(name, candidate.Commands[name]); err != nil {
			fatalf("%v", err)
		}
		fmt.Printf("manifest command %s: PASS\n", name)
	}
}

func runCommand(name string, spec commandSpec) error {
	if spec.Run == "" {
		return fmt.Errorf("manifest command %q has no run string", name)
	}

	command := exec.Command("sh", "-c", spec.Run)
	output, err := command.CombinedOutput()
	exitCode := 0
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			return fmt.Errorf("manifest command %q failed to start: %w", name, err)
		}
		exitCode = exitError.ExitCode()
	}
	if exitCode != spec.ExpectedExit {
		return fmt.Errorf("manifest command %q exit = %d, want %d\n%s", name, exitCode, spec.ExpectedExit, output)
	}
	if spec.ExpectedOutput != "" && !strings.Contains(string(output), spec.ExpectedOutput) {
		return fmt.Errorf("manifest command %q did not emit %q\n%s", name, spec.ExpectedOutput, output)
	}
	return nil
}

func fatalf(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "Practice 01 manifest check: "+format+"\n", arguments...)
	os.Exit(1)
}
