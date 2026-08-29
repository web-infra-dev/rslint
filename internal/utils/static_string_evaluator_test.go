// cspell:ignore ACFB Hrkt lookaheads truenull

package utils

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
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
		"const objectValueOfString = String(({valueOf: 1}));\n" +
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
		{name: "objectValueOfString", want: "[object Object]", ok: true},
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

func TestStaticEvaluatorBudgetFailureDoesNotPoisonLaterEvaluation(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "budget.ts")
	code := "const small = [1]; const first = small; const second = small;"
	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err, "couldn't create program")
	sourceFile := program.GetSourceFile(filePath)
	assert.Assert(t, sourceFile != nil)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()

	staticEvaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
	staticEvaluator.eslintStaticCalls = true
	staticEvaluator.stringWorkBudgetLeft = maxStaticStringWorkBudget
	staticEvaluator.bigIntWorkBudgetLeft = maxStaticBigIntWorkBudget
	staticEvaluator.evaluationDepth = 1
	staticEvaluator.evaluationStepsLeft = maxStaticEvaluationSteps
	staticEvaluator.aggregateBudgetLeft = 0
	if _, known := staticEvaluator.EvalArrayValue(findVariableInitializer(t, sourceFile, "first")); known {
		t.Fatal("budget-limited evaluation unexpectedly succeeded")
	}

	staticEvaluator.evaluationDepth = 0
	if isArray, known := staticEvaluator.EvalArrayValue(findVariableInitializer(t, sourceFile, "second")); !known || !isArray {
		t.Fatalf("fresh evaluation after a budget failure = (%v, %v), want (true, true)", isArray, known)
	}
}

func TestStaticEvaluatorEnhancedRecursionLimitsStayIsolated(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "recursion.ts")
	var code strings.Builder
	code.WriteString("const conditional0 = 1;\n")
	for depth := 1; depth <= maxStaticEvaluationDepth; depth++ {
		code.WriteString("const conditional")
		code.WriteString(strconv.Itoa(depth))
		code.WriteString(" = true ? conditional")
		code.WriteString(strconv.Itoa(depth - 1))
		code.WriteString(" : 0;\n")
	}
	code.WriteString("const alias0 = 1;\n")
	for depth := 1; depth <= maxStaticEvaluationDepth; depth++ {
		code.WriteString("const alias")
		code.WriteString(strconv.Itoa(depth))
		code.WriteString(" = alias")
		code.WriteString(strconv.Itoa(depth - 1))
		code.WriteString(";\n")
	}

	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code.String()})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err, "couldn't create program")
	sourceFile := program.GetSourceFile(filePath)
	assert.Assert(t, sourceFile != nil)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()

	enhanced := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
	enhanced.eslintStaticCalls = true
	enhanced.stringWorkBudgetLeft = maxStaticStringWorkBudget
	enhanced.stringWorkContext = &staticStringWorkContext{remaining: &enhanced.stringWorkBudgetLeft}
	enhanced.bigIntWorkBudgetLeft = maxStaticBigIntWorkBudget
	for _, test := range []struct {
		name  string
		known bool
	}{
		{name: "conditional" + strconv.Itoa(maxStaticEvaluationDepth/2-1), known: true},
		{name: "conditional" + strconv.Itoa(maxStaticEvaluationDepth/2)},
		{name: "alias" + strconv.Itoa(maxStaticEvaluationDepth-1), known: true},
		{name: "alias" + strconv.Itoa(maxStaticEvaluationDepth)},
	} {
		_, known := enhanced.EvalArrayValue(findVariableInitializer(t, sourceFile, test.name))
		if known != test.known {
			t.Fatalf("enhanced EvalArrayValue(%s) known = %v, want %v", test.name, known, test.known)
		}
	}

	legacy := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
	for _, name := range []string{
		"conditional" + strconv.Itoa(maxStaticEvaluationDepth),
		"alias" + strconv.Itoa(maxStaticEvaluationDepth),
	} {
		if isArray, known := legacy.EvalArrayValue(findVariableInitializer(t, sourceFile, name)); !known || isArray {
			t.Fatalf("legacy EvalArrayValue(%s) = (%v, %v), want (false, true)", name, isArray, known)
		}
	}
}

