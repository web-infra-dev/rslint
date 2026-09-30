package consistent_test_it

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed consistent_test_it.schema.json
var schemaJSON []byte

// Config configures a consistent-test-it rule for one test framework.
type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
}

// Runtime supplies framework-specific call semantics to the shared rule.
type Runtime struct {
	// Parse returns the describe or test registration for node, or nil. It is
	// asked on both entry and exit of every call, so it must answer the same
	// way twice or the suite depth becomes unbalanced.
	Parse func(*ast.Node) *testFramework.ParsedCall
	// Check decides whether a test registration spells preferred. When it
	// does not, it returns the node to report and the keyword named as the
	// opposite one in the message.
	Check func(node *ast.Node, parsed *testFramework.ParsedCall, preferred string) (report *ast.Node, opposite string, ok bool)
	// Fix builds the edits that respell the registration. It runs only when
	// edits are requested and may return nil.
	Fix func(node *ast.Node, parsed *testFramework.ParsedCall, preferred string) []rule.RuleFix
}

// ParseOptions returns the keywords required outside and inside a describe.
// `withinDescribe` falls back to `fn`, then to `it`.
func ParseOptions(options []any) (outside, inside string) {
	outside, inside = "test", "it"
	if len(options) == 0 {
		return
	}
	config, _ := options[0].(map[string]any)
	if value, ok := config["fn"].(string); ok && (value == "test" || value == "it") {
		outside, inside = value, value
	}
	if value, ok := config["withinDescribe"].(string); ok && (value == "test" || value == "it") {
		inside = value
	}
	return
}

// NewRule creates a consistent-test-it rule for a test framework.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			outside, inside := ParseOptions(options)
			runtime := config.Prepare(ctx)
			depth := 0
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := runtime.Parse(node)
					if parsed == nil {
						return
					}
					if parsed.Kind == testFramework.FnKindDescribe {
						depth++
						return
					}
					if parsed.Kind != testFramework.FnKindTest {
						return
					}
					preferred := outside
					id, key, suffix := "consistentMethod", "testKeyword", ""
					if depth > 0 {
						preferred = inside
						id, key, suffix = "consistentMethodWithinDescribe", "testKeywordWithinDescribe", " within describe"
					}
					report, opposite, ok := runtime.Check(node, parsed, preferred)
					if !ok {
						return
					}
					message := rule.RuleMessage{
						Id:          id,
						Description: fmt.Sprintf("Prefer using '%s' instead of '%s'%s", preferred, opposite, suffix),
						Data:        map[string]string{key: preferred, "oppositeTestKeyword": opposite},
					}
					ctx.ReportNodeWithDeferredFixes(report, message, func() []rule.RuleFix {
						return runtime.Fix(node, parsed, preferred)
					})
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					if parsed := runtime.Parse(node); parsed != nil && parsed.Kind == testFramework.FnKindDescribe {
						depth--
					}
				},
			}
		},
	}
}
