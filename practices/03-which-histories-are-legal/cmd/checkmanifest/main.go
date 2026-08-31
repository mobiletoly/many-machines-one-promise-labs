package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const manifestPath = "practices/03-which-histories-are-legal/practice.json"

type commandSpec struct {
	Run            string `json:"run"`
	ExpectedExit   *int   `json:"expectedExit"`
	ExpectedOutput string `json:"expectedOutput"`
}

type manifest struct {
	Number                  int                    `json:"number"`
	ID                      string                 `json:"id"`
	Type                    string                 `json:"type"`
	Title                   string                 `json:"title"`
	Capabilities            []string               `json:"capabilities"`
	RecommendedAfterChapter int                    `json:"recommendedAfterChapter"`
	Artifacts               map[string]string      `json:"artifacts"`
	Commands                map[string]commandSpec `json:"commands"`
}

func main() {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fatalf("read manifest: %v", err)
	}
	candidate, err := decodeManifest(data)
	if err != nil {
		fatalf("parse manifest: %v", err)
	}
	if err := validateManifest(candidate); err != nil {
		fatalf("%v", err)
	}
	for _, name := range []string{"failure", "solution", "boundary"} {
		if err := runCommand(name, candidate.Commands[name]); err != nil {
			fatalf("%v", err)
		}
		fmt.Printf("manifest command %s: PASS\n", name)
	}
}

func decodeManifest(data []byte) (manifest, error) {
	var candidate manifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&candidate); err != nil {
		return manifest{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return manifest{}, fmt.Errorf("unexpected trailing JSON value")
		}
		return manifest{}, fmt.Errorf("parse trailing JSON: %w", err)
	}
	return candidate, nil
}

func validateManifest(candidate manifest) error {
	if candidate.Number != 3 || candidate.ID != "which-histories-are-legal" || candidate.Type != "practice" ||
		candidate.Title != "Which Histories Are Legal?" || candidate.RecommendedAfterChapter != 32 {
		return fmt.Errorf("manifest identity does not match Practice 03")
	}
	wantCapabilities := []string{
		"read-your-writes-review",
		"causal-dependency-review",
		"monotonic-reads-review",
		"linearizability-history-review",
		"guarantee-boundary",
	}
	if !sameStrings(candidate.Capabilities, wantCapabilities) {
		return fmt.Errorf("manifest capabilities do not match Practice 03")
	}
	if len(candidate.Artifacts) != 4 {
		return fmt.Errorf("manifest must declare exactly four artifacts")
	}
	for _, name := range []string{"contract", "histories", "starterReview", "solutionReview"} {
		path, ok := candidate.Artifacts[name]
		if !ok {
			return fmt.Errorf("manifest artifact %q is missing", name)
		}
		if !strings.HasPrefix(path, "practices/03-which-histories-are-legal/") {
			return fmt.Errorf("manifest artifact %q leaves Practice 03", name)
		}
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("manifest artifact %q: %w", name, err)
		}
	}
	if len(candidate.Commands) != 5 {
		return fmt.Errorf("manifest must declare exactly five commands")
	}
	for _, name := range []string{"failure", "solution", "boundary", "verifyReader", "check"} {
		spec, ok := candidate.Commands[name]
		if !ok {
			return fmt.Errorf("manifest command %q is missing", name)
		}
		if spec.Run == "" || spec.ExpectedExit == nil {
			return fmt.Errorf("manifest command %q lacks run or expectedExit", name)
		}
	}
	return nil
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
	if spec.ExpectedExit == nil {
		return fmt.Errorf("manifest command %q has no expected exit", name)
	}
	if exitCode != *spec.ExpectedExit {
		return fmt.Errorf("manifest command %q exit = %d, want %d\n%s", name, exitCode, *spec.ExpectedExit, output)
	}
	if spec.ExpectedOutput != "" && !strings.Contains(string(output), spec.ExpectedOutput) {
		return fmt.Errorf("manifest command %q did not emit %q\n%s", name, spec.ExpectedOutput, output)
	}
	return nil
}

func sameStrings(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for index := range expected {
		if actual[index] != expected[index] {
			return false
		}
	}
	return true
}

func fatalf(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "Practice 03 manifest check: "+format+"\n", arguments...)
	os.Exit(1)
}
