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

// Cover absent initializers and the scope boundary of class decorators.
// Full diagnostic expectations were checked against typescript-eslint 8.67.0.
func TestNamingConventionReferenceBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NamingConventionRule,
		[]rule_tester.ValidTestCase{
			// uninitialized let bad_name; console.log(bad_name);
			{Code: "let bad_name;\nconsole.log(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// uninitialized let bad_name; function read(){return bad_name;}
			{Code: "let bad_name;\nfunction read(){return bad_name;}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// uninitialized let bad_name: number; console.log(bad_name);
			{Code: "let bad_name: number;\nconsole.log(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// uninitialized let bad_name: number; function read(){return bad_name;}
			{Code: "let bad_name: number;\nfunction read(){return bad_name;}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// uninitialized var bad_name; console.log(bad_name);
			{Code: "var bad_name;\nconsole.log(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// uninitialized var bad_name; function read(){return bad_name;}
			{Code: "var bad_name;\nfunction read(){return bad_name;}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// uninitialized declare let bad_name: number; console.log(bad_name);
			{Code: "declare let bad_name: number;\nconsole.log(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// uninitialized declare let bad_name: number; function read(){return bad_name;}
			{Code: "declare let bad_name: number;\nfunction read(){return bad_name;}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// for const of console.log(bad_item);
			{Code: "for (const bad_item of [1]) { console.log(bad_item); }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// for let of console.log(bad_item);
			{Code: "for (let bad_item of [1]) { console.log(bad_item); }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// for var of console.log(bad_item);
			{Code: "for (var bad_item of [1]) { console.log(bad_item); }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// for const in console.log(bad_item);
			{Code: "for (const bad_item in {a:1}) { console.log(bad_item); }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// for let in console.log(bad_item);
			{Code: "for (let bad_item in {a:1}) { console.log(bad_item); }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// for var in console.log(bad_item);
			{Code: "for (var bad_item in {a:1}) { console.log(bad_item); }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// review loop
			{Code: "for (const item of [1]) {\n  console.log(item);\n}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// destructured loop
			{Code: "for (const {bad_item} of [{bad_item:1}]) {console.log(bad_item);}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// renamed destructured loop
			{Code: "for (const {value:bad_item} of [{value:1}]) {console.log(bad_item);}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// array loop
			{Code: "for (const [bad_item] of [[1]]) {console.log(bad_item);}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// loop closure
			{Code: "for (const bad_item of [1]) {(() => bad_item)();}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// for init read
			{Code: "for (let bad_item: number; bad_item < 1; bad_item++) {}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// class references @register(() => bad_class) class bad_class {}
			{Code: "@register(() => bad_class)\nclass bad_class {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// class references @register(bad_class) class bad_class {}
			{Code: "@register(bad_class)\nclass bad_class {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// class references @register(() => (bad_class)) class bad_class {}
			{Code: "@register(() => (bad_class))\nclass bad_class {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// class references @first(() => bad_class) @second() export class bad_class {}
			{Code: "@first(() => bad_class)\n@second()\nexport class bad_class {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// class references @first() /* comment */ @second(() => bad_class) class bad_class {}
			{Code: "@first()\n/* comment */ @second(() => bad_class)\nclass bad_class {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// class references @register(() => bad_class) class bad_class {method(){return bad_class;}}
			{Code: "@register(() => bad_class)\nclass bad_class {method(){return bad_class;}}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// class references class bad_class { method(){return bad_class;} } console.log(bad_class);
			{Code: "class bad_class { method(){return bad_class;} }\nconsole.log(bad_class);", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// class references const good = class bad_class {method(){return bad_class;}};
			{Code: "const good = class bad_class {method(){return bad_class;}};", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// numeric initializer read
			{Code: "const bad_name=1; console.log(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// type external read
			{Code: "type bad_type={next:bad_type}; declare const good:bad_type;", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
		}, []rule_tester.InvalidTestCase{
			// uninitialized let bad_name;
			{Code: "let bad_name;\n", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// uninitialized let bad_name; bad_name=1;
			{Code: "let bad_name;\nbad_name=1;", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// uninitialized let bad_name; bad_name++;
			{Code: "let bad_name;\nbad_name++;", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// uninitialized let bad_name: number;
			{Code: "let bad_name: number;\n", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 21}}},
			// uninitialized let bad_name: number; bad_name=1;
			{Code: "let bad_name: number;\nbad_name=1;", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 21}}},
			// uninitialized let bad_name: number; bad_name++;
			{Code: "let bad_name: number;\nbad_name++;", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 21}}},
			// uninitialized var bad_name;
			{Code: "var bad_name;\n", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// uninitialized var bad_name; bad_name=1;
			{Code: "var bad_name;\nbad_name=1;", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// uninitialized var bad_name; bad_name++;
			{Code: "var bad_name;\nbad_name++;", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// uninitialized declare let bad_name: number;
			{Code: "declare let bad_name: number;\n", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 13, EndLine: 1, EndColumn: 29}}},
			// uninitialized declare let bad_name: number; bad_name=1;
			{Code: "declare let bad_name: number;\nbad_name=1;", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 13, EndLine: 1, EndColumn: 29}}},
			// uninitialized declare let bad_name: number; bad_name++;
			{Code: "declare let bad_name: number;\nbad_name++;", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 13, EndLine: 1, EndColumn: 29}}},
			// for const of bad_item=1;
			{Code: "for (const bad_item of [1]) { bad_item=1; }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 12, EndLine: 1, EndColumn: 20}}},
			// for const of
			{Code: "for (const bad_item of [1]) {  }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 12, EndLine: 1, EndColumn: 20}}},
			// for let of bad_item=1;
			{Code: "for (let bad_item of [1]) { bad_item=1; }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// for let of
			{Code: "for (let bad_item of [1]) {  }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// for var of bad_item=1;
			{Code: "for (var bad_item of [1]) { bad_item=1; }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// for var of
			{Code: "for (var bad_item of [1]) {  }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// for const in bad_item=1;
			{Code: "for (const bad_item in {a:1}) { bad_item=1; }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 12, EndLine: 1, EndColumn: 20}}},
			// for const in
			{Code: "for (const bad_item in {a:1}) {  }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 12, EndLine: 1, EndColumn: 20}}},
			// for let in bad_item=1;
			{Code: "for (let bad_item in {a:1}) { bad_item=1; }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// for let in
			{Code: "for (let bad_item in {a:1}) {  }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// for var in bad_item=1;
			{Code: "for (var bad_item in {a:1}) { bad_item=1; }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// for var in
			{Code: "for (var bad_item in {a:1}) {  }", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// loop shadowing
			{Code: "let bad_item; for (const bad_item of [1]) {console.log(bad_item);}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_item` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// class references @register(() => {const bad_class=1;return bad_class;}) class bad_class {}
			{Code: "@register(() => {const bad_class=1;return bad_class;})\nclass bad_class {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 2, Column: 7, EndLine: 2, EndColumn: 16}}},
			// class references @register(() => other_class) class bad_class {}
			{Code: "@register(() => other_class)\nclass bad_class {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 2, Column: 7, EndLine: 2, EndColumn: 16}}},
			// class references class bad_class {}
			{Code: "class bad_class {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}}},
			// class references class bad_class { method(){return bad_class;} }
			{Code: "class bad_class { method(){return bad_class;} }", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}}},
			// class references class bad_class { static value=bad_class; }
			{Code: "class bad_class { static value=bad_class; }", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}}},
			// class references class bad_class { @register(() => bad_class) method(){} }
			{Code: "class bad_class { @register(() => bad_class) method(){} }", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}}},
			// class references class bad_class { @register(() => bad_class) field=1; }
			{Code: "class bad_class { @register(() => bad_class) field=1; }", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}}},
			// class references class bad_class { [key(bad_class)](){} }
			{Code: "class bad_class { [key(bad_class)](){} }", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}}},
			// class references class bad_class extends base(bad_class) {}
			{Code: "class bad_class extends base(bad_class) {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}}},
			// named function recursive
			{Code: "function bad_function(){return bad_function();}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_function` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 22}}},
			// function initializer recursive
			{Code: "const bad_function=(() => bad_function());", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_function` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 19}}},
			// type recursion
			{Code: "type bad_type={next:bad_type};", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_type` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
		})
}

// Expectations were independently checked against typescript-eslint 8.67.0.
// Keep external-use and direct-export controls beside the self-use cases.
func TestNamingConventionUnusedAssignmentBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NamingConventionRule,
		[]rule_tester.ValidTestCase{
			// external read ()=>bad_name
			{Code: "let bad_name; bad_name=()=>bad_name; consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read function(){return bad_name;}
			{Code: "let bad_name; bad_name=function(){return bad_name;}; consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read (()=>bad_name)()
			{Code: "let bad_name; bad_name=(()=>bad_name)(); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read (function(){return bad_name;})()
			{Code: "let bad_name; bad_name=(function(){return bad_name;})(); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read ((()=>bad_name))()
			{Code: "let bad_name; bad_name=((()=>bad_name))(); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read (()=>()=>bad_name)()
			{Code: "let bad_name; bad_name=(()=>()=>bad_name)(); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read {read:()=>bad_name}
			{Code: "let bad_name; bad_name={read:()=>bad_name}; consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read {read:()=>bad_name}.read()
			{Code: "let bad_name; bad_name={read:()=>bad_name}.read(); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// self RHS consume(()=>bad_name)
			{Code: "let bad_name; bad_name=consume(()=>bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read consume(()=>bad_name)
			{Code: "let bad_name; bad_name=consume(()=>bad_name); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// self RHS new Holder(()=>bad_name)
			{Code: "let bad_name; bad_name=new Holder(()=>bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read new Holder(()=>bad_name)
			{Code: "let bad_name; bad_name=new Holder(()=>bad_name); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// self RHS (saved=()=>bad_name)
			{Code: "let bad_name; bad_name=(saved=()=>bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read (saved=()=>bad_name)
			{Code: "let bad_name; bad_name=(saved=()=>bad_name); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read (0,()=>bad_name)
			{Code: "let bad_name; bad_name=(0,()=>bad_name); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read (()=>bad_name,()=>1)
			{Code: "let bad_name; bad_name=(()=>bad_name,()=>1); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// self RHS (()=>{const callback=()=>bad_name;return callback;})()
			{Code: "let bad_name; bad_name=(()=>{const callback=()=>bad_name;return callback;})();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read (()=>{const callback=()=>bad_name;return callback;})()
			{Code: "let bad_name; bad_name=(()=>{const callback=()=>bad_name;return callback;})(); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external read (()=>consume(bad_name))()
			{Code: "let bad_name; bad_name=(()=>consume(bad_name))(); consume(bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// let bad_name=0; consume(bad_name=(()=>bad_name)());
			{Code: "let bad_name=0; consume(bad_name=(()=>bad_name)());", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// let bad_name=0; for(let good=0;good<1;good++){bad_name=(()=>bad_name)();}
			{Code: "let bad_name=0; for(let good=0;good<1;good++){bad_name=(()=>bad_name)();}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// let bad_name=0; function later(){bad_name=(()=>bad_name)();}
			{Code: "let bad_name=0; function later(){bad_name=(()=>bad_name)();}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// function* good(){let bad_name;bad_name=yield (()=>bad_name);}
			{Code: "function* good(){let bad_name;bad_name=yield (()=>bad_name);}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// let bad_name=0; bad_name=tag`${()=>bad_name}`;
			{Code: "let bad_name=0; bad_name=tag`${()=>bad_name}`;", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// let bad_name=0; bad_name=consume(()=>bad_name);
			{Code: "let bad_name=0; bad_name=consume(()=>bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// merged unused function bad_function(){} export namespace bad_function {}
			{Code: "function bad_function(){} export namespace bad_function {}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// direct exported function bad_function(){} export namespace bad_function {}
			{Code: "function bad_function(){} export namespace bad_function {}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"exported"}, "format": []any{"camelCase"}}}},
			// merged unused function bad_function(){} export namespace bad_function {export type Options={};}
			{Code: "function bad_function(){} export namespace bad_function {export type Options={};}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// direct exported function bad_function(){} export namespace bad_function {export type Options={};}
			{Code: "function bad_function(){} export namespace bad_function {export type Options={};}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"exported"}, "format": []any{"camelCase"}}}},
			// direct exported function bad_function(){} namespace bad_function {export type Options={};}
			{Code: "function bad_function(){} namespace bad_function {export type Options={};}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"exported"}, "format": []any{"camelCase"}}}},
			// merged unused export function bad_function(){} namespace bad_function {export type Options={};}
			{Code: "export function bad_function(){} namespace bad_function {export type Options={};}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// direct exported function bad_function(){} namespace Outer {export namespace bad_function {export type Options={};}}
			{Code: "function bad_function(){} namespace Outer {export namespace bad_function {export type Options={};}}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"exported"}, "format": []any{"camelCase"}}}},
			// merged unused namespace Outer {function bad_function(){} export namespace bad_function {export type Options={};}}
			{Code: "namespace Outer {function bad_function(){} export namespace bad_function {export type Options={};}}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// direct exported namespace Outer {function bad_function(){} export namespace bad_function {export type Options={};}}
			{Code: "namespace Outer {function bad_function(){} export namespace bad_function {export type Options={};}}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"exported"}, "format": []any{"camelCase"}}}},
			// direct exported function bad_function(){} export namespace other_function {export type Options={};}
			{Code: "function bad_function(){} export namespace other_function {export type Options={};}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"exported"}, "format": []any{"camelCase"}}}},
			// merged class/interface
			{Code: "class bad_class {} export interface bad_class {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// merged interface/class
			{Code: "interface bad_class {} export class bad_class {}", Options: []any{map[string]any{"selector": "interface", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// execution boundary let bad_name; class Good {field=(bad_name=()=>bad_name,0);}
			{Code: "let bad_name; class Good {field=(bad_name=()=>bad_name,0);}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// execution boundary let bad_name; class Good {field=(bad_name=(()=>bad_name)(),0);}
			{Code: "let bad_name; class Good {field=(bad_name=(()=>bad_name)(),0);}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// execution boundary let bad_name; class Good {static {bad_name=()=>bad_name;}}
			{Code: "let bad_name; class Good {static {bad_name=()=>bad_name;}}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
		}, []rule_tester.InvalidTestCase{
			// self RHS ()=>bad_name
			{Code: "let bad_name; bad_name=()=>bad_name;", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// self RHS function(){return bad_name;}
			{Code: "let bad_name; bad_name=function(){return bad_name;};", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// self RHS (()=>bad_name)()
			{Code: "let bad_name; bad_name=(()=>bad_name)();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// self RHS (function(){return bad_name;})()
			{Code: "let bad_name; bad_name=(function(){return bad_name;})();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// self RHS ((()=>bad_name))()
			{Code: "let bad_name; bad_name=((()=>bad_name))();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// self RHS (()=>()=>bad_name)()
			{Code: "let bad_name; bad_name=(()=>()=>bad_name)();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// self RHS {read:()=>bad_name}
			{Code: "let bad_name; bad_name={read:()=>bad_name};", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// self RHS {read:()=>bad_name}.read()
			{Code: "let bad_name; bad_name={read:()=>bad_name}.read();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// self RHS (0,()=>bad_name)
			{Code: "let bad_name; bad_name=(0,()=>bad_name);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// self RHS (()=>bad_name,()=>1)
			{Code: "let bad_name; bad_name=(()=>bad_name,()=>1);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// self RHS (()=>consume(bad_name))()
			{Code: "let bad_name; bad_name=(()=>consume(bad_name))();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// let bad_name=0; bad_name+=(()=>bad_name)();
			{Code: "let bad_name=0; bad_name+=(()=>bad_name)();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// let bad_name; bad_name=new (function(){return bad_name;})();
			{Code: "let bad_name; bad_name=new (function(){return bad_name;})();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// let bad_name=0; bad_name=(()=>{const bad_name=1;return bad_name;})();
			{Code: "let bad_name=0; bad_name=(()=>{const bad_name=1;return bad_name;})();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// let bad_name=0; bad_name=(other=1,bad_name+1);
			{Code: "let bad_name=0; bad_name=(other=1,bad_name+1);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// merged unused function bad_function(){} namespace bad_function {export type Options={};}
			{Code: "function bad_function(){} namespace bad_function {export type Options={};}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_function` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 22}}},
			// direct exported export function bad_function(){} namespace bad_function {export type Options={};}
			{Code: "export function bad_function(){} namespace bad_function {export type Options={};}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"exported"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_function` must match one of the following formats: camelCase", Line: 1, Column: 17, EndLine: 1, EndColumn: 29}}},
			// merged unused function bad_function(){} namespace Outer {export namespace bad_function {export type Options={};}}
			{Code: "function bad_function(){} namespace Outer {export namespace bad_function {export type Options={};}}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_function` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 22}}},
			// merged unused function bad_function(){} export namespace other_function {export type Options={};}
			{Code: "function bad_function(){} export namespace other_function {export type Options={};}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_function` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 22}}},
			// merged class control
			{Code: "class bad_class {} interface bad_class {}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_class` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}}},
			// execution boundary let bad_name; class Good {[(bad_name=()=>bad_name, "field")]=0;}
			{Code: "let bad_name; class Good {[(bad_name=()=>bad_name, \"field\")]=0;}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
		})
}

// Full expectations were independently checked against typescript-eslint 8.67.0.
// Cover callback ownership, ordered RHS state, and namespace syntax projection.
func TestNamingConventionCallbackAndNamespaceBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NamingConventionRule,
		[]rule_tester.ValidTestCase{
			// named queued callback
			{Code: "let bad_name = 0;\nbad_name = (() => {\n  function read() { console.log(bad_name); }\n  queueMicrotask(read);\n  return 1;\n})();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// computed method key
			{Code: "declare function register(make: () => string): string;\nlet bad_key = 'run';\nbad_key = register(() => ({ [bad_key]() { return 'next'; } }).run());", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// nested callback assignment
			{Code: "declare function wrap(fn: () => void): () => void;\nlet bad_task: () => void = () => {};\nbad_task = (bad_task = () => bad_task(), wrap(bad_task));", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// external queued callback control
			{Code: "let bad_name=0; function read(){console.log(bad_name);} queueMicrotask(read);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// simple namespace control
			{Code: "function bad_name() {} export namespace bad_name {export type Entry=string;}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// function block let bad_name=0;bad_name=(()=>{function read(){return bad_name;}setTimeout(read,0);return 1;})();
			{Code: "let bad_name=0;bad_name=(()=>{function read(){return bad_name;}setTimeout(read,0);return 1;})();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// function block let bad_name=0;bad_name=(()=>{try {function read(){return bad_name;}queueMicrotask(read);}catch {} return 1;})();
			{Code: "let bad_name=0;bad_name=(()=>{try {function read(){return bad_name;}queueMicrotask(read);}catch {} return 1;})();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// function block let bad_name=0;bad_name=(()=>{function read(){return bad_name;}return 1;})();
			{Code: "let bad_name=0;bad_name=(()=>{function read(){return bad_name;}return 1;})();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// function block let bad_name=0;bad_name=(()=>{queueMicrotask(()=>console.log(bad_name));return 1;})();
			{Code: "let bad_name=0;bad_name=(()=>{queueMicrotask(()=>console.log(bad_name));return 1;})();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// method position declare function register(make:()=>string):string;let bad_key="run";bad_key=register(()=>({get [bad_key](){return "next";}}).run);
			{Code: "declare function register(make:()=>string):string;let bad_key=\"run\";bad_key=register(()=>({get [bad_key](){return \"next\";}}).run);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// method position declare function register(make:()=>string):string;let bad_key="run";bad_key=register(()=>({set [bad_key](value:string){},run:"next"}).run);
			{Code: "declare function register(make:()=>string):string;let bad_key=\"run\";bad_key=register(()=>({set [bad_key](value:string){},run:\"next\"}).run);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// method position declare function register(make:()=>string):string;let bad_key="run";bad_key=register(()=>{class Good{[bad_key](){return "next";}}return "next";});
			{Code: "declare function register(make:()=>string):string;let bad_key=\"run\";bad_key=register(()=>{class Good{[bad_key](){return \"next\";}}return \"next\";});", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// RHS state let bad_task:()=>void;bad_task=(bad_task=()=>bad_task(),0 as any);
			{Code: "let bad_task:()=>void;bad_task=(bad_task=()=>bad_task(),0 as any);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// RHS state declare function wrap(fn:()=>void):()=>void;let bad_task:()=>void=()=>{};bad_task=(0,bad_task=()=>bad_task(),wrap(bad_task));
			{Code: "declare function wrap(fn:()=>void):()=>void;let bad_task:()=>void=()=>{};bad_task=(0,bad_task=()=>bad_task(),wrap(bad_task));", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// RHS state declare function wrap(fn:()=>void):()=>void;let bad_task:()=>void=()=>{};bad_task=(bad_task=()=>bad_task(),wrap(bad_task));bad_task();
			{Code: "declare function wrap(fn:()=>void):()=>void;let bad_task:()=>void=()=>{};bad_task=(bad_task=()=>bad_task(),wrap(bad_task));bad_task();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// RHS state let bad_task:()=>void;bad_task=()=>bad_task();bad_task();
			{Code: "let bad_task:()=>void;bad_task=()=>bad_task();bad_task();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// RHS state let bad_task:any;let saved:any;bad_task=(saved=()=>bad_task(),0);
			{Code: "let bad_task:any;let saved:any;bad_task=(saved=()=>bad_task(),0);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// namespace shape function bad_name(){} export namespace bad_name{export namespace Inner{export type Entry=string;}}
			{Code: "function bad_name(){} export namespace bad_name{export namespace Inner{export type Entry=string;}}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// namespace shape export function bad_name(){} export namespace bad_name.Inner{export type Entry=string;}
			{Code: "export function bad_name(){} export namespace bad_name.Inner{export type Entry=string;}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// dotted direct exported modifier
			{Code: "function bad_name(){} export namespace bad_name.Inner{export type Entry=string;}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"exported"}, "format": []any{"camelCase"}}}},
		}, []rule_tester.InvalidTestCase{
			// dotted namespace
			{Code: "function bad_name() {}\nexport namespace bad_name.Inner { export type Entry = string; }", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// method body self-read control
			{Code: "let bad_name;bad_name={run(){return bad_name;}};", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
			// direct callback assignment control
			{Code: "let bad_task:()=>void; bad_task=()=>bad_task();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_task` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 22}}},
			// method position let bad_key="run";bad_key=(()=>({[bad_key](){return "next";}}).run())();
			{Code: "let bad_key=\"run\";bad_key=(()=>({[bad_key](){return \"next\";}}).run())();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_key` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 12}}},
			// method position let bad_key="run";bad_key=(()=>({run(){return bad_key;}}).run())();
			{Code: "let bad_key=\"run\";bad_key=(()=>({run(){return bad_key;}}).run())();", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_key` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 12}}},
			// RHS state let bad_task:()=>void;bad_task=()=>bad_task();let good=1;console.log(good);
			{Code: "let bad_task:()=>void;bad_task=()=>bad_task();let good=1;console.log(good);", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `bad_task` must match one of the following formats: camelCase", Line: 1, Column: 5, EndLine: 1, EndColumn: 22}}},
			// namespace shape function bad_name(){} export namespace bad_name.Inner.Deep{export type Entry=string;}
			{Code: "function bad_name(){} export namespace bad_name.Inner.Deep{export type Entry=string;}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// namespace shape function bad_name(){} namespace bad_name.Inner{export type Entry=string;}
			{Code: "function bad_name(){} namespace bad_name.Inner{export type Entry=string;}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// namespace shape function bad_name(){} namespace bad_name{export namespace Inner{export type Entry=string;}}
			{Code: "function bad_name(){} namespace bad_name{export namespace Inner{export type Entry=string;}}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// namespace shape function bad_name(){} export namespace other_name.Inner{export type Entry=string;}
			{Code: "function bad_name(){} export namespace other_name.Inner{export type Entry=string;}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
		})
}

// Full expectations were checked against typescript-eslint 8.67.0.
// Qualified namespace declarations must not hide reads of the base binding.
func TestNamingConventionDottedNamespaceReferences(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NamingConventionRule,
		[]rule_tester.ValidTestCase{
			// same-named dotted namespace type read
			{Code: "type bad_name = string;\nexport namespace bad_name.Inner {\n export type Value = bad_name;\n}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// different namespace type read
			{Code: "type bad_name=string;export namespace Other.Inner {export type Value=bad_name;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// simple namespace control
			{Code: "type bad_name=string;export namespace bad_name {export type Value=bad_name;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// non-exported dotted read
			{Code: "type bad_name=string; namespace bad_name.Inner {type value=bad_name;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// deeper dotted read
			{Code: "type bad_name=string;export namespace bad_name.Inner.Deep {export type Value=bad_name;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// forward type declaration
			{Code: "export namespace bad_name.Inner {export type Value=bad_name;} type bad_name=string;", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// recursive alias externally read
			{Code: "type bad_name={next:bad_name};export namespace bad_name.Inner{export type Value=bad_name;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// simple and dotted mixed read
			{Code: "type bad_name=string;namespace bad_name{type self=bad_name;}export namespace bad_name.Inner{export type Value=bad_name;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// interface dotted read
			{Code: "interface bad_name{} export namespace bad_name.Inner{export type Value=bad_name;}", Options: []any{map[string]any{"selector": "interface", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// class dotted read
			{Code: "class bad_name{} export namespace bad_name.Inner{export type Value=bad_name;}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// enum dotted read
			{Code: "enum bad_name{good} export namespace bad_name.Inner{export type Value=bad_name;}", Options: []any{map[string]any{"selector": "enum", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// qualified type root reads alias
			{Code: "type bad_name=string;namespace bad_name.Inner{export type T=string;}export type U=bad_name.Inner.T;", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// value read skips type-only binding for outer const
			{Code: "const bad_name=1;namespace Outer{type bad_name=string;namespace bad_name.Inner{export const z=1;}export const y=bad_name;}", Options: []any{map[string]any{"selector": "variable", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// simple namespace keeps value read beside dotted
			{Code: "type bad_name=string;namespace bad_name{export type T=1;}namespace bad_name.Inner{export const x=1;}export const root=bad_name;", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
			// class value read beside dotted
			{Code: "class bad_name{} namespace bad_name.Inner{export const x=1;} export const root=bad_name;", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}},
		}, []rule_tester.InvalidTestCase{
			// dotted namespace without type read
			{Code: "type bad_name=string;export namespace bad_name.Inner {export type Value=number;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// recursive type control
			{Code: "type bad_name={next:bad_name};", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// simple local self read
			{Code: "type bad_name=string; namespace bad_name {type value=bad_name;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// deeper dotted no read
			{Code: "type bad_name=string;export namespace bad_name.Inner.Deep {export type Value=number;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// recursive alias no external read
			{Code: "type bad_name={next:bad_name};export namespace bad_name.Inner{export type Value=number;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// simple and dotted mixed no read
			{Code: "type bad_name=string;namespace bad_name{type self=bad_name;}export namespace bad_name.Inner{export type Value=number;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// dotted shadowed alias
			{Code: "type bad_name=string;export namespace bad_name.Inner{export type bad_name=number;export type Value=bad_name;}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// interface dotted no read
			{Code: "interface bad_name{} export namespace bad_name.Inner{export type Value=string;}", Options: []any{map[string]any{"selector": "interface", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Interface name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 11, EndLine: 1, EndColumn: 19}}},
			// class dotted no read
			{Code: "class bad_name{} export namespace bad_name.Inner{export type Value=string;}", Options: []any{map[string]any{"selector": "class", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}},
			// enum dotted no read
			{Code: "enum bad_name{good} export namespace bad_name.Inner{export type Value=string;}", Options: []any{map[string]any{"selector": "enum", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Enum name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// function type query control
			{Code: "function bad_name(){} export namespace bad_name.Inner{export type Value=typeof bad_name;}", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
			// dotted namespace value read inside
			{Code: "type bad_name = string;\nnamespace bad_name.Inner {\n  export const root = bad_name;\n}", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// interface dotted value read inside
			{Code: "interface bad_name {}\nnamespace bad_name.Inner {\n  export const root = bad_name;\n}", Options: []any{map[string]any{"selector": "interface", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Interface name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 11, EndLine: 1, EndColumn: 19}}},
			// dotted namespace value read outside
			{Code: "type bad_name=string;namespace bad_name.Inner{export const x=1;}console.log(bad_name);", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// dotted namespace qualified value read outside
			{Code: "type bad_name=string;namespace bad_name.Inner{export const x=1;}console.log(bad_name.Inner.x);", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// import-equals root is a value reference
			{Code: "type bad_name=string;namespace bad_name.Inner{export const x=1;}import y = bad_name.Inner; y;", Options: []any{map[string]any{"selector": "typeAlias", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Alias name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 6, EndLine: 1, EndColumn: 14}}},
			// function qualified type root
			{Code: "function bad_name(){} namespace bad_name.Inner{export type T=string;} export type U=bad_name.Inner.T;", Options: []any{map[string]any{"selector": "function", "modifiers": []any{"unused"}, "format": []any{"camelCase"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Function name `bad_name` must match one of the following formats: camelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}}},
		})
}
