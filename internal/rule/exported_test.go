package rule

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

func TestIsGlobalScopeBinding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileName    string
		scriptKind  core.ScriptKind
		source      string
		sourceType  string
		withoutRefs bool
		want        bool
	}{
		{name: "script var", source: `var Target = 1;`, sourceType: "script", want: true},
		{name: "script function", source: `function Target() {}`, sourceType: "script", want: true},
		{name: "script lexical destructuring", source: `const { value: Target } = source;`, sourceType: "script", want: true},
		{name: "script block hoisted var", source: `{ var Target = 1; }`, sourceType: "script", want: true},
		{name: "script type alias", source: `type Target = string;`, sourceType: "script", want: true},
		{name: "script exported type alias", source: `export type Target = string;`, sourceType: "script", want: true},
		{name: "script named default export", source: `export default function Target() {}`, sourceType: "script", want: true},
		{name: "script import equals", source: `import Target = require("target");`, sourceType: "script", want: true},
		{name: "typescript commonjs global", source: `type Target = string;`, sourceType: "commonjs", want: true},

		{name: "default module without module syntax", source: `type Target = string;`},
		{name: "explicit module", source: `type Target = string;`, sourceType: "module"},
		{name: "javascript commonjs wrapper", fileName: "/scope.js", scriptKind: core.ScriptKindJS, source: `var Target = 1;`, sourceType: "commonjs"},
		{name: "fallback script", source: `var Target = 1;`, sourceType: "script", withoutRefs: true, want: true},
		{name: "fallback module", source: `var Target = 1;`, sourceType: "module", withoutRefs: true},
		{name: "fallback javascript commonjs", fileName: "/scope.js", scriptKind: core.ScriptKindJS, source: `var Target = 1;`, sourceType: "commonjs", withoutRefs: true},
		{name: "fallback typescript commonjs", source: `type Target = string;`, sourceType: "commonjs", withoutRefs: true, want: true},
		{name: "block lexical", source: `{ let Target = 1; }`, sourceType: "script"},
		{name: "block function", source: `{ function Target() {} }`, sourceType: "script"},
		{name: "function local", source: `function f() { let Target = 1; }`, sourceType: "script"},
		{name: "function parameter", source: `function f(Target: number) {}`, sourceType: "script"},
		{name: "named function expression", source: `const Alias = function Target() {};`, sourceType: "script"},
		{name: "named class expression", source: `const Alias = class Target {};`, sourceType: "script"},
		{name: "catch parameter", source: `try {} catch (Target) {}`, sourceType: "script"},
		{name: "namespace member", source: `namespace N { const Target = 1; }`, sourceType: "script"},
		{name: "static block local", source: `class C { static { let Target = 1; } }`, sourceType: "script"},

		{name: "type alias parameter", source: `type Alias<Target> = string;`, sourceType: "script"},
		{name: "interface parameter", source: `interface Alias<Target> {}`, sourceType: "script"},
		{name: "class parameter", source: `class Alias<Target> {}`, sourceType: "script"},
		{name: "function parameter type", source: `function f<Target>() {}`, sourceType: "script"},
		{name: "call signature parameter", source: `interface I { <Target>(): void; }`, sourceType: "script"},
		{name: "construct signature parameter", source: `interface I { new <Target>(): object; }`, sourceType: "script"},
		{name: "method signature parameter", source: `interface I { f<Target>(): void; }`, sourceType: "script"},
		{name: "signature value parameter", source: `type F = (Target: number) => void;`, sourceType: "script"},
		{name: "mapped type key", source: `type M<K> = { [Target in keyof K]: K[Target] };`, sourceType: "script"},
		{name: "infer type parameter", source: `type M<T> = T extends infer Target ? string : never;`, sourceType: "script"},
		{name: "enum member", source: `enum E { Target }`, sourceType: "script"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fileName := test.fileName
			if fileName == "" {
				fileName = "/scope.ts"
			}
			scriptKind := test.scriptKind
			if scriptKind == core.ScriptKindUnknown {
				scriptKind = core.ScriptKindTS
			}
			options := LanguageOptions{SourceType: test.sourceType}
			sourceFile, refs := newBoundRefStoreWithLanguageOptions(t, fileName, scriptKind, test.source, options)
			_, _, normalizedOptions := ResolveLanguageDefaults(fileName, options)
			ctx := RuleContext{
				SourceFile:      sourceFile,
				LanguageOptions: normalizedOptions,
				Refs:            refs,
				Exported:        NewExported(map[string]bool{"Target": true}, nil),
			}
			if test.withoutRefs {
				ctx.Refs = nil
			}
			symbol := declaredIdentifierSymbol(t, sourceFile, "Target")

			if got := ctx.IsGlobalScopeBinding(symbol, "Target"); got != test.want {
				t.Errorf("IsGlobalScopeBinding() = %v, want %v", got, test.want)
			}
			if got := ctx.IsExportedGlobalBinding(symbol, "Target"); got != test.want {
				t.Errorf("IsExportedGlobalBinding() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestIsExportedGlobalBindingRequiresListedName(t *testing.T) {
	t.Parallel()

	sourceFile, refs := newBoundRefStoreWithLanguageOptions(
		t,
		"/scope.ts",
		core.ScriptKindTS,
		`var Target = 1;`,
		LanguageOptions{SourceType: "script"},
	)
	ctx := RuleContext{SourceFile: sourceFile, Refs: refs}
	if ctx.IsExportedGlobalBinding(declaredIdentifierSymbol(t, sourceFile, "Target"), "Target") {
		t.Fatal("an unlisted global binding must not be treated as exported")
	}
}

func declaredIdentifierSymbol(t *testing.T, sourceFile *ast.SourceFile, name string) *ast.Symbol {
	t.Helper()
	occurrences := identifiers(sourceFile.AsNode(), name)
	if len(occurrences) == 0 {
		t.Fatalf("no %q identifier found", name)
	}
	declaration := ast.GetDeclarationFromName(occurrences[0])
	if declaration == nil {
		declaration = occurrences[0].Parent
	}
	if declaration == nil || declaration.Symbol() == nil {
		t.Fatalf("%q declaration has no raw binder symbol", name)
	}
	return declaration.Symbol()
}
