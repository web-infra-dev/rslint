package no_thenable

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDObject = "no-thenable-object"
	messageIDExport = "no-thenable-export"
	messageIDClass  = "no-thenable-class"
)

var (
	messageObject = rule.RuleMessage{
		Id:          messageIDObject,
		Description: "Do not add `then` to an object.",
	}
	messageExport = rule.RuleMessage{
		Id:          messageIDExport,
		Description: "Do not export `then`.",
	}
	messageClass = rule.RuleMessage{
		Id:          messageIDClass,
		Description: "Do not add `then` to a class.",
	}
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/no-thenable.md
var NoThenableRule = rule.Rule{
	Name: "unicorn/no-thenable",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		state := &ruleState{
			ctx:           ctx,
			staticStrings: utils.NewStaticStringEvaluatorWithSourceFile(ctx.TypeChecker, ctx.SourceFile),
		}

		return rule.RuleListeners{
			ast.KindObjectLiteralExpression: state.checkObjectExpression,
			ast.KindPropertyDeclaration:     state.checkClassMember,
			ast.KindMethodDeclaration:       state.checkClassMember,
			ast.KindGetAccessor:             state.checkClassMember,
			ast.KindSetAccessor:             state.checkClassMember,
			ast.KindBinaryExpression:        state.checkAssignmentExpression,
			ast.KindCallExpression:          state.checkCallExpression,
			ast.KindExportDeclaration:       state.checkExportDeclaration,
			ast.KindFunctionDeclaration:     state.checkExportedDeclaration,
			ast.KindClassDeclaration:        state.checkExportedDeclaration,
			ast.KindVariableStatement:       state.checkExportedVariableStatement,
		}
	},
}

type ruleState struct {
	ctx           rule.RuleContext
	staticStrings *utils.StaticStringEvaluator
}

func (state *ruleState) checkObjectExpression(node *ast.Node) {
	objectExpression := node.AsObjectLiteralExpression()
	if objectExpression == nil || objectExpression.Properties == nil {
		return
	}

	for _, property := range objectExpression.Properties.Nodes {
		name := objectPropertyName(property)
		if name == nil || !state.isThenPropertyName(name) {
			continue
		}

		state.ctx.ReportNode(reportPropertyNameNode(name), messageObject)
	}
}

func (state *ruleState) checkClassMember(node *ast.Node) {
	if node.Parent == nil || (node.Parent.Kind != ast.KindClassDeclaration && node.Parent.Kind != ast.KindClassExpression) {
		return
	}

	name := node.Name()
	if name == nil || !state.isThenPropertyName(name) {
		return
	}

	state.ctx.ReportNode(reportPropertyNameNode(name), messageClass)
}

func (state *ruleState) checkAssignmentExpression(node *ast.Node) {
	binaryExpression := node.AsBinaryExpression()
	if binaryExpression == nil || binaryExpression.OperatorToken == nil ||
		!ast.IsAssignmentOperator(binaryExpression.OperatorToken.Kind) {
		return
	}

	left := ast.SkipParentheses(binaryExpression.Left)
	reportNode, ok := state.thenAccessReportNode(left)
	if !ok {
		return
	}

	state.ctx.ReportNode(reportNode, messageObject)
}

func (state *ruleState) checkCallExpression(node *ast.Node) {
	state.checkDefinePropertyCall(node)
	state.checkFromEntriesCall(node)
}

func (state *ruleState) checkDefinePropertyCall(node *ast.Node) {
	if !isNonOptionalMethodCall(node, []string{"Object", "Reflect"}, "defineProperty", 3, -1) {
		return
	}

	args := node.Arguments()
	if ast.IsSpreadElement(args[0]) {
		return
	}

	secondArgument := args[1]
	if !state.isStringThen(secondArgument) {
		return
	}

	state.ctx.ReportNode(reportExpressionNode(secondArgument), messageObject)
}

func (state *ruleState) checkFromEntriesCall(node *ast.Node) {
	if !isNonOptionalMethodCall(node, []string{"Object"}, "fromEntries", -1, 1) {
		return
	}

	args := node.Arguments()
	if len(args) != 1 || ast.IsSpreadElement(args[0]) {
		return
	}

	firstArgument := ast.SkipParentheses(args[0])
	if firstArgument == nil || firstArgument.Kind != ast.KindArrayLiteralExpression {
		return
	}

	outerElements := firstArgument.AsArrayLiteralExpression().Elements
	if outerElements == nil {
		return
	}

	for _, pair := range outerElements.Nodes {
		pair = ast.SkipParentheses(pair)
		if pair == nil || pair.Kind != ast.KindArrayLiteralExpression {
			continue
		}

		elements := pair.AsArrayLiteralExpression().Elements
		if elements == nil || len(elements.Nodes) == 0 {
			continue
		}

		key := elements.Nodes[0]
		if key == nil || ast.IsSpreadElement(key) || !state.isStringThen(key) {
			continue
		}

		state.ctx.ReportNode(reportExpressionNode(key), messageObject)
	}
}

