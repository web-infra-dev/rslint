package dynamic_import_chunkname

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Names that exist in a fresh `vm` context, so referencing them in a magic
// comment value does not throw a ReferenceError. `arguments` is available
// because the comment body is evaluated inside a function.
var vmContextGlobals = map[string]struct{}{
	"undefined": {}, "NaN": {}, "Infinity": {}, "globalThis": {}, "arguments": {},
	"Object": {}, "Function": {}, "Array": {}, "Number": {}, "Boolean": {}, "String": {},
	"Symbol": {}, "Date": {}, "Promise": {}, "RegExp": {}, "BigInt": {},
	"Error": {}, "AggregateError": {}, "EvalError": {}, "RangeError": {}, "ReferenceError": {},
	"SyntaxError": {}, "TypeError": {}, "URIError": {},
	"JSON": {}, "Math": {}, "Intl": {}, "Reflect": {}, "Proxy": {}, "Atomics": {}, "WebAssembly": {},
	"ArrayBuffer": {}, "SharedArrayBuffer": {}, "DataView": {},
	"Int8Array": {}, "Uint8Array": {}, "Uint8ClampedArray": {}, "Int16Array": {}, "Uint16Array": {},
	"Int32Array": {}, "Uint32Array": {}, "Float32Array": {}, "Float64Array": {},
	"BigInt64Array": {}, "BigUint64Array": {},
	"Map": {}, "Set": {}, "WeakMap": {}, "WeakSet": {}, "WeakRef": {}, "FinalizationRegistry": {},
	"parseFloat": {}, "parseInt": {}, "isFinite": {}, "isNaN": {}, "eval": {},
	"decodeURI": {}, "decodeURIComponent": {}, "encodeURI": {}, "encodeURIComponent": {},
	"escape": {}, "unescape": {}, "console": {},
	"Iterator": {}, "Float16Array": {}, "DisposableStack": {}, "AsyncDisposableStack": {}, "SuppressedError": {}, "Temporal": {},
}

// isValidWebpackCommentBody reports whether the text between the delimiters of
// a magic comment can be evaluated the way webpack does:
// `(function() {return {<body>}})()`. Upstream runs that source through
// `vm.runInNewContext`, so both syntax errors and runtime ReferenceErrors make
// the comment invalid.
//
// rslint cannot run JavaScript. It parses the same wrapper text and then looks
// for the failures that can occur in practice: syntax errors, TypeScript-only
// syntax and invalid regular expression literals anywhere in the text
// (including function and class bodies that are never called), and references
// to identifiers that a fresh `vm` context does not define. Other runtime
// errors, such as calling a non-function, are not detected.
func isValidWebpackCommentBody(body string) bool {
	text := "(function() {return {" + body + "}})()"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/webpack-comment.js",
		Path:     "/webpack-comment.js",
	}, text, core.ScriptKindJS)
	if len(sourceFile.Diagnostics()) != 0 {
		return false
	}
	root := sourceFile.AsNode()
	return !containsInvalidSyntax(root) && !containsContextError(root) && !containsStrictModeError(sourceFile) &&
		!referencesUndefinedGlobal(root)
}

// containsInvalidSyntax reports syntax that is an error in JavaScript but that
// the TypeScript parser accepts without a parse diagnostic. It visits every
// node: V8 rejects the whole source even when the offending function or class
// is never run.
func containsInvalidSyntax(root *ast.Node) bool {
	invalid := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if isTypeScriptOnlySyntax(node) || isInvalidRegexLiteral(node) {
			invalid = true
			return true
		}
		return node.ForEachChild(visit)
	}
	root.ForEachChild(visit)
	return invalid
}

// isInvalidRegexLiteral reports a regular expression literal that `new RegExp`
// would reject, since a literal is an early error wherever it appears. The
// TypeScript parser only checks that the literal is terminated.
func isInvalidRegexLiteral(node *ast.Node) bool {
	if node.Kind != ast.KindRegularExpressionLiteral {
		return false
	}
	text := node.Text()
	end := strings.LastIndexByte(text, '/')
	if end <= 0 {
		return true
	}
	pattern, flags := text[1:end], text[end+1:]
	return !isValidRegexFlags(flags) || !utils.IsValidRegexPattern(pattern, utils.ParseRegexFlags(flags))
}

// isValidRegexFlags reports whether flags is a set of known flags without
// repeats, and without both u and v.
func isValidRegexFlags(flags string) bool {
	seen := map[rune]bool{}
	for _, flag := range flags {
		if seen[flag] || !strings.ContainsRune("dgimsuvy", flag) {
			return false
		}
		seen[flag] = true
	}
	return !seen['u'] || !seen['v']
}

func isTypeScriptOnlySyntax(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression,
		ast.KindTypeParameter, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration,
		ast.KindModuleDeclaration:
		return true
	case ast.KindExpressionWithTypeArguments:
		// `class A extends B {}` is JavaScript; type arguments and `implements` are not.
		return node.AsExpressionWithTypeArguments().TypeArguments != nil ||
			node.Parent != nil && node.Parent.Kind == ast.KindHeritageClause &&
				node.Parent.AsHeritageClause().Token == ast.KindImplementsKeyword
	case ast.KindParameter:
		parameter := node.AsParameterDeclaration()
		// `this` names a TypeScript this-parameter, which is not a binding.
		return parameter.QuestionToken != nil || parameter.Type != nil || parameter.Modifiers() != nil ||
			node.Name().Kind == ast.KindIdentifier && node.Name().Text() == "this"
	case ast.KindCallExpression:
		return node.AsCallExpression().TypeArguments != nil
	case ast.KindNewExpression:
		return node.AsNewExpression().TypeArguments != nil
	case ast.KindTaggedTemplateExpression:
		return node.AsTaggedTemplateExpression().TypeArguments != nil
	case ast.KindPropertyDeclaration, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindConstructor:
		return isInvalidClassMember(node)
	}
	return ast.IsTypeNode(node)
}

// isInvalidClassMember reports class member syntax that JavaScript rejects:
// modifiers other than static and async, optional or definite markers, return
// and property types, and members without a body.
func isInvalidClassMember(member *ast.Node) bool {
	switch member.Kind {
	case ast.KindPropertyDeclaration:
		return !hasOnlyModifiers(member, ast.KindStaticKeyword) || member.PostfixToken() != nil ||
			member.AsPropertyDeclaration().Type != nil
	case ast.KindMethodDeclaration:
		return !hasOnlyModifiers(member, ast.KindStaticKeyword, ast.KindAsyncKeyword) ||
			member.PostfixToken() != nil || member.Body() == nil || member.Type() != nil
	case ast.KindGetAccessor, ast.KindSetAccessor:
		return !hasOnlyModifiers(member, ast.KindStaticKeyword) || member.Body() == nil || member.Type() != nil
	case ast.KindConstructor:
		return member.Modifiers() != nil || member.Body() == nil
	}
	return false
}

// hasOnlyModifiers reports whether every modifier of node is one of allowed.
func hasOnlyModifiers(node *ast.Node, allowed ...ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return true
	}
	for _, modifier := range modifiers.Nodes {
		if !slices.Contains(allowed, modifier.Kind) {
			return false
		}
	}
	return true
}
