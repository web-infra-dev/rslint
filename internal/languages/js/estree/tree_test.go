package estree

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestWithTreeBasicsRangesAndCleanup(t *testing.T) {
	var builder Builder

	// Multi-byte UTF-8 comment + Unicode identifier: Tree ranges must be exact UTF-8 byte offsets.
	text := "// 😀\nexport const 变量: unknown = v as string;\n"
	file := source("/unicode.ts", text)

	var seenTypes []string
	var unknownRange [2]int
	var asRange [2]int
	var rootParent NodeID = 999

	err := builder.WithTree(file, TreeOptions{SourceType: "module"}, func(tree *Tree) {
		if tree.Root() == NoNode {
			t.Fatal("expected non-empty root")
		}
		rootParent = tree.Parent(tree.Root())

		var walk func(NodeID)
		walk = func(id NodeID) {
			tid := tree.Type(id)
			seenTypes = append(seenTypes, tid.String())
			r := tree.Range(id)
			switch tid {
			case typeTSUnknownKeyword:
				unknownRange = [2]int{r.Pos(), r.End()}
			case typeTSAsExpression:
				asRange = [2]int{r.Pos(), r.End()}
				// Verify Parent/ParentField and Field lookup.
				parent := tree.Parent(id)
				if tree.Type(parent) != typeVariableDeclarator {
					t.Fatalf("TSAsExpression parent = %s, want VariableDeclarator", tree.Type(parent))
				}
				initField, _ := FieldIDByName("init")
				if tree.ParentField(id) != initField {
					t.Fatalf("TSAsExpression parentField = %s, want init", tree.ParentField(id))
				}
				typeAnnField, _ := FieldIDByName("typeAnnotation")
				ann := tree.Field(id, typeAnnField)
				if ann.Kind != ValueNode || tree.Type(ann.Node) != typeTSStringKeyword {
					t.Fatalf("TSAsExpression.typeAnnotation = %+v, want TSStringKeyword", ann)
				}
			}
			tree.ForEachChild(id, func(child NodeID) bool {
				if tree.Parent(child) != id {
					t.Fatalf("child %s parent mismatch", tree.Type(child))
				}
				walk(child)
				return false
			})
		}
		walk(tree.Root())
	})
	if err != nil {
		t.Fatalf("WithTree failed: %v", err)
	}
	if rootParent != NoNode {
		t.Fatalf("Program parent = %d, want NoNode", rootParent)
	}
	if got := text[unknownRange[0]:unknownRange[1]]; got != "unknown" {
		t.Fatalf("TSUnknownKeyword slice = %q, want %q", got, "unknown")
	}
	if got := text[asRange[0]:asRange[1]]; got != "v as string" {
		t.Fatalf("TSAsExpression slice = %q, want %q", got, "v as string")
	}
	// Deferred clearSource must have wiped all staged source strings.
	assertClearedProjectionStrings(t, &builder)
	if builder.p.file != nil || builder.p.text != "" || builder.p.triviaStarts != nil {
		t.Fatal("WithTree retained source references after return")
	}
	if len(seenTypes) == 0 {
		t.Fatal("expected visited nodes")
	}
}

