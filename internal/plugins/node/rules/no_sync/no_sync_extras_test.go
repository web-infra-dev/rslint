package no_sync

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSyncExtras(t *testing.T) {
	rule_tester.RunRuleTester(syncRoot(t), "tsconfig.json", t, &NoSyncRule,
		[]rule_tester.ValidTestCase{
			// Package patterns retain JavaScript Unicode regexp semantics.
			{Code: `import {fooSync} from "aaa"; fooSync();`, Options: map[string]any{"ignores": []any{map[string]any{"from": "package", "package": `\u{61}aa`}}}},
			// literals and constructors
			{Code: "fs[\"readFileSync\"](); fs[`readFileSync`](); new fooSync(); new fs.fooSync(); tagSync`text`;", FileName: "input.js"},
			// private identifiers
			{Code: "class C { #fooSync() {} run() { this.#fooSync(); } }", FileName: "input.js"},
			// root blocks and class initializers
			{Code: "if (true) { fs.fooSync(); } class C { x=fooSync(); static y=fooSync(); static {fooSync();} [fooSync()](){} }", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}},
			// function nested in root callee
			{Code: "factory(() => fooSync).run();", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}},
			// TypeScript callee wrappers
			{Code: "(fooSync as any)(); fooSync!(); (fooSync satisfies Function)(); (fooSync<string>)();", FileName: "input.ts"},
			// string ignore exact spelling
			{Code: "fs.fooSync(); barSync();", FileName: "input.js", Options: []any{map[string]any{"ignores": []any{"fooSync", "barSync"}}}},
			// qualified object name
			{Code: "class Task { runSync(){} } new Task().runSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{"Task.runSync"}}}}}},
			// unresolved intrinsic type
			{Code: "unknownSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}}},
			// union has no declaring symbol
			{Code: "declare let fooSync: (()=>void)|(()=>number); fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}}},
			// local file glob
			{Code: "import {fooSync} from \"./foo\"; fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "path": "**/foo.ts"}}}}},
			// package without name
			{Code: "import {fooSync} from \"aaa\"; fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "package"}}}}},
			// package regular expression
			{Code: "import {fooSync} from \"aaa\"; fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "package", "package": "a+"}}}}},
			// import alias full name
			{Code: "import {fooSync as aliasSync} from \"aaa\"; aliasSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "package", "package": "aaa", "name": []any{"fooSync"}}}}}},
			// mixed ignores after type resolution
			{Code: "class Task { runSync(){} } new Task().runSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "lib"}, "runSync"}}}},
			// constrained generic type
			{Code: "declare function original():void; function run<T extends typeof original>(fooSync:T){fooSync()}", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{"original"}}}}}},
			// namespace full name stops at module
			{Code: "namespace Local { export function fooSync(){} } Local.fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{"fooSync"}}}}}},
		},
		[]rule_tester.InvalidTestCase{
			// empty options
			{Code: "fooSync();", FileName: "input.js", Options: []any{map[string]any{}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 1, 1, 10)}},
			// explicit defaults
			{Code: "fs.fooSync();", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": false, "ignores": []any{}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 1, 1, 11)}},
			// suffix is case sensitive
			{Code: "Sync(); sync(); fooSYNC(); SyncLater(); f\\u006foSync();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("Sync", 1, 1, 1, 7), syncError("fooSync", 1, 41, 1, 55)}},
			// computed identifiers
			{Code: "fs[(fooSync)]();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 1, 1, 14)}},
			// direct identifier arguments
			{Code: "consume(fooSync); consume(...fooSync); consume({fooSync});", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 1, 1, 17)}},
			// callee descendants and duplicate shorthand nodes
			{Code: "factory({fooSync}).run();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 10, 1, 17), syncError("fooSync", 1, 10, 1, 17)}},
			// callee descendants in destructuring
			{Code: "factory(({fooSync}) => 1).run();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 11, 1, 18), syncError("fooSync", 1, 11, 1, 18)}},
			// nested callee calls
			{Code: "factory(fooSync).run(); fs[fooSync()]();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 1, 1, 17), syncError("fooSync", 1, 28, 1, 37)}},
			// multiple matches
			{Code: "fs.fooSync.barSync(otherSync);", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 1, 1, 11), syncError("barSync", 1, 1, 1, 19), syncError("otherSync", 1, 1, 1, 30)}},
			// optional and parenthesized chains
			{Code: "fs?.fooSync(); fs.fooSync?.(); fooSync?.(); (fs?.fooSync)(); (fooSync)?.();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 1, 1, 12), syncError("fooSync", 1, 16, 1, 26), syncError("fooSync", 1, 32, 1, 43), syncError("fooSync", 1, 62, 1, 75)}},
			// parenthesized callees
			{Code: "(fs.fooSync)(); ((fooSync))();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 2, 1, 12), syncError("fooSync", 1, 17, 1, 30)}},
			// function boundaries
			{Code: "function f(x=fooSync()){ fs.fooSync(); } const a=()=>fooSync(); const b={m(){fooSync()},get x(){return fooSync()}}; class C{constructor(){fooSync()} set x(v){fooSync()}}", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 14, 1, 23), syncError("fooSync", 1, 26, 1, 36), syncError("fooSync", 1, 54, 1, 63), syncError("fooSync", 1, 78, 1, 87), syncError("fooSync", 1, 104, 1, 113), syncError("fooSync", 1, 139, 1, 148), syncError("fooSync", 1, 159, 1, 168)}},
			// root callee inside outer function
			{Code: "function f(){factory(() => fooSync).run();}", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 22, 1, 35)}},
			// nested computed method key
			{Code: "function f(){ class C { [fooSync()]() {} } }", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 26, 1, 35)}},
			// shadowing and escaped identifiers
			{Code: "function f(fooSync){fooSync();} obj.f\\u006foSync();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 21, 1, 30), syncError("fooSync", 1, 33, 1, 49)}},
			// JSX identifier boundaries
			{Code: "factory(<FooSync attrSync={fooSync}/>).run();", FileName: "input.jsx", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 27, 1, 36)}},
			// TypeScript member descendants
			{Code: "(fooSync as any).call(); fs.fooSync<string>(); fs.read<TypeSync>();", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 2, 1, 16), syncError("fooSync", 1, 26, 1, 36)}},
			// JSDoc wrappers
			{Code: "(/** @type {any} */ (fooSync))(); (/** @satisfies {Function} */ (fs.fooSync))();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 1, 1, 33), syncError("fooSync", 1, 66, 1, 76)}},
			// multiline UTF-16 ranges
			{Code: "\"😀\"; fs.\n  fooSync(\n);\n((fooSync))(\n);", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 7, 2, 10), syncError("fooSync", 4, 1, 5, 2)}},
			// ignored object names still resolve message
			{Code: "class Task { runSync(){} } new Task().runSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{"other"}}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("Task.runSync", 1, 28, 1, 46)}},
			// string names use identifier spelling
			{Code: "class Task { runSync(){} } new Task().runSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{"Task.runSync"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("runSync", 1, 28, 1, 46)}},
			// empty name list
			{Code: "declare function fooSync():void; fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{}}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 34, 1, 43)}},
			// relative file pattern
			{Code: "import {fooSync} from \"./foo\"; fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "path": "foo.ts"}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 32, 1, 41)}},
			// File patterns preserve leading, trailing, and Unicode spaces.
			{Code: `import {fooSync} from "./foo"; fooSync();`, Options: map[string]any{"ignores": []any{
				map[string]any{"from": "file", "path": " **/foo.ts"},
				map[string]any{"from": "file", "path": "**/foo.ts "},
				map[string]any{"from": "file", "path": "\u00a0**/foo.ts"},
			}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 32, 1, 41)}},
			// empty file pattern (upstream throws; native ignores nothing)
			{Code: "import {fooSync} from \"./foo\"; fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "path": ""}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 32, 1, 41)}},
			// wrong package
			{Code: "import {fooSync} from \"aaa\"; fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "package", "package": "bbb"}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 30, 1, 39)}},
			// local file is not a package
			{Code: "declare function fooSync():void; fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "package"}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 34, 1, 43)}},
			// missing from property
			{Code: "declare function fooSync():void; fooSync();", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"path": "**/*.ts"}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 34, 1, 43)}},
		},
	)
}