func TestStaticEvaluatorClosedBigIntCachePreservesDepthLimits(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "bigint-cache.ts")
	var code strings.Builder
	code.WriteString("const huge = 2n ** 1024n;\n")
	code.WriteString("const dependent = huge + 1n;\n")
	code.WriteString("const selected = true ? huge : 0n;\n")
	code.WriteString("const logical = true && huge;\n")
	code.WriteString("const nested = 2n")
	for range 900 {
		code.WriteString(" | 0n")
	}
	code.WriteString(";\n")
	code.WriteString("const nestedAlias0 = nested;\n")
	for depth := 1; depth <= 200; depth++ {
		code.WriteString("const nestedAlias")
		code.WriteString(strconv.Itoa(depth))
		code.WriteString(" = nestedAlias")
		code.WriteString(strconv.Itoa(depth - 1))
		code.WriteString(";\n")
	}
	code.WriteString("const bigAlias0 = 2n ** 16n;\n")
	for depth := 1; depth <= 1100; depth++ {
		code.WriteString("const bigAlias")
		code.WriteString(strconv.Itoa(depth))
		code.WriteString(" = bigAlias")
		code.WriteString(strconv.Itoa(depth - 1))
		code.WriteString(";\n")
	}
	code.WriteString("const bigConditional0 = 2n ** 16n;\n")
	for depth := 1; depth <= 600; depth++ {
		code.WriteString("const bigConditional")
		code.WriteString(strconv.Itoa(depth))
		code.WriteString(" = true ? bigConditional")
		code.WriteString(strconv.Itoa(depth - 1))
		code.WriteString(" : 0n;\n")
	}

	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code.String()})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err, "couldn't create program")
	sourceFile := program.GetSourceFile(filePath)
	assert.Assert(t, sourceFile != nil)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()

	staticEvaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)
	staticEvaluator.eslintStaticCalls = true
	staticEvaluator.stringWorkBudgetLeft = maxStaticStringWorkBudget
	staticEvaluator.stringWorkContext = &staticStringWorkContext{remaining: &staticEvaluator.stringWorkBudgetLeft}
	staticEvaluator.bigIntWorkBudgetLeft = maxStaticBigIntWorkBudget
	staticEvaluator.bigIntCacheBudgetLeft = maxStaticBigIntCacheBudget

	huge := findVariableInitializer(t, sourceFile, "huge")
	first := staticEvaluator.evalValue(huge)
	second := staticEvaluator.evalValue(huge)
	firstBigInt, firstOK := first.value.(staticBigIntValue)
	secondBigInt, secondOK := second.value.(staticBigIntValue)
	if !first.ok || !second.ok || !firstOK || !secondOK ||
		firstBigInt.value == nil || firstBigInt.value != secondBigInt.value {
		t.Fatal("a closed BigInt expression did not reuse its immutable result")
	}
	if _, cached := staticEvaluator.bigIntExpressionCache[SkipAssertionsAndParens(huge)]; !cached {
		t.Fatal("closed BigInt expression was not cached")
	}

	dependent := findVariableInitializer(t, sourceFile, "dependent")
	if result := staticEvaluator.evalValue(dependent); !result.ok {
		t.Fatal("dependent BigInt expression unexpectedly became unknown")
	}
	if _, cached := staticEvaluator.bigIntExpressionCache[SkipAssertionsAndParens(dependent)]; cached {
		t.Fatal("an identifier-dependent BigInt expression was cached")
	}
	for _, name := range []string{"selected", "logical"} {
		initializer := findVariableInitializer(t, sourceFile, name)
		if result := staticEvaluator.evalValue(initializer); !result.ok {
			t.Fatalf("%s unexpectedly became unknown", name)
		}
		if _, cached := staticEvaluator.bigIntExpressionCache[SkipAssertionsAndParens(initializer)]; cached {
			t.Fatalf("pass-through expression %s was cached", name)
		}
	}

	nested := findVariableInitializer(t, sourceFile, "nested")
	if result := staticEvaluator.evalValue(nested); !result.ok {
		t.Fatal("bounded closed BigInt tree unexpectedly became unknown")
	}
	if _, cached := staticEvaluator.bigIntExpressionCache[SkipAssertionsAndParens(nested)]; !cached {
		t.Fatal("bounded closed BigInt tree was not cached")
	}
	if isArray, known := staticEvaluator.EvalArrayValue(
		findVariableInitializer(t, sourceFile, "nestedAlias200"),
	); known || isArray {
		t.Fatalf("cached-depth replay = (%v, %v), want (false, false)", isArray, known)
	}

	for _, name := range []string{"bigAlias400", "bigAlias800", "bigConditional400"} {
		if isArray, known := staticEvaluator.EvalArrayValue(findVariableInitializer(t, sourceFile, name)); !known || isArray {
			t.Fatalf("prewarmed EvalArrayValue(%s) = (%v, %v), want (false, true)", name, isArray, known)
		}
	}
	for _, name := range []string{"bigAlias1100", "bigConditional600"} {
		if isArray, known := staticEvaluator.EvalArrayValue(findVariableInitializer(t, sourceFile, name)); known || isArray {
			t.Fatalf("deep EvalArrayValue(%s) = (%v, %v), want (false, false)", name, isArray, known)
		}
	}
}