func TestWithTreePanicAndErrorCleanup(t *testing.T) {
	var builder Builder

	// Invalid sourceType returns error without corrupting builder.
	if err := builder.WithTree(source("/a.ts", "const x = 1;"), TreeOptions{SourceType: "bogus"}, func(*Tree) {
		t.Fatal("callback must not run on invalid sourceType")
	}); err == nil {
		t.Fatal("expected error for invalid sourceType")
	}

	// Resilient treeMode: unsupported or invalid-type constructs (JSDoc types,
	// import(Name), malformed literals/headers) must degrade locally rather than
	// aborting the file and silencing diagnostics on later statements.
	sourceField, _ := FieldIDByName("source")
	for _, tc := range []struct {
		name      string
		code      string
		wantTypes []TypeID
	}{
		{"/import_name.ts", "type Name = 'pkg';\ntype T = import(Name);\ndebugger;", []TypeID{typeTSImportType, typeDebuggerStatement}},
		{"/import_null.ts", "type D = import(null);\ndebugger;", []TypeID{typeTSImportType, typeDebuggerStatement}},
		{"/import_paren_null.ts", "type D = import((null)).Foo;\ndebugger;", []TypeID{typeTSImportType, typeDebuggerStatement}},
		{"/import_paren_name.ts", "type Name = 'pkg';\ntype T = import((Name)).Foo;\ndebugger;", []TypeID{typeTSImportType, typeDebuggerStatement}},
		{"/import_template.ts", "type T = import(`a${string}b`);\ndebugger;", []TypeID{typeTSImportType, typeDebuggerStatement}},
		{"/jsdoc_nullable.ts", "type T = ?number;\ndebugger;", []TypeID{typeTSTypeAliasDeclaration, typeDebuggerStatement}},
		{"/bad_bigint.ts", "const x = 0xn;\ndebugger;", []TypeID{typeDebuggerStatement}},
		{"/leading_zero_bigint.ts", "const x = 0_10n;\ndebugger;", []TypeID{typeLiteral, typeDebuggerStatement}},
		{"/bad_num.ts", "const x = 0x;\ndebugger;", []TypeID{typeDebuggerStatement}},
		{"/missing_brace.ts", "class C\ndebugger;", []TypeID{typeDebuggerStatement}},
	} {
		if err := builder.WithTree(source(tc.name, tc.code), TreeOptions{SourceType: "module"}, func(tree *Tree) {
			seen := make(map[TypeID]bool)
			var walk func(NodeID)
			walk = func(id NodeID) {
				tid := tree.Type(id)
				seen[tid] = true
				if (tc.name == "/import_name.ts" || tc.name == "/import_null.ts" || tc.name == "/import_paren_null.ts" || tc.name == "/import_paren_name.ts" || tc.name == "/import_template.ts") && tid == typeTSImportType {
					if got := tree.Field(id, sourceField); got.Kind != ValueUndefined {
						t.Fatalf("%s TSImportType.source = %+v, want ValueUndefined", tc.name, got)
					}
				}
				if tc.name == "/leading_zero_bigint.ts" && tid == typeLiteral {
					valueField, _ := FieldIDByName("value")
					if got := tree.Field(id, valueField); got.Kind != ValueNull {
						t.Fatalf("leading-zero BigInt 0_10n must degrade to ValueNull, got %+v", got)
					}
				}
				tree.ForEachChild(id, func(child NodeID) bool {
					walk(child)
					return false
				})
			}
			walk(tree.Root())
			for _, want := range tc.wantTypes {
				if !seen[want] {
					t.Fatalf("%s: expected projected %s, saw %v", tc.name, want, seen)
				}
			}
		}); err != nil {
			t.Fatalf("%s: unexpected error %v", tc.name, err)
		}
		assertClearedProjectionStrings(t, &builder)
	}

	// Malformed import-type options in treeMode must not leave orphan nodes or inverted spans in Tree.
	for _, tc := range []struct {
		name string
		code string
	}{
		{"/bad_import_opts_unclosed.ts", `type T = import("a", { with: { type: "json" }`},
		{"/bad_import_opts_empty.ts", `type T = import(("./m"), { });`},
		{"/bad_import_opts_comma.ts", `type T = import("./m", { , });`},
		{"/bad_import_opts_colon.ts", `type T = import("./m", { : { type: "json" } });`},
		{"/bad_import_unclosed_paren.ts", `type T = import(("./m"`},
		{"/bad_import_empty_paren.ts", `type T = import(())`},
		{"/bad_import_dangling_paren.ts", `type T = import((`},
	} {
		if err := builder.WithTree(source(tc.name, tc.code), TreeOptions{SourceType: "module"}, func(tree *Tree) {
			visited := 0
			var walk func(NodeID)
			walk = func(id NodeID) {
				visited++
				r := tree.Range(id)
				if r.Pos() < 0 || r.End() < r.Pos() || r.End() > len(tc.code)+1 {
					t.Fatalf("%s: node %s has invalid range [%d, %d]", tc.name, tree.Type(id), r.Pos(), r.End())
				}
				tree.ForEachChild(id, func(child NodeID) bool {
					walk(child)
					return false
				})
			}
			walk(tree.Root())
			if visited != tree.NodeCount() {
				t.Fatalf("%s: malformed import left orphan nodes: visited=%d NodeCount=%d", tc.name, visited, tree.NodeCount())
			}
		}); err != nil {
			t.Fatalf("WithTree(%s) failed: %v", tc.name, err)
		}
	}

	// Wire-mode Build() remains strict on unsupported/malformed syntax and non-LiteralType import().
	for _, tc := range []struct {
		name string
		code string
	}{
		{"/wire_jsdoc_nullable.ts", "type T = ?number;"},
		{"/wire_leading_zero_bigint.ts", "const x = 0_10n;"},
		{"/wire_bad_bigint.ts", "const x = 0xn;"},
		{"/wire_import_name.ts", "type Name = 'pkg'; type T = import(Name);"},
		{"/wire_import_paren_name.ts", "type Name = 'pkg'; type T = import((Name)).Foo;"},
		{"/wire_import_template.ts", "type T = import(`a${string}b`); const after = 1; // comment"},
		{"/wire_import_paren_template.ts", "type T = import((`a${string}b`)).Foo; const after = 1;"},
		{"/wire_import_empty_opts.ts", `type T = import(("./m"), { });`},
		{"/wire_import_comma_opts.ts", `type T = import("./m", { , });`},
		{"/wire_import_colon_opts.ts", `type T = import("./m", { : { type: "json" } });`},
	} {
		if _, err := builder.Build(context.Background(), source(tc.name, tc.code), Options{SourceType: "module"}); err == nil {
			t.Fatalf("Build(%s) succeeded, want error in strict wire mode", tc.name)
		}
		assertClearedProjectionStrings(t, &builder)
	}

	// JS JSDoc modifiers/heritage/prologue-declarations must be skipped in both Build and WithTree.
	nameField, _ := FieldIDByName("name")
	keyField, _ := FieldIDByName("key")
	implField, _ := FieldIDByName("implements")
	directiveField, _ := FieldIDByName("directive")
	jsCtor := "class C { /** @private */ constructor() {} }"
	if _, err := builder.Build(context.Background(), source("/jsdoc_ctor.js", jsCtor), Options{SourceType: "module"}); err != nil {
		t.Fatalf("Build(jsdoc_ctor.js) failed: %v", err)
	}
	if err := builder.WithTree(source("/jsdoc_ctor.js", jsCtor), TreeOptions{SourceType: "module"}, func(tree *Tree) {
		foundCtor := false
		var walk func(NodeID)
		walk = func(id NodeID) {
			if tree.Type(id) == typeMethodDefinition {
				k := tree.Field(id, keyField)
				if k.Kind == ValueNode && tree.Field(k.Node, nameField).Str == "constructor" {
					foundCtor = true
				}
			}
			tree.ForEachChild(id, func(child NodeID) bool {
				walk(child)
				return false
			})
		}
		walk(tree.Root())
		if !foundCtor {
			t.Fatal("expected constructor MethodDefinition with Identifier key in JS file with JSDoc @private")
		}
	}); err != nil {
		t.Fatalf("WithTree(jsdoc_ctor.js) failed: %v", err)
	}
	jsImpl := "/** @implements {Foo} */ class C {}"
	if _, err := builder.Build(context.Background(), source("/jsdoc_impl.js", jsImpl), Options{SourceType: "module"}); err != nil {
		t.Fatalf("Build(jsdoc_impl.js) failed: %v", err)
	}
	if err := builder.WithTree(source("/jsdoc_impl.js", jsImpl), TreeOptions{SourceType: "module"}, func(tree *Tree) {
		foundClass := false
		var walk func(NodeID)
		walk = func(id NodeID) {
			if tree.Type(id) == typeClassDeclaration {
				if got := tree.Field(id, implField); got.Kind != ValueUndefined {
					t.Fatalf("JS @implements ClassDeclaration.implements = %+v, want ValueUndefined", got)
				}
				foundClass = true
			}
			if tree.Type(id) == typeTSClassImplements {
				t.Fatal("JS @implements must not emit TSClassImplements")
			}
			tree.ForEachChild(id, func(child NodeID) bool {
				walk(child)
				return false
			})
		}
		walk(tree.Root())
		if !foundClass {
			t.Fatal("expected ClassDeclaration")
		}
	}); err != nil {
		t.Fatalf("WithTree(jsdoc_impl.js) failed: %v", err)
	}
	jsDirective := "/** @typedef {number} T */\n\"use strict\";"
	if _, err := builder.Build(context.Background(), source("/jsdoc_directive.js", jsDirective), Options{SourceType: "script"}); err != nil {
		t.Fatalf("Build(jsdoc_directive.js) failed: %v", err)
	}
	if err := builder.WithTree(source("/jsdoc_directive.js", jsDirective), TreeOptions{SourceType: "script"}, func(tree *Tree) {
		foundDirective := false
		var walk func(NodeID)
		walk = func(id NodeID) {
			if tree.Type(id) == typeExpressionStatement {
				if got := tree.Field(id, directiveField); got.Kind == ValueString && got.Str == "use strict" {
					foundDirective = true
				}
			}
			tree.ForEachChild(id, func(child NodeID) bool {
				walk(child)
				return false
			})
		}
		walk(tree.Root())
		if !foundDirective {
			t.Fatal("expected leading @typedef followed by \"use strict\" to preserve ExpressionStatement.directive")
		}
	}); err != nil {
		t.Fatalf("WithTree(jsdoc_directive.js) failed: %v", err)
	}

	// Degraded type annotation must not stretch Identifier.end past the identifier token.
	varAnnCode := "let x: ?number;"
	if err := builder.WithTree(source("/degraded_ann.ts", varAnnCode), TreeOptions{SourceType: "module"}, func(tree *Tree) {
		found := false
		var walk func(NodeID)
		walk = func(id NodeID) {
			if tree.Type(id) == typeIdentifier && tree.Field(id, nameField).Str == "x" {
				r := tree.Range(id)
				if got := varAnnCode[r.Pos():r.End()]; got != "x" {
					t.Fatalf("degraded annotation stretched Identifier range to %q, want %q", got, "x")
				}
				found = true
			}
			tree.ForEachChild(id, func(child NodeID) bool {
				walk(child)
				return false
			})
		}
		walk(tree.Root())
		if !found {
			t.Fatal("expected Identifier x")
		}
	}); err != nil {
		t.Fatalf("WithTree(degraded_ann.ts) failed: %v", err)
	}

	// Panic inside fn must NOT be swallowed by WithTree, and must still clear strings.
	sentinel := errors.New("callback panic")
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = builder.WithTree(source("/ok.ts", "const s = 'hello';"), TreeOptions{SourceType: "module"}, func(*Tree) {
			panic(sentinel)
		})
	}()
	if err, ok := recovered.(error); !ok || !errors.Is(err, sentinel) {
		t.Fatalf("recovered = %v, want %v", recovered, sentinel)
	}
	assertClearedProjectionStrings(t, &builder)
	if builder.p.file != nil || builder.p.text != "" {
		t.Fatal("WithTree retained source references after callback panic")
	}
}

