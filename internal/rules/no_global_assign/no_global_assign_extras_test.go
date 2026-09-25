package no_global_assign

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoGlobalAssignCachedShadowing(t *testing.T) {
	for _, test := range []struct {
		name      string
		code      string
		want      []string
		spellings []string
		options   []any
		globals   map[string]utils.GlobalAccess
	}{
		{
			name: "member names do not declare globals",
			code: `const props = { Object: 1, Array() {} };
/*report*/Object = 1; /*report*/Object++; /*report*/Array = 2;`,
			want: []string{"Object", "Object", "Array"},
		},
		{
			name: "sibling scopes keep independent answers",
			code: `function local(Object) { Object = 0; }
function f() { /*report*/Object = 1; /*report*/Object++; }
function g() { let Object; Object = 2; }
/*report*/Object = 3;`,
			want: []string{"Object", "Object", "Object"},
		},
		{
			name: "parameter initializers stay outside body declarations",
			code: `function f(x = (/*report*/Object = 1)) { var Object; Object = 2; }
function g(x = (/*report*/Object = 3)) { var Object; Object++; }
/*report*/Object = 4;`,
			want: []string{"Object", "Object", "Object"},
		},
		{
			name: "namespace loop and catch bindings stay inside their scopes",
			code: `namespace N { let Object; Object = 0; }
for (let Object of []) { Object = 0; }
try {} catch (Object) { Object = 0; }
/*report*/Object = 1;`,
			want: []string{"Object"},
		},
		{
			name: "hoisted file binding shadows every write",
			code: `function f() { Object = 1; } Object++; { var Object; }`,
		},
		{
			name: "type declarations and member names do not shadow values",
			code: `interface Object {} type Array = unknown;
function f<Object>() { /*report*/Object = 1; }
/*report*/Array = 2;`,
			want: []string{"Object", "Array"},
		},
		{
			name: "exceptions and global overrides",
			code: `/* global custom: readonly */
/*report*/custom = 1; Object = 1; String = 1; Array = 1; /*report*/custom++;`,
			want:    []string{"custom", "custom"},
			options: []any{map[string]any{"exceptions": []any{"Object"}}},
			globals: map[string]utils.GlobalAccess{
				"custom": utils.GlobalAccessWritable,
				"String": utils.GlobalAccessWritable,
				"Array":  utils.GlobalAccessOff,
			},
		},
		{
			name: "wrapped destructuring keeps write compatibility",
			code: `Object!! = 1; [/*report*/Object!! = 1] = items;
(Object satisfies any) = 0; [(/*report*/Object!! += 1)] = items;`,
			want: []string{"Object", "Object"},
		},
		{
			name: "single type wrappers remain writes",
			code: `(<any>/*report*/Object) = 1; (/*report*/Object!)++;`,
			want: []string{"Object", "Object"},
		},
		{
			name: "cached names still honor disable directives",
			code: `// eslint-disable-next-line no-global-assign
Object = 1;
/*report*/Object = 2;
/* eslint-disable no-global-assign */
Object++;
/* eslint-enable no-global-assign */
/*report*/Object = 3;
Object = 4; // eslint-disable-line no-global-assign`,
			want: []string{"Object", "Object"},
		},
		{
			name:      "escaped names preserve source ranges",
			code:      `/*report*/\u004f\u0062\u006a\u0065\u0063\u0074 = 1; /*report*/Object /* trivia */ ++;`,
			want:      []string{"Object", "Object"},
			spellings: []string{`\u004f\u0062\u006a\u0065\u0063\u0074`, "Object"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/no-global-assign.ts", Path: "/no-global-assign.ts",
			}, test.code, core.ScriptKindTS)
			binder.BindSourceFile(sourceFile)
			globalsInit, refsInit, language := rule.ResolveLanguageDefaults(sourceFile.FileName(), rule.LanguageOptions{})
			for _, withRefs := range []bool{false, true} {
				for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
					t.Run(fmt.Sprintf("refs=%t/demand=%d", withRefs, demand), func(t *testing.T) {
						comments := rule.NewCommentStore(sourceFile)
						inline, declarations := rule.ParseInlineGlobals(sourceFile, comments)
						ctx := rule.RuleContext{
							SourceFile:      sourceFile,
							LanguageOptions: language,
							Globals:         rule.NewGlobals(language, globalsInit, test.globals, inline, declarations),
							Comments:        comments,
							DisableManager:  rule.NewDisableManager(sourceFile, comments),
						}
						if withRefs {
							ctx.Refs = rule.NewRefStore(sourceFile, &core.CompilerOptions{}, nil, refsInit)
						}
						var diagnostics []rule.RuleDiagnostic
						ctx = ctx.WithDiagnosticConsumer(NoGlobalAssignRule.Name, rule.SeverityWarning, rule.DiagnosticConsumer{
							Demand: demand,
							Report: func(diagnostic rule.RuleDiagnostic) {
								diagnostics = append(diagnostics, diagnostic)
							},
						})
						listeners := NoGlobalAssignRule.Run(ctx, test.options)
						var visit func(*ast.Node) bool
						visit = func(node *ast.Node) bool {
							if listener := listeners[node.Kind]; listener != nil {
								listener(node)
							}
							node.ForEachChild(visit)
							if listener := listeners[rule.ListenerOnExit(node.Kind)]; listener != nil {
								listener(node)
							}
							return false
						}
						visit(sourceFile.AsNode())
						if len(diagnostics) != len(test.want) {
							t.Fatalf("got %d diagnostics, want %d: %v", len(diagnostics), len(test.want), diagnostics)
						}
						const marker = "/*report*/"
						position := 0
						for i, name := range test.want {
							offset := strings.Index(test.code[position:], marker)
							if offset < 0 {
								t.Fatal("missing expected diagnostic marker")
							}
							position += offset + len(marker)
							spelling := name
							if test.spellings != nil {
								spelling = test.spellings[i]
							}
							diagnostic := diagnostics[i]
							if diagnostic.Range.Pos() != position || diagnostic.Range.End() != position+len(spelling) {
								t.Errorf("range = %v, want [%d, %d)", diagnostic.Range, position, position+len(spelling))
							}
							if diagnostic.Message.Id != "globalShouldNotBeModified" ||
								diagnostic.Message.Description != "Read-only global '"+name+"' should not be modified." {
								t.Errorf("unexpected message: %v", diagnostic.Message)
							}
							if diagnostic.Severity != rule.SeverityWarning || diagnostic.FixesPtr != nil || diagnostic.Suggestions != nil {
								t.Errorf("unexpected severity or edits: %v", diagnostic)
							}
						}
					})
				}
			}
		})
	}
}