func TestStaticEvaluatorBigIntWorkBudget(t *testing.T) {
	value := new(big.Int).Lsh(big.NewInt(1), maxStaticBigIntBits-1)
	cost := max((value.BitLen()+7)/8, 32)
	staticEvaluator := &StaticStringEvaluator{bigIntWorkBudgetLeft: cost*2 - 1}
	first := staticEvalResult{value: staticBigIntValue{value: value}, ok: true}
	if !staticEvaluator.chargeStaticBigIntResult(&first) {
		t.Fatal("first bounded BigInt result unexpectedly exhausted the work budget")
	}
	if !staticEvaluator.chargeStaticBigIntResult(&first) {
		t.Fatal("the same retained BigInt result was charged twice")
	}
	second := staticEvalResult{value: staticBigIntValue{value: new(big.Int).Set(value)}, ok: true}
	if staticEvaluator.chargeStaticBigIntResult(&second) {
		t.Fatal("a fresh BigInt result exceeded the persistent work budget")
	}
	if staticEvaluator.bigIntWorkBudgetLeft != 0 {
		t.Fatalf("BigInt work left after exhaustion = %d, want 0", staticEvaluator.bigIntWorkBudgetLeft)
	}

	abstract := staticBigIntBinary(
		ast.KindAsteriskToken,
		staticBigIntValue{value: value},
		staticBigIntValue{value: value},
	)
	abstractValue, ok := abstract.value.(staticBigIntValue)
	if !abstract.ok || !ok || abstractValue.value != nil || abstract.bigIntWorkBytes <= cost {
		t.Fatalf("oversized multiplication metadata = %#v", abstract)
	}
	staticEvaluator.bigIntWorkBudgetLeft = abstract.bigIntWorkBytes - 1
	if staticEvaluator.chargeStaticBigIntResult(&abstract) {
		t.Fatal("an abstracted multiplication bypassed its transient BigInt work charge")
	}
	if staticEvaluator.bigIntWorkBudgetLeft != 0 {
		t.Fatalf("transient BigInt work left after exhaustion = %d, want 0", staticEvaluator.bigIntWorkBudgetLeft)
	}
}

func TestStaticBigIntComparisonWorkBudget(t *testing.T) {
	value := new(big.Int).Lsh(big.NewInt(1), maxStaticBigIntBits-1)
	byteLength := (value.BitLen() + 7) / 8
	workLeft := byteLength*2 - 1
	work := &staticStringWorkContext{remaining: &workLeft}
	left := staticBigIntValue{value: value, bigIntWorkContext: work}
	right := staticBigIntValue{value: new(big.Int).Set(value), bigIntWorkContext: work}
	if equal, known := staticValuesStrictEqual(left, right); !known || !equal {
		t.Fatal("the first exact BigInt comparison unexpectedly exhausted its work budget")
	}
	if equal, known := staticValuesStrictEqual(left, right); known || equal {
		t.Fatal("a repeated exact BigInt comparison bypassed its persistent work budget")
	}
	if workLeft != 0 {
		t.Fatalf("BigInt comparison work left after exhaustion = %d, want 0", workLeft)
	}

	for _, test := range []struct {
		integer int64
		number  float64
		want    int
	}{
		{integer: 1, number: 1, want: 0},
		{integer: 1, number: 1.5, want: -1},
		{integer: -1, number: -1.5, want: 1},
	} {
		comparison, ordered, known := staticValuesCompare(
			staticBigIntValue{value: big.NewInt(test.integer)},
			staticNumberValue(test.number),
		)
		if !known || !ordered || comparison != test.want {
			t.Fatalf("BigInt(%d) compare %v = (%d, %v, %v), want (%d, true, true)",
				test.integer, test.number, comparison, ordered, known, test.want)
		}
	}

	huge := new(big.Int).Lsh(big.NewInt(1), 2048)
	for _, test := range []struct {
		value *big.Int
		want  float64
	}{
		{value: huge, want: math.Inf(1)},
		{value: new(big.Int).Neg(huge), want: math.Inf(-1)},
	} {
		result := staticNumberCall(staticBigIntValue{value: test.value})
		number, ok := result.value.(staticNumberValue)
		if !result.ok || !ok || float64(number) != test.want {
			t.Fatalf("Number(large BigInt) = (%#v, %v), want (%v, true)", result.value, result.ok, test.want)
		}
	}
}

