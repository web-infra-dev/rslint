package func_name_matching

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	tsscanner "github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed func_name_matching.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/func-name-matching
var FuncNameMatchingRule = rule.Rule{
	Name:   "func-name-matching",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		ecmaVersion := ctx.LanguageOptions.EffectiveECMAVersion()

		handleVariableDeclarator := func(node *ast.Node) {
			vd := node.AsVariableDeclaration()
			if node.Name().Kind != ast.KindIdentifier || vd.Initializer == nil {
				return
			}
			init := ast.SkipParentheses(vd.Initializer)
			if init.Kind != ast.KindFunctionExpression {
				return
			}
			fnNameNode := init.Name()
			if fnNameNode == nil {
				return
			}
			varName := node.Name().AsIdentifier().Text
			funcName := fnNameNode.AsIdentifier().Text
			if shouldWarn(opts, varName, funcName) {
				reportMismatch(ctx, opts, node, varName, funcName, false)
			}
		}

		handleAssignment := func(node *ast.Node) {
			be := node.AsBinaryExpression()
			if !ast.IsAssignmentOperator(be.OperatorToken.Kind) {
				return
			}

			// NOTE: ESTree has no parenthesized-expression node, so
			// `foo = (function bar() {})` already sees a bare
			// FunctionExpression on the right upstream; tsgo keeps the
			// wrapper explicit and needs the unwrap.
			right := ast.SkipParentheses(be.Right)
			if right.Kind != ast.KindFunctionExpression {
				return
			}

			left := ast.SkipParentheses(be.Left)
			if left == nil {
				return
			}

			if left.Kind == ast.KindElementAccessExpression {
				arg := ast.SkipParentheses(left.AsElementAccessExpression().ArgumentExpression)
				if arg == nil || !isEstreeLiteralKind(arg.Kind) {
					return
				}
			}

			if !opts.includeCommonJSModuleExports && utils.IsSpecificMemberAccess(left, "module", "exports") {
				return
			}

			if left.Kind != ast.KindIdentifier &&
				left.Kind != ast.KindPropertyAccessExpression &&
				left.Kind != ast.KindElementAccessExpression {
				return
			}

			isProp := left.Kind == ast.KindPropertyAccessExpression || left.Kind == ast.KindElementAccessExpression

			var name string
			switch left.Kind {
			case ast.KindPropertyAccessExpression:
				name, _ = utils.GetStaticPropertyName(left.AsPropertyAccessExpression().Name())
			case ast.KindElementAccessExpression:
				arg := ast.SkipParentheses(left.AsElementAccessExpression().ArgumentExpression)
				name = staticLiteralText(arg)
			default: // ast.KindIdentifier
				name = left.AsIdentifier().Text
			}
			if name == "" {
				return
			}

			fnNameNode := right.Name()
			if fnNameNode == nil {
				return
			}
			funcName := fnNameNode.AsIdentifier().Text

			// NOTE: unlike the Property/PropertyDefinition listener below,
			// upstream calls its identifier-shape check here without an
			// ecmaVersion argument, which its helper treats as ES5
			// regardless of the file's configured language version.
			if !isIdentifierName(name, 5) {
				return
			}
			if shouldWarn(opts, name, funcName) {
				reportMismatch(ctx, opts, node, name, funcName, isProp)
			}
		}

		handlePropertyLike := func(node *ast.Node) {
			var rawInit *ast.Node
			switch node.Kind {
			case ast.KindPropertyAssignment:
				rawInit = node.AsPropertyAssignment().Initializer
			case ast.KindPropertyDeclaration:
				rawInit = node.AsPropertyDeclaration().Initializer
			default:
				return
			}
			if rawInit == nil {
				return
			}
			init := ast.SkipParentheses(rawInit)
			if init.Kind != ast.KindFunctionExpression {
				return
			}
			fnNameNode := init.Name()
			if fnNameNode == nil {
				return
			}
			functionName := fnNameNode.AsIdentifier().Text

			keyNode := node.Name()

			if keyNode.Kind == ast.KindIdentifier {
				propertyName := keyNode.AsIdentifier().Text

				if opts.considerPropertyDescriptor && propertyName == "value" &&
					node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
					switch {
					case isPropertyCall(nthAncestor(node, 2), "Object", "defineProperty"),
						isPropertyCall(nthAncestor(node, 2), "Reflect", "defineProperty"):
						call := nthAncestor(node, 2).AsCallExpression()
						if call.Arguments != nil && len(call.Arguments.Nodes) > 1 {
							propArg := call.Arguments.Nodes[1]
							if propArg.Kind == ast.KindStringLiteral {
								propValue := propArg.AsStringLiteral().Text
								if shouldWarn(opts, propValue, functionName) {
									reportMismatch(ctx, opts, node, propValue, functionName, true)
								}
							}
						}
					case isPropertyCall(nthAncestor(node, 4), "Object", "defineProperties"),
						isPropertyCall(nthAncestor(node, 4), "Object", "create"):
						outerProp := nthAncestor(node, 2)
						if outerProp.Name().Kind == ast.KindIdentifier {
							outerName := outerProp.Name().AsIdentifier().Text
							if shouldWarn(opts, outerName, functionName) {
								reportMismatch(ctx, opts, node, outerName, functionName, true)
							}
						}
					default:
						if shouldWarn(opts, propertyName, functionName) {
							reportMismatch(ctx, opts, node, propertyName, functionName, true)
						}
					}
					return
				}

				if shouldWarn(opts, propertyName, functionName) {
					reportMismatch(ctx, opts, node, propertyName, functionName, true)
				}
				return
			}

			if keyNode.Kind == ast.KindStringLiteral {
				keyValue := keyNode.AsStringLiteral().Text
				if isIdentifierName(keyValue, ecmaVersion) && shouldWarn(opts, keyValue, functionName) {
					reportMismatch(ctx, opts, node, keyValue, functionName, true)
				}
				return
			}

			if keyNode.Kind == ast.KindComputedPropertyName {
				inner := ast.SkipParentheses(keyNode.AsComputedPropertyName().Expression)
				if inner != nil && inner.Kind == ast.KindStringLiteral {
					keyValue := inner.AsStringLiteral().Text
					if isIdentifierName(keyValue, ecmaVersion) && shouldWarn(opts, keyValue, functionName) {
						reportMismatch(ctx, opts, node, keyValue, functionName, true)
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: handleVariableDeclarator,
			ast.KindBinaryExpression:    handleAssignment,
			ast.KindPropertyAssignment:  handlePropertyLike,
			ast.KindPropertyDeclaration: handlePropertyLike,
		}
	},
}

type funcNameMatchingOptions struct {
	nameMatches                  string
	considerPropertyDescriptor   bool
	includeCommonJSModuleExports bool
}

func parseOptions(raw []any) funcNameMatchingOptions {
	opts := funcNameMatchingOptions{nameMatches: "always"}

	var optsObj map[string]any
	if len(raw) > 0 {
		if s, ok := raw[0].(string); ok {
			opts.nameMatches = s
		} else if m, ok := raw[0].(map[string]any); ok {
			optsObj = m
		}
	}
	if optsObj == nil && len(raw) > 1 {
		if m, ok := raw[1].(map[string]any); ok {
			optsObj = m
		}
	}
	if optsObj != nil {
		if v, ok := optsObj["considerPropertyDescriptor"].(bool); ok {
			opts.considerPropertyDescriptor = v
		}
		if v, ok := optsObj["includeCommonJSModuleExports"].(bool); ok {
			opts.includeCommonJSModuleExports = v
		}
	}
	return opts
}

func shouldWarn(opts funcNameMatchingOptions, x, y string) bool {
	if opts.nameMatches == "never" {
		return x == y
	}
	return x != y
}

func reportMismatch(ctx rule.RuleContext, opts funcNameMatchingOptions, node *ast.Node, name, funcName string, isProp bool) {
	var id, desc string
	switch {
	case opts.nameMatches == "always" && isProp:
		id = "matchProperty"
		desc = fmt.Sprintf("Function name `%s` should match property name `%s`.", funcName, name)
	case opts.nameMatches == "always":
		id = "matchVariable"
		desc = fmt.Sprintf("Function name `%s` should match variable name `%s`.", funcName, name)
	case isProp:
		id = "notMatchProperty"
		desc = fmt.Sprintf("Function name `%s` should not match property name `%s`.", funcName, name)
	default:
		id = "notMatchVariable"
		desc = fmt.Sprintf("Function name `%s` should not match variable name `%s`.", funcName, name)
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          id,
		Description: desc,
		Data:        map[string]string{"name": name, "funcName": funcName},
	})
}

