package no_restricted_rstest_methods

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_restricted_rstest_methods.schema.json
var schemaJSON []byte

// restriction is one entry of the option object: the member is disallowed, and
// the author may have written a message saying what to reach for instead.
type restriction struct {
	Message    string
	HasMessage bool
}

func parseOptions(rawOptions []any) map[string]restriction {
	if len(rawOptions) == 0 {
		return nil
	}
	raw, ok := rawOptions[0].(map[string]any)
	if !ok {
		return nil
	}

	restricted := make(map[string]restriction, len(raw))
	for member, rawMessage := range raw {
		if rawMessage == nil {
			restricted[member] = restriction{}
			continue
		}
		message, ok := rawMessage.(string)
		if !ok {
			continue
		}
		if message == "" {
			restricted[member] = restriction{}
			continue
		}
		restricted[member] = restriction{Message: message, HasMessage: true}
	}
	return restricted
}

func restrictedMethodMessage(member string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "restrictedRstestMethod",
		Description: "Use of `" + member + "` is disallowed",
		Data:        map[string]string{"restriction": member},
	}
}

func restrictedMethodWithMessage(member string, message string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "restrictedRstestMethodWithMessage",
		Description: message,
		Data: map[string]string{
			"message":     message,
			"restriction": member,
		},
	}
}

var NoRestrictedRstestMethodsRule = rule.Rule{
	Name:   "rstest/no-restricted-rstest-methods",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		restricted := parseOptions(options)
		if len(restricted) == 0 {
			return rule.RuleListeners{}
		}

		report := func(memberNode *ast.Node, member string) {
			entry := restricted[member]
			if entry.HasMessage {
				ctx.ReportNode(memberNode, restrictedMethodWithMessage(member, entry.Message))
				return
			}
			ctx.ReportNode(memberNode, restrictedMethodMessage(member))
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// The two kinds of member on the utilities object are reached
				// in opposite ways, and which kind a name is cannot be known
				// until the option object is read, so both are tried here.
				if managed := rstestUtils.ParseRstestPluginManagedCall(node); managed != nil {
					if _, ok := restricted[managed.Member]; !ok {
						return
					}
					// The build hands the two members that resolve their
					// receiver back to a local declaration, so a call on the
					// file's own object is that object's, not Rstest's.
					if managed.API.ResolvesReceiver &&
						rstestUtils.ReceiverIsLocallyDeclared(ctx, managed.NamespaceNode) {
						return
					}
					report(managed.MemberNode, managed.Member)
					return
				}

				memberNode, member, receiver := calledMember(node)
				if memberNode == nil {
					return
				}
				// A plugin-managed member that did not parse above is written
				// in a shape the build does not rewrite, so the call reaches
				// the stub that throws rather than the API being restricted.
				if rstestUtils.IsPluginManagedAPI(member) {
					return
				}
				if _, ok := restricted[member]; !ok {
					return
				}
				if !rstestUtils.IsUtilitiesObject(ctx, receiver) {
					return
				}
				report(memberNode, member)
			},
		}
	},
}

// calledMember returns the member a call reaches its callee through, the name
// it is written as, and the expression it is read off.
//
// Every shape counts here, because the members matched this way are ordinary
// functions on an ordinary object: `rs['fn']()`, `rs.fn?.()` and `rs?.fn()`
// all really call `fn`. That is the opposite of the plugin-managed members,
// where the shape decides whether anything runs at all.
func calledMember(node *ast.Node) (*ast.Node, string, *ast.Node) {
	call := node.AsCallExpression()
	if call == nil {
		return nil, "", nil
	}
	callee := internalUtils.SkipAssertionsAndParens(call.Expression)
	if callee == nil {
		return nil, "", nil
	}

	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access == nil {
			return nil, "", nil
		}
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return nil, "", nil
		}
		return name, name.AsIdentifier().Text, access.Expression
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if access == nil {
			return nil, "", nil
		}
		key := internalUtils.SkipAssertionsAndParens(access.ArgumentExpression)
		name, ok := internalUtils.GetStaticStringLiteralValue(key)
		if !ok || name == "" {
			return nil, "", nil
		}
		return key, name, access.Expression
	default:
		return nil, "", nil
	}
}
