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
	IsRegistration        func(*ast.Node) bool
	RegistrationCallbacks map[*ast.Node]bool
	Snapshots             func(*ast.Node) []Snapshot
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
			type groupState struct {
				snapshots []Snapshot
				depth     int
			}
			var snapshots []Snapshot
			var groups []groupState
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
			enterGroup := func(groupDepth int) {
				groups = append(groups, groupState{snapshots: snapshots, depth: depth})
				snapshots = nil
				depth = groupDepth
			}
			exitGroup := func() {
				flush()
				last := len(groups) - 1
				snapshots = groups[last].snapshots
				depth = groups[last].depth
				groups = groups[:last]
			}
			enter := func(node *ast.Node) {
				if runtime.RegistrationCallbacks[node] {
					enterGroup(1)
					return
				}
				depth++
			}
			enterRegistered := func(node *ast.Node) {
				if runtime.RegistrationCallbacks[node] {
					enterGroup(1)
				}
			}
			exitRegistered := func(node *ast.Node) {
				if runtime.RegistrationCallbacks[node] {
					exitGroup()
				}
			}
			exit := func(node *ast.Node) {
				if runtime.RegistrationCallbacks[node] {
					exitGroup()
					return
				}
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
				ast.KindMethodDeclaration:                        enter,
				ast.KindConstructor:                              enter,
				ast.KindGetAccessor:                              enter,
				ast.KindSetAccessor:                              enter,
				ast.KindFunctionDeclaration:                      enterRegistered,
				rule.ListenerOnExit(ast.KindMethodDeclaration):   exit,
				rule.ListenerOnExit(ast.KindConstructor):         exit,
				rule.ListenerOnExit(ast.KindGetAccessor):         exit,
				rule.ListenerOnExit(ast.KindSetAccessor):         exit,
				rule.ListenerOnExit(ast.KindFunctionDeclaration): exitRegistered,
				ast.KindCallExpression: func(node *ast.Node) {
					if runtime.IsRegistration(node) {
						enterGroup(0)
						return
					}
					snapshots = append(snapshots, runtime.Snapshots(node)...)
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					if runtime.IsRegistration(node) {
						exitGroup()
					}
				},
				rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) { flush() },
			}
		},
	}
}
