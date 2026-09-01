package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
)

const manifestPath = "episodes/09-time-bounded-authority/episode.json"

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
	fmt.Println("Episode 09 manifest: PASS")
}

func decodeManifest(data []byte) (manifest, error) {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return manifest{}, err
	}

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

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder, "$"); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("unexpected trailing JSON value")
		}
		return fmt.Errorf("parse trailing JSON: %w", err)
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder, path string) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	switch delimiter {
	case '{':
		keys := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("object key at %s is not a string", path)
			}
			if _, exists := keys[key]; exists {
				return fmt.Errorf("duplicate object key %q at %s", key, path)
			}
			keys[key] = struct{}{}
			if err := scanJSONValue(decoder, path+"."+key); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim('}') {
			return fmt.Errorf("object at %s has invalid closing delimiter", path)
		}
		return nil
	case '[':
		for index := 0; decoder.More(); index++ {
			if err := scanJSONValue(decoder, fmt.Sprintf("%s[%d]", path, index)); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim(']') {
			return fmt.Errorf("array at %s has invalid closing delimiter", path)
		}
		return nil
	default:
		return fmt.Errorf("unexpected delimiter %q at %s", delimiter, path)
	}
}

func validateManifest(candidate manifest) error {
	if candidate.Number != 9 || candidate.ID != "time-bounded-authority" ||
		candidate.Title != "The Desk Read 110" || candidate.RecommendedAfterChapter != 28 {
		return fmt.Errorf("manifest identity does not match Episode 09")
	}

	wantConcepts := []string{
		"lease",
		"time-bounded-authority",
		"effect-authority",
		"monotonic-clock",
		"atomic-operation",
		"guarantee-boundary",
	}
	if !slices.Equal(candidate.Concepts, wantConcepts) {
		return fmt.Errorf("manifest concepts do not match Episode 09")
	}

	if candidate.States.Start != "episodes/09-time-bounded-authority/start" ||
		candidate.States.Solution != "episodes/09-time-bounded-authority/solution" {
		return fmt.Errorf("manifest state paths do not match Episode 09")
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
				Run:          "GOWORK=off go test -count=1 ./episodes/09-time-bounded-authority/start -run '^TestEstablishedLeaseSupportsPublicationWithoutIssuer$'",
				ExpectedExit: &wantExitZero,
			},
		},
		{
			name: "failure",
			got:  candidate.Commands.Failure,
			want: commandSpec{
				Run:            "GOWORK=off go test -count=1 -tags=failure ./episodes/09-time-bounded-authority/start -run '^TestLeaseExpiryIsEnforcedAtEffectAcceptance$'",
				ExpectedExit:   &wantExitOne,
				ExpectedOutput: "lease expiry violated: publish(operation=P-after, lease=L-88) decided at S time 110 = accepted, want lease_expired",
			},
		},
		{
			name: "solution",
			got:  candidate.Commands.Solution,
			want: commandSpec{
				Run:          "GOWORK=off go test -count=1 -tags=failure ./episodes/09-time-bounded-authority/solution -run '^TestLeaseExpiryIsEnforcedAtEffectAcceptance$'",
				ExpectedExit: &wantExitZero,
			},
		},
		{
			name: "boundary",
			got:  candidate.Commands.Boundary,
			want: commandSpec{
				Run:          "GOWORK=off go test -count=1 ./episodes/09-time-bounded-authority/solution -run '^TestBoundary'",
				ExpectedExit: &wantExitZero,
			},
		},
		{
			name: "verify",
			got:  candidate.Commands.Verify,
			want: commandSpec{
				Run:          "./scripts/verify-episode-09.sh",
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
		return fmt.Errorf("manifest command %q run string does not match Episode 09", name)
	}
	if got.ExpectedExit == nil || *got.ExpectedExit != *want.ExpectedExit {
		return fmt.Errorf("manifest command %q expected exit does not match Episode 09", name)
	}
	if got.ExpectedOutput != want.ExpectedOutput {
		return fmt.Errorf("manifest command %q expected output does not match Episode 09", name)
	}
	return nil
}

func fatalf(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "Episode 09 manifest check: "+format+"\n", arguments...)
	os.Exit(1)
}
