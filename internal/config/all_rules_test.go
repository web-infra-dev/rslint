// cspell:ignore fset elts typeinfo

package config

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	tsast "github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestAllRules_NilTypeCheckerEarlyReturnImpliesRequiresTypeInfo is the static
// counterpart to TestGapFile_OptionalTypeCheckerRules_DoNotPanic: the panic
// sweep catches rules that crash when handed a nil TypeChecker, but it cannot
// catch rules that *silently* short-circuit to zero diagnostics — those slip
// through gap-file linting in CLI mode (nil checker) and then surface as false
// positives in LSP mode (where typescript-go's project session hands the rule
// an inferred-project TypeChecker that doesn't see `parserOptions.project` lib
// types).
//
// A rule with the shape
//
//	if ctx.TypeChecker == nil { return rule.RuleListeners{} }
//
// inside its Run body is, by definition, useless without type info. It must
// declare RequiresTypeInfo: true so the linter framework filters it out for
// gap files and inferred-project files instead of leaving it silently broken.
//
// The check resolves each registered rule's Run function pointer back to its
// source file via runtime.FuncForPC, then walks that file's AST to inspect
// only the matching FuncLit. This avoids ambiguity for rules registered under
// both bare and plugin-prefixed keys (e.g. `prefer-promise-reject-errors` and
// `@typescript-eslint/prefer-promise-reject-errors`) where both packages
// happen to use the same Go var name (`PreferPromiseRejectErrorsRule`) — a
// var-name-based scan would collapse those into one entry and silently miss
// regressions in one of the two packages.
func TestAllRules_NilTypeCheckerEarlyReturnImpliesRequiresTypeInfo(t *testing.T) {
	RegisterAllRules()
	registry := GlobalRuleRegistry.GetAllRules()

	// Iterate keys in sorted order so failure output is stable, which keeps
	// CI logs and `go test -run` rerun targets deterministic.
	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parser := newRuleSourceParser()
	var failures []string
	for _, key := range keys {
		impl := registry[key]
		if impl.RequiresTypeInfo {
			continue
		}
		body, file, err := parser.runBodyFor(impl.Run)
		if err != nil {
			// We require an unambiguous file:line for every registered Run
			// so the static check can never silently skip a rule. If the
			// reflect→AST mapping fails, that's a test bug, not a false
			// negative we should hide.
			t.Fatalf("rule %q: %v", key, err)
		}
		if hasNilTCEarlyReturn(body) {
			failures = append(failures, fmt.Sprintf(
				"%s: rule %q returns rule.RuleListeners{} when ctx.TypeChecker == nil but does not declare RequiresTypeInfo: true",
				file, key,
			))
		}
	}

	if len(failures) > 0 {
		t.Fatalf("rules return rule.RuleListeners{} on nil TypeChecker but do not declare RequiresTypeInfo: true (LSP would still run them with an inferred-project checker, producing false positives that CLI hides):\n  %s",
			strings.Join(failures, "\n  "))
	}
}

// ruleSourceParser maps a rule's runtime Run function back to its source-tree
// FuncLit body, caching parsed files so the cost stays at one parse per Go
// source file regardless of how many rules live in it.
type ruleSourceParser struct {
	fset  *token.FileSet
	files map[string]*ast.File
	// modulePath and moduleRoot translate the import-path-rooted file names
	// that -trimpath builds report back into on-disk paths; both are empty
	// when the module root could not be located.
	modulePath string
	moduleRoot string
}

func newRuleSourceParser() *ruleSourceParser {
	modulePath, moduleRoot := findModule()
	return &ruleSourceParser{
		fset:       token.NewFileSet(),
		files:      make(map[string]*ast.File),
		modulePath: modulePath,
		moduleRoot: moduleRoot,
	}
}

// findModule walks up from the test's working directory (the package source
// directory, per `go test`) to the enclosing go.mod, returning its module path
// and the directory holding it. Both are empty if no go.mod is found.
func findModule() (modulePath, moduleRoot string) {
	dir, err := os.Getwd()
	if err != nil {
		return "", ""
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			return modulePathFromGoMod(string(data)), dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ""
		}
		dir = parent
	}
}

