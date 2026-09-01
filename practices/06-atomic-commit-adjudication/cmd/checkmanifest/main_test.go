package main

import (
	"os"
	"strings"
	"testing"
)

func manifestBytes(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../practice.json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestDecodeManifestRejectsUnknownField(t *testing.T) {
	data := manifestBytes(t)
	data = append(data[:len(data)-2], []byte(",\n  \"answer\": 42\n}\n")...)
	if _, err := decodeManifest(data); err == nil {
		t.Fatal("manifest with unknown field unexpectedly passed")
	}
}

func TestDecodeManifestRejectsTrailingJSON(t *testing.T) {
	data := append(manifestBytes(t), []byte("{}\n")...)
	if _, err := decodeManifest(data); err == nil {
		t.Fatal("manifest with trailing JSON unexpectedly passed")
	}
}

func TestDecodeManifestRejectsDuplicateTopLevelKey(t *testing.T) {
	data := manifestBytes(t)
	data = append(data[:len(data)-2], []byte(",\n  \"id\": \"replacement\"\n}\n")...)
	_, err := decodeManifest(data)
	if err == nil || !strings.Contains(err.Error(), `duplicate object key "id"`) {
		t.Fatalf("expected duplicate id failure, got %v", err)
	}
}

func TestDecodeManifestRejectsDuplicateNestedKey(t *testing.T) {
	data := manifestBytes(t)
	needle := []byte(`"contract": "practices/06-atomic-commit-adjudication/contract.json",`)
	replacement := []byte(`"contract": "practices/06-atomic-commit-adjudication/contract.json", "contract": "replacement",`)
	data = []byte(strings.Replace(string(data), string(needle), string(replacement), 1))
	_, err := decodeManifest(data)
	if err == nil || !strings.Contains(err.Error(), `duplicate object key "contract"`) {
		t.Fatalf("expected nested duplicate key failure, got %v", err)
	}
}

func TestValidateManifestRejectsIdentityDrift(t *testing.T) {
	candidate, err := decodeManifest(manifestBytes(t))
	if err != nil {
		t.Fatal(err)
	}
	candidate.RecommendedAfterChapter = 25
	if err := validateManifest(candidate); err == nil {
		t.Fatal("identity drift unexpectedly passed")
	}
}
