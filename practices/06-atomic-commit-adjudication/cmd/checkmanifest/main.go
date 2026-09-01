package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mobiletoly/many-machines-one-promise-labs/practices/06-atomic-commit-adjudication/internal/strictjson"
)

const manifestPath = "practices/06-atomic-commit-adjudication/practice.json"

type commandSpec struct {
	Run            string `json:"run"`
	ExpectedExit   *int   `json:"expectedExit"`
	ExpectedOutput string `json:"expectedOutput,omitempty"`
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
	if err := strictjson.Unmarshal(data, &candidate); err != nil {
		return manifest{}, err
	}
	return candidate, nil
}

func validateManifest(candidate manifest) error {
	if candidate.Number != 6 || candidate.ID != "atomic-commit-adjudication" || candidate.Type != "practice" ||
		candidate.Title != "Can This Participant Finish?" || candidate.RecommendedAfterChapter != 24 {
		return fmt.Errorf("manifest identity does not match Practice 06")
	}
	wantCapabilities := []string{
		"atomic-commit-evidence-adjudication",
		"prepared-capability-review",
		"decision-authority-review",
		"recovery-disposition-certification",
		"guarantee-boundary",
	}
	if !sameStrings(candidate.Capabilities, wantCapabilities) {
		return fmt.Errorf("manifest capabilities do not match Practice 06")
	}
	wantArtifacts := map[string]string{
		"contract":            "contract.json",
		"cases":               "cases.json",
		"participantEvidence": "participant-evidence.jsonl",
		"decisionEvidence":    "decision-evidence.jsonl",
		"starterReview":       "starter/adjudication.json",
		"solutionReview":      "solution/adjudication.json",
	}
	if len(candidate.Artifacts) != len(wantArtifacts) {
		return fmt.Errorf("manifest must declare exactly six artifacts")
	}
	for name, suffix := range wantArtifacts {
		path, ok := candidate.Artifacts[name]
		if !ok || path != "practices/06-atomic-commit-adjudication/"+suffix {
			return fmt.Errorf("manifest artifact %q does not match Practice 06", name)
		}
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("manifest artifact %q: %w", name, err)
		}
	}
	wantCommands := []string{"failure", "verifyReader", "solution", "boundary", "check"}
	if len(candidate.Commands) != len(wantCommands) {
		return fmt.Errorf("manifest must declare exactly five commands")
	}
	for _, name := range wantCommands {
		spec, ok := candidate.Commands[name]
		if !ok || spec.Run == "" || spec.ExpectedExit == nil {
			return fmt.Errorf("manifest command %q lacks run or expectedExit", name)
		}
	}
	return nil
}

func runCommand(name string, spec commandSpec) error {
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
	fmt.Fprintf(os.Stderr, "Practice 06 manifest check: "+format+"\n", arguments...)
	os.Exit(1)
}