func TestWithTreeResetBoundsLinksAndScratch(t *testing.T) {
	text := strings.Repeat("const value = { key: 'text' };\n", 20_000)
	var builder Builder
	if err := builder.WithTree(source("/large.ts", text), TreeOptions{SourceType: "module"}, func(tree *Tree) {
		if tree.NodeCount() == 0 {
			t.Fatal("expected nodes")
		}
	}); err != nil {
		t.Fatal(err)
	}
	builder.Reset()
	retained := uintptr(0)
	for _, buffer := range []any{
		builder.p.nodes,
		builder.p.values,
		builder.p.lists,
		builder.p.strings,
		builder.p.numbers,
		builder.tokens,
		builder.comments,
		builder.p.tokenCovers,
		builder.p.offsets.entries,
		builder.p.links,
		builder.p.listScratch,
	} {
		value := reflect.ValueOf(buffer)
		retained += uintptr(value.Cap()) * value.Type().Elem().Size()
	}
	if retained > 4<<20 {
		t.Fatalf("oversized scratch survived Reset: %d bytes", retained)
	}
	assertClearedProjectionStrings(t, &builder)
}

func TestWithTreeLiteralsTemplatesSparseAndUndefined(t *testing.T) {
	var builder Builder

	// 1. Sparse array holes and listOffset == 0 (first list emitted in a minimal file).
	if err := builder.WithTree(source("/sparse.ts", "[1, , 3];"), TreeOptions{SourceType: "module"}, func(tree *Tree) {
		var literals []NodeID
		var walk func(NodeID)
		walk = func(id NodeID) {
			if tree.Type(id) == typeLiteral {
				literals = append(literals, id)
			}
			tree.ForEachChild(id, func(child NodeID) bool {
				walk(child)
				return false
			})
		}
		walk(tree.Root())
		if len(literals) != 2 {
			t.Fatalf("expected 2 literals, got %d", len(literals))
		}
		list, idx0, ok0 := tree.SiblingList(literals[0])
		if !ok0 || idx0 != 0 || list.Len() != 3 || list.At(1) != NoNode {
			t.Fatalf("first literal sibling list = (%d, %v, len=%d, hole=%d)", idx0, ok0, list.Len(), list.At(1))
		}
		_, idx1, ok1 := tree.SiblingList(literals[1])
		if !ok1 || idx1 != 2 {
			t.Fatalf("second literal sibling index = (%d, %v), want (2, true)", idx1, ok1)
		}
	}); err != nil {
		t.Fatal(err)
	}

	// 2. BigInt, RegExp (valid, /v flag, invalid flags), TemplateElement, and hidden $-fields.
	if _, ok := FieldIDByName("$regexPattern"); ok {
		t.Fatal("internal $regexPattern must not be exposed")
	}
	if _, ok := FieldIDByName("$cooked"); ok {
		t.Fatal("internal $cooked must not be exposed")
	}

	valueField, _ := FieldIDByName("value")
	regexField, _ := FieldIDByName("regex")
	code := "const a = 0x10n; const b = /foo/ig; const c = /[a-z]/v; const d = tag`\\unicode`;"
	if err := builder.WithTree(source("/literals.ts", code), TreeOptions{SourceType: "module"}, func(tree *Tree) {
		var sawBigInt, sawRegex, sawVFlag, sawInvalidCooked bool
		var walk func(NodeID)
		walk = func(id NodeID) {
			switch tree.Type(id) {
			case typeLiteral:
				val := tree.Field(id, valueField)
				reg := tree.Field(id, regexField)
				switch val.Kind {
				case ValueBigInt:
					if val.Str != "16" {
						t.Fatalf("bigint value = %q, want 16", val.Str)
					}
					sawBigInt = true
				case ValueRegExp:
					if reg.Kind != ValueRegexObject {
						t.Fatalf("regex literal missing regex object: %+v", reg)
					}
					pat, flags, ok := tree.RegexInfo(id)
					if !ok {
						t.Fatal("RegexInfo returned ok=false")
					}
					if pat == "foo" && flags == "ig" {
						sawRegex = true
					} else if pat == "[a-z]" && flags == "v" {
						sawVFlag = true
					}
				}
			case typeTemplateElement:
				tv := tree.Field(id, valueField)
				if tv.Kind != ValueTemplateValue {
					t.Fatalf("TemplateElement.value kind = %d, want ValueTemplateValue", tv.Kind)
				}
				_, cookedNull, raw, ok := tree.TemplateValue(id)
				if !ok || !cookedNull || raw != `\unicode` {
					t.Fatalf("TemplateValue = (cookedNull=%v, raw=%q, ok=%v)", cookedNull, raw, ok)
				}
				sawInvalidCooked = true
			}
			tree.ForEachChild(id, func(child NodeID) bool {
				walk(child)
				return false
			})
		}
		walk(tree.Root())
		if !sawBigInt || !sawRegex || !sawVFlag || !sawInvalidCooked {
			t.Fatalf("missing literal/template checks: bigint=%v regex=%v v=%v cooked=%v", sawBigInt, sawRegex, sawVFlag, sawInvalidCooked)
		}
	}); err != nil {
		t.Fatal(err)
	}

	// 3. Absent TS fields stay ValueUndefined (not ValueNull), and JS/TS sourceType & JSDoc/rest fixes.
	typeAnnField, _ := FieldIDByName("typeAnnotation")
	returnTypeField, _ := FieldIDByName("returnType")
	accField, _ := FieldIDByName("accessibility")
	phaseField, _ := FieldIDByName("phase")
	sourceTypeField, _ := FieldIDByName("sourceType")
	optionalField, _ := FieldIDByName("optional")

	if err := builder.WithTree(source("/ts_undef.ts", "function f(x) {} class C { m() {} } type F = (...args?: number[]) => void;"), TreeOptions{SourceType: "commonjs"}, func(tree *Tree) {
		if got := tree.Field(tree.Root(), sourceTypeField); got.Kind != ValueString || got.Str != "script" {
			t.Fatalf("TS commonjs Program.sourceType = %+v, want script", got)
		}
		var walk func(NodeID)
		walk = func(id NodeID) {
			switch tree.Type(id) {
			case typeFunctionDeclaration:
				if got := tree.Field(id, returnTypeField); got.Kind != ValueUndefined {
					t.Fatalf("unannotated FunctionDeclaration.returnType = %+v, want ValueUndefined", got)
				}
			case typeMethodDefinition:
				if got := tree.Field(id, accField); got.Kind != ValueUndefined {
					t.Fatalf("MethodDefinition.accessibility = %+v, want ValueUndefined", got)
				}
			case typeIdentifier:
				parent := tree.Parent(id)
				if tree.Type(parent) == typeRestElement {
					if got := tree.Field(id, optionalField); got.Kind != ValueBool || got.Bool {
						t.Fatalf("RestElement.argument.optional = %+v, want false", got)
					}
					if got := tree.Field(id, typeAnnField); got.Kind != ValueUndefined {
						t.Fatalf("unannotated Identifier.typeAnnotation = %+v, want ValueUndefined", got)
					}
				}
			}
			tree.ForEachChild(id, func(child NodeID) bool {
				walk(child)
				return false
			})
		}
		walk(tree.Root())
	}); err != nil {
		t.Fatal(err)
	}

	jsText := "import x from 'a';\nimport defer * as ns from 'b';\nclass C {\n  /** @template T */\n  m(x) {}\n}\n"
	if err := builder.WithTree(source("/mod.cjs", jsText), TreeOptions{SourceType: "commonjs"}, func(tree *Tree) {
		if got := tree.Field(tree.Root(), sourceTypeField); got.Kind != ValueString || got.Str != "commonjs" {
			t.Fatalf("JS commonjs Program.sourceType = %+v, want commonjs", got)
		}
		var sawPlainImport, sawDeferImport bool
		var walk func(NodeID)
		walk = func(id NodeID) {
			switch tree.Type(id) {
			case typeImportDeclaration:
				got := tree.Field(id, phaseField)
				switch got.Kind {
				case ValueUndefined:
					sawPlainImport = true
				case ValueString:
					if got.Str != "defer" {
						t.Fatalf("JS import defer phase = %q, want defer", got.Str)
					}
					sawDeferImport = true
				default:
					t.Fatalf("unexpected JS ImportDeclaration.phase = %+v", got)
				}
			case typeFunctionExpression:
				r := tree.Range(id)
				if got := jsText[r.Pos():r.End()]; got != "(x) {}" {
					t.Fatalf("JS JSDoc @template method FunctionExpression = %q, want %q", got, "(x) {}")
				}
			}
			tree.ForEachChild(id, func(child NodeID) bool {
				walk(child)
				return false
			})
		}
		walk(tree.Root())
		if !sawPlainImport || !sawDeferImport {
			t.Fatalf("JS import phase checks missing: plain=%v defer=%v", sawPlainImport, sawDeferImport)
		}
	}); err != nil {
		t.Fatal(err)
	}

	// 4. ES2025 + Annex B RegExp validation: distinct-alternative duplicate capture
	// names, valid `/v` patterns, and legacy Annex B escapes yield ValueRegExp,
	// while same-alternative duplicates, unknown named backrefs, and invalid `/v`
	// sets yield ValueNull.
	for _, tc := range []struct {
		code string
		want ValueKind
	}{
		{`const r = /(?<k>a)|(?<k>b)/;`, ValueRegExp},
		{`const r = /(?<x>a)(?<x>b)/;`, ValueNull},
		{`const r = /(?:(?<k>a)|(?<k>b))(?<k>c)/;`, ValueNull},
		{`const r = /[[A--B]]/v;`, ValueRegExp},
		{`const r = /[[A&&&]]/v;`, ValueNull},
		{`const r = /(?i:foo)bar/;`, ValueRegExp},
		{`const r = /\p{Letter}/;`, ValueRegExp},
		{`const r = /\k<foo>/;`, ValueRegExp},
		{`const r = /\01/;`, ValueRegExp},
		{`const r = /[\1]/;`, ValueRegExp},
		{`const r = /\u{61}/;`, ValueRegExp},
		{`const r = /\u{NOT_HEX}/;`, ValueRegExp},
		{`const r = /\p{Script=Garay}/v;`, ValueRegExp},
		{`const r = /\p{sc=Gara}/u;`, ValueRegExp},
		// cspell:ignore Beria
		{`const r = /\p{Script=Beria_Erfe}/v;`, ValueRegExp},
		{`const r = /\p{General_Category=Garay}/v;`, ValueNull},
		{`const r = /(?<a>x)\k<b>/;`, ValueNull},
		{`const r = /a/gg;`, ValueNull},
		{`const r = /a/z;`, ValueNull},
		{`const r = /a/uv;`, ValueNull},
		{`const r = /abc`, ValueNull},
		{"const r = /a\u2028b/;", ValueNull},
		{"const r = /a\u2029b/u;", ValueNull},
	} {
		if err := builder.WithTree(source("/regex_valid.ts", tc.code), TreeOptions{SourceType: "module"}, func(tree *Tree) {
			found := false
			var walk func(NodeID)
			walk = func(id NodeID) {
				if tree.Type(id) == typeLiteral {
					if val := tree.Field(id, valueField); val.Kind != tc.want {
						t.Fatalf("%s: Literal.value = %+v, want kind %d", tc.code, val, tc.want)
					}
					found = true
				}
				tree.ForEachChild(id, func(child NodeID) bool {
					walk(child)
					return false
				})
			}
			walk(tree.Root())
			if !found {
				t.Fatalf("%s: expected Literal node", tc.code)
			}
		}); err != nil {
			t.Fatal(err)
		}
	}

	// TS preserves CRLF/CR in TemplateElement.value.raw (@typescript-eslint/parser),
	// whereas JS normalizes them to LF (espree) — for both plain and interpolated templates.
	for _, tc := range []struct {
		file     string
		wantRaws []string
	}{
		{"/tpl_crlf.ts", []string{"a\r\nb\rc", "x\r\n", "y\rz"}},
		{"/tpl_crlf.js", []string{"a\nb\nc", "x\n", "y\nz"}},
	} {
		if err := builder.WithTree(source(tc.file, "const s = `a\r\nb\rc`; const t = `x\r\n${1}y\rz`;"), TreeOptions{SourceType: "module"}, func(tree *Tree) {
			var gotRaws []string
			var walk func(NodeID)
			walk = func(id NodeID) {
				if tree.Type(id) == typeTemplateElement {
					_, _, raw, ok := tree.TemplateValue(id)
					if !ok {
						t.Fatalf("%s: TemplateValue returned ok=false", tc.file)
					}
					gotRaws = append(gotRaws, raw)
				}
				tree.ForEachChild(id, func(child NodeID) bool {
					walk(child)
					return false
				})
			}
			walk(tree.Root())
			if !reflect.DeepEqual(gotRaws, tc.wantRaws) {
				t.Fatalf("%s: TemplateElement.value.raw = %q, want %q", tc.file, gotRaws, tc.wantRaws)
			}
		}); err != nil {
			t.Fatal(err)
		}
	}

	// 5. Bounds safety on invalid NodeIDs/FieldIDs and defensive clone on TriggerKinds.
	if err := builder.WithTree(source("/bounds.ts", `const a = [1, 2]; const b = [3];`), TreeOptions{SourceType: "module"}, func(tree *Tree) {
		if tree.Type(NoNode) != UnknownType || tree.Type(NodeID(99999)) != UnknownType {
			t.Fatal("invalid NodeID did not return UnknownType")
		}
		if tree.Parent(NoNode) != NoNode || tree.ParentField(NoNode) != NoField {
			t.Fatal("invalid NodeID Parent/ParentField not sentinel")
		}
		if _, _, ok := tree.SiblingList(NoNode); ok {
			t.Fatal("NoNode must not have a sibling list")
		}
		if got := tree.Field(tree.Root(), FieldID(250)); got.Kind != ValueUndefined {
			t.Fatalf("out-of-range FieldID = %+v, want ValueUndefined", got)
		}
	}); err != nil {
		t.Fatal(err)
	}
	kinds, ok := TriggerKinds(typeMemberExpression)
	if !ok || len(kinds) == 0 {
		t.Fatal("MemberExpression should be in triggerKindsTable")
	}
	orig := kinds[0]
	kinds[0] = ast.KindUnknown
	kinds2, _ := TriggerKinds(typeMemberExpression)
	if kinds2[0] != orig {
		t.Fatal("TriggerKinds returned mutable package-global slice")
	}

	// 6. Malformed error-recovery inputs must not underflow ranges or panic with runtime errors.
	for _, bad := range []struct {
		name string
		code string
	}{
		{"/unterminated_jsx.tsx", `<A x="`},
		{"/unterminated_tpl.ts", "`abc"},
		{"/unterminated_dir.ts", `"`},
		{"/empty_jsx_expr.tsx", `const x = <div>{}</div>;`},
	} {
		_ = builder.WithTree(source(bad.name, bad.code), TreeOptions{SourceType: "module"}, func(tree *Tree) {
			var check func(NodeID)
			check = func(id NodeID) {
				r := tree.Range(id)
				if r.Pos() < 0 || r.End() < r.Pos() || r.End() > len(bad.code)+1 {
					t.Fatalf("%s: invalid range [%d, %d] for len %d", bad.name, r.Pos(), r.End(), len(bad.code))
				}
				tree.ForEachChild(id, func(child NodeID) bool {
					check(child)
					return false
				})
			}
			check(tree.Root())
		})
	}
}

