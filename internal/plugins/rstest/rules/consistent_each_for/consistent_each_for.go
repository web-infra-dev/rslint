package consistent_each_for

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed consistent_each_for.schema.json
var schemaJSON []byte

// The option keys, which are also the registration names the messages use.
const (
	registrationDescribe = "describe"
	registrationIt       = "it"
	registrationTest     = "test"
)

// preferences maps a registration name to the parameterized form required for
// it. A name absent from the map is unconfigured, and the rule leaves every
// registration made through it alone — which is why an empty options object,
// the default, turns the whole rule into a no-op.
type preferences map[string]rstestUtils.RstestParameterizedKind

func parseOptions(rawOptions []any) preferences {
	if len(rawOptions) == 0 {
		return nil
	}
	optionsMap, ok := rawOptions[0].(map[string]any)
	if !ok || len(optionsMap) == 0 {
		return nil
	}
	parsed := make(preferences, len(optionsMap))
	for _, name := range []string{registrationDescribe, registrationIt, registrationTest} {
		value, ok := optionsMap[name].(string)
		if !ok {
			continue
		}
		switch rstestUtils.RstestParameterizedKind(value) {
		case rstestUtils.RstestParameterizedEach:
			parsed[name] = rstestUtils.RstestParameterizedEach
		case rstestUtils.RstestParameterizedFor:
			parsed[name] = rstestUtils.RstestParameterizedFor
		}
	}
	if len(parsed) == 0 {
		return nil
	}
	return parsed
}

// registrationName returns the option key that governs a parsed registration.
// A suite is always governed by `describe`, including when it was reached as
// `test.describe`, where the resolved name stays `test`; a test is governed by
// whichever of `test` and `it` it was written with.
func registrationName(parsed *rstestUtils.ParsedRstestFnCall) (string, bool) {
	switch parsed.Kind {
	case rstestUtils.RstestFnTypeDescribe:
		return registrationDescribe, true
	case rstestUtils.RstestFnTypeTest:
		if parsed.Name == registrationTest || parsed.Name == registrationIt {
			return parsed.Name, true
		}
	}
	return "", false
}

// reportNode picks the accessor the diagnostic points at. The written `.each`
// or `.for` is the honest anchor, but it is only present in MemberEntries when
// it was written at this call site: a registration that picked up its
// parameterized form through an alias, such as `const cases = test.for`, has
// the semantic kind and no member node to match. The identifier that resolves
// to the parameterized API is then what the reader can act on.
func reportNode(
	node *ast.Node,
	parsed *rstestUtils.ParsedRstestFnCall,
	actual rstestUtils.RstestParameterizedKind,
) *ast.Node {
	for _, entry := range parsed.MemberEntries {
		if entry.Name == string(actual) && entry.Node != nil {
			return entry.Node
		}
	}
	if parsed.Head.Local.Node != nil {
		return parsed.Head.Local.Node
	}
	return node
}

func buildConsistentMethodMessage(
	functionName string,
	preferred rstestUtils.RstestParameterizedKind,
	actual rstestUtils.RstestParameterizedKind,
) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "consistentMethod",
		Description: "Prefer using `" + functionName + "." + string(preferred) +
			"` over `" + functionName + "." + string(actual) + "`",
		Data: map[string]string{
			"functionName": functionName,
			"preferred":    string(preferred),
			"actual":       string(actual),
		},
	}
}

var ConsistentEachForRule = rule.Rule{
	Name:   "rstest/consistent-each-for",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		configured := parseOptions(options)
		if len(configured) == 0 {
			return rule.RuleListeners{}
		}

		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil || !parsed.IsParameterized() {
					return
				}
				name, ok := registrationName(parsed)
				if !ok {
					return
				}
				preferred, ok := configured[name]
				if !ok {
					return
				}
				actual := parsed.ParameterizedKind
				if actual == preferred {
					return
				}
				ctx.ReportNode(
					reportNode(node, parsed, actual),
					buildConsistentMethodMessage(name, preferred, actual),
				)
			},
		}
	},
}
