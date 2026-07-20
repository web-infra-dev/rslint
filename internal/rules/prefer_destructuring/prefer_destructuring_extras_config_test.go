package prefer_destructuring

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

type preferDestructuringConfigCase struct {
	name    string
	options any
	want    options
}

func preferDestructuringConfigMatrix() []preferDestructuringConfigCase {
	all := destructuringConfig{array: true, object: true}
	none := destructuringConfig{}
	arrayOnly := destructuringConfig{array: true}
	objectOnly := destructuringConfig{object: true}

	return []preferDestructuringConfigCase{
		{
			name: "defaults",
			want: options{
				variableDeclarator:   all,
				assignmentExpression: all,
			},
		},
		{
			name:    "flat all",
			options: []any{map[string]any{"array": true, "object": true}},
			want: options{
				variableDeclarator:   all,
				assignmentExpression: all,
			},
		},
		{
			// A bare single object follows the same NormalizeOptions path used
			// by production configuration loading.
			name:    "flat object",
			options: map[string]any{"object": true},
			want: options{
				variableDeclarator:   objectOnly,
				assignmentExpression: objectOnly,
			},
		},
		{
			name:    "flat array",
			options: []any{map[string]any{"array": true}},
			want: options{
				variableDeclarator:   arrayOnly,
				assignmentExpression: arrayOnly,
			},
		},
		{
			name:    "flat none",
			options: map[string]any{"array": false, "object": false},
			want: options{
				variableDeclarator:   none,
				assignmentExpression: none,
			},
		},
		{
			name: "per context all",
			options: []any{map[string]any{
				"VariableDeclarator":   map[string]any{"array": true, "object": true},
				"AssignmentExpression": map[string]any{"array": true, "object": true},
			}},
			want: options{
				variableDeclarator:   all,
				assignmentExpression: all,
			},
		},
		{
			name: "declarations only",
			options: []any{map[string]any{
				"VariableDeclarator": map[string]any{"array": true, "object": true},
			}},
			want: options{
				variableDeclarator:   all,
				assignmentExpression: none,
			},
		},
		{
			name: "assignments only",
			options: []any{map[string]any{
				"AssignmentExpression": map[string]any{"array": true, "object": true},
			}},
			want: options{
				variableDeclarator:   none,
				assignmentExpression: all,
			},
		},
		{
			name: "object declarations array assignments",
			options: []any{map[string]any{
				"VariableDeclarator":   map[string]any{"array": false, "object": true},
				"AssignmentExpression": map[string]any{"array": true, "object": false},
			}},
			want: options{
				variableDeclarator:   objectOnly,
				assignmentExpression: arrayOnly,
			},
		},
		{
			name: "array declarations object assignments",
			options: []any{map[string]any{
				"VariableDeclarator":   map[string]any{"array": true, "object": false},
				"AssignmentExpression": map[string]any{"array": false, "object": true},
			}},
			want: options{
				variableDeclarator:   arrayOnly,
				assignmentExpression: objectOnly,
			},
		},
		{
			name: "enforce all",
			options: []any{
				map[string]any{"array": true, "object": true},
				map[string]any{"enforceForRenamedProperties": true},
			},
			want: options{
				variableDeclarator:   all,
				assignmentExpression: all,
				enforceRenamed:       true,
			},
		},
		{
			name: "enforce object",
			options: []any{
				map[string]any{"object": true},
				map[string]any{"enforceForRenamedProperties": true},
			},
			want: options{
				variableDeclarator:   objectOnly,
				assignmentExpression: objectOnly,
				enforceRenamed:       true,
			},
		},
		{
			name: "enforce array",
			options: []any{
				map[string]any{"array": true},
				map[string]any{"enforceForRenamedProperties": true},
			},
			want: options{
				variableDeclarator:   arrayOnly,
				assignmentExpression: arrayOnly,
				enforceRenamed:       true,
			},
		},
		{
			name: "explicit no enforcement",
			options: []any{
				map[string]any{"array": true, "object": true},
				map[string]any{"enforceForRenamedProperties": false},
			},
			want: options{
				variableDeclarator:   all,
				assignmentExpression: all,
			},
		},
		{
			name: "mixed enforced",
			options: []any{
				map[string]any{
					"VariableDeclarator":   map[string]any{"array": false, "object": true},
					"AssignmentExpression": map[string]any{"array": true, "object": false},
				},
				map[string]any{"enforceForRenamedProperties": true},
			},
			want: options{
				variableDeclarator:   objectOnly,
				assignmentExpression: arrayOnly,
				enforceRenamed:       true,
			},
		},
	}
}

func TestPreferDestructuringParseOptionsMatrix(t *testing.T) {
	for _, testCase := range preferDestructuringConfigMatrix() {
		t.Run(testCase.name, func(t *testing.T) {
			got := parseOptions(rule.NormalizeOptions(testCase.options))
			if got != testCase.want {
				t.Fatalf("parseOptions() = %#v, want %#v", got, testCase.want)
			}
		})
	}

	all := destructuringConfig{array: true, object: true}
	directAPICases := []struct {
		name string
		raw  []any
		want options
	}{
		{
			// A nil first option is schema-invalid, but the direct API fallback
			// follows upstream's default-option behavior without panicking.
			name: "nil first option",
			raw:  []any{nil},
			want: options{
				variableDeclarator:   all,
				assignmentExpression: all,
			},
		},
		{
			// Public configuration rejects this ambiguous oneOf value. Keep a
			// deterministic fallback for direct Rule.Run callers that bypass
			// schema validation.
			name: "empty object",
			raw:  []any{map[string]any{}},
			want: options{},
		},
	}
	for _, testCase := range directAPICases {
		t.Run("direct API/"+testCase.name, func(t *testing.T) {
			got := parseOptions(testCase.raw)
			if got != testCase.want {
				t.Fatalf("parseOptions() = %#v, want %#v", got, testCase.want)
			}
		})
	}
}

