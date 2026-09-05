package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// RemoveMethodCallFixes mirrors Unicorn's removeMethodCall fixer. It removes
// the dotted property and the following argument list separately so explicit
// parentheses around the receiver or callee remain intact.
func RemoveMethodCallFixes(call DotMethodCall) []rule.RuleFix {
	return []rule.RuleFix{
		rule.RuleFixRemoveRange(core.NewTextRange(call.Object.End(), call.Callee.End())),
		rule.RuleFixRemoveRange(core.NewTextRange(call.RawCallee.End(), call.Call.End())),
	}
}
