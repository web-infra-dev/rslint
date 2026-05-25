package complexity

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const defaultThreshold = 20

// ComplexityRule enforces a maximum cyclomatic complexity allowed in a program.
// https://eslint.org/docs/latest/rules/complexity
var ComplexityRule = rule.Rule{
	Name: "complexity",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		threshold, isModified := parseOptions(options)

		// counters maps a "code path owner" node to its complexity counter.
		// Owners are: function-likes, ClassStaticBlockDeclarations, and
		// PropertyDeclarations whose Initializer is a non-function expression
		// (the field-initializer code path).
		counters := map[*ast.Node]int{}

		// findOwner walks up from a branching node to the nearest enclosing
		// code path owner, mirroring ESLint's CodePathAnalyzer origin. Returns
		// nil for top-level (program) — the complexity rule does not report
		// at the program level.
		//
		// For PropertyDeclaration ancestors, we are owned by the
		// PropertyDeclaration ONLY if we came through its Initializer field
		// (and the initializer itself is not a function-like, in which case
		// the function takes over as its own code path). Walking through the
		// Name field — including ComputedPropertyName — keeps walking up so
		// computed-key expressions count toward the enclosing scope.
		findOwner := func(n *ast.Node) *ast.Node {
			cur := n.Parent
			prev := n
			for cur != nil {
				if ast.IsFunctionLikeDeclaration(cur) {
					return cur
				}
				if cur.Kind == ast.KindClassStaticBlockDeclaration {
					return cur
				}
				if cur.Kind == ast.KindPropertyDeclaration {
					propDecl := cur.AsPropertyDeclaration()
					if propDecl != nil &&
						propDecl.Initializer != nil &&
						propDecl.Initializer == prev &&
						!ast.IsFunctionLikeDeclaration(propDecl.Initializer) {
						return cur
					}
				}
				prev = cur
				cur = cur.Parent
			}
			return nil
		}

		increment := func(node *ast.Node) {
			owner := findOwner(node)
			if owner == nil {
				return
			}
			counters[owner]++
		}

		// Function-like start: seed the counter at 1 (one execution path).
		startFunc := func(node *ast.Node) {
			counters[node] = 1
		}
		endFunc := func(node *ast.Node) {
			complexity, ok := counters[node]
			if !ok {
				return
			}
			delete(counters, node)
			if complexity > threshold {
				name := utils.UpperCaseFirstASCII(getFunctionNameWithKind(node))
				loc := utils.GetFunctionHeadLoc(ctx.SourceFile, node)
				ctx.ReportRange(loc, makeMessage(name, complexity, threshold))
			}
		}

		startStaticBlock := func(node *ast.Node) {
			counters[node] = 1
		}
		endStaticBlock := func(node *ast.Node) {
			complexity, ok := counters[node]
			if !ok {
				return
			}
			delete(counters, node)
			if complexity > threshold {
				// Report at the `static` keyword token, matching ESLint's
				// `sourceCode.getFirstToken(node).loc`.
				trimmedStart := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
				tokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, trimmedStart)
				ctx.ReportRange(tokenRange, makeMessage("Class static block", complexity, threshold))
			}
		}

		// PropertyDeclaration: only push when the initializer exists and is
		// not itself a function-like (when the initializer is a function, the
		// function code path takes over). This mirrors ESLint's
		// "class-field-initializer" code-path origin.
		startProp := func(node *ast.Node) {
			propDecl := node.AsPropertyDeclaration()
			if propDecl == nil || propDecl.Initializer == nil ||
				ast.IsFunctionLikeDeclaration(propDecl.Initializer) {
				return
			}
			counters[node] = 1
		}
		endProp := func(node *ast.Node) {
			complexity, ok := counters[node]
			if !ok {
				return
			}
			delete(counters, node)
			if complexity > threshold {
				propDecl := node.AsPropertyDeclaration()
				if propDecl == nil || propDecl.Initializer == nil {
					return
				}
				loc := utils.TrimNodeTextRange(ctx.SourceFile, propDecl.Initializer)
				ctx.ReportRange(loc, makeMessage("Class field initializer", complexity, threshold))
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                              startFunc,
			rule.ListenerOnExit(ast.KindFunctionDeclaration):         endFunc,
			ast.KindFunctionExpression:                               startFunc,
			rule.ListenerOnExit(ast.KindFunctionExpression):          endFunc,
			ast.KindArrowFunction:                                    startFunc,
			rule.ListenerOnExit(ast.KindArrowFunction):               endFunc,
			ast.KindMethodDeclaration:                                startFunc,
			rule.ListenerOnExit(ast.KindMethodDeclaration):           endFunc,
			ast.KindGetAccessor:                                      startFunc,
			rule.ListenerOnExit(ast.KindGetAccessor):                 endFunc,
			ast.KindSetAccessor:                                      startFunc,
			rule.ListenerOnExit(ast.KindSetAccessor):                 endFunc,
			ast.KindConstructor:                                      startFunc,
			rule.ListenerOnExit(ast.KindConstructor):                 endFunc,
			ast.KindClassStaticBlockDeclaration:                      startStaticBlock,
			rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): endStaticBlock,
			ast.KindPropertyDeclaration:                              startProp,
			rule.ListenerOnExit(ast.KindPropertyDeclaration):         endProp,

			ast.KindCatchClause:           increment,
			ast.KindConditionalExpression: increment,
			ast.KindIfStatement:           increment,
			ast.KindWhileStatement:        increment,
			ast.KindDoStatement:           increment,
			ast.KindForStatement:          increment,
			ast.KindForInStatement:        increment,
			ast.KindForOfStatement:        increment,

			// LogicalExpression / logical-assignment AssignmentExpression both
			// collapse into BinaryExpression in tsgo; branch on the operator.
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin == nil || bin.OperatorToken == nil {
					return
				}
				switch bin.OperatorToken.Kind {
				case ast.KindAmpersandAmpersandToken,
					ast.KindBarBarToken,
					ast.KindQuestionQuestionToken,
					ast.KindAmpersandAmpersandEqualsToken,
					ast.KindBarBarEqualsToken,
					ast.KindQuestionQuestionEqualsToken:
					increment(node)
				}
			},

			// SwitchCase[test] in classic mode (CaseClause has an Expression;
			// DefaultClause is its own kind so it never reaches this listener).
			ast.KindCaseClause: func(node *ast.Node) {
				if !isModified {
					increment(node)
				}
			},
			// In modified mode, the SwitchStatement itself counts +1 and
			// individual cases do not.
			ast.KindSwitchStatement: func(node *ast.Node) {
				if isModified {
					increment(node)
				}
			},

			// Optional-chain segments: ESLint counts each `?.` once. In tsgo,
			// the QuestionDotToken on a single PropertyAccess / ElementAccess /
			// CallExpression marks a `?.` at that node.
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				pae := node.AsPropertyAccessExpression()
				if pae != nil && pae.QuestionDotToken != nil {
					increment(node)
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				eae := node.AsElementAccessExpression()
				if eae != nil && eae.QuestionDotToken != nil {
					increment(node)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				ce := node.AsCallExpression()
				if ce != nil && ce.QuestionDotToken != nil {
					increment(node)
				}
			},

			// AssignmentPattern equivalents: parameter defaults
			// (`function f(x = 1)`) and destructuring defaults
			// (`{x = 1}` / `[x = 1]` in either parameters or `let`/`const`).
			ast.KindParameter: func(node *ast.Node) {
				pd := node.AsParameterDeclaration()
				if pd != nil && pd.Initializer != nil {
					increment(node)
				}
			},
			ast.KindBindingElement: func(node *ast.Node) {
				be := node.AsBindingElement()
				if be != nil && be.Initializer != nil {
					increment(node)
				}
			},
		}
	},
}

