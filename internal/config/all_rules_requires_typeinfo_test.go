// cspell:ignore fset elts typeinfo

package config

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
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
}

func newRuleSourceParser() *ruleSourceParser {
	return &ruleSourceParser{
		fset:  token.NewFileSet(),
		files: make(map[string]*ast.File),
	}
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
