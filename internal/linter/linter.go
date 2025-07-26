package linter

import (
	"context"

	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
)

// ConfiguredRule represents a rule with its configuration level
// Note: This is duplicated from config package to avoid circular import
type ConfiguredRule struct {
	Name  string
	Level rule.DiagnosticLevel
	Run   func(ctx rule.RuleContext, options any) rule.RuleListeners
}

func RunLinter(programs []*compiler.Program, singleThreaded bool, files []*ast.SourceFile, getRulesForFile func(sourceFile *ast.SourceFile) []ConfiguredRule, onDiagnostic func(diagnostic rule.RuleDiagnostic)) error {

	queue := make(chan *ast.SourceFile, len(files))
	for _, file := range files {
		queue <- file
	}
	close(queue)

	wg := core.NewWorkGroup(singleThreaded)
	for _, program := range programs {
		checkers, done := program.GetTypeCheckers(context.Background())
		defer done()
		for _, checker := range checkers {
			wg.Queue(func() {
				registeredListeners := make(map[ast.Kind][](func(node *ast.Node)), 20)

				for file := range queue {
					rules := getRulesForFile(file)
					for _, r := range rules {
						// Skip rules that are turned off
						if r.Level == rule.DiagnosticLevelOff {
							continue
						}

						ctx := rule.RuleContext{
							SourceFile:  file,
							Program:     program,
							TypeChecker: checker,
							ReportRange: func(textRange core.TextRange, msg rule.RuleMessage) {
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:   r.Name,
									Range:      textRange,
									Message:    msg,
									Level:      r.Level,
									SourceFile: file,
								})
							},
							ReportRangeWithSuggestions: func(textRange core.TextRange, msg rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:    r.Name,
									Range:       textRange,
									Message:     msg,
									Level:       r.Level,
									Suggestions: &suggestions,
									SourceFile:  file,
								})
							},
							ReportNode: func(node *ast.Node, msg rule.RuleMessage) {
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:   r.Name,
									Range:      utils.TrimNodeTextRange(file, node),
									Message:    msg,
									Level:      r.Level,
									SourceFile: file,
								})
							},
							ReportNodeWithFixes: func(node *ast.Node, msg rule.RuleMessage, fixes ...rule.RuleFix) {
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:   r.Name,
									Range:      utils.TrimNodeTextRange(file, node),
									Message:    msg,
									Level:      r.Level,
									FixesPtr:   &fixes,
									SourceFile: file,
								})
							},

							ReportNodeWithSuggestions: func(node *ast.Node, msg rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:    r.Name,
									Range:       utils.TrimNodeTextRange(file, node),
									Message:     msg,
									Level:       r.Level,
									Suggestions: &suggestions,
									SourceFile:  file,
								})
							},
						}

						for kind, listener := range r.Run(ctx, nil) { // Pass nil for options for now
							listeners, ok := registeredListeners[kind]
							if !ok {
								listeners = make([](func(node *ast.Node)), 0, len(rules))
							}
							registeredListeners[kind] = append(listeners, listener)
						}
					}

					runListeners := func(kind ast.Kind, node *ast.Node) {
						if listeners, ok := registeredListeners[kind]; ok {
							for _, listener := range listeners {
								listener(node)
							}
						}
					}

					/* convert.ts -> allowPattern:
					catch name
					variabledeclaration name
					forinstatement initializer
					forofstatement initializer
					(propagation) allowPattern > arrayliteralexpression elements
					(propagation) allowPattern > objectliteralexpression properties
					(propagation) allowPattern > spreadassignment,spreadelement expression
					(propagation) allowPattern > propertyassignment value
					arraybindingpattern elements
					objectbindingpattern elements
					(init) binaryexpression(with '=' operator') left
					*/

					var childVisitor ast.Visitor
					var patternVisitor func(node *ast.Node)
					patternVisitor = func(node *ast.Node) {
						runListeners(node.Kind, node)
						kind := rule.ListenerOnAllowPattern(node.Kind)
						runListeners(kind, node)

						switch node.Kind {
						case ast.KindArrayLiteralExpression:
							for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
								patternVisitor(element)
							}
						case ast.KindObjectLiteralExpression:
							for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
								patternVisitor(property)
							}
						case ast.KindSpreadElement, ast.KindSpreadAssignment:
							patternVisitor(node.Expression())
						case ast.KindPropertyAssignment:
							patternVisitor(node.Initializer())
						default:
							node.ForEachChild(childVisitor)
						}

						runListeners(rule.ListenerOnExit(kind), node)
						runListeners(rule.ListenerOnExit(node.Kind), node)
					}
					childVisitor = func(node *ast.Node) bool {
						runListeners(node.Kind, node)

						switch node.Kind {
						case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
							kind := rule.ListenerOnNotAllowPattern(node.Kind)
							runListeners(kind, node)
							node.ForEachChild(childVisitor)
							runListeners(rule.ListenerOnExit(kind), node)
						default:
							if ast.IsAssignmentExpression(node, true) {
								expr := node.AsBinaryExpression()
								patternVisitor(expr.Left)
								childVisitor(expr.OperatorToken)
								childVisitor(expr.Right)
							} else {
								node.ForEachChild(childVisitor)
							}
						}

						runListeners(rule.ListenerOnExit(node.Kind), node)

						return false
					}
					file.Node.ForEachChild(childVisitor)
					clear(registeredListeners)
				}
			})
		}

	}
	wg.RunAndWait()
	return nil
}