func makeMessage(name string, complexity, threshold int) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "complex",
		Description: fmt.Sprintf(
			"%s has a complexity of %d. Maximum allowed is %d.",
			name, complexity, threshold,
		),
	}
}

// getFunctionNameWithKind mirrors ESLint's astUtils.getFunctionNameWithKind
// EXACTLY — used here in preference to utils.GetFunctionNameWithKind because
// the shared helper resolves names through VariableDeclaration parents
// (producing strings like "function 'func'" for `var func = function () {}`),
// while ESLint's complexity rule expects the bare "function" form. Other
// rslint rules accept the broader name resolution; the complexity rule's test
// suite locks in ESLint's narrower behavior.
func getFunctionNameWithKind(node *ast.Node) string {
	// Constructor short-circuit (matches ESLint). A KindConstructor is always
	// declared inside a class body in valid TS, so a single unconditional
	// check covers every case.
	if node.Kind == ast.KindConstructor {
		return "constructor"
	}

	parent := node.Parent
	if parent == nil {
		return "function"
	}

	tokens := []string{}

	// Modifier prefixes (static / private) — only meaningful on class
	// members. ESLint reads them off MethodDefinition / PropertyDefinition.
	// Object-method shorthand (`{ foo() {} }`) inherits the "method"
	// classification but never carries static / private modifiers.
	isClassMember := isDirectClassMember(node)
	isClassFieldValue := isClassFieldInitializer(node)
	if isClassMember || isClassFieldValue {
		owner := node
		if isClassFieldValue {
			owner = parent
		}
		if ast.HasSyntacticModifier(owner, ast.ModifierFlagsStatic) {
			tokens = append(tokens, "static")
		}
		if name := owner.Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
			tokens = append(tokens, "private")
		}
	}

	flags := ast.GetFunctionFlags(node)
	if flags&ast.FunctionFlagsAsync != 0 {
		tokens = append(tokens, "async")
	}
	if flags&ast.FunctionFlagsGenerator != 0 {
		tokens = append(tokens, "generator")
	}

	// Function-kind word. `MethodDeclaration` covers both class methods and
	// object-method shorthand (`{ foo() {} }`) in tsgo, so it always
	// classifies as "method". Object methods written with the `key: value`
	// form (`{ foo: function () {} }` or `{ foo: () => {} }`) reach this
	// switch as a FunctionExpression / ArrowFunction whose parent is a
	// PropertyAssignment — also classified as "method" to match ESLint.
	switch {
	case node.Kind == ast.KindGetAccessor:
		tokens = append(tokens, "getter")
	case node.Kind == ast.KindSetAccessor:
		tokens = append(tokens, "setter")
	case node.Kind == ast.KindMethodDeclaration:
		tokens = append(tokens, "method")
	case parent.Kind == ast.KindPropertyAssignment:
		tokens = append(tokens, "method")
	case isClassFieldValue:
		tokens = append(tokens, "method")
	case node.Kind == ast.KindArrowFunction:
		tokens = append(tokens, "arrow", "function")
	default:
		tokens = append(tokens, "function")
	}

	// Name resolution. ESLint only reads the name from the property key
	// when the parent is Property / MethodDefinition / PropertyDefinition,
	// or from `node.id` for plain function declarations / expressions.
	// Notably, it does NOT walk up to a VariableDeclaration to recover a
	// variable binding name — `var f = function () {}` reports as
	// "function" without the "f".
	//
	// Object-method shorthand (`{ foo() {} }`) lives in tsgo as a
	// MethodDeclaration whose parent is the ObjectLiteralExpression itself
	// (no intermediate Property node), so we read the name off the method
	// declaration directly when the kind is MethodDeclaration.
	switch {
	case parent.Kind == ast.KindPropertyAssignment:
		// `{ foo: function () {} }` / `{ foo: () => {} }` — name lives on
		// the property assignment.
		appendNameFromOwner(&tokens, parent, node)
	case isClassFieldValue:
		appendNameFromOwner(&tokens, parent, node)
	case node.Kind == ast.KindMethodDeclaration ||
		node.Kind == ast.KindGetAccessor ||
		node.Kind == ast.KindSetAccessor:
		// Class methods/accessors AND object-literal method shorthand /
		// getter / setter — name lives on the node itself in tsgo (no
		// intermediate Property node).
		appendNameFromOwner(&tokens, node, node)
	default:
		if id := nodeIdentifier(node); id != "" {
			tokens = append(tokens, fmt.Sprintf("'%s'", id))
		}
	}

	return joinTokens(tokens)
}

