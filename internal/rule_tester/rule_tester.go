package rule_tester

import (
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"gotest.tools/v3/assert"
)

type ValidTestCase struct {
	Code     string                 `json:"code"`
	FileName string                 `json:"fileName"`
	Only     bool                   `json:"only"`
	Skip     bool                   `json:"skip"`
	Options  any                    `json:"options"`
	Settings map[string]interface{} `json:"settings"`
	// LanguageOptions is the normalized per-file language configuration exposed
	// through ctx.LanguageOptions and used to construct native rule globals. Its
	// zero value means latest.
	LanguageOptions rule.LanguageOptions `json:"languageOptions"`
	// Globals simulates a config-declared `languageOptions.globals` for rules
	// that read ctx.Globals (e.g. no-undef). Values are authored exactly as in
	// an ESLint config — "readonly", "writable", "off", or their aliases — and
	// an unrecognized one fails the test. Inline `/* global */` comments in
	// Code are picked up automatically by the linter and merged with this — no
	// separate field needed for those.
	Globals  map[string]any `json:"globals"`
	TSConfig string         `json:"tsConfig"`
	Tsx      bool           `json:"tsx"`
}

type InvalidTestCaseError struct {
	MessageId string `json:"messageId"`
	// Message, if non-empty, must match the diagnostic's formatted message
	// exactly. Leave empty to skip the text assertion.
	Message     string                      `json:"message,omitempty"`
	Line        int                         `json:"line"`
	Column      int                         `json:"column"`
	EndLine     int                         `json:"endLine"`
	EndColumn   int                         `json:"endColumn"`
	Suggestions []InvalidTestCaseSuggestion `json:"suggestions"`
}

type InvalidTestCaseSuggestion struct {
	MessageId string `json:"messageId"`
	Output    string `json:"output"`
}

type InvalidTestCase struct {
	Code     string                 `json:"code"`
	FileName string                 `json:"fileName"`
	Only     bool                   `json:"only"`
	Skip     bool                   `json:"skip"`
	Output   []string               `json:"output"`
	Errors   []InvalidTestCaseError `json:"errors"`
	Settings map[string]interface{} `json:"settings"`
	// LanguageOptions is the normalized per-file language configuration exposed
	// through ctx.LanguageOptions and used to construct native rule globals. Its
	// zero value means latest.
	LanguageOptions rule.LanguageOptions `json:"languageOptions"`
	// Globals simulates a config-declared `languageOptions.globals` for rules
	// that read ctx.Globals (e.g. no-undef). Values are authored exactly as in
	// an ESLint config — "readonly", "writable", "off", or their aliases — and
	// an unrecognized one fails the test. Inline `/* global */` comments in
	// Code are picked up automatically by the linter and merged with this — no
	// separate field needed for those.
	Globals  map[string]any `json:"globals"`
	TSConfig string         `json:"tsConfig"`
	Options  any            `json:"options"`
	Tsx      bool           `json:"tsx"`
}

// TestSuite represents a complete test suite that can be loaded from JSON
type TestSuite struct {
	Valid   []ValidTestCase   `json:"valid"`
	Invalid []InvalidTestCase `json:"invalid"`
}

var defaultTestFileID atomic.Uint64

func resolveTestCaseFileName(fileName string, tsx bool) string {
	if fileName != "" {
		return fileName
	}

	extension := ".ts"
	if tsx {
		extension = ".tsx"
	}
	return "filename-" + strconv.FormatUint(defaultTestFileID.Add(1), 10) + extension
}

// ResolveTestCaseOptions puts a test case's options through the same step a
// real config goes through before the rule ever runs: the normalized options
// array is validated against the rule's declared schema, and schema-declared
// `default`s are filled in. Options a real config would reject fail the test
// here, so a test case can't lock in a configuration no user can reach; and a
// rule whose defaults live in its schema sees them in tests too, exactly as it
// would at runtime.
//
// Ad-hoc rules constructed directly in tests may declare no schema; they run
// whatever the test case passes, unvalidated.
//
// It is exported so test harnesses that call a rule's Run directly instead of
// going through [RunRuleTester] (e.g. to inspect diagnostics across multiple
// files, or under different [rule.EditDemand]s) can still exercise the same
// schema validation an ordinary config would apply.
//
// The options are deep-copied before validation because [rule.Schema.Validate]
// fills defaults in place, the way ajv's `useDefaults` does. Test cases are
// package level values shared by every case [RunRuleTester] runs, and those
// run in parallel, so mutating the caller's own maps would be a data race.
func ResolveTestCaseOptions(t *testing.T, r *rule.Rule, rawOptions any) []any {
	if r.Schema == nil {
		return rule.NormalizeOptions(rawOptions)
	}
	options := rule.NormalizeOptions(rule.DeepCopyJSON(rawOptions))
	if err := r.Schema.Validate(options); err != nil {
		t.Fatalf("test case options %#v are rejected by %s's schema: %v", rawOptions, r.Name, err)
	}
	return options
}