func TestTriggerKindsSoundness(t *testing.T) {
	corpus := map[string]string{
		"/js.js": `
			'use strict';
			with (obj) { debugger; ; }
			label: for (let i = 0; i < 1; i++) { if (i) continue label; else break label; }
			for (const k in obj) {}
			for await (const v of iter) {}
			while (a) {}
			do {} while (b);
			switch (x) { case 1: break; default: break; }
			try { throw new Error(); } catch (e) {} finally {}
			async function* gen(a = 1, ...rest) { yield await a; }
			const arrow = () => ({ a: 1, ...spread, m() {} });
			const arr = [1, , 2];
			const tpl = tag` + "`a${1}b`" + `;
			const lit = ` + "`plain`" + `;
			const u = typeof x + void y + delete z.a + ++i + i-- + (a ? b : c) + (a, b) + (a && b);
			a = 1;
			a?.b?.()!.c;
			import('mod');
			class C extends B { #p = 1; static { super.m(); } m() { return import.meta; } }
			import Def, * as NS from 'a' with { type: 'json' };
			import { x as y } from 'b';
			export { y as z } from 'b';
			export * as all from 'b';
			export default class {}
		`,
		"/ts.ts": `
			@dec abstract class AC<T extends string = string> extends Base<T> implements I<T> {
				abstract am(): void;
				abstract ap: number;
				abstract accessor aa: string;
				accessor acc = 1;
				readonly prop!: T;
				constructor(public p: number) {}
			}
			declare function df<T>(x: T): T;
			export interface I<T> extends Base<T> {
				readonly a?: T;
				m<U>(x: U): T;
				[k: string]: T;
				(x: number): void;
				new (x: number): I<T>;
			}
			export type Alias<T> = { -readonly [K in keyof T as K]-?: T[K] };
			type Kw = string | number | boolean | any | unknown | void | undefined | never | object | symbol | bigint | intrinsic | null;
			type Compound<T> = [label: T, T?, ...T[]] & (T extends infer U ? U[] : 'lit') & ` + "`a-${string}`" + ` & typeof ns.x & keyof T & this;
			type Fn = (x: unknown) => x is string;
			type Ctor = abstract new (x: number) => object;
			type Imp = import('pkg', { with: { 'resolution-mode': 'import' } }).X<string>;
			const a = v as unknown as string;
			const b = v!;
			const c = v satisfies unknown;
			const d = <string>v;
			const inst = fn<string>;
			const enum CE { A, B = 1 }
			namespace A.B { export const x = 1; }
			import Eq = require('eq');
			export = Eq;
			export as namespace GlobalNS;
		`,
		"/jsx.tsx": `
			const el = <A.B ns:attr="v" {...spread}>text<>{val}{/* c */}</><self /></A.B>;
		`,
	}

	var builder Builder
	for path, text := range corpus {
		file := source(path, text)
		visitedKinds := make(map[ast.Kind]bool)
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			visitedKinds[node.Kind] = true
			node.ForEachChild(visit)
			return false
		}
		// Match internal/linter/linter.go: starts at file.Node.ForEachChild.
		file.Node.ForEachChild(visit)

		if err := builder.WithTree(file, TreeOptions{SourceType: "module"}, func(tree *Tree) {
			var check func(NodeID)
			check = func(id NodeID) {
				tid := tree.Type(id)
				if kinds, ok := TriggerKinds(tid); ok {
					matched := false
					for _, k := range kinds {
						if visitedKinds[k] {
							matched = true
							break
						}
					}
					if !matched {
						t.Errorf("%s: projected %s but none of its TriggerKinds %v were visited by linter traversal", path, tid, kinds)
					}
				}
				tree.ForEachChild(id, func(child NodeID) bool {
					check(child)
					return false
				})
			}
			check(tree.Root())
		}); err != nil {
			t.Fatalf("%s: WithTree failed: %v", path, err)
		}
	}
}

