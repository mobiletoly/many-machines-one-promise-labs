package strictjson

import (
	"strings"
	"testing"
)

type sample struct {
	ID     string `json:"id"`
	Nested struct {
		Value string `json:"value"`
	} `json:"nested"`
}

func TestRejectsDuplicateTopLevelKey(t *testing.T) {
	var value sample
	err := Unmarshal([]byte(`{"id":"one","id":"two","nested":{"value":"v"}}`), &value)
	if err == nil || !strings.Contains(err.Error(), `duplicate object key "id"`) {
		t.Fatalf("expected duplicate-key error, got %v", err)
	}
}

func TestRejectsDuplicateNestedKey(t *testing.T) {
	var value sample
	err := Unmarshal([]byte(`{"id":"one","nested":{"value":"v","value":"w"}}`), &value)
	if err == nil || !strings.Contains(err.Error(), `duplicate object key "value"`) {
		t.Fatalf("expected nested duplicate-key error, got %v", err)
	}
}

func TestRejectsUnknownAndTrailingJSON(t *testing.T) {
	for _, input := range []string{
		`{"id":"one","nested":{"value":"v"},"extra":true}`,
		`{"id":"one","nested":{"value":"v"}} {}`,
	} {
		var value sample
		if err := Unmarshal([]byte(input), &value); err == nil {
			t.Fatalf("expected strict decode failure for %s", input)
		}
	}
}
