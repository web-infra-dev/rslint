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
// syntax, and references to identifiers that a fresh `vm` context does not
// define. Other runtime errors, such as calling a non-function, are not
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

	valid := true
	wrapperVisited := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
			valid = false
			return true
		case ast.KindFunctionExpression:
			// Only the wrapper function runs; later functions are created but
			// never called.
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
					valid = false
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	return valid
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
