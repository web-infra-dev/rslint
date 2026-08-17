package no_param_reassign

import (
	_ "embed"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed no_param_reassign.schema.json
var schemaJSON []byte

type Options struct {
	Props                               bool
	IgnorePropertyModificationsFor      []string
	IgnorePropertyModificationsForRegex []*esregexp.RegExp
}

func parseOptions(options []any) Options {
	opts := Options{Props: false}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	if props, ok := optsMap["props"].(bool); ok {
		opts.Props = props
	}
	if arr, ok := optsMap["ignorePropertyModificationsFor"].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				opts.IgnorePropertyModificationsFor = append(opts.IgnorePropertyModificationsFor, s)
			}
		}
	}
	if arr, ok := optsMap["ignorePropertyModificationsForRegex"].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				// ECMAScript + Unicode flags mirror ESLint's `new RegExp(pattern, 'u')`,
				// so patterns using lookaround, backreferences, or `\p{...}` behave the
				// same as upstream (Go's RE2 supports none of those). An unparsable
				// pattern is rejected up front by the schema's `format: "regex"`, so
				// the compile-error branch here is only defensive.
				if re, err := esregexp.Compile(s, "u"); err == nil {
					opts.IgnorePropertyModificationsForRegex = append(opts.IgnorePropertyModificationsForRegex, re)
				}
			}
		}
	}
	return opts
}

func buildAssignmentMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "assignmentToFunctionParam",
		Description: "Assignment to function parameter '" + name + "'.",
	}
}

func buildPropAssignmentMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "assignmentToFunctionParamProp",
		Description: "Assignment to property of function parameter '" + name + "'.",
	}
}

// isPropWalkStopNode marks the edge of the expression context around a
// parameter reference. The walk in isModifyingProp continues upward through
// expressions and halts once it hits a statement, declaration, or function
// boundary — mirroring ESLint's stopNodePattern. ForIn/ForOf are intentionally
// excluded so the walk can still inspect their initializer position.
func isPropWalkStopNode(node *ast.Node) bool {
	if node == nil {
		return true
	}
	switch node.Kind {
	case ast.KindForInStatement, ast.KindForOfStatement:
		return false
	case ast.KindSourceFile, ast.KindModuleBlock:
		return true
	case ast.KindVariableDeclaration,
		ast.KindParameter,
		ast.KindPropertyDeclaration,
		ast.KindBindingElement,
		ast.KindClassStaticBlockDeclaration:
		return true
	}
	if ast.IsFunctionLikeDeclaration(node) {
		return true
	}
	return ast.IsStatement(node)
}

// isModifyingProp walks up from a parameter reference to decide whether the
// reference participates in a property modification (assignment, update,
// delete, or for-in/of target). Ported from ESLint's isModifyingProp.
func isModifyingProp(ident *ast.Node) bool {
	node := ident
	parent := node.Parent
	for parent != nil && !isPropWalkStopNode(parent) {
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary != nil && binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				return binary.Left == node
			}

		case ast.KindPrefixUnaryExpression:
			pf := parent.AsPrefixUnaryExpression()
			if pf != nil && (pf.Operator == ast.KindPlusPlusToken || pf.Operator == ast.KindMinusMinusToken) {
				return true
			}

		case ast.KindPostfixUnaryExpression:
			pf := parent.AsPostfixUnaryExpression()
			if pf != nil && (pf.Operator == ast.KindPlusPlusToken || pf.Operator == ast.KindMinusMinusToken) {
				return true
			}

		case ast.KindDeleteExpression:
			return true

		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := parent.AsForInOrOfStatement()
			if stmt != nil && stmt.Initializer == node {
				return true
			}
			// Initializer is somewhere else: this statement bounds the walk.
			return false

		case ast.KindCallExpression:
			// `foo(bar.x).y = 0` — bar is an argument, not the callee.
			call := parent.AsCallExpression()
			if call != nil && call.Expression != node {
				return false
			}

		case ast.KindPropertyAccessExpression:
			// `obj[bar].x = 0` — bar cannot appear as the Name side of a
			// property access; if it did (private identifier case), stop.
			pa := parent.AsPropertyAccessExpression()
			if pa != nil && pa.Name() == node {
				return false
			}

		case ast.KindElementAccessExpression:
			// `cache[bar] = 0` — bar is the computed key, not the target.
			ea := parent.AsElementAccessExpression()
			if ea != nil && ea.ArgumentExpression == node {
				return false
			}

		case ast.KindPropertyAssignment:
			// `({bar: ...})` — bar is the key position, not a write.
			pa := parent.AsPropertyAssignment()
			if pa != nil && pa.Name() == node {
				return false
			}

		case ast.KindShorthandPropertyAssignment:
			// `({bar} = ...)` is a direct reassignment, already handled by
			// IsWriteReference; never treat it as a property modification.
			return false

		case ast.KindConditionalExpression:
			// `(bar ? a : b).x = 0` — bar in the test position is only read.
			ce := parent.AsConditionalExpression()
			if ce != nil && ce.Condition == node {
				return false
			}
		}

		node = parent
		parent = node.Parent
	}
	return false
}

