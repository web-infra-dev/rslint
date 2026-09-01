package unified_signatures

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed unified_signatures.schema.json
var schemaJSON []byte

type unifiedSignaturesOptions struct {
	ignoreDifferentlyNamedParameters  bool
	ignoreOverloadsWithDifferentJSDoc bool
}

type unifyKind uint8

const (
	unifySingleParameterDifference unifyKind = iota
	unifyExtraParameter
	unifyAllParametersAreSame
)

type unifyResult struct {
	kind                   unifyKind
	parameter0, parameter1 *ast.Node
	extraParameter         *ast.Node
	signature0, signature1 *ast.Node
	otherSignature         *ast.Node
}

type overloadAnalyzer struct {
	ctx      rule.RuleContext
	options  unifiedSignaturesOptions
	comments []*ast.CommentRange
	tokens   *scanner.Scanner
}

var UnifiedSignaturesRule = rule.CreateRule(rule.Rule{
	Name:   "unified-signatures",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		a := &overloadAnalyzer{ctx: ctx, options: parseOptions(options)}
		if a.options.ignoreOverloadsWithDifferentJSDoc && ctx.Comments != nil {
			a.comments = ctx.Comments.All()
		}
		a.checkScope(&ctx.SourceFile.Node)
		checkScope := func(node *ast.Node) { a.checkScope(node) }
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindClassDeclaration):     checkScope,
			rule.ListenerOnExit(ast.KindInterfaceDeclaration): checkScope,
			rule.ListenerOnExit(ast.KindModuleBlock):          checkScope,
			rule.ListenerOnExit(ast.KindTypeLiteral):          checkScope,
		}
	},
})

func parseOptions(options []any) unifiedSignaturesOptions {
	result := unifiedSignaturesOptions{}
	if len(options) == 0 {
		return result
	}
	option, _ := options[0].(map[string]any)
	result.ignoreDifferentlyNamedParameters, _ = option["ignoreDifferentlyNamedParameters"].(bool)
	result.ignoreOverloadsWithDifferentJSDoc, _ = option["ignoreOverloadsWithDifferentJSDoc"].(bool)
	return result
}

func (a *overloadAnalyzer) checkScope(scope *ast.Node) {
	overloadIndexes := make(map[string]int)
	overloads := make([][]*ast.Node, 0)
	for _, candidate := range scopeChildren(scope) {
		if !isOverloadSignature(candidate) {
			continue
		}
		if key, ok := a.overloadKey(candidate); ok {
			if index, exists := overloadIndexes[key]; exists {
				overloads[index] = append(overloads[index], candidate)
			} else {
				overloadIndexes[key] = len(overloads)
				overloads = append(overloads, []*ast.Node{candidate})
			}
		}
	}
	var outerTypeParameters []*ast.Node
	if scope.Kind == ast.KindClassDeclaration || scope.Kind == ast.KindInterfaceDeclaration {
		outerTypeParameters = scope.TypeParameters()
	}
	for _, group := range overloads {
		for i := range group {
			for j := i + 1; j < len(group); j++ {
				if result, ok := a.compareSignatures(group[i], group[j], outerTypeParameters); ok {
					a.report(result, len(group) == 2)
				}
			}
		}
	}
}

func scopeChildren(scope *ast.Node) []*ast.Node {
	switch scope.Kind {
	case ast.KindSourceFile, ast.KindModuleBlock:
		return scope.Statements()
	case ast.KindClassDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeLiteral:
		return scope.Members()
	default:
		return nil
	}
}

func isOverloadSignature(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindConstructor:
		return node.Body() == nil
	case ast.KindCallSignature, ast.KindConstructSignature, ast.KindMethodSignature:
		return true
	default:
		return false
	}
}

func (a *overloadAnalyzer) overloadKey(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		if name := node.Name(); name != nil {
			return name.Text(), true
		}
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault) {
			// Upstream falls back to the raw ESTree exporting-node type.
			return "ExportDefaultDeclaration", true
		}
		return "", false
	case ast.KindConstructor:
		// tsgo also uses KindConstructor for quoted methods such as
		// `"constructor"()`. ESTree keeps those as ordinary methods.
		return a.constructorOverloadKey(node)
	case ast.KindConstructSignature:
		return "11constructor", true
	case ast.KindCallSignature:
		return "11()", true
	case ast.KindMethodDeclaration, ast.KindMethodSignature:
		return a.methodOverloadKey(node)
	default:
		return "", false
	}
}

