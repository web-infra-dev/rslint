package no_document_cookie

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-document-cookie"

var globalObjectNames = [...]string{"globalThis", "window", "self", "global"}

type tracedValue uint8

const (
	tracedDocument tracedValue = iota
	tracedGlobalObject
)

// documentReferenceTracker follows the global document object through the
// flow-insensitive aliases supported by @eslint-community/eslint-utils'
// ReferenceTracker. The active stacks prevent cycles without suppressing the
// duplicate diagnostics that upstream emits when two roots reach one alias.
type documentReferenceTracker struct {
	ctx               rule.RuleContext
	identifiersByName map[string][]*ast.Node
	propertyEvaluator *utils.StaticStringEvaluator
	variableStack     map[*ast.Symbol]bool
	globalStack       map[string]bool
	problems          []*ast.Node
}

// NoDocumentCookieRule disallows assigning to document.cookie directly.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-document-cookie.js
var NoDocumentCookieRule = rule.Rule{
	Name:   "unicorn/no-document-cookie",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !sourceMayReferenceDocument(ctx.SourceFile) {
			return nil
		}

		tracker := &documentReferenceTracker{
			ctx:               ctx,
			identifiersByName: make(map[string][]*ast.Node),
		}

		return rule.RuleListeners{
			ast.KindIdentifier: tracker.collectIdentifier,
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				tracker.finish()
			},
		}
	},
}

func sourceMayReferenceDocument(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.Identifiers == nil {
		return true
	}
	if _, ok := sourceFile.Identifiers["document"]; ok {
		return true
	}
	for _, name := range globalObjectNames {
		if _, ok := sourceFile.Identifiers[name]; ok {
			return true
		}
	}
	return false
}

func (tracker *documentReferenceTracker) collectIdentifier(node *ast.Node) {
	if utils.IsNonReferenceIdentifier(node) {
		return
	}
	name := node.AsIdentifier().Text
	tracker.identifiersByName[name] = append(tracker.identifiersByName[name], node)
}

func (tracker *documentReferenceTracker) finish() {
	tracker.trackGlobalRoot("document", tracedDocument)
	for _, name := range globalObjectNames {
		tracker.trackGlobalRoot(name, tracedGlobalObject)
	}

	message := rule.RuleMessage{
		Id:          messageID,
		Description: "Do not use `document.cookie` directly.",
	}
	for _, node := range tracker.problems {
		tracker.ctx.ReportNode(node, message)
	}
}

func (tracker *documentReferenceTracker) trackGlobalRoot(name string, value tracedValue) {
	if !tracker.ctx.Globals.Access(name).IsDeclared() {
		return
	}
	references := tracker.globalReferences(name)
	for _, reference := range references {
		if utils.IsWriteReference(reference) {
			return
		}
	}
	for _, reference := range references {
		tracker.trackExpression(reference, value)
	}
}

func (tracker *documentReferenceTracker) globalReferences(name string) []*ast.Node {
	var references []*ast.Node
	for _, identifier := range tracker.identifiersByName[name] {
		if tracker.isGlobalReference(identifier, name) {
			references = append(references, identifier)
		}
	}
	return references
}

func (tracker *documentReferenceTracker) isGlobalReference(identifier *ast.Node, name string) bool {
	if tracker.ctx.Refs != nil {
		return tracker.ctx.Refs.IsGlobalReference(identifier)
	}
	return !utils.IsShadowed(identifier, name)
}

func (tracker *documentReferenceTracker) trackExpression(node *ast.Node, value tracedValue) {
	if node == nil {
		return
	}
	for node.Parent != nil && documentValuePassesThrough(node, node.Parent) {
		node = node.Parent
	}

	parent := node.Parent
	if parent == nil {
		return
	}
	if ast.IsAccessExpression(parent) && utils.AccessExpressionObject(parent) == node {
		name, ok := tracker.accessExpressionStaticName(parent)
		if !ok {
			return
		}
		switch value {
		case tracedGlobalObject:
			if name == "document" {
				tracker.trackExpression(parent, tracedDocument)
			}
		case tracedDocument:
			if name == "cookie" && isDirectAssignmentTarget(parent) {
				tracker.problems = append(tracker.problems, parent)
			}
		}
		return
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.Right == node && binary.OperatorToken != nil &&
			ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			tracker.trackAssignmentTarget(binary.Left, value)
			tracker.trackExpression(parent, value)
		}
	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		if declaration != nil && declaration.Initializer == node {
			tracker.trackAssignmentTarget(declaration.Name(), value)
		}
	case ast.KindParameter:
		parameter := parent.AsParameterDeclaration()
		if parameter != nil && parameter.Initializer == node {
			tracker.trackAssignmentTarget(parameter.Name(), value)
		}
	case ast.KindBindingElement:
		element := parent.AsBindingElement()
		if element != nil && element.Initializer == node {
			tracker.trackAssignmentTarget(element.Name(), value)
		}
	case ast.KindShorthandPropertyAssignment:
		property := parent.AsShorthandPropertyAssignment()
		if property != nil && property.ObjectAssignmentInitializer == node {
			tracker.trackAssignmentTarget(property.Name(), value)
		}
	}
}

func (tracker *documentReferenceTracker) trackAssignmentTarget(node *ast.Node, value tracedValue) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindIdentifier:
		tracker.trackIdentifier(node, value)
	case ast.KindObjectBindingPattern:
		tracker.trackObjectBindingPattern(node, value)
	case ast.KindObjectLiteralExpression:
		tracker.trackObjectAssignmentPattern(node, value)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			tracker.trackAssignmentTarget(binary.Left, value)
		}
	}
}

