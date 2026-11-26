package test_framework

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func parseFirstCall(t *testing.T, code string) (*ast.SourceFile, *ast.Node) {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)

	call := findFirstNodeOfKind(sourceFile, ast.KindCallExpression)
	if call == nil {
		t.Fatalf("no call expression in %q", code)
	}
	return sourceFile, call
}

// parseFirstNodeOfKind returns the first node of kind in traversal order, so
// nested constructs resolve to the outermost one.
func parseFirstNodeOfKind(t *testing.T, code string, kind ast.Kind) *ast.Node {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)

	node := findFirstNodeOfKind(sourceFile, kind)
	if node == nil {
		t.Fatalf("no %v node in %q", kind, code)
	}
	return node
}

func findFirstNodeOfKind(sourceFile *ast.SourceFile, kind ast.Kind) *ast.Node {
	var found *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == kind {
			found = node
			return true
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	return found
}

func TestGetMemberEntries(t *testing.T) {
	tests := []struct {
		code      string
		wantNames []string
	}{
		{
			code:      "describe.only()",
			wantNames: []string{"describe", "only"},
		},
		{
			code:      `describe[ "only" ]("math", () => {})`,
			wantNames: []string{"describe", "only"},
		},
		{
			code:      "describe[\n  \"only\"\n](\"math\", () => {})",
			wantNames: []string{"describe", "only"},
		},
		{
			code:      "test.concurrent.only()",
			wantNames: []string{"test", "concurrent", "only"},
		},
	}

	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			_, call := parseFirstCall(t, test.code)
			entries := GetMemberEntries(call)
			names := make([]string, len(entries))
			for i, entry := range entries {
				names[i] = entry.Name
			}
			if !slices.Equal(names, test.wantNames) {
				t.Fatalf("member names = %v, want %v", names, test.wantNames)
			}
		})
	}
}

func TestGetMemberEntriesRejectsDynamicElementAccess(t *testing.T) {
	_, call := parseFirstCall(t, "describe[getMode()]()")
	if entries := GetMemberEntries(call); entries != nil {
		t.Fatalf("dynamic element access entries = %#v, want nil", entries)
	}
}

func TestGetMemberEntriesRejectsNonIdentifierRoots(t *testing.T) {
	cases := []string{
		`/a/.test(text)`,
		`(/a/).test(text)`,
		`/a/["test"](text)`,
		`new RegExp("a").test(text)`,
		`this.test()`,
		`class C extends Base { m() { super.test(); } }`,
		`({ test() {} }).test()`,
		`[].describe()`,
		`"something".describe()`,
		"``.test()",
		`(a || b).test()`,
		`(cond ? a : b).test()`,
		`patterns[0].test(text)`,
		`map[getKey()].test(text)`,
		`a[""].test(text)`,
		`re!.test(text)`,
		`(re as RegExp).test(text)`,
		`import.meta.rstest.expect(1).toBe(1)`,
		`(/a/)().test(text)`,
		`this.expect(1).toBe(2)`,
	}
	for _, code := range cases {
		t.Run(code, func(t *testing.T) {
			_, call := parseFirstCall(t, code)
			if entries := GetMemberEntries(call); len(entries) != 0 {
				t.Fatalf("GetMemberEntries(%s) = %#v, want nil", code, entries)
			}
		})
	}
}

func TestGetMemberEntriesFromRoot(t *testing.T) {
	matchImportMetaRstest := func(node *ast.Node) (MemberEntry, bool) {
		node = ast.SkipParentheses(node)
		if node == nil || node.Kind != ast.KindPropertyAccessExpression {
			return MemberEntry{}, false
		}
		prop := node.AsPropertyAccessExpression()
		if prop.Name() == nil || prop.Name().Text() != "rstest" {
			return MemberEntry{}, false
		}
		expr := ast.SkipParentheses(prop.Expression)
		if expr == nil || expr.Kind != ast.KindMetaProperty {
			return MemberEntry{}, false
		}
		return MemberEntry{Name: "rstest", Node: prop.Name()}, true
	}

	_, call := parseFirstCall(t, `import.meta.rstest.expect(1).toBe(1)`)
	entries := GetMemberEntriesFromRoot(call, false, matchImportMetaRstest)
	if got := JoinMemberEntries(entries); got != "rstest.expect.toBe" {
		t.Fatalf("GetMemberEntriesFromRoot = %q, want %q", got, "rstest.expect.toBe")
	}

	for _, code := range []string{
		`import.meta.other.rstest.expect(1).toBe(1)`,
		`expect(1).toBe(1)`,
		`(import.meta.rstest.expect(1) as any).resolves.toBe(1)`,
	} {
		_, c := parseFirstCall(t, code)
		if got := GetMemberEntriesFromRoot(c, false, matchImportMetaRstest); len(got) != 0 {
			t.Fatalf("GetMemberEntriesFromRoot(%s, false) = %#v, want nil", code, got)
		}
	}

	_, wrapped := parseFirstCall(t, `(import.meta.rstest.expect(1) as any).resolves.toBe(1)`)
	entries = GetMemberEntriesFromRoot(wrapped, true, matchImportMetaRstest)
	if got := JoinMemberEntries(entries); got != "rstest.expect.resolves.toBe" {
		t.Fatalf("GetMemberEntriesFromRoot(wrapped, true) = %q, want %q", got, "rstest.expect.resolves.toBe")
	}

	if got := GetMemberEntriesFromRoot(call, false, func(n *ast.Node) (MemberEntry, bool) {
		return MemberEntry{Name: "", Node: n}, true
	}); len(got) != 0 {
		t.Fatalf("empty root name entries = %#v, want nil", got)
	}
	if got := GetMemberEntriesFromRoot(call, false, func(*ast.Node) (MemberEntry, bool) {
		return MemberEntry{Name: "rstest", Node: nil}, true
	}); len(got) != 0 {
		t.Fatalf("nil root node entries = %#v, want nil", got)
	}

	_, identCall := parseFirstCall(t, `expect(1).toBe(1)`)
	if got := JoinMemberEntries(GetMemberEntriesFromRoot(identCall, false, nil)); got != "expect.toBe" {
		t.Fatalf("GetMemberEntriesFromRoot(nil matcher) = %q, want %q", got, "expect.toBe")
	}
}