func (a *overloadAnalyzer) constructorOverloadKey(node *ast.Node) (string, bool) {
	parameters := node.ParameterList()
	if parameters == nil {
		return "", false
	}
	openParen := parameters.Pos() - 1
	if openParen < 0 {
		return "", false
	}
	if a.tokens == nil {
		a.tokens = scanner.NewScanner()
		a.tokens.SetText(a.ctx.SourceFile.Text())
		a.tokens.SetLanguageVariant(a.ctx.SourceFile.LanguageVariant)
	}
	a.tokens.ResetTokenState(utils.TrimNodeTextRange(a.ctx.SourceFile, node).Pos())
	var nameKind ast.Kind
	var nameText string
	for a.tokens.Scan(); a.tokens.Token() != ast.KindEndOfFile && a.tokens.TokenEnd() <= openParen; a.tokens.Scan() {
		nameKind = a.tokens.Token()
		nameText = a.tokens.TokenText()
	}

	if nameKind == ast.KindConstructorKeyword && !ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
		return "11constructor", true
	}
	prefix := "11"
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
		prefix = "10"
	}
	switch nameKind {
	case ast.KindConstructorKeyword, ast.KindIdentifier:
		return prefix + "identifier_constructor", true
	case ast.KindStringLiteral:
		return prefix + nameText, true
	default:
		return "", false
	}
}

func (a *overloadAnalyzer) methodOverloadKey(node *ast.Node) (string, bool) {
	name := node.Name()
	if name == nil {
		return "", false
	}
	prefix := "11"
	computed := name.Kind == ast.KindComputedPropertyName
	if computed {
		prefix = "01"
		name = ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
		prefix = prefix[:1] + "0"
	}
	switch name.Kind {
	case ast.KindPrivateIdentifier:
		return prefix + "private_identifier_" + name.Text(), true
	case ast.KindIdentifier:
		return prefix + "identifier_" + name.Text(), true
	default:
		if computed && !utils.IsESTreeLiteralKind(name.Kind) {
			// Upstream reads Literal.raw for every other computed key. Non-literal
			// expressions have no raw value and therefore share one overload group.
			return prefix + "undefined", true
		}
		return prefix + utils.TrimmedNodeText(a.ctx.SourceFile, name), true
	}
}

func (a *overloadAnalyzer) compareSignatures(left, right *ast.Node, outerTypeParameters []*ast.Node) (unifyResult, bool) {
	if !a.signaturesCanBeUnified(left, right, outerTypeParameters) {
		return unifyResult{}, false
	}
	if len(left.Parameters()) == len(right.Parameters()) {
		return a.signaturesHaveSameAmountOfParameters(left, right)
	}
	return signaturesDifferByOptionalOrRestParameter(a.ctx.SourceFile, left, right)
}

func (a *overloadAnalyzer) signaturesCanBeUnified(left, right *ast.Node, outerTypeParameters []*ast.Node) bool {
	leftParams, rightParams := left.Parameters(), right.Parameters()
	if a.options.ignoreDifferentlyNamedParameters {
		for i := range min(len(leftParams), len(rightParams)) {
			if parametersHaveSameESTreeShape(leftParams[i], rightParams[i]) && staticParameterName(leftParams[i]) != staticParameterName(rightParams[i]) {
				return false
			}
		}
	}
	if a.options.ignoreOverloadsWithDifferentJSDoc && a.blockCommentValue(left) != a.blockCommentValue(right) {
		return false
	}
	return typesAreEqual(a.ctx.SourceFile, left.Type(), right.Type()) &&
		typeParametersAreEqual(left.TypeParameters(), right.TypeParameters()) &&
		signatureUsesOuterTypeParameter(left, outerTypeParameters) == signatureUsesOuterTypeParameter(right, outerTypeParameters)
}

func (a *overloadAnalyzer) signaturesHaveSameAmountOfParameters(left, right *ast.Node) (unifyResult, bool) {
	leftParams, rightParams := left.Parameters(), right.Parameters()
	if isThisVoidParameter(firstParameter(leftParams)) || isThisVoidParameter(firstParameter(rightParams)) {
		return unifyResult{}, false
	}
	index := firstParameterDifference(a.ctx.SourceFile, leftParams, rightParams)
	if index < 0 {
		return unifyResult{kind: unifyAllParametersAreSame, signature0: left, signature1: right}, true
	}
	for i := index + 1; i < len(leftParams); i++ {
		if !parametersAreEqual(a.ctx.SourceFile, leftParams[i], rightParams[i]) {
			return unifyResult{}, false
		}
	}
	if !parametersHaveEqualSigils(leftParams[index], rightParams[index]) || isRestParameter(leftParams[index]) {
		return unifyResult{}, false
	}
	return unifyResult{
		kind:       unifySingleParameterDifference,
		parameter0: leftParams[index],
		parameter1: rightParams[index],
		signature0: left,
		signature1: right,
	}, true
}