func TestStaticEvaluatorDerivedStringBudget(t *testing.T) {
	staticEvaluator := &StaticStringEvaluator{
		evaluationGeneration: 1,
		stringBudgetLeft:     maxStaticStringBudget,
		stringWorkBudgetLeft: maxStaticStringWorkBudget,
	}
	result := staticEvalResult{
		value:                strings.Repeat("x", maxStaticStringLength),
		ok:                   true,
		stringCodeUnits:      maxStaticStringLength,
		stringCodeUnitsKnown: true,
	}
	if !staticEvaluator.chargeStaticStringResult(&result) {
		t.Fatal("first derived string unexpectedly exhausted the budget")
	}
	afterFirstCharge := staticEvaluator.stringBudgetLeft
	if !staticEvaluator.chargeStaticStringResult(&result) ||
		staticEvaluator.stringBudgetLeft != afterFirstCharge {
		t.Fatal("the same derived result was charged twice")
	}
	for range maxStaticStringBudget/maxStaticStringLength - 1 {
		fresh := result
		fresh.stringBudgetGeneration = 0
		if !staticEvaluator.chargeStaticStringResult(&fresh) {
			t.Fatal("derived string budget was exhausted too early")
		}
	}
	fresh := result
	fresh.stringBudgetGeneration = 0
	if staticEvaluator.chargeStaticStringResult(&fresh) {
		t.Fatal("derived string budget accepted an additional oversized result")
	}

	staticEvaluator.evaluationGeneration++
	staticEvaluator.stringBudgetLeft = maxStaticStringBudget
	workBudget := staticEvaluator.stringWorkBudgetLeft
	if !staticEvaluator.chargeStaticStringResult(&result) ||
		staticEvaluator.stringWorkBudgetLeft != workBudget {
		t.Fatal("a retained derived string was charged to the file work budget twice")
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

func TestStaticArrayJoinStreamsNestedValuesWithinWorkBudget(t *testing.T) {
	leaf := strings.Repeat("x", 16383)
	innerValues := make([]any, 64)
	for index := range innerValues {
		innerValues[index] = leaf
	}
	inner := staticArrayFromValues(innerValues)

	workLeft := maxStaticStringWorkBudget
	work := &staticStringWorkContext{remaining: &workLeft}
	inner.stringWorkContext = work
	joined, ok := staticArrayJoin(inner, ",")
	if !ok {
		t.Fatal("staticArrayJoin rejected a nested value just below the limit")
	}
	if len(joined) != maxStaticStringLength-1 {
		t.Fatalf("joined length = %d, want %d", len(joined), maxStaticStringLength-1)
	}
	if workLeft >= maxStaticStringWorkBudget {
		t.Fatal("staticArrayJoin did not charge streamed output")
	}

	workLeft = maxStaticStringWorkBudget
	inner.stringWorkContext = work
	outer := staticArrayFromValues([]any{inner, inner})
	outer.stringWorkContext = work
	if _, ok := staticArrayJoin(outer, ","); ok {
		t.Fatal("staticArrayJoin accepted repeated nested output above the limit")
	}
	if workLeft != 0 {
		t.Fatalf("work left after cap failure = %d, want 0", workLeft)
	}
}

func TestStaticArrayJoinPreservesUTF16AndNullishSemantics(t *testing.T) {
	loneSurrogate := ecmascript.StringFromCodeUnits([]uint16{0xD800})
	value := staticArrayFromValues([]any{"😀", staticUndefinedValue{}, loneSurrogate})
	workLeft := maxStaticStringWorkBudget
	value.stringWorkContext = &staticStringWorkContext{remaining: &workLeft}

	got, ok := staticArrayJoin(value, "|")
	if !ok {
		t.Fatal("staticArrayJoin rejected valid UTF-16 input")
	}
	want := "😀||" + loneSurrogate
	if got != want {
		t.Fatalf("staticArrayJoin = %q, want %q", got, want)
	}
	if units := ecmascript.StringCodeUnitCount(got); units != 5 {
		t.Fatalf("joined code units = %d, want 5", units)
	}
}

func TestStaticArrayJoinDepthLimitIsEnhancedOnly(t *testing.T) {
	legacy := staticArrayFromValues([]any{"x"})
	for range maxStaticArrayStringDepth + 100 {
		legacy = staticArrayFromValues([]any{legacy})
	}
	if value, ok := staticArrayJoin(legacy, ","); !ok || value != "x" {
		t.Fatalf("legacy nested join = (%q, %v), want (x, true)", value, ok)
	}

	workLeft := maxStaticStringWorkBudget
	enhanced := legacy
	enhanced.stringWorkContext = &staticStringWorkContext{remaining: &workLeft}
	if _, ok := staticArrayJoin(enhanced, ","); ok {
		t.Fatal("enhanced nested join ignored its depth limit")
	}
	if workLeft != 0 {
		t.Fatalf("enhanced nested join left %d work after its depth limit, want 0", workLeft)
	}
}

func TestStaticPrimitiveFanoutExhaustsSharedWorkBudget(t *testing.T) {
	segment := strings.Repeat("x", 100)
	rawValues := make([]any, 10500)
	for index := range rawValues {
		rawValues[index] = segment
	}
	raw := staticArrayFromValues(rawValues)
	object := &staticObjectValue{enhancedCoercion: true}
	object.addProperty("raw", raw)

	t.Run("String.raw", func(t *testing.T) {
		workLeft := maxStaticStringWorkBudget
		work := &staticStringWorkContext{remaining: &workLeft}
		if result := staticStringRawCall(work, []any{object}); result.ok {
			t.Fatal("String.raw fanout unexpectedly stayed below the output cap")
		}
		if workLeft != 0 {
			t.Fatalf("work left after String.raw cap failure = %d, want 0", workLeft)
		}
	})

	t.Run("String.concat", func(t *testing.T) {
		workLeft := maxStaticStringWorkBudget
		work := &staticStringWorkContext{remaining: &workLeft}
		if result := staticStringConcat(work, "", rawValues); result.ok {
			t.Fatal("String.concat fanout unexpectedly stayed below the output cap")
		}
		if workLeft != 0 {
			t.Fatalf("work left after String.concat cap failure = %d, want 0", workLeft)
		}
	})
}

func TestStaticDateParseKeepsHostLocalTimeAbstract(t *testing.T) {
	for _, text := range []string{
		"2021-03-14T02:30:00",
		"January 2, 2020 12:00:00",
		"Jan 01 1970 00:00:00 GMT+0000 trailing junk",
	} {
		result := staticDateParse(text)
		if !result.ok {
			t.Fatalf("staticDateParse(%q) returned unknown", text)
		}
		if _, abstract := result.value.(staticUnknownNumberValue); !abstract {
			t.Fatalf("staticDateParse(%q) = %#v, want an abstract number", text, result.value)
		}
	}

	result := staticDateParse("2021-03-14T02:30:00Z")
	milliseconds, exact := result.value.(staticNumberValue)
	if !result.ok || !exact || milliseconds != 1615689000000 {
		t.Fatalf("UTC staticDateParse = (%#v, %v), want (1615689000000, true)", result.value, result.ok)
	}
}

func TestStaticNormalizeKeepsNewerUnicodeMappingsAbstract(t *testing.T) {
	workLeft := maxStaticStringWorkBudget
	work := &staticStringWorkContext{remaining: &workLeft}
	loneSurrogate := ecmascript.StringFromCodeUnits([]uint16{0xD800})
	for _, test := range []struct {
		name  string
		value string
		form  string
	}{
		{name: "compatibility", value: "\U0001CCD6", form: "NFKC"},
		{name: "decomposition", value: "\U000105C9", form: "NFD"},
		{name: "composition", value: "\U000105D2\u0307", form: "NFC"},
		{name: "combining class", value: "A\u0345\u0897", form: "NFD"},
		{name: "mixed surrogate", value: loneSurrogate + "\U0001CCD6", form: "NFKC"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := staticStringPrototypeCall(work, test.value, "normalize", []any{test.form})
			abstract, ok := result.value.(staticUnknownStringValue)
			if !result.ok || !ok || !abstract.truthy {
				t.Fatalf("normalize(%q, %q) = (%#v, %v), want a truthy abstract string", test.value, test.form, result.value, result.ok)
			}
		})
	}

	result := staticStringPrototypeCall(work, "e\u0301", "normalize", []any{"NFC"})
	if value, ok := result.value.(string); !result.ok || !ok || value != "é" {
		t.Fatalf("Unicode 15 normalize control = (%#v, %v), want (é, true)", result.value, result.ok)
	}

	if !staticNormalizationNeedsNewerUnicodeTables("é", "NFC", "17.0.0") {
		t.Fatal("a future normalization table kept non-ASCII output exact")
	}
	if staticNormalizationNeedsNewerUnicodeTables("ASCII", "NFC", "17.0.0") {
		t.Fatal("a future normalization table made ASCII output abstract")
	}
}

func TestStaticNormalizePreservesSurrogateBarriers(t *testing.T) {
	high := ecmascript.StringFromCodeUnits([]uint16{0xD800})
	low := ecmascript.StringFromCodeUnits([]uint16{0xDC00})
	for _, test := range []struct {
		name  string
		value string
		form  string
		want  string
	}{
		{name: "only surrogate", value: high, form: "NFC", want: high},
		{name: "normalize after high", value: high + "\u212A", form: "NFC", want: high + "K"},
		{name: "normalize before low", value: "\u212A" + low, form: "NFC", want: "K" + low},
		{name: "compose after high", value: high + "A\u030A", form: "NFC", want: high + "Å"},
		{name: "compose before low", value: "A\u030A" + low, form: "NFC", want: "Å" + low},
		{name: "compatibility after high", value: high + "\uFB01", form: "NFKC", want: high + "fi"},
		{name: "two barriers", value: "\u212A" + high + "\u212A" + low + "\u212A", form: "NFC", want: "K" + high + "K" + low + "K"},
	} {
		t.Run(test.name, func(t *testing.T) {
			workLeft := maxStaticStringWorkBudget
			result := staticStringPrototypeCall(
				&staticStringWorkContext{remaining: &workLeft},
				test.value,
				"normalize",
				[]any{test.form},
			)
			value, ok := result.value.(string)
			if !result.ok || !ok || value != test.want {
				t.Fatalf("normalize(%q, %q) = (%q, %v), want (%q, true)", test.value, test.form, value, result.ok, test.want)
			}
		})
	}
}

func TestStaticCaseConversionKeepsNewerUnicodeSemanticsAbstract(t *testing.T) {
	for _, test := range []struct {
		name   string
		value  string
		method string
	}{
		{name: "lowercase mapping", value: "\uA7CE", method: "toLowerCase"},
		{name: "uppercase mapping", value: "\uA7CF", method: "toUpperCase"},
		{name: "Beria mapping", value: "\U00016EA0", method: "toLowerCase"},
		{name: "cased context", value: "\u0295Σ", method: "toLowerCase"},
		{name: "ignorable context", value: "AΣ\u1ACFB", method: "toLowerCase"},
	} {
		t.Run(test.name, func(t *testing.T) {
			workLeft := maxStaticStringWorkBudget
			result := staticStringPrototypeCall(
				&staticStringWorkContext{remaining: &workLeft}, test.value, test.method, nil,
			)
			abstract, ok := result.value.(staticUnknownStringValue)
			if !result.ok || !ok || !abstract.truthy {
				t.Fatalf("%s(%q) = (%#v, %v), want a truthy abstract string", test.method, test.value, result.value, result.ok)
			}
		})
	}

	workLeft := maxStaticStringWorkBudget
	result := staticStringPrototypeCall(
		&staticStringWorkContext{remaining: &workLeft}, "\U00010D70", "toUpperCase", nil,
	)
	if value, ok := result.value.(string); !result.ok || !ok || value != "\U00010D50" {
		t.Fatalf("Unicode 16 case control = (%#v, %v), want (U+10D50, true)", result.value, result.ok)
	}

	result = staticStringPrototypeCall(
		&staticStringWorkContext{remaining: &workLeft}, "\u1ACF", "toLowerCase", nil,
	)
	if value, ok := result.value.(string); !result.ok || !ok || value != "\u1ACF" {
		t.Fatalf("context-free identity control = (%#v, %v), want (U+1ACF, true)", result.value, result.ok)
	}
}

func TestStaticRegExpConstructorPatternValidation(t *testing.T) {
	for _, test := range []struct {
		name    string
		pattern string
		flags   string
		want    bool
	}{
		{name: "quantifier", pattern: `a{1,2}`, flags: "u", want: true},
		{name: "raw closing bracket in Annex B", pattern: `]`, want: true},
		{name: "general category", pattern: `\p{L}`, flags: "u", want: true},
		{name: "script property", pattern: `\p{Script=Greek}`, flags: "u", want: true},
		{name: "unicode sets", pattern: `[[A--B]]`, flags: "v", want: true},
		{name: "named backreference", pattern: `(?<a>a)\k<a>`, flags: "u", want: true},
		{name: "escaped named capture", pattern: `(?<\u0061>a)`, flags: "u", want: true},
		{name: "extended escaped named capture", pattern: `(?<\u{1d49c}>a)`, want: true},
		{name: "Annex B named identity escape", pattern: `\k<a>`, want: true},
		{name: "Annex B property identity escape", pattern: `\p{IsGreek}`, want: true},
		{name: "Annex B braced u identity escape in class", pattern: `(?<a>a)[(?<\u{61}>)]`, want: true},
		{name: "Unicode 16 script", pattern: `\p{Script=Garay}`, flags: "u", want: true},
		{name: "distinct nested names", pattern: `(?:(?<a>a))(?<b>b)`, flags: "u", want: true},
		{name: "raw closing bracket under u", pattern: `]`, flags: "u"},
		{name: "reversed quantifier", pattern: `a{2,1}`},
		{name: "range bounded by class escape", pattern: `[a-\d]`, flags: "u"},
		{name: "quantified assertion", pattern: `(?=a)+`, flags: "u"},
		{name: "inline modifier unsupported by Node 22", pattern: `(?i:a)`},
		{name: "negative inline modifier unsupported by Node 22", pattern: `(?-i:a)`},
		{name: "duplicate nested name", pattern: `(?:(?<a>a))(?<a>b)`},
		{name: "duplicate escaped name", pattern: `(?<a>a)|(?<\u0061>b)`, flags: "u"},
		{name: "missing named backreference", pattern: `\k<a>`, flags: "u"},
		{name: "unknown named backreference", pattern: `(?<a>a)\k<b>`},
		{name: "dotnet atomic group", pattern: `(?>a)`},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := staticRegExpConstructorPatternValid(test.pattern, test.flags); got != test.want {
				t.Fatalf("staticRegExpConstructorPatternValid(%q, %q) = %v, want %v", test.pattern, test.flags, got, test.want)
			}
		})
	}
}

