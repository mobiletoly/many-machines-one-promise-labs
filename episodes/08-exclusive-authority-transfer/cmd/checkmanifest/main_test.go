package main

import (
	"os"
	"testing"
)

func TestEpisode08ManifestSatisfiesExactContract(t *testing.T) {
	candidate := readManifest(t)
	if err := validateManifest(candidate); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeManifestRejectsUnknownField(t *testing.T) {
	data := readManifestBytes(t)
	data = append(data[:len(data)-2], []byte(",\n  \"answer\": 42\n}\n")...)
	if _, err := decodeManifest(data); err == nil {
		t.Fatal("manifest with unknown field unexpectedly passed")
	}
}

func TestDecodeManifestRejectsTrailingJSON(t *testing.T) {
	data := append(readManifestBytes(t), []byte("{}\n")...)
	if _, err := decodeManifest(data); err == nil {
		t.Fatal("manifest with trailing JSON unexpectedly passed")
	}
}

func TestValidateManifestRejectsContractDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*manifest)
	}{
		{
			name: "number",
			mutate: func(candidate *manifest) {
				candidate.Number = 999
			},
		},
		{
			name: "id",
			mutate: func(candidate *manifest) {
				candidate.ID = "wrong-episode"
			},
		},
		{
			name: "start state path",
			mutate: func(candidate *manifest) {
				candidate.States.Start = "episodes/08-exclusive-authority-transfer/elsewhere"
			},
		},
		{
			name: "failure command",
			mutate: func(candidate *manifest) {
				candidate.Commands.Failure.Run = "go test ./..."
			},
		},
		{
			name: "missing failure exit",
			mutate: func(candidate *manifest) {
				candidate.Commands.Failure.ExpectedExit = nil
			},
		},
		{
			name: "failure output",
			mutate: func(candidate *manifest) {
				candidate.Commands.Failure.ExpectedOutput = "different failure"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := readManifest(t)
			test.mutate(&candidate)
			if err := validateManifest(candidate); err == nil {
				t.Fatal("manifest contract drift unexpectedly passed")
			}
		})
	}
}

func readManifest(t *testing.T) manifest {
	t.Helper()
	candidate, err := decodeManifest(readManifestBytes(t))
	if err != nil {
		t.Fatal(err)
	}
	return candidate
}

func readManifestBytes(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../episode.json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}
