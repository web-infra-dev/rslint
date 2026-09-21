// cspell:ignore arraybuffers
package es_syntax

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/referencetracker"
)

func (c *syntaxChecker) checkBuiltins() {
	tracker := referencetracker.New(c.ctx)
	for _, f := range features {
		if _, ok := c.enabled[f.name]; !ok || len(f.reads) == 0 {
			continue
		}
		traces := map[string]*referencetracker.Trace{}
		for _, path := range f.reads {
			properties := traces
			var trace *referencetracker.Trace
			for _, part := range strings.Split(path, ".") {
				trace = properties[part]
				if trace == nil {
					trace = &referencetracker.Trace{Properties: map[string]*referencetracker.Trace{}}
					properties[part] = trace
				}
				properties = trace.Properties
			}
			trace.Read = func(node *ast.Node) {
				if f.name != "subclassing-builtins" || extendingClass(node) != nil {
					c.report(f.name, node)
				}
			}
		}
		tracker.TrackGlobals(traces)
	}
	// RegExp constructors are tracked through aliases and global-object access,
	// just like es-x. A shadowed constructor is not a builtin.
	tracker.TrackGlobals(map[string]*referencetracker.Trace{"RegExp": {Call: c.regexpCall, Construct: c.regexpCall}})
	if _, ok := c.enabled["resizable-and-growable-arraybuffers"]; ok {
		check := func(node *ast.Node) {
			args := node.Arguments()
			if len(args) > 1 && args[0].Kind != ast.KindSpreadElement && args[1].Kind != ast.KindSpreadElement {
				c.report("resizable-and-growable-arraybuffers", args[1])
			}
		}
		tracker.TrackGlobals(map[string]*referencetracker.Trace{"ArrayBuffer": {Construct: check}, "SharedArrayBuffer": {Construct: check}})
	}
	if _, ok := c.enabled["error-cause"]; ok {
		traces := map[string]*referencetracker.Trace{}
		for _, name := range []string{"Error", "EvalError", "RangeError", "ReferenceError", "SyntaxError", "TypeError", "URIError", "AggregateError"} {
			index := 1
			if name == "AggregateError" {
				index = 2
			}
			traces[name] = &referencetracker.Trace{
				Construct: func(node *ast.Node) {
					if c.hasCause(node, index) {
						c.report("error-cause", node)
					}
				},
				Read: func(node *ast.Node) {
					class := extendingClass(node)
					if class == nil {
						return
					}
					// Only the first matching super call is reported for each superclass
					// reference by the pinned upstream rule.
					var visit func(*ast.Node) bool
					visit = func(child *ast.Node) bool {
						if child != class && ast.IsClassLike(child) {
							return false
						}
						if child.Kind == ast.KindCallExpression && child.Expression().Kind == ast.KindSuperKeyword && c.hasCause(child, index) {
							c.report("error-cause", child)
							return true
						}
						return child.ForEachChild(visit)
					}
					class.ForEachChild(visit)
				},
			}
		}
		tracker.TrackGlobals(traces)
	}
}

func extendingClass(node *ast.Node) *ast.Node {
	node = utils.OutermostParenthesizedExpression(node)
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindExpressionWithTypeArguments && parent.Expression() == node {
		parent = parent.Parent
	}
	if parent != nil && parent.Kind == ast.KindHeritageClause && parent.AsHeritageClause().Token == ast.KindExtendsKeyword && ast.IsClassLike(parent.Parent) {
		return parent.Parent
	}
	return nil
}
func (c *syntaxChecker) hasCause(node *ast.Node, index int) bool {
	args := node.Arguments()
	if len(args) <= index {
		return false
	}
	for _, arg := range args[:index] {
		if arg.Kind == ast.KindSpreadElement {
			return false
		}
	}
	options := utils.ESTreeRuntimeExpression(args[index])
	if options.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, property := range options.Properties() {
		if property.Kind != ast.KindSpreadAssignment {
			if name, ok := c.static().EvalPropertyName(property.Name()); ok && name == "cause" {
				return true
			}
		}
	}
	return false
}

