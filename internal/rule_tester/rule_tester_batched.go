// cspell:ignore batchable subtest
package rule_tester

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"gotest.tools/v3/assert"
)

// RunRuleTesterBatched is a drop-in replacement for RunRuleTester that
// amortizes the cost of building TypeScript programs across every test case.
//
// Mechanism:
//
//  1. Every batchable case (no custom FileName / TSConfig) is written into a
//     single shared overlay VFS as a uniquely-named file under rootDir.
//  2. We build ONE program containing the lib + all case files.
//  3. We run linter.RunLinter ONCE with Scope.Dirs = [rootDir]. This uses
//     the framework's existing fast path (isDirAllowed — pure prefix check,
//     no os.Stat), so the per-call cost stays O(N) instead of the O(N²)
//     blow-up that Scope.Files triggers via its os.Stat fallback.
//  4. OnDiagnostic routes each diagnostic into a map keyed by SourceFile
//     path. Each subtest then looks up its own diagnostics from this map.
//
// Falls back to per-case standalone build for cases with custom FileName /
// TSConfig, and for the 2nd+ iteration of auto-fix loops (post-fix source
// no longer matches the overlay).
func RunRuleTesterBatched(rootDir string, tsconfigPath string, t *testing.T, r *rule.Rule, validTestCases []ValidTestCase, invalidTestCases []InvalidTestCase) {
	t.Parallel()

	onlyMode := slices.ContainsFunc(validTestCases, func(c ValidTestCase) bool { return c.Only }) ||
		slices.ContainsFunc(invalidTestCases, func(c InvalidTestCase) bool { return c.Only })

	type caseRef struct {
		kind     string // "v" or "i"
		idx      int
		options  any
		settings map[string]interface{}
	}

	overlay := make(map[string]string)
	pathToCase := make(map[string]caseRef)
	validPaths := make([]string, len(validTestCases))
	invalidPaths := make([]string, len(invalidTestCases))

	for i, c := range validTestCases {
		if c.FileName != "" || c.TSConfig != "" {
			continue
		}
		p := batchedFilePath(rootDir, "v", i, c.Tsx)
		overlay[p] = c.Code
		validPaths[i] = p
		pathToCase[p] = caseRef{kind: "v", idx: i, options: c.Options, settings: c.Settings}
	}
	for i, c := range invalidTestCases {
		if c.FileName != "" || c.TSConfig != "" {
			continue
		}
		p := batchedFilePath(rootDir, "i", i, c.Tsx)
		overlay[p] = c.Code
		invalidPaths[i] = p
		pathToCase[p] = caseRef{kind: "i", idx: i, options: c.Options, settings: c.Settings}
	}

	diagsByPath := make(map[string][]rule.RuleDiagnostic)
	sharedLintOk := false

	if len(overlay) > 0 {
		fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), overlay)
		host := utils.CreateCompilerHost(rootDir, fs)
		program, err := utils.CreateProgram(true, fs, rootDir, tsconfigPath, host)
		if err != nil {
			t.Logf("RunRuleTesterBatched: shared program build failed, falling back per-case: %v", err)
		} else {
			var mu sync.Mutex
			_, lintErr := linter.RunLinter(linter.RunLinterOptions{
				Programs:       []*compiler.Program{program},
				SingleThreaded: true,
				// Scope.Dirs uses the framework's tspath.StartsWithDirectory
				// prefix check — no os.Stat fallback, so walk cost is O(N)
				// over program files rather than O(N) per case.
				Scope:        linter.FileScope{Dirs: []string{rootDir}},
				ExcludePaths: []string{},
				GetRulesForFile: func(sf *ast.SourceFile) []linter.ConfiguredRule {
					ref, ok := pathToCase[sf.FileName()]
					if !ok {
						// Real fixture files (empty react.tsx / file.ts)
						// or anything else under rootDir that's not one of
						// our virtual cases — no rule, no diagnostics.
						return nil
					}
					caseOptions := ref.options
					caseSettings := ref.settings
					return []linter.ConfiguredRule{{
						Name:     "test",
						Settings: caseSettings,
						Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return r.Run(ctx, caseOptions)
						},
					}}
				},
				OnDiagnostic: func(d rule.RuleDiagnostic) {
					mu.Lock()
					defer mu.Unlock()
					p := d.SourceFile.FileName()
					diagsByPath[p] = append(diagsByPath[p], d)
				},
			})
			assert.NilError(t, lintErr, "error running shared linter")
			sharedLintOk = true
		}
	}

	canBatchValid := func(i int, c ValidTestCase) bool {
		return sharedLintOk && validPaths[i] != "" && c.TSConfig == "" && c.FileName == ""
	}
	canBatchInvalid := func(i int, c InvalidTestCase) bool {
		return sharedLintOk && invalidPaths[i] != "" && c.TSConfig == "" && c.FileName == ""
	}

	lintStandalone := func(t *testing.T, code string, options any, settings map[string]interface{}, tsconfigOverride, fileName string) []rule.RuleDiagnostic {
		var mu sync.Mutex
		out := make([]rule.RuleDiagnostic, 0, 3)

		fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
		host := utils.CreateCompilerHost(rootDir, fs)

		tsc := tsconfigPath
		if tsconfigOverride != "" {
			tsc = tsconfigOverride
		}

		program, err := utils.CreateProgram(true, fs, rootDir, tsc, host)
		assert.NilError(t, err, "couldn't create program. code: "+code)

		sf := program.GetSourceFile(fileName)
		_, err = linter.RunLinter(linter.RunLinterOptions{
			Programs:       []*compiler.Program{program},
			SingleThreaded: true,
			Scope:          linter.FileScope{Files: []string{sf.FileName()}},
			ExcludePaths:   []string{},
			GetRulesForFile: func(sf *ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     "test",
					Settings: settings,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.Run(ctx, options)
					},
				}}
			},
			OnDiagnostic: func(d rule.RuleDiagnostic) {
				mu.Lock()
				defer mu.Unlock()
				out = append(out, d)
			},
		})
		assert.NilError(t, err, "error running linter (standalone). code:\n", code)
		return out
	}

	for i, testCase := range validTestCases {
		t.Run("valid-"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			if (onlyMode && !testCase.Only) || testCase.Skip {
				t.SkipNow()
			}

			var diagnostics []rule.RuleDiagnostic
			if canBatchValid(i, testCase) {
				diagnostics = diagsByPath[validPaths[i]]
			} else {
				fileName := defaultFileName(testCase.Tsx, testCase.FileName)
				diagnostics = lintStandalone(t, testCase.Code, testCase.Options, testCase.Settings, testCase.TSConfig, fileName)
			}

			if len(diagnostics) != 0 {
				t.Errorf("Expected valid test case not to contain errors. Code:\n%v", testCase.Code)
				for j, d := range diagnostics {
					t.Errorf("error %v - (%v-%v) %v", j+1, d.Range.Pos(), d.Range.End(), d.Message.Description)
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

			fileName := defaultFileName(testCase.Tsx, testCase.FileName)

			var initialDiagnostics []rule.RuleDiagnostic
			outputs := make([]string, 0, 1)
			code := testCase.Code

			for it := range 10 {
				var diagnostics []rule.RuleDiagnostic
				if it == 0 && canBatchInvalid(i, testCase) {
					diagnostics = diagsByPath[invalidPaths[i]]
				} else {
					diagnostics = lintStandalone(t, code, testCase.Options, testCase.Settings, testCase.TSConfig, fileName)
				}
				if it == 0 {
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
				for j, expected := range testCase.Output {
					assert.Equal(t, expected, outputs[j], "Expected code after fix")
				}
			} else {
				t.Errorf("Expected to have %v outputs but had %v: %v", len(testCase.Output), len(outputs), outputs)
			}

			if len(initialDiagnostics) != len(testCase.Errors) {
				t.Fatalf("Expected invalid test case to contain exactly %v errors (reported %v errors - %v). Code:\n%v", len(testCase.Errors), len(initialDiagnostics), initialDiagnostics, testCase.Code)
			}

			for ei, expected := range testCase.Errors {
				diagnostic := initialDiagnostics[ei]

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
					for si, expectedSuggestion := range expected.Suggestions {
						suggestion := (*diagnostic.Suggestions)[si]
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

func batchedFilePath(rootDir, kind string, i int, tsx bool) string {
	ext := ".ts"
	if tsx {
		ext = ".tsx"
	}
	return tspath.ResolvePath(rootDir, fmt.Sprintf("__batched_%s_%d%s", kind, i, ext))
}

func defaultFileName(tsx bool, override string) string {
	if override != "" {
		return override
	}
	if tsx {
		return "react.tsx"
	}
	return "file.ts"
}
