package require_unicode_regexp

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type regexpCallTracker map[*ast.Node]bool

var regexpGlobalObjectNames = [...]string{"globalThis", "window", "self", "global"}

// sourceMayUseRegexpConstructor reports whether the file spells one of the
// syntactic roots from which this tracker can reach the built-in constructor.
// SourceFile.HasIdentifier collects normalized names, so escaped identifiers
// such as R\u0065gExp remain observable. Unknown source files are kept.
func sourceMayUseRegexpConstructor(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.AsNode().Kind != ast.KindSourceFile {
		return true
	}
	if sourceFile.HasIdentifier("RegExp") {
		return true
	}
	for _, name := range regexpGlobalObjectNames {
		if sourceFile.HasIdentifier(name) {
			return true
		}
	}
	return false
}

func newRegexpCallTracker(ctx rule.RuleContext) regexpCallTracker {
	calls := regexpCallTracker{}
	for _, node := range rule.TrackGlobalCalls(ctx, map[string]*rule.GlobalCallTrace{
		"RegExp": {Call: true, Construct: true},
	}, rule.GlobalCallOptions{Deduplicate: true}) {
		calls[node] = true
	}
	return calls
}

func (tracker regexpCallTracker) isCall(node *ast.Node) bool {
	return tracker[node]
}
