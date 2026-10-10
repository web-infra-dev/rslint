// This suite is generated from every test case in eslint-plugin-unicorn v77.0.0's
// test/name-replacements.js. The checked-in JSON keeps the upstream grouping and
// expectations reviewable; this adapter supplies rslint's native rule context.
package name_replacements_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/name_replacements"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type upstreamGroup struct {
	Kind  string        `json:"kind"`
	Cases upstreamCases `json:"cases"`
}

type upstreamCases struct {
	TesterOptions upstreamTesterOptions `json:"testerOptions"`
	Valid         []json.RawMessage     `json:"valid"`
	Invalid       []json.RawMessage     `json:"invalid"`
}

type upstreamTesterOptions struct {
	LanguageOptions struct {
		Globals       map[string]any `json:"globals"`
		ParserOptions struct {
			ECMAFeatures struct {
				JSX bool `json:"jsx"`
			} `json:"ecmaFeatures"`
		} `json:"parserOptions"`
	} `json:"languageOptions"`
}

type upstreamCase struct {
	Code     string          `json:"code"`
	Filename string          `json:"filename"`
	Language string          `json:"language"`
	Options  any             `json:"options"`
	Output   json.RawMessage `json:"output"`
	Errors   json.RawMessage `json:"errors"`
}

type upstreamError struct {
	Message     any             `json:"message"`
	MessageID   string          `json:"messageId"`
	Line        int             `json:"line"`
	Column      int             `json:"column"`
	EndLine     int             `json:"endLine"`
	EndColumn   int             `json:"endColumn"`
	Suggestions json.RawMessage `json:"suggestions"`
}

type upstreamSuggestion struct {
	MessageID string `json:"messageId"`
	Output    string `json:"output"`
}

func decodeUpstreamCase(raw json.RawMessage) (upstreamCase, error) {
	var code string
	if err := json.Unmarshal(raw, &code); err == nil {
		return upstreamCase{Code: code}, nil
	}
	var result upstreamCase
	err := json.Unmarshal(raw, &result)
	return result, err
}

// normalizeRegexpOptions converts JavaScript RegExp literals captured from
// upstream into the string form native rslint configuration can carry. A
// case-insensitive RegExp cannot be represented by this option's string form,
// so those few cases remain explained skips instead of changing expectations.
func normalizeRegexpOptions(value any) (any, bool) {
	switch value := value.(type) {
	case []any:
		result := make([]any, len(value))
		for index, item := range value {
			var unsupported bool
			result[index], unsupported = normalizeRegexpOptions(item)
			if unsupported {
				return nil, true
			}
		}
		return result, false
	case map[string]any:
		if pattern, ok := value["__regexp"].(string); ok {
			flags, _ := value["flags"].(string)
			if strings.Contains(flags, "i") {
				return nil, true
			}
			return pattern, false
		}
		result := make(map[string]any, len(value))
		for key, item := range value {
			var unsupported bool
			result[key], unsupported = normalizeRegexpOptions(item)
			if unsupported {
				return nil, true
			}
		}
		return result, false
	default:
		return value, false
	}
}

func upstreamExpectedErrors(t *testing.T, raw json.RawMessage) (int, []upstreamError) {
	t.Helper()
	var count int
	if err := json.Unmarshal(raw, &count); err == nil {
		return count, nil
	}
	var errors []upstreamError
	if err := json.Unmarshal(raw, &errors); err != nil {
		t.Fatalf("decode upstream errors: %v", err)
	}
	return len(errors), errors
}

func upstreamSuggestionCount(t *testing.T, raw json.RawMessage) (int, []upstreamSuggestion) {
	t.Helper()
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}
	var count int
	if err := json.Unmarshal(raw, &count); err == nil {
		return count, nil
	}
	var suggestions []upstreamSuggestion
	if err := json.Unmarshal(raw, &suggestions); err != nil {
		t.Fatalf("decode upstream suggestions: %v", err)
	}
	return len(suggestions), suggestions
}

func configuredGlobals(raw map[string]any) map[string]utils.GlobalAccess {
	result := make(map[string]utils.GlobalAccess, len(raw))
	for name, value := range raw {
		if access, ok := utils.NormalizeGlobalAccess(value); ok {
			result[name] = access
		}
	}
	return result
}

func unsupportedNativeTestFilename(filename string) bool {
	if filename == "" {
		return false
	}
	switch filepath.Ext(filename) {
	case ".js", ".jsx", ".mjs", ".cjs", ".ts", ".tsx", ".mts", ".cts":
		return false
	default:
		return true
	}
}

