package require_mock_type_parameters

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed require_mock_type_parameters.schema.json
var schemaJSON []byte

// mockFactory is the utilities-object member whose type parameter is the
// mock's own call signature. `fn: <T extends FunctionLike = FunctionLike>(fn?:
// T) => Mock<T>` falls back to a signature that takes anything and returns
// `any` whenever no implementation is passed to infer it from, so every
// argument and the return value go unchecked at every call of the mock.
const mockFactory = "fn"

// moduleLoaders are the members that hand back a whole module. Each is typed
// `<T = Record<string, unknown>>(path: string) => …`, and nothing about the
// path argument tells the compiler what the module exports, so the default is
// what a caller gets: an index signature with no named exports and no
// signatures on them.
//
// All four are plugin-managed, so they are matched by the syntax at the call
// site rather than by resolving the receiver. The four do not share one shape:
// the shared table has `importActual` and `requireActual` reading a bracketed
// string, an optional call and a local declaration of the receiver, while
// `importMock` and `requireMock` are matched on the name as written like the
// mock family. ParseRstestPluginManagedCall applies each member's own shape, so
// this list only says which members this rule checks.
var moduleLoaders = map[string]bool{
	"importActual":  true,
	"importMock":    true,
	"requireActual": true,
	"requireMock":   true,
}

func missingTypeParameterMessage(member string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingTypeParameter",
		Description: "'" + member + "' is called without a type parameter.",
		Data:        map[string]string{"member": member},
	}
}

type options struct {
	CheckImportFunctions bool
}

func parseOptions(rawOptions []any) options {
	opts := options{}
	if len(rawOptions) == 0 {
		return opts
	}

	optsMap, _ := rawOptions[0].(map[string]any)
	if check, ok := optsMap["checkImportFunctions"].(bool); ok {
		opts.CheckImportFunctions = check
	}

	return opts
}

var RequireMockTypeParametersRule = rule.Rule{
	Name:   "rstest/require-mock-type-parameters",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil {
					return
				}
				// The type parameter is written on the call, so a call that
				// already carries one is what this rule asks for. An empty
				// list does not parse, so any list at all is an explicit one.
				if call.TypeArguments != nil && len(call.TypeArguments.Nodes) > 0 {
					return
				}

				if opts.CheckImportFunctions {
					if loader := rstestUtils.ParseRstestPluginManagedCall(node); loader != nil &&
						moduleLoaders[loader.Member] &&
						loadsOneModule(call) &&
						!receiverTakesTheCall(ctx, loader) {
						ctx.ReportNode(loader.MemberNode, missingTypeParameterMessage(loader.Member))
						return
					}
				}

				member, receiver := rstestUtils.CalledPlainMember(call.Expression)
				if member == nil || member.Text() != mockFactory {
					return
				}
				if !rstestUtils.IsUtilitiesObject(ctx, receiver) {
					return
				}
				ctx.ReportNode(member, missingTypeParameterMessage(mockFactory))
			},
		}
	},
}

// receiverTakesTheCall reports whether this file's own declaration of the
// receiver name takes the call away from the rewrite, so that the loader runs
// as an ordinary method of an ordinary object and the type parameter this rule
// asks for would be meaningless.
//
// Only the members whose rewrite resolves the receiver can be taken that way.
// The others are rewritten on the name as written, so a local `const rs = { … }`
// is bypassed and the call is still Rstest's.
func receiverTakesTheCall(ctx rule.RuleContext, loader *rstestUtils.RstestPluginManagedCall) bool {
	return loader.API.ResolvesReceiver &&
		rstestUtils.ReceiverIsLocallyDeclared(ctx, loader.NamespaceNode)
}

// loadsOneModule reports whether Rstest's build actually rewrites this call.
//
// The four module loaders are rewritten into a module reference the bundler
// resolves, so the path has to be readable at build time and there is nothing
// else to pass: a call with any other argument count fails the build outright
// ("Invalid function call: `rs.importActual` function expects 1 argument"),
// and a path that is not a plain quoted string — an identifier, a template
// literal, a spread — is left as the runtime stub, which throws. A type
// parameter repairs none of those, so the call is left alone.
func loadsOneModule(call *ast.CallExpression) bool {
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return false
	}
	path := internalUtils.SkipAssertionsAndParens(call.Arguments.Nodes[0])
	return path != nil && path.Kind == ast.KindStringLiteral
}
