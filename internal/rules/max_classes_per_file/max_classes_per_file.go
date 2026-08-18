package max_classes_per_file

import (
	_ "embed"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed max_classes_per_file.schema.json
var schemaJSON []byte

const defaultMax = 1

// MaxClassesPerFileRule enforces a maximum number of classes per file.
// https://eslint.org/docs/latest/rules/max-classes-per-file
var MaxClassesPerFileRule = rule.Rule{
	Name:   "max-classes-per-file",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		classCount := 0

		listeners := rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				classCount++
			},
		}
		if !opts.ignoreExpressions {
			listeners[ast.KindClassExpression] = func(node *ast.Node) {
				classCount++
			}
		}

		// ESLint reports on "Program:exit" using the source range of the
		// Program's own body (first statement start to last statement end),
		// not the range of any particular class. The end-of-file token is
		// always the last node the linter visits, so it stands in for
		// "Program:exit" here.
		statements := ctx.SourceFile.Statements
		if statements == nil || len(statements.Nodes) == 0 {
			return listeners
		}
		firstStatement := statements.Nodes[0]
		lastStatement := statements.Nodes[len(statements.Nodes)-1]

		listeners[rule.ListenerOnExit(ast.KindEndOfFile)] = func(node *ast.Node) {
			if classCount <= opts.max {
				return
			}
			textRange := core.NewTextRange(
				utils.TrimNodeTextRange(ctx.SourceFile, firstStatement).Pos(),
				lastStatement.End(),
			)
			ctx.ReportRange(textRange, rule.RuleMessage{
				Id: "maximumExceeded",
				Description: "File has too many classes (" +
					strconv.Itoa(classCount) + "). Maximum allowed is " +
					strconv.Itoa(opts.max) + ".",
				Data: map[string]string{
					"classCount": strconv.Itoa(classCount),
					"max":        strconv.Itoa(opts.max),
				},
			})
		}

		return listeners
	},
}

type ruleOptions struct {
	max               int
	ignoreExpressions bool
}

// parseOptions mirrors ESLint's destructuring of a bare max or an
// { ignoreExpressions, max } object: a bare number never carries
// ignoreExpressions, and an object without max falls back to the default.
func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{max: defaultMax}
	if len(options) == 0 {
		return opts
	}
	opts.max = utils.ResolveLegacyMaxOption(options[0], defaultMax)
	if m, ok := options[0].(map[string]any); ok {
		if v, ok := m["ignoreExpressions"].(bool); ok {
			opts.ignoreExpressions = v
		}
	}
	return opts
}
