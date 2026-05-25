package no_useless_computed_key

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type Options struct {
	EnforceForClassMembers bool
}

func parseOptions(options any) Options {
	opts := Options{EnforceForClassMembers: true}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["enforceForClassMembers"].(bool); ok {
			opts.EnforceForClassMembers = v
		}
	}
	return opts
}

// isIdentPart reports whether c can appear inside a JS identifier (ASCII).
// Used to decide if two adjacent tokens would fuse when we delete the
// surrounding brackets. Mirrors `astUtils.canTokensBeAdjacent` enough for
// the cases the rule can emit (tokens before `[` are keywords / `*` / `{`;
// the key text starts with either an identifier char, a digit, a quote,
// or `.`).
func isIdentPart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_' || c == '$'
}

// isObjectLiteralAssignmentPattern reports whether an ObjectLiteralExpression
// is being used as a destructuring target — either directly as the LHS of
// `=`, the Initializer of a for-in/of statement, or nested inside another
// such pattern. tsgo reuses ObjectLiteralExpression for assignment patterns
// (unlike ESTree's ObjectPattern), so we have to recover the distinction
// from context to match ESLint's Property-vs-Property behavior: the
// `__proto__` carve-out applies only to value-position object literals.
func isObjectLiteralAssignmentPattern(objLit *ast.Node) bool {
	cur := objLit
	for {
		parent := cur.Parent
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			cur = parent
		case ast.KindBinaryExpression:
			be := parent.AsBinaryExpression()
			if be == nil || be.OperatorToken == nil {
				return false
			}
			return be.OperatorToken.Kind == ast.KindEqualsToken && be.Left == cur
		case ast.KindForInStatement, ast.KindForOfStatement:
			fs := parent.AsForInOrOfStatement()
			return fs != nil && fs.Initializer == cur
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
			// Nested in an outer object pattern (e.g. `({y: {['x']: a}} = b)`).
			// Walk up to the enclosing ObjectLiteralExpression, then continue.
			cur = parent.Parent
			if cur == nil {
				return false
			}
		case ast.KindArrayLiteralExpression, ast.KindSpreadAssignment, ast.KindSpreadElement:
			cur = parent
		default:
			return false
		}
	}
}

// hasCommentsBetween reports whether a line or block comment starts inside
// `[start, end)`, skipping over the contents of string and template literals
// so that an embedded `//` or `/*` sequence doesn't false-positive.
func hasCommentsBetween(text string, start, end int) bool {
	i := start
	for i < end {
		c := text[i]
		switch c {
		case '/':
			if i+1 < end && (text[i+1] == '/' || text[i+1] == '*') {
				return true
			}
			i++
		case '\'', '"':
			quote := c
			i++
			for i < end && text[i] != quote {
				if text[i] == '\\' && i+1 < end {
					i += 2
					continue
				}
				i++
			}
			if i < end {
				i++
			}
		case '`':
			i++
			for i < end && text[i] != '`' {
				if text[i] == '\\' && i+1 < end {
					i += 2
					continue
				}
				i++
			}
			if i < end {
				i++
			}
		default:
			i++
		}
	}
	return false
}

