package valid_expect_with_promise

import (
	"context"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestValidExpectWithPromiseTypeCapability(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/capability.txtar").Materialize(t, ""))
	fs := bundled.WrapFS(osvfs.FS())
	for _, config := range []string{"tsconfig.json", "loose.json"} {
		t.Run(config, func(t *testing.T) {
			raw, err := utils.CreateProgram(true, fs, dir, config, utils.CreateCompilerHost(dir, fs))
			if err != nil {
				t.Fatal(err)
			}
			program := lintprogram.NewFromCompiler(raw)
			file := program.GetSourceFile(tspath.ResolvePath(dir, "entry.ts"))
			if file == nil {
				t.Fatal("missing entry.ts")
			}
			strict := config == "tsconfig.json"
			if !program.Options().Strict.IsTrue() || (config == "loose.json" && !program.Options().StrictNullChecks.IsFalse()) {
				t.Fatal("compiler options were not preserved")
			}
			seen := map[string]bool{}
			probe := rule.ConfiguredRule{Name: "type-capability", RequiresTypeInfo: true, Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
						call := node.AsCallExpression()
						if !ast.IsIdentifier(call.Expression) || call.Expression.Text() != "probe" {
							return
						}
						name, subject := call.Arguments.Nodes[0].Text(), call.Arguments.Nodes[1]
						seen[name] = true
						typ := ctx.TypeChecker.GetTypeAtLocation(subject)
						wantPromise := name == "external" || name == "derived" || name == "ambient" || name == "narrowed" || name == "constraint" || (name == "nullable" && !strict)
						if got := utils.IsPromiseLike(ctx.Program(), ctx.TypeChecker, typ); got != wantPromise {
							t.Errorf("%s IsPromiseLike = %v, want %v", name, got, wantPromise)
						}
						if name == "thenable" || name == "chainable" {
							if got := isStrictThenable(ctx.TypeChecker, subject, typ); got != (name == "thenable") {
								t.Errorf("%s strict thenable = %v", name, got)
							}
							if !utils.IsThenableType(ctx.TypeChecker, subject, typ) {
								t.Errorf("%s is not recognized by existing thenable helper", name)
							}
						}
						if name == "overload" || name == "ambientOverload" {
							returned := callableReturnType(ctx.TypeChecker, subject, typ)
							if returned.Flags()&checker.TypeFlagsNumber == 0 {
								t.Errorf("%s no-argument return is not number", name)
							}
						}
						if name == "constraint" && typ.Flags()&checker.TypeFlagsTypeParameter == 0 {
							t.Error("lost type parameter")
						}
						if name == "nullable" && utils.IsUnionType(typ) != strict {
							t.Error("strictNullChecks did not affect union")
						}
						if name == "awaited" && typ.Flags()&checker.TypeFlagsNumber == 0 {
							t.Error("awaited value is not number")
						}
						if name == "constructor" || name == "derived" {
							if got := utils.IsPromiseConstructorLike(ctx.Program(), ctx.TypeChecker, typ); got != (name == "constructor") {
								t.Errorf("%s constructor predicate = %v", name, got)
							}
						}
						if name == "derived" {
							symbol := checker.Type_symbol(typ)
							if symbol == nil || len(symbol.Declarations) == 0 || !strings.HasSuffix(ast.GetSourceFileOfNode(symbol.Declarations[0]).FileName(), "/values.ts") {
								t.Error("lost cross-file class symbol")
							}
						}
						if name == "external" {
							if !utils.IsSymbolFromDefaultLibrary(ctx.Program(), checker.Type_symbol(typ)) {
								t.Error("Promise symbol is not from default library")
							}
							awaited := checker.Checker_getAwaitedType(ctx.TypeChecker, typ)
							if awaited == nil || awaited.Flags()&checker.TypeFlagsNumber == 0 {
								t.Error("cannot resolve awaited Promise<number>")
							}
						}
						if name == "ambient" || name == "thenable" {
							symbol := ctx.Refs.Resolve(subject)
							if symbol == nil || len(symbol.Declarations) == 0 || !strings.HasSuffix(ast.GetSourceFileOfNode(symbol.Declarations[0]).FileName(), "/ambient.d.ts") {
								t.Errorf("%s lost ambient symbol", name)
							}
						}
					}}
				},
			}
			runCapabilityRules(t, program, file, []rule.ConfiguredRule{probe})
			if len(seen) != 12 {
				t.Fatalf("visited %d probes, want 12", len(seen))
			}
			sourceOnly, err := lintprogram.NewFromBoundSources(raw, []*ast.SourceFile{file})
			if err != nil {
				t.Fatal(err)
			}
			configured := rule.ConfiguredRule{
				Name:             ValidExpectWithPromiseRule.Name,
				RequiresTypeInfo: ValidExpectWithPromiseRule.RequiresTypeInfo,
				Severity:         rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return ValidExpectWithPromiseRule.Run(ctx, []any{map[string]any{"checkThenables": true}})
				},
			}
			result, diagnostics := runCapabilityRules(t, program, file, []rule.ConfiguredRule{configured})
			want := 4
			if strict {
				want++
			}
			if len(diagnostics) != want {
				t.Fatalf("rule diagnostics = %d, want %d: %+v", len(diagnostics), want, diagnostics)
			}
			if _, ok := result.ExecutedRules[configured.Name]; !ok {
				t.Fatal("type-aware rule was not executed")
			}
			if diagnostics := raw.GetSemanticDiagnostics(context.Background(), file); len(diagnostics) != 0 {
				t.Fatalf("synthetic calls polluted compiler diagnostics: %v", diagnostics)
			}
			clear(seen)
			result, diagnostics = runCapabilityRules(t, sourceOnly, file, []rule.ConfiguredRule{probe, configured})
			if len(seen) != 0 || len(diagnostics) != 0 || len(result.ExecutedRules) != 0 {
				t.Fatal("source-only program ran type-aware probe")
			}
		})
	}
}

