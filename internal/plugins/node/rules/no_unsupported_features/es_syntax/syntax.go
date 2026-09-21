package es_syntax

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (c *syntaxChecker) visit(node *ast.Node) {
	source := c.ctx.SourceFile
	if node.Modifiers() != nil && ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
		c.report("modules", node)
	}
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		if node.Body() == nil {
			return
		}
		loc := utils.ESTreeFunctionRange(source, node)
		flags := ast.GetFunctionFlags(node)
		if flags&ast.FunctionFlagsAsync != 0 {
			c.reportRange("async-functions", node, loc)
		}
		if flags&ast.FunctionFlagsGenerator != 0 {
			c.reportRange("generators", node, loc)
		}
		if flags&ast.FunctionFlagsAsyncGenerator == ast.FunctionFlagsAsyncGenerator {
			c.reportRange("async-iteration", node, loc)
		}
		if node.Kind == ast.KindArrowFunction {
			c.reportRange("arrow-functions", node, loc)
		}
		params := utils.ESTreeParameters(node)
		if len(params) > 0 {
			c.trailingFunctionComma(node, params[len(params)-1])
		}
		if node.Kind == ast.KindFunctionDeclaration {
			if node.Parent.Kind == ast.KindBlock && !ast.IsFunctionLike(node.Parent.Parent) {
				c.reportRange("block-scoped-functions", node, loc)
			}
			if node.Parent.Kind == ast.KindIfStatement {
				c.reportRange("function-declarations-in-if-statement-clauses-without-block", node, loc)
			}
			if node.Parent.Kind == ast.KindLabeledStatement {
				c.report("labelled-function-declarations", node.Parent)
			}
		}
		if node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor {
			c.property(node)
			if node.Kind == ast.KindMethodDeclaration && node.Parent.Kind == ast.KindObjectLiteralExpression {
				c.report("property-shorthands", node)
			}
			if node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor {
				c.report("accessor-properties", node)
			}
			if node.Name().Kind == ast.KindPrivateIdentifier {
				c.report("class-fields", node.Name())
			}
		}
	case ast.KindVariableDeclarationList:
		if node.Flags&ast.NodeFlagsBlockScoped != 0 && !ast.IsVarUsing(node) && !ast.IsVarAwaitUsing(node) {
			target := node
			if node.Parent.Kind == ast.KindVariableStatement {
				target = node.Parent
			}
			loc := core.NewTextRange(scanner.GetTokenPosOfNode(node, source, false), target.End())
			c.reportRange("block-scoped-variables", node, loc)
		}
	case ast.KindVariableDeclaration:
		if init := node.Initializer(); init != nil && node.Parent.Kind == ast.KindVariableDeclarationList && node.Parent.Parent.Kind == ast.KindForInStatement {
			c.report("initializers-in-for-in", init)
		}
	case ast.KindParameter:
		p := node.AsParameterDeclaration()
		if p.Initializer != nil {
			c.report("default-parameters", node)
		}
		if p.DotDotDotToken != nil {
			c.report("rest-parameters", node)
		}
	case ast.KindBindingElement:
		c.property(node)
		if node.AsBindingElement().DotDotDotToken != nil && node.Parent.Kind == ast.KindObjectBindingPattern {
			c.report("rest-spread-properties", node)
		}
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		c.destructuring(node)
		c.trailingComma(node)
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		c.destructuring(node)
		c.trailingComma(node)
	case ast.KindClassDeclaration, ast.KindClassExpression:
		c.reportRange("classes", node, core.NewTextRange(utils.FindFunctionKeywordPos(source, node), node.End()))
	case ast.KindClassStaticBlockDeclaration:
		c.report("class-static-block", node)
	case ast.KindPropertyDeclaration:
		if utils.IsPlainClassMember(node) && !ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient) && !ast.HasSyntacticModifier(node.Parent, ast.ModifierFlagsAmbient) {
			c.report("class-fields", propertyKey(node))
		}
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		c.property(node)
	case ast.KindNumericLiteral, ast.KindBigIntLiteral:
		loc := utils.TrimNodeTextRange(source, node)
		text := source.Text()[loc.Pos():loc.End()]
		if node.Kind == ast.KindBigIntLiteral {
			c.report("bigint", node)
		} else if strings.HasPrefix(text, "0b") || strings.HasPrefix(text, "0B") {
			c.report("binary-numeric-literals", node)
		} else if strings.HasPrefix(text, "0o") || strings.HasPrefix(text, "0O") {
			c.report("octal-numeric-literals", node)
		}
		if strings.Contains(text, "_") {
			c.report("numeric-separators", node)
		}
	case ast.KindStringLiteral, ast.KindIdentifier:
		c.escapes(node)
		if node.Kind == ast.KindIdentifier {
			c.legacyAccessorIdentifier(node)
		}
	case ast.KindPrivateIdentifier:
		if node.Parent.Kind == ast.KindBinaryExpression && node.Parent.AsBinaryExpression().OperatorToken.Kind == ast.KindInKeyword && node.Parent.AsBinaryExpression().Left == node {
			c.report("class-fields", node)
			c.report("private-in", node)
			return
		}
		switch node.Parent.Kind {
		case ast.KindPropertyDeclaration, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			if utils.IsPlainClassMember(node.Parent) {
				return
			}
		}
		c.report("class-fields", node)
	case ast.KindRegularExpressionLiteral:
		text := node.Text()
		end := strings.LastIndex(text, "/")
		if end > 0 {
			c.regexp(node, text[1:end], text[end+1:])
		}
	case ast.KindForOfStatement:
		c.report("for-of-loops", node)
		if node.AsForInOrOfStatement().AwaitModifier != nil {
			c.report("async-iteration", node)
			c.topLevelAwait(node)
		}
	case ast.KindAwaitExpression:
		c.topLevelAwait(node)
	case ast.KindCatchClause:
		if node.AsCatchClause().VariableDeclaration == nil {
			c.report("optional-catch-binding", node)
		} else {
			c.shadowCatchParameter(node)
		}
	case ast.KindBinaryExpression:
		expr := node.AsBinaryExpression()
		switch expr.OperatorToken.Kind {
		case ast.KindAsteriskAsteriskToken, ast.KindAsteriskAsteriskEqualsToken:
			c.report("exponential-operators", node)
		case ast.KindQuestionQuestionToken:
			c.report("nullish-coalescing-operators", expr.OperatorToken)
		case ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken:
			c.report("logical-assignment-operators", expr.OperatorToken)
		}
	case ast.KindCallExpression, ast.KindNewExpression:
		args := node.Arguments()
		for _, argument := range args {
			if argument.Kind == ast.KindSpreadElement {
				c.report("spread-elements", argument)
			}
		}
		if node.Expression().Kind == ast.KindImportKeyword {
			c.report("dynamic-import", node)
		} else if len(args) > 0 {
			c.trailingFunctionComma(node, args[len(args)-1])
		}
		if node.Kind == ast.KindCallExpression && node.QuestionDotToken() != nil {
			c.report("optional-chaining", node.QuestionDotToken())
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if utils.IsInJsxTagName(node) {
			return
		}
		if token := node.QuestionDotToken(); token != nil {
			c.report("optional-chaining", token)
		}
		c.checkPrototype(node)
		if node.Kind == ast.KindPropertyAccessExpression && isKeyword(node.Name()) {
			c.report("keyword-properties", node)
		}
		if _, enabled := c.enabled["legacy-object-prototype-accessor-methods"]; enabled {
			if name, ok := utils.NewStaticStringEvaluatorWithoutScope().EvalAccessExpressionName(node); ok && isLegacyAccessor(name) {
				c.report("legacy-object-prototype-accessor-methods", propertyKey(node))
			}
		}
	case ast.KindSpreadAssignment:
		c.report("rest-spread-properties", node)
	case ast.KindMetaProperty:
		if node.AsMetaProperty().KeywordToken == ast.KindNewKeyword {
			c.report("new-target", node)
		} else {
			c.report("import-meta", node)
		}
	case ast.KindSuperKeyword:
		for parent := node.Parent; parent != nil; parent = parent.Parent {
			if ast.IsFunctionLikeDeclaration(parent) && parent.Kind != ast.KindArrowFunction {
				if parent.Kind == ast.KindMethodDeclaration && parent.Parent.Kind == ast.KindObjectLiteralExpression {
					c.report("object-super-properties", node)
				}
				break
			}
		}
	case ast.KindTaggedTemplateExpression:
		c.report("template-literals", node)
	case ast.KindTemplateExpression, ast.KindNoSubstitutionTemplateLiteral:
		if node.Parent.Kind != ast.KindTaggedTemplateExpression {
			c.report("template-literals", node)
		}
		malformed := false
		if node.Kind == ast.KindNoSubstitutionTemplateLiteral {
			malformed = node.TemplateLiteralLikeData().TemplateFlags&ast.TokenFlagsContainsInvalidEscape != 0
			c.escapes(node)
		} else {
			template := node.AsTemplateExpression()
			malformed = template.Head.TemplateLiteralLikeData().TemplateFlags&ast.TokenFlagsContainsInvalidEscape != 0
			for _, span := range template.TemplateSpans.Nodes {
				malformed = malformed || span.AsTemplateSpan().Literal.TemplateLiteralLikeData().TemplateFlags&ast.TokenFlagsContainsInvalidEscape != 0
			}
		}
		if malformed {
			c.report("malformed-template-literals", node)
		}
	case ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
		c.escapes(node)
	case ast.KindImportDeclaration, ast.KindExportDeclaration, ast.KindExportAssignment:
		c.report("modules", node)
		if node.Kind == ast.KindExportDeclaration {
			clause := node.AsExportDeclaration().ExportClause
			if clause != nil && clause.Kind == ast.KindNamespaceExport {
				c.report("export-ns-from", node)
			}
		}
		c.moduleNames(node)
	}
}