func TestPreferDestructuringSchemaMatrix(t *testing.T) {
	valid := []struct {
		name    string
		options []any
	}{
		{name: "nil"},
		{name: "empty", options: []any{}},
		{name: "flat array", options: []any{map[string]any{"array": true}}},
		{name: "flat object", options: []any{map[string]any{"object": false}}},
		{
			name: "per context empty nested object",
			options: []any{map[string]any{
				"VariableDeclarator": map[string]any{},
			}},
		},
		{
			name: "per context both",
			options: []any{map[string]any{
				"VariableDeclarator":   map[string]any{"array": true, "object": false},
				"AssignmentExpression": map[string]any{"array": false, "object": true},
			}},
		},
		{
			name: "renamed true",
			options: []any{
				map[string]any{"object": true},
				map[string]any{"enforceForRenamedProperties": true},
			},
		},
		{
			name: "renamed false",
			options: []any{
				map[string]any{"array": true},
				map[string]any{"enforceForRenamedProperties": false},
			},
		},
		{
			name: "empty enforcement object",
			options: []any{
				map[string]any{"array": true},
				map[string]any{},
			},
		},
	}
	for _, testCase := range valid {
		t.Run("valid/"+testCase.name, func(t *testing.T) {
			if err := PreferDestructuringRule.Schema.Validate(testCase.options); err != nil {
				t.Fatalf("expected options to be valid, got %v", err)
			}
		})
	}

	invalid := []struct {
		name    string
		options []any
	}{
		{
			// Both oneOf object shapes accept an object with no properties, so
			// Draft 4 oneOf correctly rejects it for matching twice.
			name:    "ambiguous empty first object",
			options: []any{map[string]any{}},
		},
		{name: "null first option", options: []any{nil}},
		{name: "non object first option", options: []any{"object"}},
		{
			name: "mixed flat and per context shapes",
			options: []any{map[string]any{
				"object":             true,
				"VariableDeclarator": map[string]any{"object": true},
			}},
		},
		{
			name:    "wrong flat value type",
			options: []any{map[string]any{"array": "true"}},
		},
		{
			name:    "unknown flat property",
			options: []any{map[string]any{"object": true, "unknown": true}},
		},
		{
			name: "wrong nested value type",
			options: []any{map[string]any{
				"VariableDeclarator": map[string]any{"object": 1},
			}},
		},
		{
			name: "unknown nested property",
			options: []any{map[string]any{
				"AssignmentExpression": map[string]any{"unknown": true},
			}},
		},
		{
			name: "unknown enforcement property",
			options: []any{
				map[string]any{"object": true},
				map[string]any{"unknown": true},
			},
		},
		{
			name: "wrong enforcement value type",
			options: []any{
				map[string]any{"object": true},
				map[string]any{"enforceForRenamedProperties": 1},
			},
		},
		{
			name: "null enforcement object",
			options: []any{
				map[string]any{"object": true},
				nil,
			},
		},
		{
			name: "too many options",
			options: []any{
				map[string]any{"object": true},
				map[string]any{"enforceForRenamedProperties": true},
				map[string]any{},
			},
		},
	}
	for _, testCase := range invalid {
		t.Run("invalid/"+testCase.name, func(t *testing.T) {
			if err := PreferDestructuringRule.Schema.Validate(testCase.options); err == nil {
				t.Fatal("expected options to be rejected")
			}
		})
	}
}

func TestPreferDestructuringRuntimeConfigMatrix(t *testing.T) {
	lines := []string{
		"const foo = object.foo;",
		"const local = object.remote;",
		"const value = array[0];",
		"foo = object.foo;",
		"local = object.remote;",
		"value = array[0];",
	}
	code := strings.Join(lines, "\n")

	configurations := preferDestructuringConfigMatrix()
	valid := make([]rule_tester.ValidTestCase, 0, 1)
	invalid := make([]rule_tester.InvalidTestCase, 0, len(configurations)-1)
	for _, configuration := range configurations {
		want := configuration.want
		errors := make([]rule_tester.InvalidTestCaseError, 0, 6)
		if want.variableDeclarator.object {
			errors = append(errors, preferError("object", 1, 7, 1, len(lines[0])))
		}
		if want.variableDeclarator.object && want.enforceRenamed {
			errors = append(errors, preferError("object", 2, 7, 2, len(lines[1])))
		}
		if want.variableDeclarator.array {
			errors = append(errors, preferError("array", 3, 7, 3, len(lines[2])))
		}
		if want.assignmentExpression.object {
			errors = append(errors, preferError("object", 4, 1, 4, len(lines[3])))
		}
		if want.assignmentExpression.object && want.enforceRenamed {
			errors = append(errors, preferError("object", 5, 1, 5, len(lines[4])))
		}
		if want.assignmentExpression.array {
			errors = append(errors, preferError("array", 6, 1, 6, len(lines[5])))
		}

		if len(errors) == 0 {
			valid = append(valid, rule_tester.ValidTestCase{
				Code:    code,
				Options: configuration.options,
			})
			continue
		}

		var output []string
		if want.variableDeclarator.object {
			output = []string{
				"const {foo} = object;\n" +
					lines[1] + "\n" +
					lines[2] + "\n" +
					lines[3] + "\n" +
					lines[4] + "\n" +
					lines[5],
			}
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:    code,
			Output:  output,
			Options: configuration.options,
			Errors:  errors,
		})
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		valid,
		invalid,
	)
}
