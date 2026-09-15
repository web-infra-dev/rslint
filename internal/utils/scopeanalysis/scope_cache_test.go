package scopeanalysis

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

func cacheContext() rule.RuleContext {
	return rule.RuleContext{SourceFile: parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/scope.tsx", Path: "/scope.tsx",
	}, `const x = 1; function f(a = x) { let x = a; return x; } f(x); missing;`, core.ScriptKindTSX)}.WithFileCache(rule.NewFileCache())
}

func TestGetKeysAndIsolation(t *testing.T) {
	ctx := cacheContext()
	options := []scope.Options{
		{},
		{CollectReferences: true},
		{CollectReferences: true, ReferenceNames: map[string]struct{}{}},
		{CollectReferences: true, ReferenceNames: map[string]struct{}{"x": {}}},
		{CollectReferences: true, ReferenceNames: map[string]struct{}{"x": {}, "missing": {}}},
	}
	for _, reverse := range []bool{false, true} {
		ctx := ctx.WithFileCache(rule.NewFileCache())
		managers := make([]*scope.Manager, len(options))
		for step := range options {
			i := step
			if reverse {
				i = len(options) - step - 1
			}
			managers[i] = Get(ctx, options[i])
			want := scope.Build(ctx.SourceFile, options[i])
			if !reflect.DeepEqual(managers[i], want) {
				t.Fatalf("options %d differ from uncached build", i)
			}
			for _, reference := range managers[i].References {
				managers[i].Acquire(reference.Identifier)
				want.Acquire(reference.Identifier)
			}
			if !reflect.DeepEqual(managers[i], want) {
				t.Fatalf("options %d differ after Acquire", i)
			}
		}
		for i, manager := range managers {
			if manager != Get(ctx, options[i]) {
				t.Fatalf("options %d not reused", i)
			}
			for j := range i {
				if manager == managers[j] {
					t.Fatalf("options %d and %d incorrectly shared", i, j)
				}
			}
			if manager == Get(ctx.WithFileCache(rule.NewFileCache()), options[i]) {
				t.Fatal("different file passes shared a manager")
			}
		}
	}
	uncached := rule.RuleContext{SourceFile: ctx.SourceFile}
	first := Get(uncached, scope.Options{})
	second := Get(uncached, scope.Options{})
	if first == second {
		t.Fatal("uncached context shared a manager")
	}
}

func TestGetNameSetContent(t *testing.T) {
	ctx := cacheContext()
	names := map[string]struct{}{"x": {}, "missing": {}}
	first := Get(ctx, scope.Options{CollectReferences: true, ReferenceNames: names})
	second := Get(ctx, scope.Options{CollectReferences: true, ReferenceNames: map[string]struct{}{"missing": {}, "x": {}}})
	if first != second {
		t.Fatal("equivalent name sets did not share")
	}
	delete(names, "missing")
	if first == Get(ctx, scope.Options{CollectReferences: true, ReferenceNames: names}) {
		t.Fatal("mutated name set reused stale manager")
	}
	if Get(ctx, scope.Options{}) != Get(ctx, scope.Options{ReferenceNames: names}) {
		t.Fatal("ignored names changed declaration-only cache")
	}
	for _, pair := range [][2][]string{{{"x\x00y"}, {"x", "y"}}, {{""}, {}}, {{"a", "bc"}, {"ab", "c"}}} {
		var values [2]*scope.Manager
		for i, set := range pair {
			names := make(map[string]struct{}, len(set))
			for _, name := range set {
				names[name] = struct{}{}
			}
			values[i] = Get(ctx, scope.Options{CollectReferences: true, ReferenceNames: names})
		}
		if values[0] == values[1] {
			t.Fatal("different name sets collided")
		}
	}
}

func TestDeclarationsReusesAnyExistingTree(t *testing.T) {
	ctx := cacheContext()
	filtered := Get(ctx, scope.Options{
		CollectReferences: true,
		ReferenceNames:    map[string]struct{}{"x": {}},
	})
	if got := Declarations(ctx); got != filtered {
		t.Fatal("declaration query did not reuse the existing scope tree")
	}

	ctx = ctx.WithFileCache(rule.NewFileCache())
	declarations := Declarations(ctx)
	if got := Get(ctx, scope.Options{}); got != declarations {
		t.Fatal("declaration query did not publish its scope tree")
	}
}

func TestReferencesReusesOnlyCompleteReferences(t *testing.T) {
	ctx := cacheContext()
	complete := Get(ctx, scope.Options{CollectReferences: true})
	if got := References(ctx, map[string]struct{}{"x": {}}); got != complete {
		t.Fatal("filtered query did not reuse complete references")
	}

	ctx = ctx.WithFileCache(rule.NewFileCache())
	first := References(ctx, map[string]struct{}{"x": {}})
	second := References(ctx, map[string]struct{}{"missing": {}})
	if first == second {
		t.Fatal("incompatible filtered queries shared a manager")
	}
}

func TestDeclarationsReusesCompleteReferences(t *testing.T) {
	ctx := cacheContext()
	complete := Get(ctx, scope.Options{CollectReferences: true})
	if got := Declarations(ctx); got != complete {
		t.Fatal("declaration query did not reuse complete references")
	}
	if len(complete.References) == 0 {
		t.Fatal("complete reference manager lost its references")
	}
}

func BenchmarkGet(b *testing.B) {
	ctx := rule.RuleContext{SourceFile: parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/scope.ts", Path: "/scope.ts",
	}, strings.Repeat(`{ let x = value; function f(a = x) { let y = a; return y; } f(x); }`, 100), core.ScriptKindTS)}
	for _, shared := range []bool{false, true} {
		name := "uncached"
		if shared {
			name = "shared"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				pass := ctx
				if shared {
					pass = ctx.WithFileCache(rule.NewFileCache())
				}
				for range 5 {
					Get(pass, scope.Options{CollectReferences: true})
				}
			}
		})
	}
}