func TestStaticRegExpConstructorConversionEarlyErrors(t *testing.T) {
	workLeft := maxStaticStringWorkBudget
	work := &staticStringWorkContext{remaining: &workLeft}
	for _, pattern := range []string{"\\\n", "\\\r", "\\\u2028", "\\\u2029"} {
		if staticRegExpConstructorSourceValid(pattern, "u") {
			t.Fatalf("staticRegExpConstructorSourceValid(%q, %q) = true", pattern, "u")
		}
		if !staticRegExpConstructorSourceValid(pattern, "") {
			t.Fatalf("staticRegExpConstructorSourceValid(%q, %q) = false", pattern, "")
		}
		if _, ok := staticRegExpSource(work, pattern); !ok {
			t.Fatalf("staticRegExpSource(%q) unexpectedly failed in Annex B mode", pattern)
		}
	}
	for _, pattern := range []string{"\n", "\\\\\n", "\r", "\\\\\u2028"} {
		if _, ok := staticRegExpSource(work, pattern); !ok {
			t.Fatalf("staticRegExpSource(%q) unexpectedly failed", pattern)
		}
	}
	for _, pattern := range []string{"[/]", "[[/]]", `[\q{a/b}]`} {
		if staticRegExpConstructorSourceValid(pattern, "v") {
			t.Fatalf("staticRegExpConstructorSourceValid(%q, %q) = true", pattern, "v")
		}
	}
	for _, pattern := range []string{`[\/]`, `[a/b]`, `a/b`} {
		flags := "v"
		want := pattern != `[a/b]`
		if got := staticRegExpConstructorSourceValid(pattern, flags); got != want {
			t.Fatalf("staticRegExpConstructorSourceValid(%q, %q) = %v, want %v", pattern, flags, got, want)
		}
	}
	for _, flags := range []string{"u", "v"} {
		for _, property := range []string{"Hrkt", "Katakana_Or_Hiragana"} {
			pattern := `\p{Script=` + property + `}`
			if staticRegExpConstructorSourceValid(pattern, flags) {
				t.Fatalf("staticRegExpConstructorSourceValid(%q, %q) = true", pattern, flags)
			}
		}
		if !staticRegExpConstructorSourceValid(`\p{Script=Greek}`, flags) {
			t.Fatalf("staticRegExpConstructorSourceValid(%q, %q) = false", `\p{Script=Greek}`, flags)
		}
		for _, pattern := range []string{`\p{`, `\P{`, `\p{\p{L}`} {
			if staticRegExpConstructorSourceValid(pattern, flags) {
				t.Fatalf("staticRegExpConstructorSourceValid(%q, %q) = true", pattern, flags)
			}
		}
	}
	for _, pattern := range []string{`\p{`, `\P{`} {
		if !staticRegExpConstructorSourceValid(pattern, "") {
			t.Fatalf("legacy staticRegExpConstructorSourceValid(%q, %q) = false", pattern, "")
		}
	}
	if !staticRegExpConstructorSourceValid(`\\p{`, "u") {
		t.Fatal("an escaped backslash was mistaken for a Unicode property escape")
	}
	if staticRegExpConstructorSourceValid(`\\\p{Script=Hrkt}`, "u") {
		t.Fatal("a Unicode property escape after an escaped backslash was skipped")
	}
}

