package prefer_at

import (
	"path"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Tests and examples from eslint-plugin-unicorn v79.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-at.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-at.md
func TestPreferAtUpstream(t *testing.T) {
	const (
		jsFile    = "file.js"
		tsFile    = "file.ts"
		typedFile = "typed.ts"
	)
	checkAll := map[string]any{"checkAllIndexAccess": true}
	testRule := PreferAtRule
	testRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// ESLint exercises both the syntax-only parser and parser services.
		if path.Base(ctx.SourceFile.FileName()) != typedFile {
			ctx.TypeChecker = nil
		}
		return PreferAtRule.Run(ctx, options)
	}

	suites := []struct {
		name    string
		valid   []rule_tester.ValidTestCase
		invalid []rule_tester.InvalidTestCase
	}{

		{
			name: "Keyword spacing",
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "function f() {return(_).last(getItems(/* keep */));}", FileName: jsFile,
					Output: []string{"function f() {return getItems(/* keep */).at(-1);}"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 21, 1, 29, "_.last"),
					},
				},
				{
					Code: "function f() {throw(_).last(getItems(/* keep */));}", FileName: jsFile,
					Output: []string{"function f() {throw getItems(/* keep */).at(-1);}"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 20, 1, 28, "_.last"),
					},
				},
				{
					Code: "if (flag) {}else(_.last)(getItems(/* keep */));", FileName: jsFile,
					Output: []string{"if (flag) {}else getItems(/* keep */).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 18, 1, 24, "_.last"),
					},
				},
				{
					Code: "do(_).last(getItems(/* keep */));while (flag);", FileName: jsFile,
					Output: []string{"do getItems(/* keep */).at(-1);while (flag);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 3, 1, 11, "_.last"),
					},
				},
			},
		},
		{
			name: "Last-element calls and retained comments",
			valid: []rule_tester.ValidTestCase{
				{Code: `_.last(document.querySelectorAll(/* keep */ "li"));`, FileName: jsFile},
				{Code: "function f() { return _.last((/* keep */ arguments)); }", FileName: jsFile},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "lodash.last(((items /* keep */)));", FileName: jsFile,
					Output: []string{"((items /* keep */)).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 12, "lodash.last"),
					},
				},
				{
					Code: "underscore.last((items // keep\r\n));", FileName: jsFile,
					Output: []string{"(items // keep\r\n).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 16, "underscore.last"),
					},
				},
				{
					Code: "before()\n_.last([1, /* keep */ 2]);", FileName: jsFile,
					Output: []string{"before()\n;[1, /* keep */ 2].at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 2, 1, 2, 7, "_.last"),
					},
				},
				{
					Code: "_.last(items as /* keep */ number[]);", FileName: tsFile,
					Output: []string{"(items as /* keep */ number[]).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: "utils.lastElement(getItems(/* keep */));", FileName: jsFile, Options: map[string]any{"getLastElementFunctions": []any{"utils.lastElement"}},
					Output: []string{"getItems(/* keep */).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 18, "utils.lastElement"),
					},
				},
				{
					Code: "_./* discard */last(getItems(/* keep */));", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 20, "_.last"),
					},
				},
				{
					Code: "_.last(/* discard */ getItems(/* keep */));", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: "_.last(getItems(/* keep */) /* discard */);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: "_.last /* discard */ ((/* keep */ items));", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
			},
		},
		{
			name: "Last-element call comment placement",
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "_.last(getItems(/* keep */));", FileName: jsFile,
					Output: []string{"getItems(/* keep */).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: "_.last((/* keep */ items));", FileName: jsFile,
					Output: []string{"(/* keep */ items).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
			},
		},
		{
			name: "JSDoc DOM types in JavaScript",
			valid: []rule_tester.ValidTestCase{
				{Code: "/** @param {NodeList | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }", FileName: jsFile},
				{Code: "/** @returns {HTMLCollection | (Null | Undefined)} */ function read() { return source; } const nodes = read(); nodes[nodes.length - 1];", FileName: jsFile},
				{Code: "/** @param {() => (NodeList | (null | undefined))} read */ function run(read) { const nodes = read(); return _.last(nodes); }", FileName: jsFile},
				{Code: "/** @param {(null | undefined) | HTMLCollectionOf<Element>} nodes */ function run(nodes) { return nodes?.[0]; }", FileName: jsFile, Options: checkAll},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }", FileName: jsFile,
					Output: []string{"/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 105, 1, 121),
					},
				},
				{
					Code: "/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }", FileName: jsFile,
					Output: []string{"/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 89, 1, 105),
					},
				},
			},
		},
		{
			name: "JSDoc DOM types with the TypeScript parser",
			valid: []rule_tester.ValidTestCase{
				{Code: "/** @param {NodeList | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }", FileName: tsFile},
				{Code: "/** @returns {HTMLCollection | (Null | Undefined)} */ function read() { return source; } const nodes = read(); nodes[nodes.length - 1];", FileName: tsFile},
				{Code: "/** @param {() => (NodeList | (null | undefined))} read */ function run(read) { const nodes = read(); return _.last(nodes); }", FileName: tsFile},
				{Code: "/** @param {(null | undefined) | HTMLCollectionOf<Element>} nodes */ function run(nodes) { return nodes?.[0]; }", FileName: tsFile, Options: checkAll},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }", FileName: tsFile,
					Output: []string{"/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 105, 1, 121),
					},
				},
				{
					Code: "/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }", FileName: tsFile,
					Output: []string{"/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 89, 1, 105),
					},
				},
			},
		},
		{
			name: "JSDoc DOM types with project service",
			valid: []rule_tester.ValidTestCase{
				{Code: "/** @param {NodeList | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }", FileName: typedFile},
				{Code: "/** @returns {HTMLCollection | (Null | Undefined)} */ function read() { return source; } const nodes = read(); nodes[nodes.length - 1];", FileName: typedFile},
				{Code: "/** @param {() => (NodeList | (null | undefined))} read */ function run(read) { const nodes = read(); return _.last(nodes); }", FileName: typedFile},
				{Code: "/** @param {(null | undefined) | HTMLCollectionOf<Element>} nodes */ function run(nodes) { return nodes?.[0]; }", FileName: typedFile, Options: checkAll},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }", FileName: typedFile,
					Output: []string{"/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 105, 1, 121),
					},
				},
				{
					Code: "/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }", FileName: typedFile,
					Output: []string{"/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 89, 1, 105),
					},
				},
			},
		},
		{
			name: "Assignment and update targets",
			valid: []rule_tester.ValidTestCase{
				{Code: "(array[array.length - 1] as number) = 3;", FileName: tsFile},
				{Code: "(array[array.length - 1] as number) += 3;", FileName: tsFile},
				{Code: "(array[array.length - 1] as number)++;", FileName: tsFile},
				{Code: "--(array[array.length - 1] as number);", FileName: tsFile},
				{Code: "delete (array[array.length - 1] as number);", FileName: tsFile},
				{Code: "delete (array.slice(-1)[0] as number);", FileName: tsFile},
				{Code: "(<number>array[array.length - 1]) = 3;", FileName: tsFile},
				{Code: "(<number>array[array.length - 1]) += 3;", FileName: tsFile},
				{Code: "(<number>array[array.length - 1])++;", FileName: tsFile},
				{Code: "--(<number>array[array.length - 1]);", FileName: tsFile},
				{Code: "delete (<number>array[array.length - 1]);", FileName: tsFile},
				{Code: "delete (<number>array.slice(-1)[0]);", FileName: tsFile},
				{Code: "array[array.length - 1]! = 3;", FileName: tsFile},
				{Code: "array[array.length - 1]! += 3;", FileName: tsFile},
				{Code: "array[array.length - 1]!++;", FileName: tsFile},
				{Code: "--array[array.length - 1]!;", FileName: tsFile},
				{Code: "delete array[array.length - 1]!;", FileName: tsFile},
				{Code: "delete array.slice(-1)[0]!;", FileName: tsFile},
				{Code: "(array[array.length - 1] satisfies number) = 3;", FileName: tsFile},
				{Code: "(array[array.length - 1] satisfies number) += 3;", FileName: tsFile},
				{Code: "(array[array.length - 1] satisfies number)++;", FileName: tsFile},
				{Code: "--(array[array.length - 1] satisfies number);", FileName: tsFile},
				{Code: "delete (array[array.length - 1] satisfies number);", FileName: tsFile},
				{Code: "delete (array.slice(-1)[0] satisfies number);", FileName: tsFile},
				{Code: "((array[array.length - 1] as number)!) = 3;", FileName: tsFile},
				{Code: "((array[array.length - 1] as number)!) += 3;", FileName: tsFile},
				{Code: "((array[array.length - 1] as number)!)++;", FileName: tsFile},
				{Code: "--((array[array.length - 1] as number)!);", FileName: tsFile},
				{Code: "delete ((array[array.length - 1] as number)!);", FileName: tsFile},
				{Code: "delete ((array.slice(-1)[0] as number)!);", FileName: tsFile},
				{Code: "delete array?.[array.length - 1];", FileName: jsFile},
				{Code: "delete array?.slice(-1)[0];", FileName: jsFile},
				{Code: "delete ((array?.[array.length - 1] as number)!);", FileName: tsFile},
				{Code: "([array[array.length - 1]!] = values);", FileName: tsFile},
				{Code: "({value: (array[array.length - 1] as number)} = object);", FileName: tsFile},
				{Code: "for (array[array.length - 1]! of values) {}", FileName: tsFile},
				{Code: "(array[0] as number) = 3;", FileName: tsFile, Options: checkAll},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "const value = (array[array.length - 1] as number);", FileName: tsFile,
					Output: []string{"const value = (array.at(-1) as number);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 22, 1, 38),
					},
				},
			},
		},
		{
			name: "Method, tag, and constructor contexts",
			valid: []rule_tester.ValidTestCase{
				{Code: "array[array.length - 1]()", FileName: jsFile},
				{Code: "array[array.length - 1]?.()", FileName: jsFile},
				{Code: "array[array.length - 1]`template`", FileName: jsFile},
				{Code: "array?.[array.length - 1]()", FileName: jsFile},
				{Code: "(array?.[array.length - 1])()", FileName: jsFile},
				{Code: "array.slice(-1)[0]()", FileName: jsFile},
				{Code: "array.slice(-1)[0]`template`", FileName: jsFile},
				{Code: "(array?.slice(-1)[0])()", FileName: jsFile},
				{Code: "(array[array.length - 1] as number)()", FileName: tsFile},
				{Code: "(array.slice(-1)[0] as number)`template`", FileName: tsFile},
				{Code: "(<number>array[array.length - 1])()", FileName: tsFile},
				{Code: "(<number>array.slice(-1)[0])`template`", FileName: tsFile},
				{Code: "array[array.length - 1]!()", FileName: tsFile},
				{Code: "array.slice(-1)[0]!`template`", FileName: tsFile},
				{Code: "(array[array.length - 1] satisfies number)()", FileName: tsFile},
				{Code: "(array.slice(-1)[0] satisfies number)`template`", FileName: tsFile},
				{Code: "((array[array.length - 1] as number)!)()", FileName: tsFile},
				{Code: "((array.slice(-1)[0] as number)!)`template`", FileName: tsFile},
				{Code: "array[0]()", FileName: jsFile, Options: checkAll},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "new array[array.length - 1]()", FileName: jsFile,
					Output: []string{"new (array.at(-1))()"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 11, 1, 27),
					},
				},
				{
					Code: "new array[array.length - 1].Constructor()", FileName: jsFile,
					Output: []string{"new (array.at(-1)).Constructor()"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 11, 1, 27),
					},
				},
				{
					Code: "new array[array.length - 1].tag`template`()", FileName: jsFile,
					Output: []string{"new (array.at(-1)).tag`template`()"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 11, 1, 27),
					},
				},
				{
					Code: "new (array[array.length - 1])()", FileName: jsFile,
					Output: []string{"new (array.at(-1))()"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 12, 1, 28),
					},
				},
				{
					Code: "new (array[array.length - 1].Constructor)()", FileName: jsFile,
					Output: []string{"new (array.at(-1).Constructor)()"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 12, 1, 28),
					},
				},
				{
					Code: "new array[array.length - /* keep */ 1]()", FileName: jsFile,
					Output: []string{"new (array.at(- /* keep */ 1))()"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 11, 1, 38),
					},
				},
				{
					Code: "new array[array.length - 1]!.Constructor()", FileName: tsFile,
					Output: []string{"new (array.at(-1))!.Constructor()"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 11, 1, 27),
					},
				},
				{
					Code: "new array[0]()", FileName: jsFile, Options: checkAll,
					Output: []string{"new (array.at(0))()"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 11, 1, 12),
					},
				},
			},
		},
		{
			name: "DOM collection receivers",
			valid: []rule_tester.ValidTestCase{
				{Code: "const nodes = element.children; nodes[nodes.length - 1];", FileName: jsFile},
				{Code: "const nodes = element.children; nodes.slice(-1)[0];", FileName: jsFile},
				{Code: "const nodes = element.children; _.last(nodes);", FileName: jsFile},
				{Code: "_.last(element.children);", FileName: jsFile},
				{Code: "const nodes = element.children; nodes[0];", FileName: jsFile, Options: checkAll},
				{Code: "const nodes = parent.childNodes; nodes[nodes.length - 1];", FileName: jsFile},
				{Code: "const nodes = parent.childNodes; nodes.slice(-1)[0];", FileName: jsFile},
				{Code: "const nodes = parent.childNodes; _.last(nodes);", FileName: jsFile},
				{Code: "_.last(parent.childNodes);", FileName: jsFile},
				{Code: "const nodes = parent.childNodes; nodes[0];", FileName: jsFile, Options: checkAll},
				{Code: `const nodes = document.querySelectorAll("li"); nodes[nodes.length - 1];`, FileName: jsFile},
				{Code: `const nodes = document.querySelectorAll("li"); nodes.slice(-1)[0];`, FileName: jsFile},
				{Code: `const nodes = document.querySelectorAll("li"); _.last(nodes);`, FileName: jsFile},
				{Code: `_.last(document.querySelectorAll("li"));`, FileName: jsFile},
				{Code: `const nodes = document.querySelectorAll("li"); nodes[0];`, FileName: jsFile, Options: checkAll},
				{Code: `const nodes = document.getElementsByClassName("item"); nodes[nodes.length - 1];`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByClassName("item"); nodes.slice(-1)[0];`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByClassName("item"); _.last(nodes);`, FileName: jsFile},
				{Code: `_.last(document.getElementsByClassName("item"));`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByClassName("item"); nodes[0];`, FileName: jsFile, Options: checkAll},
				{Code: `const nodes = document.getElementsByName("item"); nodes[nodes.length - 1];`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByName("item"); nodes.slice(-1)[0];`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByName("item"); _.last(nodes);`, FileName: jsFile},
				{Code: `_.last(document.getElementsByName("item"));`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByName("item"); nodes[0];`, FileName: jsFile, Options: checkAll},
				{Code: `const nodes = document.getElementsByTagName("li"); nodes[nodes.length - 1];`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByTagName("li"); nodes.slice(-1)[0];`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByTagName("li"); _.last(nodes);`, FileName: jsFile},
				{Code: `_.last(document.getElementsByTagName("li"));`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByTagName("li"); nodes[0];`, FileName: jsFile, Options: checkAll},
				{Code: `const nodes = document.getElementsByTagNameNS(namespace, "li"); nodes[nodes.length - 1];`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByTagNameNS(namespace, "li"); nodes.slice(-1)[0];`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByTagNameNS(namespace, "li"); _.last(nodes);`, FileName: jsFile},
				{Code: `_.last(document.getElementsByTagNameNS(namespace, "li"));`, FileName: jsFile},
				{Code: `const nodes = document.getElementsByTagNameNS(namespace, "li"); nodes[0];`, FileName: jsFile, Options: checkAll},
				{Code: `const nodes = document.querySelectorAll("li"); const alias = nodes; alias[alias.length - 1];`, FileName: jsFile},
				{Code: "const nodes = parent?.childNodes; nodes?.[nodes.length - 1];", FileName: jsFile},
				{Code: `const nodes = document?.querySelectorAll?.("li"); _.last(nodes);`, FileName: jsFile},
				{Code: "const nodes = element.children; nodes.slice(-1).pop();", FileName: jsFile},
				{Code: "const nodes = element.children; nodes.slice(-1).shift();", FileName: jsFile},
				{Code: "const nodes = element.children; nodes.slice(-2, -1)[0];", FileName: jsFile},
				{Code: "element.children.slice(-1)[0];", FileName: jsFile},
				{Code: "lodash.last(element.children);", FileName: jsFile},
				{Code: "underscore.last(element.childNodes);", FileName: jsFile},
				{Code: `const nodes = document.querySelectorAll("li"); utils.lastElement(nodes);`, FileName: jsFile, Options: map[string]any{"getLastElementFunctions": []any{"utils.lastElement"}}},
				{Code: "function foo(nodes: NodeList) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: tsFile},
				{Code: "function foo(nodes: NodeList) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: typedFile},
				{Code: "function foo(nodes: NodeListOf<Element>) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: tsFile},
				{Code: "function foo(nodes: NodeListOf<Element>) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: typedFile},
				{Code: "function foo(nodes: HTMLCollection) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: tsFile},
				{Code: "function foo(nodes: HTMLCollection) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: typedFile},
				{Code: "function foo(nodes: HTMLCollectionOf<Element>) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: tsFile},
				{Code: "function foo(nodes: HTMLCollectionOf<Element>) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: typedFile},
				{Code: "function foo<T extends NodeList>(nodes: T) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: tsFile},
				{Code: "function foo<T extends NodeList>(nodes: T) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: typedFile},
				{Code: `const nodes = document.querySelectorAll("li") as NodeListOf<Element>; (nodes as NodeListOf<Element>)[nodes.length - 1];`, FileName: tsFile},
				{Code: "const nodes = <NodeList>element.childNodes; _.last(<NodeList>nodes);", FileName: tsFile},
				{Code: "const nodes = element.children satisfies HTMLCollection; nodes![nodes!.length - 1];", FileName: tsFile},
				{Code: "(element.children as string[])[element.children.length - 1];\n(element.children as string[]).slice(-1)[0];\n_.last(element.children as string[]);", FileName: tsFile},
				{Code: "type Nodes = NodeListOf<Element>; function foo(nodes: Nodes) { _.last(nodes); }", FileName: tsFile},
				{Code: "function foo(nodes: NodeList | HTMLCollection) { _.last(nodes); }", FileName: tsFile},
				{Code: "function foo(nodes: NodeList | undefined) { return nodes?.[nodes.length - 1]; }", FileName: tsFile},
				{Code: "function foo(nodes: HTMLCollection | null) { return nodes?.[0]; }", FileName: tsFile, Options: checkAll},
				{Code: "function foo(nodes: NodeList) { return nodes[0]; }", FileName: tsFile, Options: checkAll},
				{Code: "declare function getNodes(): NodeListOf<Element>; const nodes = getNodes(); nodes[nodes.length - 1];", FileName: typedFile},
				{Code: "declare function getNodes(): NodeListOf<Element>; const nodes = getNodes(); _.last(nodes);", FileName: tsFile},
				{Code: "declare function getNodes(): NodeList; const nodes = getNodes?.(); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);", FileName: tsFile},
				{Code: "declare function getNodes(): NodeList; const nodes = getNodes?.(); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);", FileName: typedFile},
				{Code: "declare function getNodes(): HTMLCollection; _.last(getNodes?.());", FileName: tsFile},
				{Code: `function getNodes() { return document.querySelectorAll("li"); } const nodes = getNodes(); nodes[nodes.length - 1];`, FileName: typedFile},
				{Code: `let nodes = document.querySelectorAll("li"); _.last(nodes);`, FileName: typedFile},
				{Code: "const {children: nodes} = document.body; _.last(nodes);", FileName: typedFile},
				{Code: "function foo(value: unknown) { if (value instanceof NodeList) { return value[value.length - 1]; } }", FileName: typedFile},
				{Code: "const nodes: unknown = element.children; _.last(nodes);", FileName: tsFile},
				{Code: "const nodes: NodeList | string[] = element.children; _.last(nodes);", FileName: tsFile},
				{Code: "declare const source: {nodes: HTMLCollection}; _.last(source.nodes);", FileName: typedFile},
				{Code: "type Nodes = ReturnType<() => NodeListOf<Element>>; declare const nodes: Nodes; _.last(nodes);", FileName: typedFile},
				{Code: "function foo(nodes: NodeList | undefined) { return nodes?.[nodes.length - 1]; }", FileName: typedFile},
			},
		},
		{
			name: "Negative index access",
			valid: []rule_tester.ValidTestCase{
				{Code: "array.at(-1)", FileName: jsFile},
				{Code: "array[array.length - 0];", FileName: jsFile},
				{Code: "array[array.length + 1]", FileName: jsFile},
				{Code: "array[array.length + -1]", FileName: jsFile},
				{Code: "const modes = new Set(['foo']); modes.clear(); array[modes.size ? 1 : 'x'];", FileName: jsFile},
				{Code: "const modes = new Set(['foo']); modes.clear(); array[(modes.size && 1) || index];", FileName: jsFile},
				{Code: "const modes = new Set(['foo']); modes.clear(); array[(modes.size ? 1 : index) as number];", FileName: typedFile},
				{Code: "foo[bar.length - 1]", FileName: jsFile},
				{Code: "array[array.length - 0.5]", FileName: jsFile},
				{Code: "array[array.length - 1.5]", FileName: jsFile},
				{Code: "array[array.length - 1] = 1", FileName: jsFile},
				{Code: "array[array.length - 1] %= 1", FileName: jsFile},
				{Code: "++ array[array.length - 1]", FileName: jsFile},
				{Code: "array[array.length - 1] --", FileName: jsFile},
				{Code: "delete array[array.length - 1]", FileName: jsFile},
				{Code: "class Foo {bar; #bar; baz() {return this.#bar[this.bar.length - 1]}}", FileName: jsFile},
				{Code: "([array[array.length - 1]] = [])", FileName: jsFile},
				{Code: "({foo: array[array.length - 1] = 9} = {})", FileName: jsFile},
				{Code: "function foo() {return arguments[arguments.length - 1]}", FileName: jsFile},
				{Code: "parent?.childNodes[parent.childNodes.length - 1];", FileName: jsFile},
				{Code: "element.children[element.children.length - 1];", FileName: jsFile},
				{Code: `document.querySelectorAll("li")[document.querySelectorAll("li").length - 1];`, FileName: jsFile},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "array[array.length - 1];", FileName: jsFile,
					Output: []string{"array.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 7, 1, 23),
					},
				},
				{
					Code: "array?.[array.length - 1];", FileName: jsFile,
					Output: []string{"array?.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 9, 1, 25),
					},
				},
				{
					Code: "array[array.length -1];", FileName: jsFile,
					Output: []string{"array.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 7, 1, 22),
					},
				},
				{
					Code: "array[array.length - /* comment */ 1];", FileName: jsFile,
					Output: []string{"array.at(- /* comment */ 1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 7, 1, 37),
					},
				},
				{
					Code: "array[array.length - 1.];", FileName: jsFile,
					Output: []string{"array.at(- 1.);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 7, 1, 24),
					},
				},
				{
					Code: "array[array.length - 0b1];", FileName: jsFile,
					Output: []string{"array.at(- 0b1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 7, 1, 25),
					},
				},
				{
					Code: "array[array.length - 9];", FileName: jsFile,
					Output: []string{"array.at(-9);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 7, 1, 23),
					},
				},
				{
					Code: "array[0][array[0].length - 1];", FileName: jsFile,
					Output: []string{"array[0].at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 10, 1, 29),
					},
				},
				{
					Code: "array[(( array.length )) - 1];", FileName: jsFile,
					Output: []string{"array.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 7, 1, 29),
					},
				},
				{
					Code: "array[array.length - (( 1 ))];", FileName: jsFile,
					Output: []string{"array.at(- (( 1 )));"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 7, 1, 29),
					},
				},
				{
					Code: "array[(( array.length - 1 ))];", FileName: jsFile,
					Output: []string{"array.at((( -1 )));"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 10, 1, 26),
					},
				},
				{
					Code: "(( array ))[array.length - 1];", FileName: jsFile,
					Output: []string{"(( array )).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 13, 1, 29),
					},
				},
				{
					Code: "(( array[array.length - 1] ));", FileName: jsFile,
					Output: []string{"(( array.at(-1) ));"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 10, 1, 26),
					},
				},
				{
					Code: "array[array.length - 1].pop().shift()[0];", FileName: jsFile,
					Output: []string{"array.at(-1).pop().shift()[0];"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 7, 1, 23),
					},
				},
				{
					Code: "a = array[array.length - 1]", FileName: jsFile,
					Output: []string{"a = array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 11, 1, 27),
					},
				},
				{
					Code: "const a = array[array.length - 1]", FileName: jsFile,
					Output: []string{"const a = array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 17, 1, 33),
					},
				},
				{
					Code: "const {a = array[array.length - 1]} = {}", FileName: jsFile,
					Output: []string{"const {a = array.at(-1)} = {}"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 18, 1, 34),
					},
				},
				{
					Code: "typeof array[array.length - 1]", FileName: jsFile,
					Output: []string{"typeof array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 14, 1, 30),
					},
				},
				{
					Code: "class Foo {bar; baz() {return this.bar[this.bar.length - 1]}}", FileName: jsFile,
					Output: []string{"class Foo {bar; baz() {return this.bar.at(-1)}}"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 40, 1, 59),
					},
				},
				{
					Code: "class Foo {#bar; baz() {return this.#bar[this.#bar.length - 1]}}", FileName: jsFile,
					Output: []string{"class Foo {#bar; baz() {return this.#bar.at(-1)}}"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 42, 1, 62),
					},
				},
				{
					Code: "(array as Foo[])[array.length - 1]", FileName: tsFile,
					Output: []string{"(array as Foo[]).at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 18, 1, 34),
					},
				},
				{
					Code: "(array as Foo[])[(array as Foo[]).length - 1]", FileName: tsFile,
					Output: []string{"(array as Foo[]).at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 18, 1, 45),
					},
				},
				{
					Code: "<Foo[]>array[array.length - 1]", FileName: tsFile,
					Output: []string{"<Foo[]>array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 14, 1, 30),
					},
				},
				{
					Code: "<Foo[]>array[(<Foo[]>array).length - 1]", FileName: tsFile,
					Output: []string{"<Foo[]>array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 14, 1, 39),
					},
				},
				{
					Code: "array![array.length - 1]", FileName: tsFile,
					Output: []string{"array!.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 8, 1, 24),
					},
				},
				{
					Code: "array![array!.length - 1]", FileName: tsFile,
					Output: []string{"array!.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 8, 1, 25),
					},
				},
			},
		},
		{
			name: "String index access",
			valid: []rule_tester.ValidTestCase{
				{Code: "array[0]", FileName: jsFile},
				{Code: "unknown[0]", FileName: jsFile},
				{Code: "1[0]", FileName: jsFile},
				{Code: `({0: "a"})[0]`, FileName: jsFile},
				{Code: `const object = {0: "a"}; object[0]`, FileName: jsFile},
				{Code: "string[unknown]", FileName: jsFile},
				{Code: "string[-1]", FileName: jsFile},
				{Code: "string[1.5]", FileName: jsFile},
				{Code: "string[1n]", FileName: jsFile},
				{Code: "string[0] = value", FileName: jsFile},
				{Code: `"string"[1]`, FileName: jsFile},
				{Code: "`string`[1]", FileName: jsFile},
				{Code: "String(value)[0]", FileName: jsFile},
				{Code: "String.fromCodePoint(65)[0]", FileName: jsFile},
				{Code: "(typeof value)[0]", FileName: jsFile},
				{Code: `const string = "string"; string[1]`, FileName: jsFile},
				{Code: "String.fromCharCode(65)[0]", FileName: jsFile},
				{Code: "function foo(value: string | number) { return value[0]; }", FileName: tsFile},
				{Code: "function foo(value: number) { return value[0]; }", FileName: tsFile},
				{Code: "const value = other; const other = value; value[0];", FileName: jsFile},
				{Code: "type Value = Other; type Other = Value; function foo(value: Value) { return value[0]; }", FileName: tsFile},
				{Code: "(value as string)[0]", FileName: tsFile},
				{Code: "(value satisfies string)[0]", FileName: tsFile},
				{Code: "(<string>value)[0]", FileName: tsFile},
				{Code: "function foo(value: string) { return value![0]; }", FileName: tsFile},
				{Code: "function foo(value: string) { return value[0]; }", FileName: tsFile},
				{Code: `const value: string | number = "string"; value[0];`, FileName: tsFile},
				{Code: `const value: unknown = "string"; value[0];`, FileName: tsFile},
				{Code: "type StringAlias = string; function foo(value: StringAlias) { return value[0]; }", FileName: tsFile},
				{Code: `type StringUnion = "a" | "b"; function foo(value: StringUnion) { return value[0]; }`, FileName: tsFile},
				{Code: "declare function getValue(): string; getValue()[0];", FileName: typedFile},
				{Code: "type Value = ReturnType<() => string>; declare const value: Value; value[0];", FileName: typedFile},
				{Code: `function foo(value: string | number) { if (typeof value === "string") { return value[0]; } }`, FileName: typedFile},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: `"string"["string".length - 1]`, FileName: jsFile,
					Output: []string{`"string".at(-1)`},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 10, 1, 29),
					},
				},
				{
					Code: `const string = "string"; string[string.length - 1]`, FileName: jsFile,
					Output: []string{`const string = "string"; string.at(-1)`},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 33, 1, 50),
					},
				},
			},
		},
		{
			name: "charAt",
			valid: []rule_tester.ValidTestCase{
				{Code: "string.charAt(string.length - 0);", FileName: jsFile},
				{Code: "string.charAt(string.length + 1)", FileName: jsFile},
				{Code: "string.charAt(string.length + -1)", FileName: jsFile},
				{Code: "foo.charAt(bar.length - 1)", FileName: jsFile},
				{Code: "string?.charAt?.(string.length - 1);", FileName: jsFile},
				{Code: "string.charAt(9);", FileName: jsFile},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "string.charAt(string.length - 1);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 15, 1, 32, "string.at(- 1);"),
					},
				},
				{
					Code: "string?.charAt(string.length - 1);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 16, 1, 33, "string?.at(- 1);"),
					},
				},
				{
					Code: "string.charAt(string.length - 0o11);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 15, 1, 35, "string.at(- 0o11);"),
					},
				},
				{
					Code: "some.string.charAt(some.string.length - 1);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 20, 1, 42, "some.string.at(- 1);"),
					},
				},
				{
					Code: "string.charAt((( string.length )) - 0xFF);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 15, 1, 41, "string.at(- 0xFF);"),
					},
				},
				{
					Code: "string.charAt(string.length - (( 1 )));", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 15, 1, 38, "string.at(- (( 1 )));"),
					},
				},
				{
					Code: "string.charAt((( string.length - 1 )));", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 18, 1, 35, "string.at((( - 1 )));"),
					},
				},
				{
					Code: "(( string )).charAt(string.length - 1);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 21, 1, 38, "(( string )).at(- 1);"),
					},
				},
				{
					Code: "(( string.charAt ))(string.length - 1);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 21, 1, 38, "(( string.at ))(- 1);"),
					},
				},
				{
					Code: "(( string.charAt(string.length - 1) ));", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 18, 1, 35, "(( string.at(- 1) ));"),
					},
				},
				{
					Code: "(string as string).charAt((string as string).length - 1);", FileName: tsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 27, 1, 56, "(string as string).at(- 1);"),
					},
				},
				{
					Code: "<string>string.charAt((<string>string).length - 1);", FileName: tsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 23, 1, 50, "<string>string.at(- 1);"),
					},
				},
				{
					Code: "string!.charAt(string!.length - 1);", FileName: tsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 16, 1, 34, "string!.at(- 1);"),
					},
				},
				{
					Code: "string.charAt(string /* keep */ .length - 1);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 15, 1, 44),
					},
				},
			},
		},
		{
			name: "Single-character substring",
			valid: []rule_tester.ValidTestCase{
				{Code: "string.substring()", FileName: jsFile},
				{Code: "string.substring(0)", FileName: jsFile},
				{Code: "string.substring(index)", FileName: jsFile},
				{Code: "string.substring(0, 2)", FileName: jsFile},
				{Code: "string.substring(-1, 0)", FileName: jsFile},
				{Code: "string.substring(1.5, 2.5)", FileName: jsFile},
				{Code: "string.substring(index, index + 2)", FileName: jsFile},
				{Code: "string.substring(index, other + 1)", FileName: jsFile},
				{Code: "string.substring(other, index + 1)", FileName: jsFile},
				{Code: "string.substring(index, index + length)", FileName: jsFile},
				{Code: "string.substring(index, length + index)", FileName: jsFile},
				{Code: "string.substring(...arguments)", FileName: jsFile},
				{Code: "string.substring(index, index + 1, extraArgument)", FileName: jsFile},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "string.substring(0, 1)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 23, "string.at(0)"),
					},
				},
				{
					Code: "string.substring(1, 2)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 23, "string.at(1)"),
					},
				},
				{
					Code: "string.substring(2, 1)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 23, "string.at(1)"),
					},
				},
				{
					Code: "string.substring(index, index + 1)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 35, "string.at(index)"),
					},
				},
				{
					Code: "string.substring(index, 1 + index)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 35, "string.at(index)"),
					},
				},
				{
					Code: "string.substring(index - 1, index)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 35, "string.at(index - 1)"),
					},
				},
				{
					Code: "string.substring(index + 1, index)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 35, "string.at(index)"),
					},
				},
				{
					Code: "string.substring(1 + index, index)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 35, "string.at(index)"),
					},
				},
				{
					Code: "string.substring(index, /* comment */ index + 1)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 49),
					},
				},
				{
					Code: "string.substring(index, index + /* comment */ 1)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 49),
					},
				},
				{
					Code: "string.substring(0, /* between */ 1)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 37),
					},
				},
				{
					Code: "string.substring((( index )), (( index )) + 1)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 47, "string.at((( index )))"),
					},
				},
				{
					Code: "string.substring((( index )) - 1, (( index )))", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 47, "string.at((( index )) - 1)"),
					},
				},
				{
					Code: "(( string )).substring(index, index + 1)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 41, "(( string )).at(index)"),
					},
				},
				{
					Code: "string?.substring(index, index + 1)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 36, "string?.at(index)"),
					},
				},
				{
					Code: "string.substring?.(index, index + 1)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 1, 1, 37, "string.at?.(index)"),
					},
				},
			},
		},
		{
			name: "One-argument slice",
			valid: []rule_tester.ValidTestCase{
				{Code: "array.slice(-1)", FileName: jsFile},
				{Code: "new array.slice(-1)", FileName: jsFile},
				{Code: "array.slice(-0)[0]", FileName: jsFile},
				{Code: "array.slice(-9).pop()", FileName: jsFile},
				{Code: "array.slice(-1.1)[0]", FileName: jsFile},
				{Code: "array.slice(-1)?.[0]", FileName: jsFile},
				{Code: "array.slice?.(-1)[0]", FileName: jsFile},
				{Code: "array.notSlice(-1)[0]", FileName: jsFile},
				{Code: "array.slice()[0]", FileName: jsFile},
				{Code: "array.slice(...[-1])[0]", FileName: jsFile},
				{Code: "array.slice(-1).shift?.()", FileName: jsFile},
				{Code: "array.slice(-1)?.shift()", FileName: jsFile},
				{Code: "array.slice(-1).shift(...[])", FileName: jsFile},
				{Code: "new array.slice(-1).shift()", FileName: jsFile},
				{Code: "array.slice(-1)[0] += 1", FileName: jsFile},
				{Code: "++ array.slice(-1)[0]", FileName: jsFile},
				{Code: "array.slice(-1)[0] --", FileName: jsFile},
				{Code: "delete array.slice(-1)[0]", FileName: jsFile},
				{Code: "array.slice(-9)[0]", FileName: jsFile},
				{Code: "array.slice(-9).shift()", FileName: jsFile},
				{Code: "array.slice(-0xA)[0b000]", FileName: jsFile},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "array.slice(-1)[0]", FileName: jsFile,
					Output: []string{"array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array?.slice(-1)[0]", FileName: jsFile,
					Output: []string{"array?.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 8, 1, 13),
					},
				},
				{
					Code: "array.slice(-1).pop()", FileName: jsFile,
					Output: []string{"array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice(-1.0).shift()", FileName: jsFile,
					Output: []string{"array.at(-1.0)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice(-1)[(( 0 ))];", FileName: jsFile,
					Output: []string{"array.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice(-(( 1 )))[0];", FileName: jsFile,
					Output: []string{"array.at(-(( 1 )));"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice((( -1 )))[0];", FileName: jsFile,
					Output: []string{"array.at((( -1 )));"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "(( array.slice(-1) ))[0];", FileName: jsFile,
					Output: []string{"(( array.at(-1) ));"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 10, 1, 15),
					},
				},
				{
					Code: "(( array )).slice(-1)[0];", FileName: jsFile,
					Output: []string{"(( array )).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 13, 1, 18),
					},
				},
				{
					Code: "(( array.slice(-1)[0] ));", FileName: jsFile,
					Output: []string{"(( array.at(-1) ));"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 10, 1, 15),
					},
				},
				{
					Code: "(( array.slice(-1) )).pop();", FileName: jsFile,
					Output: []string{"(( array.at(-1) ));"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 10, 1, 15),
					},
				},
				{
					Code: "(( array.slice(-1).pop ))();", FileName: jsFile,
					Output: []string{"(( array.at(-1) ));"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 10, 1, 15),
					},
				},
				{
					Code: "(( array.slice(-1).pop() ));", FileName: jsFile,
					Output: []string{"(( array.at(-1) ));"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 10, 1, 15),
					},
				},
				{
					Code: "array.slice(-1)[0].pop().shift().slice(-1)", FileName: jsFile,
					Output: []string{"array.at(-1).pop().shift().slice(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "(array as Foo[]).slice(-1)[0]", FileName: tsFile,
					Output: []string{"(array as Foo[]).at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 18, 1, 23),
					},
				},
				{
					Code: "(array as Foo[]).slice(-1).pop()", FileName: tsFile,
					Output: []string{"(array as Foo[]).at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 18, 1, 23),
					},
				},
				{
					Code: "(<Foo[]>array).slice(-1).pop()", FileName: tsFile,
					Output: []string{"(<Foo[]>array).at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 16, 1, 21),
					},
				},
				{
					Code: "<Foo[]>array.slice(-1)[0]", FileName: tsFile,
					Output: []string{"<Foo[]>array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 14, 1, 19),
					},
				},
				{
					Code: "<Foo[]>array.slice(-1).shift()", FileName: tsFile,
					Output: []string{"<Foo[]>array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 14, 1, 19),
					},
				},
				{
					Code: "array!.slice(-1)[0]", FileName: tsFile,
					Output: []string{"array!.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 8, 1, 13),
					},
				},
				{
					Code: "array!.slice(-1).pop()", FileName: tsFile,
					Output: []string{"array!.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 8, 1, 13),
					},
				},
			},
		},
		{
			name: "Two-argument slice with adjacent bounds",
			valid: []rule_tester.ValidTestCase{
				{Code: "array.slice(-9.1, -8.1)[0]", FileName: jsFile},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "array.slice(-9, -8)[0]", FileName: jsFile,
					Output: []string{"array.at(-9)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice(-9, -0o10)[0]", FileName: jsFile,
					Output: []string{"array.at(-9)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice(-9, -8).pop()", FileName: jsFile,
					Output: []string{"array.at(-9)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice(-9, -8).shift()", FileName: jsFile,
					Output: []string{"array.at(-9)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice((( -9 )), (( -8 )), ).shift()", FileName: jsFile,
					Output: []string{"array.at((( -9 )))"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "(( array.slice(-9, -8).shift ))()", FileName: jsFile,
					Output: []string{"(( array.at(-9) ))"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 10, 1, 15),
					},
				},
			},
		},
		{
			name: "Other two-argument slices",
			valid: []rule_tester.ValidTestCase{
				{Code: "array.slice(-unknown, -unknown2)[0]", FileName: jsFile},
				{Code: "array.slice(-9.1, unknown)[0]", FileName: jsFile},
				{Code: "array.slice(-9, unknown).pop()", FileName: jsFile},
				{Code: "array.slice(-9, ...unknown)[0]", FileName: jsFile},
				{Code: "array.slice(...[-9], unknown)[0]", FileName: jsFile},
				{Code: "array.slice(-9, unknown)[0]", FileName: jsFile},
				{Code: "array.slice(-0o11, -7)[0]", FileName: jsFile},
				{Code: "array.slice(-9, unknown).shift()", FileName: jsFile},
				{Code: "const KNOWN = -8; array.slice(-9, KNOWN).shift()", FileName: jsFile},
				{Code: "array.slice(-9, 0)[0]", FileName: jsFile},
				{Code: "(( (( array.slice( ((-9)), ((unknown)), ).shift ))() ));", FileName: jsFile},
				{Code: "array.slice(-9, (a, really, _really, complicated, second) => argument)[0]", FileName: jsFile},
			},
		},
		{
			name: "Last-element functions",
			valid: []rule_tester.ValidTestCase{
				{Code: "new _.last(array)", FileName: jsFile},
				{Code: "_.last(array, 2)", FileName: jsFile},
				{Code: "_.last(...array)", FileName: jsFile},
				{Code: "function foo() {return _.last(arguments)}", FileName: jsFile},
				{Code: "function foo() {return lodash.last(arguments)}", FileName: jsFile},
				{Code: "function foo() {return underscore.last(arguments)}", FileName: jsFile},
				{Code: "function foo() {return getLast(arguments)}", FileName: jsFile, Options: map[string]any{"getLastElementFunctions": []any{"getLast"}}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "_.last(array)", FileName: jsFile,
					Output: []string{"array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: "lodash.last(array)", FileName: jsFile,
					Output: []string{"array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 12, "lodash.last"),
					},
				},
				{
					Code: "underscore.last(array)", FileName: jsFile,
					Output: []string{"array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 16, "underscore.last"),
					},
				},
				{
					Code: "_.last(foo.bar)", FileName: jsFile,
					Output: []string{"foo.bar.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: "_.last(new Array)", FileName: jsFile,
					Output: []string{"(new Array).at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: "const foo = []\n_.last([bar])", FileName: jsFile,
					Output: []string{"const foo = []\n;[bar].at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 2, 1, 2, 7, "_.last"),
					},
				},
				{
					Code: "const foo = []\n_.last( new Array )", FileName: jsFile,
					Output: []string{"const foo = []\n;(new Array).at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 2, 1, 2, 7, "_.last"),
					},
				},
				{
					Code: "const foo = []\n_.last( (( new Array )) )", FileName: jsFile,
					Output: []string{"const foo = []\n;(( new Array )).at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 2, 1, 2, 7, "_.last"),
					},
				},
				{
					Code: "if (foo) _.last([bar])", FileName: jsFile,
					Output: []string{"if (foo) [bar].at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 10, 1, 16, "_.last"),
					},
				},
				{
					Code: "_.last(getLast(utils.lastOne(array)))", FileName: jsFile, Options: map[string]any{"getLastElementFunctions": []any{"getLast", "  utils.lastOne  "}},
					Output: []string{"getLast(utils.lastOne(array)).at(-1)", "utils.lastOne(array).at(-1).at(-1)", "array.at(-1).at(-1).at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
						upstreamError("get-last-function", 1, 8, 1, 15, "getLast"),
						upstreamError("get-last-function", 1, 16, 1, 29, "utils.lastOne"),
					},
				},
			},
		},
		{
			name: "checkAllIndexAccess",
			valid: []rule_tester.ValidTestCase{
				{Code: "const object = {value: true}; Object.defineProperty(object, 'value', {get() { return false; }}); array[object.value ? 1 : index];", FileName: jsFile, Options: checkAll},
				{Code: "++array[1]", FileName: jsFile, Options: checkAll},
				{Code: "const offset = 5;const extraArgument = 6;string.charAt(offset + 9, extraArgument)", FileName: jsFile, Options: checkAll},
				{Code: "array[unknown]", FileName: jsFile, Options: checkAll},
				{Code: "array[-1]", FileName: jsFile, Options: checkAll},
				{Code: "array[1.5]", FileName: jsFile, Options: checkAll},
				{Code: "array[1n]", FileName: jsFile, Options: checkAll},
				{Code: "function foo() {return arguments[0]}", FileName: jsFile, Options: checkAll},
				{Code: "const object = {1: 1, a: 2}; object[1]", FileName: jsFile, Options: checkAll},
				{Code: "({1: 1})[1]", FileName: jsFile, Options: checkAll},
				{Code: "const object = {1: 1} as const; object[1]", FileName: tsFile, Options: checkAll},
				{Code: "const object = {1: 1} satisfies Record<number, number>; object[1]", FileName: tsFile, Options: checkAll},
				{Code: "const object = {1: 1}; (object as Record<number, number>)[1]", FileName: tsFile, Options: checkAll},
				{Code: "const object = {1: 1}; object![1]", FileName: tsFile, Options: checkAll},
				{Code: "parent.childNodes[0]", FileName: jsFile, Options: checkAll},
				{Code: "element.children[0]", FileName: jsFile, Options: checkAll},
				{Code: `document.querySelectorAll("li")[0]`, FileName: jsFile, Options: checkAll},
				{Code: `document.getElementsByClassName("item")[0]`, FileName: jsFile, Options: checkAll},
				{Code: `document.getElementsByTagName("li")[0]`, FileName: jsFile, Options: checkAll},
				{Code: `document.getElementsByTagNameNS(namespace, "li")[0]`, FileName: jsFile, Options: checkAll},
				{Code: `document.getElementsByName("item")[0]`, FileName: jsFile, Options: checkAll},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "array[0]", FileName: jsFile, Options: checkAll,
					Output: []string{"array.at(0)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 7, 1, 8),
					},
				},
				{
					Code: "array[1]", FileName: jsFile, Options: checkAll,
					Output: []string{"array.at(1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 7, 1, 8),
					},
				},
				{
					Code: "array[5 + 9]", FileName: jsFile, Options: checkAll,
					Output: []string{"array.at(5 + 9)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 7, 1, 12),
					},
				},
				{
					Code: "const offset = 5;array[offset + 9]", FileName: jsFile, Options: checkAll,
					Output: []string{"const offset = 5;array.at(offset + 9)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 24, 1, 34),
					},
				},
				{
					Code: "let object = {1: 1, a: 2}; object[1]", FileName: jsFile, Options: checkAll,
					Output: []string{"let object = {1: 1, a: 2}; object.at(1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 35, 1, 36),
					},
				},
				{
					Code: `"string"[1]`, FileName: jsFile, Options: checkAll,
					Output: []string{`"string".at(1)`},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 10, 1, 11),
					},
				},
				{
					Code: "`string`[1]", FileName: jsFile, Options: checkAll,
					Output: []string{"`string`.at(1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 10, 1, 11),
					},
				},
				{
					Code: `const string = "string"; string[1]`, FileName: jsFile, Options: checkAll,
					Output: []string{`const string = "string"; string.at(1)`},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 33, 1, 34),
					},
				},
				{
					Code: "new Uint8Array([1])[0]", FileName: jsFile, Options: checkAll,
					Output: []string{"new Uint8Array([1]).at(0)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 21, 1, 22),
					},
				},
				{
					Code: "array[array.length - 1]", FileName: jsFile, Options: checkAll,
					Output: []string{"array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 7, 1, 23),
					},
				},
				{
					Code: "string.charAt(9)", FileName: jsFile, Options: checkAll,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at", 1, 15, 1, 16, "string.at(9)"),
					},
				},
				{
					Code: "string.charAt(5 + 9)", FileName: jsFile, Options: checkAll,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at", 1, 15, 1, 20, "string.at(5 + 9)"),
					},
				},
				{
					Code: "const offset = 5;string.charAt(offset + 9)", FileName: jsFile, Options: checkAll,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at", 1, 32, 1, 42, "const offset = 5;string.at(offset + 9)"),
					},
				},
				{
					Code: "string.charAt(unknown)", FileName: jsFile, Options: checkAll,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at", 1, 15, 1, 22, "string.at(unknown)"),
					},
				},
				{
					Code: "string.charAt(1.5)", FileName: jsFile, Options: checkAll,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at", 1, 15, 1, 18, "string.at(1.5)"),
					},
				},
				{
					Code: "string.charAt(1n)", FileName: jsFile, Options: checkAll,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at", 1, 15, 1, 17, "string.at(1n)"),
					},
				},
				{
					Code: "string.charAt(string.length - 1)", FileName: jsFile, Options: checkAll,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 15, 1, 32, "string.at(- 1)"),
					},
				},
				{
					Code: "foo.charAt(bar.length - 1)", FileName: jsFile, Options: checkAll,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at", 1, 12, 1, 26, "foo.at(bar.length - 1)"),
					},
				},
			},
		},
		{
			name: "Negative charAt and last-element comments",
			valid: []rule_tester.ValidTestCase{
				{Code: "string.charAt(-1)", FileName: jsFile, Options: checkAll},
				{Code: "string.charAt(-1.5)", FileName: jsFile, Options: checkAll},
				{Code: "string.charAt(-0x1)", FileName: jsFile, Options: checkAll},
				{Code: "string.charAt(-0b1)", FileName: jsFile, Options: checkAll},
				{Code: "string.charAt(1 - 2)", FileName: jsFile, Options: checkAll},
				{Code: "const index = -1; string.charAt(index)", FileName: jsFile, Options: checkAll},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "string.charAt(Symbol.iterator)", FileName: jsFile, Options: checkAll,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at", 1, 15, 1, 30, "string.at(Symbol.iterator)"),
					},
				},
				{
					Code: "const a = _.last(/* keep */ foo);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 11, 1, 17, "_.last"),
					},
				},
				{
					Code: "const a = _.last(foo);", FileName: jsFile,
					Output: []string{"const a = foo.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 11, 1, 17, "_.last"),
					},
				},
			},
		},
		{
			name: "Comments in the length access",
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "const a = array[array /* keep */ .length - 1];", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 17, 1, 45),
					},
				},
				{
					Code: "const a = array[/* keep */ array.length - 1];", FileName: jsFile,
					Output: []string{"const a = array.at(/* keep */ -1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 28, 1, 44),
					},
				},
				{
					Code: "const a = array[(/* keep */ array.length) - 1];", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 17, 1, 46),
					},
				},
				{
					Code: "const a = array[array.length - 1];", FileName: jsFile,
					Output: []string{"const a = array.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 17, 1, 33),
					},
				},
				{
					Code: "const a = array[1 /* keep */ + 1];", FileName: jsFile, Options: checkAll,
					Output: []string{"const a = array.at(1 /* keep */ + 1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 17, 1, 33),
					},
				},
			},
		},
		{
			name: "Comments in the slice call",
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "array.slice(-1).pop(/* keep */)", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice(/* keep */ -1).pop()", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice(-1).shift()", FileName: jsFile,
					Output: []string{"array.at(-1)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "array.slice(-1)[0 /* keep */]", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 7, 1, 12),
					},
				},
				{
					Code: "foo(array.slice(-1)[0], /* keep */ bar)", FileName: jsFile,
					Output: []string{"foo(array.at(-1), /* keep */ bar)"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 11, 1, 16),
					},
				},
			},
		},
		{
			name: "Unknown and non-DOM receivers",
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "const nodes = getNodes(); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);", FileName: jsFile,
					Output: []string{"const nodes = getNodes(); nodes.at(-1); nodes.at(-1); nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 33, 1, 49),
						upstreamError("slice", 1, 58, 1, 63),
						upstreamError("get-last-function", 1, 72, 1, 78, "_.last"),
					},
				},
				{
					Code: "_.last(getNodes());", FileName: jsFile,
					Output: []string{"getNodes().at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: `const nodes = Array.from(document.querySelectorAll("li")); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);`, FileName: jsFile,
					Output: []string{`const nodes = Array.from(document.querySelectorAll("li")); nodes.at(-1); nodes.at(-1); nodes.at(-1);`},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 66, 1, 82),
						upstreamError("slice", 1, 91, 1, 96),
						upstreamError("get-last-function", 1, 105, 1, 111, "_.last"),
					},
				},
				{
					Code: "const nodes = [...element.children]; nodes[nodes.length - 1];", FileName: jsFile,
					Output: []string{"const nodes = [...element.children]; nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 44, 1, 60),
					},
				},
				{
					Code: "_.last(Array.from(element.children));", FileName: jsFile,
					Output: []string{"Array.from(element.children).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: "_.last([...element.childNodes]);", FileName: jsFile,
					Output: []string{"[...element.childNodes].at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: "const nodes = [1, 2]; nodes[nodes.length - 1];", FileName: jsFile,
					Output: []string{"const nodes = [1, 2]; nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 29, 1, 45),
					},
				},
				{
					Code: `const string = "abc"; string[string.length - 1];`, FileName: jsFile,
					Output: []string{`const string = "abc"; string.at(-1);`},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 30, 1, 47),
					},
				},
				{
					Code: "const bytes = new Uint8Array([1, 2]); bytes[bytes.length - 1];", FileName: jsFile,
					Output: []string{"const bytes = new Uint8Array([1, 2]); bytes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 45, 1, 61),
					},
				},
				{
					Code: "const nodes = alias; const alias = nodes; _.last(nodes);", FileName: jsFile,
					Output: []string{"const nodes = alias; const alias = nodes; nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 43, 1, 49, "_.last"),
					},
				},
				{
					Code: "let nodes = element.children; nodes = []; _.last(nodes);", FileName: jsFile,
					Output: []string{"let nodes = element.children; nodes = []; nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 43, 1, 49, "_.last"),
					},
				},
				{
					Code: "const {children: nodes} = element; _.last(nodes);", FileName: jsFile,
					Output: []string{"const {children: nodes} = element; nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 36, 1, 42, "_.last"),
					},
				},
				{
					Code: "const nodes = element.children; function foo(nodes) { return _.last(nodes); }", FileName: jsFile,
					Output: []string{"const nodes = element.children; function foo(nodes) { return nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 62, 1, 68, "_.last"),
					},
				},
				{
					Code: `_.last(document["querySelectorAll"]("li"));`, FileName: jsFile,
					Output: []string{`document["querySelectorAll"]("li").at(-1);`},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 1, 1, 7, "_.last"),
					},
				},
				{
					Code: "function foo(value: any) { return value[value.length - 1]; }", FileName: tsFile,
					Output: []string{"function foo(value: any) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 41, 1, 57),
					},
				},
				{
					Code: "function foo(value: any) { return value[value.length - 1]; }", FileName: typedFile,
					Output: []string{"function foo(value: any) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 41, 1, 57),
					},
				},
				{
					Code: "function foo(value: unknown) { return value[value.length - 1]; }", FileName: tsFile,
					Output: []string{"function foo(value: unknown) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 45, 1, 61),
					},
				},
				{
					Code: "function foo(value: unknown) { return value[value.length - 1]; }", FileName: typedFile,
					Output: []string{"function foo(value: unknown) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 45, 1, 61),
					},
				},
				{
					Code: "function foo(value: NodeList | string[]) { return value[value.length - 1]; }", FileName: tsFile,
					Output: []string{"function foo(value: NodeList | string[]) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 57, 1, 73),
					},
				},
				{
					Code: "function foo(value: NodeList | string[]) { return value[value.length - 1]; }", FileName: typedFile,
					Output: []string{"function foo(value: NodeList | string[]) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 57, 1, 73),
					},
				},
				{
					Code: "function foo(value: string[]) { return value[value.length - 1]; }", FileName: tsFile,
					Output: []string{"function foo(value: string[]) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 46, 1, 62),
					},
				},
				{
					Code: "function foo(value: string[]) { return value[value.length - 1]; }", FileName: typedFile,
					Output: []string{"function foo(value: string[]) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 46, 1, 62),
					},
				},
				{
					Code: "function foo(value: readonly number[]) { return value[value.length - 1]; }", FileName: tsFile,
					Output: []string{"function foo(value: readonly number[]) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 55, 1, 71),
					},
				},
				{
					Code: "function foo(value: readonly number[]) { return value[value.length - 1]; }", FileName: typedFile,
					Output: []string{"function foo(value: readonly number[]) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 55, 1, 71),
					},
				},
				{
					Code: "function foo(value: string) { return value[value.length - 1]; }", FileName: tsFile,
					Output: []string{"function foo(value: string) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 44, 1, 60),
					},
				},
				{
					Code: "function foo(value: string) { return value[value.length - 1]; }", FileName: typedFile,
					Output: []string{"function foo(value: string) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 44, 1, 60),
					},
				},
				{
					Code: "function foo(value: Uint8Array) { return value[value.length - 1]; }", FileName: tsFile,
					Output: []string{"function foo(value: Uint8Array) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 48, 1, 64),
					},
				},
				{
					Code: "function foo(value: Uint8Array) { return value[value.length - 1]; }", FileName: typedFile,
					Output: []string{"function foo(value: Uint8Array) { return value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 48, 1, 64),
					},
				},
				{
					Code: "const nodes = Array.from(element.children); nodes[0];", FileName: jsFile, Options: checkAll,
					Output: []string{"const nodes = Array.from(element.children); nodes.at(0);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 51, 1, 52),
					},
				},
				{
					Code: "function foo(value: any) { return value[0]; }", FileName: typedFile, Options: checkAll,
					Output: []string{"function foo(value: any) { return value.at(0); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 41, 1, 42),
					},
				},
				{
					Code: `const nodes = Array.from(document.querySelectorAll("li")); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);`, FileName: typedFile,
					Output: []string{`const nodes = Array.from(document.querySelectorAll("li")); nodes.at(-1); nodes.at(-1); nodes.at(-1);`},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 66, 1, 82),
						upstreamError("slice", 1, 91, 1, 96),
						upstreamError("get-last-function", 1, 105, 1, 111, "_.last"),
					},
				},
				{
					Code: "const nodes = condition ? element.children : []; nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);", FileName: jsFile,
					Output: []string{"const nodes = condition ? element.children : []; nodes.at(-1); nodes.at(-1); nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 56, 1, 72),
						upstreamError("slice", 1, 81, 1, 86),
						upstreamError("get-last-function", 1, 95, 1, 101, "_.last"),
					},
				},
				{
					Code: "const nodes = element.children; _.last(nodes as string[]);", FileName: tsFile,
					Output: []string{"const nodes = element.children; (nodes as string[]).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 33, 1, 39, "_.last"),
					},
				},
				{
					Code: "function foo(value: unknown) { if (value instanceof NodeList) { return value[value.length - 1]; } }", FileName: tsFile,
					Output: []string{"function foo(value: unknown) { if (value instanceof NodeList) { return value.at(-1); } }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 78, 1, 94),
					},
				},
				{
					Code: "function foo(value: any) { value.slice(-1)[0]; _.last(value); }", FileName: tsFile,
					Output: []string{"function foo(value: any) { value.at(-1); value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 34, 1, 39),
						upstreamError("get-last-function", 1, 48, 1, 54, "_.last"),
					},
				},
				{
					Code: "function foo(value: any) { value.slice(-1)[0]; _.last(value); }", FileName: typedFile,
					Output: []string{"function foo(value: any) { value.at(-1); value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 34, 1, 39),
						upstreamError("get-last-function", 1, 48, 1, 54, "_.last"),
					},
				},
				{
					Code: "function foo(value: unknown) { value.slice(-1)[0]; _.last(value); }", FileName: tsFile,
					Output: []string{"function foo(value: unknown) { value.at(-1); value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 38, 1, 43),
						upstreamError("get-last-function", 1, 52, 1, 58, "_.last"),
					},
				},
				{
					Code: "function foo(value: unknown) { value.slice(-1)[0]; _.last(value); }", FileName: typedFile,
					Output: []string{"function foo(value: unknown) { value.at(-1); value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 38, 1, 43),
						upstreamError("get-last-function", 1, 52, 1, 58, "_.last"),
					},
				},
				{
					Code: "function foo(value: NodeList | string[]) { value.slice(-1)[0]; _.last(value); }", FileName: tsFile,
					Output: []string{"function foo(value: NodeList | string[]) { value.at(-1); value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 50, 1, 55),
						upstreamError("get-last-function", 1, 64, 1, 70, "_.last"),
					},
				},
				{
					Code: "function foo(value: NodeList | string[]) { value.slice(-1)[0]; _.last(value); }", FileName: typedFile,
					Output: []string{"function foo(value: NodeList | string[]) { value.at(-1); value.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 50, 1, 55),
						upstreamError("get-last-function", 1, 64, 1, 70, "_.last"),
					},
				},
				{
					Code: "function foo<T extends NodeList | string[]>(nodes: T) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: tsFile,
					Output: []string{"function foo<T extends NodeList | string[]>(nodes: T) { nodes.at(-1); nodes.at(-1); nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 63, 1, 79),
						upstreamError("slice", 1, 88, 1, 93),
						upstreamError("get-last-function", 1, 102, 1, 108, "_.last"),
					},
				},
				{
					Code: "function foo<T extends NodeList | string[]>(nodes: T) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: typedFile,
					Output: []string{"function foo<T extends NodeList | string[]>(nodes: T) { nodes.at(-1); nodes.at(-1); nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 63, 1, 79),
						upstreamError("slice", 1, 88, 1, 93),
						upstreamError("get-last-function", 1, 102, 1, 108, "_.last"),
					},
				},
				{
					Code: "function foo(nodes: NodeList | string[] | undefined) { nodes?.[nodes.length - 1]; nodes?.slice(-1)[0]; _.last(nodes); }", FileName: tsFile,
					Output: []string{"function foo(nodes: NodeList | string[] | undefined) { nodes?.at(-1); nodes?.at(-1); nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 64, 1, 80),
						upstreamError("slice", 1, 90, 1, 95),
						upstreamError("get-last-function", 1, 104, 1, 110, "_.last"),
					},
				},
				{
					Code: "function foo(nodes: NodeList | string[] | undefined) { nodes?.[nodes.length - 1]; nodes?.slice(-1)[0]; _.last(nodes); }", FileName: typedFile,
					Output: []string{"function foo(nodes: NodeList | string[] | undefined) { nodes?.at(-1); nodes?.at(-1); nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 64, 1, 80),
						upstreamError("slice", 1, 90, 1, 95),
						upstreamError("get-last-function", 1, 104, 1, 110, "_.last"),
					},
				},
				{
					Code: "function foo(nodes: NodeList | string[]) { if (nodes instanceof NodeList) { return nodes[nodes.length - 1]; } return nodes[nodes.length - 1]; }", FileName: typedFile,
					Output: []string{"function foo(nodes: NodeList | string[]) { if (nodes instanceof NodeList) { return nodes[nodes.length - 1]; } return nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 124, 1, 140),
					},
				},
				{
					Code: "declare function getNodes(): any; const nodes = getNodes?.(); _.last(nodes);", FileName: tsFile,
					Output: []string{"declare function getNodes(): any; const nodes = getNodes?.(); nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 63, 1, 69, "_.last"),
					},
				},
				{
					Code: "declare function getNodes(): unknown; const nodes = getNodes?.(); _.last(nodes);", FileName: tsFile,
					Output: []string{"declare function getNodes(): unknown; const nodes = getNodes?.(); nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 67, 1, 73, "_.last"),
					},
				},
				{
					Code: "declare function getNodes(): NodeList | string[]; const nodes = getNodes?.(); _.last(nodes);", FileName: tsFile,
					Output: []string{"declare function getNodes(): NodeList | string[]; const nodes = getNodes?.(); nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 79, 1, 85, "_.last"),
					},
				},
				{
					Code: "declare function getNodes(): string[]; const nodes = getNodes?.(); _.last(nodes);", FileName: tsFile,
					Output: []string{"declare function getNodes(): string[]; const nodes = getNodes?.(); nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 68, 1, 74, "_.last"),
					},
				},
				{
					Code: "declare function getNodes(): NodeList; const nodes = getNodes?.() as string[]; _.last(nodes);", FileName: tsFile,
					Output: []string{"declare function getNodes(): NodeList; const nodes = getNodes?.() as string[]; nodes.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 80, 1, 86, "_.last"),
					},
				},
				{
					Code: "function foo(slot: HTMLSlotElement) { const nodes = slot.assignedElements(); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: typedFile,
					Output: []string{"function foo(slot: HTMLSlotElement) { const nodes = slot.assignedElements(); nodes.at(-1); nodes.at(-1); nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 84, 1, 100),
						upstreamError("slice", 1, 109, 1, 114),
						upstreamError("get-last-function", 1, 123, 1, 129, "_.last"),
					},
				},
				{
					Code: "function foo(form: HTMLFormElement) { const nodes = Array.from(form.elements); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }", FileName: typedFile,
					Output: []string{"function foo(form: HTMLFormElement) { const nodes = Array.from(form.elements); nodes.at(-1); nodes.at(-1); nodes.at(-1); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 86, 1, 102),
						upstreamError("slice", 1, 111, 1, 116),
						upstreamError("get-last-function", 1, 125, 1, 131, "_.last"),
					},
				},
			},
		},
		{
			name: "Upstream runtime edit checks",
			valid: []rule_tester.ValidTestCase{
				{Code: "const array = [function () { return this === array; }]; array[array.length - 1]();", FileName: jsFile},
				{Code: "const array = [function () { return this.length; }]; array.slice(-1)[0]();", FileName: jsFile},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "function read() {return(_).last(getItems(/* keep */));} read();", FileName: jsFile,
					Output: []string{"function read() {return getItems(/* keep */).at(-1);} read();"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 24, 1, 32, "_.last"),
					},
				},
				{
					Code: "const array = [class { value = 1; }]; new array[array.length - 1]().value;", FileName: jsFile,
					Output: []string{"const array = [class { value = 1; }]; new (array.at(-1))().value;"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 49, 1, 65),
					},
				},
				{
					Code: "const array = [{Constructor: class { value = 1; }}]; new array[array.length - 1].Constructor().value;", FileName: jsFile,
					Output: []string{"const array = [{Constructor: class { value = 1; }}]; new (array.at(-1)).Constructor().value;"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 64, 1, 80),
					},
				},
				{
					Code: "const array = [{tag() { return class { value = 1; }; }}]; new array[array.length - 1].tag`template`().value;", FileName: jsFile,
					Output: []string{"const array = [{tag() { return class { value = 1; }; }}]; new (array.at(-1)).tag`template`().value;"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 69, 1, 85),
					},
				},
			},
		},
		{
			name: "Documentation examples",
			valid: []rule_tester.ValidTestCase{
				{Code: "const foo = array.at(-1);", FileName: jsFile},
				{Code: "const foo = array.at(-5);", FileName: jsFile},
				{Code: "const foo = string.at(-5);", FileName: jsFile},
				{Code: "const foo = string.at(index);", FileName: jsFile},
				{Code: "const foo = array[100];", FileName: jsFile},
				{Code: "const foo = array.at(array.length - 1);", FileName: jsFile},
				{Code: "array[array.length - 1] = foo;", FileName: jsFile},
				{Code: "function foo() {\n\treturn arguments[arguments.length - 1];\n}", FileName: jsFile},
				{Code: "const foo = element.children[element.children.length - 1];", FileName: jsFile},
				{Code: "const elements = document.querySelectorAll('div');\nconst lastElement = elements[elements.length - 1];", FileName: jsFile},
				{Code: "const lastElement = Array.from(elements).at(-1);", FileName: jsFile},
				{Code: "const foo = bar[unknownProperty];", FileName: jsFile, Options: checkAll},
				{Code: "const foo = string.charAt(-1);", FileName: jsFile, Options: checkAll},
				{Code: "const foo = bar.at(-1);", FileName: jsFile, Options: map[string]any{"getLastElementFunctions": []any{"utils.lastElement"}}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{
					Code: "const foo = array[array.length - 1];", FileName: jsFile,
					Output: []string{"const foo = array.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 19, 1, 35),
					},
				},
				{
					Code: "const foo = array.slice(-1)[0];", FileName: jsFile,
					Output: []string{"const foo = array.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 19, 1, 24),
					},
				},
				{
					Code: "const foo = array.slice(-1).pop();", FileName: jsFile,
					Output: []string{"const foo = array.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 19, 1, 24),
					},
				},
				{
					Code: "const foo = array.slice(-1).shift();", FileName: jsFile,
					Output: []string{"const foo = array.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 19, 1, 24),
					},
				},
				{
					Code: "const foo = lodash.last(array);", FileName: jsFile,
					Output: []string{"const foo = array.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 13, 1, 24, "lodash.last"),
					},
				},
				{
					Code: "const foo = array[array.length - 5];", FileName: jsFile,
					Output: []string{"const foo = array.at(-5);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("negative-index", 1, 19, 1, 35),
					},
				},
				{
					Code: "const foo = string.charAt(string.length - 5);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at-negative", 1, 27, 1, 44, "const foo = string.at(- 5);"),
					},
				},
				{
					Code: "const foo = string.substring(index, index + 1);", FileName: jsFile,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-substring", 1, 13, 1, 47, "const foo = string.at(index);"),
					},
				},
				{
					Code: "const lastElement = Array.from(elements).slice(-1)[0];", FileName: jsFile,
					Output: []string{"const lastElement = Array.from(elements).at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("slice", 1, 42, 1, 47),
					},
				},
				{
					Code: "const foo = bar[10];", FileName: jsFile, Options: checkAll,
					Output: []string{"const foo = bar.at(10);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("index", 1, 17, 1, 19),
					},
				},
				{
					Code: "const foo = string.charAt(unknownIndex);", FileName: jsFile, Options: checkAll,
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("string-char-at", 1, 27, 1, 39, "const foo = string.at(unknownIndex);"),
					},
				},
				{
					Code: "const foo = utils.lastElement(bar);", FileName: jsFile, Options: map[string]any{"getLastElementFunctions": []any{"utils.lastElement"}},
					Output: []string{"const foo = bar.at(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						upstreamError("get-last-function", 1, 13, 1, 30, "utils.lastElement"),
					},
				},
			},
		},
	}
	for _, suite := range suites {
		t.Run(suite.name, func(t *testing.T) {
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testRule, suite.valid, suite.invalid)
		})
	}
}

func upstreamError(id string, line, column, endLine, endColumn int, detail ...string) rule_tester.InvalidTestCaseError {
	message := upstreamMessages[id]
	if id == "get-last-function" {
		message = "Prefer `.at(-1)` over `" + detail[0] + "(…)` to get the last element."
		detail = detail[1:]
	}
	diagnostic := rule_tester.InvalidTestCaseError{
		MessageId: id, Message: message, Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
	if len(detail) != 0 {
		diagnostic.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{MessageId: "use-at", Output: detail[0]}}
	}
	return diagnostic
}

var upstreamMessages = map[string]string{
	"negative-index":          "Prefer `.at(…)` over `[….length - index]`.",
	"index":                   "Prefer `.at(…)` over index access.",
	"string-char-at-negative": "Prefer `String#at(…)` over `String#charAt(….length - index)`.",
	"string-substring":        "Prefer `String#at(…)` over `String#substring(…)` when getting one character.",
	"slice":                   "Prefer `.at(…)` over the first element from `.slice(…)`.",
	"string-char-at":          "Prefer `String#at(…)` over `String#charAt(…)`.",
}
