package rule

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type fileCacheTestKey struct{}

func TestCachedByFileReusesWithinOneFile(t *testing.T) {
	ctx := RuleContext{}.WithFileCache(NewFileCache())
	created := 0
	build := func() *int {
		created++
		return new(int)
	}

	first := CachedByFile(ctx, fileCacheTestKey{}, build)
	second := CachedByFile(ctx, fileCacheTestKey{}, build)
	if first != second {
		t.Fatal("same file cache did not reuse its value")
	}
	if created != 1 {
		t.Fatalf("builder called %d times, want 1", created)
	}
}

func TestCachedByFileSeparatesFiles(t *testing.T) {
	first := CachedByFile(
		RuleContext{}.WithFileCache(NewFileCache()),
		fileCacheTestKey{},
		func() *int { return new(int) },
	)
	second := CachedByFile(
		RuleContext{}.WithFileCache(NewFileCache()),
		fileCacheTestKey{},
		func() *int { return new(int) },
	)
	if first == second {
		t.Fatal("different file caches reused a value")
	}
}

func TestCachedByFileWithoutCacheBuildsEveryTime(t *testing.T) {
	created := 0
	build := func() *int {
		created++
		return new(int)
	}
	ctx := RuleContext{}
	first := CachedByFile(ctx, fileCacheTestKey{}, build)
	second := CachedByFile(ctx, fileCacheTestKey{}, build)
	if first == second {
		t.Fatal("context without a file cache unexpectedly reused a value")
	}
	if created != 2 {
		t.Fatalf("builder called %d times, want 2", created)
	}
}

func TestRuleContextProcessCurrentDirectoryUsesFileSharedState(t *testing.T) {
	ctx := RuleContext{}.WithFileCache(
		NewFileCacheWithProcessCurrentDirectory("/repo"),
	)
	if got := ctx.ProcessCurrentDirectory(); got != "/repo" {
		t.Fatalf("ProcessCurrentDirectory() = %q, want /repo", got)
	}
	if got := (&RuleContext{}).ProcessCurrentDirectory(); got != "" {
		t.Fatalf("empty ProcessCurrentDirectory() = %q, want empty", got)
	}
}

func TestRuleContextVariableUsageMarks(t *testing.T) {
	file, refs := newBoundRefStore(t, "/marks.jsx", core.ScriptKindJSX,
		`const Button = 0; function render(Button) { return <Button />; } consume(Button);`)
	nodes := identifiers(file.AsNode(), "Button")
	if len(nodes) != 4 {
		t.Fatalf("got %d Button identifiers, want 4", len(nodes))
	}
	outer, inner := nodes[0].Parent.Symbol(), nodes[1].Parent.Symbol()
	ctx := (RuleContext{SourceFile: file, Refs: refs}).WithFileCache(NewFileCache())
	if !ctx.MarkVariableAsUsed("Button", nodes[2]) {
		t.Fatal("could not mark local component")
	}
	otherRule := ctx
	if !otherRule.IsVariableMarkedAsUsed(inner) || otherRule.IsVariableMarkedAsUsed(outer) {
		t.Fatal("mark did not select the shadowing parameter")
	}
	if ctx.MarkVariableAsUsed("missing", nodes[2]) || ctx.IsGlobalMarkedAsUsed("Button") {
		t.Fatal("mark leaked to an unresolved name or global")
	}
	nextPass := ctx.WithFileCache(NewFileCache())
	if nextPass.IsVariableMarkedAsUsed(inner) {
		t.Fatal("usage mark leaked across lint passes sharing a bound source")
	}
	if !nextPass.MarkVariableAsUsed("Button", nodes[3]) || !nextPass.IsVariableMarkedAsUsed(outer) {
		t.Fatal("could not mark the outer binding from its scope")
	}
	if ctx.IsVariableMarkedAsUsed(outer) {
		t.Fatal("later lint pass changed earlier marks")
	}
}

func TestRuleContextGlobalUsageMarks(t *testing.T) {
	file, refs := newBoundRefStore(t, "/globals.jsx", core.ScriptKindJSX,
		`export function render(React) { return <div />; } consume(<div />);`)
	ctx := (RuleContext{
		SourceFile: file,
		Refs:       refs,
		Globals: NewGlobals(LanguageOptions{}, GlobalsInit{},
			map[string]utils.GlobalAccess{"React": utils.GlobalAccessReadonly}, nil, nil),
	}).WithFileCache(NewFileCache())
	parameter := identifiers(file.AsNode(), "React")[0]
	if !ctx.MarkVariableAsUsed("React", parameter.Parent.Parent.Body()) || ctx.IsGlobalMarkedAsUsed("React") {
		t.Fatal("local mark changed a shadowed global")
	}
	if !ctx.MarkVariableAsUsed("React", file.Statements.Nodes[1]) || !ctx.IsGlobalMarkedAsUsed("React") {
		t.Fatal("global use was not marked")
	}
	if (RuleContext{}).MarkVariableAsUsed("React", parameter) || ctx.MarkVariableAsUsed("React", nil) ||
		ctx.MarkVariableAsUsed("", parameter) || (RuleContext{}).IsVariableMarkedAsUsed(nil) ||
		(RuleContext{}).IsGlobalMarkedAsUsed("React") {
		t.Fatal("empty context or missing location/name should not have marks")
	}
}

func TestRuleContextMarksUseESLintFunctionScope(t *testing.T) {
	file, refs := newBoundRefStore(t, "/parameters.jsx", core.ScriptKindJSX,
		`const React = {}; export function render(view = React) { var React = {}; return view; }`)
	nodes := identifiers(file.AsNode(), "React")
	if len(nodes) != 3 {
		t.Fatalf("got %d React identifiers, want 3", len(nodes))
	}
	outer, inner := nodes[0].Parent.Symbol(), nodes[2].Parent.Symbol()
	ctx := (RuleContext{SourceFile: file, Refs: refs}).WithFileCache(NewFileCache())
	if refs.ResolveInFile(nodes[1]) != outer {
		t.Fatal("ordinary reference must not see the body declaration")
	}
	if !ctx.MarkVariableAsUsed("React", nodes[1]) || !ctx.IsVariableMarkedAsUsed(inner) || ctx.IsVariableMarkedAsUsed(outer) {
		t.Fatal("explicit mark must select the function scope's body declaration")
	}
	if refs.ResolveInFile(nodes[1]) != outer {
		t.Fatal("explicit mark changed ordinary reference resolution")
	}
}
