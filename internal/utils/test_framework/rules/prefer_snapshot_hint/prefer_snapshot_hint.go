// Package prefer_snapshot_hint shares snapshot hint policy and lexical grouping
// between the Jest and Rstest adapters.
package prefer_snapshot_hint

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Snapshot contains only the framework-neutral parts of a snapshot assertion.
type Snapshot struct {
	Matcher *ast.Node
	Args    []*ast.Node
	// Properties distinguishes the (properties?, hint?) overload from (hint?).
	Properties bool
}

type Runtime struct {
	IsRegistration func(*ast.Node) bool
	Snapshots      func(*ast.Node) []Snapshot
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
	// Rstest accepts interpolated strings as hints, since they are always strings.
	AllowInterpolatedHints bool
}

func missingHint(snapshot Snapshot, allowInterpolated bool) bool {
	if len(snapshot.Args) == 0 {
		return true
	}
	if !snapshot.Properties {
		return len(snapshot.Args) != 1
	}
	if len(snapshot.Args) == 2 {
		return false
	}
	arg := ast.SkipParentheses(snapshot.Args[0])
	return arg == nil || (arg.Kind != ast.KindStringLiteral &&
		arg.Kind != ast.KindNoSubstitutionTemplateLiteral &&
		(!allowInterpolated || arg.Kind != ast.KindTemplateExpression))
}

//go:embed prefer_snapshot_hint.schema.json
var schemaJSON []byte

var schema = rule.NewSchema(schemaJSON)

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: schema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			always := len(options) > 0 && options[0] == "always"
			runtime := config.Prepare(ctx)
			var snapshots []Snapshot
			var depths []int
			depth := 0
			flush := func() {
				if always || len(snapshots) > 1 {
					for _, snapshot := range snapshots {
						if missingHint(snapshot, config.AllowInterpolatedHints) {
							ctx.ReportNode(snapshot.Matcher, rule.RuleMessage{
								Id:          "missingHint",
								Description: "You should provide a hint for this snapshot",
							})
						}
					}
				}
				snapshots = nil
			}
			enter := func(*ast.Node) { depth++ }
			exit := func(*ast.Node) {
				depth--
				if always || depth == 0 {
					flush()
				}
			}
			return rule.RuleListeners{
				ast.KindFunctionExpression:                      enter,
				ast.KindArrowFunction:                           enter,
				rule.ListenerOnExit(ast.KindFunctionExpression): exit,
				rule.ListenerOnExit(ast.KindArrowFunction):      exit,
				// ESTree represents method, constructor and accessor bodies as
				// FunctionExpressions; tsgo gives them their own node kinds.
				ast.KindMethodDeclaration:                      enter,
				ast.KindConstructor:                            enter,
				ast.KindGetAccessor:                            enter,
				ast.KindSetAccessor:                            enter,
				rule.ListenerOnExit(ast.KindMethodDeclaration): exit,
				rule.ListenerOnExit(ast.KindConstructor):       exit,
				rule.ListenerOnExit(ast.KindGetAccessor):       exit,
				rule.ListenerOnExit(ast.KindSetAccessor):       exit,
				ast.KindCallExpression: func(node *ast.Node) {
					if runtime.IsRegistration(node) {
						depths = append(depths, depth)
						depth = 0
						return
					}
					snapshots = append(snapshots, runtime.Snapshots(node)...)
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					if runtime.IsRegistration(node) {
						depth = depths[len(depths)-1]
						depths = depths[:len(depths)-1]
					}
				},
				rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) { flush() },
			}
		},
	}
}