func lintUpstreamCase(t *testing.T, testCase upstreamCase, globals map[string]any, typescript bool, tsx bool, index int) []rule.RuleDiagnostic {
	t.Helper()
	root := fixtures.GetRootDir()
	extension := ".js"
	if typescript {
		extension = ".ts"
	}
	if tsx {
		extension = ".tsx"
	}
	fileName := testCase.Filename
	if fileName == "" {
		fileName = "name-replacements-upstream-" + strconv.Itoa(index) + extension
	}
	resolvedOptions := rule_tester.ResolveTestCaseOptions(t, &name_replacements.NameReplacementsRule, testCase.Options)
	resolvedFileName := tspath.ResolvePath(root.Dir, fileName)
	fs := utils.NewOverlayVFS(root.FS, map[string]string{resolvedFileName: testCase.Code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	var sourceProgram *lintprogram.Program
	var sourceFile *ast.SourceFile
	var err error
	if strings.HasPrefix(filepath.Base(fileName), ".") {
		sourceProgram, err = lintprogram.NewFromRoots(lintprogram.RootOptions{
			RootFileNames:   []string{resolvedFileName},
			Host:            host,
			CompilerOptions: lintprogram.SourceOnlyCompilerOptions(),
			SingleThreaded:  true,
		})
		if sourceProgram != nil {
			sourceFile = sourceProgram.GetSourceFile(resolvedFileName)
		}
	} else {
		var compilerProgram *compiler.Program
		compilerProgram, err = utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
		if compilerProgram != nil {
			sourceProgram = lintprogram.NewFromCompiler(compilerProgram)
			sourceFile = compilerProgram.GetSourceFile(fileName)
		}
	}
	if err != nil {
		t.Fatalf("create program: %v\ncode:\n%s", err, testCase.Code)
	}
	if sourceFile == nil {
		t.Fatalf("source file %q missing from program", fileName)
	}

	var diagnosticsMu sync.Mutex
	var diagnostics []rule.RuleDiagnostic
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{sourceProgram},
		TargetsByProgram: [][]string{{sourceFile.FileName()}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name: "unicorn/name-replacements",
				Environment: &rule.RuleEnvironment{
					Globals: configuredGlobals(globals),
				},
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return name_replacements.NameReplacementsRule.Run(ctx, resolvedOptions)
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("prepare lint plan: %v", err)
	}
	_, err = linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnosticsMu.Lock()
				diagnostics = append(diagnostics, diagnostic)
				diagnosticsMu.Unlock()
			},
		},
	})
	if err != nil {
		t.Fatalf("run linter: %v", err)
	}
	return diagnostics
}

func assertUpstreamMessage(t *testing.T, actual string, expected any) {
	t.Helper()
	switch expected := expected.(type) {
	case string:
		if actual != expected {
			t.Errorf("message = %q, want %q", actual, expected)
		}
	case map[string]any:
		pattern, _ := expected["__regexp"].(string)
		if pattern != "" {
			compiled, err := regexp.Compile(pattern)
			if err != nil || !compiled.MatchString(actual) {
				t.Errorf("message %q does not match %q", actual, pattern)
			}
		}
	}
}

func assertUpstreamInvalid(t *testing.T, testCase upstreamCase, diagnostics []rule.RuleDiagnostic) {
	t.Helper()
	expectedCount, expectedErrors := upstreamExpectedErrors(t, testCase.Errors)
	if len(diagnostics) != expectedCount {
		t.Fatalf("diagnostic count = %d, want %d\ncode:\n%s\ndiagnostics: %v", len(diagnostics), expectedCount, testCase.Code, diagnostics)
	}
	for index, expected := range expectedErrors {
		actual := diagnostics[index]
		if expected.MessageID != "" && actual.Message.Id != expected.MessageID {
			t.Errorf("diagnostic %d message ID = %q, want %q", index, actual.Message.Id, expected.MessageID)
		}
		if expected.Message != nil {
			assertUpstreamMessage(t, actual.Message.Description, expected.Message)
		}
		start := actual.Range.Pos()
		end := actual.Range.End()
		line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(actual.SourceFile, start)
		endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(actual.SourceFile, end)
		if expected.Line != 0 && expected.Line != line+1 {
			t.Errorf("diagnostic %d line = %d, want %d", index, line+1, expected.Line)
		}
		if expected.Column != 0 && expected.Column != int(column)+1 {
			t.Errorf("diagnostic %d column = %d, want %d", index, column+1, expected.Column)
		}
		if expected.EndLine != 0 && expected.EndLine != endLine+1 {
			t.Errorf("diagnostic %d end line = %d, want %d", index, endLine+1, expected.EndLine)
		}
		if expected.EndColumn != 0 && expected.EndColumn != int(endColumn)+1 {
			t.Errorf("diagnostic %d end column = %d, want %d", index, endColumn+1, expected.EndColumn)
		}
		expectedSuggestionCount, expectedSuggestions := upstreamSuggestionCount(t, expected.Suggestions)
		actualSuggestionCount := 0
		if actual.Suggestions != nil {
			actualSuggestionCount = len(*actual.Suggestions)
		}
		if actualSuggestionCount != expectedSuggestionCount {
			t.Errorf("diagnostic %d suggestion count = %d, want %d", index, actualSuggestionCount, expectedSuggestionCount)
			continue
		}
		for suggestionIndex, expectedSuggestion := range expectedSuggestions {
			actualSuggestion := (*actual.Suggestions)[suggestionIndex]
			if actualSuggestion.Message.Id != expectedSuggestion.MessageID {
				t.Errorf("suggestion %d message ID = %q, want %q", suggestionIndex, actualSuggestion.Message.Id, expectedSuggestion.MessageID)
			}
			output, _, _ := linter.ApplyRuleFixes(testCase.Code, []rule.RuleSuggestion{actualSuggestion})
			if output != expectedSuggestion.Output {
				t.Errorf("suggestion %d output = %q, want %q", suggestionIndex, output, expectedSuggestion.Output)
			}
		}
	}

	wantOutput := ""
	hasOutput := len(testCase.Output) > 0 && string(testCase.Output) != "null"
	if hasOutput {
		if err := json.Unmarshal(testCase.Output, &wantOutput); err != nil {
			t.Fatalf("decode expected output: %v", err)
		}
	}
	actualOutput, _, fixed := linter.ApplyRuleFixes(testCase.Code, diagnostics)
	if hasOutput {
		if !fixed || actualOutput != wantOutput {
			t.Errorf("autofix output = %q (fixed=%t), want %q", actualOutput, fixed, wantOutput)
		}
	} else if fixed {
		t.Errorf("unexpected autofix output %q", actualOutput)
	}
}

