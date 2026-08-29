package no_array_concat_in_loop_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_concat_in_loop"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoArrayConcatInLoopScopeParity exercises the name-only variable lookup
// used by upstream. In particular, TypeScript type bindings stop lookup even
// for value expressions, while JSDoc-generated syntax declares nothing.
func TestNoArrayConcatInLoopScopeParity(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_array_concat_in_loop.NoArrayConcatInLoopRule,
		[]rule_tester.ValidTestCase{
			// Authored type bindings shadow the outer value in scope-manager's
			// name-only lookup, even though TypeScript resolves through them.
			tsValid(`let result = [];
{
	interface result {}
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			tsValid(`let result = [];
{
	type result = number;
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			tsValid(`let result = [];
function append<result>() {
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			tsValid(`interface result {}
let result = [];
for (const chunk of chunks) result = result.concat(chunk);`),
			tsValid(`import type {result} from "./types";
let result = [];
for (const chunk of chunks) result = result.concat(chunk);`),
			tsValid(`function append<result>() {
	let result = [];
	for (const chunk of chunks) result = result.concat(chunk);
}`),

			// Class and function-type scopes are visible in positions where the
			// TypeScript resolver intentionally hides their type parameters.
			tsValid(`let result = [];
class Box<result> {
	static { for (const chunk of chunks) result = result.concat(chunk); }
}`),
			tsValid(`let result = [];
for (const chunk of chunks) {
	class Box<result> { [result = result.concat(chunk)]() {} }
}`),
			tsValid(`let result = [];
for (const chunk of chunks) {
	type Shape = { [result = result.concat(chunk)]<result>(): void };
}`),
			tsValid(`let result = [];
for (const chunk of chunks) {
	class Box {
		append<Value extends { [result = result.concat(chunk)]: unknown }>() {}
	}
}`),
			tsValid(`let result = [];
for (const chunk of chunks) {
	class Box {
		append(): { [result = result.concat(chunk)]: unknown } { return {}; }
	}
}`),
			tsValid(`let result = [];
for (const chunk of chunks) {
	class Box {
		append(@((result = result.concat(chunk))) value: unknown) {}
	}
}`),

			// A function scope contains its body declarations while a named
			// function expression's self binding remains one scope farther out.
			tsValid(`let result = [];
const append = function result() {
	for (const chunk of chunks) result = result.concat(chunk);
};`),
			tsValid(`let result = [];
function append(value = class {
	static { for (const chunk of chunks) result = result.concat(chunk); }
}) {
	let result = [value];
}`),

			// Non-arrow runtime functions provide an implicit, definition-less
			// arguments variable. An authored declaration is tested separately.
			tsValid(`let arguments = [];
function append() {
	for (const chunk of chunks) arguments = arguments.concat(chunk);
}`),

			// Enum members bind only within their own enum declaration.
			tsValid(`let result = [];
for (const chunk of chunks) {
	enum State { result = 0, next = (result = result.concat(chunk), 1) }
}`),
			tsValid(`let result = [];
for (const chunk of chunks) {
	enum State { "result" = 0, next = (result = result.concat(chunk), 1) }
}`),

			// Reopened namespaces have distinct lexical scopes even though the
			// TypeScript binder merges their export symbol tables.
			tsValid(`let result = [0];
namespace Store { export let result = []; }
namespace Store {
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			tsValid(`let result = [];
namespace Store {
	export let result = [0];
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			tsValid(`let result = [];
namespace Store {
	export import result = Other;
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			tsValid(`let result = [];
for (const chunk of chunks) {
	switch (result = result.concat(chunk)) {
		case 0: let result = [chunk];
	}
}`),

			// A function under a scope-less statement binds in the nearest
			// existing scope. A braced counterpart appears in the invalid list.
			tsValid(`function append() {
	let result = [];
	if (condition) function result() {}
	for (const chunk of chunks) result = result.concat(chunk);
}`),

			// Direct class decorators acquire the class scope. A nested scope in
			// the decorator deliberately drops it and appears below as invalid.
			tsValid(`let result = [];
for (const chunk of chunks) {
	@((result = result.concat(chunk))) class Box<result> {}
}`),
			tsValid(`let result = [];
for (const chunk of chunks) {
	class Box<result> { @((result = result.concat(chunk))) method() {} }
}`),
			tsValid(`let result = [];
for (const chunk of chunks) {
	const Box = @((result = result.concat(chunk))) class result {};
}`),
		},
		[]rule_tester.InvalidTestCase{
			// Parsing JSDoc may add binder nodes, but comments do not
			// define variables in ESLint. Synthetic-only scopes must be skipped.
			invalidConcat(`/** @typedef {number} result */
let result = [];
for (const chunk of chunks) result = result.concat(chunk);`),
			invalidConcat(`/** @callback result @returns {number} */
let result = [];
for (const chunk of chunks) result = result.concat(chunk);`),
			invalidConcat(`/** @import {Value as result} from './types.js' */
let result = [];
for (const chunk of chunks) result = result.concat(chunk);`),
			invalidConcat(`let result = [];
/** @template result */
function append() {
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			invalidConcat(`/** @template result */
function append() {
	let result = [];
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			invalidConcat(`/** @typedef {number} result.value */
let result = [];
for (const chunk of chunks) result = result.concat(chunk);`),
			invalidConcat(`/** @type {unknown[]} */
let result = [];
for (const chunk of chunks) result = result.concat(chunk);`),
			invalidConcat(`let result = [];
function append() {
	/** @typedef {number} result */
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			invalidConcat(`let result = [];
/** @template result */
class Box {
	static { for (const chunk of chunks) result = result.concat(chunk); }
}`),

			// A closer authored variable wins over an outer type parameter.
			invalidConcatTS(`function append<result>() {
	{
		let result = [];
		for (const chunk of chunks) result = result.concat(chunk);
	}
}`),
			invalidConcatTS(`class Box<result> {
	static {
		let result = [];
		for (const chunk of chunks) result = result.concat(chunk);
	}
}`),

			// Parameter initializers acquire the function scope, including body
			// declarations that TypeScript's runtime resolver walks past.
			invalidConcatTS(`let result = [0];
function append(value = class {
	static { for (const chunk of chunks) result = result.concat(chunk); }
}) {
	let result = [];
}`),
			invalidConcatTS(`const append = function result(value = class {
	static { for (const chunk of chunks) result = result.concat(chunk); }
}) {
	let result = [];
};`),

			// Arrows inherit arguments, while an authored function-scope
			// declaration takes precedence over the implicit variable.
			invalidConcatTS(`let arguments = [];
const append = () => {
	for (const chunk of chunks) arguments = arguments.concat(chunk);
};`),
			invalidConcatTS(`function append() {
	let arguments = [];
	for (const chunk of chunks) arguments = arguments.concat(chunk);
}`),

			// A braced function stays in its block. Nested var declarations are
			// collected exactly once in the surrounding variable scope.
			invalidConcatTS(`function append() {
	let result = [];
	if (condition) { function result() {} }
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			invalidConcatTS(`function append() {
	if (condition) { var result = []; }
	for (const chunk of chunks) result = result.concat(chunk);
}`),

			// Neither segment of a dotted namespace declaration binds a lexical
			// variable, and declarations in other reopened bodies do not leak.
			invalidConcatTS(`let result = [];
namespace result.inner {
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			invalidConcatTS(`let result = [];
namespace Store { export let result = [0]; }
namespace Store {
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			invalidConcatTS(`let result = [];
namespace Store { export import result = Other; }
namespace Store {
	for (const chunk of chunks) result = result.concat(chunk);
}`),
			invalidConcatTS(`let result = [];
namespace Store {
	export {result};
	for (const chunk of chunks) result = result.concat(chunk);
}`),

			// A member of another enum declaration does not enter this enum's
			// lexical scope.
			invalidConcatTS(`let result = [];
for (const chunk of chunks) {
	enum State { next = (result = result.concat(chunk), 1) }
	enum State { result = 0 }
}`),

			// Runtime method type parameters do not cover the computed key, but
			// a TS method-signature type parameter does cover its computed key.
			invalidConcatTS(`let result = [];
for (const chunk of chunks) {
	class Box { [result = result.concat(chunk)]<result>() {} }
}`),
			invalidConcatTS(`let result = [];
for (const chunk of chunks) {
	type Shape = { [result = result.concat(chunk)]<other>(): void };
}`),

			// A nested decorator scope is parented outside the decorated class.
			invalidConcatTS(`let result = [];
@((() => {
	for (const chunk of chunks) result = result.concat(chunk);
})())
class Box<result> {}`),

			// Bodyless TypeScript declarations are not any of unicorn's runtime
			// function node kinds, so they do not hide an enclosing loop.
			invalidConcatTS(`let result = [];
for (const chunk of chunks) {
	function append(value = (result = result.concat(chunk))): void;
}`),
			invalidConcatTS(`let result = [];
for (const chunk of chunks) {
	abstract class Box {
		abstract append(value = (result = result.concat(chunk))): void;
	}
}`),
		},
	)
}
