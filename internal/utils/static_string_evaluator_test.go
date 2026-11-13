// cspell:ignore truenull

package utils

import (
	"math"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"gotest.tools/v3/assert"
)

func TestStaticStringEvaluatorPropertyNames(t *testing.T) {
	for _, test := range []struct {
		key   string
		want  string
		known bool
	}{
		{`fn`, "fn", true},
		{`["f" + "n"]`, "fn", true},
		{"[`f${\"n\"}`]", "fn", true},
		{`['fn'.toString()]`, "fn", true},
		{`[unknown.toString()]`, "", false},
		{`[["fn"]]`, "fn", true},
		{`[""]`, "", true},
		{`[-0]`, "0", true},
		{`[true]`, "true", true},
		{`[null]`, "null", true},
		{`[1n + 2n]`, "3", true},
		{`[key]`, "", false},
		{`[undefined]`, "", false},
		{`[String("fn")]`, "", false},
		{`#fn`, "", false},
	} {
		t.Run(test.key, func(t *testing.T) {
			evaluator := NewStaticStringEvaluatorWithoutScope()
			access := "." + test.key
			kind := ast.KindPropertyAccessExpression
			if strings.HasPrefix(test.key, "[") {
				access = test.key
				kind = ast.KindElementAccessExpression
			}
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/property.ts", Path: "/property.ts",
			}, `const key = "fn"; class C { #fn; method() { object`+access+`; } }`, core.ScriptKindTS)
			member := findFirstNodeOfKind(t, source, kind)
			if got, known := evaluator.EvalAccessExpressionName(member); got != test.want || known != test.known {
				t.Fatalf("member name = (%q, %v), want (%q, %v)", got, known, test.want, test.known)
			}
			source = parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/property.ts", Path: "/property.ts",
			}, `const key = "fn"; class C { `+test.key+` = 0; }`, core.ScriptKindTS)
			property := findFirstNodeOfKind(t, source, ast.KindPropertyDeclaration)
			if got, known := evaluator.EvalPropertyName(property.Name()); got != test.want || known != test.known {
				t.Fatalf("property name = (%q, %v), want (%q, %v)", got, known, test.want, test.known)
			}
		})
	}
	// The existing literal-only helper retains its narrower contract.
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/property.ts", Path: "/property.ts",
	}, `({["f" + "n"]: 0})`, core.ScriptKindTS)
	property := findFirstNodeOfKind(t, source, ast.KindPropertyAssignment)
	if _, known := GetStaticPropertyName(property.Name()); known {
		t.Fatal("literal-only helper unexpectedly folded a computed expression")
	}
}

