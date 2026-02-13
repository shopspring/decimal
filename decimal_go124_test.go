//go:build go1.24

package decimal

import (
	"encoding/json"
	"testing"
)

// `omitzero` is supported by encoding/json starting in Go 1.24.
// Keep this test in a go1.24-gated file so older CI jobs skip it.
func TestJSONOmitZeroTag(t *testing.T) {
	type Nested struct {
		Amount Decimal `json:"amount,omitzero"`
	}

	type Parent struct {
		Nested Nested `json:"nested,omitzero"`
	}

	tests := []struct {
		name     string
		parent   Parent
		expected string
	}{
		{
			name: "Decimal{} empty value",
			parent: Parent{
				Nested: Nested{
					Amount: Decimal{},
				},
			},
			expected: "{}",
		},
		{
			name: "Zero constant",
			parent: Parent{
				Nested: Nested{
					Amount: Zero,
				},
			},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.parent)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(b))
			}
		})
	}
}
