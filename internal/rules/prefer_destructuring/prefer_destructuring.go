package prefer_destructuring

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed prefer-destructuring.schema.json
var schemaJSON []byte

type destructuringConfig struct {
	array  bool
	object bool
}

type options struct {
	variableDeclarator   destructuringConfig
	assignmentExpression destructuringConfig
	enforceRenamed       bool
}

func parseOptions(raw []any) options {
	opts := options{
		variableDeclarator:   destructuringConfig{array: true, object: true},
		assignmentExpression: destructuringConfig{array: true, object: true},
	}

	if len(raw) > 0 && raw[0] != nil {
		// Supplying the first option replaces the defaults. ESLint accepts either
		// the legacy flat shape or separate settings for declarations and
		// assignments; omitted properties in either shape are disabled.
		opts.variableDeclarator = destructuringConfig{}
		opts.assignmentExpression = destructuringConfig{}

		if enabledTypes, ok := raw[0].(map[string]interface{}); ok {
			_, hasArray := enabledTypes["array"]
			_, hasObject := enabledTypes["object"]
			if hasArray || hasObject {
				config := parseDestructuringConfig(enabledTypes)
				opts.variableDeclarator = config
				opts.assignmentExpression = config
			} else {
				if config, ok := enabledTypes["VariableDeclarator"].(map[string]interface{}); ok {
					opts.variableDeclarator = parseDestructuringConfig(config)
				}
				if config, ok := enabledTypes["AssignmentExpression"].(map[string]interface{}); ok {
					opts.assignmentExpression = parseDestructuringConfig(config)
				}
			}
		}
	}

	if len(raw) > 1 {
		if enforcement, ok := raw[1].(map[string]interface{}); ok {
			if value, ok := enforcement["enforceForRenamedProperties"].(bool); ok {
				opts.enforceRenamed = value
			}
		}
	}

	return opts
}

func parseDestructuringConfig(raw map[string]interface{}) destructuringConfig {
	config := destructuringConfig{}
	if value, ok := raw["array"].(bool); ok {
		config.array = value
	}
	if value, ok := raw["object"].(bool); ok {
		config.object = value
	}
	return config
}

func preferDestructuringMessage(kind string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferDestructuring",
		Description: "Use " + kind + " destructuring.",
		Data:        map[string]string{"type": kind},
	}
}

func identifierName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return ""
	}
	return node.AsIdentifier().Text
}

func matchingObjectPropertyName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name == nil || !ast.IsIdentifier(name) {
			return "", false
		}
		return name.AsIdentifier().Text, true
	case ast.KindElementAccessExpression:
		argument := node.AsElementAccessExpression().ArgumentExpression
		if argument == nil {
			return "", false
		}
		argument = ast.SkipParentheses(argument)
		// Keep this narrower than utils.AccessExpressionStaticName: ESLint's
		// same-name branch accepts string Literals, but not template literals,
		// numeric keys, dynamic expressions, or TypeScript assertions.
		if !ast.IsStringLiteral(argument) {
			return "", false
		}
		return argument.AsStringLiteral().Text, true
	default:
		return "", false
	}
}

func shouldFix(leftNode, rightNode *ast.Node) bool {
	leftName := identifierName(leftNode)
	if leftName == "" || !ast.IsPropertyAccessExpression(rightNode) {
		return false
	}
	name := rightNode.AsPropertyAccessExpression().Name()
	return name != nil &&
		ast.IsIdentifier(name) &&
		leftName == name.AsIdentifier().Text
}

const precedenceOfAssignmentExpression = 1