func signaturesDifferByOptionalOrRestParameter(sourceFile *ast.SourceFile, left, right *ast.Node) (unifyResult, bool) {
	leftParams, rightParams := left.Parameters(), right.Parameters()
	shorter, longer, shorterSignature := leftParams, rightParams, left
	if len(leftParams) > len(rightParams) {
		shorter, longer, shorterSignature = rightParams, leftParams, right
	}
	if isThisParameter(firstParameter(leftParams)) != isThisParameter(firstParameter(rightParams)) ||
		isThisVoidParameter(firstParameter(leftParams)) || isThisVoidParameter(firstParameter(rightParams)) {
		return unifyResult{}, false
	}
	for i := len(shorter) + 1; i < len(longer); i++ {
		if !parameterMayBeMissing(longer[i]) {
			return unifyResult{}, false
		}
	}
	for i := range shorter {
		if !typesAreEqual(sourceFile, parameterTypeAnnotation(shorter[i]), parameterTypeAnnotation(longer[i])) {
			return unifyResult{}, false
		}
	}
	if len(shorter) > 0 && isRestParameter(shorter[len(shorter)-1]) {
		return unifyResult{}, false
	}
	return unifyResult{kind: unifyExtraParameter, extraParameter: longer[len(longer)-1], otherSignature: shorterSignature}, true
}

func (a *overloadAnalyzer) report(result unifyResult, onlyTwo bool) {
	otherLine := 0
	switch result.kind {
	case unifySingleParameterDifference:
		if !onlyTwo {
			otherLine = lineOfRange(a.ctx.SourceFile, utils.TrimNodeTextRange(a.ctx.SourceFile, result.parameter0))
		}
		a.ctx.ReportNode(result.parameter1, rule.RuleMessage{
			Id:          "singleParameterDifference",
			Description: fmt.Sprintf("%s taking `%s`.", failureStringStart(otherLine), a.unifiedTypeText(parameterTypeAnnotation(result.parameter0), parameterTypeAnnotation(result.parameter1))),
		})
	case unifyExtraParameter:
		if !onlyTwo {
			otherLine = lineOfRange(a.ctx.SourceFile, signatureReportRange(a.ctx.SourceFile, result.otherSignature))
		}
		messageID, suffix := "omittingSingleParameter", " with an optional parameter."
		if isRestParameter(result.extraParameter) {
			messageID, suffix = "omittingRestParameter", " with a rest parameter."
		}
		a.ctx.ReportNode(result.extraParameter, rule.RuleMessage{Id: messageID, Description: failureStringStart(otherLine) + suffix})
	case unifyAllParametersAreSame:
		if !onlyTwo {
			otherLine = lineOfRange(a.ctx.SourceFile, signatureReportRange(a.ctx.SourceFile, result.signature0))
		}
		a.ctx.ReportRange(signatureReportRange(a.ctx.SourceFile, result.signature1), rule.RuleMessage{Id: "allParametersAreSame", Description: failureStringStart(otherLine) + " with identical parameters."})
	}
}

func signatureReportRange(sourceFile *ast.SourceFile, signature *ast.Node) core.TextRange {
	trimmed := utils.TrimNodeTextRange(sourceFile, signature)
	start := trimmed.Pos()
	// ESTree reports the inner declaration, not its ExportNamedDeclaration
	// wrapper. tsgo includes `export` in the declaration range.
	for _, modifier := range signature.ModifierNodes() {
		if modifier.Kind != ast.KindExportKeyword {
			continue
		}
		start = scanner.SkipTrivia(sourceFile.Text(), modifier.End())
		if ast.HasSyntacticModifier(signature, ast.ModifierFlagsDefault) {
			start = scanner.SkipTrivia(sourceFile.Text(), start+len("default"))
		}
	}
	if (signature.Kind == ast.KindMethodDeclaration || signature.Kind == ast.KindConstructor) && signature.ParameterList() != nil {
		start = signature.ParameterList().Pos() - 1
	}
	// For generic class members, typescript-eslint reports the type-parameter
	// list rather than the member key or opening parenthesis.
	if signature.Kind == ast.KindMethodDeclaration && len(signature.TypeParameters()) > 0 {
		start = signature.TypeParameters()[0].Pos() - 1
	}
	return core.NewTextRange(start, trimmed.End())
}