// isPropertyCall reports whether node is a CallExpression invoking
// `<objName>.<methodName>(...)`, mirroring upstream's own isPropertyCall
// helper. utils.IsSpecificMemberAccess already unwraps parenthesized and
// optional-chained callees.
func isPropertyCall(node *ast.Node, objName, methodName string) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	return utils.IsSpecificMemberAccess(node.AsCallExpression().Expression, objName, methodName)
}

// nthAncestor walks n Parent links up from node, returning nil if the chain
// runs out early.
func nthAncestor(node *ast.Node, n int) *ast.Node {
	for i := 0; i < n && node != nil; i++ {
		node = node.Parent
	}
	return node
}

// isEstreeLiteralKind reports whether kind is one of the node kinds ESLint's
// ESTree AST reports as `Literal` (excluding template literals, which ESTree
// types as `TemplateLiteral`).
func isEstreeLiteralKind(kind ast.Kind) bool {
	switch kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral, ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	}
	return false
}

// staticLiteralText stringifies one of the isEstreeLiteralKind node kinds the
// same way ESLint's astUtils.getStaticPropertyName does (String(value)).
func staticLiteralText(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		return utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text)
	case ast.KindBigIntLiteral:
		return utils.NormalizeBigIntLiteral(node.AsBigIntLiteral().Text)
	case ast.KindRegularExpressionLiteral:
		return node.AsRegularExpressionLiteral().Text
	case ast.KindNullKeyword:
		return "null"
	case ast.KindTrueKeyword:
		return "true"
	case ast.KindFalseKeyword:
		return "false"
	}
	return ""
}