func TestStaticStringEvaluator(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "file.ts")
	code := "\n" +
		"const direct = \"then\";\n" +
		"const left = \"th\";\n" +
		"const concat = left + \"en\";\n" +
		"const template = `${left}en`;\n" +
		"const asserted = (\"then\" as string);\n" +
		"const satisfiesValue = \"then\" satisfies string;\n" +
		"const nested = `${concat}`;\n" +
		"const templateBoolean = `${false}`;\n" +
		"const conditional = true ? \"then\" : \"no\";\n" +
		"const conditionalFromConst = direct ? direct : \"no\";\n" +
		"const unresolvedConditional = flag ? \"then\" : \"no\";\n" +
		"const logicalOr = \"\" || \"then\";\n" +
		"const logicalAnd = \"then\" && \"then\";\n" +
		"const nullish = null ?? \"then\";\n" +
		"const undefinedNullish = undefined ?? \"then\";\n" +
		"const stringCall = String(\"then\");\n" +
		"const stringNumberCall = String(1 + 2);\n" +
		"const reviewStringAlias = String; const reviewSymbolString = reviewStringAlias(Symbol.iterator);\n" +
		"const reviewNumberAlias = Number; const reviewParseAlias = reviewNumberAlias.parseInt; const reviewParsed = String(reviewParseAlias('10px'));\n" +
		"const reviewGlobalParseAlias = parseInt; const reviewGlobalParsed = String(reviewGlobalParseAlias('0x10'));\n" +
		"const reviewMathAlias = Math; const reviewMaxAlias = reviewMathAlias.max; const reviewInfAlias = Infinity; const reviewMax = String(reviewMaxAlias(reviewInfAlias, 2));\n" +
		"const reviewAbsAlias = reviewMathAlias.abs; const reviewAbs = String(reviewAbsAlias(-1));\n" +
		"const reviewMathKey = Math.abs(-1); const reviewMathKeyUse = String(reviewMathKey + 1);\n" +
		"const reviewTextAlias = ' foo '; const reviewTrimmed = reviewTextAlias.trim();\n" +
		"const reviewChainAlias = null?.x; const reviewChainBoundary = reviewChainAlias.y;\n" +
		"const stringNoArgumentCall = String();\n" +
		"const stringToString = \"GET\".toString();\n" +
		"const emptyToString = \"\".toString();\n" +
		"const computedToString = \"GET\"[\"to\" + \"String\"]();\n" +
		"const optionalToString = \"GET\"?.toString?.();\n" +
		"const stringToStringExtraArgument = \"GET\".toString(123);\n" +
		"const stringToStringUnknownArgument = \"GET\".toString(unknown);\n" +
		"const customToString = ({toString() { return \"GET\"; }}).toString();\n" +
		"const uppercase = \"get\".toUpperCase();\n" +
		"const uppercaseExpansion = \"ß\".toUpperCase();\n" +
		"const uppercaseStaticExtraArgument = \"get\".toUpperCase(1);\n" +
		"const uppercaseExtraArgument = \"get\".toUpperCase(unknown);\n" +
		"const lowercase = \"HEAD\".toLowerCase();\n" +
		"const lowercaseFinalSigma = \"ΟΣ\".toLowerCase();\n" +
		"const fromCharCode = String.fromCharCode(71, 69, 84);\n" +
		"const fromCharCodeNegativeModulo = String.fromCharCode(-4294967225);\n" +
		"const fromCharCodeHexString = String.fromCharCode(\"0x47\");\n" +
		"const fromCharCodeNEL = String.fromCharCode(\"\u008571\");\n" +
		"const fromCharCodeBOM = String.fromCharCode(\"\uFEFF71\");\n" +
		"const fromCharCodeSpread = String.fromCharCode(...[72, 69, 65, 68]);\n" +
		"const arrayOfFirst = Array.of(\"GET\")[0];\n" +
		"const arrayOfSpread = Array.of(...[\"HEAD\"])[0];\n" +
		"const spreadNested = [\"a\", ...[\"b\", ...[\"c\"]], \"d\"].join(\"\");\n" +
		"const spreadEmpty = [...[], ...[]].join(\",\");\n" +
		"const spreadSparse = [...[, , \"x\"]].join(\"-\");\n" +
		"const spreadSparseLength = String([...[, , \"x\"]].length);\n" +
		"const spreadPosition = String([0, ...[1, 2], 3, ...[4]][3]);\n" +
		"const spreadValues = [\"a\", \"b\"]; const spreadAlias = [...spreadValues].join(\"\");\n" +
		"const spreadCallArguments = String.fromCharCode(...[...[65, 66]]);\n" +
		"const spreadStringArguments = Array.of(...\"AB\").join(\"\");\n" +
		"const spreadUnknown = [...unknown].join(\"\");\n" +
		"const spreadCall = [...getValues()].join(\"\");\n" +
		"const spreadNonIterable = [...42].join(\"\");\n" +
		"const spreadCustomIterator = [...{*[Symbol.iterator]() { yield \"x\"; }}].join(\"\");\n" +
		"const objectSpread = ({...{value: \"x\"}}).value;\n" +
		"const objectSpreadOrder = ({value: \"first\", ...{value: \"middle\"}, value: \"last\"}).value;\n" +
		"const objectSpreadNullish = ({value: \"x\", ...null, ...undefined, ...1, ...true, ...1n}).value;\n" +
		"const objectSpreadArray = ({...[\"a\", \"b\"]})[1];\n" +
		"const objectSpreadHole = ({__proto__: {0: \"inherited\"}, ...[, ]})[0];\n" +
		"const objectSpreadUndefined = String(({__proto__: {0: \"inherited\"}, ...[undefined]})[0]);\n" +
		"const objectSpreadDensifiedHole = String(({__proto__: {0: \"inherited\"}, ...[...[, ]]})[0]);\n" +
		"const objectSpreadArrayLength = String(({...[1, 2]}).length);\n" +
		"const objectSpreadInherited = String(({...{__proto__: {value: \"inherited\"}}}).value);\n" +
		"const objectSpreadProtoData = ({...{[\"__proto__\"]: \"own\"}}).__proto__;\n" +
		"const objectSpreadRegExp = String(({.../a/}).lastIndex);\n" +
		"const objectSpreadUnknown = ({...unknown}).value;\n" +
		"const objectSpreadGetter = ({...{get value() { return \"x\"; }}}).value;\n" +
		"const stringSlice = \"xGETy\".slice(1, 4);\n" +
		"const stringSliceDefault = \"GET\".slice(undefined);\n" +
		"const stringSliceUtf16 = \"😀GETx\".slice(2, 5);\n" +
		"const stringSubstring = \"xHEADy\".substring(1, 5);\n" +
		"const stringSubstringReversed = \"HEAD\".substring(4, 0);\n" +
		"const stringSubstr = \"xGETy\".substr(1, 3);\n" +
		"const stringSubstrNegative = \"xxGET\".substr(-3);\n" +
		"const stringSubstrEmptyLength = \"GET\".substr(0, 0);\n" +
		"const stringCharAt = \"GETx\".charAt(0);\n" +
		"const stringCharAtDefault = \"GET\".charAt();\n" +
		"const stringCharAtOutOfRange = \"GET\".charAt(9);\n" +
		"const stringConcat = \"G\".concat(\"ET\");\n" +
		"const stringConcatMany = \"HE\".concat(\"A\", \"D\");\n" +
		"const stringConcatCoercion = \"n\".concat(1, true, null);\n" +
		"const stringConcatUnknown = \"G\".concat(unknownValue);\n" +
		"const unaryPlusString = String(+\"71\");\n" +
		"const unaryPlusBoolean = String(+true);\n" +
		"const unaryMinusString = String(-\"1\");\n" +
		"const unaryTildeString = String(~\"1\");\n" +
		"const unaryPlusInvalid = String(+\"nope\");\n" +
		"const unaryPlusUnknown = String(+unknownValue);\n" +
		"const strictEqualNumbers = +\"71\" === 71 ? \"then\" : \"no\";\n" +
		"const strictEqualStrings = \"a\" === \"a\" ? \"then\" : \"no\";\n" +
		"const strictUnequalStrings = \"a\" !== \"b\" ? \"then\" : \"no\";\n" +
		"const strictEqualAcrossKinds = (1 as any) === \"1\" ? \"no\" : \"then\";\n" +
		"const strictEqualNullish = (null as any) === undefined ? \"no\" : \"then\";\n" +
		"const strictEqualObjects = ({} as any) === {} ? \"then\" : \"no\";\n" +
		"const strictEqualUnknown = unknownValue === \"a\" ? \"then\" : \"no\";\n" +
		"const StringAlias = String;\n" +
		"const aliasFromCharCode = StringAlias.fromCharCode(71, 69, 84);\n" +
		"const ArrayAlias = Array;\n" +
		"const aliasArrayOf = ArrayAlias.of(\"GET\")[0];\n" +
		"{ const String = {fromCharCode: () => \"GET\"}; const shadowedFromCharCode = String.fromCharCode(71); }\n" +
		"{ const Array = {of: () => [\"GET\"]}; const shadowedArrayOf = Array.of(\"GET\")[0]; }\n" +
		"const stringRaw = String.raw`then`;\n" +
		"const stringRawSubstitution = String.raw`th${\"e\"}n`;\n" +
		"const stringRawEscapes = String.raw`\\p{Script=Kawi}\\n`;\n" +
		"const stringRawEscapedSubstitution = String.raw`\\p{Script=${\"Ka\"}wi}\\n`;\n" +
		"const stringRawEmptyParts = String.raw`${\"\\n\"}${1}`;\n" +
		"const stringRawNewlines = String.raw`a\r\nb\rc`;\n" +
		"const stringRawInvalidEscape = String.raw`\\u{110000}`;\n" +
		"const RawString = String;\n" +
		"const stringRawAlias = RawString.raw`then`;\n" +
		"let MutableRawString = String;\n" +
		"MutableRawString = {raw: value => \"then\"} as any;\n" +
		"const stringRawMutableAlias = MutableRawString.raw`then`;\n" +
		"const typedRawStringAlias: StringConstructor = fake;\n" +
		"const stringRawTypedAlias = typedRawStringAlias.raw`then`;\n" +
		"{ const String = value => \"then\"; const shadowedStringCall = String(\"then\"); }\n" +
		"{ const String = { raw: value => \"then\" }; const shadowedStringRaw = String.raw`then`; }\n" +
		"let letStatic = \"then\";\n" +
		"var varStatic = \"then\";\n" +
		"const letStaticUse = letStatic;\n" +
		"const varStaticUse = varStatic;\n" +
		"let letWritten = \"then\";\n" +
		"letWritten = \"other\";\n" +
		"const letWrittenUse = letWritten;\n" +
		"var varWritten = \"then\";\n" +
		"varWritten++;\n" +
		"const varWrittenUse = varWritten;\n" +
		"let destructuredWritten = \"then\";\n" +
		"({destructuredWritten} = other);\n" +
		"const destructuredWrittenUse = destructuredWritten;\n" +
		"const notThen = \"not-then\";\n" +
		"const cycle = cycle;\n" +
		"let letValue = \"then\";\n" +
		"const letUse = letValue;\n" +
		"const numeric = 1 + 2;\n" +
		"const unknownUse = unknownValue;\n" +
		"const controlConst = 1;\n" +
		"let controlLet = 1;\n" +
		"var controlVar = \"x\";\n" +
		"const controlArray = [1];\n" +
		"const controlObject = {value: true};\n" +
		"const controlConstUse = controlConst;\n" +
		"const controlLetUse = controlLet;\n" +
		"const controlVarUse = controlVar;\n" +
		"const controlArrayUse = controlArray;\n" +
		"const controlObjectUse = controlObject.value;\n" +
		"const stableArray = [\"\", \"message\"];\n" +
		"const stableObject = {message: \"value\"};\n" +
		"const stableArrayUse = stableArray[0];\n" +
		"const writtenArray = [\"\"];\n" +
		"writtenArray[0] = \"message\";\n" +
		"const writtenArrayUse = writtenArray[0];\n" +
		"const filledArray = [\"\"];\n" +
		"filledArray.fill(\"message\");\n" +
		"const filledArrayUse = filledArray[0];\n" +
		"const mutationMethod = \"fill\";\n" +
		"const computedMutatedArray = [\"\"];\n" +
		"computedMutatedArray[mutationMethod](\"message\");\n" +
		"const computedMutatedArrayUse = computedMutatedArray[0];\n" +
		"const updatedArray = [\"\"];\n" +
		"updatedArray[0] += \"message\";\n" +
		"const updatedArrayUse = updatedArray[0];\n" +
		"const slicedArray = [\"\"];\n" +
		"slicedArray.slice();\n" +
		"const slicedArrayUse = slicedArray[0];\n" +
		"const writtenObject = {message: \"\"};\n" +
		"writtenObject.message = \"message\";\n" +
		"const writtenObjectUse = writtenObject.message;\n" +
		"const protoMessage = ({__proto__: {message: \"message\"}}).message;\n" +
		"const quotedProtoMessage = ({\"__proto__\": {message: \"message\"}}).message;\n" +
		"const ownProtoMessage = ({__proto__: {message: \"message\"}, message: \"\"}).message;\n" +
		"const computedProtoMessage = ({[\"__proto__\"]: {message: \"message\"}}).message;\n" +
		"const nullProtoString = String(({__proto__: null}));\n" +
		"const regexSource = /g/u.source;\n" +
		"const regexComputedSource = /gi/u[\"source\"];\n" +
		"const regexFlags = /x/mi.flags;\n" +
		"const regexGlobal = String(/x/g.global);\n" +
		"const regexHasIndices = String(/x/d.hasIndices);\n" +
		"const regexUnicode = String(/x/u.unicode);\n" +
		"const regexLastIndex = String(/x/g.lastIndex);\n" +
		"const regexString = String(/x/mi);\n" +
		"const regexAlias = /alias/u;\n" +
		"const regexAliasSource = regexAlias.source;\n" +
		"const mutatedRegex = /before/u;\n" +
		"mutatedRegex.lastIndex = 1;\n" +
		"const mutatedRegexSource = mutatedRegex.source;\n" +
		"const regexUnicodeSets = /x/v.unicodeSets;\n" +
		"const regexProtoSource = ({__proto__: /u/u}).source;\n" +
		"const regexProtoLastIndex = String(({__proto__: /g/u}).lastIndex);\n" +
		"const regexProtoOwnSource = ({__proto__: /u/u, source: \"own\"}).source;\n" +
		"const regexProtoString = String({__proto__: /u/u});\n" +
		"const emptyJoin = [].join();\n" +
		"const joined = [\"error\", \"message\"].join(\": \");\n" +
		"const nestedEmptyJoin = [[]].join();\n" +
		"const overflowEmptyJoin = [\"\", \"\", \"\"].join(\"\");\n" +
		"const undefinedSeparatorJoin = [].join(undefined);\n" +
		"const boundJoinArray = [\"\"];\n" +
		"const boundJoin = boundJoinArray.join();\n" +
		"const mutatedJoinArray = [\"\"];\n" +
		"mutatedJoinArray.fill(\"message\");\n" +
		"const mutatedJoin = mutatedJoinArray.join();\n"

	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err, "couldn't create program")

	sourceFile := program.GetSourceFile(filePath)
	assert.Assert(t, sourceFile != nil)

	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()

	staticEvaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
	tests := []struct {
		name string
		want string
		ok   bool
	}{
		{name: "direct", want: "then", ok: true},
		{name: "reviewSymbolString", want: "Symbol(Symbol.iterator)", ok: true},
		{name: "reviewParsed", want: "10", ok: true},
		{name: "reviewGlobalParsed", want: "16", ok: true},
		{name: "reviewAbs", want: "1", ok: true},
		{name: "reviewMathKeyUse", want: "2", ok: true},
		{name: "reviewMax", want: "Infinity", ok: true},
		{name: "reviewTrimmed", want: "foo", ok: true},
		{name: "reviewChainBoundary"},
		{name: "concat", want: "then", ok: true},
		{name: "template", want: "then", ok: true},
		{name: "asserted", want: "then", ok: true},
		{name: "satisfiesValue", want: "then", ok: true},
		{name: "nested", want: "then", ok: true},
		{name: "templateBoolean", want: "false", ok: true},
		{name: "conditional", want: "then", ok: true},
		{name: "conditionalFromConst", want: "then", ok: true},
		{name: "unresolvedConditional"},
		{name: "logicalOr", want: "then", ok: true},
		{name: "logicalAnd", want: "then", ok: true},
		{name: "nullish", want: "then", ok: true},
		{name: "undefinedNullish", want: "then", ok: true},
		{name: "stringCall", want: "then", ok: true},
		{name: "stringNumberCall", want: "3", ok: true},
		{name: "stringNoArgumentCall", want: "", ok: true},
		{name: "stringToString", want: "GET", ok: true},
		{name: "emptyToString", want: "", ok: true},
		{name: "computedToString", want: "GET", ok: true},
		{name: "optionalToString", want: "GET", ok: true},
		{name: "stringToStringExtraArgument", want: "GET", ok: true},
		{name: "stringToStringUnknownArgument"},
		{name: "customToString"},
		{name: "uppercase", want: "GET", ok: true},
		{name: "uppercaseExpansion", want: "SS", ok: true},
		{name: "uppercaseStaticExtraArgument", want: "GET", ok: true},
		{name: "uppercaseExtraArgument"},
		{name: "lowercase", want: "head", ok: true},
		{name: "lowercaseFinalSigma", want: "ος", ok: true},
		{name: "fromCharCode", want: "GET", ok: true},
		{name: "fromCharCodeNegativeModulo", want: "G", ok: true},
		{name: "fromCharCodeHexString", want: "G", ok: true},
		{name: "fromCharCodeNEL", want: "\x00", ok: true},
		{name: "fromCharCodeBOM", want: "G", ok: true},
		{name: "fromCharCodeSpread", want: "HEAD", ok: true},
		{name: "arrayOfFirst", want: "GET", ok: true},
		{name: "arrayOfSpread", want: "HEAD", ok: true},
		{name: "spreadNested", want: "abcd", ok: true},
		{name: "spreadEmpty", want: "", ok: true},
		{name: "spreadSparse", want: "--x", ok: true},
		{name: "spreadSparseLength", want: "3", ok: true},
		{name: "spreadPosition", want: "3", ok: true},
		{name: "spreadAlias", want: "ab", ok: true},
		{name: "spreadCallArguments", want: "AB", ok: true},
		{name: "spreadStringArguments", want: "AB", ok: true},
		{name: "spreadUnknown"},
		{name: "spreadCall"},
		{name: "spreadNonIterable"},
		{name: "spreadCustomIterator"},
		{name: "objectSpread", want: "x", ok: true},
		{name: "objectSpreadOrder", want: "last", ok: true},
		{name: "objectSpreadNullish", want: "x", ok: true},
		{name: "objectSpreadArray", want: "b", ok: true},
		{name: "objectSpreadHole", want: "inherited", ok: true},
		{name: "objectSpreadUndefined", want: "undefined", ok: true},
		{name: "objectSpreadDensifiedHole", want: "undefined", ok: true},
		{name: "objectSpreadArrayLength", want: "undefined", ok: true},
		{name: "objectSpreadInherited", want: "undefined", ok: true},
		{name: "objectSpreadProtoData", want: "own", ok: true},
		{name: "objectSpreadRegExp", want: "undefined", ok: true},
		{name: "objectSpreadUnknown"},
		{name: "objectSpreadGetter"},
		{name: "stringSlice", want: "GET", ok: true},
		{name: "stringSliceDefault", want: "GET", ok: true},
		{name: "stringSliceUtf16", want: "GET", ok: true},
		{name: "stringSubstring", want: "HEAD", ok: true},
		{name: "stringSubstringReversed", want: "HEAD", ok: true},
		{name: "stringSubstr", want: "GET", ok: true},
		{name: "stringSubstrNegative", want: "GET", ok: true},
		{name: "stringSubstrEmptyLength", want: "", ok: true},
		{name: "stringCharAt", want: "G", ok: true},
		{name: "stringCharAtDefault", want: "G", ok: true},
		{name: "stringCharAtOutOfRange", want: "", ok: true},
		{name: "stringConcat", want: "GET", ok: true},
		{name: "stringConcatMany", want: "HEAD", ok: true},
		{name: "stringConcatCoercion", want: "n1truenull", ok: true},
		{name: "stringConcatUnknown"},
		{name: "unaryPlusString", want: "71", ok: true},
		{name: "unaryPlusBoolean", want: "1", ok: true},
		{name: "unaryMinusString", want: "-1", ok: true},
		{name: "unaryTildeString", want: "-2", ok: true},
		{name: "unaryPlusInvalid", want: "NaN", ok: true},
		{name: "unaryPlusUnknown"},
		{name: "strictEqualNumbers", want: "then", ok: true},
		{name: "strictEqualStrings", want: "then", ok: true},
		{name: "strictUnequalStrings", want: "then", ok: true},
		{name: "strictEqualAcrossKinds", want: "then", ok: true},
		{name: "strictEqualNullish", want: "then", ok: true},
		{name: "strictEqualObjects"},
		{name: "strictEqualUnknown"},
		{name: "aliasFromCharCode", want: "GET", ok: true},
		{name: "aliasArrayOf", want: "GET", ok: true},
		{name: "shadowedFromCharCode"},
		{name: "shadowedArrayOf"},
		{name: "stringRaw", want: "then", ok: true},
		{name: "stringRawSubstitution", want: "then", ok: true},
		{name: "stringRawEscapes", want: `\p{Script=Kawi}\n`, ok: true},
		{name: "stringRawEscapedSubstitution", want: `\p{Script=Kawi}\n`, ok: true},
		{name: "stringRawEmptyParts", want: "\n1", ok: true},
		{name: "stringRawNewlines", want: "a\nb\nc", ok: true},
		{name: "stringRawInvalidEscape", want: `\u{110000}`, ok: true},
		{name: "stringRawAlias", want: "then", ok: true},
		{name: "stringRawMutableAlias"},
		{name: "stringRawTypedAlias"},
		{name: "shadowedStringCall"},
		{name: "shadowedStringRaw"},
		{name: "letStatic", want: "then", ok: true},
		{name: "varStatic", want: "then", ok: true},
		{name: "letStaticUse", want: "then", ok: true},
		{name: "varStaticUse", want: "then", ok: true},
		{name: "letWritten", want: "then", ok: true},
		{name: "letWrittenUse"},
		{name: "varWritten", want: "then", ok: true},
		{name: "varWrittenUse"},
		{name: "destructuredWritten", want: "then", ok: true},
		{name: "destructuredWrittenUse"},
		{name: "notThen", want: "not-then", ok: true},
		{name: "cycle"},
		{name: "letUse", want: "then", ok: true},
		{name: "numeric"},
		{name: "unknownUse"},
		{name: "stableArrayUse", want: "", ok: true},
		{name: "writtenArrayUse"},
		{name: "filledArrayUse"},
		{name: "computedMutatedArrayUse"},
		{name: "updatedArrayUse"},
		{name: "slicedArrayUse", want: "", ok: true},
		{name: "writtenObjectUse"},
		{name: "protoMessage", want: "message", ok: true},
		{name: "quotedProtoMessage", want: "message", ok: true},
		{name: "ownProtoMessage", want: "", ok: true},
		{name: "computedProtoMessage"},
		{name: "nullProtoString"},
		{name: "regexSource", want: "g", ok: true},
		{name: "regexComputedSource", want: "gi", ok: true},
		{name: "regexFlags", want: "im", ok: true},
		{name: "regexGlobal", want: "true", ok: true},
		{name: "regexHasIndices", want: "true", ok: true},
		{name: "regexUnicode", want: "true", ok: true},
		{name: "regexLastIndex", want: "0", ok: true},
		{name: "regexString", want: "/x/im", ok: true},
		{name: "regexAliasSource", want: "alias", ok: true},
		{name: "mutatedRegexSource"},
		{name: "regexUnicodeSets"},
		{name: "regexProtoSource"},
		{name: "regexProtoLastIndex", want: "0", ok: true},
		{name: "regexProtoOwnSource", want: "own", ok: true},
		{name: "regexProtoString"},
		{name: "emptyJoin", want: "", ok: true},
		{name: "joined", want: "error: message", ok: true},
		{name: "nestedEmptyJoin", want: "", ok: true},
		{name: "overflowEmptyJoin", want: "", ok: true},
		{name: "undefinedSeparatorJoin", want: "", ok: true},
		{name: "boundJoin", want: "", ok: true},
		{name: "mutatedJoin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := staticEvaluator.Eval(findVariableInitializer(t, sourceFile, tt.name))
			if got != tt.want || ok != tt.ok {
				t.Fatalf("Eval(%s) = (%q, %v), want (%q, %v)", tt.name, got, ok, tt.want, tt.ok)
			}
		})
	}

	if isArray, known := staticEvaluator.EvalArrayValue(findVariableInitializer(t, sourceFile, "stableArray")); !known || !isArray {
		t.Fatalf("EvalArrayValue(stableArray) = (%v, %v), want (true, true)", isArray, known)
	}
	if isArray, known := staticEvaluator.EvalArrayValue(findVariableInitializer(t, sourceFile, "stableObject")); !known || isArray {
		t.Fatalf("EvalArrayValue(stableObject) = (%v, %v), want (false, true)", isArray, known)
	}
	if isArray, known := staticEvaluator.EvalArrayValue(findVariableInitializer(t, sourceFile, "unknownUse")); known || isArray {
		t.Fatalf("EvalArrayValue(unknownUse) = (%v, %v), want (false, false)", isArray, known)
	}

	if _, ok := staticEvaluator.EvalControlFlowValue(findVariableInitializer(t, sourceFile, "controlConstUse")); !ok {
		t.Fatal("EvalControlFlowValue(const) did not resolve")
	}
	for _, name := range []string{"controlLetUse", "controlVarUse", "controlObjectUse"} {
		if _, ok := staticEvaluator.EvalControlFlowValue(findVariableInitializer(t, sourceFile, name)); ok {
			t.Fatalf("EvalControlFlowValue(%s) unexpectedly resolved", name)
		}
	}
	if isArray, known := staticEvaluator.EvalControlFlowArrayValue(findVariableInitializer(t, sourceFile, "controlArrayUse")); !known || !isArray {
		t.Fatalf("EvalControlFlowArrayValue(controlArrayUse) = (%v, %v), want (true, true)", isArray, known)
	}
}