// appendNameFromOwner mirrors ESLint's name-resolution preference order
// inside getFunctionNameWithKind: prefer the property key (private identifier
// → static-name → empty), and fall back to `node.id` only for nameless
// FunctionExpressions on a Property.
func appendNameFromOwner(tokens *[]string, owner *ast.Node, node *ast.Node) {
	if name := owner.Name(); name != nil {
		if name.Kind == ast.KindPrivateIdentifier {
			*tokens = append(*tokens, fmt.Sprintf("'%s'", name.AsPrivateIdentifier().Text))
			return
		}
		if s, ok := utils.GetStaticPropertyName(name); ok {
			*tokens = append(*tokens, fmt.Sprintf("'%s'", s))
			return
		}
	}
	if id := nodeIdentifier(node); id != "" {
		*tokens = append(*tokens, fmt.Sprintf("'%s'", id))
	}
}

// nodeIdentifier returns the identifier on the function node itself
// (`node.id` in ESTree), or the empty string when there is none.
func nodeIdentifier(node *ast.Node) string {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		if n := node.AsFunctionDeclaration().Name(); n != nil && n.Kind == ast.KindIdentifier {
			return n.AsIdentifier().Text
		}
	case ast.KindFunctionExpression:
		if n := node.AsFunctionExpression().Name(); n != nil && n.Kind == ast.KindIdentifier {
			return n.AsIdentifier().Text
		}
	}
	return ""
}