func TestAccessorPropertyAndName(t *testing.T) {
	tests := []struct {
		code     string
		wantName string
		wantKind ast.Kind
	}{
		{`a.then()`, "then", ast.KindIdentifier},
		{`(a.then)()`, "then", ast.KindIdentifier},
		{`a["then"]()`, "then", ast.KindStringLiteral},
		{"a[`catch`]()", "catch", ast.KindNoSubstitutionTemplateLiteral},
		{`a[("then")]()`, "then", ast.KindStringLiteral},
		{`a[key]()`, "key", ast.KindIdentifier},
		{`a[0]()`, "", ast.KindNumericLiteral},
		{`a[getKey()]()`, "", ast.KindCallExpression},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			_, call := parseFirstCall(t, test.code)
			propNode, name := AccessorPropertyAndName(call.AsCallExpression().Expression)
			if propNode == nil || propNode.Kind != test.wantKind || name != test.wantName {
				t.Fatalf("AccessorPropertyAndName(%s) = (%v, %q), want (%v, %q)", test.code, propNode, name, test.wantKind, test.wantName)
			}
		})
	}
	_, bareCall := parseFirstCall(t, `a()`)
	if propNode, name := AccessorPropertyAndName(bareCall.AsCallExpression().Expression); propNode != nil || name != "" {
		t.Fatalf("AccessorPropertyAndName(identifier) = (%v, %q), want (nil, \"\")", propNode, name)
	}
	if propNode, name := AccessorPropertyAndName(nil); propNode != nil || name != "" {
		t.Fatalf("AccessorPropertyAndName(nil) = (%v, %q), want (nil, \"\")", propNode, name)
	}
}

func TestMemberChainTransparentExpressionPolicy(t *testing.T) {
	tests := []string{
		`(expect(value) as any).resolves.toBe(1)`,
		`(<any>expect(value)).resolves.toBe(1)`,
		`(expect(value) satisfies unknown).resolves.toBe(1)`,
		`expect(value)!.resolves.toBe(1)`,
		`(((expect(value) as any)!) satisfies unknown).resolves.toBe(1)`,
	}
	for _, code := range tests {
		t.Run(code, func(t *testing.T) {
			_, call := parseFirstCall(t, code)
			if entries := GetMemberEntries(call); len(entries) != 0 {
				t.Fatalf("default member entries = %#v, want nil past transparent wrapper", entries)
			}
			entries := GetMemberEntriesThroughTransparentExpressions(call)
			names := make([]string, len(entries))
			for i, entry := range entries {
				names[i] = entry.Name
			}
			want := []string{"expect", "resolves", "toBe"}
			if !slices.Equal(names, want) {
				t.Fatalf("transparent member names = %v, want %v", names, want)
			}

			root := ResolveFirstIdentifierThroughTransparentExpressions(call.AsCallExpression().Expression)
			if root == nil || root.Kind != ast.KindIdentifier || root.AsIdentifier().Text != "expect" {
				t.Fatalf("transparent first identifier = %#v, want expect", root)
			}
		})
	}
}

func TestTransparentTraversalDoesNotCrossNonReceiverPositions(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{`consume(expect(value) as any)`, "consume"},
		{`receiver[expect(value) as any]()`, "receiver"},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			_, call := parseFirstCall(t, test.code)
			root := ResolveFirstIdentifierThroughTransparentExpressions(call.AsCallExpression().Expression)
			if root == nil || root.Kind != ast.KindIdentifier || root.AsIdentifier().Text != test.want {
				t.Fatalf("first identifier = %#v, want %s", root, test.want)
			}
		})
	}
}