func (c *syntaxChecker) checkPrototype(node *ast.Node) {
	if len(c.prototypes) == 0 {
		return
	}
	name, ok := c.static().EvalAccessExpressionName(node)
	if !ok {
		return
	}
	indices := c.prototypes[name]
	if len(indices) == 0 {
		return
	}
	object, _ := utils.MemberExpressionParts(node)
	settings, _ := c.ctx.Settings["es-x"].(map[string]any)
	aggressive, _ := settings["aggressive"].(bool)
	objectType := ""
	resolved := false
	for _, index := range indices {
		f := features[index]
		for _, prototype := range f.prototypes {
			if !slices.Contains(prototype.properties, name) {
				continue
			}
			if !resolved {
				objectType = c.expressionType(object)
				resolved = true
			}
			if c.matchesPrototype(node, object, objectType, prototype.class, aggressive) {
				c.report(f.name, node)
				break
			}
		}
	}
}

// es-x's default, syntax-only receiver inference. Binding and mutation facts
// come from the existing evaluator and reference store, not a second scope walk.
func (c *syntaxChecker) expressionType(node *ast.Node) string {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return ""
	}
	if c.expressionTypes == nil {
		c.expressionTypes = map[*ast.Node]string{}
	}
	if value, ok := c.expressionTypes[node]; ok {
		return value
	}
	c.expressionTypes[node] = ""
	value := c.inferExpressionType(node)
	c.expressionTypes[node] = value
	return value
}
func (c *syntaxChecker) inferExpressionType(node *ast.Node) string {
	switch node.Kind {
	case ast.KindArrayLiteralExpression:
		return "Array"
	case ast.KindObjectLiteralExpression:
		return "Object"
	case ast.KindRegularExpressionLiteral:
		return "RegExp"
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression, ast.KindTypeOfExpression:
		return "String"
	case ast.KindNumericLiteral, ast.KindPostfixUnaryExpression:
		return "Number"
	case ast.KindBigIntLiteral:
		return "BigInt"
	case ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindDeleteExpression:
		return "Boolean"
	case ast.KindNullKeyword:
		return "null"
	case ast.KindVoidExpression:
		return "undefined"
	case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression:
		return "Function"
	case ast.KindIdentifier:
		if c.globalIdentifier(node) {
			if builtinConstructor(node.Text()) {
				return "Function"
			}
			switch node.Text() {
			case "undefined":
				return "undefined"
			case "NaN", "Infinity":
				return "Number"
			case "Intl":
				return "Object"
			}
		}
		if init, ok := c.static().ResolveIdentifierInitializer(node); ok {
			return c.expressionType(init)
		}
		if symbol := c.ctx.Refs.ResolveInFile(node); symbol != nil && len(symbol.Declarations) == 1 {
			kind := symbol.Declarations[0].Kind
			if kind == ast.KindFunctionDeclaration || kind == ast.KindFunctionExpression {
				return "Function"
			}
		}
	case ast.KindConditionalExpression:
		cond := node.AsConditionalExpression()
		a, b := c.expressionType(cond.WhenTrue), c.expressionType(cond.WhenFalse)
		if a == b {
			return a
		}
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		callee := utils.ESTreeCallCallee(node.Expression())
		if node.Kind == ast.KindTaggedTemplateExpression {
			callee = utils.ESTreeCallCallee(node.AsTaggedTemplateExpression().Tag)
		}
		if callee == nil {
			return ""
		}
		if callee.Kind == ast.KindIdentifier && c.globalIdentifier(callee) && builtinConstructor(callee.Text()) {
			return callee.Text()
		}
		if callee.Kind == ast.KindPropertyAccessExpression {
			receiver, name := utils.MemberExpressionParts(callee)
			if c.globalIdentifier(receiver) && receiver.Text() == "Intl" && slices.Contains([]string{"Collator", "DateTimeFormat", "ListFormat", "NumberFormat", "PluralRules", "RelativeTimeFormat", "Segmenter"}, name.Text()) {
				return "Intl." + name.Text()
			}
		}
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		switch unary.Operator {
		case ast.KindExclamationToken:
			return "Boolean"
		case ast.KindPlusToken, ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return "Number"
		case ast.KindMinusToken, ast.KindTildeToken:
			t := c.expressionType(unary.Operand)
			if t == "BigInt" {
				return t
			}
			if t != "" {
				return "Number"
			}
		}
	case ast.KindBinaryExpression:
		b := node.AsBinaryExpression()
		op := b.OperatorToken.Kind
		if op == ast.KindEqualsToken || op == ast.KindCommaToken {
			return c.expressionType(b.Right)
		}
		left, right := c.expressionType(b.Left), c.expressionType(b.Right)
		switch op {
		case ast.KindPlusToken, ast.KindPlusEqualsToken:
			if left == "String" || right == "String" {
				return "String"
			}
			if left == "BigInt" || right == "BigInt" {
				return "BigInt"
			}
			if right == "Number" || left == "Number" && (right == "null" || right == "undefined") {
				return "Number"
			}
			if right != "" {
				return "String"
			}
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			if left == right {
				return left
			}
		case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken, ast.KindLessThanToken, ast.KindLessThanEqualsToken, ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken, ast.KindInKeyword, ast.KindInstanceOfKeyword:
			return "Boolean"
		case ast.KindLessThanLessThanToken, ast.KindLessThanLessThanEqualsToken, ast.KindGreaterThanGreaterThanToken, ast.KindGreaterThanGreaterThanEqualsToken, ast.KindGreaterThanGreaterThanGreaterThanToken, ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
			return "Number"
		default:
			if left == "BigInt" || right == "BigInt" {
				return "BigInt"
			}
			if left != "" || right != "" {
				return "Number"
			}
		}
	}
	return ""
}
func builtinConstructor(name string) bool {
	return strings.Contains(" String Number Boolean Symbol BigInt Object Function Array RegExp Date Promise Int8Array Uint8Array Uint8ClampedArray Int16Array Uint16Array Int32Array Uint32Array Float32Array Float64Array BigInt64Array BigUint64Array ArrayBuffer SharedArrayBuffer ", " "+name+" ")
}
func (c *syntaxChecker) globalIdentifier(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindIdentifier && c.ctx.Globals.Access(node.Text()).IsDeclared() && c.ctx.Refs.IsGlobalNameReference(node, node.Text(), ast.SymbolFlagsAll)
}
func isLegacyAccessor(name string) bool {
	return name == "__defineGetter__" || name == "__defineSetter__" || name == "__lookupGetter__" || name == "__lookupSetter__"
}
func (c *syntaxChecker) legacyAccessorIdentifier(node *ast.Node) {
	if !isLegacyAccessor(node.Text()) || !c.ctx.Refs.IsGlobalReference(node) {
		return
	}
	outer := utils.OutermostParenthesizedExpression(node)
	if _, key := utils.MemberExpressionParts(outer.Parent); key == outer {
		return
	}
	if outer.Parent.Kind == ast.KindComputedPropertyName {
		owner := outer.Parent.Parent.Parent
		if owner.Kind == ast.KindObjectLiteralExpression || owner.Kind == ast.KindObjectBindingPattern {
			return
		}
	}
	c.report("legacy-object-prototype-accessor-methods", node)
}
func (c *syntaxChecker) shadowCatchParameter(node *ast.Node) {
	if _, ok := c.enabled["shadow-catch-param"]; !ok {
		return
	}
	clause := node.AsCatchClause()
	name := clause.VariableDeclaration.Name()
	if name.Kind != ast.KindIdentifier {
		return
	}
	var visit func(*ast.Node) bool
	visit = func(child *ast.Node) bool {
		if ast.IsFunctionLikeDeclaration(child) || ast.IsClassLike(child) || child.Kind == ast.KindModuleDeclaration {
			return false
		}
		if child.Kind == ast.KindVariableDeclaration && child.Parent.Kind == ast.KindVariableDeclarationList && child.Parent.Flags&ast.NodeFlagsBlockScoped == 0 {
			matches := false
			utils.CollectBindingNames(child.Name(), func(_ *ast.Node, binding string) { matches = matches || binding == name.Text() })
			if matches {
				c.report("shadow-catch-param", child)
			}
		}
		return child.ForEachChild(visit)
	}
	clause.Block.ForEachChild(visit)
}
