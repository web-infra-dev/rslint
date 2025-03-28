package linter

import (
	"fmt"
	"slices"
	"strings"

	"none.none/tsgolint/internal/rule"
	"none.none/tsgolint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

type ConfiguredRule struct {
	Name string
	Run func (ctx rule.RuleContext) rule.RuleListeners
}

func RunLinter(singleThreaded bool, fs vfs.FS, fileNames []string, getRulesForFile func (sourceFile *ast.SourceFile) []ConfiguredRule, cwd string, tsconfigPath string, onDiagnostic func(diagnostic rule.RuleDiagnostic)) error {
	program, err := utils.CreateProgram(singleThreaded, fs, cwd, tsconfigPath)

	if err != nil {
		return err
	}


	var files []*ast.SourceFile
	if len(fileNames) == 0 {
		files = utils.Filter(program.SourceFiles(), func(f *ast.SourceFile) bool {
			// TODO: case-aware comparison
			return strings.HasPrefix(string(f.Path()), cwd)
		})
	} else {
		files = make([]*ast.SourceFile, len(fileNames))
		for i, fileName := range fileNames {
			filePath := tspath.ToPath(fileName, cwd, program.Host().FS().UseCaseSensitiveFileNames())
			file := program.GetSourceFileByPath(filePath)
			if file == nil {
				return fmt.Errorf("unknown file %v", filePath)
			}
			files[i] = file
		}
	}

	queue := make(chan *ast.SourceFile, len(files))
	slices.SortFunc(files, func(a *ast.SourceFile, b *ast.SourceFile) int {
		return len(b.Text) - len(a.Text)
	})
	for _, file := range files  {
		queue <- file
	}
	close(queue)

	wg := core.NewWorkGroup(singleThreaded)
	for _, checker := range program.GetTypeCheckers() {
		wg.Queue(func() {
			registeredListeners := make(map[ast.Kind][](func (node *ast.Node)), 20)

			for file := range queue {
				rules := getRulesForFile(file)
				for _, r := range rules {
					ctx := rule.RuleContext{
						SourceFile: file,
						Program:     program,
						TypeChecker: checker,
						ReportRange: func(textRange core.TextRange, msg rule.RuleMessage) {
							onDiagnostic(rule.RuleDiagnostic{
								RuleName: r.Name,
								Range:      textRange,
								Message:    msg,
								SourceFile: file,
							})
						},
						ReportNode: func(node *ast.Node, msg rule.RuleMessage) {
							onDiagnostic(rule.RuleDiagnostic{
								RuleName: r.Name,
								Range:      utils.TrimNodeTextRange(file, node),
								Message:    msg,
								SourceFile: file,
							})
						},
						ReportNodeWithFixes: func(node *ast.Node, msg rule.RuleMessage, fixes ...rule.RuleFix) {
							onDiagnostic(rule.RuleDiagnostic{
								RuleName: r.Name,
								Range:      utils.TrimNodeTextRange(file, node),
								Message:    msg,
								FixesPtr:      &fixes,
								SourceFile: file,
							})
						},

						ReportNodeWithSuggestions: func(node *ast.Node, msg rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
							onDiagnostic(rule.RuleDiagnostic{
								RuleName: r.Name,
								Range:      utils.TrimNodeTextRange(file, node),
								Message:     msg,
								Suggestions: &suggestions,
								SourceFile:  file,
							})
						},
					}

					for kind, listener := range r.Run(ctx) {
						listeners, ok := registeredListeners[kind]
						if !ok {
							listeners = make([](func (node *ast.Node)), 0, len(rules))
						}
						registeredListeners[kind] = append(listeners, listener)
					}
				}

				utils.VisitAllChildrenWithExit(&file.Node, func(node *ast.Node) func() {
					listeners, ok := registeredListeners[node.Kind]
					if ok {
						for _, listener := range listeners {
							listener(node)
						}
					}
					listenersOnExit, ok := registeredListeners[rule.ListenerOnExit(node.Kind)]
					if ok {
						return func() {
							for _, listener := range listenersOnExit {
								listener(node)
							}
						}
					}
					return nil
				})
				clear(registeredListeners)
			}
		})
	}
	wg.RunAndWait()

	return nil
}
