package require_mock_type_parameters

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
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
// site rather than by resolving the receiver. Their rewrite reads a wider set
// of call shapes than the hoisted module-mock APIs do, so it is read by
// parseModuleLoaderCall rather than by the shared parser: see the comment
// there.
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
					if loader := parseModuleLoaderCall(call); loader != nil &&
						loadsOneModule(call) &&
						!isShadowed(ctx, loader.namespace) {
						ctx.ReportNode(loader.memberNode, missingTypeParameterMessage(loader.member))
						return
					}
				}

				member, receiver := calledMember(call.Expression)
				if member == nil || member.Text() != mockFactory {
					return
				}
				if !isUtilitiesObject(ctx, receiver) {
					return
				}
				ctx.ReportNode(member, missingTypeParameterMessage(mockFactory))
			},
		}
	},
}

// moduleLoaderCall is one of the four module loaders as it is written at the
// call site: the member name, the node that names it, and the receiver.
type moduleLoaderCall struct {
	member     string
	memberNode *ast.Node
	namespace  *ast.Node
}

// parseModuleLoaderCall matches a module-loader call the way Rstest's rewrite
// matches it, which is not the way the hoisted module-mock APIs are matched.
// Measured on rstest 0.11.8, with `rs` imported from `@rstest/core`:
//
//	rs.importActual('./m')      rewritten    rs.mock('./m')      rewritten
//	rs['importActual']('./m')   rewritten    rs['mock']('./m')   throws
//	rs[`importActual`]('./m')   rewritten
//	rs.importActual?.('./m')    rewritten    rs.mock?.('./m')    throws
//	rs?.importActual('./m')     throws
//
// So a computed key and an optional call reach the loader rewrite and are
// matched here, while an optional receiver does not and is left alone. The
// shared parser keeps the narrower shape the mock family needs; widening it
// there would report calls that really do stay as the throwing stub.
//
// Parentheses and type-only syntax are transparent on both sides, for the same
// reason they are in the shared parser.
func parseModuleLoaderCall(call *ast.CallExpression) *moduleLoaderCall {
	callee := internalUtils.SkipAssertionsAndParens(call.Expression)
	if callee == nil {
		return nil
	}

	var member string
	var memberNode, namespace *ast.Node
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access == nil || access.QuestionDotToken != nil {
			return nil
		}
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return nil
		}
		member, memberNode, namespace = name.AsIdentifier().Text, name, access.Expression
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if access == nil || access.QuestionDotToken != nil {
			return nil
		}
		// A key that is not a plain string is not a name the build can read,
		// and a substitution is only known at run time.
		key := internalUtils.SkipAssertionsAndParens(access.ArgumentExpression)
		if key == nil ||
			(key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNoSubstitutionTemplateLiteral) {
			return nil
		}
		member, memberNode, namespace = key.Text(), key, access.Expression
	default:
		return nil
	}

	if !moduleLoaders[member] {
		return nil
	}

	namespace = internalUtils.SkipAssertionsAndParens(namespace)
	if namespace == nil || namespace.Kind != ast.KindIdentifier {
		return nil
	}
	if name := namespace.AsIdentifier().Text; name != "rs" && name != "rstest" {
		return nil
	}
	return &moduleLoaderCall{member: member, memberNode: memberNode, namespace: namespace}
}

// calledMember returns the property name a call reaches its callee through,
// and the expression it is read off. Both are nil when the callee is anything
// but a plain dotted member.
//
// A computed member is not matched: it is written to reach a property whose
// name is not a fixed identifier, and reporting the string inside the brackets
// would point at a value rather than at a member.
func calledMember(callee *ast.Node) (*ast.Node, *ast.Node) {
	callee = internalUtils.SkipAssertionsAndParens(callee)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return nil, nil
	}
	access := callee.AsPropertyAccessExpression()
	if access == nil {
		return nil, nil
	}
	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return nil, nil
	}
	return name, access.Expression
}

// isUtilitiesObject reports whether receiver names Rstest's utilities object.
//
// `fn` is an ordinary function on an ordinary object, so it is reached through
// ordinary bindings: a renamed import is the same function, and a local
// declaration of `rs` really is a different object. That is the opposite of
// how the plugin-managed members are matched, and it is why the receiver is
// resolved here and read as written there.
//
// The recognizer stays with this rule rather than being shared. Modeling the
// utilities object in general — its mock, timer and spy surface — is separate
// work, and `require_test_timeout` keeps its own narrower recognizer for the
// same reason.
func isUtilitiesObject(ctx rule.RuleContext, receiver *ast.Node) bool {
	receiver = internalUtils.SkipAssertionsAndParens(receiver)
	if receiver == nil {
		return false
	}

	if receiver.Kind == ast.KindIdentifier {
		// `rs` and `rstest` name the same object
		// (packages/core/src/runtime/api/public.ts), either may be imported
		// under a further name, and under `globals: true` both are on
		// `globalThis` with no import at all. A local declaration of the name
		// shadows both, and resolution reports that by returning no name.
		//
		// ctx.Refs.Resolve places an import and a local declaration from the
		// binder alone, so a file that declares the name is recognized whether
		// or not a TypeChecker is configured.
		name, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbol(
			receiver.AsIdentifier().Text,
			receiver,
			ctx.Refs.Resolve(receiver),
			ctx.SourceFile,
			rstestUtils.RstestImportModule,
		)
		return name == "rs" || name == "rstest"
	}

	// A namespace import or a whole-module require reaches the same object
	// through one more member: `core.rs.fn()`.
	member, namespace := calledMember(receiver)
	if member == nil || (member.Text() != "rs" && member.Text() != "rstest") {
		return false
	}
	namespace = internalUtils.SkipAssertionsAndParens(namespace)
	if namespace == nil || namespace.Kind != ast.KindIdentifier {
		return false
	}
	return testFramework.IsModuleNamespaceSymbol(
		ctx.Refs.Resolve(namespace),
		rstestUtils.RstestImportModule,
	)
}

// isShadowed reports whether the receiver name is declared in this file by
// something other than an import.
//
// Unlike the hoisted module-mock APIs, which are rewritten purely on the name
// written at the call site, the four module loaders are left as the runtime
// stub when the receiver is a local binding: a file that declares
// `const rs = { importActual: … }` — or takes `rs` as a parameter — runs its
// own object, while a file that reaches `rs` through an import runs the
// rewrite, whichever module the import came from. So an import of any kind is
// not shadowing here, and every other declaration is.
func isShadowed(ctx rule.RuleContext, receiver *ast.Node) bool {
	symbol := ctx.Refs.Resolve(receiver)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
			continue
		}
		switch declaration.Kind {
		case ast.KindImportSpecifier,
			ast.KindImportClause,
			ast.KindNamespaceImport,
			ast.KindImportEqualsDeclaration:
			continue
		}
		return true
	}
	return false
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
