package youtrack

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/quick"
)

func TestIssueCustomFieldsDecodeDisplayValues(t *testing.T) {
	var issue Issue
	err := json.Unmarshal([]byte(`{
		"id":"1",
		"idReadable":"ABC-1",
		"summary":"One",
		"project":{"shortName":"ABC"},
		"customFields":[
			{"name":"State","value":{"name":"Open"}},
			{"name":"Assignee","value":[{"login":"jane"},{"login":"max"}]},
			{"name":"Priority","value":{"presentation":"Major"}},
			{"name":"Text","value":{"text":"details"}},
			{"name":"Scalar","value":42}
		]
	}`), &issue)
	if err != nil {
		t.Fatalf("decode issue: %v", err)
	}

	tests := map[string]string{
		"State":    "Open",
		"Assignee": "jane, max",
		"Priority": "Major",
		"Text":     "details",
		"Scalar":   "42",
	}
	for name, want := range tests {
		if got := issue.CustomFieldValue(name); got != want {
			t.Fatalf("CustomFieldValue(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestCustomFieldStringValueProperty(t *testing.T) {
	property := func(name, value string) bool {
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" || value == "" {
			return true
		}

		data, err := json.Marshal(map[string]any{
			"name":  name,
			"value": value,
		})
		if err != nil {
			return false
		}
		var field CustomField
		if err := json.Unmarshal(data, &field); err != nil {
			return false
		}
		return field.Name == name && field.Value == value
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func FuzzCustomFieldDisplayValue(f *testing.F) {
	for _, seed := range []string{
		`"Open"`,
		`{"name":"Open"}`,
		`{"login":"jane"}`,
		`{"presentation":"Major"}`,
		`[{"login":"jane"},{"login":"max"}]`,
		`42`,
		`null`,
		`not json`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		value := CustomFieldDisplayValue([]byte(raw))
		if strings.Contains(value, "\n") {
			t.Fatalf("CustomFieldDisplayValue(%q) contains newline: %q", raw, value)
		}
	})
}
