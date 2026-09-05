package no_unsafe_string_replacement

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-unsafe-string-replacement"

func unsafeReplacementMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: "Do not use a non-literal replacement value with `String#" + method + "()`.",
		Data:        map[string]string{"method": method},
	}
}

// NoUnsafeStringReplacementRule mirrors eslint-plugin-unicorn's
// no-unsafe-string-replacement. Upstream classifies the receiver syntactically
// before consulting type information; this rule always uses type information
// and declares RequiresTypeInfo, so it is skipped rather than guessing on
// files where none is available. See the rule doc for the divergences that
// follow from dropping the syntactic classifier.
var NoUnsafeStringReplacementRule = rule.Rule{
	Name:             "unicorn/no-unsafe-string-replacement",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				checkCall(ctx, node)
			},
		}
	},
}

func checkCall(ctx rule.RuleContext, node *ast.Node) {
	argumentsLength := 2
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods:             []string{"replace", "replaceAll"},
		ArgumentsLength:     &argumentsLength,
		RejectSpreadElement: true,
		AllowOptionalCall:   true,
		AllowOptionalMember: true,
	})
	if !ok {
		return
	}

	replacement := node.Arguments()[1]
	if isAllowedReplacement(ctx, replacement) ||
		isPlainObjectReplacement(ctx, replacement) ||
		unicornutil.IsKnownNonStringType(ctx, call.Object) {
		return
	}

	// ESTree does not preserve argument parentheses in the reported node range.
	// Keep TS assertion wrappers, which upstream includes in the range.
	reportNode := ast.SkipParentheses(replacement)
	method := call.Property.AsIdentifier().Text
	ctx.ReportNode(reportNode, unsafeReplacementMessage(method))
}

func isAllowedReplacement(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	case ast.KindTaggedTemplateExpression:
		return isStaticStringRawTaggedTemplate(ctx, node)
	default:
		return false
	}
}

func isStaticStringRawTaggedTemplate(ctx rule.RuleContext, node *ast.Node) bool {
	tagged := node.AsTaggedTemplateExpression()
	if tagged == nil || tagged.Template == nil ||
		tagged.Template.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return false
	}

	tag := ast.SkipParentheses(tagged.Tag)
	if tag == nil || !ast.IsPropertyAccessExpression(tag) || ast.IsOptionalChainRoot(tag) {
		return false
	}
	access := tag.AsPropertyAccessExpression()
	name := access.Name()
	object := ast.SkipParentheses(access.Expression)
	return name != nil && ast.IsIdentifier(name) && name.AsIdentifier().Text == "raw" &&
		isGlobalStringReference(ctx, object)
}

func isGlobalStringReference(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || !ast.IsIdentifier(node) || node.AsIdentifier().Text != "String" ||
		!ctx.Globals.Access("String").IsDeclared() {
		return false
	}
	if ctx.Refs != nil {
		if symbol := ctx.Refs.Resolve(node); symbol != nil {
			return !utils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
		}
	}
	return !utils.IsShadowed(node, "String")
}

var objectCoercionPropertyNames = map[string]bool{
	"__proto__": true,
	"toString":  true,
	"valueOf":   true,
}

func isPlainObjectReplacement(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if initializer := utils.GetConstVariableInitializer(node, ctx.TypeChecker); initializer != nil {
		node = utils.SkipAssertionsAndParens(initializer)
	}
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}

	object := node.AsObjectLiteralExpression()
	if object == nil || object.Properties == nil {
		return true
	}
	for _, property := range object.Properties.Nodes {
		if !isPlainObjectProperty(property) {
			return false
		}
	}
	return true
}

func isPlainObjectProperty(property *ast.Node) bool {
	if property == nil ||
		(property.Kind != ast.KindPropertyAssignment &&
			property.Kind != ast.KindShorthandPropertyAssignment) {
		return false
	}
	name := property.Name()
	if name == nil || name.Kind == ast.KindComputedPropertyName {
		return false
	}
	staticName, ok := utils.GetStaticPropertyName(name)
	if !ok && name.Kind == ast.KindBigIntLiteral {
		staticName, ok = utils.NormalizeBigIntLiteral(name.AsBigIntLiteral().Text), true
	}
	return ok && !objectCoercionPropertyNames[staticName]
}
