package no_thenable_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_thenable"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageIDObject = "no-thenable-object"
	messageIDExport = "no-thenable-export"
	messageIDClass  = "no-thenable-class"

	messageObject = "Do not add `then` to an object."
	messageExport = "Do not export `then`."
	messageClass  = "Do not add `then` to a class."
)

// TestNoThenableUpstream migrates the full valid/invalid suite from upstream
// test/no-thenable.js 1:1. Position assertions cover line/column for every
// invalid case. rslint-specific lock-in cases live in the
// no_thenable_extras_test.go file.
func TestNoThenableUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_thenable.NoThenableRule,
		[]rule_tester.ValidTestCase{
			// ---- `object` ----
			jsValid(`const then = {}`),
			jsValid(`const notThen = then`),
			jsValid(`const then = then.then`),
			jsValid(`const foo = {notThen: 1}`),
			jsValid(`const foo = {notThen() {}}`),
			jsValid(`const foo = {[then]: 1}`),
			jsValid(`const NOT_THEN = "no-then";const foo = {[NOT_THEN]: 1}`),
			jsValid(`function foo({then}) {}`),
			jsValid(`({[Symbol.prototype]: 1})`),

			// ---- `class` ----
			jsValid(`class then {}`),
			jsValid(`class Foo {notThen}`),
			jsValid(`class Foo {notThen() {}}`),
			jsValid(`class Foo {[then]}`),
			jsValid(`class Foo {#then}`),
			jsValid(`class Foo {#then() {}}`),
			jsValid(`class Foo {[then]() {}}`),
			jsValid(`class Foo {get notThen() {}}`),
			jsValid(`class Foo {get #then() {}}`),
			jsValid(`class Foo {get [then]() {}}`),
			jsValid(`class Foo {static notThen}`),
			jsValid(`class Foo {static notThen() {}}`),
			jsValid(`class Foo {static #then}`),
			jsValid(`class Foo {static #then() {}}`),
			jsValid(`class Foo {static [then]}`),
			jsValid(`class Foo {static [then]() {}}`),
			jsValid(`class Foo {static get notThen() {}}`),
			jsValid(`class Foo {static get #then() {}}`),
			jsValid(`class Foo {static get [then]() {}}`),
			jsValid(`class Foo {notThen = then}`),
			jsValid(`class Foo {[Symbol.property]}`),
			jsValid(`class Foo {static [Symbol.property]}`),
			jsValid(`class Foo {get [Symbol.property]() {}}`),
			jsValid(`class Foo {[Symbol.property]() {}}`),
			jsValid(`class Foo {static get [Symbol.property]() {}}`),

			// ---- Assign ----
			jsValid(`foo[then] = 1`),
			jsValid(`foo.notThen = 1`),
			jsValid(`then.notThen = then.then`),
			jsValid(`const NOT_THEN = "no-then";foo[NOT_THEN] = 1`),
			jsValid(`foo.then ++`),
			jsValid(`++ foo.then`),
			jsValid(`delete foo.then`),
			jsValid(`typeof foo.then`),
			jsValid(`foo.then != 1`),
			jsValid(`foo[Symbol.property] = 1`),

			// ---- `Object.fromEntries` ----
			jsValid(`Object.fromEntries([then, 1])`),
			jsValid(`Object.fromEntries([,,])`),
			jsValid(`Object.fromEntries([[,,],[]])`),
			jsValid(`const NOT_THEN = "not-then";Object.fromEntries([[NOT_THEN, 1]])`),
			jsValid(`Object.fromEntries([[["then", 1]]])`),
			jsValid(`NotObject.fromEntries([["then", 1]])`),
			jsValid(`Object.notFromEntries([["then", 1]])`),
			jsValid(`Object.fromEntries?.([["then", 1]])`),
			jsValid(`Object?.fromEntries([["then", 1]])`),
			jsValid(`Object.fromEntries([[..."then", 1]])`),
			jsValid(`Object.fromEntries([["then", 1]], extraArgument)`),
			jsValid(`Object.fromEntries(...[["then", 1]])`),
			jsValid(`Object.fromEntries([[Symbol.property, 1]])`),

			// ---- `{Object,Reflect}.defineProperty` ----
			jsValid(`Object.defineProperty(foo, then, 1)`),
			jsValid(`Object.defineProperty(foo, "not-then", 1)`),
			jsValid(`const then = "no-then";Object.defineProperty(foo, then, 1)`),
			jsValid(`Reflect.defineProperty(foo, then, 1)`),
			jsValid(`Reflect.defineProperty(foo, "not-then", 1)`),
			jsValid(`const then = "no-then";Reflect.defineProperty(foo, then, 1)`),
			jsValid(`Object.defineProperty(foo, "then", )`),
			jsValid(`Object.defineProperty(...foo, "then", 1)`),
			jsValid(`Object.defineProperty(foo, ...["then", 1])`),
			jsValid(`Object.defineProperty(foo, Symbol.property, 1)`),
			jsValid(`Reflect.defineProperty(foo, Symbol.property, 1)`),

			// ---- `export` ----
			jsValid(`export {default} from "then"`),
			jsValid(`const then = 1; export {then as notThen}`),
			jsValid(`export default then`),
			jsValid(`export function notThen(){}`),
			jsValid(`export class notThen {}`),
			jsValid(`export default function then (){}`),
			jsValid(`export default class then {}`),
			jsValid(`export default function (){}`),
			jsValid(`export default class {}`),

			// ---- `export variables` ----
			jsValid(`export const notThen = 1`),
			jsValid(`export const {then: notThen} = 1`),
			jsValid(`export const {then: notThen = then} = 1`),
		},
		[]rule_tester.InvalidTestCase{
			// ---- `object` ----
			objectInvalid(`const foo = {then: 1}`, `then`),
			objectInvalid(`const foo = {["then"]: 1}`, `"then"`),
			objectInvalid("const foo = {[`then`]: 1}", "`then`"),
			objectInvalid(`const THEN = "then";const foo = {[THEN]: 1}`, `THEN`, 2),
			objectInvalid(`const foo = {then() {}}`, `then`),
			objectInvalid(`const foo = {["then"]() {}}`, `"then"`),
			objectInvalid("const foo = {[`then`]() {}}", "`then`"),
			objectInvalid(`const THEN = "then";const foo = {[THEN]() {}}`, `THEN`, 2),
			objectInvalid(`const foo = {get then() {}}`, `then`),
			objectInvalid(`const foo = {set then(v) {}}`, `then`),
			objectInvalid(`const foo = {get ["then"]() {}}`, `"then"`),
			objectInvalid("const foo = {get [`then`]() {}}", "`then`"),
			objectInvalid(`const THEN = "then";const foo = {get [THEN]() {}}`, `THEN`, 2),

			// ---- `class` ----
			classInvalid(`class Foo {then}`, `then`),
			classInvalid(`const Foo = class {then}`, `then`),
			classInvalid(`class Foo {["then"]}`, `"then"`),
			classInvalid("class Foo {[`then`]}", "`then`"),
			classInvalid(`const THEN = "then";class Foo {[THEN]}`, `THEN`, 2),
			classInvalid(`class Foo {then() {}}`, `then`),
			classInvalid(`class Foo {["then"]() {}}`, `"then"`),
			classInvalid("class Foo {[`then`]() {}}", "`then`"),
			classInvalid(`const THEN = "then";class Foo {[THEN]() {}}`, `THEN`, 2),
			classInvalid(`class Foo {static then}`, `then`),
			classInvalid(`class Foo {static ["then"]}`, `"then"`),
			classInvalid("class Foo {static [`then`]}", "`then`"),
			classInvalid(`const THEN = "then";class Foo {static [THEN]}`, `THEN`, 2),
			classInvalid(`class Foo {static then() {}}`, `then`),
			classInvalid(`class Foo {static ["then"]() {}}`, `"then"`),
			classInvalid("class Foo {static [`then`]() {}}", "`then`"),
			classInvalid(`const THEN = "then";class Foo {static [THEN]() {}}`, `THEN`, 2),
			classInvalid(`class Foo {get then() {}}`, `then`),
			classInvalid(`class Foo {get ["then"]() {}}`, `"then"`),
			classInvalid("class Foo {get [`then`]() {}}", "`then`"),
			classInvalid(`const THEN = "then";class Foo {get [THEN]() {}}`, `THEN`, 2),
			classInvalid(`class Foo {set then(v) {}}`, `then`),
			classInvalid(`class Foo {set ["then"](v) {}}`, `"then"`),
			classInvalid("class Foo {set [`then`](v) {}}", "`then`"),
			classInvalid(`const THEN = "then";class Foo {set [THEN](v) {}}`, `THEN`, 2),
			classInvalid(`class Foo {static get then() {}}`, `then`),
			classInvalid(`class Foo {static set then(v) {}}`, `then`),
			classInvalid(`class Foo {static get ["then"]() {}}`, `"then"`),
			classInvalid("class Foo {static get [`then`]() {}}", "`then`"),
			classInvalid(`const THEN = "then";class Foo {static get [THEN]() {}}`, `THEN`, 2),

			// ---- Assign ----
			objectInvalid(`foo.then = 1`, `then`),
			objectInvalid(`foo["then"] = 1`, `"then"`),
			objectInvalid("foo[`then`] = 1", "`then`"),
			objectInvalid(`const THEN = "then";foo[THEN] = 1`, `THEN`, 2),
			objectInvalid(`foo.then += 1`, `then`),
			objectInvalid(`foo.then ||= 1`, `then`),
			objectInvalid(`foo.then ??= 1`, `then`),

			// ---- `{Object,Reflect}.defineProperty` ----
			objectInvalid(`Object.defineProperty(foo, "then", 1)`, `"then"`),
			objectInvalid("Object.defineProperty(foo, `then`, 1)", "`then`"),
			objectInvalid(`const THEN = "then";Object.defineProperty(foo, THEN, 1)`, `THEN`, 2),
			objectInvalid(`Reflect.defineProperty(foo, "then", 1)`, `"then"`),
			objectInvalid("Reflect.defineProperty(foo, `then`, 1)", "`then`"),
			objectInvalid(`const THEN = "then";Reflect.defineProperty(foo, THEN, 1)`, `THEN`, 2),

			// ---- `Object.fromEntries` ----
			objectInvalid(`Object.fromEntries([["then", 1]])`, `"then"`),
			objectInvalid(`Object.fromEntries([["then"]])`, `"then"`),
			objectInvalid("Object.fromEntries([[`then`, 1]])", "`then`"),
			objectInvalid(`const THEN = "then";Object.fromEntries([[THEN, 1]])`, `THEN`, 2),
			objectInvalid(`Object.fromEntries([foo, ["then", 1]])`, `"then"`),

			// ---- `export` ----
			exportInvalid(`const then = 1; export {then}`, `then`, 2),
			exportInvalid(`const notThen = 1; export {notThen as then}`, `then`),
			exportInvalid(`export {then} from "foo"`, `then`),
			exportInvalid(`export function then() {}`, `then`),
			exportInvalid(`export async function then() {}`, `then`),
			exportInvalid(`export function * then() {}`, `then`),
			exportInvalid(`export async function * then() {}`, `then`),
			exportInvalid(`export class then {}`, `then`),

			// ---- `export variables` ----
			exportInvalid(`export const then = 1`, `then`),
			exportInvalid(`export let then = 1`, `then`),
			exportInvalid(`export var then = 1`, `then`),
			exportInvalid(`export const [then] = 1`, `then`),
			exportInvalid(`export let [then] = 1`, `then`),
			exportInvalid(`export var [then] = 1`, `then`),
			exportInvalid(`export const [, then] = 1`, `then`),
			exportInvalid(`export let [, then] = 1`, `then`),
			exportInvalid(`export var [, then] = 1`, `then`),
			exportInvalid(`export const [, ...then] = 1`, `then`),
			exportInvalid(`export let [, ...then] = 1`, `then`),
			exportInvalid(`export var [, ...then] = 1`, `then`),
			exportInvalid(`export const {then} = 1`, `then`),
			exportInvalid(`export let {then} = 1`, `then`),
			exportInvalid(`export var {then} = 1`, `then`),
			exportInvalid(`export const {foo, ...then} = 1`, `then`),
			exportInvalid(`export let {foo, ...then} = 1`, `then`),
			exportInvalid(`export var {foo, ...then} = 1`, `then`),
			exportInvalid(`export const {foo: {bar: [{baz: then}]}} = 1`, `then`),
			exportInvalid(`export const notThen = 1, then = 1`, `then`),
		},
	)
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js"}
}