// resolveTestCaseGlobals normalizes a test case's ESLint-authored globals the
// way the config loader normalizes `languageOptions.globals`.
func resolveTestCaseGlobals(t *testing.T, rawGlobals map[string]any) map[string]utils.GlobalAccess {
	if len(rawGlobals) == 0 {
		return nil
	}
	globals := make(map[string]utils.GlobalAccess, len(rawGlobals))
	for name, value := range rawGlobals {
		access, ok := utils.NormalizeGlobalAccess(value)
		if !ok {
			t.Fatalf("test case global %q has invalid access %#v; expected \"readonly\", \"writable\", or \"off\"", name, value)
		}
		globals[name] = access
	}
	return globals
}

func resolveTestCaseLanguageOptions(t *testing.T, languageOptions rule.LanguageOptions) rule.LanguageOptions {
	if languageOptions.SourceType != "" && !rule.IsValidSourceType(languageOptions.SourceType) {
		t.Fatalf("test case sourceType %q is invalid; expected module, script, or commonjs", languageOptions.SourceType)
	}
	return languageOptions
}

func RunRuleTester(root Root, tsconfigPath string, t *testing.T, r *rule.Rule, validTestCases []ValidTestCase, invalidTestCases []InvalidTestCase) {
	t.Parallel()

	onlyMode := slices.ContainsFunc(validTestCases, func(c ValidTestCase) bool { return c.Only }) ||
		slices.ContainsFunc(invalidTestCases, func(c InvalidTestCase) bool { return c.Only })
	if onlyMode {
		t.Fatal("focused rule test cases (Only) are not allowed; use go test -run to select tests")
	}
	if len(validTestCases)+len(invalidTestCases) == 0 {
		t.Fatal("rule test suite must not be empty")
	}

	runLinter := func(t *testing.T, code string, rawOptions any, settings map[string]interface{}, languageOptions rule.LanguageOptions, rawGlobals map[string]any, tsconfigPathOverride string, fileName string) []rule.RuleDiagnostic {
		options := ResolveTestCaseOptions(t, r, rawOptions)
		globals := resolveTestCaseGlobals(t, rawGlobals)
		languageOptions = resolveTestCaseLanguageOptions(t, languageOptions)

		var diagnosticsMu sync.Mutex
		diagnostics := make([]rule.RuleDiagnostic, 0, 3)

		fs := utils.NewOverlayVFS(root.FS, map[string]string{tspath.ResolvePath(root.Dir, fileName): code})
		host := utils.CreateCompilerHost(root.Dir, fs)

		tsconfigPath := tsconfigPath
		if tsconfigPathOverride != "" {
			tsconfigPath = tsconfigPathOverride
		}

		program, err := utils.CreateProgram(true, fs, root.Dir, tsconfigPath, host)
		assert.NilError(t, err, "couldn't create program. code: "+code)

		sourceFile := program.GetSourceFile(fileName)
		allowedFiles := []string{sourceFile.FileName()}
		programs := []*lintprogram.Program{lintprogram.NewFromCompiler(program)}
		lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
			Programs:         programs,
			TargetsByProgram: [][]string{allowedFiles},
			SingleThreaded:   true,
			GetRulesForFile: func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{
					{
						Name: "test",
						Environment: &rule.RuleEnvironment{
							Settings:        settings,
							LanguageOptions: languageOptions,
							Globals:         globals,
						},
						Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return r.Run(ctx, options)
						},
					},
				}
			},
		})
		assert.NilError(t, err, "error preparing lint plan. code:\n", code)

		_, err = linter.RunLinter(linter.RunLinterOptions{
			SingleThreaded:  true,
			LintPlan:        lintPlan,
			ProjectAnalysis: true,
			Consumer: rule.DiagnosticConsumer{
				Demand: rule.EditDemandAll,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnosticsMu.Lock()
					defer diagnosticsMu.Unlock()

					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})

		assert.NilError(t, err, "error running linter. code:\n", code)

		return diagnostics
	}

	for i, testCase := range validTestCases {
		t.Run("valid-"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			if testCase.Skip {
				t.SkipNow()
			}

			fileName := resolveTestCaseFileName(testCase.FileName, testCase.Tsx)

			diagnostics := runLinter(t, testCase.Code, testCase.Options, testCase.Settings, testCase.LanguageOptions, testCase.Globals, testCase.TSConfig, fileName)
			if len(diagnostics) != 0 {
				// TODO: pretty errors
				t.Errorf("Expected valid test case not to contain errors. Code:\n%v", testCase.Code)
				for i, d := range diagnostics {
					t.Errorf("error %v - (%v-%v) %v", i+1, d.Range.Pos(), d.Range.End(), d.Message.Description)
				}
				t.FailNow()
			}
		})
	}

	for i, testCase := range invalidTestCases {
		t.Run("invalid-"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			if testCase.Skip {
				t.SkipNow()
			}

			var initialDiagnostics []rule.RuleDiagnostic
			outputs := make([]string, 0, 1)
			code := testCase.Code

			fileName := resolveTestCaseFileName(testCase.FileName, testCase.Tsx)

			for i := range 10 {
				diagnostics := runLinter(t, code, testCase.Options, testCase.Settings, testCase.LanguageOptions, testCase.Globals, testCase.TSConfig, fileName)
				if i == 0 {
					initialDiagnostics = diagnostics
				}

				fixedCode, _, fixed := linter.ApplyRuleFixes(code, diagnostics)

				if !fixed {
					break
				}
				code = fixedCode
				outputs = append(outputs, fixedCode)
			}

			if len(testCase.Output) == len(outputs) {
				for i, expected := range testCase.Output {
					assert.Equal(t, expected, outputs[i], "Expected code after fix")
				}
			} else {
				t.Errorf("Expected to have %v outputs but had %v: %v", len(testCase.Output), len(outputs), outputs)
			}

			if len(initialDiagnostics) != len(testCase.Errors) {
				t.Fatalf("Expected invalid test case to contain exactly %v errors (reported %v errors - %v). Code:\n%v", len(testCase.Errors), len(initialDiagnostics), initialDiagnostics, testCase.Code)
			}

			for i, expected := range testCase.Errors {
				diagnostic := initialDiagnostics[i]

				if expected.MessageId != diagnostic.Message.Id {
					t.Errorf("Invalid message id %v. Expected %v", diagnostic.Message.Id, expected.MessageId)
				}

				if expected.Message != "" && expected.Message != diagnostic.Message.Description {
					t.Errorf("Invalid message text %q. Expected %q", diagnostic.Message.Description, expected.Message)
				}
				lineIndex, columnIndex := scanner.GetECMALineAndUTF16CharacterOfPosition(diagnostic.SourceFile, diagnostic.Range.Pos())
				line, column := lineIndex+1, int(columnIndex)+1
				endLineIndex, endColumnIndex := scanner.GetECMALineAndUTF16CharacterOfPosition(diagnostic.SourceFile, diagnostic.Range.End())
				endLine, endColumn := endLineIndex+1, int(endColumnIndex)+1

				if expected.Line != 0 && expected.Line != line {
					t.Errorf("Error line should be %v. Got %v", expected.Line, line)
				}
				if expected.Column != 0 && expected.Column != column {
					t.Errorf("Error column should be %v. Got %v", expected.Column, column)
				}
				if expected.EndLine != 0 && expected.EndLine != endLine {
					t.Errorf("Error end line should be %v. Got %v", expected.EndLine, endLine)
				}
				if expected.EndColumn != 0 && expected.EndColumn != endColumn {
					t.Errorf("Error end column should be %v. Got %v", expected.EndColumn, endColumn)
				}

				suggestionsCount := 0
				if diagnostic.Suggestions != nil {
					suggestionsCount = len(*diagnostic.Suggestions)
				}
				if len(expected.Suggestions) != suggestionsCount {
					t.Errorf("Expected to have %v suggestions but had %v", len(expected.Suggestions), suggestionsCount)
				} else {
					for i, expectedSuggestion := range expected.Suggestions {
						suggestion := (*diagnostic.Suggestions)[i]
						if expectedSuggestion.MessageId != suggestion.Message.Id {
							t.Errorf("Invalid suggestion message id %v. Expected %v", suggestion.Message.Id, expectedSuggestion.MessageId)
						} else {
							output, _, _ := linter.ApplyRuleFixes(testCase.Code, []rule.RuleSuggestion{suggestion})

							assert.Equal(t, expectedSuggestion.Output, output, "Expected code after suggestion fix")
						}
					}
				}
			}
		})
	}
}