func TestStaticRegExpMalformedDelimiterScanning(t *testing.T) {
	malformedProperties := strings.Repeat(`\p{`, 200_000)
	if staticRegExpConstructorSourceValid(malformedProperties, "u") {
		t.Fatal("malformed Unicode properties were accepted")
	}
	if !staticRegExpConstructorSourceValid(malformedProperties, "") {
		t.Fatal("legacy property identity escapes were rejected")
	}
	if got := staticRegExpValidatorSource(malformedProperties, "u", false); got != malformedProperties {
		t.Fatal("malformed Unicode properties were rewritten by the validator bridge")
	}

	malformedBackreferences := strings.Repeat(`\k<`, 200_000)
	if got := staticRegExpNormalizeGroupNames(malformedBackreferences, false); got != malformedBackreferences {
		t.Fatal("malformed legacy backreferences were changed during group-name normalization")
	}
	escapedGroupThenMalformed := `(?<\u0061>a)` + strings.Repeat(`(?<`, 200_000)
	normalized := staticRegExpNormalizeGroupNames(escapedGroupThenMalformed, false)
	if !strings.HasPrefix(normalized, `(?<a>a)`) ||
		!strings.HasSuffix(normalized, strings.Repeat(`(?<`, 2)) {
		t.Fatal("a malformed suffix discarded an earlier group-name normalization")
	}
}