func modulePathFromGoMod(contents string) string {
	for line := range strings.SplitSeq(contents, "\n") {
		if fields := strings.Fields(line); len(fields) >= 2 && fields[0] == "module" {
			return strings.Trim(fields[1], `"`)
		}
	}
	return ""
}

// resolveSourcePath maps the file name runtime reports for a function to a path
// that can be opened. Ordinary builds report absolute paths; -trimpath builds
// report the file's import path instead (`<module path>/internal/...`), which
// only resolves once the module path prefix is swapped for the module root.
func (p *ruleSourceParser) resolveSourcePath(file string) (string, error) {
	if filepath.IsAbs(file) {
		return file, nil
	}
	if p.modulePath == "" || p.moduleRoot == "" {
		return "", fmt.Errorf("cannot resolve trimmed path %s: no enclosing go.mod found", file)
	}
	rel, ok := strings.CutPrefix(path.Clean(filepath.ToSlash(file)), p.modulePath+"/")
	if !ok {
		return "", fmt.Errorf("cannot resolve trimmed path %s: not inside module %s", file, p.modulePath)
	}
	return filepath.Join(p.moduleRoot, filepath.FromSlash(rel)), nil
}

// runBodyFor returns the *ast.BlockStmt of the FuncLit registered as a rule's
// Run, plus the source file path. It uses runtime.FuncForPC on the function
// pointer to resolve (file, entry-line), then locates the innermost FuncLit
// in the parsed file whose body brackets contain that line.
func (p *ruleSourceParser) runBodyFor(run any) (*ast.BlockStmt, string, error) {
	rv := reflect.ValueOf(run)
	if !rv.IsValid() || rv.Kind() != reflect.Func {
		return nil, "", errors.New("Run is not a function value")
	}
	pc := rv.Pointer()
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return nil, "", errors.New("runtime.FuncForPC returned nil for Run pointer")
	}
	// fn.Entry() is the function's entry PC; FileLine on the entry typically
	// reports the line of the function declaration's opening brace. That's
	// inside the FuncLit body's range, which is what we want.
	file, line := fn.FileLine(fn.Entry())
	if file == "" {
		return nil, "", errors.New("FileLine returned empty file for Run pointer")
	}
	file, err := p.resolveSourcePath(file)
	if err != nil {
		return nil, "", err
	}
	parsed, err := p.parseFile(file)
	if err != nil {
		return nil, file, err
	}
	body := findFuncLitBodyAtLine(p.fset, parsed, line)
	if body == nil {
		return nil, file, fmt.Errorf("could not locate FuncLit at %s:%d", file, line)
	}
	return body, file, nil
}

func (p *ruleSourceParser) parseFile(path string) (*ast.File, error) {
	if cached, ok := p.files[path]; ok {
		return cached, nil
	}
	f, err := parser.ParseFile(p.fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	p.files[path] = f
	return f, nil
}

// findFuncLitBodyAtLine walks the file and returns the body of the smallest
// function-like node whose body brackets contain the given line. Both
// FuncLits (used for inline `Run: func(...) {...}`) and FuncDecls (used for
// `Run: run` references to a top-level function) are considered. Smallest-wins
// because outer Run bodies can themselves contain inner FuncLits (listener
// closures); we want the outer Run.
func findFuncLitBodyAtLine(fset *token.FileSet, file *ast.File, line int) *ast.BlockStmt {
	var found *ast.BlockStmt
	var foundSpan int
	consider := func(body *ast.BlockStmt) {
		if body == nil {
			return
		}
		startLine := fset.Position(body.Lbrace).Line
		endLine := fset.Position(body.Rbrace).Line
		if line < startLine || line > endLine {
			return
		}
		span := endLine - startLine
		if found == nil || span < foundSpan {
			found = body
			foundSpan = span
		}
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.FuncLit:
			consider(v.Body)
		case *ast.FuncDecl:
			consider(v.Body)
		}
		return true
	})
	return found
}