type staticEvaluatorTestResolver struct{ resolver binder.NameResolver }

func TestStaticStringEvaluatorRegExp(t *testing.T) {
	root := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(root.Dir, "regexp.ts")
	code := `
		const literal = /a/gi;
		const alias = literal;
		let stable = new RegExp("x", "y");
		const stableUse = stable;
		let changed = /x/;
		changed = /y/g;
		const changedUse = changed;
		const mutatedLiteral = /x/;
		mutatedLiteral[Symbol.match] = () => null;
		const mutatedLiteralUse = mutatedLiteral;
		const aliasedLiteral = /x/;
		const mutationAlias = aliasedLiteral;
		mutationAlias[Symbol.match] = () => null;
		const mutatedAliasUse = aliasedLiteral;
		let mutatedConstructor = new RegExp("x");
		Reflect.set(mutatedConstructor, "test", () => false);
		const mutatedConstructorUse = mutatedConstructor;
		const constructor = new RegExp("a" + "b", "u");
		const copied = new RegExp(/a/gi);
		const overridden = new RegExp(/a/g, "y");
		const empty = new RegExp();
		const primitive = new RegExp(null);
		const unicodeSets = new RegExp("[a&&b]", "v");
		const invalidPattern = new RegExp("[", "u");
		const duplicateFlags = new RegExp("a", "gg");
		const invalidFlags = new RegExp("a", "z");
		const incompatibleFlags = new RegExp("a", "uv");
		const coercion = new RegExp({toString() { effect(); return "a"; }});
		const array = new RegExp(["a"]);
		const spread = new RegExp(...["a"]);
		const effect = new RegExp(read());
		const object = {pattern: /a/};
		Object.defineProperty(object, "pattern", {get() { return "a"; }});
		const mutableMember = object.pattern;
		const memberAlias = mutableMember || /b/;
		const called = parseInt("1");
		const calledAlias = called || /b/;
		const missingMember = object.missing;
		const literalMember = ({pattern: /a/}).pattern;
		const cycle = cycle;
		const textValue = "a";
		const numberValue = 1;
		const nullValue = null;
		const undefinedValue = undefined;
		const symbolValue = Symbol.iterator;
		const arrayValue = [];
		const objectValue = {};
		{ const RegExp = Custom; const shadowed = new RegExp("a"); }
	`
	fs := NewOverlayVFS(root.FS, map[string]string{filePath: code})
	program, err := CreateProgram(true, fs, root.Dir, "tsconfig.json", CreateCompilerHost(root.Dir, fs))
	assert.NilError(t, err)
	source := program.GetSourceFile(filePath)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()
	evaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, source)
	for _, test := range []struct {
		name, pattern, flags string
		known                bool
	}{
		{"literal", "a", "gi", true},
		{"alias", "a", "gi", true},
		{"stableUse", "x", "y", true},
		{"constructor", "ab", "u", true},
		{"copied", "a", "gi", true},
		{"overridden", "a", "y", true},
		{"empty", "", "", true},
		{"primitive", "null", "", true},
		{"unicodeSets", "[a&&b]", "v", true},
		{"changedUse", "", "", false},
		{"mutatedLiteralUse", "", "", false},
		{"mutatedAliasUse", "", "", false},
		{"mutatedConstructorUse", "", "", false},
		{"invalidPattern", "", "", false},
		{"duplicateFlags", "", "", false},
		{"invalidFlags", "", "", false},
		{"incompatibleFlags", "", "", false},
		{"coercion", "", "", false},
		{"array", "", "", false},
		{"spread", "", "", false},
		{"effect", "", "", false},
		{"mutableMember", "", "", false},
		{"literalMember", "a", "", true},
		{"cycle", "", "", false},
		{"shadowed", "", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			pattern, flags, known := evaluator.EvalRegExp(findVariableInitializer(t, source, test.name))
			if known != test.known || pattern != test.pattern || flags != test.flags {
				t.Fatalf("EvalRegExp = (%q, %q, %v), want (%q, %q, %v)", pattern, flags, known, test.pattern, test.flags, test.known)
			}
		})
	}
	for _, test := range []struct {
		name  string
		known bool
	}{
		{"mutableMember", true},
		{"memberAlias", true},
		{"calledAlias", false},
		{"missingMember", false},
		{"literalMember", false},
		{"changedUse", false},
		{"cycle", false},
	} {
		node := findVariableInitializer(t, source, test.name)
		if got := evaluator.HasNonUndefinedValueIgnoringPropertyMutations(node); got != test.known {
			t.Errorf("HasNonUndefinedValueIgnoringPropertyMutations(%s) = %v, want %v", test.name, got, test.known)
		}
	}
	for _, test := range []struct {
		name      string
		primitive bool
	}{
		{"textValue", true}, {"numberValue", true}, {"nullValue", true},
		{"undefinedValue", true}, {"symbolValue", true},
		{"arrayValue", false}, {"objectValue", false}, {"literal", false},
	} {
		value, known := evaluator.EvalValue(findVariableInitializer(t, source, test.name))
		if !known || IsPrimitiveStaticValue(value) != test.primitive {
			t.Errorf("IsPrimitiveStaticValue(%s): known = %v, want primitive = %v", test.name, known, test.primitive)
		}
	}
}

func (r *staticEvaluatorTestResolver) Resolve(node *ast.Node) *ast.Symbol {
	return r.resolver.Resolve(node, node.Text(), ast.SymbolFlagsValue|ast.SymbolFlagsAlias, nil, true, false)
}

