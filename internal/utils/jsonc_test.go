package utils

import (
	"encoding/json"
	"testing"
)

func TestParseJSONC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "basic JSON",
			input: `{
				"key": "value",
				"number": 42
			}`,
			expected: map[string]interface{}{
				"key":    "value",
				"number": float64(42),
			},
			wantErr: false,
		},
		{
			name: "JSON with single-line comments",
			input: `{
				"key": "value", // This is a comment
				"number": 42 // Another comment
			}`,
			expected: map[string]interface{}{
				"key":    "value",
				"number": float64(42),
			},
			wantErr: false,
		},
		{
			name: "JSON with multi-line comments",
			input: `{
				"key": "value", /* This is a 
				multi-line comment */
				"number": 42
			}`,
			expected: map[string]interface{}{
				"key":    "value",
				"number": float64(42),
			},
			wantErr: false,
		},
		{
			name: "JSON with trailing commas",
			input: `{
				"key": "value",
				"number": 42,
			}`,
			expected: map[string]interface{}{
				"key":    "value",
				"number": float64(42),
			},
			wantErr: false,
		},
		{
			name: "JSON with comments and trailing commas",
			input: `{
				"key": "value", // comment
				"number": 42, /* multi-line
				comment */
			}`,
			expected: map[string]interface{}{
				"key":    "value",
				"number": float64(42),
			},
			wantErr: false,
		},
		{
			name: "JSON array with trailing comma",
			input: `{
				"array": [1, 2, 3,],
				"nested": {
					"key": "value",
				}
			}`,
			expected: map[string]interface{}{
				"array": []interface{}{float64(1), float64(2), float64(3)},
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := ParseJSONC([]byte(tt.input), &result)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSONC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Compare the results
				expectedJSON, _ := json.Marshal(tt.expected)
				resultJSON, _ := json.Marshal(result)

				if string(expectedJSON) != string(resultJSON) {
					t.Errorf("ParseJSONC() = %v, want %v", string(resultJSON), string(expectedJSON))
				}
			}
		})
	}
}

func TestStripJSONComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "no comments",
			input: `{"key": "value"}`,
		},
		{
			name: "single line comment inside",
			input: `{
				"key": "value" // comment
			}`,
		},
		{
			name:  "multi-line comment",
			input: `{"key": /* comment */ "value"}`,
		},
		{
			name:  "trailing comma",
			input: `{"key": "value",}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripJSONComments(tt.input)

			// Verify that the result is valid JSON by trying to parse it
			var parsed map[string]interface{}
			err := json.Unmarshal([]byte(result), &parsed)
			if err != nil {
				t.Errorf("StripJSONComments() produced invalid JSON: %v, result: %q", err, result)
				return
			}

			// Verify that the key-value pair is preserved
			if parsed["key"] != "value" {
				t.Errorf("StripJSONComments() = %v, expected key 'value' to be preserved", result)
			}
		})
	}
}