func TestStaticRegExpConstructorEngineLimits(t *testing.T) {
	if !staticRegExpConstructorSourceValid(strings.Repeat("(a)", maxStaticRegExpCaptures), "") {
		t.Fatal("maximum supported capture count was rejected")
	}
	if staticRegExpConstructorSourceValid(strings.Repeat("(a)", maxStaticRegExpCaptures+1), "") {
		t.Fatal("capture count above the engine limit was accepted")
	}
	if staticRegExpConstructorSourceValid(strings.Repeat("(?<a>a)", maxStaticRegExpCaptures+1), "") {
		t.Fatal("named captures above the engine limit were accepted")
	}
	for _, test := range []struct {
		name    string
		pattern string
		flags   string
	}{
		{name: "non-capturing groups", pattern: strings.Repeat("(?:a)", maxStaticRegExpCaptures+1)},
		{name: "lookaheads", pattern: strings.Repeat("(?=a)(?!a)", maxStaticRegExpCaptures+1)},
		{name: "lookbehinds", pattern: strings.Repeat("(?<=a)(?<!a)", maxStaticRegExpCaptures+1)},
		{name: "escaped parentheses", pattern: strings.Repeat(`\(`, maxStaticRegExpCaptures+1)},
		{name: "class parentheses", pattern: strings.Repeat(`[(?<a>)]`, maxStaticRegExpCaptures+1)},
		{name: "class strings", pattern: strings.Repeat(`[\q{\(}]`, maxStaticRegExpCaptures+1), flags: "v"},
	} {
		if !staticRegExpConstructorSourceValid(test.pattern, test.flags) {
			t.Fatalf("%s were counted as captures", test.name)
		}
	}
	if !staticRegExpConstructorSourceValid(
		strings.Repeat("[", maxStaticRegExpUnicodeSetsNesting)+"a"+
			strings.Repeat("]", maxStaticRegExpUnicodeSetsNesting),
		"v",
	) {
		t.Fatal("maximum supported Unicode-sets nesting was rejected")
	}
	if staticRegExpConstructorSourceValid(
		strings.Repeat("[", maxStaticRegExpUnicodeSetsNesting+1)+"a"+
			strings.Repeat("]", maxStaticRegExpUnicodeSetsNesting+1),
		"v",
	) {
		t.Fatal("Unicode-sets nesting above the engine limit was accepted")
	}
}