func objectInvalid(code string, target string, occurrence ...int) rule_tester.InvalidTestCase {
	return invalid(code, target, messageIDObject, messageObject, occurrence...)
}

func classInvalid(code string, target string, occurrence ...int) rule_tester.InvalidTestCase {
	return invalid(code, target, messageIDClass, messageClass, occurrence...)
}

func exportInvalid(code string, target string, occurrence ...int) rule_tester.InvalidTestCase {
	return invalid(code, target, messageIDExport, messageExport, occurrence...)
}

func invalid(code string, target string, messageID string, message string, occurrence ...int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, messageID, message, occurrence...),
		},
	}
}

func expectedError(code string, target string, messageID string, message string, occurrence ...int) rule_tester.InvalidTestCaseError {
	nth := 1
	if len(occurrence) > 0 {
		nth = occurrence[0]
	}

	offset := nthIndex(code, target, nth)
	if offset < 0 {
		panic("target not found in no-thenable test: " + target + " in " + code)
	}

	line, column := lineColumnForOffset(code, offset)
	endLine, endColumn := lineColumnForOffset(code, offset+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func nthIndex(s string, target string, nth int) int {
	searchStart := 0
	for i := range nth {
		offset := strings.Index(s[searchStart:], target)
		if offset < 0 {
			return -1
		}
		searchStart += offset
		if i == nth-1 {
			return searchStart
		}
		searchStart += len(target)
	}
	return -1
}

func lineColumnForOffset(code string, offset int) (int, int) {
	line := 1
	column := 1
	for i := range offset {
		if code[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
