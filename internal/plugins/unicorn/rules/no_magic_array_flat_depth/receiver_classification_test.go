// cspell:ignore ACFB Hrkt

package no_magic_array_flat_depth_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_magic_array_flat_depth"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMagicArrayFlatDepthLargeStaticMapReceiver(t *testing.T) {
	var source strings.Builder
	source.WriteString("new Map([")
	for index := range 2000 {
		if index > 0 {
			source.WriteByte(',')
		}
		source.WriteByte('[')
		source.WriteString(strconv.Itoa(index))
		source.WriteByte(',')
		source.WriteString(strconv.Itoa(index))
		source.WriteByte(']')
	}
	source.WriteString("]).get(-1).flat(2)")
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_magic_array_flat_depth.NoMagicArrayFlatDepthRule,
		[]rule_tester.ValidTestCase{jsValid(source.String())},
		nil,
	)
}

func TestNoMagicArrayFlatDepthSourceOnlyRecursionGuard(t *testing.T) {
	conditional := func(depth int, tail func(*strings.Builder, string)) string {
		var source strings.Builder
		source.WriteString("const value0 = 1;\n")
		for index := 1; index <= depth; index++ {
			source.WriteString("const value")
			source.WriteString(strconv.Itoa(index))
			source.WriteString(" = true ? value")
			source.WriteString(strconv.Itoa(index - 1))
			source.WriteString(" : 0;\n")
		}
		tail(&source, "value"+strconv.Itoa(depth))
		return source.String()
	}
	alias := func(depth int, tail func(*strings.Builder, string)) string {
		var source strings.Builder
		source.WriteString("const value0 = 1;\n")
		for index := 1; index <= depth; index++ {
			source.WriteString("const value")
			source.WriteString(strconv.Itoa(index))
			source.WriteString(" = value")
			source.WriteString(strconv.Itoa(index - 1))
			source.WriteString(";\n")
		}
		tail(&source, "value"+strconv.Itoa(depth))
		return source.String()
	}
	direct := func(source *strings.Builder, value string) {
		source.WriteString(value)
		source.WriteString(".flat(2);\n")
	}
	mathAbs := func(source *strings.Builder, value string) {
		source.WriteString("Math.abs(")
		source.WriteString(value)
		source.WriteString(").flat(2);\n")
	}
	warmArguments := func(source *strings.Builder, _ string) {
		source.WriteString("Math.max(")
		for index := 400; index <= 2000; index += 400 {
			if index > 400 {
				source.WriteByte(',')
			}
			source.WriteString("value")
			source.WriteString(strconv.Itoa(index))
		}
		source.WriteString(").flat(2);\n")
	}
	warmRoots := func(source *strings.Builder, _ string) {
		for index := 400; index <= 2000; index += 400 {
			source.WriteString("value")
			source.WriteString(strconv.Itoa(index))
			source.WriteString(".flat(2);\n")
		}
	}
	warmAliasRoots := func(source *strings.Builder, _ string) {
		for index := 400; index <= 2000; index += 400 {
			source.WriteString("Math.abs(value")
			source.WriteString(strconv.Itoa(index))
			source.WriteString(").flat(2);\n")
		}
	}
	bigIntConditionalCache := func() string {
		var source strings.Builder
		source.WriteString("const closed = ")
		for range 700 {
			source.WriteString("true ? (")
		}
		source.WriteString("2n ** 16n")
		for range 700 {
			source.WriteString(") : 0n")
		}
		source.WriteString(";\nBoolean(closed).flat(2);\nconst cacheAlias0 = closed;\n")
		for depth := 1; depth <= 700; depth++ {
			source.WriteString("const cacheAlias")
			source.WriteString(strconv.Itoa(depth))
			source.WriteString(" = cacheAlias")
			source.WriteString(strconv.Itoa(depth - 1))
			source.WriteString(";\n")
		}
		source.WriteString("Boolean(cacheAlias700).flat(2);\n")
		return source.String()
	}
	bigIntOperatorCache := func() string {
		var source strings.Builder
		source.WriteString("const closed = 2n")
		for range 700 {
			source.WriteString(" | 0n")
		}
		source.WriteString(";\nBoolean(closed).flat(2);\nconst cacheAlias0 = closed;\n")
		for depth := 1; depth <= 400; depth++ {
			source.WriteString("const cacheAlias")
			source.WriteString(strconv.Itoa(depth))
			source.WriteString(" = cacheAlias")
			source.WriteString(strconv.Itoa(depth - 1))
			source.WriteString(";\n")
		}
		source.WriteString("Boolean(cacheAlias400).flat(2);\n")
		return source.String()
	}
	classificationCacheReplay := func() string {
		var source strings.Builder
		source.WriteString("const value0 = 1;\n")
		for depth := 1; depth <= 400; depth++ {
			source.WriteString("const value")
			source.WriteString(strconv.Itoa(depth))
			source.WriteString(" = value")
			source.WriteString(strconv.Itoa(depth - 1))
			source.WriteString(";\n")
		}
		source.WriteString("value400.flat(2);\n")
		for range 700 {
			source.WriteString("(true ? ")
		}
		source.WriteString("value400")
		for range 700 {
			source.WriteString(" : 0)")
		}
		source.WriteString(".flat(2);\n")
		return source.String()
	}
	classificationFanout := func(source *strings.Builder, value string) {
		for range 1000 {
			direct(source, value)
		}
	}
	classificationDistinctRoots := func(source *strings.Builder, value string) {
		for index := range 1000 {
			source.WriteString("const root")
			source.WriteString(strconv.Itoa(index))
			source.WriteString(" = ")
			source.WriteString(value)
			source.WriteString("; root")
			source.WriteString(strconv.Itoa(index))
			source.WriteString(".flat(2);\n")
		}
	}
	classificationDiamond := func() string {
		var source strings.Builder
		source.WriteString("const diamond0 = 1;\n")
		for depth := 1; depth <= 24; depth++ {
			source.WriteString("const diamond")
			source.WriteString(strconv.Itoa(depth))
			source.WriteString(" = flag ? diamond")
			source.WriteString(strconv.Itoa(depth - 1))
			source.WriteString(" : diamond")
			source.WriteString(strconv.Itoa(depth - 1))
			source.WriteString(";\n")
		}
		for range 50 {
			source.WriteString("diamond24.flat(2);\n")
		}
		return source.String()
	}
	classificationUnknownDiamond := func() string {
		var source strings.Builder
		source.WriteString("const unknownDiamond0 = globalThis.flag ? 1 : [];\n")
		for depth := 1; depth <= 24; depth++ {
			source.WriteString("const unknownDiamond")
			source.WriteString(strconv.Itoa(depth))
			source.WriteString(" = globalThis.flag ? unknownDiamond")
			source.WriteString(strconv.Itoa(depth - 1))
			source.WriteString(" : unknownDiamond")
			source.WriteString(strconv.Itoa(depth - 1))
			source.WriteString(";\n")
		}
		source.WriteString("unknownDiamond24.flat(2);\n")
		return source.String()
	}
	classificationTerminalPoison := func() string {
		var source strings.Builder
		source.WriteString("const poison0 = 1;\n")
		for depth := 1; depth <= 400; depth++ {
			source.WriteString("const poison")
			source.WriteString(strconv.Itoa(depth))
			source.WriteString(" = poison")
			source.WriteString(strconv.Itoa(depth - 1))
			source.WriteString(";\n")
		}
		for range 700 {
			source.WriteString("(true ? ")
		}
		source.WriteString("poison400")
		for range 700 {
			source.WriteString(" : 0)")
		}
		source.WriteString(".flat(2);\npoison400.flat(2);\n")
		return source.String()
	}
	classificationTerminalRecovery := func() string {
		var source strings.Builder
		source.WriteString(alias(1100, func(source *strings.Builder, value string) {
			direct(source, value)
			source.WriteString("value77.flat(2);\n")
		}))
		return source.String()
	}
	errors := make([]rule_tester.InvalidTestCaseError, 4)
	for index := range errors {
		errors[index] = rule_tester.InvalidTestCaseError{MessageId: messageID, Message: messageString}
	}
	aliasErrors := errors[:3]
	cycleErrors := errors[:2]
	fanoutErrors := make([]rule_tester.InvalidTestCaseError, 1000)
	for index := range fanoutErrors {
		fanoutErrors[index] = rule_tester.InvalidTestCaseError{MessageId: messageID, Message: messageString}
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_magic_array_flat_depth.NoMagicArrayFlatDepthRule,
		[]rule_tester.ValidTestCase{
			jsValid(conditional(128, mathAbs)),
			jsValid(conditional(511, direct)),
			jsValid(alias(128, mathAbs)),
			jsValid(alias(512, classificationFanout)),
			jsValid(alias(512, classificationDistinctRoots)),
			jsValid(classificationDiamond()),
		},
		[]rule_tester.InvalidTestCase{
			depthInvalid(conditional(512, direct)),
			depthInvalid(conditional(1100, mathAbs)),
			depthInvalid(alias(1100, mathAbs)),
			depthInvalid(conditional(2000, warmArguments)),
			{Code: conditional(2000, warmRoots), FileName: "file.js", Errors: errors},
			{Code: alias(2000, warmAliasRoots), FileName: "file.js", Errors: aliasErrors},
			depthInvalid(bigIntConditionalCache()),
			depthInvalid(bigIntOperatorCache()),
			depthInvalid(classificationCacheReplay()),
			depthInvalid(classificationUnknownDiamond()),
			{Code: alias(1100, classificationFanout), FileName: "file.js", Errors: fanoutErrors},
			{Code: alias(1100, classificationDistinctRoots), FileName: "file.js", Errors: fanoutErrors},
			{Code: classificationTerminalPoison(), FileName: "file.js", Errors: errors[:1]},
			{Code: classificationTerminalRecovery(), FileName: "file.js", Errors: errors[:1]},
			{
				Code: "const cycleA = cycleB; const cycleB = cycleA;\n" +
					"cycleA.flat(2); cycleB.flat(2);",
				FileName: "file.js",
				Errors:   cycleErrors,
			},
			{
				Code: "const mixedCycleA = globalThis.flag ? 1 : mixedCycleB; const mixedCycleB = mixedCycleA;\n" +
					"mixedCycleA.flat(2); mixedCycleB.flat(2);",
				FileName: "file.js",
				Errors:   cycleErrors,
			},
		},
	)
}