func TestStaticSymbolDescriptionsAndHostDependentIdentity(t *testing.T) {
	iterator := staticSymbolValue{description: "iterator", wellKnown: true}
	if description, known := staticSymbolDescription(iterator); !known || description != "Symbol.iterator" {
		t.Fatalf("well-known description = (%q, %v), want (%q, true)", description, known, "Symbol.iterator")
	}
	registry := staticSymbolValue{description: "iterator", global: true}
	if description, known := staticSymbolDescription(registry); !known || description != "iterator" {
		t.Fatalf("registry description = (%q, %v), want (%q, true)", description, known, "iterator")
	}
	dispose := staticSymbolValue{description: "dispose", hostDependent: true}
	if _, known := staticSymbolDescription(dispose); known {
		t.Fatal("Symbol.dispose unexpectedly had a host-independent description")
	}
	if equal, known := staticSymbolIdentityEqual(dispose, dispose); !known || !equal {
		t.Fatalf("Symbol.dispose self identity = (%v, %v), want (true, true)", equal, known)
	}
	node22Dispose := staticSymbolValue{description: "nodejs.dispose", global: true}
	if _, known := staticSymbolIdentityEqual(dispose, node22Dispose); known {
		t.Fatal("Symbol.dispose registry identity unexpectedly ignored the host version")
	}
}

func TestStaticCollectionBigIntKeysUseBoundedBinaryWork(t *testing.T) {
	value := new(big.Int).Lsh(big.NewInt(1), maxStaticBigIntBits-1)
	byteLength := (value.BitLen() + 7) / 8
	workLeft := byteLength * 3
	collection := staticCollectionValue{
		kind: "Set", identity: &staticIdentity{},
		stringWorkContext: &staticStringWorkContext{remaining: &workLeft},
	}
	key := staticBigIntValue{value: value}
	if !staticCollectionStore(&collection, key, nil) {
		t.Fatal("the first bounded bigint key was rejected")
	}
	if staticCollectionStore(&collection, key, nil) {
		t.Fatal("a repeated bigint key was materialized after exhausting work")
	}
	if workLeft != 0 {
		t.Fatalf("work left after bigint key cap = %d, want 0", workLeft)
	}

	equalKey, equalOK := staticCanonicalCollectionKey(
		staticBigIntValue{value: new(big.Int).Set(value)},
		nil,
	)
	originalKey, originalOK := staticCanonicalCollectionKey(key, nil)
	negativeKey, negativeOK := staticCanonicalCollectionKey(
		staticBigIntValue{value: new(big.Int).Neg(value)},
		nil,
	)
	if !equalOK || !originalOK || equalKey != originalKey {
		t.Fatal("separately materialized equal bigints did not share a collection key")
	}
	if !negativeOK || negativeKey == originalKey {
		t.Fatal("opposite-sign bigints shared a collection key")
	}

	textWorkLeft := len("large key") - 1
	if _, ok := staticCanonicalCollectionKey(
		"large key",
		&staticStringWorkContext{remaining: &textWorkLeft},
	); ok || textWorkLeft != 0 {
		t.Fatal("an oversized string key bypassed collection hashing work")
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
