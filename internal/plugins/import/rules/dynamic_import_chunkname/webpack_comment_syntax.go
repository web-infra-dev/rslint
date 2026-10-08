package dynamic_import_chunkname

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
		if isTypeScriptOnlySyntax(node) || isInvalidRegexLiteral(node) {
			invalid = true
			return true
		}
		return node.ForEachChild(visit)
	}
	root.ForEachChild(visit)
	return invalid
}

// isInvalidRegexLiteral reports a regular expression literal whose pattern or
// flags are an early error. The TypeScript parser only checks that the literal
// is terminated, so `/(/` or `/[z-a]/` parse without a diagnostic.
func isInvalidRegexLiteral(node *ast.Node) bool {
	return node.Kind == ast.KindRegularExpressionLiteral && !ecmascript.IsValidRegexLiteral(node.Text())
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
// not searched, and code that short-circuiting skips is not searched either.
func referencesUndefinedGlobal(root *ast.Node) bool {
	checker := &referenceChecker{evaluator: utils.NewStaticStringEvaluator(nil)}
	root.ForEachChild(checker.visit)
	return checker.found
}

type referenceChecker struct {
	evaluator      *utils.StaticStringEvaluator
	wrapperVisited bool
	found          bool
}

// visit reports true to stop the walk. An operand is searched only when it is
// known to be evaluated: when a condition or the left operand cannot be folded
// to a constant, the code it may skip is left alone rather than reported.
func (c *referenceChecker) visit(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionExpression:
		if c.wrapperVisited {
			return false
		}
		c.wrapperVisited = true
	case ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindClassExpression,
		ast.KindClassDeclaration, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return false
	case ast.KindIdentifier:
		if isEvaluatedReference(node) {
			if _, ok := vmContextGlobals[node.Text()]; !ok {
				c.found = true
				return true
			}
		}
	case ast.KindBinaryExpression:
		return c.visitBinary(node.AsBinaryExpression())
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		if c.visit(conditional.Condition) {
			return true
		}
		if truthy, known := c.evaluator.EvalControlFlowTruthiness(conditional.Condition); known {
			if truthy {
				return c.visit(conditional.WhenTrue)
			}
			return c.visit(conditional.WhenFalse)
		}
		return false
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression:
		if node.Flags&ast.NodeFlagsOptionalChain != 0 {
			return c.visitOptionalChain(node)
		}
	}
	return node.ForEachChild(c.visit)
}

func (c *referenceChecker) visitBinary(binary *ast.BinaryExpression) bool {
	switch binary.OperatorToken.Kind {
	case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
		if c.visit(binary.Left) {
			return true
		}
		// The right operand runs only for a falsy (`||`), truthy (`&&`) or
		// nullish (`??`) left operand.
		var evaluatesRight, known bool
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken:
			var truthy bool
			truthy, known = c.evaluator.EvalControlFlowTruthiness(binary.Left)
			evaluatesRight = !truthy
		case ast.KindAmpersandAmpersandToken:
			evaluatesRight, known = c.evaluator.EvalControlFlowTruthiness(binary.Left)
		default:
			evaluatesRight, known = c.evaluator.EvalControlFlowNullish(binary.Left)
		}
		return known && evaluatesRight && c.visit(binary.Right)
	case ast.KindEqualsToken:
		// Assigning to an undeclared name creates a global in sloppy mode, so
		// the target is not a read. Destructuring targets are left alone.
		switch ast.SkipParentheses(binary.Left).Kind {
		case ast.KindIdentifier, ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
			return c.visit(binary.Right)
		}
	}
	return binary.Node.ForEachChild(c.visit)
}

// visitOptionalChain searches the operand of a `?.` chain, and the rest of the
// chain only when no `?.` in it can short-circuit.
func (c *referenceChecker) visitOptionalChain(node *ast.Node) bool {
	var expression *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		expression = node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		expression = node.AsElementAccessExpression().Expression
	default:
		expression = node.AsCallExpression().Expression
	}
	if c.visit(expression) {
		return true
	}
	if !c.chainEvaluated(node) {
		return false
	}
	return node.ForEachChild(func(child *ast.Node) bool {
		if child == expression {
			return false
		}
		return c.visit(child)
	})
}

// chainEvaluated reports whether every `?.` in the chain below node is known
// to continue, that is, its operand is a constant that is not null or undefined.
func (c *referenceChecker) chainEvaluated(node *ast.Node) bool {
	for node != nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {
		var expression *ast.Node
		var questionDot *ast.Node
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			access := node.AsPropertyAccessExpression()
			expression, questionDot = access.Expression, access.QuestionDotToken
		case ast.KindElementAccessExpression:
			access := node.AsElementAccessExpression()
			expression, questionDot = access.Expression, access.QuestionDotToken
		case ast.KindCallExpression:
			call := node.AsCallExpression()
			expression, questionDot = call.Expression, call.QuestionDotToken
		default:
			return true
		}
		if questionDot != nil {
			if nullish, known := c.evaluator.EvalControlFlowNullish(expression); !known || nullish {
				return false
			}
		}
		node = expression
	}
	return true
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