// es6ReservedWords are the ECMAScript keywords and literal reserved words
// esutils' keyword.isReservedWordES6 rejects when called without its
// `strict` flag (as this rule always calls it) — the additional strict-mode
// reserved words (implements/interface/package/private/protected/public/
// static/let) never apply here.
var es6ReservedWords = map[string]bool{
	"if": true, "in": true, "do": true,
	"var": true, "for": true, "new": true, "try": true,
	"this": true, "else": true, "case": true, "void": true, "with": true, "enum": true,
	"while": true, "break": true, "catch": true, "throw": true, "const": true, "yield": true, "class": true, "super": true,
	"return": true, "typeof": true, "delete": true, "switch": true, "export": true, "import": true,
	"default": true, "finally": true, "extends": true,
	"function": true, "continue": true, "debugger": true,
	"instanceof": true,
	"null":       true, "true": true, "false": true,
}

// isIdentifierName mirrors esutils' keyword.isIdentifierES5 / isIdentifierES6
// as this rule calls them: identifier-shape validity plus non-strict-mode
// reserved-word rejection. "yield" is the only word the two editions
// disagree on: esutils' ES5 helper explicitly special-cases it back to
// non-reserved outside strict mode, while its ES6 keyword table always
// treats it as reserved.
//
// Divergence: character-shape validity is answered by tsgo's own scanner
// tables (github.com/microsoft/typescript-go/shim/scanner.IsValidIdentifier),
// which track a current Unicode identifier snapshot rather than the frozen
// ES5.1/Unicode-v9 tables esutils ships for its ES5 path. A small number of
// codepoints added to Unicode's identifier properties after that snapshot
// (e.g. U+1885) are treated as valid identifier characters here at every
// ecmaVersion, where ESLint's ES5 mode rejects them.
func isIdentifierName(name string, ecmaVersion int) bool {
	if !tsscanner.IsValidIdentifier(name) {
		return false
	}
	if name == "yield" {
		return ecmaVersion < 2015
	}
	return !es6ReservedWords[name]
}
