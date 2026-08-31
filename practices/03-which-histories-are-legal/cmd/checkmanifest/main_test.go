package main

import (
	"os"
	"strings"
	"testing"
)

func TestDecodeManifestRejectsUnknownAndTrailingJSON(t *testing.T) {
	data, err := os.ReadFile("../../practice.json")
	if err != nil {
		t.Fatal(err)
	}
	unknown := strings.Replace(string(data), `"capabilities":`, `"capabilites": ["typo"], "capabilities":`, 1)
	if _, err := decodeManifest([]byte(unknown)); err == nil {
		t.Fatal("manifest with unknown field unexpectedly passed")
	}
	if _, err := decodeManifest(append(data, []byte("\n{}\n")...)); err == nil {
		t.Fatal("manifest with trailing JSON unexpectedly passed")
	}
}