func (state *ruleState) checkExportDeclaration(node *ast.Node) {
	exportDeclaration := node.AsExportDeclaration()
	if exportDeclaration == nil || exportDeclaration.IsTypeOnly ||
		exportDeclaration.ExportClause == nil ||
		exportDeclaration.ExportClause.Kind != ast.KindNamedExports {
		return
	}

	namedExports := exportDeclaration.ExportClause.AsNamedExports()
	if namedExports == nil || namedExports.Elements == nil {
		return
	}

	for _, specifierNode := range namedExports.Elements.Nodes {
		specifier := specifierNode.AsExportSpecifier()
		if specifier == nil || specifier.IsTypeOnly {
			continue
		}

		exported := specifier.Name()
		if isIdentifierNamed(exported, "then") {
			state.ctx.ReportNode(exported, messageExport)
		}
	}
}

func (state *ruleState) checkExportedDeclaration(node *ast.Node) {
	if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) ||
		ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault) ||
		!isIdentifierNamed(node.Name(), "then") {
		return
	}

	state.ctx.ReportNode(node.Name(), messageExport)
}

func (state *ruleState) checkExportedVariableStatement(node *ast.Node) {
	if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
		return
	}

	variableStatement := node.AsVariableStatement()
	if variableStatement == nil || variableStatement.DeclarationList == nil {
		return
	}

	declarationList := variableStatement.DeclarationList.AsVariableDeclarationList()
	if declarationList == nil || declarationList.Declarations == nil {
		return
	}

	for _, declarationNode := range declarationList.Declarations.Nodes {
		declaration := declarationNode.AsVariableDeclaration()
		if declaration == nil || declaration.Name() == nil {
			continue
		}

		utils.CollectBindingNames(declaration.Name(), func(identifier *ast.Node, name string) {
			if name == "then" {
				state.ctx.ReportNode(identifier, messageExport)
			}
		})
	}
}

func objectPropertyName(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindPropertyAssignment,
		ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return node.Name()
	default:
		return nil
	}
}

func (state *ruleState) isThenPropertyName(name *ast.Node) bool {
	value, ok := state.staticPropertyName(name)
	return ok && value == "then"
}

func (state *ruleState) staticPropertyName(name *ast.Node) (string, bool) {
	if name == nil {
		return "", false
	}

	if name.Kind == ast.KindComputedPropertyName {
		return state.staticStrings.Eval(name.AsComputedPropertyName().Expression)
	}

	return utils.GetStaticPropertyName(name)
}

func (state *ruleState) thenAccessReportNode(node *ast.Node) (*ast.Node, bool) {
	if node == nil {
		return nil, false
	}

	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if isIdentifierNamed(name, "then") {
			return name, true
		}
	case ast.KindElementAccessExpression:
		argument := node.AsElementAccessExpression().ArgumentExpression
		if state.isStringThen(argument) {
			return reportExpressionNode(argument), true
		}
	}

	return nil, false
}

func (state *ruleState) isStringThen(node *ast.Node) bool {
	value, ok := state.staticStrings.Eval(node)
	return ok && value == "then"
}

func reportPropertyNameNode(name *ast.Node) *ast.Node {
	if name != nil && name.Kind == ast.KindComputedPropertyName {
		return reportExpressionNode(name.AsComputedPropertyName().Expression)
	}
	return name
}

func reportExpressionNode(node *ast.Node) *ast.Node {
	unwrapped := utils.SkipAssertionsAndParens(node)
	if unwrapped != nil {
		return unwrapped
	}
	return node
}

// isNonOptionalMethodCall mirrors unicorn's isMethodCall with computed:false,
// optionalCall:false, and optionalMember:false. The wider
// utils.IsSpecificMemberAccess intentionally accepts bracket access and
// optional chains, which would diverge from this rule's upstream matcher.
func isNonOptionalMethodCall(node *ast.Node, objects []string, method string, minimumArguments int, argumentsLength int) bool {
	call := node.AsCallExpression()
	if call == nil || call.QuestionDotToken != nil {
		return false
	}

	args := node.Arguments()
	if minimumArguments >= 0 && len(args) < minimumArguments {
		return false
	}
	if argumentsLength >= 0 && len(args) != argumentsLength {
		return false
	}

	rawCallee := call.Expression
	callee := ast.SkipParentheses(rawCallee)
	if callee == nil || (rawCallee != callee && ast.IsOptionalChain(callee)) ||
		callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}

	propertyAccess := callee.AsPropertyAccessExpression()
	if propertyAccess == nil || propertyAccess.QuestionDotToken != nil ||
		!isIdentifierNamed(propertyAccess.Name(), method) {
		return false
	}

	object := ast.SkipParentheses(propertyAccess.Expression)
	return object != nil && object.Kind == ast.KindIdentifier &&
		slices.Contains(objects, object.AsIdentifier().Text)
}

func isIdentifierNamed(node *ast.Node, name string) bool {
	return node != nil && node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == name
}
