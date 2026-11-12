package unicornutil

import (
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func booleanValueAnnotation(ctx rule.RuleContext, declaration *ast.Node) (*ast.Node, bool) {
	annotation, optional := ValueTypeAnnotation(ctx, declaration)
	if annotation != nil && !utils.IsJSDocSyntaxNode(annotation) {
		return annotation, optional
	}
	// A native initializer assertion takes precedence over the binding's JSDoc.
	if declaration.Kind == ast.KindVariableDeclaration {
		for value := utils.ESTreeRuntimeExpression(declaration.Initializer()); value != nil; value = utils.ESTreeRuntimeExpression(value.Expression()) {
			switch value.Kind {
			case ast.KindAsExpression, ast.KindTypeAssertionExpression:
				return nil, optional
			case ast.KindSatisfiesExpression, ast.KindNonNullExpression:
				continue
			}
			break
		}
	}
	// tsgo leaves Closure-style function types unparsed, including the parameter
	// name after them. Adapt only that spelling; let its parser read the result.
	if annotation == nil || annotation.Kind == ast.KindTypeReference && ast.IsIdentifier(annotation.AsTypeReferenceNode().TypeName) && annotation.AsTypeReferenceNode().TypeName.Text() == "function" {
		if adapted, optional := closureCallableAnnotation(ctx, declaration); adapted != nil {
			return adapted, optional
		}
	}
	return annotation, optional
}

func closureCallableAnnotation(ctx rule.RuleContext, declaration *ast.Node) (*ast.Node, bool) {
	host := declaration
	if declaration.Kind == ast.KindParameter {
		host = declaration.Parent
	}
	for ; host != nil; host = ast.GetNextJSDocCommentLocation(host) {
		for _, doc := range host.JSDoc(ctx.SourceFile) {
			if doc.AsJSDoc().Tags == nil {
				continue
			}
			for _, tag := range doc.AsJSDoc().Tags.Nodes {
				var expression *ast.Node
				switch {
				case declaration.Kind == ast.KindParameter && ast.IsIdentifier(declaration.Name()) && tag.Kind == ast.KindJSDocParameterTag:
					expression = tag.AsJSDocParameterOrPropertyTag().TypeExpression
				case declaration.Kind == ast.KindVariableDeclaration && tag.Kind == ast.KindJSDocTypeTag:
					expression = tag.AsJSDocTypeTag().TypeExpression
				}
				if expression == nil {
					continue
				}
				text := ctx.SourceFile.Text()[expression.Pos():tag.End()]
				tokens := scanner.NewScanner()
				tokens.SetText(text)
				tokens.SetSkipJSDocLeadingAsterisks(true)
				if tokens.Scan() != ast.KindOpenBraceToken || tokens.Scan() != ast.KindFunctionKeyword || tokens.Scan() != ast.KindOpenParenToken {
					continue
				}
				depth := 1
				for depth > 0 {
					switch tokens.Scan() {
					case ast.KindOpenParenToken:
						depth++
					case ast.KindCloseParenToken:
						depth--
					case ast.KindEndOfFile:
						return nil, false
					}
				}
				if tokens.Scan() != ast.KindColonToken {
					continue
				}
				start := tokens.TokenEnd()
				for depth = 1; depth > 0; {
					switch tokens.Scan() {
					case ast.KindOpenBraceToken:
						depth++
					case ast.KindCloseBraceToken:
						depth--
					case ast.KindEndOfFile:
						return nil, false
					}
				}
				result := text[start:tokens.TokenStart()]
				optional := ast.HasQuestionToken(declaration)
				if declaration.Kind == ast.KindParameter {
					kind := tokens.Scan()
					bracketed := kind == ast.KindOpenBracketToken
					if bracketed {
						optional = true
						tokens.Scan()
					}
					if tokens.TokenValue() != declaration.Name().Text() {
						continue
					}
					if bracketed {
						if next := tokens.Scan(); next != ast.KindCloseBracketToken && next != ast.KindEqualsToken {
							continue
						}
					} else if end := tokens.TokenEnd(); end < len(text) {
						next, _ := utf8.DecodeRuneInString(text[end:])
						if !ecmascript.IsWhiteSpaceOrLineTerminator(next) {
							continue
						}
					}
				}
				file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/annotation.ts", Path: "/annotation.ts"}, "/** @type {() => "+result+"} */ let value;", core.ScriptKindTS)
				value := file.Statements.Nodes[0].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0]
				adapted, _ := ValueTypeAnnotation(rule.RuleContext{SourceFile: file}, value)
				if adapted != nil && adapted.Kind == ast.KindFunctionType {
					// Resolve any names in the annotation's original lexical scope.
					adapted.Parent = declaration
					return adapted, optional
				}
			}
		}
	}
	return nil, false
}
