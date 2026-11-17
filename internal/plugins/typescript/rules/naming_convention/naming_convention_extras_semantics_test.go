package naming_convention

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

// Expectations were checked with @typescript-eslint/eslint-plugin 8.67.0.
// Auto-accessor type filters intentionally follow the documented semantics;
// upstream accepts the option but forgets to apply it to this selector.
func TestNamingConventionSemanticRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NamingConventionRule,
		[]rule_tester.ValidTestCase{
			// leading allowDouble __foo
			{Code: "const __foo=1;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "leadingUnderscore": "allowDouble"}}},
			// trailing allowDouble foo__
			{Code: "const foo__=1;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "trailingUnderscore": "allowDouble"}}},
			// type null/string
			{Code: "declare const bad_name: null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type null/number
			{Code: "declare const bad_name: null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type null/boolean
			{Code: "declare const bad_name: null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type null/function
			{Code: "declare const bad_name: null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type null/array
			{Code: "declare const bad_name: null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type undefined/string
			{Code: "declare const bad_name: undefined;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type undefined/number
			{Code: "declare const bad_name: undefined;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type undefined/boolean
			{Code: "declare const bad_name: undefined;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type undefined/function
			{Code: "declare const bad_name: undefined;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type undefined/array
			{Code: "declare const bad_name: undefined;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type never/string
			{Code: "declare const bad_name: never;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type never/number
			{Code: "declare const bad_name: never;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type never/boolean
			{Code: "declare const bad_name: never;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type never/function
			{Code: "declare const bad_name: never;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type never/array
			{Code: "declare const bad_name: never;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type any/string
			{Code: "declare const bad_name: any;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type any/number
			{Code: "declare const bad_name: any;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type any/boolean
			{Code: "declare const bad_name: any;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type any/function
			{Code: "declare const bad_name: any;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type any/array
			{Code: "declare const bad_name: any;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type unknown/string
			{Code: "declare const bad_name: unknown;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type unknown/number
			{Code: "declare const bad_name: unknown;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type unknown/boolean
			{Code: "declare const bad_name: unknown;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type unknown/function
			{Code: "declare const bad_name: unknown;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type unknown/array
			{Code: "declare const bad_name: unknown;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type string | number/string
			{Code: "declare const bad_name: string | number;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type string | number/number
			{Code: "declare const bad_name: string | number;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type string | number/boolean
			{Code: "declare const bad_name: string | number;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type string | number/function
			{Code: "declare const bad_name: string | number;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type string | number/array
			{Code: "declare const bad_name: string | number;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type string | null/number
			{Code: "declare const bad_name: string | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type string | null/boolean
			{Code: "declare const bad_name: string | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type string | null/function
			{Code: "declare const bad_name: string | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type string | null/array
			{Code: "declare const bad_name: string | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type boolean/string
			{Code: "declare const bad_name: boolean;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type boolean/number
			{Code: "declare const bad_name: boolean;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type boolean/function
			{Code: "declare const bad_name: boolean;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type boolean/array
			{Code: "declare const bad_name: boolean;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type true | false/string
			{Code: "declare const bad_name: true | false;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type true | false/number
			{Code: "declare const bad_name: true | false;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type true | false/function
			{Code: "declare const bad_name: true | false;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type true | false/array
			{Code: "declare const bad_name: true | false;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type "a" | "b"/number
			{Code: "declare const bad_name: \"a\" | \"b\";", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type "a" | "b"/boolean
			{Code: "declare const bad_name: \"a\" | \"b\";", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type "a" | "b"/function
			{Code: "declare const bad_name: \"a\" | \"b\";", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type "a" | "b"/array
			{Code: "declare const bad_name: \"a\" | \"b\";", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type 1 | 2/string
			{Code: "declare const bad_name: 1 | 2;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type 1 | 2/boolean
			{Code: "declare const bad_name: 1 | 2;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type 1 | 2/function
			{Code: "declare const bad_name: 1 | 2;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type 1 | 2/array
			{Code: "declare const bad_name: 1 | 2;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type string & {brand:true}/string
			{Code: "declare const bad_name: string & {brand:true};", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type string & {brand:true}/number
			{Code: "declare const bad_name: string & {brand:true};", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type string & {brand:true}/boolean
			{Code: "declare const bad_name: string & {brand:true};", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type string & {brand:true}/function
			{Code: "declare const bad_name: string & {brand:true};", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type string & {brand:true}/array
			{Code: "declare const bad_name: string & {brand:true};", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type () => void/string
			{Code: "declare const bad_name: () => void;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type () => void/number
			{Code: "declare const bad_name: () => void;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type () => void/boolean
			{Code: "declare const bad_name: () => void;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type () => void/array
			{Code: "declare const bad_name: () => void;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type (() => void) | null/string
			{Code: "declare const bad_name: (() => void) | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type (() => void) | null/number
			{Code: "declare const bad_name: (() => void) | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type (() => void) | null/boolean
			{Code: "declare const bad_name: (() => void) | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type (() => void) | null/array
			{Code: "declare const bad_name: (() => void) | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}},
			// type string[] | null/string
			{Code: "declare const bad_name: string[] | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type string[] | null/number
			{Code: "declare const bad_name: string[] | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type string[] | null/boolean
			{Code: "declare const bad_name: string[] | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type string[] | null/function
			{Code: "declare const bad_name: string[] | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type readonly string[]/string
			{Code: "declare const bad_name: readonly string[];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type readonly string[]/number
			{Code: "declare const bad_name: readonly string[];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type readonly string[]/boolean
			{Code: "declare const bad_name: readonly string[];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type readonly string[]/function
			{Code: "declare const bad_name: readonly string[];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type [string,number]/string
			{Code: "declare const bad_name: [string,number];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type [string,number]/number
			{Code: "declare const bad_name: [string,number];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type [string,number]/boolean
			{Code: "declare const bad_name: [string,number];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type [string,number]/function
			{Code: "declare const bad_name: [string,number];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// type number[] | string[]/string
			{Code: "declare const bad_name: number[] | string[];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// type number[] | string[]/number
			{Code: "declare const bad_name: number[] | string[];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}},
			// type number[] | string[]/boolean
			{Code: "declare const bad_name: number[] | string[];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}},
			// type number[] | string[]/function
			{Code: "declare const bad_name: number[] | string[];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}},
			// empty types
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{}}}},
			// autoAccessor type gate
			{Code: "class C { accessor bad_name=1; }", Options: []any{map[string]any{"selector": "autoAccessor", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// export export const good=function bad_name(){};
			{Code: "export const good=function bad_name(){};", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"exported"}}}},
			// export export const good=class bad_name {};
			{Code: "export const good=class bad_name {};", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"exported"}}}},
			// global namespace N {const bad_name=1;function bad_func(){}}
			{Code: "namespace N {const bad_name=1;function bad_func(){}}", Options: []any{map[string]any{"selector": []any{"variable", "function"}, "format": []any{"camelCase"}, "modifiers": []any{"global"}}}},
			// global { var bad_name=1; let bad_other=2; }
			{Code: "{ var bad_name=1; let bad_other=2; }", Options: []any{map[string]any{"selector": []any{"variable", "function"}, "format": []any{"camelCase"}, "modifiers": []any{"global"}}}},
			// global for(let bad_name=0;bad_name<1;bad_name++) {}
			{Code: "for(let bad_name=0;bad_name<1;bad_name++) {}", Options: []any{map[string]any{"selector": []any{"variable", "function"}, "format": []any{"camelCase"}, "modifiers": []any{"global"}}}},
			// global const good=function bad_name(){};
			{Code: "const good=function bad_name(){};", Options: []any{map[string]any{"selector": []any{"variable", "function"}, "format": []any{"camelCase"}, "modifiers": []any{"global"}}}},
			// unused let bad_name=1;console.log(bad_name);
			{Code: "let bad_name=1;console.log(bad_name);", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused declare function good(bad_arg:string):void;
			{Code: "declare function good(bad_arg:string):void;", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused const good=function bad_func(){};
			{Code: "const good=function bad_func(){};", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused const good=class bad_class{};
			{Code: "const good=class bad_class{};", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused export const bad_name=1;
			{Code: "export const bad_name=1;", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused enum good {bad_member}
			{Code: "enum good {bad_member}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused function good(bad_arg:string){return bad_arg;}
			{Code: "function good(bad_arg:string){return bad_arg;}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// accepted autoAccessor types 1
			{Code: "class C { accessor bad_name=1; }", Options: []any{map[string]any{"selector": "autoAccessor", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// accepted accessor types 1
			{Code: "class C { accessor bad_name=1; }", Options: []any{map[string]any{"selector": "accessor", "format": []any{"camelCase"}, "types": []any{"string"}}}},
			// modifier public: enum Good {bad_name}
			{Code: "enum Good {bad_name}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"public"}}}},
			// modifier const: enum Good {bad_name}
			{Code: "enum Good {bad_name}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"const"}}}},
			// modifier readonly: enum Good {bad_name}
			{Code: "enum Good {bad_name}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"readonly"}}}},
			// modifier public: const enum bad_name {good}
			{Code: "const enum bad_name {good}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"public"}}}},
			// modifier const: const enum bad_name {good}
			{Code: "const enum bad_name {good}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"const"}}}},
			// modifier readonly: const enum bad_name {good}
			{Code: "const enum bad_name {good}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"readonly"}}}},
			// modifier const: interface Good {readonly bad_name:()=>void}
			{Code: "interface Good {readonly bad_name:()=>void}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"const"}}}},
			// modifier readonly: interface Good {readonly bad_name:()=>void}
			{Code: "interface Good {readonly bad_name:()=>void}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"readonly"}}}},
			// modifier const: type Good={readonly bad_name:()=>void}
			{Code: "type Good={readonly bad_name:()=>void}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"const"}}}},
			// modifier readonly: type Good={readonly bad_name:()=>void}
			{Code: "type Good={readonly bad_name:()=>void}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"readonly"}}}},
			// modifier const: class Good {readonly bad_name=()=>{}}
			{Code: "class Good {readonly bad_name=()=>{}}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"const"}}}},
			// unused boundary let bad_name=1;bad_name||=1;
			{Code: "let bad_name=1;bad_name||=1;", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused boundary let bad_name=1;foo(bad_name++);
			{Code: "let bad_name=1;foo(bad_name++);", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused boundary let bad_name=1;for(;;){bad_name=bad_name+1;}
			{Code: "let bad_name=1;for(;;){bad_name=bad_name+1;}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused boundary let bad_name=1;function f(){bad_name=bad_name+1;}
			{Code: "let bad_name=1;function f(){bad_name=bad_name+1;}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused boundary function good(){for(const bad_key in obj){return true;}}
			{Code: "function good(){for(const bad_key in obj){return true;}}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused boundary function good(this:unknown){}
			{Code: "function good(this:unknown){}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// unused boundary function bad_func(){};export default bad_func;
			{Code: "function bad_func(){};export default bad_func;", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// used parameter with predicate
			{Code: "function good(bad_arg:unknown): bad_arg is string { return typeof bad_arg === \"string\"; }", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}},
			// regexp character class "[a/b]"
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": "[a/b]", "match": true}}}},
		}, []rule_tester.InvalidTestCase{
			// leading allowDouble _foo
			{Code: "const _foo=1;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "leadingUnderscore": "allowDouble"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `_foo` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 11}}},
			// leading allowDouble foo_
			{Code: "const foo_=1;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "leadingUnderscore": "allowDouble"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `foo_` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 11}}},
			// leading allowDouble foo__
			{Code: "const foo__=1;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "leadingUnderscore": "allowDouble"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `foo__` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 12}}},
			// trailing allowDouble _foo
			{Code: "const _foo=1;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "trailingUnderscore": "allowDouble"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `_foo` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 11}}},
			// trailing allowDouble __foo
			{Code: "const __foo=1;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "trailingUnderscore": "allowDouble"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `__foo` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 12}}},
			// trailing allowDouble foo_
			{Code: "const foo_=1;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "trailingUnderscore": "allowDouble"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `foo_` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 11}}},
			// meta accessor priority
			{Code: "class C { get bad_name() {} accessor bad_field=1; }", Options: []any{map[string]any{"selector": "accessor", "format": []any{"snake_case"}}, map[string]any{"selector": "memberLike", "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Classic Accessor name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 23}, {MessageId: "doesNotMatchFormat", Message: "Auto Accessor name `bad_field` must match one of the following formats: camelCase", Line: 1, Column: 38, EndLine: 1, EndColumn: 47}}},
			// parenthesized methods
			{Code: "class C { bad_name = (()=>{}); } const obj={bad_name:(function(){})}; type T={bad_name:(()=>void)};", Options: []any{map[string]any{"selector": "method", "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class Method name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 11, EndLine: 1, EndColumn: 19}, {MessageId: "doesNotMatchFormat", Message: "Object Literal Method name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 45, EndLine: 1, EndColumn: 53}, {MessageId: "doesNotMatchFormat", Message: "Type Method name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 79, EndLine: 1, EndColumn: 87}}},
			// parenthesized async
			{Code: "const bad_name=(async()=>{}); class C{bad_name=(async()=>{});} const obj={bad_name:(async()=>{})};", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"async"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}, {MessageId: "doesNotMatchFormat", Message: "Class Method name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 39, EndLine: 1, EndColumn: 47}, {MessageId: "doesNotMatchFormat", Message: "Object Literal Method name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 75, EndLine: 1, EndColumn: 83}}},
			// string imports
			{Code: "import { \"name\" as bad_name, \"default\" as bad_other } from \"x\";", Options: []any{map[string]any{"selector": "import", "format": []any{"camelCase"}, "modifiers": []any{"default"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Import name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 20, EndLine: 1, EndColumn: 28}, {MessageId: "doesNotMatchFormat", Message: "Import name `bad_other` must match one of the following formats: camelCase", Line: 1, Column: 43, EndLine: 1, EndColumn: 52}}},
			// rest vs shorthand
			{Code: "const {bad_name,...bad_rest}=x; function f({bad_name,...bad_rest}){}", Options: []any{map[string]any{"selector": []any{"variable", "parameter"}, "format": []any{"camelCase"}, "modifiers": []any{"destructured"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}, {MessageId: "doesNotMatchFormat", Message: "Parameter name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 45, EndLine: 1, EndColumn: 53}}},
			// typed spans
			{Code: "let bad_name!: string; function f(bad_arg?: string,...bad_rest:string[]){}", Options: []any{map[string]any{"selector": "variableLike", "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 22}, {MessageId: "doesNotMatchFormat", Message: "Parameter name `bad_arg` must match one of the following formats: camelCase", Line: 1, Column: 35, EndLine: 1, EndColumn: 51}, {MessageId: "doesNotMatchFormat", Message: "Parameter name `bad_rest` must match one of the following formats: camelCase", Line: 1, Column: 55, EndLine: 1, EndColumn: 63}}},
			// type string | null/string
			{Code: "declare const bad_name: string | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 38}}},
			// type boolean/boolean
			{Code: "declare const bad_name: boolean;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 32}}},
			// type true | false/boolean
			{Code: "declare const bad_name: true | false;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"boolean"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 37}}},
			// type "a" | "b"/string
			{Code: "declare const bad_name: \"a\" | \"b\";", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"string"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 34}}},
			// type 1 | 2/number
			{Code: "declare const bad_name: 1 | 2;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"number"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 30}}},
			// type () => void/function
			{Code: "declare const bad_name: () => void;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 35}}},
			// type (() => void) | null/function
			{Code: "declare const bad_name: (() => void) | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"function"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 44}}},
			// type string[] | null/array
			{Code: "declare const bad_name: string[] | null;", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 40}}},
			// type readonly string[]/array
			{Code: "declare const bad_name: readonly string[];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 42}}},
			// type [string,number]/array
			{Code: "declare const bad_name: [string,number];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 40}}},
			// type number[] | string[]/array
			{Code: "declare const bad_name: number[] | string[];", Options: []any{map[string]any{"selector": "variable", "format": []any{"camelCase"}, "types": []any{"array"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 44}}},
			// export const bad_name=1; export {bad_name};
			{Code: "const bad_name=1; export {bad_name};", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"exported"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// export const bad_name=1; export default bad_name;
			{Code: "const bad_name=1; export default bad_name;", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"exported"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// export const bad_name=1; export {bad_name}; function f(){const bad_name=2;}
			{Code: "const bad_name=1; export {bad_name}; function f(){const bad_name=2;}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"exported"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// export namespace N {const bad_name=1;export {bad_name};}
			{Code: "namespace N {const bad_name=1;export {bad_name};}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"exported"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 20, EndLine: 1, EndColumn: 28}}},
			// global const bad_name=1;
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": []any{"variable", "function"}, "format": []any{"camelCase"}, "modifiers": []any{"global"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// global for(var bad_name=0;bad_name<1;bad_name++) {}
			{Code: "for(var bad_name=0;bad_name<1;bad_name++) {}", Options: []any{map[string]any{"selector": []any{"variable", "function"}, "format": []any{"camelCase"}, "modifiers": []any{"global"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 9, EndLine: 1, EndColumn: 17}}},
			// unused let bad_name=1;bad_name=2;
			{Code: "let bad_name=1;bad_name=2;", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// unused let bad_name=1;bad_name++;
			{Code: "let bad_name=1;bad_name++;", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// unused function bad_func(){bad_func();}
			{Code: "function bad_func(){bad_func();}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_func` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// unused type bad_type={next:bad_type};
			{Code: "type bad_type={next:bad_type};", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_type` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// unused interface bad_type {next:bad_type}
			{Code: "interface bad_type {next:bad_type}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Interface name `bad_type` must match one of the following formats: camelCase", Line: 1, Column: 11, EndLine: 1, EndColumn: 19}}},
			// unused const bad_func=()=>bad_func();
			{Code: "const bad_func=()=>bad_func();", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_func` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// unused class C {set good(bad_arg:string){}}
			{Code: "class C {set good(bad_arg:string){}}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `C` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 8}}},
			// unused const bad_value=1;type Good=typeof bad_value;
			{Code: "const bad_value=1;type Good=typeof bad_value;", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_value` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}, {MessageId: "doesNotMatchFormat", Message: "Type Alias name `Good` must match one of the following formats: camelCase", Line: 1, Column: 24, EndLine: 1, EndColumn: 28}}},
			// unused enum bad_enum {good}
			{Code: "enum bad_enum {good}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Enum name `bad_enum` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// numeric names
			{Code: "const o={0x10:1,1e3:2,1_000:3};", Options: []any{map[string]any{"selector": "objectLiteralProperty", "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Object Literal Property name `16` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 14}, {MessageId: "doesNotMatchFormat", Message: "Object Literal Property name `1000` must match one of the following formats: camelCase", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}, {MessageId: "doesNotMatchFormat", Message: "Object Literal Property name `1000` must match one of the following formats: camelCase", Line: 1, Column: 23, EndLine: 1, EndColumn: 28}}},
			// unicode spans
			{Code: "const good=\"😀\"; const bad_name=1;\r\nfunction f(bad_arg?:string){}", Options: []any{map[string]any{"selector": "variableLike", "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 24, EndLine: 1, EndColumn: 32}, {MessageId: "doesNotMatchFormat", Message: "Parameter name `bad_arg` must match one of the following formats: camelCase", Line: 2, Column: 12, EndLine: 2, EndColumn: 27}}},
			// accepted autoAccessor types "value"
			{Code: "class C { accessor bad_name=\"value\"; }", Options: []any{map[string]any{"selector": "autoAccessor", "format": []any{"camelCase"}, "types": []any{"string"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Auto Accessor name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 20, EndLine: 1, EndColumn: 28}}},
			// accepted accessor types "value"
			{Code: "class C { accessor bad_name=\"value\"; }", Options: []any{map[string]any{"selector": "accessor", "format": []any{"camelCase"}, "types": []any{"string"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Auto Accessor name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 20, EndLine: 1, EndColumn: 28}}},
			// modifier public: interface Good {readonly bad_name:()=>void}
			{Code: "interface Good {readonly bad_name:()=>void}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"public"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Method name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 26, EndLine: 1, EndColumn: 34}}},
			// modifier public: type Good={readonly bad_name:()=>void}
			{Code: "type Good={readonly bad_name:()=>void}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"public"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Method name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 21, EndLine: 1, EndColumn: 29}}},
			// modifier public: class Good {readonly bad_name=()=>{}}
			{Code: "class Good {readonly bad_name=()=>{}}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"public"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class Method name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 22, EndLine: 1, EndColumn: 30}}},
			// modifier readonly: class Good {readonly bad_name=()=>{}}
			{Code: "class Good {readonly bad_name=()=>{}}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"readonly"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class Method name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 22, EndLine: 1, EndColumn: 30}}},
			// unused boundary let bad_name=1;bad_name=bad_name+1;
			{Code: "let bad_name=1;bad_name=bad_name+1;", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// unused boundary let bad_name=1;bad_name+=1;
			{Code: "let bad_name=1;bad_name+=1;", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// unused boundary function good(){for(const bad_key in obj){return true;foo();}}
			{Code: "function good(){for(const bad_key in obj){return true;foo();}}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_key` must match one of the following formats: camelCase", Line: 1, Column: 27, EndLine: 1, EndColumn: 34}}},
			// unused boundary class bad_class {static value=bad_class;}
			{Code: "class bad_class {static value=bad_class;}", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}}},
			// regexp rendering "/"
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": "/", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /\\//u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// regexp rendering "\n"
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": "\n", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /\\n/u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// regexp rendering "\r"
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": "\r", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /\\r/u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// regexp rendering " "
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": " ", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /\\u2028/u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// regexp rendering " "
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": " ", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /\\u2029/u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// regexp rendering "\\/"
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": "\\/", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /\\//u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// regexp rendering "a/b"
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": "a/b", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /a\\/b/u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// regexp rendering "\\\\/"
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": "\\\\/", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /\\\\\\//u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// unused type predicate
			{Code: "function good(bad_arg:unknown): bad_arg is string { return true; }", Options: []any{map[string]any{"selector": "default", "format": []any{"camelCase"}, "modifiers": []any{"unused"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Parameter name `bad_arg` must match one of the following formats: camelCase", Line: 1, Column: 15, EndLine: 1, EndColumn: 30}}},
			// regexp character class "[/]"
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": "[/]", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /[/]/u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// regexp character class "[\\/]"
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": "[\\/]", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /[\\/]/u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// regexp character class "\\[/"
			{Code: "const bad_name=1;", Options: []any{map[string]any{"selector": "variable", "format": nil, "custom": map[string]any{"regex": "\\[/", "match": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom", Message: "Variable name `bad_name` must match the RegExp: /\\[\\//u", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
		})
}