func failureStringStart(otherLine int) string {
	if otherLine == 0 {
		return "These overloads can be combined into one signature"
	}
	return fmt.Sprintf("This overload and the one on line %d can be combined into one signature", otherLine)
}

func (a *overloadAnalyzer) unifiedTypeText(left, right *ast.Node) string {
	// A sole annotated type still passes through the same formatting as a
	// union member: transparent parentheses are removed, while function-like
	// types are wrapped to preserve their union spelling.
	if left == nil {
		return unionMemberText(a.ctx.SourceFile, right)
	}
	if right == nil {
		return unionMemberText(a.ctx.SourceFile, left)
	}
	members := append(unionMembers(left), unionMembers(right)...)
	unique := make([]*ast.Node, 0, len(members))
	for _, member := range members {
		duplicate := false
		for _, other := range unique {
			if typesAreEqual(a.ctx.SourceFile, member, other) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			unique = append(unique, member)
		}
	}
	var text strings.Builder
	for i, member := range unique {
		if i > 0 {
			text.WriteString(" | ")
		}
		text.WriteString(unionMemberText(a.ctx.SourceFile, member))
	}
	return text.String()
}

func unionMemberText(sourceFile *ast.SourceFile, member *ast.Node) string {
	member = skipParenthesizedType(member)
	text := utils.TrimmedNodeText(sourceFile, member)
	if member.Kind == ast.KindConditionalType || member.Kind == ast.KindConstructorType || member.Kind == ast.KindFunctionType {
		return "(" + text + ")"
	}
	return text
}

func unionMembers(node *ast.Node) []*ast.Node {
	node = skipParenthesizedType(node)
	if node == nil {
		return nil
	}
	if node.Kind != ast.KindUnionType {
		return []*ast.Node{node}
	}
	result := []*ast.Node{}
	for _, member := range node.AsUnionTypeNode().Types.Nodes {
		result = append(result, unionMembers(member)...)
	}
	return result
}

func (a *overloadAnalyzer) blockCommentValue(node *ast.Node) string {
	start, end := node.Pos(), scanner.SkipTrivia(a.ctx.SourceFile.Text(), node.Pos())
	var found *ast.CommentRange
	for _, comment := range a.comments {
		if comment.Pos() >= start && comment.End() <= end && comment.Kind == ast.KindMultiLineCommentTrivia {
			found = comment
		}
	}
	if found == nil {
		return ""
	}
	// Prefix distinguishes an absent comment from an explicitly empty block comment.
	return "\x00" + utils.CommentValue(a.ctx.SourceFile.Text(), found)
}

func firstParameter(parameters []*ast.Node) *ast.Node {
	if len(parameters) == 0 {
		return nil
	}
	return parameters[0]
}

func isThisParameter(parameter *ast.Node) bool {
	return parameter != nil && parameter.Name() != nil && parameter.Name().Kind == ast.KindIdentifier && parameter.Name().Text() == "this"
}

func isThisVoidParameter(parameter *ast.Node) bool {
	if !isThisParameter(parameter) {
		return false
	}
	typeNode := skipParenthesizedType(parameterTypeAnnotation(parameter))
	return typeNode != nil && typeNode.Kind == ast.KindVoidKeyword
}

func isRestParameter(parameter *ast.Node) bool {
	return parameter != nil && parameter.AsParameterDeclaration().DotDotDotToken != nil
}

func isOptionalParameter(parameter *ast.Node) bool {
	return parameter != nil && parameter.AsParameterDeclaration().QuestionToken != nil
}

func parameterMayBeMissing(parameter *ast.Node) bool {
	return isRestParameter(parameter) || isOptionalParameter(parameter)
}

func parametersHaveEqualSigils(left, right *ast.Node) bool {
	return isRestParameter(left) == isRestParameter(right) && isOptionalParameter(left) == isOptionalParameter(right)
}

