package main

import (
	"os"
	"testing"
)

func TestDecodeManifestRejectsUnknownField(t *testing.T) {
	data, err := os.ReadFile("../../practice.json")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data[:len(data)-2], []byte(",\n  \"answer\": 42\n}\n")...)
	if _, err := decodeManifest(data); err == nil {
		t.Fatal("manifest with unknown field unexpectedly passed")
	}
}

func TestDecodeManifestRejectsTrailingJSON(t *testing.T) {
	data, err := os.ReadFile("../../practice.json")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, []byte("{}\n")...)
	if _, err := decodeManifest(data); err == nil {
		t.Fatal("manifest with trailing JSON unexpectedly passed")
	}
}