func TestStaticStringEvaluatorMutationAnalysisOrder(t *testing.T) {
	for _, test := range []struct {
		name         string
		declarations string
		method       string
	}{
		{"array", `const method = ["reverse"];`, `method[0]`},
		{"object", `const method = {name: "reverse"};`, `method.name`},
		{"alias", `const names = ["reverse"]; const method = names[0];`, `method`},
		{"nested", `const names = ["reverse"]; const method = {name: names[0]};`, `method.name`},
		{"arrayCoercion", `const method = ["reverse"];`, `method`},
		{"callInitializer", `const method = Array.of("reverse");`, `method[0]`},
		{"letBinding", `let method = ["reverse"];`, `method[0]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := fixtures.GetRootDir()
			filePath := tspath.ResolvePath(root.Dir, "file.ts")
			code := test.declarations + `
const values = ["", "ok"];
const untouched = ["stable"];
const cycleA = [cycleB[0]];
const cycleB = [cycleA[0]];
const uncertain = ["ignored"];
uncertain[cycleA[0]]();
const selfMutating = ["reverse", "pop"];
selfMutating[selfMutating[0]]();
const prime = ` + test.method + `;
values[` + test.method + `]();
const result = values[0];
const stable = untouched[0];
const cyclic = cycleA[0];
const selfRead = selfMutating[0];`
			fs := NewOverlayVFS(root.FS, map[string]string{filePath: code})
			program, err := CreateProgram(true, fs, root.Dir, "tsconfig.json", CreateCompilerHost(root.Dir, fs))
			assert.NilError(t, err)
			sourceFile := program.GetSourceFile(filePath)
			assert.Assert(t, sourceFile != nil)
			typeChecker, done := program.GetTypeChecker(t.Context())
			defer done()
			for _, resolution := range []string{"checker", "references"} {
				for _, first := range []string{"prime", "result", "stable", "cyclic", "selfRead"} {
					t.Run(resolution+"/"+first, func(t *testing.T) {
						staticEvaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
						if resolution == "references" {
							resolver := &staticEvaluatorTestResolver{resolver: binder.NameResolver{Globals: sourceFile.Locals}}
							staticEvaluator = NewStaticStringEvaluatorWithReferenceResolver(nil, sourceFile, resolver)
						}
						staticEvaluator.EvalValue(findVariableInitializer(t, sourceFile, first))
						// Reuse the instance after both successful and failed evaluations.
						for range 2 {
							if value, known := staticEvaluator.Eval(findVariableInitializer(t, sourceFile, "result")); known {
								t.Fatalf("mutated array folded to %q after evaluating %s first", value, first)
							}
							if value, known := staticEvaluator.EvalToString(findVariableInitializer(t, sourceFile, "prime")); !known || value != "reverse" {
								t.Fatalf("method = (%q, %v), want (reverse, true)", value, known)
							}
							if value, known := staticEvaluator.Eval(findVariableInitializer(t, sourceFile, "stable")); !known || value != "stable" {
								t.Fatalf("unmodified array = (%q, %v), want (stable, true)", value, known)
							}
							if value, known := staticEvaluator.EvalValue(findVariableInitializer(t, sourceFile, "cyclic")); known {
								t.Fatalf("cyclic initializer folded to %v", value)
							}
							if value, known := staticEvaluator.Eval(findVariableInitializer(t, sourceFile, "selfRead")); known {
								t.Fatalf("self-modifying array folded to %q", value)
							}
						}
					})
				}
			}
		})
	}

	t.Run("alias mutations", func(t *testing.T) {
		for _, test := range []struct {
			name, mutation string
			known          bool
		}{
			{"iterator alias", `const alias = values; alias[Symbol.iterator] = function*() { yield "changed"; };`, false},
			{"alias chain", `const alias = values; const second = alias; const third = second; third[0] = "changed";`, false},
			{"conditional alias", `const alias = true ? values : []; alias[Symbol.iterator] = function*() { yield "changed"; };`, false},
			{"conditional mutation target", `(true ? values : [])[Symbol.iterator] = function*() { yield "changed"; };`, false},
			{"literal container mutation target", `({values}).values[Symbol.iterator] = function*() { yield "changed"; };`, false},
			{"conditional mutator receiver", `(true ? values : []).reverse();`, false},
			{"literal container mutator receiver", `({values}).values.reverse();`, false},
			{"conditional computed mutator receiver", `const method = "fill"; (true ? values : [])[method]("changed");`, false},
			{"conditional alternate alias", `const alias = false ? [] : values; alias[0] = "changed";`, false},
			{"logical alias", `const alias = values || []; alias[0] = "changed";`, false},
			{"logical right alias", `const alias = true && values; alias[0] = "changed";`, false},
			{"nullish alias", `const alias = null ?? values; alias[0] = "changed";`, false},
			{"comma alias", `const alias = (0, values); alias[0] = "changed";`, false},
			{"assignment alias", `let other; const alias = (other = values); alias[0] = "changed";`, false},
			{"object container", `const holder = {values}; holder.values[Symbol.iterator] = function*() { yield "changed"; };`, false},
			{"object binding alias", `const holder = {values}; const {values: alias} = holder; alias[0] = "changed";`, false},
			{"array binding alias", `const [alias] = [values]; alias[0] = "changed";`, false},
			{"nested binding alias", `const {items: [alias]} = {items: [values]}; alias[0] = "changed";`, false},
			{"binding default alias", `const {alias = values} = {}; alias[0] = "changed";`, false},
			{"outer binding default alias", `const {items: {alias} = {alias: values}} = {}; alias[0] = "changed";`, false},
			{"object rest nested mutation", `const {...copy} = {values}; copy.values[0] = "changed";`, false},
			{"array rest nested mutation", `const [...copy] = [values]; copy[0][0] = "changed";`, false},
			{"object rest copy", `const {...copy} = values; copy[0] = "changed";`, true},
			{"array rest copy", `const [...copy] = values; copy[0] = "changed";`, true},
			{"nested array rest copy", `const [[...copy]] = [values]; copy[0] = "changed";`, true},
			{"binding replacement", `let [alias] = [values]; alias = [];`, true},
			{"assign container", `const holder = Object.assign({}, {values}); holder.values[0] = "changed";`, false},
			{"assign multiple sources", `const holder = Object.assign({}, {other: []}, {values}); holder.values[0] = "changed";`, false},
			{"assign spread arguments", `const holder = Object.assign(...[{}, {values}]); holder.values[0] = "changed";`, false},
			{"assign existing target", `const holder = Object.assign(values, {}); holder[0] = "changed";`, false},
			{"assign target after empty spread", `const holder = Object.assign(...[], values); holder[0] = "changed";`, false},
			{"assign nested target", `const target = {values}; const holder = Object.assign(target, {}); holder.values[0] = "changed";`, false},
			{"assign property replacement", `const holder = Object.assign({}, {values}); holder.values = [];`, true},
			{"assign array copy", `const copy = Object.assign([], values); copy[0] = "changed";`, true},
			{"assign object copy", `const copy = Object.assign({}, values); copy[0] = "changed";`, true},
			{"assign copies spread sources", `const copy = Object.assign({}, ...[values]); copy[0] = "changed";`, true},
			{"renamed object container", `const holder = {items: values}; holder.items.reverse();`, false},
			{"array container", `const holder = [values]; holder[0].reverse();`, false},
			{"nested container alias", `const holder = {items: [values]}; const alias = holder.items[0]; alias[0] = "changed";`, false},
			{"nested object container", `const holder = {outer: {inner: values}}; holder.outer.inner[0] = "changed";`, false},
			{"frozen container", `const holder = Object.freeze({values}); holder.values[0] = "changed";`, false},
			{"array factory container", `const holder = Array.of(values); holder[0][0] = "changed";`, false},
			{"computed array factory container", `const method = "of"; const holder = Array[method](values); holder[0][0] = "changed";`, false},
			{"aliased array factory container", `const ArrayAlias = Array; const holder = ArrayAlias.of(values); holder[0][0] = "changed";`, false},
			{"array factory spread container", `const holder = Array.of(...[values]); holder[0][0] = "changed";`, false},
			{"container mutation after replacement", `const holder = {values}; holder.other = []; holder.values[0] = "changed";`, false},
			{"object spread container", `const holder = {values}; const copy = {...holder}; copy.values[0] = "changed";`, false},
			{"array spread container", `const holder = [values]; const copy = [...holder]; copy[0][0] = "changed";`, false},
			{"nested property definition", `const holder = {values}; Object.defineProperty(holder.values, "0", {value: "changed"});`, false},
			{"nested explicit receiver", `const holder = {values}; Reflect.set({}, "0", "changed", holder.values);`, false},
			{"pass-through alias", `const alias = Object.preventExtensions(values); alias[0] = "changed";`, false},
			{"array copy", `const alias = [...values]; alias[0] = "changed";`, true},
			{"array factory copy", `const alias = Array.of(...values); alias[0] = "changed";`, true},
			{"object copy", `const alias = {...values}; alias[0] = "changed";`, true},
			{"container property replacement", `const holder = {values}; holder.values = [];`, true},
			{"conditional test is not an alias", `const alias = values ? [] : []; alias[0] = "changed";`, true},
			{"comma left is not an alias", `const alias = (values, []); alias[0] = "changed";`, true},
			{"computed key is not an alias", `const holder = {[values]: []}; holder.stable[0] = "changed";`, true},
			{"computed index is not an alias", `const holder = {stable: []}; const alias = holder[values]; alias[0] = "changed";`, true},
			{"method body is not an alias", `const holder = {method() { return values; }}; holder.method.other = "changed";`, true},
			{"unmodified indirect alias", `const holder = {values}; const alias = true ? holder.values : []; alias[0];`, true},
			{"circular container aliases", `const first = {second}; const second = {first}; first.second.other = "changed";`, true},
			{"circular container with reachable alias", `const first = {second}; const second = {first, values}; first.second.values[0] = "changed";`, false},
			{"circular access aliases", `const first = second.value; const second = first.value; first.other = "changed";`, true},
			{"alias mutator", `const alias = values; alias.reverse();`, false},
			{"computed alias mutator", `const alias = values; const method = "fill"; alias[method]("changed");`, false},
			{"mutable alias", `let alias = values; alias[0] = "changed"; alias = [];`, false},
			{"mutable alias conservatism", `let alias = values; alias = []; alias[0] = "changed";`, false},
			{"alias replacement", `let alias = values; alias = [];`, true},
			{"unmodified alias", `const alias = values; const second = alias; second[0];`, true},
			{"circular aliases", `const first = second; const second = first; first[0] = "changed";`, true},
			{"self alias", `const alias = alias; alias[0] = "changed";`, true},
			{"getter definition", `Object.defineProperty(values, "0", {get() { return unknown; }});`, false},
			{"getter definition through alias", `const alias = values; Object.defineProperty(alias, "0", {get() { return unknown; }});`, false},
			{"multiple property definitions", `Object.defineProperties(values, {0: {get() { return unknown; }}});`, false},
			{"assign", `Object.assign(values, {0: "changed"});`, false},
			{"assign expanded arguments", `Object.assign(...[values, {0: "changed"}]);`, false},
			{"prototype mutation", `Object.setPrototypeOf(values, {[Symbol.iterator]() { return unknown; }});`, false},
			{"reflect definition", `Reflect.defineProperty(values, "0", {get() { return unknown; }});`, false},
			{"reflect assignment", `Reflect.set(values, "0", "changed");`, false},
			{"reflect explicit receiver", `Reflect.set({}, Symbol.iterator, function*() { yield "changed"; }, values);`, false},
			{"reflect deletion", `Reflect.deleteProperty(values, "0");`, false},
			{"shadowed property definition", `{ const Object = {defineProperty() {}}; Object.defineProperty(values, "0", {}); }`, true},
		} {
			t.Run(test.name, func(t *testing.T) {
				root := fixtures.GetRootDir()
				filePath := tspath.ResolvePath(root.Dir, "alias-mutation.ts")
				code := `const values = ["stable"]; ` + test.mutation + ` const result = [...values].join(""); const spread = [...values];`
				fs := NewOverlayVFS(root.FS, map[string]string{filePath: code})
				program, err := CreateProgram(true, fs, root.Dir, "tsconfig.json", CreateCompilerHost(root.Dir, fs))
				assert.NilError(t, err)
				sourceFile := program.GetSourceFile(filePath)
				assert.Assert(t, sourceFile != nil)
				typeChecker, done := program.GetTypeChecker(t.Context())
				defer done()
				for _, resolution := range []string{"checker", "references"} {
					t.Run(resolution, func(t *testing.T) {
						staticEvaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
						if resolution == "references" {
							resolver := &staticEvaluatorTestResolver{resolver: binder.NameResolver{Globals: sourceFile.Locals}}
							staticEvaluator = NewStaticStringEvaluatorWithReferenceResolver(nil, sourceFile, resolver)
						}
						value, known := staticEvaluator.Eval(findVariableInitializer(t, sourceFile, "result"))
						if known != test.known || known && value != "stable" {
							t.Fatalf("spread result = (%q, %v), want known = %v", value, known, test.known)
						}
						if _, known := staticEvaluator.EvalValueIfNoSideEffects(findVariableInitializer(t, sourceFile, "spread")); known != test.known {
							t.Fatalf("spread safety = %v, want %v", known, test.known)
						}
					})
				}
			})
		}
	})
}

func TestStaticStringEvaluatorCallEscapes(t *testing.T) {
	for _, test := range []struct {
		name, operation, expression string
		known                       bool
	}{
		{"captured prototype reference", "const holder = {__proto__: methods}; unknown(methods);", "holder.current", false},
		{"captured quoted prototype reference", "const holder = {\"__proto__\": methods}; unknown(methods);", "holder.current", false},
		{"captured prototype alias", "const proto = methods; const holder = {__proto__: proto}; unknown(methods);", "holder.current", false},
		{"captured nested prototype", "const holder = {__proto__: {__proto__: methods}}; unknown(methods);", "holder.current", false},
		{"stable prototype reference", "const holder = {__proto__: methods};", "holder.current", true},
		{"primitive copied from prototype", "const holder = {__proto__: methods}; const key = holder.current; unknown(methods);", "key", true},
		{"null prototype capture", "const holder = {__proto__: null, current: methods.current}; unknown(methods);", "holder.current", true},
		{"ignored primitive prototype", "const holder = {__proto__: 1, current: methods.current}; unknown(methods);", "holder.current", true},
		{"own property shadows prototype", "const holder = {__proto__: methods, current: \"catch\"}; unknown(methods);", "holder.current", true},
		{"object spread omits prototype", "const holder = {__proto__: methods, current: \"catch\"}; unknown({...holder});", "methods.current", true},
		{"computed proto is own reference", "const holder = {[\"__proto__\"]: methods}; unknown(methods);", "holder.__proto__.current", false},
		{"nested array capture after iterator mutation", "Array.prototype[Symbol.iterator] = function*() { yield \"then\"; }; const copy = [...[\"catch\"]]; const key = copy[0];", "key", false},
		{"nested string capture after iterator mutation", "String.prototype[Symbol.iterator] = function*() { yield \"then\"; }; const copy = [...\"catch\"]; const key = copy;", "key.join(\"\")", false},
		{"nested array capture with ordinary iterator", "const copy = [...[\"catch\"]]; const key = copy[0];", "key", true},
		{"nested string capture with ordinary iterator", "const copy = [...\"catch\"]; const key = copy;", "key.join(\"\")", true},
		{"array capture with unrelated string iterator mutation", "String.prototype[Symbol.iterator] = function*() { yield \"then\"; }; const copy = [...[\"catch\"]]; const key = copy[0];", "key", true},
		{"binding transfer in argument", "let holder = {}; unknown(holder, (holder = methods, 0));", "methods.current", false},
		{"property transfer in argument", "const holder = {}; unknown(holder, (holder.item = methods, 0));", "methods.current", false},
		{"assign transfer in argument", "const holder = {}; unknown(holder, (Object.assign(holder, {methods}), 0));", "methods.current", false},
		{"define property transfer in argument", "const holder = {}; unknown(holder, (Object.defineProperty(holder, \"item\", {value: methods}), 0));", "methods.current", false},
		{"define properties transfer in argument", "const holder = {}; unknown(holder, (Object.defineProperties(holder, {item: {value: methods}}), 0));", "methods.current", false},
		{"reflect set transfer in argument", "const holder = {}; unknown(holder, (Reflect.set(holder, \"item\", methods), 0));", "methods.current", false},
		{"reflect define transfer in argument", "const holder = {}; unknown(holder, (Reflect.defineProperty(holder, \"item\", {value: methods}), 0));", "methods.current", false},
		{"builtin alias transfer in argument", "const holder = {}; const assign = Object.assign; unknown(holder, (assign(holder, {methods}), 0));", "methods.current", false},
		{"computed builtin transfer in argument", "const holder = {}; const key = \"assign\"; unknown(holder, (Object[key](holder, {methods}), 0));", "methods.current", false},
		{"transfer in constructor argument", "const holder = {}; new callable(holder, (Object.assign(holder, {methods}), 0));", "methods.current", false},
		{"transfer in template substitution", "const holder = {}; callable`${holder}${(Object.assign(holder, {methods}), 0)}`;", "methods.current", false},
		{"transfer in callee expression", "const holder = {}; (Object.assign(holder, {methods}), unknown)(holder);", "methods.current", false},
		{"transfer in optional call argument", "const holder = {}; callable?.(holder, (Object.assign(holder, {methods}), 0));", "methods.current", false},
		{"nested transfer under argument", "const holder = {}; unknown(holder, callable((Object.assign(holder, {methods}), 0)));", "methods.current", false},
		{"primitive assign in argument", "const holder = {}; unknown(holder, (Object.assign(holder, methods), 0));", "methods.current", true},
		{"transfer in following argument after earlier call", "const holder = {}; function noop(...args: any[]) {} noop(unknown(holder), (Object.assign(holder, {methods}), 0));", "methods.current", true},
		{"transfer after invocation in same expression", "const holder = {}; (unknown(holder), Object.assign(holder, {methods}));", "methods.current", true},
		{"transfer before invocation in same expression", "const holder = {}; (Object.assign(holder, {methods}), unknown(holder));", "methods.current", false},
		{"set prototype transfer", `const holder = {}; Object.setPrototypeOf(holder, {methods}); unknown(holder);`, `methods.current`, false},
		{"reflect prototype transfer", `const holder = {}; Reflect.setPrototypeOf(holder, {methods}); unknown(holder);`, `methods.current`, false},
		{"prototype transfer in argument", `const holder = {}; unknown(holder, (Object.setPrototypeOf(holder, {methods}), 0));`, `methods.current`, false},
		{"prototype transfer after call", `const holder = {}; unknown(holder); Object.setPrototypeOf(holder, {methods});`, `methods.current`, true},
		{"own captured property shadows prototype", `const holder = {__proto__: methods, current: methods.current}; unknown(methods);`, `holder.current`, true},
		{"own captured property through alias", `const holder = {__proto__: methods, current: methods.current}; const alias = holder; unknown(methods);`, `alias.current`, true},
		{"own captured property mutated through alias", `const holder = {__proto__: methods, current: methods.current}; const alias = holder; unknown(alias);`, `holder.current`, false},
		{"own captured property with spread", `const holder = {__proto__: methods, ...{current: methods.current}}; unknown(methods);`, `holder.current`, true},
		{"parameter mutation", `function change(holder: typeof methods) { holder.current = "then"; } change(methods);`, `methods.current`, false},
		{"unknown function", `unknown(methods);`, `methods.current`, false},
		{"alias", `const alias = methods; unknown(alias);`, `methods.current`, false},
		{"later binding transfer", `let alias = {current: "then"}; alias = methods; unknown(alias);`, `methods.current`, false},
		{"conditional binding transfer", `let alias = {current: "then"}; if (condition) alias = methods; unknown(alias);`, `methods.current`, false},
		{"binding transfer after call", `let alias = {current: "then"}; unknown(alias); alias = methods;`, `methods.current`, true},
		{"overwritten binding transfer", `let alias = {current: "then"}; alias = methods; alias = {current: "then"}; unknown(alias);`, `methods.current`, true},
		{"captured primitive", `const key = methods.current; unknown(methods);`, `key`, true},
		{"captured primitive alias", `const key = methods.current; unknown(methods); const second = key;`, `second`, true},
		{"captured primitive derived value", `const key = methods.current; unknown(methods); const second = key + "";`, `second`, true},
		{"primitive captured in container", `const holder = {key: methods.current}; unknown(methods);`, `holder.key`, true},
		{"reference retained in container", `const holder = {methods}; unknown(methods);`, `holder.methods.current`, false},
		{"capture after mutation", `unknown(methods); const key = methods.current;`, `key`, false},
		{"assigned container transfer", `const holder = {}; Object.assign(holder, {methods}); unknown(holder);`, `methods.current`, false},
		{"reassigned holder transfer", `const holder = {}; let alias = {}; alias = holder; alias.item = methods; unknown(holder);`, `methods.current`, false},
		{"holder transfer after unrelated call", `const holder = {}; let alias = {}; unknown(holder); alias = holder; alias.item = methods;`, `methods.current`, true},
		{"defined container transfer", `const holder = {}; Object.defineProperty(holder, "item", {value: methods}); unknown(holder);`, `methods.current`, false},
		{"reflect container transfer", `const holder = {}; Reflect.defineProperty(holder, "item", {value: methods}); unknown(holder);`, `methods.current`, false},
		{"copied primitive container", `const holder = {}; Object.assign(holder, methods); unknown(holder);`, `methods.current`, true},
		{"object wrapper", `unknown({methods});`, `methods.current`, false},
		{"array wrapper", `unknown([methods]);`, `methods.current`, false},
		{"nested wrapper", `unknown({nested: [methods]});`, `methods.current`, false},
		{"destructured alias", `const {item: alias} = {item: methods}; unknown(alias);`, `methods.current`, false},
		{"conditional argument", `unknown(condition ? methods : {});`, `methods.current`, false},
		{"logical argument", `unknown(methods || {});`, `methods.current`, false},
		{"sequence argument", `unknown((0, methods));`, `methods.current`, false},
		{"method argument", `callable.change(methods);`, `methods.current`, false},
		{"method receiver", `methods.change();`, `methods.current`, false},
		{"call receiver", `callable.call(methods);`, `methods.current`, false},
		{"apply arguments", `callable.apply(null, [methods]);`, `methods.current`, false},
		{"bound argument", `callable.bind(null, methods);`, `methods.current`, false},
		{"constructor", `new callable(methods);`, `methods.current`, false},
		{"spread arguments", `unknown(...[methods]);`, `methods.current`, false},
		{"optional call", `callable?.(methods);`, `methods.current`, false},
		{"tagged template", "callable`key ${methods}`;", `methods.current`, false},
		{"array argument", `unknown(names);`, `names[0]`, false},
		{"shallow freeze", `unknown(Object.freeze({methods}));`, `methods.current`, false},
		{"object rest nested reference", `const {...copy} = {methods}; unknown(copy);`, `methods.current`, false},
		{"spread nested reference", `unknown({...{methods}});`, `methods.current`, false},
		{"factory nested reference", `unknown(Array.of(methods));`, `methods.current`, false},
		{"assign nested reference", `unknown(Object.assign({}, {methods}));`, `methods.current`, false},
		{"returned closure reference", `unknown(() => methods);`, `methods.current`, false},
		{"function declaration reference", `function get() { return methods; } unknown(get);`, `methods.current`, false},
		{"nested returned callback", `unknown(() => () => methods);`, `methods.current`, false},
		{"stored after initializer", `const holder = {}; holder.item = methods; unknown(holder);`, `methods.current`, false},
		{"stored through holder alias", `const holder = {}; const alias = holder; alias.item = methods; unknown(holder);`, `methods.current`, false},
		{"stored through destructured alias", `const holder = {}; const {item: alias} = {item: holder}; alias.item = methods; unknown(holder);`, `methods.current`, false},
		{"stored through default alias", `const holder = {}; const {alias = holder} = {}; alias.item = methods; unknown(holder);`, `methods.current`, false},
		{"stored through nested default alias", `const holder = {}; const {outer: {alias} = {alias: holder}} = {}; alias.item = methods; unknown(holder);`, `methods.current`, false},
		{"stored through nested holder", `const holder = {nested: {}}; holder.nested.item = methods; unknown(holder);`, `methods.current`, false},
		{"stored through referenced member", `const other = {current: "catch"}; const holder = {nested: other}; holder.nested.item = methods; unknown(other);`, `methods.current`, false},
		{"callback publishes reference", `unknown((holder: any) => { holder.item = methods; });`, `methods.current`, false},
		{"callback throws reference", `unknown(() => { throw methods; });`, `methods.current`, false},
		{"generator yields reference", `unknown(function*() { yield methods; });`, `methods.current`, false},
		{"class returns reference", `class Factory { constructor() { return methods; } } unknown(new Factory());`, `methods.current`, false},
		{"getter exposes reference", `unknown({get item() { return methods; }});`, `methods.current`, false},
		{"primitive binding", `let key = "catch"; unknown(key);`, `key`, true},
		{"primitive member argument", `unknown(methods.current);`, `methods.current`, true},
		{"primitive callback result", `unknown(() => methods.current);`, `methods.current`, true},
		{"unused nested function", `unknown(() => { function unused() { return methods; } });`, `methods.current`, true},
		{"primitive object spread", `unknown({...methods});`, `methods.current`, true},
		{"primitive array spread", `unknown([...names]);`, `names[0]`, true},
		{"primitive object rest", `const {...copy} = methods; unknown(copy);`, `methods.current`, true},
		{"primitive array rest", `const [...copy] = names; unknown(copy);`, `names[0]`, true},
		{"primitive array factory copy", `unknown(Array.of(...names));`, `names[0]`, true},
		{"primitive assign copy", `unknown(Object.assign({}, methods));`, `methods.current`, true},
		{"private stored reference", `const holder = {}; holder.item = methods;`, `methods.current`, true},
		{"store into fresh rest copy", `const holder = {}; const {...copy} = holder; copy.item = methods; unknown(holder);`, `methods.current`, true},
		{"string conversion", `const key = String(names);`, `key`, true},
		{"array join", `const key = names.join("");`, `key`, true},
		{"empty function", `function noop(value: unknown) {} noop(methods);`, `methods.current`, true},
		{"empty method", `const handler = {noop(value: unknown) {}}; handler.noop(methods);`, `methods.current`, true},
		{"overwritten empty function", `let noop = (value: unknown) => {}; noop = callable; noop(methods);`, `methods.current`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := fixtures.GetRootDir()
			filePath := tspath.ResolvePath(root.Dir, "call-escape.ts")
			code := `const methods = {current: "catch"}; const names = ["catch"];
declare function unknown(...args: any[]): any;
declare const callable: any;
declare const condition: boolean;
` + test.operation + ` const result = ` + test.expression + `;`
			fs := NewOverlayVFS(root.FS, map[string]string{filePath: code})
			program, err := CreateProgram(true, fs, root.Dir, "tsconfig.json", CreateCompilerHost(root.Dir, fs))
			assert.NilError(t, err)
			sourceFile := program.GetSourceFile(filePath)
			assert.Assert(t, sourceFile != nil)
			typeChecker, done := program.GetTypeChecker(t.Context())
			defer done()
			for _, resolution := range []string{"checker", "references"} {
				t.Run(resolution, func(t *testing.T) {
					evaluator := NewStaticStringEvaluatorWithCallEscapes(typeChecker, sourceFile, nil)
					if resolution == "references" {
						resolver := &staticEvaluatorTestResolver{resolver: binder.NameResolver{Globals: sourceFile.Locals}}
						evaluator = NewStaticStringEvaluatorWithCallEscapes(nil, sourceFile, resolver)
					}
					for range 2 {
						value, known := evaluator.Eval(findVariableInitializer(t, sourceFile, "result"))
						if known != test.known || known && value != "catch" {
							t.Fatalf("result = (%q, %v), want known = %v", value, known, test.known)
						}
					}
				})
			}
		})
	}
}

func TestStaticStringEvaluatorIteratorPrototypeMutations(t *testing.T) {
	for _, test := range []struct {
		name, mutation              string
		arrayMutated, stringMutated bool
	}{
		{"array assignment", `Array.prototype[Symbol.iterator] = function*() { yield "changed"; };`, true, false},
		{"string assignment", `String.prototype[Symbol.iterator] = function*() { yield "changed"; };`, false, true},
		{"property definition", `Object.defineProperty(Array.prototype, Symbol.iterator, {value: function*() { yield "changed"; }});`, true, false},
		{"spread property definition", `Object.defineProperty(...[Array.prototype, Symbol.iterator, {value: function*() { yield "changed"; }}]);`, true, false},
		{"empty spread before target", `Object.defineProperty(...[], String.prototype, Symbol.iterator, {value: function*() { yield "changed"; }});`, false, true},
		{"spread explicit receiver", `Reflect.set(...[{}, Symbol.iterator, function*() {}, String.prototype]);`, false, true},
		{"multiple definitions", `Object.defineProperties(String.prototype, {[Symbol.iterator]: {value: function*() { yield "changed"; }}});`, false, true},
		{"object assign", `Object.assign(Array.prototype, {[Symbol.iterator]: function*() { yield "changed"; }});`, true, false},
		{"reflect definition", `Reflect.defineProperty(String.prototype, Symbol.iterator, {value: function*() { yield "changed"; }});`, false, true},
		{"reflect set", `Reflect.set(Array.prototype, Symbol.iterator, function*() { yield "changed"; });`, true, false},
		{"explicit receiver", `Reflect.set({}, Symbol.iterator, function*() { yield "changed"; }, String.prototype);`, false, true},
		{"delete", `delete Array.prototype[Symbol.iterator];`, true, false},
		{"reflect deletion", `Reflect.deleteProperty(String.prototype, Symbol.iterator);`, false, true},
		{"computed object method", `const method = "defineProperty"; Object[method](Array.prototype, Symbol.iterator, {});`, true, false},
		{"computed reflect method", `const method = "set"; Reflect[method](String.prototype, Symbol.iterator, function*() {});`, false, true},
		{"prototype alias", `const proto = Array.prototype; proto[Symbol.iterator] = function*() {};`, true, false},
		{"constructor alias", `const Constructor = String; Constructor.prototype[Symbol.iterator] = function*() {};`, false, true},
		{"destructured array prototype", `const {prototype: proto} = Array; proto[Symbol.iterator] = function*() {};`, true, false},
		{"destructured string prototype", `const {prototype: proto} = String; proto[Symbol.iterator] = function*() {};`, false, true},
		{"TypeScript wrapper", `(Array.prototype as object)[Symbol.iterator] = function*() {};`, true, false},
		{"unmodified prototypes", `Array.prototype[Symbol.iterator]; String.prototype[Symbol.iterator];`, false, false},
		{"shadowed array", `{ const Array = {prototype: {}}; Array.prototype[Symbol.iterator] = function*() {}; }`, false, false},
		{"shadowed string", `{ const String = {prototype: {}}; String.prototype[Symbol.iterator] = function*() {}; }`, false, false},
		{"shadowed object", `{ const Object = {defineProperty() {}}; Object.defineProperty(Array.prototype, Symbol.iterator, {}); }`, false, false},
		{"shadowed reflect", `{ const Reflect = {set() {}}; Reflect.set(String.prototype, Symbol.iterator, null); }`, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := fixtures.GetRootDir()
			filePath := tspath.ResolvePath(root.Dir, "iterator-mutation.ts")
			code := test.mutation + `
const arraySpread = [...["stable"]];
const stringSpread = [..."stable"];
const arraySource = ["stable"];
const aliasSpread = [...arraySource];
const callSpread = Array.of(...["stable"]);
const arrayProperty = ({...["stable"]})[0];
const stringProperty = ({..."stable"})[0];
`
			fs := NewOverlayVFS(root.FS, map[string]string{filePath: code})
			program, err := CreateProgram(true, fs, root.Dir, "tsconfig.json", CreateCompilerHost(root.Dir, fs))
			assert.NilError(t, err)
			sourceFile := program.GetSourceFile(filePath)
			assert.Assert(t, sourceFile != nil)
			typeChecker, done := program.GetTypeChecker(t.Context())
			defer done()
			for _, resolution := range []string{"checker", "references"} {
				for _, first := range []string{"arraySpread", "stringSpread", "aliasSpread", "arrayProperty"} {
					t.Run(resolution+"/"+first, func(t *testing.T) {
						evaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
						if resolution == "references" {
							resolver := &staticEvaluatorTestResolver{resolver: binder.NameResolver{Globals: sourceFile.Locals}}
							evaluator = NewStaticStringEvaluatorWithReferenceResolver(nil, sourceFile, resolver)
						}
						evaluator.EvalValue(findVariableInitializer(t, sourceFile, first))
						for _, query := range []struct {
							name    string
							mutated bool
						}{
							{"arraySpread", test.arrayMutated},
							{"stringSpread", test.stringMutated},
							{"aliasSpread", test.arrayMutated},
							{"callSpread", test.arrayMutated},
						} {
							initializer := findVariableInitializer(t, sourceFile, query.name)
							if _, known := evaluator.EvalValue(initializer); known == query.mutated {
								t.Fatalf("%s known = %v, want %v", query.name, known, !query.mutated)
							}
							if query.name != "callSpread" {
								if _, known := evaluator.EvalValueIfNoSideEffects(initializer); known == query.mutated {
									t.Fatalf("%s safety = %v, want %v", query.name, known, !query.mutated)
								}
								if effect := evaluator.HasSideEffect(initializer, false); effect != query.mutated {
									t.Fatalf("%s side effect = %v, want %v", query.name, effect, query.mutated)
								}
							}
						}
						for name, want := range map[string]string{"arrayProperty": "stable", "stringProperty": "s"} {
							if value, known := evaluator.Eval(findVariableInitializer(t, sourceFile, name)); !known || value != want {
								t.Fatalf("%s = (%q, %v), want (%q, true)", name, value, known, want)
							}
						}
					})
				}
			}
		})
	}
}

func TestStaticStringEvaluatorConfiguredGlobals(t *testing.T) {
	for _, expression := range []string{
		"String.raw`\\n`", `String("x")`, `Array.of("x")[0]`,
		`Object.freeze(["x"])[0]`, `undefined`, `Math.PI`, `Number.EPSILON`,
		`Infinity`, `NaN`, `String(Symbol.iterator)`, `Number.parseInt("10")`,
		`parseInt("10")`, `Math.max(Infinity, 2)`,
	} {
		t.Run(expression, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/globals.js", Path: "/globals.js"}, "const value = "+expression, core.ScriptKindJS)
			evaluator := NewStaticStringEvaluatorWithSourceFile(nil, source)
			initializer := findVariableInitializer(t, source, "value")
			for _, access := range []GlobalAccess{GlobalAccessOff, GlobalAccessReadonly, GlobalAccessWritable} {
				evaluator.GlobalAccess = func(string) GlobalAccess { return access }
				if _, known := evaluator.EvalToString(initializer); known != access.IsDeclared() {
					t.Fatalf("EvalToString with %s globals: known = %v", access, known)
				}
			}
		})
	}
}

func TestStaticStringEvaluatorReviewValues(t *testing.T) {
	for _, test := range []struct {
		expression   string
		want         string
		known        bool
		withoutScope bool
	}{
		{`String(Symbol.iterator)`, "Symbol(Symbol.iterator)", true, false},
		{`String(Symbol.for("foo"))`, "Symbol(foo)", true, false},
		{`String(Symbol.for(""))`, "Symbol()", true, false},
		{`String(...[Symbol.iterator])`, "Symbol(Symbol.iterator)", true, false},
		{`String(Symbol.iterator, unknown)`, "", false, false},
		{"`${Symbol.iterator}`", "", false, false},
		{`Symbol.iterator + ""`, "", false, false},
		{`Symbol.for(Symbol.iterator)`, "", false, false},
		{`Number.parseInt("10px")`, "10", true, false},
		{`Number["parseInt"]("0x10")`, "16", true, false},
		{`parseInt("1e2")`, "1", true, false},
		{`parseInt("10", 4294967298)`, "2", true, false},
		{`parseInt("10", undefined)`, "10", true, false},
		{`parseInt()`, "NaN", true, false},
		{`parseInt(10n)`, "10", true, false},
		{`parseInt(Symbol.iterator)`, "", false, false},
		{`parseInt("10", 2n)`, "", false, false},
		{`parseInt("10", Symbol.iterator)`, "", false, false},
		{`Math.max(Infinity, 2)`, "Infinity", true, false},
		{`Math.min(-Infinity, 2)`, "-Infinity", true, false},
		{`String(Math.max(1, 2))`, "2", true, false},
		{`Math.max()`, "-Infinity", true, false},
		{`Math.min()`, "Infinity", true, false},
		{`Math.max(NaN, 2)`, "NaN", true, false},
		{`Math.max(NaN, Symbol.iterator)`, "", false, false},
		{`Math.min(NaN, 2n)`, "", false, false},
		{`1 / Math.max(-0, 0)`, "Infinity", true, false},
		{`1 / Math.min(-0, 0)`, "-Infinity", true, false},
		{`Math.abs(-1) + 1`, "2", true, false},
		{`String(Math.abs(-1))`, "1", true, false},
		{`Math.abs()`, "NaN", true, false},
		{`Math.abs(-1, Symbol.iterator)`, "1", true, false},
		{`Math.abs(-1, 2n)`, "1", true, false},
		{`Math.abs(-1, unknown())`, "", false, false},
		{`Math.abs(Symbol.iterator)`, "", false, false},
		{`Math.abs(2n)`, "", false, false},
		{`Math.clz32()`, "32", true, false},
		{`Math.imul(2)`, "0", true, false},
		{`Math.pow(1, NaN)`, "NaN", true, false},
		{`Math.pow(-1, Infinity)`, "NaN", true, false},
		{`Math.pow(NaN, 0)`, "1", true, false},
		{`Math.round(-1.5)`, "-1", true, false},
		{`Math.round(0.49999999999999994)`, "0", true, false},
		{`1 / Math.round(-0.5)`, "-Infinity", true, false},
		{`1 / Math.sign(-0)`, "-Infinity", true, false},
		{`1 / Math.abs(-0)`, "Infinity", true, false},
		{`Math.hypot(Infinity, NaN)`, "Infinity", true, false},
		{`Math.hypot(NaN, Infinity)`, "Infinity", true, false},
		{`Math.hypot(Number.MAX_VALUE, Number.MAX_VALUE, NaN)`, "NaN", true, false},
		{`Math.hypot(Number.MAX_VALUE, Number.MAX_VALUE, NaN, Infinity)`, "Infinity", true, false},
		{`Math.hypot(Infinity, Symbol.iterator)`, "", false, false},
		{`Math.hypot()`, "0", true, false},
		{`Math.random()`, "", false, false},
		{`"\uFEFF foo \u2029".trim()`, "foo", true, true},
		{`"\u0085foo\u0085".trim()`, "\u0085foo\u0085", true, true},
		{`null?.x`, "undefined", true, true},
		{`null?.[unknown]`, "undefined", true, true},
		{`null?.x.y(unknown)`, "undefined", true, true},
		{`null?.x!.y`, "undefined", true, true},
		{`(null?.x)?.()`, "undefined", true, true},
		{`(null?.x).y`, "", false, false},
		{`"x"?.trim?.()`, "x", true, true},
		{`String?.("x")`, "x", true, false},
		{`Symbol?.for("x")`, "", false, false},
		{`({})?.missing.y`, "", false, false},
		{`({}).toString?.()`, "", false, false},
		{`({toString: undefined}).toString?.()`, "undefined", true, false},
		{`({__proto__: null}).toString?.()`, "undefined", true, true},
		{`null.x`, "", false, false},
	} {
		t.Run(test.expression, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/review.ts", Path: "/review.ts"}, "const value = "+test.expression, core.ScriptKindTS)
			initializer := findVariableInitializer(t, source, "value")
			for _, scope := range []bool{false, true} {
				evaluator := NewStaticStringEvaluatorWithoutScope()
				if scope {
					evaluator = NewStaticStringEvaluatorWithSourceFile(nil, source)
				}
				wantKnown := test.known && (scope || test.withoutScope)
				got, known := evaluator.EvalToString(initializer)
				if known != wantKnown || known && got != test.want {
					t.Fatalf("EvalToString with scope=%v = (%q, %v), want (%q, %v)", scope, got, known, test.want, wantKnown)
				}
			}
		})
	}
}

func TestIsMutatingArrayMethod(t *testing.T) {
	for _, name := range []string{
		"copyWithin",
		"fill",
		"pop",
		"push",
		"reverse",
		"shift",
		"sort",
		"splice",
		"unshift",
	} {
		if !isMutatingArrayMethod(name) {
			t.Errorf("isMutatingArrayMethod(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"", "concat", "join", "map", "slice"} {
		if isMutatingArrayMethod(name) {
			t.Errorf("isMutatingArrayMethod(%q) = true, want false", name)
		}
	}
}

func TestStaticArrayJoinRejectsExcessiveOutput(t *testing.T) {
	value := &staticArrayValue{
		length:   2048,
		overflow: make([]any, 2046),
	}
	for index := range value.length {
		value.set(index, staticUndefinedValue{})
	}
	if _, ok := staticArrayJoin(value, strings.Repeat("x", 1024)); ok {
		t.Fatal("staticArrayJoin accepted output above the static string limit")
	}
}

func TestStaticStringEvaluatorSpreadLimits(t *testing.T) {
	text := strings.Repeat("x", maxStaticArrayLength)
	for _, test := range []struct {
		name, expression string
		known            bool
	}{
		{"at limit", `[..."` + text + `"]`, true},
		{"deep nesting", strings.Repeat("[", 64) + "0" + strings.Repeat(", ...[]]", 64), true},
		{"string overflow", `[..."` + text + `x"]`, false},
		{"element overflow", `[...[..."` + text + `"], 1]`, false},
		{"array overflow", `[...[..."` + text + `"], ...[1]]`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/spread.ts", Path: "/spread.ts",
			}, "const value = "+test.expression+";", core.ScriptKindTS)
			evaluator := NewStaticStringEvaluator(nil)
			isArray, known := evaluator.EvalArrayValue(findVariableInitializer(t, sourceFile, "value"))
			assert.Equal(t, known, test.known)
			assert.Equal(t, isArray, test.known)
		})
	}
	for _, test := range []struct {
		name, expression string
		known            bool
	}{
		{"object overflow", `{..."` + strings.Repeat("x", maxStaticObjectProperties+1) + `"}`, false},
		{"nested objects", strings.Repeat("{value:", 64) + "0" + strings.Repeat(", ...{}}", 64), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/spread.ts", Path: "/spread.ts",
			}, "const value = "+test.expression+";", core.ScriptKindTS)
			_, known := NewStaticStringEvaluator(nil).EvalValue(findVariableInitializer(t, sourceFile, "value"))
			assert.Equal(t, known, test.known)
		})
	}
}

func TestStaticStringEvaluatorUTF16Semantics(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "utf16.ts")
	code := `
const loneLength = String('\uD800'.length);
const loneToString = '\uD83D'.toString();
const astralToString = '😀'.toString();
const indexAfterLone = String('\uD800x'.indexOf('x'));
const astralHighIndex = String('😀'.indexOf('\uD83D'));
const astralLowIndex = String('😀'.indexOf('\uDE00'));
const positiveInfinityStart = String('abc'.indexOf('a', 1 / 0));
const negativeInfinityStart = String('abc'.indexOf('a', -1 / 0));
const hugePositiveStart = String('abc'.indexOf('a', 1e100));
const nanStart = String('abc'.indexOf('a', 0 / 0));
const missingNeedle = String('undefined!'.indexOf());
const missingNeedleAbsent = String('abc'.indexOf());
const emptyNeedleInfinity = String('abc'.indexOf('', 1 / 0));
const emptyNeedleMidPair = String('😀'.indexOf('', 1));
const sliceHigh = '😀'.slice(0, 1);
const sliceLow = '😀'.slice(1, 2);
const substringHigh = '😀'.substring(0, 1);
const substrLow = '😀'.substr(1, 1);
const charAtHigh = '😀'.charAt(0);
const charAtLow = '😀'.charAt(1);
const fromCharCodeHigh = String.fromCharCode(0xD83D);
const fromCharCodePair = String.fromCharCode(0xD83D, 0xDE00);
const concatPair = '\uD83D' + '\uDE00';
` + "const templatePair = `${'\\uD83D'}${'\\uDE00'}`;\n" + `
const joinPair = ['\uD83D', '\uDE00'].join('');
const astralUpper = ('\uD801' + '\uDC28').toUpperCase();
const astralLower = ('\uD801' + '\uDC00').toLowerCase();
const strictPair = ('\uD83D' + '\uDE00') === '😀' ? 'yes' : 'no';
const objectPair = ({['\uD83D' + '\uDE00']: 'yes'})['😀'];
const indexedHigh = '😀'[0];
const indexedLow = '😀'[1];
const indexedLone = '\uD800'[0];
const spreadAstral = [...'A😀B'].join('-');
const spreadAstralLength = String([...'A😀B'].length);
const spreadLone = [...'\uD800x'][0];
const spreadPair = String([...('\uD83D' + '\uDE00')].length);
const objectSpreadAstralHigh = ({...'😀'})[0];
const objectSpreadAstralLow = ({...'😀'})[1];
const objectSpreadLone = ({...'\uD800x'})[0];
`
	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err, "couldn't create program")
	sourceFile := program.GetSourceFile(filePath)
	assert.Assert(t, sourceFile != nil)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()
	staticEvaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)

	high := ecmascript.StringFromCodeUnits([]uint16{0xD83D})
	low := ecmascript.StringFromCodeUnits([]uint16{0xDE00})
	tests := []struct {
		name string
		want string
	}{
		{name: "loneLength", want: "1"},
		{name: "loneToString", want: high},
		{name: "astralToString", want: "😀"},
		{name: "indexAfterLone", want: "1"},
		{name: "astralHighIndex", want: "0"},
		{name: "astralLowIndex", want: "1"},
		{name: "positiveInfinityStart", want: "-1"},
		{name: "negativeInfinityStart", want: "0"},
		{name: "hugePositiveStart", want: "-1"},
		{name: "nanStart", want: "0"},
		{name: "missingNeedle", want: "0"},
		{name: "missingNeedleAbsent", want: "-1"},
		{name: "emptyNeedleInfinity", want: "3"},
		{name: "emptyNeedleMidPair", want: "1"},
		{name: "sliceHigh", want: high},
		{name: "sliceLow", want: low},
		{name: "substringHigh", want: high},
		{name: "substrLow", want: low},
		{name: "charAtHigh", want: high},
		{name: "charAtLow", want: low},
		{name: "fromCharCodeHigh", want: high},
		{name: "fromCharCodePair", want: "😀"},
		{name: "concatPair", want: "😀"},
		{name: "templatePair", want: "😀"},
		{name: "joinPair", want: "😀"},
		{name: "astralUpper", want: "𐐀"},
		{name: "astralLower", want: "𐐨"},
		{name: "strictPair", want: "yes"},
		{name: "objectPair", want: "yes"},
		{name: "indexedHigh", want: high},
		{name: "indexedLow", want: low},
		{name: "indexedLone", want: ecmascript.StringFromCodeUnits([]uint16{0xD800})},
		{name: "spreadAstral", want: "A-😀-B"},
		{name: "spreadAstralLength", want: "3"},
		{name: "spreadLone", want: ecmascript.StringFromCodeUnits([]uint16{0xD800})},
		{name: "spreadPair", want: "1"},
		{name: "objectSpreadAstralHigh", want: high},
		{name: "objectSpreadAstralLow", want: low},
		{name: "objectSpreadLone", want: ecmascript.StringFromCodeUnits([]uint16{0xD800})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := staticEvaluator.Eval(findVariableInitializer(t, sourceFile, tt.name))
			if !ok || got != tt.want {
				t.Fatalf("Eval(%s) = (%q, %v), want code units %v", tt.name, got, ok, ecmascript.StringCodeUnits(tt.want))
			}
		})
	}
}

func TestStaticStringIndexOfCommonPrefix(t *testing.T) {
	text := strings.Repeat("a", 100_000)
	needle := strings.Repeat("a", 50_000) + "b"
	result := staticStringIndexOf(text, []any{needle, staticNumberValue(math.Inf(-1))})
	value, ok := staticValueToNumber(result.value)
	if !result.ok || !ok || value != -1 {
		t.Fatalf("staticStringIndexOf common prefix = (%v, %v), want -1", result.value, result.ok)
	}
}

func findVariableInitializer(t testing.TB, sourceFile *ast.SourceFile, bindingName string) *ast.Node {
	t.Helper()

	var initializer *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind == ast.KindVariableDeclaration {
			declaration := node.AsVariableDeclaration()
			if declaration != nil && declaration.Name() != nil &&
				declaration.Name().Kind == ast.KindIdentifier &&
				declaration.Name().AsIdentifier().Text == bindingName {
				initializer = declaration.Initializer
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	visit(sourceFile.AsNode())
	if initializer == nil {
		t.Fatalf("missing initializer for %q", bindingName)
	}
	return initializer
}

func TestStaticStringEvaluatorControlFlowSafety(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "control-flow.ts")
	code := `
		let loose = 1;
		let flag;
		const scalar = 1;
		const regularExpression = /a/g;
		const regexpAlias = regularExpression;
		let mutableRegExp = /a/;
		const mutableRegExpAlias = mutableRegExp;
		const array = [1];
		const literalMember = ({value: "x"}).value;
		const arrayIndex = [1][0];
		const arrayLength = [1]["length"];
		const globalNumber = Math["PI"];
		const globalInfinity = 0 / Infinity;
		const globalNaN = NaN & 0;
		const template = ` + "`abc`" + `;
		const templateLength = ` + "`abc`.length" + `;
		const templateAliasLength = template.length;
		const templateIndex = ` + "`abc`[1]" + `;
		const templateGlobalNumber = Math[(` + "`PI`" + `)];
		const parenthesizedGlobalNumber = Math[("PI")];
		const assertedGlobalNumber = (Math as typeof Math).PI;
		const nonNullGlobalNumber = Math!.PI;
		const assertedObjectMember = ({value: 1} as const).value;
		const assertedStringMember = ("abc" as string).length;
		const substitutedGlobalNumber = Math[` + "`P${'I'}`" + `];
		const substitutedLength = ` + "`a${1}`.length" + `;
		const frozen = Object.freeze([1]);
		const instantiatedArray = (Object.freeze<number[]>)([1]);
		const frozenRegExp = (Object.freeze<RegExp>)(/a/);
		const sealedRegExp = (Object.seal<RegExp>)(/a/);
		const fixedRegExp = (Object.preventExtensions<RegExp>)(/a/);
		const optionalRegExp = (Object.freeze<RegExp>)?.(/a/);
		const optionalMemberRegExp = (Object?.freeze<RegExp>)(/a/);
		const computedRegExp = (Object["freeze"]<RegExp>)(/a/);
		const frozenAlias = frozen;
		const safeAlias = array;
		const skippedMutable = true ? array : loose;
		const beforeDeclaration = later;
		const skippedBeforeDeclaration = true ? array : later;
		const later = "x";
		const reachedMutable = loose;
		const voidMutable = void loose ?? array;
		const assignment = flag = true;
		const assignmentAlias = assignment;
		const assignmentCondition = (flag = true) ? 1 : unknown;
		const object = {value: [1]};
		Object.defineProperty(object, "value", {get() { return unknown; }});
		const getter = object.value;
		const getterAlias = getter;
		const call = String("x");
		const skippedSideEffect = true ? "x" : call;
		const cycle = cycle;
		const unknownValue = unknown;
		let mutationMethod = "fill";
		const mutated = [""];
		mutated[mutationMethod]("changed");
		const mutationUse = mutated[0];
	`
	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err)
	sourceFile := program.GetSourceFile(filePath)
	assert.Assert(t, sourceFile != nil)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()

	staticEvaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
	for _, test := range []struct {
		name     string
		known    bool
		isArray  bool
		isRegExp bool
		// Known for value evaluation, but rejected as a control-flow input.
		valueOnly bool
	}{
		{name: "scalar", known: true},
		{name: "regularExpression", known: true, isRegExp: true},
		{name: "regexpAlias", known: true, isRegExp: true},
		{name: "mutableRegExpAlias", valueOnly: true},
		{name: "array", known: true, isArray: true},
		{name: "literalMember", known: true},
		{name: "arrayIndex", known: true},
		{name: "arrayLength", known: true},
		{name: "globalNumber", known: true},
		{name: "globalInfinity", known: true},
		{name: "globalNaN", known: true},
		{name: "templateLength", known: true},
		{name: "templateAliasLength", known: true},
		{name: "templateIndex", known: true},
		{name: "templateGlobalNumber", known: true},
		{name: "parenthesizedGlobalNumber", known: true},
		{name: "assertedGlobalNumber"},
		{name: "nonNullGlobalNumber"},
		{name: "assertedObjectMember"},
		{name: "assertedStringMember"},
		{name: "substitutedGlobalNumber"},
		{name: "substitutedLength"},
		{name: "frozen", known: true, isArray: true},
		{name: "instantiatedArray", known: true, isArray: true},
		{name: "frozenRegExp", known: true, isRegExp: true},
		{name: "sealedRegExp", known: true, isRegExp: true},
		{name: "fixedRegExp", known: true, isRegExp: true},
		{name: "optionalRegExp"},
		{name: "optionalMemberRegExp"},
		{name: "computedRegExp"},
		{name: "frozenAlias", known: true, isArray: true},
		{name: "safeAlias", known: true, isArray: true},
		{name: "skippedMutable", known: true, isArray: true},
		{name: "beforeDeclaration", valueOnly: true},
		{name: "skippedBeforeDeclaration", known: true, isArray: true},
		{name: "reachedMutable", valueOnly: true},
		{name: "voidMutable", valueOnly: true},
		{name: "assignment"},
		{name: "assignmentAlias"},
		{name: "assignmentCondition"},
		{name: "getter"},
		{name: "getterAlias"},
		{name: "call"},
		{name: "skippedSideEffect"},
		{name: "cycle"},
		{name: "unknownValue"},
	} {
		t.Run(test.name, func(t *testing.T) {
			node := findVariableInitializer(t, sourceFile, test.name)
			_, known := staticEvaluator.EvalControlFlowValue(node)
			assert.Equal(t, known, test.known)
			isArray, known := staticEvaluator.EvalControlFlowArrayValue(node)
			assert.Equal(t, known, test.known)
			assert.Equal(t, isArray, test.isArray)
			isRegExp, known := staticEvaluator.EvalControlFlowRegExpValue(node)
			assert.Equal(t, known, test.known)
			assert.Equal(t, isRegExp, test.isRegExp)
			_, known = staticEvaluator.EvalControlFlowTruthiness(node)
			assert.Equal(t, known, test.known)
			// Merge rules need the same side-effect/member safety, while stable
			// let bindings and forward const values do not control a branch.
			_, known = staticEvaluator.EvalValueIfNoSideEffects(node)
			assert.Equal(t, known, test.known || test.valueOnly)
		})
	}

	// Scope-free callers must not gain built-in global resolution.
	withoutScope := NewStaticStringEvaluatorWithoutScope()
	global := findVariableInitializer(t, sourceFile, "globalNumber")
	_, known := withoutScope.EvalValue(global)
	assert.Assert(t, !known, "scope-free evaluation resolved a global constant")
	_, known = withoutScope.EvalControlFlowValue(global)
	assert.Assert(t, !known, "scope-free control-flow evaluation resolved a global constant")

	// A rejected conservative evaluation must not change ordinary evaluation,
	// including its stable let resolution and scalar assignment semantics.
	for _, name := range []string{"reachedMutable", "assignment", "assignmentAlias", "call"} {
		_, known := staticEvaluator.EvalValue(findVariableInitializer(t, sourceFile, name))
		assert.Assert(t, known, "ordinary evaluation changed for %s", name)
	}

	// Computing the mutation cache during conservative evaluation must still
	// resolve a stable let binding used as a computed mutating method name.
	freshEvaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
	freshEvaluator.EvalControlFlowValue(findVariableInitializer(t, sourceFile, "safeAlias"))
	_, known = freshEvaluator.Eval(findVariableInitializer(t, sourceFile, "mutationUse"))
	assert.Assert(t, !known, "conservative evaluation hid a computed property mutation")

	for _, evaluator := range []*StaticStringEvaluator{nil, staticEvaluator} {
		_, known := evaluator.EvalValueIfNoSideEffects(nil)
		assert.Assert(t, !known)
		_, known = evaluator.EvalControlFlowValue(nil)
		assert.Assert(t, !known)
		isArray, known := evaluator.EvalControlFlowArrayValue(nil)
		assert.Assert(t, !known && !isArray)
		isRegExp, known := evaluator.EvalControlFlowRegExpValue(nil)
		assert.Assert(t, !known && !isRegExp)
		truthy, known := evaluator.EvalControlFlowTruthiness(nil)
		assert.Assert(t, !known && !truthy)
	}
}

func TestStaticStringEvaluatorControlFlowNullish(t *testing.T) {
	for _, test := range []struct {
		expression string
		nullish    bool
		known      bool
	}{
		{"null", true, true},
		{"undefined", true, true},
		{"void 0", true, true},
		{"(null)", true, true},
		{"null ?? undefined", true, true},
		{"0", false, true},
		{"''", false, true},
		{"false", false, true},
		{"NaN", false, true},
		{"'a'", false, true},
		{"[]", false, true},
		{"null ?? 'a'", false, true},
		{"unknown", false, false},
		{"fn()", false, false},
		{"(unknown = null)", false, false},
	} {
		t.Run(test.expression, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/nullish.js"}, "const value = "+test.expression+";", core.ScriptKindJS)
			node := findVariableInitializer(t, source, "value")
			nullish, known := NewStaticStringEvaluator(nil).EvalControlFlowNullish(node)
			if nullish != test.nullish || known != test.known {
				t.Fatalf("EvalControlFlowNullish = (%v, %v), want (%v, %v)", nullish, known, test.nullish, test.known)
			}
		})
	}
	var evaluator *StaticStringEvaluator
	if nullish, known := evaluator.EvalControlFlowNullish(nil); nullish || known {
		t.Fatalf("nil evaluator = (%v, %v), want (false, false)", nullish, known)
	}
}

func TestStaticStringEvaluatorTypeof(t *testing.T) {
	for _, test := range []struct {
		expression, want string
		known            bool
	}{
		{"typeof 1", "number", true},
		{"typeof 'x'", "string", true},
		{"typeof 1n", "bigint", true},
		{"typeof false", "boolean", true},
		{"typeof null", "object", true},
		{"typeof void 0", "undefined", true},
		{"typeof []", "object", true},
		{"typeof ({})", "object", true},
		{"typeof /x/", "object", true},
		{"typeof typeof 1", "string", true},
		{"typeof unknown", "", false},
		{"typeof fn()", "", false},
	} {
		t.Run(test.expression, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/typeof.js"}, "const value = "+test.expression+";", core.ScriptKindJS)
			node := findVariableInitializer(t, source, "value")
			got, known := NewStaticStringEvaluatorWithoutScope().Eval(node)
			if got != test.want || known != test.known {
				t.Fatalf("Eval = (%q, %v), want (%q, %v)", got, known, test.want, test.known)
			}
		})
	}
}

func TestStaticStringEvaluatorBigInt(t *testing.T) {
	for _, test := range []struct {
		expression, want string
		known            bool
	}{
		{"42n", "42", true}, {"0x2_an", "42", true}, {"0o52n", "42", true}, {"0b101010n", "42", true},
		{"'' + 42n", "42", true}, {"`pkg${40n + 2n}`", "pkg42", true},
		{"[40n + 2n, 40n].join('/')", "42/40", true},
		{"(7n - 1n) * 7n", "42", true}, {"-43n / 2n", "-21", true}, {"-43n % 2n", "-1", true},
		{"2n ** 5n", "32", true}, {"0n ** 0n", "1", true}, {"(-2n) ** 3n", "-8", true},
		{"0n ** 999999999999999999999n", "0", true},
		{"1n ** 999999999999999999999n", "1", true},
		{"(-1n) ** 999999999999999999999n", "-1", true},
		{"(-1n) ** 999999999999999999998n", "1", true},
		{"(-1n) ** (1n << 1048575n)", "1", true},
		{"(-1n) ** ((1n << 1048575n) + 1n)", "-1", true},
		{"6n | 3n", "7", true}, {"6n & 3n", "2", true}, {"6n ^ 3n", "5", true}, {"~42n", "-43", true},
		{"42n << 2n", "168", true}, {"42n << -1n", "21", true}, {"42n >> -1n", "84", true},
		{"-43n >> 1n", "-22", true}, {"-1n >> 999999999999999999999n", "-1", true},
		{"0n << 999999999999999999999n", "0", true},
		{"(1n << 65536n) ? 'fs' : 'path'", "fs", true},
		{"(2n ** 65536n) === (1n << 65536n)", "true", true},
		{"(1n << 1048575n) >> 1048575n", "1", true},
		{"((1n << 524287n) * (1n << 524287n)) >> 1048574n", "1", true},
		{"1n << 1048576n", "", false},
		{"(1n << 1048575n) + (1n << 1048575n)", "", false},
		{"(1n << 524288n) * (1n << 524288n)", "", false},
		{"2n ** 1048576n", "", false},
		{"0n ? 'bad' : 'fs'", "fs", true}, {"!0n", "true", true}, {"0n || 'fs'", "fs", true}, {"1n && 'fs'", "fs", true},
		{"1n === 1n", "true", true}, {"1n === 1", "false", true}, {"1n !== 1", "true", true},
		{"1n == '1'", "true", true}, {"1n == true", "true", true}, {"0n == null", "false", true},
		{"1n + []", "1", true}, {"[1] + 2n", "12", true}, {"1n === []", "false", true}, {"[] !== 1n", "true", true},
		{"2n > 1.5", "true", true}, {"1.5 < 2n", "true", true}, {"2n <= '2'", "true", true},
		{"9007199254740993n > 9007199254740992", "true", true},
		{"2n >= 'invalid'", "false", true}, {"0n != 0/0", "true", true}, {"2n < 1/0", "true", true},
		{"2n > -1/0", "true", true}, {"1n == [1]", "true", true},
		{"40n + 2", "", false}, {"+42n", "", false}, {"1n / 0n", "", false}, {"1n % 0n", "", false},
		{"2n ** -1n", "", false}, {"42n >>> 1n", "", false}, {"2n ** 999999999999n", "", false},
		{"0n ** -1n", "", false}, {"1n ** -1n", "", false}, {"(-1n) ** -1n", "", false},
		{"1n << 999999999999n", "", false}, {"'x'.charAt(0n)", "", false},
	} {
		t.Run(test.expression, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/bigint.js", Path: "/bigint.js"}, "const value = "+test.expression+";", core.ScriptKindJS)
			node := findVariableInitializer(t, source, "value")
			got, known := NewStaticStringEvaluatorWithoutScope().EvalToString(node)
			if known != test.known || known && got != test.want {
				t.Fatalf("EvalToString = (%q, %v), want (%q, %v)", got, known, test.want, test.known)
			}
		})
	}
}

func TestStaticStringEvaluatorBinaryExpressions(t *testing.T) {
	for _, test := range []struct{ expression, want string }{
		{"0 < 1 ? 'fs' : 'path'", "fs"}, {"0 + true", "1"}, {"0 - '1'", "-1"},
		{"true + false", "1"}, {"null + false", "0"}, {"void 0 + 1", "NaN"},
		{"'6' * true", "6"}, {"'6' / '2'", "3"}, {"'-7' % '2'", "-1"},
		{"'2' ** true", "2"}, {"'1' ** (1/0)", "NaN"}, {"'1' ** (0/0)", "NaN"},
		{"1 / ('-0' - 0)", "-Infinity"}, {"'bad' - false", "NaN"},
		{"true << 33", "2"}, {"'-8' >> true", "-4"}, {"'-1' >>> false", "4294967295"},
		{"'4294967295' >>> 32", "4294967295"}, {"(1/0) | false", "0"},
		{"'6' & true", "0"}, {"'6' ^ true", "7"},
		{"'2' < '10'", "false"}, {"'2' < 10", "true"}, {"'2' <= '2'", "true"},
		{"'10' > '2'", "false"}, {"'2' >= '2'", "true"},
		{"'\U00010000' < '\uE000'", "true"},
		{"null < 1", "true"}, {"void 0 >= 0", "false"}, {"void 0 <= 0", "false"},
		{"0 == false", "true"}, {"0 != false", "false"}, {"'0' == false", "true"},
		{"null == 0", "false"}, {"null == void 0", "true"}, {"void 0 == false", "false"},
		{"'fs' == 'fs'", "true"}, {"'fs' != 'path'", "true"},
		{"[] == false", "true"}, {"[1] == true", "true"}, {"({}) == '[object Object]'", "true"},
		{"[1] + 2", "12"}, {"1 + [2]", "12"}, {"[] + []", ""},
		{"({}) + 1", "[object Object]1"}, {"/fs/ + 1", "/fs/1"},
		{"[1] - 2", "-1"}, {"[] < 1", "true"},
	} {
		t.Run(test.expression, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/binary.js", Path: "/binary.js"}, "const value = "+test.expression+";", core.ScriptKindJS)
			node := findVariableInitializer(t, source, "value")
			got, known := NewStaticStringEvaluatorWithoutScope().EvalToString(node)
			if !known || got != test.want {
				t.Fatalf("EvalToString = (%q, %v), want %q", got, known, test.want)
			}
		})
	}
}

func TestStaticStringEvaluatorGlobalNumbers(t *testing.T) {
	for _, test := range []struct {
		code, want string
		known      bool
	}{
		{"const value = NaN;", "NaN", true},
		{"const value = -Infinity;", "-Infinity", true},
		{"const value = String(Infinity);", "Infinity", true},
		{"const value = [NaN, Infinity].join(',');", "NaN,Infinity", true},
		{"function f(NaN) { const value = NaN; }", "", false},
		{"function f(Infinity) { const value = Infinity; }", "", false},
	} {
		t.Run(test.code, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/globals.js", Path: "/globals.js"}, test.code, core.ScriptKindJS)
			node := findVariableInitializer(t, source, "value")
			got, known := NewStaticStringEvaluatorWithSourceFile(nil, source).EvalToString(node)
			assert.Equal(t, known, test.known)
			assert.Equal(t, got, test.want)
			_, known = NewStaticStringEvaluatorWithoutScope().EvalValue(node)
			assert.Assert(t, !known, "scope-free evaluation resolved a global")
		})
	}
}
