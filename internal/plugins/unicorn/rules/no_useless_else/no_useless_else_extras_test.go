package no_useless_else_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_else"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessElseExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{FileName: "case.js", Code: `function run(enabled, values) {
  if (enabled) {
    try {
      [...values, process.exit()];
    } catch {}
  } else {
    fallback();
  }
}`},
		{FileName: "case.js", Code: `function run(enabled) {
  if (enabled) {
    try {
      process.exit(1n / 0n);
    } catch {}
  } else {
    fallback();
  }
}`},
		{FileName: "case.js", Code: `function run(enabled, operation) {
  if (enabled) {
    try {
      try {
        throw 0;
      } catch {
        operation();
        return;
      } finally {}
    } catch {}
  } else {
    fallback();
  }
}`},
		{FileName: "case.js", Code: `outer: while (condition) {
  if (ready) {
    inner: {
      try {
        break outer;
      } finally {
        break inner;
      }
    }
  } else {
    fallback();
  }
}`},
		{FileName: "case.js", Code: `function run(ready) {
  if (ready) {
    return;
  }
  throw 0;
  if (ready) {
    throw 1;
  } else {
    fallback();
  }
}`},
		{FileName: "case.js", Code: `function run(ready) {
  if (ready) {
    class C {
      method() { process.exit(); }
      value = process.exit();
    }
  } else {
    fallback();
  }
}`},
		{FileName: "case.js", Code: `function run(ready, callback) {
  if (ready) {
    (callback?.handle)?.(process.exit());
  } else {
    fallback();
  }
}`},
		{FileName: "case.js", Code: `function run(ready) {
  if (ready) {
    process.exit();
  } else {
    fallback();
  }
}`, Globals: map[string]any{"process": "off"}},
		{FileName: "case.js", Code: `with (scope) {
  if (ready) {
    process.exit();
  } else {
    fallback();
  }
}`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{FileName: "case.ts", Code: `namespace N {
  if (ready) {
    throw new Error();
  } else {
    fallback();
  }
}`},
	}
	invalid := []rule_tester.InvalidTestCase{
		{
			FileName: "case.js", Code: `function run(ready) {
  if (ready) {
    inner: {
      try {
        break inner;
      } finally {
        return;
      }
    }
  } else {
    fallback();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(10, 5)},
			Output: []string{`function run(ready) {
  if (ready) {
    inner: {
      try {
        break inner;
      } finally {
        return;
      }
    }
  }
  fallback();
}`},
		},
		{
			FileName: "case.ts", Code: `function run(ready: boolean) {
  if (ready) {
    (process.exit as (code: number) => never)(1);
  } else {
    fallback();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 5)},
			Output: []string{`function run(ready: boolean) {
  if (ready) {
    (process.exit as (code: number) => never)(1);
  }
  fallback();
}`},
		},
		{
			FileName: "case.js", Code: `if (ready) {
  class C {
    static { process.exit(); }
  }
} else {
  fallback();
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(5, 3)},
			Output: []string{`if (ready) {
  class C {
    static { process.exit(); }
  }
}
fallback();`},
		},
		{
			FileName: "case.js", Code: `function run(ready) {
  if (ready) {
    return;
  } else {
    { let value = 1; use(value); }
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 5)},
			Output: []string{`function run(ready) {
  if (ready) {
    return;
  }
  { let value = 1; use(value); }
}`},
		},
		{
			FileName: "case.ts", Code: `function run(ready: boolean) {
  if (ready) {
    return;
  } else {
    using resource = acquire();
    work(resource);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 5)},
		},
		{
			FileName: "case.ts", Code: `async function run(ready: boolean) {
  if (ready) {
    return;
  } else {
    await using resource = acquire();
    work(resource);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 5)},
		},
		{
			FileName: "case.js", Code: `function run(ready) {
  if (ready) { return; } else {
    // eslint-disable-next-line no-alert
  }
  alert();
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(2, 26)},
		},
		{
			FileName: "case.js", Code: "if (ready) { throw \"😀\"; } else { fallback(); }\u2028end();",
			Errors: []rule_tester.InvalidTestCaseError{elseAt(1, 28)},
			Output: []string{"if (ready) { throw \"😀\"; }\u2028fallback();\u2028end();"},
		},
		{
			FileName: "case.js", Code: `function run(ready) {
  if (ready) { return; } else {
    value = get()
  }
  (finish)();
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(2, 26)},
		},
		{
			FileName: "case.js", Code: "function run(ready) {\n  if (ready) throw error\n  else `side effect`\n}",
			Errors: []rule_tester.InvalidTestCaseError{elseAt(3, 3)},
		},
		{
			FileName: "case.js", Code: `function run(ready) {
  if (ready) { return; } else { fallback(); } /* keep */
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(2, 26)},
		},
		{
			FileName: "case.js", Code: `function run(ready) {
  if (ready) { return; } else {
    // Keep this comment.
  }
  work();
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(2, 26)},
			Output: []string{`function run(ready) {
  if (ready) { return; }
  // Keep this comment.
  work();
}`},
		},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_else.NoUselessElseRule, valid, invalid)
}

func TestNoUselessElseEditDemand(t *testing.T) {
	const source = "function run(ready) {\n  if (ready) { return; } else { work(); }\n}"
	const fixedSource = "function run(ready) {\n  if (ready) { return; }\n  work();\n}"
	r := &no_useless_else.NoUselessElseRule
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	compilerProgram, file, err := helper.CreateTestProgram(source, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	program := lintprogram.NewFromCompiler(compilerProgram)
	var baseline rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: program,
			File:    file.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: r.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return r.Run(ctx, nil)
				}}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			}},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: expected one diagnostic, got %v", demand, diagnostics)
		}
		got := diagnostics[0]
		start := strings.Index(source, "else")
		if got.Range.Pos() != start || got.Range.End() != start+len("else") ||
			got.Message.Id != "no-useless-else" || got.RuleName != r.Name ||
			got.Message.Description != "Unexpected `else` after a statement that exits." ||
			got.Severity != rule.SeverityError || got.Suggestions != nil {
			t.Errorf("demand %d: unexpected diagnostic: %+v", demand, got)
		}
		if demand == rule.EditDemandAll {
			baseline = got
		}
		wantFix := demand&rule.EditDemandAutofix != 0
		want := source
		if wantFix {
			want = fixedSource
			if !reflect.DeepEqual(got.Fixes(), baseline.Fixes()) {
				t.Errorf("demand %d: got fixes %v, want %v", demand, got.Fixes(), baseline.Fixes())
			}
		}
		output, _, applied := linter.ApplyRuleFixes(source, diagnostics)
		if applied != wantFix || output != want {
			t.Errorf("demand %d: got %q (%t), want %q (%t)", demand, output, applied, want, wantFix)
		}
	}
}
