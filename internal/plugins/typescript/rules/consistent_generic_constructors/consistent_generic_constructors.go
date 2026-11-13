package consistent_generic_constructors

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed consistent_generic_constructors.schema.json
var schemaJSON []byte

type ConsistentGenericConstructorsOptions struct {
	Style string `json:"style"`
}

func parseOptions(options []any) ConsistentGenericConstructorsOptions {
	opts := ConsistentGenericConstructorsOptions{Style: "constructor"}
	if len(options) > 0 {
		if style, ok := options[0].(string); ok {
			opts.Style = style
		}
	}
	return opts
}

// ConsistentGenericConstructorsRule enforces consistent generic specifier style in constructor signatures.
var ConsistentGenericConstructorsRule = rule.CreateRule(rule.Rule{
	Name:   "consistent-generic-constructors",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
})

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := parseOptions(options)
	sourceFile := ctx.SourceFile
	text := sourceFile.Text()

	checkNode := func(node, name, typeAnnotation, initializer *ast.Node) {
		if initializer == nil {
			return
		}
		initializer = ast.SkipParentheses(initializer)
		if initializer == nil || initializer.Kind != ast.KindNewExpression {
			return
		}
		newExpr := initializer.AsNewExpression()
		callee := ast.SkipParentheses(newExpr.Expression)
		if callee == nil || callee.Kind != ast.KindIdentifier {
			return
		}
		var typeRef *ast.TypeReferenceNode
		if typeAnnotation != nil {
			lhs := ast.SkipTypeParentheses(typeAnnotation)
			if lhs.Kind != ast.KindTypeReference {
				return
			}
			typeRef = lhs.AsTypeReferenceNode()
			if typeRef.TypeName.Kind != ast.KindIdentifier || typeRef.TypeName.Text() != callee.Text() {
				return
			}
		}

		// A parameter declaration can carry decorators outside the upstream
		// AssignmentPattern. Its diagnostic starts at the binding name.
		reportRange := utils.TrimNodeTextRange(sourceFile, node)
		if node.Kind == ast.KindParameter {
			reportRange = reportRange.WithPos(utils.TrimNodeTextRange(sourceFile, name).Pos())
		}
		if opts.Style == "type-annotation" {
			// An existing annotation (including a non-generic Foo) is preserved.
			if typeAnnotation != nil || newExpr.TypeArguments == nil {
				return
			}
			ctx.ReportRangeWithDeferredFixes(reportRange, rule.RuleMessage{
				Id:          "preferTypeAnnotation",
				Description: "The generic type arguments should be specified as part of the type annotation.",
			}, func() []rule.RuleFix {
				argsRange, ok := typeArgumentsRange(sourceFile, newExpr.TypeArguments)
				if !ok {
					return nil
				}
				// ComputedPropertyName includes the square brackets in tsgo.
				// Optional/definite property markers must precede the annotation.
				attachAfter := name
				if node.Kind == ast.KindPropertyDeclaration {
					if marker := node.PostfixToken(); marker != nil {
						attachAfter = marker
					}
				}
				return []rule.RuleFix{
					rule.RuleFixRemoveRange(argsRange),
					rule.RuleFixInsertAfter(attachAfter, ": "+utils.TrimmedNodeText(sourceFile, callee)+text[argsRange.Pos():argsRange.End()]),
				}
			})
			return
		}
		if ctx.LanguageOptions.IsolatedDeclarations || typeRef == nil || typeRef.TypeArguments == nil || newExpr.TypeArguments != nil {
			return
		}
		ctx.ReportRangeWithDeferredFixes(reportRange, rule.RuleMessage{
			Id:          "preferConstructor",
			Description: "The generic type arguments should be specified as part of the constructor type arguments.",
		}, func() []rule.RuleFix {
			argsRange, ok := typeArgumentsRange(sourceFile, typeRef.TypeArguments)
			if !ok {
				return nil
			}
			colon, ok := utils.TokenBeforePosition(sourceFile, utils.TrimNodeTextRange(sourceFile, typeAnnotation).Pos())
			if !ok || colon.Kind != ast.KindColonToken {
				return nil
			}
			annotationRange := core.NewTextRange(colon.Start, typeAnnotation.End())
			var moved strings.Builder
			for _, comment := range utils.CommentsInSpan(ctx.Comments.All(), annotationRange.Pos(), annotationRange.End()) {
				if comment.Pos() >= annotationRange.Pos() && comment.End() <= annotationRange.End() &&
					(comment.Pos() < argsRange.Pos() || comment.End() > argsRange.End()) {
					moved.WriteString(text[comment.Pos():comment.End()])
					if comment.Kind == ast.KindSingleLineCommentTrivia {
						terminator := ecmascript.LineTerminatorSequenceAt(text, comment.End())
						if terminator == "" {
							// Do not move an unterminated line comment into executable code.
							return nil
						}
						moved.WriteString(terminator)
					}
				}
			}
			moved.WriteString(text[argsRange.Pos():argsRange.End()])
			// Keep constructor grouping intact: inserting inside `(Foo)` would
			// turn construction into a call to Foo followed by another construction.
			calleeEnd := newExpr.Expression
			next, ok := utils.TokenAtOrAfter(sourceFile, calleeEnd.End())
			if !ok || next.Kind != ast.KindOpenParenToken {
				moved.WriteString("()")
			}
			return []rule.RuleFix{
				rule.RuleFixRemoveRange(annotationRange),
				rule.RuleFixInsertAfter(calleeEnd, moved.String()),
			}
		})
	}

	return rule.RuleListeners{
		ast.KindVariableDeclaration: func(node *ast.Node) {
			decl := node.AsVariableDeclaration()
			checkNode(node, decl.Name(), decl.Type, decl.Initializer)
		},
		ast.KindPropertyDeclaration: func(node *ast.Node) {
			decl := node.AsPropertyDeclaration()
			checkNode(node, decl.Name(), decl.Type, decl.Initializer)
		},
		ast.KindParameter: func(node *ast.Node) {
			// Upstream selects only direct parameters of function declarations
			// and expressions (including methods/constructors), not arrows or
			// parameter properties or default values nested in destructuring patterns.
			if node.Parent == nil || node.Parent.Kind == ast.KindArrowFunction || ast.IsParameterPropertyDeclaration(node, node.Parent) {
				return
			}
			decl := node.AsParameterDeclaration()
			checkNode(node, decl.Name(), decl.Type, decl.Initializer)
		},
	}
}

// The list range excludes its delimiters and may end before trailing trivia.
// Use source tokens so comments containing angle brackets cannot truncate it.
func typeArgumentsRange(sourceFile *ast.SourceFile, args *ast.NodeList) (core.TextRange, bool) {
	if args == nil || len(args.Nodes) == 0 {
		return core.TextRange{}, false
	}
	opening, ok := utils.TokenBeforePosition(sourceFile, args.Pos())
	if !ok || opening.Kind != ast.KindLessThanToken {
		return core.TextRange{}, false
	}
	closing, ok := utils.TokenAtOrAfter(sourceFile, args.End())
	if !ok || closing.Kind != ast.KindGreaterThanToken || closing.Start < opening.End {
		return core.TextRange{}, false
	}
	return core.NewTextRange(opening.Start, closing.End), true
}