func TestNoSyncWithoutTypeInfo(t *testing.T) {
	r := NoSyncRule
	r.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return NoSyncRule.Run(ctx, options)
	}
	rule_tester.RunRuleTester(syncRoot(t), "tsconfig.json", t, &r,
		[]rule_tester.ValidTestCase{
			{Code: "fooSync();", Options: map[string]any{"ignores": []any{"fooSync"}}},
			{Code: "fooSync();", Options: map[string]any{"allowAtRootLevel": true}},
		},
		[]rule_tester.InvalidTestCase{
			// Without a checker, object ignores cannot suppress a diagnostic.
			{Code: "fooSync();", Options: map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 1, 1, 10)}},
		},
	)
}

func TestNoSyncFileGlobDifference(t *testing.T) {
	rule_tester.RunRuleTester(syncRoot(t), "tsconfig.json", t, &NoSyncRule, nil,
		[]rule_tester.InvalidTestCase{
			// Upstream picomatch accepts regex-style groups. The existing glob
			// matcher treats these parentheses literally; use **/foo.ts instead.
			{Code: `import {fooSync} from "./foo"; fooSync();`, Options: map[string]any{"ignores": []any{map[string]any{"from": "file", "path": "**/foo.(ts)"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 32, 1, 41)}},
		},
	)
}

func TestNoSyncParentShapes(t *testing.T) {
	rule_tester.RunRuleTester(syncRoot(t), "tsconfig.json", t, &NoSyncRule,
		[]rule_tester.ValidTestCase{
			// JSX dotted and namespaced names
			{Code: "factory(<ns:FooSync tag:attrSync={1}><Foo.BarSync /></ns:FooSync>).run();", FileName: "input.jsx"},
			// method decorators are outside the function value
			{Code: "class A{@factory(fooSync) method(){}}", FileName: "input.ts", Options: []any{map[string]any{"allowAtRootLevel": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// function parameters
			{Code: "factory((fooSync)=>{}).run(); factory(function(fooSync){}).run();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 9, 1, 22), syncError("fooSync", 1, 39, 1, 58)}},
			// rest and array patterns
			{Code: "factory((...fooSync)=>{}).run(); factory(([fooSync, ...restSync])=>{}).run();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 10, 1, 20), syncError("fooSync", 1, 43, 1, 65), syncError("restSync", 1, 53, 1, 64)}},
			// binding defaults and property aliases
			{Code: "factory(({fooSync=1,x=otherSync,key:renamedSync=otherSync})=>{}).run();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 11, 1, 20), syncError("fooSync", 1, 11, 1, 20), syncError("otherSync", 1, 21, 1, 32), syncError("renamedSync", 1, 37, 1, 58), syncError("otherSync", 1, 37, 1, 58)}},
			// assignment shorthand default
			{Code: "factory(({x = fooSync} = obj)).run();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 11, 1, 22)}},
			// method function values
			{Code: "factory({m(fooSync){},set x(fooSync){}}).run();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 11, 1, 22), syncError("fooSync", 1, 28, 1, 39)}},
			// constructors and generic method values
			{Code: "factory(class {constructor(fooSync){} method<T>(fooSync:T){}}).run();", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 27, 1, 38), syncError("fooSync", 1, 45, 1, 61)}},
			// parameter properties
			{Code: "factory(class {constructor(public fooSync:string){}}).run();", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 28, 1, 49)}},
			// typed parameter defaults
			{Code: "factory((fooSync:number=1)=>{}).run();", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 10, 1, 26)}},
			// type nodes inside member callees
			{Code: "factory(function <TypeSync>(value: TypeSync){}).run();", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{syncError("TypeSync", 1, 19, 1, 27), syncError("TypeSync", 1, 36, 1, 44)}},
		},
	)
}

func TestNoSyncSymbolsAndSyntax(t *testing.T) {
	rule_tester.RunRuleTester(syncRoot(t), "tsconfig.json", t, &NoSyncRule,
		[]rule_tester.ValidTestCase{
			// anonymous name ignore
			{Code: "const fooSync=()=>{}; fooSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{"__function"}}}}},
			// cache entries stay independent
			{Code: "import {fooSync} from \"aaa\"; fooSync(); fooSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "file"}, map[string]any{"from": "package", "package": "aaa"}}}},
		},
		[]rule_tester.InvalidTestCase{
			// object method symbol
			{Code: "const obj={fooSync(){}}; obj.fooSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("__object.fooSync", 1, 26, 1, 37)}},
			// anonymous function symbol
			{Code: "const fooSync=()=>{}; fooSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("__function", 1, 23, 1, 32)}},
			// call signature symbol
			{Code: "type Fn=()=>void; declare const fooSync:Fn; fooSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("__type", 1, 45, 1, 54)}},
			// type literal method symbol
			{Code: "declare const obj:{fooSync():void}; obj.fooSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("__type.fooSync", 1, 37, 1, 48)}},
			// anonymous class symbol
			{Code: "const Task=class {fooSync(){}}; new Task().fooSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("__class.fooSync", 1, 33, 1, 51)}},
			// private method type alias
			{Code: "class Task { #run(){} go(){ const fooSync=this.#run; fooSync(); } }", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("Task.#run", 1, 54, 1, 63)}},
			// user underscores
			{Code: "function __fooSync(){} __fooSync(); function ___fooSync(){} ___fooSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("__fooSync", 1, 24, 1, 35), syncError("___fooSync", 1, 61, 1, 73)}},
			// same type through aliases
			{Code: "function work(){} const fooSync=work,barSync=work; fooSync(); barSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{"fooSync"}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("work", 1, 52, 1, 61), syncError("work", 1, 63, 1, 72)}},
			// intrinsic type names
			{Code: "declare const fooSync: any, barSync: any; fooSync(); barSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "lib", "name": []any{"fooSync"}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("barSync", 1, 54, 1, 63)}},
			// cache types stay independent
			{Code: "import {fooSync} from \"./foo\"; function barSync(){} fooSync(); barSync(); fooSync(); barSync();", FileName: "input.ts", Options: map[string]any{"ignores": []any{map[string]any{"from": "file", "path": "**/foo.ts"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("barSync", 1, 64, 1, 73), syncError("barSync", 1, 86, 1, 95)}},
			// abstract and optional parameters
			{Code: "factory(class { abstract m(fooSync:string):void; }).run(); factory((fooSync?:number)=>{}).run();", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 27, 1, 49), syncError("fooSync", 1, 68, 1, 89)}},
			// binding and type literal keys
			{Code: "factory(({fooSync}: {fooSync:number})=>{}).run();", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 11, 1, 18), syncError("fooSync", 1, 11, 1, 18), syncError("fooSync", 1, 22, 1, 36)}},
			// computed method keys
			{Code: "factory(class { [fooSync](){} }).run();", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 17, 1, 30)}},
			// optional member descendant
			{Code: "factory((fs?.fooSync)).run();", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 10, 1, 21)}},
			// computed property and label
			{Code: "factory(()=>{ fooSync: while(false) {break fooSync;} return { [barSync]:1 }; }).run();", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 15, 1, 53), syncError("fooSync", 1, 38, 1, 52), syncError("barSync", 1, 63, 1, 74)}},
			// parameter decorator root
			{Code: "class Task{ method(@factory(fooSync) value:string){} }", FileName: "input.ts", Options: map[string]any{"allowAtRootLevel": true}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 21, 1, 37)}},
			// nested optional call root
			{Code: "function f(){ (fs?.fooSync)(); fs?.fooSync(); }", FileName: "input.ts", Options: map[string]any{"allowAtRootLevel": true}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 32, 1, 43)}},
		},
	)
}
