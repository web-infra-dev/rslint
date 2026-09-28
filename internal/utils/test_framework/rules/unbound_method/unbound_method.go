package unbound_method

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	unboundMethod "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/unbound_method"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Config describes one test framework's unbound-method rule. The framework
// rules keep every check of @typescript-eslint/unbound-method and only exempt
// member references the framework knows are never invoked unbound.
type Config struct {
	Name string
	// Prepare returns the exemption for one file. It is called with each
	// member access before the base rule checks it; returning true skips the
	// member. Destructuring checks are never exempted.
	Prepare func(ctx rule.RuleContext) func(member *ast.Node) bool
}

// NewRule creates an unbound-method rule for a test framework.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:             config.Name,
		Schema:           unboundMethod.UnboundMethodRule.Schema,
		RequiresTypeInfo: true,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			return unboundMethod.CreateListeners(ctx, options, config.Prepare(ctx))
		},
	}
}