func TestValidExpectWithPromiseNoLib(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/capability.txtar").Materialize(t, "no-lib"))
	fs := bundled.WrapFS(osvfs.FS())
	raw, err := utils.CreateProgram(true, fs, dir, "tsconfig.json", utils.CreateCompilerHost(dir, fs))
	if err != nil {
		t.Fatal(err)
	}
	program := lintprogram.NewFromCompiler(raw)
	file := program.GetSourceFile(tspath.ResolvePath(dir, "entry.ts"))
	if file == nil || !program.Options().NoLib.IsTrue() {
		t.Fatal("noLib fixture was not loaded")
	}
	for _, enabled := range []bool{false, true} {
		_, diagnostics := runCapabilityRules(t, program, file, []rule.ConfiguredRule{{
			Name: ValidExpectWithPromiseRule.Name, RequiresTypeInfo: true, Severity: rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return ValidExpectWithPromiseRule.Run(ctx, []any{map[string]any{"checkThenables": enabled}})
			},
		}})
		want := "unneededRejectResolve"
		if enabled {
			want = "poorlyExpectedPromise"
		}
		if len(diagnostics) != 1 || diagnostics[0].Message.Id != want {
			t.Fatalf("checkThenables=%v: diagnostics = %+v, want %s", enabled, diagnostics, want)
		}
	}
}

func runCapabilityRules(t *testing.T, program *lintprogram.Program, file *ast.SourceFile, rules []rule.ConfiguredRule) (*linter.LintResult, []rule.RuleDiagnostic) {
	t.Helper()
	plan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs: []*lintprogram.Program{program}, TargetsByProgram: [][]string{{file.FileName()}}, SingleThreaded: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule { return rules },
	})
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []rule.RuleDiagnostic
	result, err := linter.RunLinter(linter.RunLinterOptions{SingleThreaded: true, LintPlan: plan,
		Consumer: rule.DiagnosticConsumer{Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
	})
	if err != nil {
		t.Fatal(err)
	}
	return result, diagnostics
}
