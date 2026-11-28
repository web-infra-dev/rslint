package valid_expect_in_promise

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/valid_expect_in_promise"
)

// sourceMayContainPromiseChain reports whether the file contains a call accepted
// by testFramework.IsPromiseChainCall, which recognizes `then`, `catch` and
// `finally` reached either as a property access or as a bracket access keyed by
// a string or no-substitution template literal.
//
// The source file's identifier table answers nearly every form exactly and in
// O(1). A property access spells the member as an identifier, and a keyword-named
// property is interned like any other; the parser also interns the cooked text of
// a string, no-substitution template or numeric element-access key, so
// promise["then"](), promise[`finally`]() and the escaped promise["\x74hen"]()
// put the member name in the same table. The table excludes the `catch` of a
// `try`/`catch`, which stays a keyword token and never reaches it; gating on the
// raw source text instead let every file with a `try`/`catch`, and every file
// with a backslash, through to the AST walk this function exists to avoid.
//
// The one form the table cannot answer is a key behind parentheses, as in
// promise[("then")](): the parser interns a key that is a literal itself, while
// IsPromiseChainCall reaches through parentheses with ast.SkipParentheses. Such
// a call opens its bracket on a parenthesis, and its raw text spells the member
// name unless the key is escaped, so requiring both keeps the fallback sound
// without paying for the brackets a test file spells everywhere else.
func sourceMayContainPromiseChain(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.AsNode().Kind != ast.KindSourceFile {
		return true
	}
	// The compiler initializes its name cache on demand; a file with no names
	// has no promise chain either way.
	for _, name := range promiseChainMemberNames {
		if sourceFile.HasIdentifier(name) {
			return nodeContainsPromiseChain(sourceFile.AsNode())
		}
	}

	// The bracket scan runs first: it answers almost every file with a single
	// IndexByte pass, while the four substring searches below cost a pass each.
	text := sourceFile.Text()
	if !bracketOpensOnParenthesis(text) {
		return false
	}
	if !strings.Contains(text, "then") &&
		!strings.Contains(text, "catch") &&
		!strings.Contains(text, "finally") &&
		!strings.Contains(text, `\`) {
		return false
	}
	return nodeContainsPromiseChain(sourceFile.AsNode())
}

// bracketOpensOnParenthesis reports whether any `[` in the text is followed,
// past whitespace and comments, by `(`. It reads the raw text rather than the
// AST, so a bracket inside a string or a regular expression counts too; that
// only costs a walk the caller would otherwise skip, and never hides one.
func bracketOpensOnParenthesis(text string) bool {
	for index := strings.IndexByte(text, '['); index >= 0; {
		if next := scanner.SkipTrivia(text, index+1); next < len(text) && text[next] == '(' {
			return true
		}
		offset := strings.IndexByte(text[index+1:], '[')
		if offset < 0 {
			return false
		}
		index += offset + 1
	}
	return false
}

var promiseChainMemberNames = [...]string{"then", "catch", "finally"}

func nodeContainsPromiseChain(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if testFramework.IsPromiseChainCall(node) {
		return true
	}
	// ForEachChild reports whether any visit returned true, so passing the
	// recursion itself keeps the walk free of a capturing closure.
	return node.ForEachChild(nodeContainsPromiseChain)
}

var ValidExpectInPromiseRule = shared.NewRule(shared.Config{
	Name:               "rstest/valid-expect-in-promise",
	MessageDescription: "This promise should either be returned or awaited to ensure the assertions in its chain are called",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		if !sourceMayContainPromiseChain(ctx.SourceFile) {
			return shared.Runtime{}
		}
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			TestCallbackFunctions: analysis.Callbacks().Functions,
			IsAssertionCall: func(node *ast.Node) bool {
				return analysis.IsExpectCall(node) || analysis.IsAssertCall(node)
			},
			IsAsyncAssertionSink: func(callNode, value *ast.Node) bool {
				if callNode == nil || callNode.Kind != ast.KindCallExpression {
					return false
				}
				top := rstestUtils.FindTopMostCallExpression(callNode)
				parsed := analysis.ParseExpectCall(top)
				if parsed == nil ||
					(!slices.Contains(parsed.Modifiers, "resolves") &&
						!slices.Contains(parsed.Modifiers, "rejects")) ||
					parsed.Head == nil || parsed.Head.AsCallExpression().Arguments == nil ||
					len(parsed.Head.AsCallExpression().Arguments.Nodes) == 0 {
					return false
				}
				return ast.SkipParentheses(parsed.Head.AsCallExpression().Arguments.Nodes[0]) ==
					ast.SkipParentheses(value)
			},
		}
	},
})
