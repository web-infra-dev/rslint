package no_dupe_keys

import (
	"strconv"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeKeysRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeKeysRule,
		[]rule_tester.ValidTestCase{
			{Code: `var x = { a: 1, b: 2 };`},
			{Code: `var x = { a: 1, b: 2, c: 3 };`},
			{Code: `var x = { get a() {}, set a(v) {} };`},
			{Code: `var x = { set a(v) {}, get a() {} };`},
			{Code: `var x = { [Symbol()]: 1, [Symbol()]: 2 };`},
			{Code: `var x = { "": 1, " ": 2 };`},
			// __proto__ as proto setter is allowed to appear multiple times
			{Code: `var x = { __proto__: foo, __proto__: bar };`},
			{Code: `var x = { "__proto__": foo, "__proto__": bar };`},
			{Code: `var x = { __proto__: null, ["__proto__"]: null };`},
			{Code: `var x = { ["__proto__"]: null, __proto__: null };`},
			// Assignment destructuring is represented as an object literal by tsgo,
			// but as an ObjectPattern skipped by the upstream rule.
			{Code: `({ a: first, a: second } = source);`},
			{Code: `({ ["a"]: first, ["a"]: second } = source);`},
			{Code: `((({ nested: { a: first, a: second } }))) = source;`},
			{Code: `([{ a: first, a: second }] = source);`},
			{Code: `for ({ a: first, a: second } of source) {}`},
			{Code: `for ({ a: first, a: second } in source) {}`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var x = { a: 1, a: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: `var x = { a: 1, b: 2, a: 3 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `var x = { a: 1, b: 2, a: 3, a: 4 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
					{MessageId: "unexpected", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: `var x = { "a": 1, "a": 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: `var x = { "a": 1, a: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: `var x = { get a() {}, set a(v) {}, get a() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 40, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code: `var x = { set a(v) {}, get a() {}, set a(v) {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 40, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code: `var x = { get a() {}, set a(v) {}, a: 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code: `var x = { a: 1, get a() {}, set a(v) {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
					{MessageId: "unexpected", Line: 1, Column: 33, EndLine: 1, EndColumn: 34},
				},
			},
			// Computed __proto__ is a regular property, not a proto setter
			{
				Code: `var x = { ["__proto__"]: 1, ["__proto__"]: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 30, EndLine: 1, EndColumn: 41},
				},
			},
			// Numeric literal equivalence: 0x1 and 1 are the same key
			{
				Code: `var x = { 0x1: "a", 1: "b" };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
				},
			},
			// BigInt literal equivalence: 0x1n and 1n normalize to the same key
			{
				Code: `var x = { [0x1n]: "a", [1n]: "b" };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25, EndLine: 1, EndColumn: 27},
				},
			},
			// Template literal computed property
			{
				Code: "var x = { [`key`]: 1, [`key`]: 2 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24, EndLine: 1, EndColumn: 29},
				},
			},
			// Numeric overflow to Infinity: 1e309 and 1e999 both normalize to "Infinity"
			{
				Code: `var x = { [1e309]: "a", [1e999]: "b" };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26, EndLine: 1, EndColumn: 31},
				},
			},
			{
				// A TypeScript-only wrapper prevents the object from becoming an
				// ESTree ObjectPattern, so the upstream rule still checks it.
				Code: `({ a: first, a: second }! = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `({ value = { a: 1, a: 2 } } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: `({ [{ a: 1, a: 2 }]: target } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `target = ({ a: 1, a: 2 });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: `({ a: first, a: second } as any) = source;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `var x = { a: 1, ["a"]: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var x = { 1n: "a", 1: "b" };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var x = { 1_0: "a", 10: "b" };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var x = { "/x/g": "a", [/x/g]: "b" };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var x = { ["__proto__"]: null, __proto__ };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
		},
	)
}

type referencePropertyInfo struct {
	get  bool
	set  bool
	init bool
}

func registerReferenceProperty(info referencePropertyInfo, kind propertyKind) (referencePropertyInfo, bool) {
	var duplicate bool
	switch kind {
	case propertyKindGet:
		duplicate = info.get || info.init
		info.get = true
	case propertyKindSet:
		duplicate = info.set || info.init
		info.set = true
	default:
		duplicate = info.init || info.get || info.set
		info.init = true
	}
	return info, duplicate
}

func TestRegisterPropertyExhaustive(t *testing.T) {
	const operationCount = 9
	var visit func(propertyState, referencePropertyInfo, int)
	visit = func(state propertyState, reference referencePropertyInfo, depth int) {
		if depth == operationCount {
			return
		}
		for kind := propertyKindInit; kind <= propertyKindSet; kind++ {
			gotState, gotDuplicate := registerProperty(state, kind)
			wantState, wantDuplicate := registerReferenceProperty(reference, kind)
			if gotDuplicate != wantDuplicate {
				t.Fatalf("depth %d kind %d: duplicate = %v, want %v", depth, kind, gotDuplicate, wantDuplicate)
			}
			visit(gotState, wantState, depth+1)
		}
	}
	visit(0, referencePropertyInfo{}, 0)
}

func TestObjectStateMatchesReference(t *testing.T) {
	nameCount := inlineObjectStateCapacity*2 + 3
	state := objectState{overflowCapacityHint: nameCount}
	reference := make(map[string]referencePropertyInfo, nameCount)
	seed := uint64(0x9e3779b97f4a7c15)

	for operation := range 10_000 {
		seed = seed*6364136223846793005 + 1
		nameIndex := operation
		if nameIndex >= nameCount {
			nameIndex = int(seed % uint64(nameCount))
		}
		var name string
		switch nameIndex {
		case 0:
			name = ""
		case 1:
			name = "__proto__"
		default:
			name = "property" + strconv.Itoa(nameIndex)
		}
		kind := propertyKind(seed % 3)

		wantInfo, wantDuplicate := registerReferenceProperty(reference[name], kind)
		reference[name] = wantInfo
		if gotDuplicate := state.register(name, kind); gotDuplicate != wantDuplicate {
			t.Fatalf("operation %d (%q, kind=%d): duplicate = %v, want %v", operation, name, kind, gotDuplicate, wantDuplicate)
		}
		if operation < inlineObjectStateCapacity && state.overflow != nil {
			t.Fatalf("operation %d: state promoted before inline capacity was exhausted", operation)
		}
		if operation == inlineObjectStateCapacity && state.overflow == nil {
			t.Fatal("state did not promote after inline capacity was exhausted")
		}
	}

	if state.overflow == nil {
		t.Fatal("adversarial sequence did not exercise overflow storage")
	}
	for name, info := range reference {
		var want propertyState
		if info.init {
			want |= propertyStateInit
		}
		if info.get {
			want |= propertyStateGet
		}
		if info.set {
			want |= propertyStateSet
		}
		if got := state.overflow[name]; got != want {
			t.Fatalf("state for %q = %d, want %d", name, got, want)
		}
	}
}