// hasNilTCEarlyReturn returns true if any statement in the function body is
// `if ctx.TypeChecker == nil { return rule.RuleListeners{} }` (or the same
// shape with `return nil`). The check is conservative: it only inspects
// top-level statements of Run, since that's the documented "useless without
// TC" pattern. Helpers that nil-guard internally are intentionally allowed.
func hasNilTCEarlyReturn(body *ast.BlockStmt) bool {
	for _, stmt := range body.List {
		ifStmt, ok := stmt.(*ast.IfStmt)
		if !ok {
			continue
		}
		if !isNilTCComparison(ifStmt.Cond) {
			continue
		}
		if returnsEmptyListeners(ifStmt.Body) {
			return true
		}
	}
	return false
}

func isNilTCComparison(expr ast.Expr) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.EQL {
		return false
	}
	// Match `ctx.TypeChecker == nil` (in either order).
	matchesField := func(e ast.Expr) bool {
		sel, ok := e.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "TypeChecker" {
			return false
		}
		x, ok := sel.X.(*ast.Ident)
		return ok && x.Name == "ctx"
	}
	matchesNil := func(e ast.Expr) bool {
		id, ok := e.(*ast.Ident)
		return ok && id.Name == "nil"
	}
	return (matchesField(bin.X) && matchesNil(bin.Y)) ||
		(matchesField(bin.Y) && matchesNil(bin.X))
}

func returnsEmptyListeners(body *ast.BlockStmt) bool {
	if len(body.List) == 0 {
		return false
	}
	ret, ok := body.List[0].(*ast.ReturnStmt)
	if !ok {
		return false
	}
	if len(ret.Results) == 0 {
		return false
	}
	switch r := ret.Results[0].(type) {
	case *ast.Ident:
		return r.Name == "nil"
	case *ast.CompositeLit:
		// rule.RuleListeners{}
		sel, ok := r.Type.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if sel.Sel.Name != "RuleListeners" {
			return false
		}
		x, ok := sel.X.(*ast.Ident)
		return ok && x.Name == "rule" && len(r.Elts) == 0
	}
	return false
}

// TestAllRules_DeclaredSchemasCompile compiles every registered rule's
// declared options schema. Schemas compile lazily at first use (so a bad
// schema would otherwise only surface when its rule is enabled by some
// user's config); this sweep front-loads that failure into CI, playing the
// role a MustCompile-at-init would — without making every rslint process pay
// startup compilation for hundreds of schemas.
//
// It also enforces that every registered rule declares a schema. Only
// ESLint-plugin placeholder rules run without one — the Node worker's own
// ESLint validates their options.
func TestAllRules_DeclaredSchemasCompile(t *testing.T) {
	RegisterAllRules()
	for name, ruleImpl := range GlobalRuleRegistry.GetAllRules() {
		if ruleImpl.IsEslintPluginRule {
			continue
		}
		if ruleImpl.Schema == nil {
			t.Errorf("rule %s declares no options schema; every registered rule must declare one", name)
			continue
		}
		if _, err := ruleImpl.Schema.Compile(); err != nil {
			t.Errorf("rule %s: options schema failed to compile: %v", name, err)
		}
	}
}

