// cspell:ignore truenull

package utils

import (
	"math"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"gotest.tools/v3/assert"
)

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
		"const stringNoArgumentCall = String();\n" +
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

func TestStaticStringEvaluatorUTF16Semantics(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "utf16.ts")
	code := `
const loneLength = String('\uD800'.length);
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
