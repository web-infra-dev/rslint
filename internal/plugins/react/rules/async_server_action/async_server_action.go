package async_server_action

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var AsyncServerActionRule = rule.Rule{
	Name:   "react/async-server-action",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBlock: func(node *ast.Node) {
				function := node.Parent
				if !ast.IsFunctionLikeDeclaration(function) {
					return
				}
				if ast.GetFunctionFlags(function)&(ast.FunctionFlagsAsync|ast.FunctionFlagsGenerator) != 0 {
					return
				}
				statements := node.AsBlock().Statements.Nodes
				if len(statements) == 0 || statements[0].Kind != ast.KindExpressionStatement {
					return
				}
				// Upstream checks the first expression's decoded value, not its
				// directive flag; parentheses and escaped strings also match.
				expression := utils.ESTreeRuntimeExpression(statements[0].AsExpressionStatement().Expression)
				if expression.Kind != ast.KindStringLiteral || expression.Text() != "use server" {
					return
				}
				location := utils.ESTreeFunctionRange(ctx.SourceFile, function)
				ctx.ReportRangeWithDeferredSuggestions(location, rule.RuleMessage{
					Id:          "asyncServerAction",
					Description: "Server Actions must be async",
				}, func() []rule.RuleSuggestion {
					isMethod := function.Kind == ast.KindMethodDeclaration || function.Kind == ast.KindConstructor ||
						(ast.IsAccessor(function) &&
							function.Parent.Kind != ast.KindObjectLiteralExpression)
					key := function.Name()
					if isMethod && key != nil && key.Kind == ast.KindComputedPropertyName {
						key = utils.ESTreeRuntimeExpression(key.AsComputedPropertyName().Expression)
					}
					name := ""
					constructorStart := 0
					if function.Kind == ast.KindConstructor {
						start := function.Pos()
						if modifiers := function.Modifiers(); modifiers != nil {
							start = modifiers.End()
						}
						// tsgo also represents a quoted 'constructor' key as a
						// constructor, but only an identifier names the suggestion.
						token := scanner.GetScannerForSourceFile(ctx.SourceFile, start)
						constructorStart = token.TokenStart()
						if token.Token() != ast.KindStringLiteral {
							name = "constructor"
						}
					} else if key != nil && key.Kind == ast.KindIdentifier && (isMethod ||
						function.Kind == ast.KindFunctionDeclaration || function.Kind == ast.KindFunctionExpression) {
						name = key.Text()
					}
					var data map[string]string
					if name != "" {
						data = map[string]string{"functionName": "`" + name + "`"}
					}
					description := "Make this function `async`"
					if name != "" {
						description = "Make `" + name + "` an `async` function"
					}
					insertion := location.Pos()
					if function.Kind == ast.KindConstructor {
						insertion = constructorStart
					} else if isMethod {
						// Match upstream's key insertion, including its limitations
						// for computed methods and accessors documented by this rule.
						insertion = utils.TrimNodeTextRange(ctx.SourceFile, key).Pos()
					}
					// Upstream supplies a suggestion description without a message ID.
					return []rule.RuleSuggestion{{
						Message: rule.RuleMessage{Description: description, Data: data},
						FixesArr: []rule.RuleFix{{
							Range: location.WithPos(insertion).WithEnd(insertion),
							Text:  "async ",
						}},
					}}
				})
			},
		}
	},
}