// isSameBinding returns true when the given identifier resolves to the same
// binding as one of the given parameter declarations.
//
// With TypeChecker: compares symbols by ValueDeclaration rather than by
// pointer — TypeScript creates two distinct symbols for `public`/`private`
// parameter properties (one for the local parameter, one for the class
// field), and both symbols share the same Parameter declaration. Pointer
// comparison would miss that.
//
// Without TypeChecker: falls back to a scope walk — the reference refers to
// the parameter unless an intermediate scope introduces its own binding with
// the same name.
func isSameBinding(ident *ast.Node, binding *paramBinding, fn *ast.Node, ctx rule.RuleContext) bool {
	if ctx.TypeChecker != nil {
		if !binding.symbolResolved {
			binding.symbol = ctx.TypeChecker.GetSymbolAtLocation(binding.ident)
			binding.symbolResolved = true
		}
		paramSymbol := binding.symbol
		if paramSymbol == nil {
			return !utils.IsNameShadowedBetween(ident, fn, binding.name)
		}
		refSym := utils.GetReferenceSymbol(ident, ctx.TypeChecker)
		if refSym == nil {
			return false
		}
		if refSym == paramSymbol {
			return true
		}
		// Parameter properties: two symbols share a Parameter decl.
		return binding.decl != nil && refSym.ValueDeclaration == binding.decl
	}
	return !utils.IsNameShadowedBetween(ident, fn, binding.name)
}

func isIgnoredPropertyAssignment(opts Options, name string) bool {
	if slices.Contains(opts.IgnorePropertyModificationsFor, name) {
		return true
	}
	for _, re := range opts.IgnorePropertyModificationsForRegex {
		if re.TestOrTimeout(name) {
			return true
		}
	}
	return false
}

// paramBinding holds the bits needed to decide whether an identifier elsewhere
// in the function refers to this parameter. The symbol is resolved lazily on
// the first possible write; most functions never need it. The declaration node
// (Parameter or BindingElement) remains the fallback discriminator.
type paramBinding struct {
	ident                   *ast.Node
	name                    string
	decl                    *ast.Node // `ident.Parent` — Parameter or BindingElement
	symbol                  *ast.Symbol
	symbolResolved          bool
	ignorePropertyMutations bool
}

// checkFunction walks every identifier inside `fn` and reports reassignments
// or property modifications that target one of its parameters.
func checkFunction(fn *ast.Node, opts Options, ctx rule.RuleContext) {
	// Collect all parameter bindings, flattening nested destructuring so that
	// `function foo({a, b: [c]})` yields three entries.
	var bindings []paramBinding
	for _, param := range fn.Parameters() {
		if param == nil || param.Name() == nil {
			continue
		}
		utils.CollectBindingNames(param.Name(), func(ident *ast.Node, n string) {
			bindings = append(bindings, paramBinding{
				ident:                   ident,
				name:                    n,
				decl:                    ident.Parent,
				ignorePropertyMutations: opts.Props && isIgnoredPropertyAssignment(opts, n),
			})
		})
	}
	if len(bindings) == 0 {
		return
	}

	// Skip the declaration identifiers themselves when walking.
	skipSet := make(map[*ast.Node]struct{}, len(bindings))
	// Avoid binding work for identifiers that cannot name a parameter.
	byName := make(map[string]*paramBinding, len(bindings))
	for i := range bindings {
		skipSet[bindings[i].ident] = struct{}{}
		byName[bindings[i].name] = &bindings[i]
	}

	// Guard against reporting the same identifier twice.
	reported := make(map[*ast.Node]struct{})

	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier {
			name := node.AsIdentifier().Text
			if b := byName[name]; b != nil {
				_, skip := skipSet[node]
				_, already := reported[node]
				if !skip && !already {
					isDirectWrite := utils.IsWriteReference(node)
					isPropertyWrite := !isDirectWrite && opts.Props &&
						!b.ignorePropertyMutations && isModifyingProp(node)
					if (isDirectWrite || isPropertyWrite) && isSameBinding(node, b, fn, ctx) {
						reported[node] = struct{}{}
						if isDirectWrite {
							ctx.ReportNode(node, buildAssignmentMessage(name))
						} else {
							ctx.ReportNode(node, buildPropAssignmentMessage(name))
						}
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}

	walk(fn)
}

var NoParamReassignRule = rule.Rule{
	Name:   "no-param-reassign",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		handler := func(node *ast.Node) {
			checkFunction(node, opts, ctx)
		}
		// Cover every function-like declaration kind — plain/arrow/async/
		// generator functions, class methods, constructors, and accessors.
		// Signature-only kinds (MethodSignature, CallSignature, etc.) have
		// no body and are intentionally omitted.
		return rule.RuleListeners{
			ast.KindFunctionDeclaration: handler,
			ast.KindFunctionExpression:  handler,
			ast.KindArrowFunction:       handler,
			ast.KindMethodDeclaration:   handler,
			ast.KindConstructor:         handler,
			ast.KindGetAccessor:         handler,
			ast.KindSetAccessor:         handler,
		}
	},
}
