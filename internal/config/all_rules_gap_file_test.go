package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// gapFileFixtureSources is a bundle of TS / TSX constructs chosen to exercise
// as many rule listeners as possible — identifier references, spread
// arguments, JSX attributes (plain / spread / shorthand), class components
// with `this.state`, createElement calls, imports/exports, function and
// arrow declarations. Any rule whose listeners touch these constructs will
// have its TypeChecker-dependent code paths invoked.
var gapFileFixtureSources = map[string]string{
	"fixture.tsx": `
		import * as React from "react";
		export const DANGER = { __html: "<b>x</b>" };

		const props = { dangerouslySetInnerHTML: DANGER };
		const style = "not-an-object";
		const moreProps = { className: "x", ...props };

		export function Inline() {
			return <div {...props}>hi</div>;
		}

		export function StyleAsIdent() {
			return <div style={style} />;
		}

		export function StyleAsShorthand() {
			return React.createElement("div", { style });
		}

		export function SpreadCall() {
			return React.createElement("div", moreProps, "child");
		}

		export class Greeter extends React.Component<{}, { name: string }> {
			state = { name: "world" };
			bump() {
				const { name } = this.state;
				this.setState({ name: name + "!" });
			}
			render() {
				return <span>{this.state.name}</span>;
			}
		}

		export const identity = <T,>(x: T): T => x;
		export const nested = () => identity(props);
	`,
	"fixture.ts": `
		export const a = 1;
		export const b = a + 1;
		export function f(x: number): number { return x + 1; }
		export type Alias = { n: number };
		export const obj: Alias = { n: 2 };
		export const { n } = obj;
		export const arr = [a, b, n];
	`,
}

// TestGapFile_OptionalTypeCheckerRules_DoNotPanic is a regression sweep for
// the bug class behind https://github.com/web-infra-dev/rslint/issues/781.
//
// Rules that do NOT set RequiresTypeInfo: true are scheduled on "gap files"
// — files in the program but outside typeInfoFiles (see linter.go) — with a
// nil ctx.TypeChecker. A rule that calls a checker-dependent helper without
// a nil guard crashes the whole lint goroutine.
//
// This test runs EVERY currently-registered non-type-aware rule against a
// gap-file fixture and asserts no panic. It is intentionally a sweep, not a
// targeted test: any new rule that forgets to nil-guard TypeChecker use will
// be caught here without the rule author having to remember to add a test.
//
// To verify the sweep is real, a probe rule confirms that the gap-file path
// is actually being exercised (i.e. TypeChecker is in fact nil for each
// listener invocation), guarding against future linter changes that might
// silently skip gap files.
func TestGapFile_OptionalTypeCheckerRules_DoNotPanic(t *testing.T) {
	RegisterAllRules()

	program, paths := createGapFileProgram(t, gapFileFixtureSources)

	// Empty typeInfoFiles → every fixture file is a gap file, so every rule
	// on every file receives a nil TypeChecker.
	gap := map[string]struct{}{}

	configured := collectNonTypeAwareRules(t)
	if len(configured) == 0 {
		t.Fatal("expected at least one non-type-aware rule; registry looks empty")
	}

	linter.RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []linter.ConfiguredRule { return configured },
		false,
		func(d rule.RuleDiagnostic) {},
		gap,
		nil,
	)

	// Verify the harness really did pass nil TypeChecker. A listener needs
	// to fire at least once for this to be meaningful, so we attach an
	// Identifier listener (every fixture has many of those).
	var sawNilChecker, sawAnyListener bool
	linter.RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     "gap-probe",
				Severity: rule.SeverityWarning,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return rule.RuleListeners{
						ast.KindIdentifier: func(n *ast.Node) {
							sawAnyListener = true
							if ctx.TypeChecker == nil {
								sawNilChecker = true
							}
						},
					}
				},
			}}
		},
		false,
		func(d rule.RuleDiagnostic) {},
		gap,
		nil,
	)
	if !sawAnyListener {
		t.Fatal("probe listener never fired; test fixture is not being traversed")
	}
	if !sawNilChecker {
		t.Fatal("expected gap files to yield a nil TypeChecker on every listener call; the regression path is not being exercised")
	}
	_ = paths
}

// collectNonTypeAwareRules returns a ConfiguredRule for every registered rule
// that does not set RequiresTypeInfo: true. Each rule is run with nil
// options — the point is to exercise the listener / TypeChecker plumbing,
// not to test correctness of the report payloads.
func collectNonTypeAwareRules(t *testing.T) []linter.ConfiguredRule {
	t.Helper()
	all := GlobalRuleRegistry.GetAllRules()
	out := make([]linter.ConfiguredRule, 0, len(all))
	for name, impl := range all {
		if impl.RequiresTypeInfo {
			continue
		}
		ruleImpl := impl
		out = append(out, linter.ConfiguredRule{
			Name:             name,
			Severity:         rule.SeverityWarning,
			RequiresTypeInfo: false,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return ruleImpl.Run(ctx, nil)
			},
		})
	}
	return out
}

// createGapFileProgram mirrors internal/linter.createTestProgramWithFiles
// but is duplicated here to keep this package free of test-only exports
// from internal/linter.
func createGapFileProgram(t *testing.T, sourceFiles map[string]string) (*compiler.Program, map[string]string) {
	t.Helper()
	tmpDir := t.TempDir()

	paths := make(map[string]string, len(sourceFiles))
	rootFiles := make([]string, 0, len(sourceFiles))
	for name, content := range sourceFiles {
		p := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		paths[name] = tspath.NormalizePath(p)
		rootFiles = append(rootFiles, paths[name])
	}

	// Build a program from explicit root file names rather than a tsconfig
	// include glob — tsgo's tsconfig-include resolution does not pick up
	// .tsx files across all configurations, which would silently defeat the
	// point of this test (no JSX file → no JSX listener fired → no
	// regression coverage).
	compilerOptions := &core.CompilerOptions{
		Jsx:             core.JsxEmitPreserve,
		Target:          core.ScriptTargetESNext,
		Module:          core.ModuleKindCommonJS,
		ESModuleInterop: core.TSTrue,
		Strict:          core.TSFalse,
		AllowJs:         core.TSFalse,
		SkipLibCheck:    core.TSTrue,
		Lib:             []string{"lib.es2020.full.d.ts", "lib.dom.d.ts"},
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, compilerOptions, rootFiles, host)
	if err != nil {
		t.Fatalf("create program: %v", err)
	}
	return program, paths
}