func TestVisitorKeysMatchTypeScriptESLint(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not available")
	}
	script := `
const vkPath = require.resolve("@typescript-eslint/visitor-keys", {
  paths: [process.cwd(), process.cwd() + "/../../../../packages/rslint", process.cwd() + "/../../../../packages/rslint-test-tools"]
});
process.stdout.write(JSON.stringify(require(vkPath).visitorKeys));
`
	out, err := exec.Command("node", "-e", script).Output()
	if err != nil {
		t.Skipf("@typescript-eslint/visitor-keys not resolvable: %v", err)
	}
	var upstream map[string][]string
	if err := json.Unmarshal(out, &upstream); err != nil {
		t.Fatal(err)
	}
	for id := range schemas {
		want, ok := upstream[schemas[id].name]
		if !ok {
			t.Fatalf("missing upstream visitor keys for %s", schemas[id].name)
		}
		got := schemas[id].visitorKeys
		if len(got) == 0 && len(want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s visitorKeys = %v, want %v", schemas[id].name, got, want)
		}
	}
}

// TestEstreeImportBoundary ensures native Go rules stay on direct tsgo AST
// traversal: only internal/linter (JS plugin frames) and
// internal/rules/no_restricted_syntax (ESTree selectors) may import this
// package in production code.
func TestEstreeImportBoundary(t *testing.T) {
	const estreeImportPath = "github.com/web-infra-dev/rslint/internal/languages/js/estree"

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate boundary test")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", ".."))

	var violations []string
	for _, scanRoot := range []string{"cmd", "internal"} {
		absDir := filepath.Join(repoRoot, scanRoot)
		err := filepath.WalkDir(absDir, func(filePath string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				switch entry.Name() {
				case "testdata", "fixtures", "node_modules", ".git":
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}
			relPath, err := filepath.Rel(repoRoot, filePath)
			if err != nil {
				return err
			}
			relPath = filepath.ToSlash(relPath)
			pkgPath := filepath.ToSlash(filepath.Dir(relPath))

			fileSet := token.NewFileSet()
			file, err := parser.ParseFile(fileSet, filePath, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, imp := range file.Imports {
				importPath, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					return err
				}
				if importPath != estreeImportPath {
					continue
				}
				if pkgPath != "internal/linter" && pkgPath != "internal/rules/no_restricted_syntax" {
					violations = append(violations, relPath+": unauthorized production import of internal/languages/js/estree")
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s: %v", scanRoot, err)
		}
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("estree import boundary violations:\n%s", strings.Join(violations, "\n"))
	}
}

func BenchmarkWithTreeWindow(b *testing.B) {
	cases := map[string]string{
		"ASCII":       strings.Repeat("function f(a) { return { value: a + 2, text: 'abc' }; }\n", 1000),
		"Unicode":     strings.Repeat("const café = { 字段: '\\uD800', text: '😃' };\n", 1000),
		"LongString":  "const text = '" + strings.Repeat("a", 1<<20) + "';",
		"LongUnicode": "const text = '" + strings.Repeat("世界😀", (1<<20)/10) + "';",
	}
	for name, text := range cases {
		file := source("/input.ts", text)
		for _, window := range []int{1, 32} {
			b.Run(fmt.Sprintf("%s/%d", name, window), func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(text) * window))
				for b.Loop() {
					var builder Builder
					for range window {
						if err := builder.WithTree(file, TreeOptions{SourceType: "module"}, func(tree *Tree) {
							if tree.Root() == NoNode {
								b.Fatal("missing root")
							}
						}); err != nil {
							b.Fatal(err)
						}
					}
				}
			})
		}
	}
}