// https://eslint.org/docs/latest/rules/no-useless-computed-key
var NoUselessComputedKeyRule = rule.Rule{
	Name: "no-useless-computed-key",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// check inspects a container node (PropertyAssignment / Method /
		// PropertyDeclaration / BindingElement), determines whether it has
		// a useless computed key, and reports + fixes accordingly.
		//
		// Listening on containers (not ComputedPropertyName) is required so
		// the rule also fires inside destructuring patterns: rslint's
		// `patternVisitor` does not recurse into property-name positions.
		check := func(container *ast.Node, cpn *ast.Node) {
			computed := cpn.AsComputedPropertyName()
			if computed == nil || computed.Expression == nil {
				return
			}

			// Only literal values (string / number) qualify. Parentheses
			// around the literal are transparent to the check, matching
			// ESLint's paren-insensitive AST.
			inner := ast.SkipParentheses(computed.Expression)
			var value string
			isString := false
			switch inner.Kind {
			case ast.KindStringLiteral:
				value = inner.AsStringLiteral().Text
				isString = true
			case ast.KindNumericLiteral:
				value = inner.AsNumericLiteral().Text
			default:
				return
			}

			switch container.Kind {
			case ast.KindPropertyAssignment:
				gp := container.Parent
				if gp == nil || gp.Kind != ast.KindObjectLiteralExpression {
					return
				}
				// In a value-position object literal, `{ ["__proto__"]: v }`
				// defines an own property whereas `{ __proto__: v }` sets
				// the prototype — the computed form carries distinct
				// meaning. In a destructuring pattern (assignment LHS,
				// for-in/of target), no such distinction exists, so the
				// computed form is genuinely useless.
				if isString && value == "__proto__" && !isObjectLiteralAssignmentPattern(gp) {
					return
				}
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				gp := container.Parent
				if gp == nil {
					return
				}
				switch gp.Kind {
				case ast.KindObjectLiteralExpression:
					if isString && value == "__proto__" && !isObjectLiteralAssignmentPattern(gp) {
						return
					}
				case ast.KindClassDeclaration, ast.KindClassExpression:
					if !opts.EnforceForClassMembers {
						return
					}
					// typescript-eslint emits `TSAbstractMethodDefinition` for
					// abstract methods/accessors, which ESLint's core rule
					// does not listen for — so the rule is a no-op there.
					// Match that behavior exactly.
					if ast.HasSyntacticModifier(container, ast.ModifierFlagsAbstract) {
						return
					}
					if ast.IsStatic(container) {
						if isString && value == "prototype" {
							return
						}
					} else {
						if isString && value == "constructor" {
							return
						}
					}
				default:
					return
				}
			case ast.KindPropertyDeclaration:
				gp := container.Parent
				if gp == nil {
					return
				}
				if gp.Kind != ast.KindClassDeclaration && gp.Kind != ast.KindClassExpression {
					return
				}
				if !opts.EnforceForClassMembers {
					return
				}
				// Same framework-level gaps as methods:
				//   - `abstract` fields → `TSAbstractPropertyDefinition`
				//   - `accessor` fields (auto-accessor) → `AccessorProperty`
				// Neither matches ESLint's `PropertyDefinition` selector, so
				// ESLint's rule skips them. Skip here too.
				if ast.HasSyntacticModifier(container, ast.ModifierFlagsAbstract) {
					return
				}
				if ast.HasSyntacticModifier(container, ast.ModifierFlagsAccessor) {
					return
				}
				if ast.IsStatic(container) {
					if isString && (value == "constructor" || value == "prototype") {
						return
					}
				} else {
					if isString && value == "constructor" {
						return
					}
				}
			case ast.KindBindingElement:
				// Destructuring pattern — always report.
			default:
				return
			}

			sf := ctx.SourceFile
			sourceText := sf.Text()

			leftBracketPos := scanner.SkipTrivia(sourceText, cpn.Pos())
			endPos := cpn.End()
			keyStart := scanner.SkipTrivia(sourceText, inner.Pos())
			keyEnd := inner.End()
			keyRaw := sourceText[keyStart:keyEnd]

			msg := rule.RuleMessage{
				Id:          "unnecessarilyComputedProperty",
				Description: fmt.Sprintf("Unnecessarily computed property [%s] found.", keyRaw),
			}

			// Suppress auto-fix when a comment sits anywhere between the
			// brackets — preserving comments across a token replacement
			// would require splicing them back in and isn't worth the
			// complexity, matching ESLint's behavior.
			if hasCommentsBetween(sourceText, leftBracketPos+1, endPos-1) {
				ctx.ReportNode(container, msg)
				return
			}

			replacement := keyRaw
			if leftBracketPos > 0 && len(keyRaw) > 0 {
				prev := sourceText[leftBracketPos-1]
				// Only add a separator space when the previous token and
				// the key would fuse into one identifier (e.g. `get` + `2`
				// → `get2`). `get` + `'foo'` is fine because `'` can't
				// continue an identifier.
				if isIdentPart(prev) && isIdentPart(keyRaw[0]) {
					replacement = " " + keyRaw
				}
			}

			fix := rule.RuleFixReplaceRange(
				core.NewTextRange(leftBracketPos, endPos),
				replacement,
			)
			ctx.ReportNodeWithFixes(container, msg, fix)
		}

		nameIfComputed := func(n *ast.Node) *ast.Node {
			if n == nil {
				return nil
			}
			name := n.Name()
			if name == nil || name.Kind != ast.KindComputedPropertyName {
				return nil
			}
			return name
		}

		handle := func(node *ast.Node) {
			if cpn := nameIfComputed(node); cpn != nil {
				check(node, cpn)
			}
		}

		bindingElementHandle := func(node *ast.Node) {
			be := node.AsBindingElement()
			if be == nil || be.PropertyName == nil {
				return
			}
			if be.PropertyName.Kind != ast.KindComputedPropertyName {
				return
			}
			check(node, be.PropertyName)
		}

		return rule.RuleListeners{
			ast.KindPropertyAssignment:  handle,
			ast.KindMethodDeclaration:   handle,
			ast.KindGetAccessor:         handle,
			ast.KindSetAccessor:         handle,
			ast.KindPropertyDeclaration: handle,
			ast.KindBindingElement:      bindingElementHandle,
		}
	},
}
