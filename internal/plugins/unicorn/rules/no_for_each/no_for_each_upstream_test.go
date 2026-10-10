// Ported from eslint-plugin-unicorn v77.0.0 test/no-for-each.js.
// cspell:ignore callbag
package no_for_each_test

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_for_each"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	errorMessageID      = "no-for-each/error"
	suggestionMessageID = "no-for-each/suggestion"
)

func valid(code, fileName string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: fileName}
}

func forEachError(code string, occurrence int) rule_tester.InvalidTestCaseError {
	start := 0
	offset := -1
	for range occurrence + 1 {
		relative := strings.Index(code[start:], "forEach")
		if relative < 0 {
			panic("forEach not found in test case")
		}
		offset = start + relative
		start = offset + len("forEach")
	}
	prefix := code[:offset]
	line := strings.Count(prefix, "\n") + 1
	lastNewline := strings.LastIndex(prefix, "\n")
	column := offset + 1
	if lastNewline >= 0 {
		column = offset - lastNewline
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: errorMessageID,
		Message:   "Use `for…of` instead of `.forEach(…)`.",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len("forEach"),
	}
}

func invalid(code, fileName string, output string) rule_tester.InvalidTestCase {
	testCase := rule_tester.InvalidTestCase{
		Code:     code,
		FileName: fileName,
		Errors:   []rule_tester.InvalidTestCaseError{forEachError(code, 0)},
	}
	if output != "" {
		testCase.Output = []string{output}
	}
	return testCase
}

//go:embed testdata/no_for_each_v77.json
var upstreamManifestJSON []byte

type upstreamManifest struct {
	Upstream string                 `json:"upstream"`
	Commit   string                 `json:"commit"`
	Valid    []upstreamManifestCase `json:"valid"`
	Invalid  []upstreamManifestCase `json:"invalid"`
}

type upstreamManifestCase struct {
	ID           string                  `json:"id"`
	Group        string                  `json:"group"`
	Code         string                  `json:"code"`
	FileName     string                  `json:"filename"`
	SourceType   string                  `json:"sourceType"`
	TypeScript   bool                    `json:"typescript"`
	GlobalReturn bool                    `json:"globalReturn"`
	Outputs      []string                `json:"outputs"`
	Errors       []upstreamManifestError `json:"errors"`
}

type upstreamManifestError struct {
	MessageID   string                       `json:"messageId"`
	Message     string                       `json:"message"`
	Line        int                          `json:"line"`
	Column      int                          `json:"column"`
	EndLine     int                          `json:"endLine"`
	EndColumn   int                          `json:"endColumn"`
	Suggestions []upstreamManifestSuggestion `json:"suggestions"`
}

type upstreamManifestSuggestion struct {
	MessageID string `json:"messageId"`
	Output    string `json:"output"`
}

func TestNoForEachUpstream(t *testing.T) {
	var manifest upstreamManifest
	if err := json.Unmarshal(upstreamManifestJSON, &manifest); err != nil {
		t.Fatalf("decode upstream manifest: %v", err)
	}
	if manifest.Upstream != "eslint-plugin-unicorn@77.0.0" ||
		manifest.Commit != "15f1d646dad6857d1a9cd39939e25cda25d7e4c9" ||
		len(manifest.Valid) != 28 || len(manifest.Invalid) != 348 {
		t.Fatalf("unexpected upstream manifest identity/counts: %q, %d valid, %d invalid",
			manifest.Upstream, len(manifest.Valid), len(manifest.Invalid))
	}
	// ESLint's parserOptions.globalReturn accepts top-level return statements.
	// Rslint follows TypeScript's parser and suppresses native rules on that
	// malformed product input, so the one valid and twelve invalid cases in the
	// dedicated global-return group stay present as explicit skips.
	globalReturnCases := 0
	for _, testCase := range manifest.Valid {
		if testCase.GlobalReturn {
			globalReturnCases++
		}
	}
	for _, testCase := range manifest.Invalid {
		if testCase.GlobalReturn {
			globalReturnCases++
		}
	}
	if globalReturnCases != 13 {
		t.Fatalf("global-return exclusions = %d, want 13", globalReturnCases)
	}

	validCases := make([]rule_tester.ValidTestCase, 0, len(manifest.Valid))
	for _, testCase := range manifest.Valid {
		validCases = append(validCases, rule_tester.ValidTestCase{
			Code:            testCase.Code,
			FileName:        manifestFileName(testCase),
			Skip:            testCase.GlobalReturn,
			LanguageOptions: languageOptions(testCase.SourceType),
		})
	}
	invalidCases := make([]rule_tester.InvalidTestCase, 0, len(manifest.Invalid))
	for _, testCase := range manifest.Invalid {
		errors := make([]rule_tester.InvalidTestCaseError, 0, len(testCase.Errors))
		for _, expected := range testCase.Errors {
			suggestions := make([]rule_tester.InvalidTestCaseSuggestion, 0, len(expected.Suggestions))
			for _, suggestion := range expected.Suggestions {
				suggestions = append(suggestions, rule_tester.InvalidTestCaseSuggestion{
					MessageId: suggestion.MessageID,
					Output:    suggestion.Output,
				})
			}
			errors = append(errors, rule_tester.InvalidTestCaseError{
				MessageId:   expected.MessageID,
				Message:     expected.Message,
				Line:        expected.Line,
				Column:      expected.Column,
				EndLine:     expected.EndLine,
				EndColumn:   expected.EndColumn,
				Suggestions: suggestions,
			})
		}
		invalidCases = append(invalidCases, rule_tester.InvalidTestCase{
			Code:            testCase.Code,
			FileName:        manifestFileName(testCase),
			Skip:            testCase.GlobalReturn,
			Output:          testCase.Outputs,
			Errors:          errors,
			LanguageOptions: languageOptions(testCase.SourceType),
		})
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&no_for_each.NoForEachRule, validCases, invalidCases)
}

func languageOptions(sourceType string) rule.LanguageOptions {
	if sourceType == "" {
		return rule.LanguageOptions{}
	}
	return rule.LanguageOptions{SourceType: sourceType}
}

func manifestFileName(testCase upstreamManifestCase) string {
	if testCase.FileName != "" {
		return testCase.FileName
	}
	if testCase.TypeScript {
		return "file.ts"
	}
	return "file.js"
}
