package prefer_import_in_mock

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed prefer_import_in_mock.schema.json
var schemaJSON []byte

// utilityNames are the two names the module mock APIs are reachable under.
// `@rstest/core` exports the same utilities object twice, as `rstest` and as
// `rs`, and registers both onto `globalThis` under `globals: true`.
//
// The names are matched at the call site rather than resolved through an
// alias, because the module mock APIs are rewritten by a syntactic transform:
// a call written through a renamed binding is left untransformed and throws
// "[Rstest] mock() was not transformed by Rstest" at run time. Reporting such
// a call would offer a fix that keeps it broken, so it is left to whatever
// diagnoses the rename.
var utilityNames = map[string]bool{
	"rstest": true,
	"rs":     true,
}

// mockMethods are the module mock APIs whose first parameter is typed
// `string | Promise<T>`. `mockRequire` and `doMockRequire` take a `string`
// only — they mock the CommonJS entry, and a `Promise` is not part of their
// signature — so a path passed to either one stays as written.
var mockMethods = map[string]bool{
	"mock":   true,
	"doMock": true,
}

func buildPreferImportMessage(path string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferImport",
		Description: fmt.Sprintf("Replace %s with import(%s)", path, path),
		Data:        map[string]string{"path": path},
	}
}

type options struct {
	Fixable bool
}

func parseOptions(rawOptions []any) options {
	opts := options{Fixable: true}
	if len(rawOptions) == 0 {
		return opts
	}

	optsMap, _ := rawOptions[0].(map[string]any)
	if fixable, ok := optsMap["fixable"].(bool); ok {
		opts.Fixable = fixable
	}

	return opts
}

var PreferImportInMockRule = rule.Rule{
	Name:   "rstest/prefer-import-in-mock",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				pathArgument := mockPathArgument(ctx, node)
				if pathArgument == nil {
					return
				}

				pathText := utils.TrimmedNodeText(ctx.SourceFile, pathArgument)
				message := buildPreferImportMessage(pathText)
				if !opts.Fixable {
					ctx.ReportNode(pathArgument, message)
					return
				}

				// The path is wrapped as written rather than re-emitted from
				// the parsed string value, so an escape sequence, a quote
				// inside the path, and the project's quote style all survive.
				ctx.ReportNodeWithFixes(
					pathArgument,
					message,
					rule.RuleFixReplace(
						ctx.SourceFile,
						pathArgument,
						"import("+pathText+")",
					),
				)
			},
		}
	},
}

// mockPathArgument returns the module path of a `rs.mock` / `rs.doMock` call
// that is written as a plain string, or nil when the call is anything else.
func mockPathArgument(ctx rule.RuleContext, node *ast.Node) *ast.Node {
	call := node.AsCallExpression()
	if call == nil || call.QuestionDotToken != nil {
		return nil
	}

	if !isMockCallee(ctx, call.Expression) {
		return nil
	}

	// An explicit type argument, `rs.mock<{ value: number }>('./module')`,
	// already states the mocked module's shape. Wrapping the path would pin
	// that argument against the real module's type instead, so the call is
	// left alone.
	if call.TypeArguments != nil && len(call.TypeArguments.Nodes) > 0 {
		return nil
	}

	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return nil
	}
	pathArgument := call.Arguments.Nodes[0]
	if pathArgument == nil || pathArgument.Kind != ast.KindStringLiteral {
		return nil
	}
	return pathArgument
}

// isMockCallee reports whether expression is `rs.mock`, `rs.doMock`, or the
// same two under the `rstest` name, read off the utilities object exported by
// `@rstest/core` or off the Rstest global.
func isMockCallee(ctx rule.RuleContext, expression *ast.Node) bool {
	if expression == nil || expression.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := expression.AsPropertyAccessExpression()
	if access == nil || access.QuestionDotToken != nil {
		return false
	}

	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier || !mockMethods[name.Text()] {
		return false
	}

	receiver := access.Expression
	if receiver == nil || receiver.Kind != ast.KindIdentifier {
		return false
	}
	localName := receiver.Text()
	if !utilityNames[localName] {
		return false
	}

	// A local `rs` or `rstest` that is neither imported from `@rstest/core`
	// nor the framework global resolves to the empty name, which keeps a
	// same-named binding from another library out of the rule.
	resolvedName, _, _ := testFramework.ResolveFunctionIdentifierReference(
		localName,
		receiver,
		ctx.TypeChecker,
		ctx.SourceFile,
		rstestUtils.RstestImportModule,
	)
	return utilityNames[resolvedName]
}
