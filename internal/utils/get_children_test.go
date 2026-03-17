package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
)

// parseJSFile parses JavaScript code with allowJs enabled to trigger JSDoc reparse.
func parseJSFile(t *testing.T, code string) *ast.SourceFile {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "test.js")
	fs := NewOverlayVFSForFile(filePath, code)
	host := CreateCompilerHost(rootDir, fs)
	program, err := CreateProgramFromOptions(true, &core.CompilerOptions{
		AllowJs:  core.TSTrue,
		CheckJs:  core.TSTrue,
		Strict:   core.TSTrue,
		Target:   core.ScriptTargetESNext,
		NoEmit:   core.TSTrue,
	}, []string{filePath}, host)
	if err != nil {
		t.Fatalf("couldn't create program: %v", err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatal("source file not found")
	}
	return sourceFile
}

// assertForEachCommentNoPanic verifies that ForEachComment completes without panicking.
func assertForEachCommentNoPanic(t *testing.T, sourceFile *ast.SourceFile) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ForEachComment panicked: %v", r)
		}
	}()
	var comments []*ast.CommentRange
	ForEachComment(&sourceFile.Node, func(comment *ast.CommentRange) {
		comments = append(comments, comment)
	}, sourceFile)
}

// TestGetChildren_SkipsReparsedNodes tests that GetChildren skips reparsed
// nodes created from JSDoc annotations, preventing token cache parent mismatches.
// This is the fix for https://github.com/web-infra-dev/rslint/issues/364
func TestGetChildren_SkipsReparsedNodes(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "JSDoc @type with array literal",
			code: `/** @type {number[]} */
const x = [1, 2, 3];`,
		},
		{
			name: "JSDoc @type with object literal",
			code: `/** @type {{ name: string, value: number }} */
const config = { name: "test", value: 42 };`,
		},
		{
			name: "JSDoc @type on var declaration",
			code: `/** @type {number[]} */
var z = [7, 8, 9];`,
		},
		{
			name: "JSDoc @satisfies with array literal",
			code: `/** @satisfies {number[]} */
const y = [4, 5, 6];`,
		},
		{
			name: "JSDoc @satisfies with object literal",
			code: `/** @satisfies {Record<string, number>} */
const scores = { math: 95, english: 88 };`,
		},
		{
			name: "JSDoc @type on return statement",
			code: `function foo() {
  /** @type {number[]} */
  return [1, 2, 3];
}`,
		},
		{
			name: "JSDoc @param with types",
			code: `/**
 * @param {number} x
 * @param {string} y
 */
function test(x, y) {
  return x + y;
}`,
		},
		{
			name: "JSDoc @param with array type",
			code: `/**
 * @param {number[]} items
 * @returns {number[]}
 */
function processItems(items) {
  return items;
}`,
		},
		{
			name: "JSDoc @param on arrow function expression",
			code: `/**
 * @param {number} x
 */
const square = (x) => x * x;`,
		},
		{
			name: "JSDoc @typedef",
			code: `/**
 * @typedef {Object} Config
 * @property {string} name
 * @property {number} value
 */

/** @type {Config} */
const config = { name: "test", value: 42 };`,
		},
		{
			name: "JSDoc @callback",
			code: `/**
 * @callback Transformer
 * @param {number} x
 * @returns {number}
 */

/** @type {Transformer} */
const double = (x) => x * 2;`,
		},
		{
			name: "JSDoc @overload",
			code: `/**
 * @overload
 * @param {string} x
 * @returns {string}
 */
/**
 * @overload
 * @param {number} x
 * @returns {number}
 */
/**
 * @param {string | number} x
 * @returns {string | number}
 */
function identity(x) {
  return x;
}`,
		},
		{
			name: "JSDoc @template",
			code: `/**
 * @template T
 * @param {T[]} arr
 * @param {(a: T, b: T) => number} compareFn
 * @returns {T[]}
 */
function sortBy(arr, compareFn) {
  return [...arr].sort(compareFn);
}`,
		},
		{
			name: "JSDoc @type with nested array",
			code: `/** @type {number[][]} */
const matrix = [[1, 2], [3, 4]];`,
		},
		{
			name: "JSDoc @type with function type",
			code: `/** @type {(x: number) => number} */
const double = (x) => x * 2;`,
		},
		{
			name: "JSDoc @type with union type",
			code: `/** @type {string | number[]} */
var mixed = [1, 2, 3];`,
		},
		{
			name: "JSDoc @type with Promise",
			code: `/** @type {Promise<number[]>} */
const result = Promise.resolve([1, 2, 3]);`,
		},
		{
			name: "JSDoc on destructured parameter",
			code: `/**
 * @param {{ name: string, items: number[] }} options
 */
function withOptions({ name, items }) {
  return { name, total: items.length };
}`,
		},
		{
			name: "multiple JSDoc annotations in one file",
			code: `/** @type {number[]} */
const a = [1, 2];

/** @type {string[]} */
const b = ["x", "y"];

/**
 * @param {number} x
 * @returns {number}
 */
function inc(x) { return x + 1; }

/** @satisfies {Record<string, number>} */
const c = { foo: 1 };`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := parseJSFile(t, tt.code)
			assertForEachCommentNoPanic(t, sourceFile)
		})
	}
}
