package ipc

import (
    json "github.com/go-json-experiment/json"
	"testing"
)

func TestAPIProjectPathsUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single string",
			input:    `"tsconfig.json"`,
			expected: []string{"tsconfig.json"},
		},
		{
			name:     "array of strings",
			input:    `["tsconfig.json", "packages/*/tsconfig.json"]`,
			expected: []string{"tsconfig.json", "packages/*/tsconfig.json"},
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paths ProjectPaths
			err := json.Unmarshal([]byte(tt.input), &paths)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if len(paths) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(paths))
			}

			for i, expected := range tt.expected {
				if i >= len(paths) {
					t.Errorf("Expected %s at index %d, but paths is too short", expected, i)
					continue
				}
				if paths[i] != expected {
					t.Errorf("Expected %s at index %d, got %s", expected, i, paths[i])
				}
			}
		})
	}
}

func TestAPIParserOptionsUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ProjectPaths
	}{
		{
			name:     "single project string",
			input:    `{"projectService": false, "project": "tsconfig.json"}`,
			expected: ProjectPaths{"tsconfig.json"},
		},
		{
			name:     "multiple project strings",
			input:    `{"projectService": false, "project": ["tsconfig.json", "packages/*/tsconfig.json"]}`,
			expected: ProjectPaths{"tsconfig.json", "packages/*/tsconfig.json"},
		},
		{
			name:     "no project field",
			input:    `{"projectService": false}`,
			expected: ProjectPaths{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts ParserOptions
			err := json.Unmarshal([]byte(tt.input), &opts)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if len(opts.Project) != len(tt.expected) {
				t.Errorf("Expected project length %d, got %d", len(tt.expected), len(opts.Project))
			}

			for i, expected := range tt.expected {
				if i >= len(opts.Project) {
					t.Errorf("Expected %s at index %d, but project is too short", expected, i)
					continue
				}
				if opts.Project[i] != expected {
					t.Errorf("Expected %s at index %d, got %s", expected, i, opts.Project[i])
				}
			}
		})
	}
}