// isDirectClassMember reports whether the function-like node is a method /
// accessor whose parent is the enclosing class (not its initializer).
func isDirectClassMember(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind != ast.KindClassDeclaration && parent.Kind != ast.KindClassExpression {
		return false
	}
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return true
	}
	return false
}

// isClassFieldInitializer reports whether the function-like node is the
// direct initializer of a class field (e.g. `class C { x = () => {} }`).
// Mirrors ESLint's `parent.type === "PropertyDefinition" && parent.value === node`.
func isClassFieldInitializer(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindPropertyDeclaration {
		return false
	}
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return parent.AsPropertyDeclaration().Initializer == node
	}
	return false
}

func joinTokens(tokens []string) string {
	return strings.Join(tokens, " ")
}

// parseOptions extracts the threshold and modified-variant flag from the
// rule's options. Mirrors ESLint's `option.maximum || option.max` coercion:
// when both are present, `maximum` wins only if truthy. The bare-number form
// (`["error", 2]` or just `2` from the CLI) sets the threshold directly.
func parseOptions(options any) (int, bool) {
	threshold := defaultThreshold
	isModified := false

	if options == nil {
		return threshold, isModified
	}
	// Bare number / array-wrapped number forms.
	if arr, ok := options.([]interface{}); ok {
		if len(arr) == 0 {
			return threshold, isModified
		}
		options = arr[0]
	}
	if n, ok := utils.CoerceInt(options); ok {
		return n, isModified
	}
	// Object form: `{ max?: number, maximum?: number, variant?: "classic"|"modified" }`.
	m := utils.GetOptionsMap(options)
	if m == nil {
		return threshold, isModified
	}
	_, hasMaximum := m["maximum"]
	_, hasMax := m["max"]
	if hasMaximum {
		if v, ok := utils.CoerceInt(m["maximum"]); ok && v != 0 {
			threshold = v
		} else if hasMax {
			if v, ok := utils.CoerceInt(m["max"]); ok {
				threshold = v
			}
		}
	} else if hasMax {
		if v, ok := utils.CoerceInt(m["max"]); ok {
			threshold = v
		}
	}
	if v, ok := m["variant"].(string); ok && v == "modified" {
		isModified = true
	}
	return threshold, isModified
}
