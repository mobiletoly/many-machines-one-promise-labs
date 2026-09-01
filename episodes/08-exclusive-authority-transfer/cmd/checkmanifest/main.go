package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
)

const manifestPath = "episodes/08-exclusive-authority-transfer/episode.json"

type commandSpec struct {
	Run            string `json:"run"`
	ExpectedExit   *int   `json:"expectedExit"`
	ExpectedOutput string `json:"expectedOutput"`
}

type commandSet struct {
	HappyPath commandSpec `json:"happyPath"`
	Failure   commandSpec `json:"failure"`
	Solution  commandSpec `json:"solution"`
	Boundary  commandSpec `json:"boundary"`
	Verify    commandSpec `json:"verify"`
}

type stateSet struct {
	Start    string `json:"start"`
	Solution string `json:"solution"`
}

type manifest struct {
	Number                  int        `json:"number"`
	ID                      string     `json:"id"`
	Title                   string     `json:"title"`
	Concepts                []string   `json:"concepts"`
	RecommendedAfterChapter int        `json:"recommendedAfterChapter"`
	BasedOn                 string     `json:"basedOn"`
	RetainedProperties      []string   `json:"retainedProperties"`
	States                  stateSet   `json:"states"`
	Commands                commandSet `json:"commands"`
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
	fmt.Println("Episode 08 manifest: PASS")
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
	if candidate.Number != 8 || candidate.ID != "exclusive-authority-transfer" ||
		candidate.Title != "Booth A Has It, Booth B Still Does" ||
		candidate.RecommendedAfterChapter != 12 || candidate.BasedOn != "07-partitioned-authority" {
		return fmt.Errorf("manifest identity does not match Episode 08")
	}

	wantConcepts := []string{
		"exclusive-authority",
		"authority-transfer",
		"relinquishment-before-grant",
		"atomic-operation",
		"guarantee-boundary",
	}
	if !slices.Equal(candidate.Concepts, wantConcepts) {
		return fmt.Errorf("manifest concepts do not match Episode 08")
	}

	wantRetainedProperties := []string{
		"global-invariant",
		"logical-operation-identity",
		"matching-retry-idempotency",
		"conflicting-reuse-rejection",
	}
	if !slices.Equal(candidate.RetainedProperties, wantRetainedProperties) {
		return fmt.Errorf("manifest retained properties do not match Episode 08")
	}

	if candidate.States.Start != "episodes/08-exclusive-authority-transfer/start" ||
		candidate.States.Solution != "episodes/08-exclusive-authority-transfer/solution" {
		return fmt.Errorf("manifest state paths do not match Episode 08")
	}

	wantExitZero := 0
	wantExitOne := 1
	wantCommands := []struct {
		name string
		got  commandSpec
		want commandSpec
	}{
		{
			name: "happyPath",
			got:  candidate.Commands.HappyPath,
			want: commandSpec{
				Run:          "GOWORK=off go test -count=1 ./episodes/08-exclusive-authority-transfer/start -run '^TestGrantFirstWorkflowCanReachOneFinalHolder$'",
				ExpectedExit: &wantExitZero,
			},
		},
		{
			name: "failure",
			got:  candidate.Commands.Failure,
			want: commandSpec{
				Run:            "GOWORK=off go test -count=1 -tags=failure ./episodes/08-exclusive-authority-transfer/start -run '^TestExclusiveAuthoritySurvivesGrantDelivery$'",
				ExpectedExit:   &wantExitOne,
				ExpectedOutput: "exclusive authority violated: right R-100 confirmed by A-301 at booth-a and B-401 at booth-b during X-100",
			},
		},
		{
			name: "solution",
			got:  candidate.Commands.Solution,
			want: commandSpec{
				Run:          "GOWORK=off go test -count=1 -tags=failure ./episodes/08-exclusive-authority-transfer/solution -run '^TestExclusiveAuthoritySurvivesGrantDelivery$'",
				ExpectedExit: &wantExitZero,
			},
		},
		{
			name: "boundary",
			got:  candidate.Commands.Boundary,
			want: commandSpec{
				Run:          "GOWORK=off go test -count=1 ./episodes/08-exclusive-authority-transfer/solution -run '^TestBoundaryLostGrantLeavesAuthorityGap$'",
				ExpectedExit: &wantExitZero,
			},
		},
		{
			name: "verify",
			got:  candidate.Commands.Verify,
			want: commandSpec{
				Run:          "./scripts/verify-episode-08.sh",
				ExpectedExit: &wantExitZero,
			},
		},
	}

	for _, command := range wantCommands {
		if err := validateCommand(command.name, command.got, command.want); err != nil {
			return err
		}
	}
	return nil
}

func validateCommand(name string, got, want commandSpec) error {
	if got.Run != want.Run {
		return fmt.Errorf("manifest command %q run string does not match Episode 08", name)
	}
	if got.ExpectedExit == nil || *got.ExpectedExit != *want.ExpectedExit {
		return fmt.Errorf("manifest command %q expected exit does not match Episode 08", name)
	}
	if got.ExpectedOutput != want.ExpectedOutput {
		return fmt.Errorf("manifest command %q expected output does not match Episode 08", name)
	}
	return nil
}

func fatalf(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "Episode 08 manifest check: "+format+"\n", arguments...)
	os.Exit(1)
}