func TestNameReplacementsUpstreamV77(t *testing.T) {
	data, err := os.ReadFile("testdata/name_replacements_v77.json")
	if err != nil {
		t.Fatal(err)
	}
	var groups []upstreamGroup
	if err := json.Unmarshal(data, &groups); err != nil {
		t.Fatal(err)
	}

	caseIndex := 0
	for groupIndex, group := range groups {
		t.Run(fmt.Sprintf("group-%02d-%s", groupIndex, group.Kind), func(t *testing.T) {
			// rslint's native parser currently accepts JavaScript and TypeScript
			// source files. Keep Unicorn's Vue, CSS, HTML, JSON, YAML, TOML, and
			// Markdown cases in the fixture as explicit skipped upstream groups.
			if group.Kind == "vue" || group.Kind == "non-javascript" {
				t.Skip("native rslint does not parse this upstream language")
			}
			typescript := group.Kind == "typescript"
			tsx := group.Cases.TesterOptions.LanguageOptions.ParserOptions.ECMAFeatures.JSX
			for validIndex, raw := range group.Cases.Valid {
				testCase, err := decodeUpstreamCase(raw)
				if err != nil {
					t.Fatal(err)
				}
				index := caseIndex
				caseIndex++
				t.Run("valid-"+strconv.Itoa(validIndex), func(t *testing.T) {
					if testCase.Language != "" {
						t.Skip("native rslint does not parse this upstream language")
					}
					if unsupportedNativeTestFilename(testCase.Filename) {
						t.Skip("the native test program cannot load this non-code or hidden filename")
					}
					normalized, unsupported := normalizeRegexpOptions(testCase.Options)
					if unsupported {
						t.Skip("native configuration cannot represent a JavaScript RegExp with flags")
					}
					testCase.Options = normalized
					diagnostics := lintUpstreamCase(t, testCase, group.Cases.TesterOptions.LanguageOptions.Globals, typescript, tsx, index)
					if len(diagnostics) != 0 {
						t.Fatalf("expected no diagnostics, got %v\ncode:\n%s", diagnostics, testCase.Code)
					}
				})
			}
			for invalidIndex, raw := range group.Cases.Invalid {
				testCase, err := decodeUpstreamCase(raw)
				if err != nil {
					t.Fatal(err)
				}
				index := caseIndex
				caseIndex++
				t.Run("invalid-"+strconv.Itoa(invalidIndex), func(t *testing.T) {
					if testCase.Language != "" {
						t.Skip("native rslint does not parse this upstream language")
					}
					if unsupportedNativeTestFilename(testCase.Filename) {
						t.Skip("the native test program cannot load this non-code or hidden filename")
					}
					normalized, unsupported := normalizeRegexpOptions(testCase.Options)
					if unsupported {
						t.Skip("native configuration cannot represent a JavaScript RegExp with flags")
					}
					testCase.Options = normalized
					diagnostics := lintUpstreamCase(t, testCase, group.Cases.TesterOptions.LanguageOptions.Globals, typescript, tsx, index)
					assertUpstreamInvalid(t, testCase, diagnostics)
				})
			}
		})
	}
}