func parametersHaveSameESTreeShape(left, right *ast.Node) bool {
	leftHasInitializer, rightHasInitializer := parameterHasInitializer(left), parameterHasInitializer(right)
	if leftHasInitializer || rightHasInitializer {
		return leftHasInitializer == rightHasInitializer
	}
	if isRestParameter(left) != isRestParameter(right) {
		return false
	}
	// ESTree represents both as RestElement regardless of the binding pattern
	// carried by its argument. The argument name comparison below decides
	// whether differently named rest parameters suppress the diagnostic.
	if isRestParameter(left) {
		return true
	}
	leftName, rightName := left.Name(), right.Name()
	return leftName != nil && rightName != nil && leftName.Kind == rightName.Kind
}

func parametersAreEqual(sourceFile *ast.SourceFile, left, right *ast.Node) bool {
	return parametersHaveEqualSigils(left, right) && typesAreEqual(sourceFile, parameterTypeAnnotation(left), parameterTypeAnnotation(right))
}

func parameterHasInitializer(parameter *ast.Node) bool {
	return parameter != nil && parameter.AsParameterDeclaration().Initializer != nil
}

func parameterTypeAnnotation(parameter *ast.Node) *ast.Node {
	if parameter == nil || parameterHasInitializer(parameter) {
		return nil
	}
	return parameter.Type()
}

func firstParameterDifference(sourceFile *ast.SourceFile, left, right []*ast.Node) int {
	for i := 0; i < len(left) && i < len(right); i++ {
		if !parametersAreEqual(sourceFile, left[i], right[i]) {
			return i
		}
	}
	return -1
}

func staticParameterName(parameter *ast.Node) string {
	if parameter == nil || parameterHasInitializer(parameter) || parameter.Name() == nil || parameter.Name().Kind != ast.KindIdentifier {
		return ""
	}
	return parameter.Name().Text()
}

func typesAreEqual(sourceFile *ast.SourceFile, left, right *ast.Node) bool {
	left, right = skipParenthesizedType(left), skipParenthesizedType(right)
	return left == right || left != nil && right != nil && utils.TrimmedNodeText(sourceFile, left) == utils.TrimmedNodeText(sourceFile, right)
}

func typeParametersAreEqual(left, right []*ast.Node) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		leftParam, rightParam := left[i].AsTypeParameterDeclaration(), right[i].AsTypeParameterDeclaration()
		if left[i].Name().Text() != right[i].Name().Text() || !constraintsAreEqual(leftParam.Constraint, rightParam.Constraint) {
			return false
		}
	}
	return true
}

func constraintsAreEqual(left, right *ast.Node) bool {
	left, right = skipParenthesizedType(left), skipParenthesizedType(right)
	return left == right || left != nil && right != nil && estreeConstraintKind(left) == estreeConstraintKind(right)
}

func estreeConstraintKind(node *ast.Node) ast.Kind {
	if node.Kind == ast.KindLiteralType {
		if literal := node.AsLiteralTypeNode().Literal; literal != nil && literal.Kind == ast.KindNullKeyword {
			return ast.KindNullKeyword
		}
	}
	return node.Kind
}

func skipParenthesizedType(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
	}
	return node
}

func signatureUsesOuterTypeParameter(signature *ast.Node, outerTypeParameters []*ast.Node) bool {
	if len(outerTypeParameters) == 0 {
		return false
	}
	names := make(map[string]struct{}, len(outerTypeParameters))
	for _, parameter := range outerTypeParameters {
		names[parameter.Name().Text()] = struct{}{}
	}
	for _, parameter := range signature.Parameters() {
		if typeContainsTypeParameter(parameterTypeAnnotation(parameter), names) {
			return true
		}
	}
	return false
}

func typeContainsTypeParameter(typeNode *ast.Node, names map[string]struct{}) bool {
	if typeNode == nil {
		return false
	}
	switch typeNode.Kind {
	case ast.KindTypeReference:
		typeName := typeNode.AsTypeReferenceNode().TypeName
		if typeName.Kind == ast.KindIdentifier {
			_, ok := names[typeName.Text()]
			return ok
		}
	case ast.KindArrayType:
		return typeContainsTypeParameter(typeNode.AsArrayTypeNode().ElementType, names)
	case ast.KindTypeOperator:
		return typeContainsTypeParameter(typeNode.AsTypeOperatorNode().Type, names)
	case ast.KindParenthesizedType:
		return typeContainsTypeParameter(typeNode.AsParenthesizedTypeNode().Type, names)
	}
	return false
}

func lineOfRange(sourceFile *ast.SourceFile, textRange core.TextRange) int {
	return scanner.ComputeLineOfPosition(sourceFile.ECMALineMap(), textRange.Pos()) + 1
}