func (tracker *documentReferenceTracker) trackObjectBindingPattern(node *ast.Node, value tracedValue) {
	if value != tracedGlobalObject {
		return
	}
	pattern := node.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil {
		return
	}
	for _, elementNode := range pattern.Elements.Nodes {
		element := elementNode.AsBindingElement()
		if element == nil || element.DotDotDotToken != nil || element.Name() == nil {
			continue
		}
		propertyName := element.PropertyName
		if propertyName == nil {
			propertyName = element.Name()
		}
		if name, ok := tracker.staticPropertyName(propertyName); ok && name == "document" {
			tracker.trackAssignmentTarget(element.Name(), tracedDocument)
		}
	}
}

func (tracker *documentReferenceTracker) trackObjectAssignmentPattern(node *ast.Node, value tracedValue) {
	if value != tracedGlobalObject {
		return
	}
	object := node.AsObjectLiteralExpression()
	if object == nil || object.Properties == nil {
		return
	}
	for _, propertyNode := range object.Properties.Nodes {
		switch propertyNode.Kind {
		case ast.KindPropertyAssignment:
			property := propertyNode.AsPropertyAssignment()
			if name, ok := tracker.staticPropertyName(property.Name()); ok && name == "document" {
				tracker.trackAssignmentTarget(property.Initializer, tracedDocument)
			}
		case ast.KindShorthandPropertyAssignment:
			property := propertyNode.AsShorthandPropertyAssignment()
			if name, ok := tracker.staticPropertyName(property.Name()); ok && name == "document" {
				tracker.trackAssignmentTarget(property.Name(), tracedDocument)
			}
		}
	}
}

func (tracker *documentReferenceTracker) trackIdentifier(identifier *ast.Node, value tracedValue) {
	if tracker.ctx.Refs != nil {
		if symbol := tracker.ctx.Refs.Resolve(identifier); utils.IsValueSymbolDeclaredInFile(symbol, tracker.ctx.SourceFile) {
			tracker.trackVariable(symbol, value)
			return
		}
	}
	if symbol := documentBindingSymbol(identifier); symbol != nil {
		tracker.trackVariable(symbol, value)
		return
	}
	name := identifier.AsIdentifier().Text
	if tracker.ctx.Globals.Access(name).IsDeclared() && tracker.isGlobalReference(identifier, name) {
		tracker.trackGlobalVariable(name, value)
	}
}

func documentBindingSymbol(identifier *ast.Node) *ast.Symbol {
	if identifier == nil || identifier.Kind != ast.KindIdentifier ||
		!utils.IsDeclarationIdentifier(identifier) {
		return nil
	}
	return identifier.Parent.Symbol()
}

func (tracker *documentReferenceTracker) trackVariable(symbol *ast.Symbol, value tracedValue) {
	if tracker.ctx.Refs == nil || symbol == nil {
		return
	}
	if tracker.variableStack != nil && tracker.variableStack[symbol] {
		return
	}
	if tracker.variableStack == nil {
		tracker.variableStack = make(map[*ast.Symbol]bool)
	}
	tracker.variableStack[symbol] = true
	defer delete(tracker.variableStack, symbol)

	for _, reference := range tracker.ctx.Refs.References(symbol) {
		if !ast.IsWriteOnlyAccess(reference) {
			tracker.trackExpression(reference, value)
		}
	}
}

func (tracker *documentReferenceTracker) trackGlobalVariable(name string, value tracedValue) {
	if tracker.globalStack != nil && tracker.globalStack[name] {
		return
	}
	if tracker.globalStack == nil {
		tracker.globalStack = make(map[string]bool)
	}
	tracker.globalStack[name] = true
	defer delete(tracker.globalStack, name)

	for _, reference := range tracker.identifiersByName[name] {
		if !ast.IsWriteOnlyAccess(reference) && tracker.isGlobalReference(reference, name) {
			tracker.trackExpression(reference, value)
		}
	}
}

func (tracker *documentReferenceTracker) accessExpressionStaticName(node *ast.Node) (string, bool) {
	if node.Kind == ast.KindElementAccessExpression {
		argument := node.AsElementAccessExpression().ArgumentExpression
		if tracker.propertyEvaluator == nil {
			tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
		}
		return tracker.propertyEvaluator.EvalToString(argument)
	}
	return utils.AccessExpressionStaticName(node)
}

func (tracker *documentReferenceTracker) staticPropertyName(node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindComputedPropertyName {
		if tracker.propertyEvaluator == nil {
			tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
		}
		return tracker.propertyEvaluator.EvalToString(node.AsComputedPropertyName().Expression)
	}
	return utils.GetStaticPropertyName(node)
}

func documentValuePassesThrough(node *ast.Node, parent *ast.Node) bool {
	if ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments) {
		return parent.Expression() == node
	}
	if parent.Kind == ast.KindConditionalExpression {
		conditional := parent.AsConditionalExpression()
		return conditional.WhenTrue == node || conditional.WhenFalse == node
	}
	if parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return false
	}
	switch binary.OperatorToken.Kind {
	case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
		return binary.Left == node || binary.Right == node
	case ast.KindCommaToken:
		return binary.Right == node
	default:
		return false
	}
}

func isDirectAssignmentTarget(node *ast.Node) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	return binary != nil && binary.Left == current && binary.OperatorToken != nil &&
		ast.IsAssignmentOperator(binary.OperatorToken.Kind)
}
