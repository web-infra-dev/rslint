package prefer_import_in_mock

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// transparentWrappers are the expression wrappers Rstest's mock transform
// never sees. TypeScript's type-only syntax — `as`, `satisfies`, `!` — is
// erased before the transform runs, and the transform reads through
// parentheses, so `(rs as any).mock('./dep')`, `rs!.mock('./dep')` and
// `(rs.mock)(('./dep'))` are all rewritten exactly like the bare form.
const transparentWrappers = ast.OEKParentheses |
	ast.OEKTypeAssertions |
	ast.OEKNonNullAssertions |
	ast.OEKSatisfies

// unwrap descends past the wrappers the transform cannot see.
func unwrap(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipOuterExpressions(node, transparentWrappers)
}

//go:embed prefer_import_in_mock.schema.json
var schemaJSON []byte

// utilityNames are the two names the module mock APIs are reachable under.
// `@rstest/core` exports the same utilities object twice, as `rstest` and as
// `rs`, and registers both onto `globalThis` under `globals: true`.
//
// The receiver is matched on the name written at the call site alone, with no
// attempt to resolve where the binding came from, because that is what Rstest
// does: its mock transform rewrites `rs.mock` and `rstest.mock` by spelling.
// A local object, and even a function parameter, that happens to be named `rs`
// has its `rs.mock(...)` hoisted and applied like any other; conversely a
// binding renamed away from both names, `import { rstest as vi }`, is never
// rewritten and throws "[Rstest] mock() was not transformed by Rstest" at run
// time. Resolving the identifier instead of reading it would disagree with the
// transform in both directions.
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
				argument, path := mockPathArgument(node)
				if path == nil {
					return
				}

				// The path is reused as written rather than re-emitted from
				// the parsed string value, so an escape sequence, a quote
				// inside the path, and the project's quote style all survive.
				pathText := utils.TrimmedNodeText(ctx.SourceFile, path)
				message := buildPreferImportMessage(pathText)
				if !opts.Fixable {
					ctx.ReportNode(argument, message)
					return
				}

				ctx.ReportNodeWithDeferredFixes(argument, message, func() []rule.RuleFix {
					// The whole argument is replaced, so a `('./dep')` loses
					// its parentheses and a `'./dep' as string` loses an
					// assertion that no longer describes a promise. A comment
					// written in the space that goes would be deleted with it,
					// so the diagnostic then stands without a fix.
					argumentRange := utils.TrimNodeTextRange(ctx.SourceFile, argument)
					pathRange := utils.TrimNodeTextRange(ctx.SourceFile, path)
					if commentBetween(ctx, argumentRange, pathRange) {
						return nil
					}
					return []rule.RuleFix{
						rule.RuleFixReplaceRange(argumentRange, "import("+pathText+")"),
					}
				})
			},
		}
	},
}

// mockPathArgument returns the first argument of a `rs.mock` / `rs.doMock`
// call that Rstest's mock transform rewrites, together with the plain string
// naming the module inside it. Both are nil when the call is anything else.
func mockPathArgument(node *ast.Node) (*ast.Node, *ast.Node) {
	call := node.AsCallExpression()
	if call == nil || call.QuestionDotToken != nil {
		return nil, nil
	}

	if !isMockCallee(call.Expression) || !isTransformablePosition(node) {
		return nil, nil
	}

	// An explicit type argument, `rs.mock<{ value: number }>('./module')`,
	// already states the mocked module's shape. Wrapping the path would pin
	// that argument against the real module's type instead, so the call is
	// left alone.
	if call.TypeArguments != nil && len(call.TypeArguments.Nodes) > 0 {
		return nil, nil
	}

	if call.Arguments == nil {
		return nil, nil
	}
	// The transform reads the path and the optional factory out of the
	// argument list positionally. It gives up on a spread, whose elements are
	// only known at run time, and a third argument fails the build outright;
	// neither failure is repaired by wrapping the path.
	arguments := call.Arguments.Nodes
	if len(arguments) == 0 || len(arguments) > 2 {
		return nil, nil
	}
	for _, argument := range arguments {
		if argument == nil || argument.Kind == ast.KindSpreadElement {
			return nil, nil
		}
	}

	// A template literal path is rejected by Rstest's build even without a
	// substitution, so only a quoted string reaches this far.
	path := unwrap(arguments[0])
	if path == nil || path.Kind != ast.KindStringLiteral {
		return nil, nil
	}
	return arguments[0], path
}

// commentBetween reports whether a comment sits inside outer but outside
// inner, which is exactly the text a whole-argument replacement would drop.
func commentBetween(ctx rule.RuleContext, outer core.TextRange, inner core.TextRange) bool {
	comments := ctx.Comments.All()
	return utils.HasCommentInSpan(comments, outer.Pos(), inner.Pos()) ||
		utils.HasCommentInSpan(comments, inner.End(), outer.End())
}

// isTransformablePosition reports whether the call stands on its own as a
// statement, which is the only place the mock transform can lift it out of.
// Rstest moves the call above the module's imports, so a call whose value is
// consumed — an argument to another call, a variable initializer, an operand
// of a comma expression, an awaited expression — is either left untransformed,
// and throws, or is lifted out of an expression that no longer parses without
// it. The wrappers the transform cannot see do not change the position.
func isTransformablePosition(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if ast.IsOuterExpression(parent, transparentWrappers) {
			continue
		}
		return parent.Kind == ast.KindExpressionStatement
	}
	return false
}

// isMockCallee reports whether expression is `rs.mock`, `rs.doMock`, or the
// same two under the `rstest` name. Wrappers the transform cannot see are
// unwrapped around both the callee and the receiver; an optional chain and a
// computed member are not rewritten by the transform, so they are rejected.
func isMockCallee(expression *ast.Node) bool {
	callee := unwrap(expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access == nil || access.QuestionDotToken != nil {
		return false
	}

	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier || !mockMethods[name.Text()] {
		return false
	}

	receiver := unwrap(access.Expression)
	return receiver != nil &&
		receiver.Kind == ast.KindIdentifier &&
		utilityNames[receiver.Text()]
}
