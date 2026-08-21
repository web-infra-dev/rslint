package no_this_alias

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_this_alias.schema.json
var schemaJSON []byte

type NoThisAliasOptions struct {
	AllowDestructuring bool     `json:"allowDestructuring"`
	AllowedNames       []string `json:"allowedNames"`
}

var (
	thisAssignmentMessage = rule.RuleMessage{
		Id:          "thisAssignment",
		Description: "Unexpected aliasing of 'this' to local variable.",
	}
	thisDestructureMessage = rule.RuleMessage{
		Id:          "thisDestructure",
		Description: "Unexpected aliasing of members of 'this' to local variables.",
	}
)

// aliasTarget returns the ESTree-equivalent VariableDeclarator id or
// AssignmentExpression left-hand side for a `this` value. ESTree elides
// parentheses, while the TypeScript AST preserves them.
func aliasTarget(node *ast.Node) *ast.Node {
	value := node
	parent := value.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		value = parent
		parent = value.Parent
	}

	if parent == nil {
		return nil
	}

	switch parent.Kind {
	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		if declaration.Initializer == value {
			return declaration.Name()
		}
	case ast.KindBinaryExpression:
		if !ast.IsAssignmentExpression(parent, false) {
			return nil
		}
		expression := parent.AsBinaryExpression()
		if expression.Right == value {
			return ast.SkipParentheses(expression.Left)
		}
	}

	return nil
}

func parseOptions(options []any) NoThisAliasOptions {
	opts := NoThisAliasOptions{AllowDestructuring: true}
	if len(options) == 0 {
		return opts
	}

	optsMap, ok := options[0].(map[string]interface{})
	if !ok {
		return opts
	}
	if allowDestructuring, ok := optsMap["allowDestructuring"].(bool); ok {
		opts.AllowDestructuring = allowDestructuring
	}
	if allowedNames, ok := optsMap["allowedNames"].([]interface{}); ok {
		opts.AllowedNames = make([]string, 0, len(allowedNames))
		for _, name := range allowedNames {
			if name, ok := name.(string); ok {
				opts.AllowedNames = append(opts.AllowedNames, name)
			}
		}
	}
	return opts
}

func reportTarget(ctx *rule.RuleContext, target *ast.Node, message rule.RuleMessage) {
	targetRange := target.Loc.WithPos(scanner.SkipTrivia(ctx.SourceFile.Text(), target.Pos()))
	// TSESTree includes a variable declarator's definite-assignment marker and
	// type annotation in the id range, including for binding patterns.
	if target.Parent != nil && target.Parent.Kind == ast.KindVariableDeclaration {
		declaration := target.Parent.AsVariableDeclaration()
		if declaration.Name() == target {
			if declaration.ExclamationToken != nil && declaration.ExclamationToken.End() > targetRange.End() {
				targetRange = targetRange.WithEnd(declaration.ExclamationToken.End())
			}
			if declaration.Type != nil && declaration.Type.End() > targetRange.End() {
				targetRange = targetRange.WithEnd(declaration.Type.End())
			}
		}
	}
	ctx.ReportRange(targetRange, message)
}

func defaultThisListener(ctx *rule.RuleContext) func(node *ast.Node) {
	return func(node *ast.Node) {
		target := aliasTarget(node)
		if target == nil || target.Kind != ast.KindIdentifier {
			return
		}
		reportTarget(ctx, target, thisAssignmentMessage)
	}
}

func configuredThisListener(ctx *rule.RuleContext, opts NoThisAliasOptions) func(node *ast.Node) {
	return func(node *ast.Node) {
		target := aliasTarget(node)
		if target == nil {
			return
		}

		if target.Kind == ast.KindIdentifier {
			name := target.AsIdentifier().Text
			for _, allowedName := range opts.AllowedNames {
				if name == allowedName {
					return
				}
			}
			reportTarget(ctx, target, thisAssignmentMessage)
			return
		}

		if !opts.AllowDestructuring {
			reportTarget(ctx, target, thisDestructureMessage)
		}
	}
}

var NoThisAliasRule = rule.CreateRule(rule.Rule{
	Name:   "no-this-alias",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if len(options) == 0 {
			return rule.RuleListeners{ast.KindThisKeyword: defaultThisListener(&ctx)}
		}
		opts := parseOptions(options)
		if opts.AllowDestructuring && len(opts.AllowedNames) == 0 {
			return rule.RuleListeners{ast.KindThisKeyword: defaultThisListener(&ctx)}
		}
		return rule.RuleListeners{ast.KindThisKeyword: configuredThisListener(&ctx, opts)}
	},
})
