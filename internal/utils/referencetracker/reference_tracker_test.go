package referencetracker

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func trackerContext(code string) rule.RuleContext {
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/tracker.tsx", Path: "/tracker.tsx",
	}, code, core.ScriptKindTSX)
	binder.BindSourceFile(source)
	globals, refs, language := rule.ResolveLanguageDefaults(source.FileName(), rule.LanguageOptions{SourceType: "module"})
	return rule.RuleContext{
		SourceFile:      source,
		LanguageOptions: language,
		Globals: rule.NewGlobals(language, globals, map[string]utils.GlobalAccess{
			"api": utils.GlobalAccessReadonly,
		}, nil, nil),
		Refs: rule.NewRefStore(source, &core.CompilerOptions{}, nil, refs),
	}.WithFileCache(rule.NewFileCache())
}

func collectEvents(ctx rule.RuleContext, tracker *Tracker) []string {
	var events []string
	collect := func(kind string) func(*ast.Node) {
		return func(node *ast.Node) {
			events = append(events, kind+": "+utils.TrimmedNodeText(ctx.SourceFile, node))
		}
	}
	tracker.TrackGlobals(map[string]*Trace{"api": {Properties: map[string]*Trace{
		"fn":   {Read: collect("read"), Call: collect("call")},
		"Ctor": {Construct: collect("new")},
	}}})
	return events
}

func TestGlobalReferences(t *testing.T) {
	for _, test := range []struct {
		name string
		code string
		want []string
	}{
		{"calls and constructors", `api.fn(); new api.Ctor();`, []string{"read: api.fn", "call: api.fn()", "new: new api.Ctor()"}},
		{"computed alias", `const {["f" + "n"]: run} = api; const invoke = run; invoke();`, []string{`read: ["f" + "n"]: run`, "call: invoke()"}},
		{"template binding key", "const {[`fn`]: run} = api; run();", []string{"read: [`fn`]: run", "call: run()"}},
		{"assignment default", `let run; ({fn: run = fallback} = api); run();`, []string{"read: fn: run = fallback", "call: run()"}},
		{"parameter default", `function f(target = api) { target.fn(); }`, []string{"read: target.fn", "call: target.fn()"}},
		{"converging paths retain duplicates", `const target = flag ? api : globalThis.api; target.fn();`, []string{"read: target.fn", "call: target.fn()", "read: target.fn", "call: target.fn()"}},
		{"cycle guard follows variable identity", `let target = api; let run = target.fn; target = run; target();`, []string{"read: target.fn"}},
		{"aliases are flow insensitive", `let run = api.fn; run = other; run();`, []string{"read: api.fn", "call: run()"}},
		{"transparent receiver", `const target = api as object; target!.fn();`, []string{"read: target!.fn", "call: target!.fn()"}},
		{"optional member", `api?.fn();`, []string{"read: api?.fn", "call: api?.fn()"}},
		{"root write disables earlier reads", `api.fn(); api = other;`, nil},
		{"shadowed root", `function f(api) { api.fn(); }`, nil},
		{"write to shadowed root", `function f(api) { api = other; } api.fn();`, []string{"read: api.fn", "call: api.fn()"}},
		{"no scope for keys", `const key = "fn"; api[key]();`, nil},
		{"rest is not an alias", `const {...rest} = api; rest.fn();`, nil},
		{"array patterns are not aliases", `const [run] = api; run();`, nil},
		{"comma keeps final value", `const target = (api, other); target.fn();`, nil},
		{"JSX tag is not a property read", `const element = <api.fn />;`, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := trackerContext(test.code)
			if got := collectEvents(ctx, New(ctx)); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("events = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSharedIndexKeepsRuleAndFileStateSeparate(t *testing.T) {
	ctx := trackerContext(`const target = api; target.fn();`)
	first, second := New(ctx), New(ctx)
	if first.names != second.names {
		t.Fatal("trackers for the same file did not share the name index")
	}
	if first.names == New(trackerContext(`api.fn();`)).names {
		t.Fatal("different files shared a name index")
	}
	want := []string{"read: target.fn", "call: target.fn()"}
	for _, tracker := range []*Tracker{first, second, first} {
		if got := collectEvents(ctx, tracker); !reflect.DeepEqual(got, want) {
			t.Fatalf("events = %q, want %q; traversal state leaked", got, want)
		}
	}
	ctx.Globals = rule.NewGlobals(ctx.LanguageOptions, rule.GlobalsInit{}, map[string]utils.GlobalAccess{
		"api": utils.GlobalAccessOff,
	}, nil, nil)
	if got := collectEvents(ctx, New(ctx)); len(got) != 0 {
		t.Fatalf("shared index retained another rule's globals: %q", got)
	}
}

func TestGlobalAssignmentAliases(t *testing.T) {
	ctx := trackerContext(`({api} = globalThis); api.fn();`)
	tracker := New(ctx)
	want := []string{"read: api.fn", "call: api.fn()"}
	if got := collectEvents(ctx, tracker); !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %q, want %q", got, want)
	}
}

func TestContextWithoutReferences(t *testing.T) {
	ctx := trackerContext(`api.fn(); function f(api) { api.fn(); }`)
	ctx.Refs = nil
	want := []string{"read: api.fn", "call: api.fn()"}
	if got := collectEvents(ctx, New(ctx)); !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %q, want %q", got, want)
	}
}

func TestReplacementReferences(t *testing.T) {
	for _, test := range []struct {
		code string
		want []string
	}{
		{
			"const target = api; target.fn();",
			[]string{"read: target.fn", "call: target.fn()"},
		},
		{
			"const {api: target} = globalThis; target.fn();",
			[]string{"read: target.fn", "call: target.fn()"},
		},
		{"let target = api; target = other; target.fn();", nil},
		{"const target = flag ? api : other; target.fn();", nil},
		{"function f(target = api) { target.fn(); }", nil},
		{"const {api: target = other} = globalThis; target.fn();", nil},
		{"(sideEffect(), api).fn();", nil},
		{"(target = api).fn();", nil},
	} {
		t.Run(test.code, func(t *testing.T) {
			ctx := trackerContext(test.code)
			if got := collectEvents(ctx, NewForReplacement(ctx)); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("events = %q, want %q", got, test.want)
			}
		})
	}
}
