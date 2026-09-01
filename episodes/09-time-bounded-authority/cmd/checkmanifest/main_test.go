package main

import (
	"bytes"
	"os"
	"testing"
)

func TestEpisode09ManifestSatisfiesExactContract(t *testing.T) {
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

func TestDecodeManifestRejectsDuplicateTopLevelKey(t *testing.T) {
	data := bytes.Replace(
		readManifestBytes(t),
		[]byte("\"number\": 9,"),
		[]byte("\"number\": 999,\n  \"number\": 9,"),
		1,
	)
	if _, err := decodeManifest(data); err == nil {
		t.Fatal("manifest with duplicate top-level key unexpectedly passed")
	}
}

func TestDecodeManifestRejectsDuplicateNestedKey(t *testing.T) {
	data := bytes.Replace(
		readManifestBytes(t),
		[]byte("\"run\": \"GOWORK=off go test -count=1 -tags=failure ./episodes/09-time-bounded-authority/start -run '^TestLeaseExpiryIsEnforcedAtEffectAcceptance$'\""),
		[]byte("\"run\": \"go test ./...\",\n      \"run\": \"GOWORK=off go test -count=1 -tags=failure ./episodes/09-time-bounded-authority/start -run '^TestLeaseExpiryIsEnforcedAtEffectAcceptance$'\""),
		1,
	)
	if _, err := decodeManifest(data); err == nil {
		t.Fatal("manifest with duplicate nested key unexpectedly passed")
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
				candidate.States.Start = "episodes/09-time-bounded-authority/elsewhere"
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
