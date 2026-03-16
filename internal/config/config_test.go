package config

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestProjectPathsUnmarshalJSON(t *testing.T) {
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

func TestParserOptionsUnmarshalJSON(t *testing.T) {
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

func TestParserOptionsProjectServicePtr(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectNil     bool
		expectedValue bool
	}{
		{
			name:      "not set",
			input:     `{}`,
			expectNil: true,
		},
		{
			name:          "explicitly true",
			input:         `{"projectService": true}`,
			expectNil:     false,
			expectedValue: true,
		},
		{
			name:          "explicitly false",
			input:         `{"projectService": false}`,
			expectNil:     false,
			expectedValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts ParserOptions
			err := json.Unmarshal([]byte(tt.input), &opts)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}
			if tt.expectNil {
				if opts.ProjectService != nil {
					t.Errorf("Expected ProjectService to be nil, got %v", *opts.ProjectService)
				}
			} else {
				if opts.ProjectService == nil {
					t.Fatalf("Expected ProjectService to be non-nil")
				}
				if *opts.ProjectService != tt.expectedValue {
					t.Errorf("Expected ProjectService to be %v, got %v", tt.expectedValue, *opts.ProjectService)
				}
			}
		})
	}
}

// TestRegisterAllRules_ConcurrentSafe verifies that RegisterAllRules can be
// called from multiple goroutines without panicking (concurrent map writes).
// Run with -race to detect data races: go test -race ./internal/config/...
func TestRegisterAllRules_ConcurrentSafe(t *testing.T) {
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			RegisterAllRules()
		}()
	}
	wg.Wait()

	// Verify rules were actually registered
	rules := GlobalRuleRegistry.GetAllRules()
	if len(rules) == 0 {
		t.Error("Expected rules to be registered after concurrent calls")
	}
}
