package dynamic_import_chunkname

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
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
}

// isValidWebpackCommentBody reports whether the text between the delimiters of
// a magic comment can be evaluated the way webpack does:
// `(function() {return {<body>}})()`. Upstream runs that source through
// `vm.runInNewContext`, so both syntax errors and runtime ReferenceErrors make
// the comment invalid.
//
// rslint cannot run JavaScript. It parses the same wrapper text and then looks
// for the failures that can occur in practice: syntax errors, TypeScript-only
// syntax anywhere in the text (including function and class bodies that are
// never called), and references to identifiers that a fresh `vm` context does
// not define. Other runtime errors, such as calling a non-function, are not
// detected.
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
	return !containsInvalidSyntax(root) && !referencesUndefinedGlobal(root)
}

// containsInvalidSyntax reports syntax that is an error in JavaScript but that
// the TypeScript parser accepts without a parse diagnostic. It visits every
// node: V8 rejects the whole source even when the offending function or class
// is never run.
func containsInvalidSyntax(root *ast.Node) bool {
	invalid := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if isTypeScriptOnlySyntax(node) {
			invalid = true
			return true
		}
		return node.ForEachChild(visit)
	}
	root.ForEachChild(visit)
	return invalid
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
		return parameter.QuestionToken != nil || parameter.Type != nil || parameter.Modifiers() != nil
	}
	return ast.IsTypeNode(node)
}

// referencesUndefinedGlobal reports whether evaluating the wrapper would read
// an identifier that does not exist in a fresh `vm` context. Only the wrapper
// function runs, so the bodies of functions and classes created inside it are
// not searched.
func referencesUndefinedGlobal(root *ast.Node) bool {
	undefinedReference := false
	wrapperVisited := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindFunctionExpression:
			if wrapperVisited {
				return false
			}
			wrapperVisited = true
		case ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindClassExpression,
			ast.KindClassDeclaration, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			return false
		case ast.KindIdentifier:
			if isEvaluatedReference(node) {
				if _, ok := vmContextGlobals[node.Text()]; !ok {
					undefinedReference = true
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	root.ForEachChild(visit)
	return undefinedReference
}

// isEvaluatedReference reports whether the identifier is read as a variable
// when the expression is evaluated.
func isEvaluatedReference(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Expression == node
	case ast.KindTypeOfExpression:
		// `typeof missing` does not throw.
		return false
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindPropertyDeclaration,
		ast.KindParameter, ast.KindVariableDeclaration, ast.KindBindingElement, ast.KindFunctionExpression,
		ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement, ast.KindMetaProperty,
		ast.KindQualifiedName:
		return false
	}
	return true
}