func TestNoMagicArrayFlatDepthReceiverClassification(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `declare const value: Set<number>; value.flat(2);`, FileName: "file.ts"},
	}
	for _, code := range []string{
		`(() => {}).flat(2)`,
		`(class {}).flat(2)`,
		`Number(1).flat(2)`,
		`String("x").flat(2)`,
		`Boolean(0).flat(2)`,
		`BigInt(1).flat(2)`,
		`Number().flat(2)`,
		`String().flat(2)`,
		`Boolean().flat(2)`,
		`String.fromCharCode(65).flat(2)`,
		`String.fromCodePoint(65).flat(2)`,
		`String.raw({raw: ["x"]}).flat(2)`,
		`Math.abs(-2).flat(2)`,
		`Math.max(1, 2).flat(2)`,
		`Math.abs().flat(2)`,
		`Math.max().flat(2)`,
		`Array.isArray([]).flat(2)`,
		`parseInt("2", 10).flat(2)`,
		`parseFloat("2").flat(2)`,
		`parseInt().flat(2)`,
		`parseFloat().flat(2)`,
		`Number.parseInt("2", 10).flat(2)`,
		`Object().flat(2)`,
		`Object(null).flat(2)`,
		`Object.is(1, 1).flat(2)`,
		`Object.isFrozen({}).flat(2)`,
		`decodeURI("x").flat(2)`,
		`Date(0).flat(2)`,
		`RegExp("x").flat(2)`,
		`RegExp("\\p{L}", "u").flat(2)`,
		`RegExp("\\p{Script=Greek}", "u").flat(2)`,
		`RegExp("(?<a>a)\\k<a>", "u").flat(2)`,
		`RegExp("(?:(?<a>a))(?<b>b)", "u").flat(2)`,
		`RegExp("(?<\\u0061>a)", "u").flat(2)`,
		`RegExp("(?<\\u{1d49c}>a)").flat(2)`,
		`RegExp("(?<a>a)[(?<\\u{61}>)]").flat(2)`,
		`RegExp("\\k<a>").flat(2)`,
		`RegExp("\\p{IsGreek}").flat(2)`,
		`RegExp("\\p{").flat(2)`,
		`RegExp("\\k<").flat(2)`,
		`RegExp("\\p{Script=Garay}", "u").flat(2)`,
		`RegExp("[[A--B]]", "v").flat(2)`,
		`RegExp("".padEnd(98301, "(a)")).flat(2)`,
		`RegExp("".padEnd(1000, "[") + "a" + "".padEnd(1000, "]"), "v").flat(2)`,
		`(new RegExp(/a/g).global ? 1 : []).flat(2)`,
		`(new RegExp(/a/g, undefined).flags === "g" ? 1 : []).flat(2)`,
		`(new RegExp(RegExp("a", "u")).unicode ? 1 : []).flat(2)`,
		`(new RegExp(/a/g, "i").flags === "i" ? 1 : []).flat(2)`,
		`const re = /a/g; (RegExp(re) === re ? 1 : []).flat(2)`,
		`RegExp({[Symbol.match]: false, source: "[", flags: "u"}).flat(2)`,
		`RegExp({[Symbol.match]: 0, source: "[", flags: "u"}).flat(2)`,
		`RegExp({[Symbol.match]: null, source: "[", flags: "u"}).flat(2)`,
		`RegExp({[Symbol.match]: undefined, source: "[", flags: "u"}).flat(2)`,
		`RegExp({source: "[", flags: "u"}).flat(2)`,
		`(RegExp({[Symbol.match]: true, source: "x", flags: "g"}).flags === "g" ? 1 : []).flat(2)`,
		`(RegExp({[Symbol.match]: true, source: "x", flags: "g"}, "").flags === "" ? 1 : []).flat(2)`,
		`(RegExp({[Symbol.match]: true}).source === "(?:)" ? 1 : []).flat(2)`,
		`(RegExp({[Symbol.match]: true, source: undefined, flags: undefined}).source === "(?:)" ? 1 : []).flat(2)`,
		`(RegExp({__proto__: {[Symbol.match]: true, source: "x", flags: "g"}}).flags === "g" ? 1 : []).flat(2)`,
		`(RegExp({[Symbol.match]: true, source: "x", flags: "g"}, undefined).flags === "g" ? 1 : []).flat(2)`,
		`(RegExp({[Symbol.match]: true, source: "x", flags: "g"}, "i").flags === "i" ? 1 : []).flat(2)`,
		`(new RegExp({[Symbol.match]: true, source: "x", flags: "g"}).flags === "g" ? 1 : []).flat(2)`,
		`RegExp({[Symbol.match]: true, constructor: RegExp, source: "[", flags: "u"}).flat(2)`,
		`(RegExp(RegExp.prototype) === RegExp.prototype ? 1 : []).flat(2)`,
		`(new RegExp(RegExp.prototype).source === "(?:)" ? 1 : []).flat(2)`,
		`const inherited = {__proto__: RegExp.prototype}; (RegExp(inherited) ? 1 : []).flat(2)`,
		`[1].at(0).flat(2)`,
		`[1].every(Boolean).flat(2)`,
		`[1].find(Boolean).flat(2)`,
		`[1].findIndex(Boolean).flat(2)`,
		`[1].some(Boolean).flat(2)`,
		`new Map().get("x").flat(2)`,
		`new Set([1]).has(1).flat(2)`,
		`Object.freeze({}).flat(2)`,
		`Object.freeze().flat(2)`,
		`Boolean(Symbol.iterator).flat(2)`,
		`String(Symbol.iterator).flat(2)`,
		`(Symbol.iterator.description === "Symbol.iterator" ? 1 : []).flat(2)`,
		`(String(Symbol.iterator) === "Symbol(Symbol.iterator)" ? 1 : []).flat(2)`,
		`(typeof Symbol.dispose === "symbol" ? 1 : []).flat(2)`,
		`const value = 1; value.flat(2)`,
		`const value = {}; value.flat(2)`,
		`const value = "x"; value.flat(2)`,
		`const value = function () {}; value.flat(2)`,
		`const value = () => {}; value.flat(2)`,
		`const value = Math.abs(-2); value.flat(2)`,
		`(flag ? 2 : 1).flat(2)`,
		`(flag ? Math.abs(-2) : 1).flat(2)`,
		`(true ? 1 : []).flat(2)`,
		`(unknown, 1).flat(2)`,
		`(1, Math.PI).flat(2)`,
		`(false || 1).flat(2)`,
		`(0 && []).flat(2)`,
		`(Math.abs(-2) && 1).flat(2)`,
		`(Math.abs(0) && []).flat(2)`,
		`(value = 1).flat(2)`,
		`const M = Math; M.abs(-2).flat(2)`,
		`let M = Math; M.abs(-2).flat(2)`,
		`const abs = Math.abs; abs(-2).flat(2)`,
		`const integer = parseInt; integer("2", 10).flat(2)`,
		`const toNumber = Number; toNumber(1).flat(2)`,
		`const A = Array; A.isArray([]).flat(2)`,
		`const Math = {abs: Number}; Math.abs(-2).flat(2)`,
		`const Array = {of: Number}; Array.of(1).flat(2)`,
		`"".padStart(0, Symbol.iterator).flat(2)`,
		`"".padEnd(Infinity, "").flat(2)`,
		`(2n ** 1000001n).flat(2)`,
		`(1n << 1000001n).flat(2)`,
		`(1n >> -1000001n).flat(2)`,
		`(Number(new Date(new String("2020-01-01"))) === 1577836800000 ? 1 : []).flat(2)`,
		`(Object.is(Number(parseInt), NaN) ? 1 : []).flat(2)`,
		`(BigInt([]) === 0n ? 1 : []).flat(2)`,
		`(BigInt(new Date(0)) === 0n ? 1 : []).flat(2)`,
		`((1e-7).toExponential(20) === "9.99999999999999954748e-8" ? 1 : []).flat(2)`,
		`((1.5).toPrecision(1) === "2" ? 1 : []).flat(2)`,
		`((1.25).toExponential(1) === "1.3e+0" ? 1 : []).flat(2)`,
		`((1.25).toPrecision(2) === "1.3" ? 1 : []).flat(2)`,
	} {
		valid = append(valid, jsValid(code))
	}

	invalid := make([]rule_tester.InvalidTestCase, 0)
	for _, code := range []string{
		`Array.of(1).flat(2)`,
		`Object.freeze([]).flat(2)`,
		`Object([]).flat(2)`,
		`Object.seal([]).flat(2)`,
		`Object.preventExtensions([]).flat(2)`,
		`Math.random().flat(2)`,
		`Math.abs(value).flat(2)`,
		`Array.isArray(value).flat(2)`,
		`parseInt(value).flat(2)`,
		`Object(value).flat(2)`,
		`Object.freeze(value).flat(2)`,
		`Number(value).flat(2)`,
		`String(value).flat(2)`,
		`Boolean(value).flat(2)`,
		`BigInt(value).flat(2)`,
		`BigInt().flat(2)`,
		`BigInt("x").flat(2)`,
		`BigInt(1.2).flat(2)`,
		`BigInt(null).flat(2)`,
		`Symbol().flat(2)`,
		`Math.abs(1n).flat(2)`,
		`Math.abs(Uint8Array.prototype).flat(2)`,
		`Number(BigInt64Array.prototype).flat(2)`,
		`Number(Symbol.iterator).flat(2)`,
		`parseInt(Symbol.iterator).flat(2)`,
		`String.fromCharCode(Symbol.iterator).flat(2)`,
		`decodeURI("%").flat(2)`,
		`RegExp("[").flat(2)`,
		`RegExp("\\p{Greek}", "u").flat(2)`,
		`RegExp("\\p{Hyphen}", "u").flat(2)`,
		`RegExp("[\\P{Latin}]", "u").flat(2)`,
		`RegExp("]", "u").flat(2)`,
		`RegExp("a{2,1}").flat(2)`,
		`RegExp("[a-\\d]", "u").flat(2)`,
		`RegExp("(?=a)+", "u").flat(2)`,
		`RegExp("(?i:a)").flat(2)`,
		`RegExp("(?-i:a)").flat(2)`,
		`RegExp("(?:(?<a>a))(?<a>b)").flat(2)`,
		`RegExp("(?<a>a)|(?<\\u0061>b)", "u").flat(2)`,
		`RegExp("\\k<a>", "u").flat(2)`,
		`RegExp("(?<a>a)\\k<b>").flat(2)`,
		`RegExp("(?>a)").flat(2)`,
		`RegExp("[/]", "v").flat(2)`,
		`RegExp("\\p{Script=Hrkt}", "u").flat(2)`,
		`RegExp("\\p{", "u").flat(2)`,
		`RegExp("".padEnd(600000, "\\p{"), "u").flat(2)`,
		`RegExp("".padEnd(98304, "(a)")).flat(2)`,
		`RegExp("".padEnd(4000, "[") + "a" + "".padEnd(4000, "]"), "v").flat(2)`,
		`(new RegExp(/a/g).global ? [] : 1).flat(2)`,
		`(new RegExp(/a/g, undefined).flags === "g" ? [] : 1).flat(2)`,
		`(new RegExp(RegExp("a", "u")).unicode ? [] : 1).flat(2)`,
		`const re = /a/g; (RegExp(re) === re ? [] : 1).flat(2)`,
		`const re = /a/g; (new RegExp(re) === re ? 1 : []).flat(2)`,
		`RegExp({[Symbol.match]: true, source: "[", flags: "u"}).flat(2)`,
		`RegExp({[Symbol.match]: true, source: "x", flags: "gg"}).flat(2)`,
		`RegExp({[Symbol.match]: true, source: "x", get flags() { throw new Error(); }}).flat(2)`,
		`RegExp({[Symbol.match]: 1, source: "[", flags: "u"}).flat(2)`,
		`RegExp({__proto__: {[Symbol.match]: true, source: "[", flags: "u"}}).flat(2)`,
		`RegExp({[Symbol.match]: {}, source: "[", flags: "u"}).flat(2)`,
		`RegExp({[Symbol.match]: Symbol.iterator, source: "[", flags: "u"}).flat(2)`,
		`RegExp({[Symbol.match]: true, source: Symbol.iterator, flags: ""}).flat(2)`,
		`RegExp({[Symbol.match]: true, source: "x", flags: Symbol.iterator}).flat(2)`,
		`(new RegExp({[Symbol.match]: true, source: "[", flags: "u"}).source ? 1 : []).flat(2)`,
		`(new RegExp({[Symbol.match]: true, source: "x", flags: "gg"}).source ? 1 : []).flat(2)`,
		`(RegExp(RegExp.prototype) === RegExp.prototype ? [] : 1).flat(2)`,
		`(new RegExp(RegExp.prototype) === RegExp.prototype ? 1 : []).flat(2)`,
		`(RegExp(RegExp.prototype, undefined) === RegExp.prototype ? [] : 1).flat(2)`,
		`(RegExp(RegExp.prototype, "") === RegExp.prototype ? 1 : []).flat(2)`,
		`(new RegExp(RegExp.prototype).source !== "(?:)" ? 1 : []).flat(2)`,
		`(RegExp(RegExp.prototype, "").source !== "(?:)" ? 1 : []).flat(2)`,
		`(new RegExp(RegExp.prototype, "g").flags !== "g" ? 1 : []).flat(2)`,
		`const inherited = {__proto__: RegExp.prototype}; (new RegExp(inherited) ? 1 : []).flat(2)`,
		`(Symbol.iterator.description === "iterator" ? 1 : []).flat(2)`,
		`(String(Symbol.iterator) === "Symbol(iterator)" ? 1 : []).flat(2)`,
		`Symbol.dispose.flat(2)`,
		`Symbol.asyncDispose.flat(2)`,
		`(Symbol.dispose.description === "dispose" ? 1 : []).flat(2)`,
		`(Symbol.dispose === Symbol.for("nodejs.dispose") ? 1 : []).flat(2)`,
		`(Symbol.dispose === Symbol.for("nodejs.dispose") ? [] : 1).flat(2)`,
		`(Symbol.keyFor(Symbol.dispose) === undefined ? 1 : []).flat(2)`,
		`(new Set([Symbol.dispose]).has(Symbol.for("nodejs.dispose")) ? 1 : []).flat(2)`,
		`decodeURI("%FF").flat(2)`,
		`(new Date(Date.prototype) ? 1 : []).flat(2)`,
		`(new Date({[Symbol.toPrimitive]: 0}) ? 1 : []).flat(2)`,
		`(new Date({valueOf: 0, toString: 0}) ? 1 : []).flat(2)`,
		`[[]].at(0).flat(2)`,
		`[1].filter(Boolean).flat(2)`,
		`[[]].find(Boolean).flat(2)`,
		`new Map([[1, []]]).get(1).flat(2)`,
		`Object.keys({a: 1}).flat(2)`,
		`Object.values({a: 1}).flat(2)`,
		`Array.from([1]).flat(2)`,
		`JSON.parse("[]").flat(2)`,
		`const value = []; value.flat(2)`,
		`const value = Array.of(1); value.flat(2)`,
		`const A = Array; A.of(1).flat(2)`,
		`const of = Array.of; of(1).flat(2)`,
		`const value = Math.PI; value.flat(2)`,
		`const a = b; const b = a; a.flat(2)`,
		`let value = 1; value.flat(2)`,
		`var value = 1; value.flat(2)`,
		`const {value = 1} = {}; value.flat(2)`,
		`(flag ? [] : []).flat(2)`,
		`(flag ? [] : 1).flat(2)`,
		`(flag ? Math.PI : 1).flat(2)`,
		`(true ? [] : 1).flat(2)`,
		`(unknown, []).flat(2)`,
		`(true && []).flat(2)`,
		`(flag || 1).flat(2)`,
		`(flag && 1).flat(2)`,
		`(flag ?? 1).flat(2)`,
		`(value = []).flat(2)`,
		`undefined.flat(2)`,
		`NaN.flat(2)`,
		`Infinity.flat(2)`,
		`Number.NaN.flat(2)`,
		`Math.PI.flat(2)`,
		`Symbol.iterator.flat(2)`,
		`new Uint8Array().flat(2)`,
		`({flat() {}}).flat(2)`,
		`"x".flat(2)`,
		`let M = Math; M = other; M.abs(-2).flat(2)`,
		`const Math = {abs(value) { return value; }}; Math.abs(-2).flat(2)`,
		`(Object.is((-0) ** 3, -0) ? [] : 1).flat(2)`,
		`(Object.is(Math.sin(0), 0) ? [] : 1).flat(2)`,
		`(Object.isFrozen(Object.freeze({})) ? 1 : []).flat(2)`,
		`(Boolean.prototype.toLocaleString === Object.prototype.toLocaleString ? [] : 1).flat(2)`,
		`(new Set([Number.parseInt]).has(parseInt) ? [] : 1).flat(2)`,
		`new Map([[Number.parseInt, []]]).get(parseInt).flat(2)`,
		`(/x/.lastIndex === 0 ? [] : 1).flat(2)`,
		`"".padStart(Infinity, "x").flat(2)`,
		`"".padEnd(Infinity).flat(2)`,
		`(2n ** 2147483648n).flat(2)`,
		`(1n << 2147483648n).flat(2)`,
		`(1n >> -2147483648n).flat(2)`,
		`(Number(new Date(new String("2020-01-01"))) === 1577836800000 ? [] : 1).flat(2)`,
		`(Object.is(Number(parseInt), NaN) ? [] : 1).flat(2)`,
		`(BigInt([]) === 0n ? [] : 1).flat(2)`,
		`(BigInt(new Date(0)) === 0n ? [] : 1).flat(2)`,
		`((1e-7).toExponential(20) === "9.99999999999999954748e-8" ? [] : 1).flat(2)`,
		`((1.5).toPrecision(1) === "2" ? [] : 1).flat(2)`,
		`((1.25).toExponential(1) === "1.3e+0" ? [] : 1).flat(2)`,
		`((1.25).toPrecision(2) === "1.3" ? [] : 1).flat(2)`,
		`((1e21).toString(36) === "5v1j4f4ds7c000" ? [] : 1).flat(2)`,
		`("\u{1CCD6}".normalize("NFKC") === "A" ? [] : 1).flat(2)`,
		`("\uD800e\u0301".normalize("NFC") === "\uD800é" ? [] : 1).flat(2)`,
		`("\uD800\u{1CCD6}".normalize("NFKC") === "\uD800A" ? [] : 1).flat(2)`,
		`("A\u0345\u0897".normalize("NFD") === "A\u0897\u0345" ? [] : 1).flat(2)`,
		`("\uA7CE".toLowerCase() === "\uA7CE" ? [] : 1).flat(2)`,
		`("\uA7CF".toUpperCase() === "\uA7CF" ? [] : 1).flat(2)`,
		`("\u0295Σ".toLowerCase() === "\u0295ς" ? [] : 1).flat(2)`,
		`("AΣ\u1ACFB".toLowerCase() === "aς\u1ACFb" ? [] : 1).flat(2)`,
		`using value = {}; value.flat(2)`,
		`await using value = {}; value.flat(2)`,
	} {
		invalid = append(invalid, depthInvalid(code))
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_magic_array_flat_depth.NoMagicArrayFlatDepthRule,
		valid,
		invalid,
	)
}
