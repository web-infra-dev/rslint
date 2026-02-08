package init_declarations

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type Options struct {
	Mode              string
	IgnoreForLoopInit bool
}

func parseOptions(options any) Options {
	opts := Options{
		Mode:              "always",
		IgnoreForLoopInit: false,
	}

	if options == nil {
		return opts
	}

	switch optionValue := options.(type) {
	case []interface{}:
		if len(optionValue) > 0 {
			if mode, ok := optionValue[0].(string); ok && mode != "" {
				opts.Mode = mode
			}
		}
		if len(optionValue) > 1 {
			if optsMap, ok := optionValue[1].(map[string]interface{}); ok {
				if ignore, ok := optsMap["ignoreForLoopInit"].(bool); ok {
					opts.IgnoreForLoopInit = ignore
				}
			}
		}
	case map[string]interface{}:
		if mode, ok := optionValue["mode"].(string); ok && mode != "" {
			opts.Mode = mode
		}
		if ignore, ok := optionValue["ignoreForLoopInit"].(bool); ok {
			opts.IgnoreForLoopInit = ignore
		}
	}

	return opts
}

func buildInitializedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "initialized",
		Description: fmt.Sprintf("Variable '%s' should be initialized on declaration.", name),
	}
}

func buildNotInitializedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notInitialized",
		Description: fmt.Sprintf("Variable '%s' should not be initialized on declaration.", name),
	}
}

func getDeclarationNameText(sourceFile *ast.SourceFile, name *ast.Node) string {
	if name == nil || sourceFile == nil {
		return ""
	}
	if name.Kind == ast.KindIdentifier {
		if id := name.AsIdentifier(); id != nil {
			return id.Text
		}
	}
	nameRange := utils.TrimNodeTextRange(sourceFile, name)
	return sourceFile.Text()[nameRange.Pos():nameRange.End()]
}

func isConstDeclaration(varDecl *ast.VariableDeclaration) bool {
	if varDecl == nil || varDecl.Parent == nil {
		return false
	}
	if varDecl.Parent.Kind != ast.KindVariableDeclarationList {
		return false
	}
	declList := varDecl.Parent.AsVariableDeclarationList()
	if declList == nil {
		return false
	}
	return declList.Flags&ast.NodeFlagsConst != 0
}

func getForLoopKind(varDecl *ast.VariableDeclaration) (ast.Kind, bool) {
	if varDecl == nil || varDecl.Parent == nil || varDecl.Parent.Kind != ast.KindVariableDeclarationList {
		return ast.KindUnknown, false
	}
	declList := varDecl.Parent.AsVariableDeclarationList()
	if declList == nil || declList.Parent == nil {
		return ast.KindUnknown, false
	}
	switch declList.Parent.Kind {
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		return declList.Parent.Kind, true
	default:
		return ast.KindUnknown, false
	}
}

func isInAmbientContext(sourceFile *ast.SourceFile, node *ast.Node) bool {
	if sourceFile != nil && sourceFile.IsDeclarationFile {
		return true
	}

	for current := node; current != nil; current = current.Parent {
		if current.Kind == ast.KindVariableStatement {
			stmt := current.AsVariableStatement()
			if stmt != nil && utils.IncludesModifier(stmt, ast.KindDeclareKeyword) {
				return true
			}
		}

		if current.Kind == ast.KindModuleDeclaration {
			moduleDecl := current.AsModuleDeclaration()
			if moduleDecl != nil {
				if utils.IncludesModifier(moduleDecl, ast.KindDeclareKeyword) {
					return true
				}
				if ast.HasSyntacticModifier(current, ast.ModifierFlagsAmbient) {
					return true
				}
				if current.Flags&ast.NodeFlagsAmbient != 0 {
					return true
				}
			}
		}
	}

	return false
}

var InitDeclarationsRule = rule.CreateRule(rule.Rule{
	Name: "init-declarations",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				varDecl := node.AsVariableDeclaration()
				if varDecl == nil {
					return
				}

				nameNode := varDecl.Name()
				if nameNode == nil {
					return
				}

				if isInAmbientContext(ctx.SourceFile, node) {
					return
				}

				if opts.Mode == "never" && isConstDeclaration(varDecl) {
					return
				}

				forLoopKind, isForLoopInit := getForLoopKind(varDecl)
				if isForLoopInit && opts.IgnoreForLoopInit {
					return
				}

				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
				nameText := getDeclarationNameText(ctx.SourceFile, nameNode)

				hasInitializer := varDecl.Initializer != nil
				isForInOrOf := forLoopKind == ast.KindForInStatement || forLoopKind == ast.KindForOfStatement

				switch opts.Mode {
				case "always":
					if !hasInitializer && !isForInOrOf {
						ctx.ReportRange(nameRange, buildInitializedMessage(nameText))
					}
				case "never":
					if isForInOrOf {
						ctx.ReportRange(nameRange, buildNotInitializedMessage(nameText))
						return
					}
					if hasInitializer {
						reportRange := nameRange.WithEnd(varDecl.Initializer.End())
						ctx.ReportRange(reportRange, buildNotInitializedMessage(nameText))
					}
				}
			},
		}
	},
})