func (c *syntaxChecker) property(node *ast.Node) {
	name := node.Name()
	if node.Kind == ast.KindBindingElement {
		name = node.AsBindingElement().PropertyName
	}
	if name == nil {
		return
	}
	if name.Kind == ast.KindComputedPropertyName {
		c.report("computed-properties", node)
	} else if node.Parent.Kind != ast.KindClassDeclaration && node.Parent.Kind != ast.KindClassExpression && isKeyword(name) {
		c.report("keyword-properties", node)
	}
}
func propertyKey(node *ast.Node) *ast.Node {
	if node.Kind == ast.KindElementAccessExpression {
		return ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
	}
	name := node.Name()
	if name != nil && name.Kind == ast.KindComputedPropertyName {
		return ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	}
	return name
}
func (c *syntaxChecker) destructuring(node *ast.Node) {
	target := utils.OutermostParenthesizedExpression(node)
	parent := target.Parent
	if parent == nil {
		return
	}
	if parent.Kind == ast.KindVariableDeclaration && parent.Parent.Kind == ast.KindCatchClause {
		return
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindParameter:
		if parent.Name() == target {
			c.report("destructuring", node)
		}
	case ast.KindForOfStatement, ast.KindForInStatement:
		if parent.Initializer() == target {
			c.report("destructuring", node)
		}
	case ast.KindBinaryExpression:
		if parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && parent.AsBinaryExpression().Left == target {
			c.report("destructuring", node)
		}
	}
}
func (c *syntaxChecker) trailingFunctionComma(node, last *ast.Node) {
	if token, ok := utils.TokenAtOrAfter(c.ctx.SourceFile, last.End()); ok && token.Kind == ast.KindCommaToken {
		c.reportRange("trailing-function-commas", node, token.Range())
	}
}
func (c *syntaxChecker) trailingComma(node *ast.Node) {
	if token, ok := utils.TokenBeforePosition(c.ctx.SourceFile, node.End()-1); ok && token.Kind == ast.KindCommaToken {
		c.report("trailing-commas", node)
	}
}
func (c *syntaxChecker) topLevelAwait(node *ast.Node) {
	for child, parent := node, node.Parent; parent != nil; child, parent = parent, parent.Parent {
		if ast.IsFunctionLikeDeclaration(parent) && child != parent.Name() && child.Kind != ast.KindDecorator {
			return
		}
	}
	c.report("top-level-await", node)
}
func (c *syntaxChecker) escapes(node *ast.Node) {
	loc := utils.TrimNodeTextRange(c.ctx.SourceFile, node)
	text := c.ctx.SourceFile.Text()[loc.Pos():loc.End()]
	for i := 0; i < len(text); i++ {
		if text[i] == '\\' {
			if strings.HasPrefix(text[i:], `\u{`) {
				j := i + 3
				for j < len(text) && strings.ContainsRune("0123456789abcdefABCDEF", rune(text[j])) {
					j++
				}
				if j > i+3 && j < len(text) && text[j] == '}' {
					c.reportRange("unicode-codepoint-escapes", node, core.NewTextRange(loc.Pos()+i, loc.Pos()+j+1))
					i = j
					continue
				}
			}
			i++
		} else if node.Kind == ast.KindStringLiteral && (strings.HasPrefix(text[i:], "\u2028") || strings.HasPrefix(text[i:], "\u2029")) {
			// Upstream provides a point location. rslint represents it as an empty range.
			c.reportRange("json-superset", node, core.NewTextRange(loc.Pos()+i, loc.Pos()+i))
			i += 2
		}
	}
}
func (c *syntaxChecker) moduleNames(node *ast.Node) {
	// Only the ESTree import/export name roles accept arbitrary string names.
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindStringLiteral && (n.Parent.Kind == ast.KindImportSpecifier && n.Parent.AsImportSpecifier().PropertyName == n || n.Parent.Kind == ast.KindExportSpecifier || n.Parent.Kind == ast.KindNamespaceExport) {
			c.report("arbitrary-module-namespace-names", n)
		}
		return n.ForEachChild(visit)
	}
	node.ForEachChild(visit)
}
func isKeyword(node *ast.Node) bool {
	// Match es-x's ES3 reserved words, including byte/goto/native. The scanner's
	// current keyword set also contains TypeScript words that ES3 allowed here.
	return node != nil && node.Kind == ast.KindIdentifier && strings.Contains(" abstract boolean break byte case catch char class const continue debugger default delete do double else enum export extends false final finally float for function goto if implements import in instanceof int interface long native new null package private protected public return short static super switch synchronized this throw throws transient true try typeof var void volatile while with ", " "+node.Text()+" ")
}
