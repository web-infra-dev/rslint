package rule_tester

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"gotest.tools/v3/assert"
)

type ValidTestCase struct {
	Code     string
	FileName string
	Only     bool
	Skip     bool
	Options  any
	TSConfig string
	Tsx      bool
}

type InvalidTestCaseError struct {
	MessageId   string
	Line        int
	Column      int
	EndLine     int
	EndColumn   int
	Suggestions []InvalidTestCaseSuggestion
}

type InvalidTestCaseSuggestion struct {
	MessageId string
	Output    string
}

type InvalidTestCase struct {
	Code     string
	FileName string
	Only     bool
	Skip     bool
	Output   []string
	Errors   []InvalidTestCaseError
	TSConfig string
	Options  any
	Tsx      bool
}

// TestSuite represents a complete test suite that can be loaded from JSON
type TestSuite struct {
	Valid   []ValidTestCase   `json:"valid"`
	Invalid []InvalidTestCase `json:"invalid"`
}

// ESLintTestCase represents test cases from ESLint format
type ESLintTestCase struct {
	Code     string                 `json:"code"`
	Filename string                 `json:"filename,omitempty"`
	Options  []interface{}          `json:"options,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	Only     bool                   `json:"only,omitempty"`
	Skip     bool                   `json:"skip,omitempty"`
	Parser   string                 `json:"parser,omitempty"`
}

// ESLintInvalidTestCase represents invalid test cases from ESLint format
type ESLintInvalidTestCase struct {
	ESLintTestCase
	Output      string              `json:"output,omitempty"`
	Errors      []ESLintError       `json:"errors"`
	Suggestions []ESLintSuggestion  `json:"suggestions,omitempty"`
}

// ESLintError represents an error in ESLint format
type ESLintError struct {
	Message     string              `json:"message,omitempty"`
	MessageId   string              `json:"messageId,omitempty"`
	Type        string              `json:"type,omitempty"`
	Line        int                 `json:"line,omitempty"`
	Column      int                 `json:"column,omitempty"`
	EndLine     int                 `json:"endLine,omitempty"`
	EndColumn   int                 `json:"endColumn,omitempty"`
	Suggestions []ESLintSuggestion  `json:"suggestions,omitempty"`
}

// ESLintSuggestion represents a suggestion in ESLint format
type ESLintSuggestion struct {
	MessageId string `json:"messageId,omitempty"`
	Desc      string `json:"desc,omitempty"`
	Output    string `json:"output,omitempty"`
}

// ESLintTestSuite represents a complete ESLint test suite
type ESLintTestSuite struct {
	Valid   []ESLintTestCase        `json:"valid"`
	Invalid []ESLintInvalidTestCase `json:"invalid"`
}

func RunRuleTester(rootDir string, tsconfigPath string, t *testing.T, r *rule.Rule, validTestCases []ValidTestCase, invalidTestCases []InvalidTestCase) {
	t.Parallel()

	onlyMode := slices.ContainsFunc(validTestCases, func(c ValidTestCase) bool { return c.Only }) ||
		slices.ContainsFunc(invalidTestCases, func(c InvalidTestCase) bool { return c.Only })

	runLinter := func(t *testing.T, code string, options any, tsconfigPathOverride string, fileName string) []rule.RuleDiagnostic {
		var diagnosticsMu sync.Mutex
		diagnostics := make([]rule.RuleDiagnostic, 0, 3)

		fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
		host := utils.CreateCompilerHost(rootDir, fs)

		tsconfigPath := tsconfigPath
		if tsconfigPathOverride != "" {
			tsconfigPath = tsconfigPathOverride
		}

		program, err := utils.CreateProgram(true, fs, rootDir, tsconfigPath, host)
		assert.NilError(t, err, "couldn't create program. code: "+code)

		sourceFile := program.GetSourceFile(fileName)
		allowedFiles := []string{sourceFile.FileName()}

		_, err = linter.RunLinter(
			[]*compiler.Program{program},
			true,
			allowedFiles,
			[]string{}, // No files to skip in test environment
			func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{
					{
						Name:     "test",
						Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return r.Run(ctx, options)
						},
					},
				}
			},
			func(diagnostic rule.RuleDiagnostic) {
				diagnosticsMu.Lock()
				defer diagnosticsMu.Unlock()

				diagnostics = append(diagnostics, diagnostic)
			},
		)

		assert.NilError(t, err, "error running linter. code:\n", code)

		return diagnostics
	}

	for i, testCase := range validTestCases {
		t.Run("valid-"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			if (onlyMode && !testCase.Only) || testCase.Skip {
				t.SkipNow()
			}

			fileName := "file.ts"
			if testCase.Tsx {
				fileName = "react.tsx"
			}
			if testCase.FileName != "" {
				fileName = testCase.FileName
			}

			diagnostics := runLinter(t, testCase.Code, testCase.Options, testCase.TSConfig, fileName)
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

			if (onlyMode && !testCase.Only) || testCase.Skip {
				t.SkipNow()
			}

			var initialDiagnostics []rule.RuleDiagnostic
			outputs := make([]string, 0, 1)
			code := testCase.Code

			fileName := "file.ts"
			if testCase.Tsx {
				fileName = "react.tsx"
			}
			if testCase.FileName != "" {
				fileName = testCase.FileName
			}

			for i := range 10 {
				diagnostics := runLinter(t, code, testCase.Options, testCase.TSConfig, fileName)
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

				lineIndex, columnIndex := scanner.GetLineAndCharacterOfPosition(diagnostic.SourceFile, diagnostic.Range.Pos())
				line, column := lineIndex+1, columnIndex+1
				endLineIndex, endColumnIndex := scanner.GetLineAndCharacterOfPosition(diagnostic.SourceFile, diagnostic.Range.End())
				endLine, endColumn := endLineIndex+1, endColumnIndex+1

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

// LoadTestSuiteFromJSON loads test cases from a JSON file
func LoadTestSuiteFromJSON(filePath string) (*TestSuite, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, err
	}

	return &suite, nil
}

// LoadESLintTestSuiteFromJSON loads ESLint-format test cases from a JSON file
func LoadESLintTestSuiteFromJSON(filePath string) (*ESLintTestSuite, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var suite ESLintTestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, err
	}

	return &suite, nil
}

// ConvertESLintTestCase converts an ESLint test case to our internal format
func ConvertESLintTestCase(tc ESLintTestCase) ValidTestCase {
	var options any
	if len(tc.Options) > 0 {
		// If single option, unwrap it; otherwise keep as array
		if len(tc.Options) == 1 {
			options = tc.Options[0]
		} else {
			options = tc.Options
		}
	}

	fileName := "file.ts"
	tsx := false
	if tc.Filename != "" {
		fileName = tc.Filename
		if filepath.Ext(fileName) == ".tsx" || filepath.Ext(fileName) == ".jsx" {
			tsx = true
		}
	}

	return ValidTestCase{
		Code:     tc.Code,
		FileName: fileName,
		Only:     tc.Only,
		Skip:     tc.Skip,
		Options:  options,
		Tsx:      tsx,
	}
}

// ConvertESLintInvalidTestCase converts an ESLint invalid test case to our internal format
func ConvertESLintInvalidTestCase(tc ESLintInvalidTestCase) InvalidTestCase {
	var options any
	if len(tc.Options) > 0 {
		if len(tc.Options) == 1 {
			options = tc.Options[0]
		} else {
			options = tc.Options
		}
	}

	fileName := "file.ts"
	tsx := false
	if tc.Filename != "" {
		fileName = tc.Filename
		if filepath.Ext(fileName) == ".tsx" || filepath.Ext(fileName) == ".jsx" {
			tsx = true
		}
	}

	// Convert errors
	errors := make([]InvalidTestCaseError, len(tc.Errors))
	for i, err := range tc.Errors {
		suggestions := make([]InvalidTestCaseSuggestion, len(err.Suggestions))
		for j, sug := range err.Suggestions {
			suggestions[j] = InvalidTestCaseSuggestion{
				MessageId: sug.MessageId,
				Output:    sug.Output,
			}
		}

		errors[i] = InvalidTestCaseError{
			MessageId:   err.MessageId,
			Line:        err.Line,
			Column:      err.Column,
			EndLine:     err.EndLine,
			EndColumn:   err.EndColumn,
			Suggestions: suggestions,
		}
	}

	// Handle output
	var output []string
	if tc.Output != "" {
		output = []string{tc.Output}
	}

	return InvalidTestCase{
		Code:     tc.Code,
		FileName: fileName,
		Only:     tc.Only,
		Skip:     tc.Skip,
		Output:   output,
		Errors:   errors,
		Options:  options,
		Tsx:      tsx,
	}
}

// ConvertESLintTestSuite converts an entire ESLint test suite to our internal format
func ConvertESLintTestSuite(suite *ESLintTestSuite) *TestSuite {
	valid := make([]ValidTestCase, len(suite.Valid))
	for i, tc := range suite.Valid {
		valid[i] = ConvertESLintTestCase(tc)
	}

	invalid := make([]InvalidTestCase, len(suite.Invalid))
	for i, tc := range suite.Invalid {
		invalid[i] = ConvertESLintInvalidTestCase(tc)
	}

	return &TestSuite{
		Valid:   valid,
		Invalid: invalid,
	}
}

// RunRuleTesterFromJSON loads and runs tests from a JSON file
func RunRuleTesterFromJSON(rootDir string, tsconfigPath string, testFilePath string, t *testing.T, r *rule.Rule) error {
	suite, err := LoadTestSuiteFromJSON(testFilePath)
	if err != nil {
		return err
	}

	RunRuleTester(rootDir, tsconfigPath, t, r, suite.Valid, suite.Invalid)
	return nil
}

// RunRuleTesterFromESLintJSON loads and runs ESLint-format tests from a JSON file
func RunRuleTesterFromESLintJSON(rootDir string, tsconfigPath string, testFilePath string, t *testing.T, r *rule.Rule) error {
	eslintSuite, err := LoadESLintTestSuiteFromJSON(testFilePath)
	if err != nil {
		return err
	}

	suite := ConvertESLintTestSuite(eslintSuite)
	RunRuleTester(rootDir, tsconfigPath, t, r, suite.Valid, suite.Invalid)
	return nil
}