// TestAllRules_SilentOnNilTypeCheckerImpliesRequiresTypeInfo is the runtime
// counterpart to TestAllRules_NilTypeCheckerEarlyReturnImpliesRequiresTypeInfo.
//
// The static check matches `if ctx.TypeChecker == nil { return RuleListeners{} }`
// at the top of Run. Some rules instead push the nil-TC gate into a helper
// (e.g. no-obj-calls's checkCallee, no-const-assign's checkIdentifierWrite,
// no-use-before-define's checkIdentifier)
// — every listener funnels through that helper, so the rule emits zero
// diagnostics without TC even though Run itself is unguarded.
//
// This test runs each rule on a fixture *known to trigger it* under two
// configurations:
//
//  1. The rule's normal Run with a real, non-nil TypeChecker — must produce
//     ≥1 diagnostic (proves the fixture is correct and the rule actually
//     fires).
//  2. The same fixture but with `ctx.TypeChecker` forcibly set to nil before
//     calling Run — if the rule emits 0 diagnostics here while emitting ≥1
//     above, the rule is silently broken in CLI gap-file mode and would also
//     misbehave in LSP inferred-project mode, so it MUST declare
//     RequiresTypeInfo: true.
//
// Adding a new rule to this table is the regression guard: if you write a
// rule whose logic is meaningless without a TypeChecker, this test forces you
// to declare the flag.
func TestAllRules_SilentOnNilTypeCheckerImpliesRequiresTypeInfo(t *testing.T) {
	RegisterAllRules()
	registry := GlobalRuleRegistry.GetAllRules()

	cases := []struct {
		ruleKey  string
		fileName string
		source   string
	}{
		{
			ruleKey:  "no-undef",
			fileName: "no-undef.ts",
			source:   `someUndefinedThing();`,
		},
		{
			ruleKey:  "no-obj-calls",
			fileName: "no-obj-calls.ts",
			source:   `Math();`,
		},
		{
			ruleKey:  "no-const-assign",
			fileName: "no-const-assign.ts",
			source:   `const x = 1; x = 2;`,
		},
		{
			ruleKey:  "no-ex-assign",
			fileName: "no-ex-assign.ts",
			source:   `try { throw 1; } catch (e) { e = 2; }`,
		},
		{
			ruleKey:  "prefer-const",
			fileName: "prefer-const.ts",
			source:   `let neverReassigned = 1; console.log(neverReassigned);`,
		},
		{
			ruleKey:  "no-unmodified-loop-condition",
			fileName: "no-unmodified-loop-condition.ts",
			source:   `let n = 1; while (n < 10) { console.log(n); }`,
		},
		{
			ruleKey:  "no-loop-func",
			fileName: "no-loop-func.ts",
			source:   `for (var i = 0; i < 3; i++) { setTimeout(function() { console.log(i); }); }`,
		},
		{
			ruleKey:  "@typescript-eslint/no-use-before-define",
			fileName: "no-use-before-define.ts",
			source:   `useBefore(); function useBefore() {}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.ruleKey, func(t *testing.T) {
			impl, ok := registry[tc.ruleKey]
			if !ok {
				t.Fatalf("rule %q is not registered", tc.ruleKey)
			}

			withTC := countDiagnosticsForRule(t, tc.fileName, tc.source, impl, true)
			if withTC == 0 {
				t.Fatalf("fixture for %q produced 0 diagnostics with a real TypeChecker; the test fixture is wrong (it must trigger the rule so the without-TC comparison is meaningful)", tc.ruleKey)
			}

			withoutTC := countDiagnosticsForRule(t, tc.fileName, tc.source, impl, false)
			if withoutTC == 0 && !impl.RequiresTypeInfo {
				t.Fatalf("rule %q emits %d diagnostics with TypeChecker but 0 without; it is silently useless on gap files / LSP inferred-project files and MUST declare RequiresTypeInfo: true", tc.ruleKey, withTC)
			}
		})
	}
}

// countDiagnosticsForRule runs a single rule on a single-file program and
// returns how many diagnostics it produced. When withTypeChecker is false, the
// rule receives a nil TypeChecker — simulating the gap-file / inferred-project
// path that the existing FilterNonTypeAwareRules infrastructure is meant to
// guard against.
func countDiagnosticsForRule(t *testing.T, fileName, source string, impl rule.Rule, withTypeChecker bool) int {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := tspath.NormalizePath(filepath.Join(tmpDir, fileName))
	if err := os.WriteFile(filePath, []byte(source), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	compilerOptions := &core.CompilerOptions{
		Target:          core.ScriptTargetESNext,
		Module:          core.ModuleKindCommonJS,
		ESModuleInterop: core.TSTrue, //nolint:staticcheck
		SkipLibCheck:    core.TSTrue,
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, compilerOptions, []string{filePath}, host)
	if err != nil {
		t.Fatalf("create program: %v", err)
	}

	configured := linter.ConfiguredRule{
		Name:     impl.Name,
		Severity: rule.SeverityWarning,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return impl.Run(ctx, nil)
		},
	}

	sourceProgram := lintprogram.NewFromCompiler(program)
	if withTypeChecker {
		// The compiler-capable Program supplies the checker.
	} else {
		// A source-only Program matches CLI gap-file behavior without a parallel
		// production policy map.
		sourceProgram, err = lintprogram.NewFromBoundSources(program, program.SourceFiles())
		if err != nil {
			t.Fatalf("create source-only Program: %v", err)
		}
	}

	count := 0
	linter.RunLinterInProgram(sourceProgram, nil, nil, utils.ExcludePaths,
		func(sf *tsast.SourceFile) []linter.ConfiguredRule {
			if sf.FileName() != filePath {
				return nil
			}
			return []linter.ConfiguredRule{configured}
		},
		false,
		func(d rule.RuleDiagnostic) { count++ },
		nil,
	)
	return count
}

// Compile-time anchor: keep the linter import live in case future trims of
// the imports remove the only call site by accident.
var _ = compiler.Program{}

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
// Rules that do NOT set RequiresTypeInfo: true are scheduled on source-only
// gap Programs with a nil ctx.TypeChecker. A rule that calls a
// checker-dependent helper without a nil guard crashes the lint goroutine.
//
// This test runs EVERY currently-registered non-type-aware rule against a
// gap-file fixture and asserts no panic. It is intentionally a sweep, not a
// targeted test: any new rule that forgets to nil-guard TypeChecker use will
// be caught here without the rule author having to remember to add a test.
//
// A probe rule is attached alongside the sweep so every listener invocation
// is observed under the exact same run — it verifies that the harness really
// did hand the rules a nil TypeChecker, guarding against future linter
// changes that might silently skip gap files.
func TestGapFile_OptionalTypeCheckerRules_DoNotPanic(t *testing.T) {
	RegisterAllRules()

	program := createGapFileProgram(t, gapFileFixtureSources)

	sourceProgram, err := lintprogram.NewFromBoundSources(program, program.SourceFiles())
	if err != nil {
		t.Fatalf("create source-only Program: %v", err)
	}

	sweep := collectNonTypeAwareRules(t)
	if len(sweep) == 0 {
		t.Fatal("expected at least one non-type-aware rule; registry looks empty")
	}

	var sawNilChecker, sawAnyListener bool
	probe := linter.ConfiguredRule{
		Name:     "gap-probe",
		Severity: rule.SeverityWarning,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return rule.RuleListeners{
				tsast.KindIdentifier: func(n *tsast.Node) {
					sawAnyListener = true
					if ctx.TypeChecker == nil {
						sawNilChecker = true
					}
				},
			}
		},
	}
	configured := append(sweep, probe)

	linter.RunLinterInProgram(sourceProgram, nil, nil, utils.ExcludePaths,
		func(sf *tsast.SourceFile) []linter.ConfiguredRule { return configured },
		false,
		func(d rule.RuleDiagnostic) {},
		nil,
	)

	if !sawAnyListener {
		t.Fatal("probe listener never fired; test fixture is not being traversed")
	}
	if !sawNilChecker {
		t.Fatal("expected gap files to yield a nil TypeChecker on every listener call; the regression path is not being exercised")
	}
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
			Name:     name,
			Severity: rule.SeverityWarning,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return ruleImpl.Run(ctx, nil)
			},
		})
	}
	return out
}

// createGapFileProgram builds a tsgo program from an in-memory source map.
// Root file names are passed explicitly because, in local experiments, a
// tsconfig-driven include glob did not reliably pick up .tsx files across
// the setups this test needs — a missed .tsx file would silently neuter the
// sweep (no JSX listener fired → no regression coverage).
func createGapFileProgram(t *testing.T, sourceFiles map[string]string) *compiler.Program {
	t.Helper()
	tmpDir := t.TempDir()

	rootFiles := make([]string, 0, len(sourceFiles))
	for name, content := range sourceFiles {
		p := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		rootFiles = append(rootFiles, tspath.NormalizePath(p))
	}

	compilerOptions := &core.CompilerOptions{
		Jsx:             core.JsxEmitPreserve,
		Target:          core.ScriptTargetESNext,
		Module:          core.ModuleKindCommonJS,
		ESModuleInterop: core.TSTrue, //nolint:staticcheck
		SkipLibCheck:    core.TSTrue,
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, compilerOptions, rootFiles, host)
	if err != nil {
		t.Fatalf("create program: %v", err)
	}
	return program
}
