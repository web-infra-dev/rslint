package linter

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
)

type ConfiguredRule struct {
	Name     string
	Severity rule.DiagnosticSeverity
	Run      func(ctx rule.RuleContext) rule.RuleListeners
}

func RunLinterInProgram(program *compiler.Program, allowFiles []string, skipFiles []string, getRulesForFile RuleHandler, onDiagnostic DiagnosticHandler) int32 {
	checker, done := program.GetTypeChecker(context.Background())
	defer done()
	var lintedFileCount int32 = 0
	for _, file := range program.GetSourceFiles() {
		p := string(file.Path())
		// skip lint node_modules and bundled files
		// FIXME: we may have better api to tell whether a file is a bundled file or not
		skipFile := false
		for _, skipPattern := range skipFiles {
			if strings.Contains(p, skipPattern) {
				skipFile = true
				break
			}
		}
		if skipFile {
			continue
		}
		// only lint allowedFiles if allowedFiles is not empty
		if allowFiles != nil {
			found := false
			for _, filePath := range allowFiles {

				if filePath == file.FileName() {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		lintedFileCount++

		registeredListeners := make(map[ast.Kind][](func(node *ast.Node)), 20)
		{
			rules := getRulesForFile(file)
			if len(rules) == 0 {
				continue
			}

			comments := make([]*ast.CommentRange, 0)
			utils.ForEachComment(&file.Node, func(comment *ast.CommentRange) { comments = append(comments, comment) }, file)

			// Create disable manager for this file
			disableManager := rule.NewDisableManager(file, comments)

			for _, r := range rules {
				ctx := rule.RuleContext{
					SourceFile:     file,
					Program:        program,
					TypeChecker:    checker,
					DisableManager: disableManager,
					ReportRange: func(textRange core.TextRange, msg rule.RuleMessage) {
						// Check if rule is disabled at this position
						if disableManager.IsRuleDisabled(r.Name, textRange.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:   r.Name,
							Range:      textRange,
							Message:    msg,
							SourceFile: file,
							Severity:   r.Severity,
						})
					},
					ReportRangeWithSuggestions: func(textRange core.TextRange, msg rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
						// Check if rule is disabled at this position
						if disableManager.IsRuleDisabled(r.Name, textRange.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:    r.Name,
							Range:       textRange,
							Message:     msg,
							Suggestions: &suggestions,
							SourceFile:  file,
							Severity:    r.Severity,
						})
					},
					ReportRangeWithFixes: func(textRange core.TextRange, msg rule.RuleMessage, fixes ...rule.RuleFix) {
						// Check if rule is disabled at this position
						if disableManager.IsRuleDisabled(r.Name, textRange.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:   r.Name,
							Range:      textRange,
							Message:    msg,
							FixesPtr:   &fixes,
							SourceFile: file,
							Severity:   r.Severity,
						})
					},
					ReportNode: func(node *ast.Node, msg rule.RuleMessage) {
						// Check if rule is disabled at this position
						if disableManager.IsRuleDisabled(r.Name, node.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:   r.Name,
							Range:      utils.TrimNodeTextRange(file, node),
							Message:    msg,
							SourceFile: file,
							Severity:   r.Severity,
						})
					},
					ReportNodeWithFixes: func(node *ast.Node, msg rule.RuleMessage, fixes ...rule.RuleFix) {
						// Check if rule is disabled at this position
						if disableManager.IsRuleDisabled(r.Name, node.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:   r.Name,
							Range:      utils.TrimNodeTextRange(file, node),
							Message:    msg,
							FixesPtr:   &fixes,
							SourceFile: file,
							Severity:   r.Severity,
						})
					},

					ReportNodeWithSuggestions: func(node *ast.Node, msg rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
						// Check if rule is disabled at this position
						if disableManager.IsRuleDisabled(r.Name, node.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:    r.Name,
							Range:       utils.TrimNodeTextRange(file, node),
							Message:     msg,
							Suggestions: &suggestions,
							SourceFile:  file,
							Severity:    r.Severity,
						})
					},
				}

				for kind, listener := range r.Run(ctx) {
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

	}
	return lintedFileCount
}

type RuleHandler = func(sourceFile *ast.SourceFile) []ConfiguredRule
type DiagnosticHandler = func(diagnostic rule.RuleDiagnostic)

// when allowedFiles is passed as nil which means all files are allowed
// when allowedFiles is passed as slice, only files in the slice are allowed
func RunLinter(programs []*compiler.Program, singleThreaded bool, allowFiles []string, excludedPaths []string, getRulesForFile RuleHandler, onDiagnostic DiagnosticHandler) (int32, error) {

	wg := core.NewWorkGroup(singleThreaded)

	var lintedFileCount atomic.Int32
	for _, program := range programs {
		{
			wg.Queue(func() {
				fileCount := RunLinterInProgram(program, allowFiles, excludedPaths, getRulesForFile, onDiagnostic)
				lintedFileCount.Add(fileCount)
			})
		}

	}
	wg.RunAndWait()
	return lintedFileCount.Load(), nil
}