func objectDestructuringFix(
	ctx rule.RuleContext,
	rightNode *ast.Node,
	reportNode *ast.Node,
) (rule.RuleFix, bool) {
	objectNode := utils.AccessExpressionObject(rightNode)
	if objectNode == nil {
		return rule.RuleFix{}, false
	}

	// ESTree does not retain ParenthesizedExpression nodes. Use the innermost
	// receiver for both source text and comment accounting. Since that receiver
	// is nested inside the declarator, a comment in either surrounding gap is
	// exactly a comment the replacement would discard.
	objectNode = ast.SkipParentheses(objectNode)
	reportRange := utils.TrimNodeTextRange(ctx.SourceFile, reportNode)
	objectRange := utils.TrimNodeTextRange(ctx.SourceFile, objectNode)
	comments := ctx.Comments.All()
	if utils.HasCommentInSpan(comments, reportRange.Pos(), objectRange.Pos()) ||
		utils.HasCommentInSpan(comments, objectRange.End(), reportRange.End()) {
		return rule.RuleFix{}, false
	}

	objectText := utils.TrimmedNodeText(ctx.SourceFile, objectNode)
	if utils.EslintLikePrecedence(objectNode) < precedenceOfAssignmentExpression {
		objectText = "(" + objectText + ")"
	}

	propertyName := rightNode.AsPropertyAccessExpression().Name().AsIdentifier().Text
	replacement := "{" + propertyName + "} = " + objectText
	return rule.RuleFixReplace(ctx.SourceFile, reportNode, replacement), true
}

func reportObject(
	ctx rule.RuleContext,
	leftNode, rightNode, reportNode *ast.Node,
	canFix bool,
) {
	message := preferDestructuringMessage("object")
	if canFix && shouldFix(leftNode, rightNode) {
		if fix, ok := objectDestructuringFix(ctx, rightNode, reportNode); ok {
			ctx.ReportNodeWithFixes(reportNode, message, fix)
			return
		}
	}
	ctx.ReportNode(reportNode, message)
}

func performCheck(
	ctx rule.RuleContext,
	opts options,
	leftNode, rightNode, reportNode *ast.Node,
	isVariableDeclarator bool,
) {
	if rightNode == nil {
		return
	}

	// ESTree discards parentheses around an expression, whereas tsgo keeps
	// them. Only parentheses are transparent here: TypeScript assertions
	// around the whole RHS remain non-member expressions, matching
	// @typescript-eslint/parser's ESTree output.
	rightNode = ast.SkipParentheses(rightNode)
	if !ast.IsAccessExpression(rightNode) {
		return
	}

	// ESLint sees an optional member expression behind a ChainExpression and
	// deliberately ignores it because the equivalent destructuring can throw.
	if ast.IsOptionalChain(rightNode) {
		return
	}

	if rightNode.Kind == ast.KindPropertyAccessExpression {
		name := rightNode.AsPropertyAccessExpression().Name()
		if name != nil && ast.IsPrivateIdentifier(name) {
			return
		}
	}

	objectNode := utils.AccessExpressionObject(rightNode)
	if objectNode != nil && ast.SkipParentheses(objectNode).Kind == ast.KindSuperKeyword {
		return
	}

	config := opts.assignmentExpression
	if isVariableDeclarator {
		config = opts.variableDeclarator
	}

	if utils.IsIntegerElementAccess(rightNode) {
		if config.array {
			ctx.ReportNode(reportNode, preferDestructuringMessage("array"))
		}
		return
	}

	if !config.object {
		return
	}

	if opts.enforceRenamed {
		reportObject(ctx, leftNode, rightNode, reportNode, isVariableDeclarator)
		return
	}

	leftName := identifierName(leftNode)
	propertyName, ok := matchingObjectPropertyName(rightNode)
	if !ok || leftName == "" || leftName != propertyName {
		return
	}
	reportObject(ctx, leftNode, rightNode, reportNode, isVariableDeclarator)
}

// PreferDestructuringRule requires destructuring instead of direct array index
// or object property access.
// https://eslint.org/docs/latest/rules/prefer-destructuring
var PreferDestructuringRule = rule.Rule{
	Name:   "prefer-destructuring",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration == nil || declaration.Initializer == nil {
					return
				}

				parent := node.Parent
				if parent != nil && parent.Kind == ast.KindVariableDeclarationList &&
					(ast.IsVarUsing(parent) || ast.IsVarAwaitUsing(parent)) {
					return
				}

				performCheck(
					ctx,
					opts,
					declaration.Name(),
					declaration.Initializer,
					node,
					true,
				)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsAssignmentExpression(node, true) {
					return
				}
				assignment := node.AsBinaryExpression()
				performCheck(ctx, opts, assignment.Left, assignment.Right, node, false)
			},
		}
	},
}