func TestResolveFirstIdentifier(t *testing.T) {
	_, call := parseFirstCall(t, "((test).each)`table`('name', () => {})")
	identifier := ResolveFirstIdentifier(call.AsCallExpression().Expression)
	if identifier == nil || identifier.Kind != ast.KindIdentifier || identifier.AsIdentifier().Text != "test" {
		t.Fatalf("first identifier = %#v, want test", identifier)
	}

	_, commaCall := parseFirstCall(t, "(ignored, test)()")
	identifier = ResolveFirstIdentifier(commaCall.AsCallExpression().Expression)
	if identifier == nil || identifier.Kind != ast.KindIdentifier || identifier.AsIdentifier().Text != "test" {
		t.Fatalf("comma expression identifier = %#v, want test", identifier)
	}

	_, logicalCall := parseFirstCall(t, "(ignored || test)()")
	if identifier := ResolveFirstIdentifier(logicalCall.AsCallExpression().Expression); identifier != nil {
		t.Fatalf("logical expression identifier = %#v, want nil", identifier)
	}

	_, missingReceiverCall := parseFirstCall(t, "().test()")
	if identifier := ResolveFirstIdentifier(missingReceiverCall.AsCallExpression().Expression); identifier != nil {
		t.Fatalf("missing receiver identifier = %#v, want nil", identifier)
	}
}

func TestCalleeChainName(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{`foo();`, "foo"},
		{`foo.bar();`, "foo.bar"},
		{`foo.bar.baz();`, "foo.bar.baz"},
		{`foo["bar"]();`, "foo.bar"},
		{"foo[`bar`]();", "foo.bar"},
		// Identifier keys are supported accessor names upstream, matching
		// getNodeName's treatment of computed identifier properties.
		{`foo[bar]();`, "foo.bar"},
		// Unsupported computed keys break the chain entirely.
		{`foo[bar.baz]();`, ""},
		{`foo[1]();`, ""},
		{"foo[`bar${x}`]();", ""},
		// Calls and news are peeled.
		{`foo().bar();`, "foo.bar"},
		{`new Foo().bar();`, "Foo.bar"},
		// new (require("x")).y() parses entirely as a NewExpression, so the
		// outer call has to be forced with parentheses to exercise the peel.
		{`(new (require("x")).y)();`, "require.y"},
		// Parentheses and optional chaining.
		{`(foo.bar)();`, "foo.bar"},
		{`foo?.bar();`, "foo.bar"},
		// Tagged templates resolve through their tag.
		{"foo.bar`x`();", "foo.bar"},
		// A non-identifier root or incomplete property access yields no name.
		{`"str".includes();`, ""},
		{`foo.();`, ""},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			_, call := parseFirstCall(t, test.code)
			if got := CalleeChainName(call.AsCallExpression().Expression); got != test.want {
				t.Errorf("CalleeChainName = %q, want %q", got, test.want)
			}
		})
	}

	if got := CalleeChainName(nil); got != "" {
		t.Errorf("CalleeChainName(nil) = %q, want empty", got)
	}
}

func TestIsFunction(t *testing.T) {
	functionForms := []struct {
		name string
		code string
		kind ast.Kind
	}{
		{"function declaration", `function f() {}`, ast.KindFunctionDeclaration},
		{"function expression", `const f = function () {};`, ast.KindFunctionExpression},
		{"arrow function", `const f = () => {};`, ast.KindArrowFunction},
		{"method declaration", `class C { m() {} }`, ast.KindMethodDeclaration},
		{"constructor", `class C { constructor() {} }`, ast.KindConstructor},
		{"get accessor", `class C { get p() { return 1; } }`, ast.KindGetAccessor},
		{"set accessor", `class C { set p(v) {} }`, ast.KindSetAccessor},
	}
	for _, form := range functionForms {
		t.Run(form.name, func(t *testing.T) {
			if !IsFunction(parseFirstNodeOfKind(t, form.code, form.kind)) {
				t.Errorf("IsFunction should be true for %s", form.name)
			}
		})
	}

	nonFunctionForms := []struct {
		name string
		code string
		kind ast.Kind
	}{
		{"call expression", `f();`, ast.KindCallExpression},
		{"class declaration", `class C {}`, ast.KindClassDeclaration},
		{"identifier", `x;`, ast.KindIdentifier},
		{"property access", `a.b;`, ast.KindPropertyAccessExpression},
	}
	for _, form := range nonFunctionForms {
		t.Run(form.name, func(t *testing.T) {
			if IsFunction(parseFirstNodeOfKind(t, form.code, form.kind)) {
				t.Errorf("IsFunction should be false for %s", form.name)
			}
		})
	}

	if IsFunction(nil) {
		t.Error("IsFunction(nil) should be false")
	}
}
