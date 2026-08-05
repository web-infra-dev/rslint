package consistent_return

import (
	_ "embed"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

//go:embed consistent_return.schema.json
var schemaJSON []byte

func buildMissingReturnMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturn",
		Description: "Expected to return a value at the end of " + name + ".",
		Data:        map[string]string{"name": name},
	}
}

func buildMissingReturnValueMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturnValue",
		Description: name + " expected a return value.",
		Data:        map[string]string{"name": name},
	}
}

func buildUnexpectedReturnValueMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedReturnValue",
		Description: name + " expected no return value.",
		Data:        map[string]string{"name": name},
	}
}

type options struct {
	treatUndefinedAsUnspecified bool
}

func parseOptions(raw any) options {
	opts := options{treatUndefinedAsUnspecified: false}
	optsMap := utils.GetOptionsMap(raw)
	if optsMap == nil {
		return opts
	}
	if value, ok := optsMap["treatUndefinedAsUnspecified"].(bool); ok {
		opts.treatUndefinedAsUnspecified = value
	}
	return opts
}

// https://eslint.org/docs/latest/rules/consistent-return
var ConsistentReturnRule = rule.Rule{
	Name:   "consistent-return",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		opts := parseOptions(rule.LegacyUnwrapOptions(_options))

		// Every diagnostic depends on the code path a `return` belongs to, so
		// the returns are collected first and judged once the file is walked.
		// Collecting them also bounds the cost of the reachability question:
		// only a code path that returns a value ever needs a graph.
		var returns []*ast.Node
		listeners := rule.RuleListeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				returns = append(returns, node)
			},
		}

		statements := ctx.SourceFile.Statements
		if statements == nil || len(statements.Nodes) == 0 {
			return listeners
		}
		lastTopLevelNode := statements.Nodes[len(statements.Nodes)-1]
		listeners[rule.ListenerOnExit(lastTopLevelNode.Kind)] = func(node *ast.Node) {
			if node == lastTopLevelNode {
				report(&ctx, returns, opts)
			}
		}
		return listeners
	},
}

// diagnostic is one pending report. ESLint emits the per-`return` reports while
// it walks and the "missing return" reports at each function's exit, then sorts
// everything by position before handing it to the user; collecting and sorting
// here reproduces that order.
type diagnostic struct {
	textRange core.TextRange
	message   rule.RuleMessage
}

// codePathInfo mirrors ESLint's `funcInfo`: the first `return` reached in a code
// path decides whether every later one in it must also carry a value.
type codePathInfo struct {
	seenReturn     bool
	hasReturnValue bool
	name           string
	build          func(string) rule.RuleMessage
}

func report(ctx *rule.RuleContext, returns []*ast.Node, opts options) {
	infos := make(map[*ast.Node]*codePathInfo)
	var rootOrder []*ast.Node
	var diagnostics []diagnostic

	// The returns arrive in source order, which is the order ESLint's traversal
	// reaches them in. Consistency is decided by that order alone — a `return`
	// no code path can reach still counts.
	for _, node := range returns {
		root := cfg.RootOf(node)
		if root == nil || !isCheckedRoot(root) {
			continue
		}
		info := infos[root]
		if info == nil {
			info = &codePathInfo{}
			infos[root] = info
			rootOrder = append(rootOrder, root)
		}

		hasReturnValue := returnsValue(node, opts)
		if !info.seenReturn {
			info.seenReturn = true
			info.hasReturnValue = hasReturnValue
			info.name = inconsistentReturnName(root)
			if hasReturnValue {
				info.build = buildMissingReturnValueMessage
			} else {
				info.build = buildUnexpectedReturnValueMessage
			}
			continue
		}
		if info.hasReturnValue != hasReturnValue {
			diagnostics = append(diagnostics, diagnostic{
				textRange: utils.TrimNodeTextRange(ctx.SourceFile, node),
				message:   info.build(info.name),
			})
		}
	}

	for _, root := range rootOrder {
		if !infos[root].hasReturnValue || allowsImplicitReturn(root) {
			continue
		}
		if !cfg.Build(root, cfg.Hooks[struct{}]{}).EndReachable {
			continue
		}
		diagnostics = append(diagnostics, diagnostic{
			textRange: implicitReturnRange(ctx.SourceFile, root),
			message:   buildMissingReturnMessage(implicitReturnName(root)),
		})
	}

	slices.SortStableFunc(diagnostics, func(a, b diagnostic) int {
		return a.textRange.Pos() - b.textRange.Pos()
	})
	for _, d := range diagnostics {
		ctx.ReportRange(d.textRange, d.message)
	}
}

// isCheckedRoot reports whether a code path root is one ESLint judges. Upstream
// listens on `Program:exit` plus the three ESTree function types, which between
// them cover the source file and every tsgo function-like kind. A class static
// block and a class field initializer are left out, and neither may hold a
// `return`.
func isCheckedRoot(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindSourceFile,
		ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return true
	}
	return false
}

// returnsValue reports whether a `return` statement hands back a value. Under
// `treatUndefinedAsUnspecified` the identifier `undefined` and any `void`
// expression are read as handing back nothing.
func returnsValue(node *ast.Node, opts options) bool {
	expression := node.AsReturnStatement().Expression
	if expression == nil {
		return false
	}
	if !opts.treatUndefinedAsUnspecified {
		return true
	}
	// ESTree has no parenthesis node, so `return (undefined)` reads the same as
	// `return undefined` upstream.
	expression = ast.SkipParentheses(expression)
	if expression.Kind == ast.KindVoidExpression {
		return false
	}
	return expression.Kind != ast.KindIdentifier || expression.AsIdentifier().Text != "undefined"
}

// allowsImplicitReturn reports whether a code path may end without a value even
// though it returns one elsewhere: a class constructor, and a function whose own
// name starts with an uppercase letter, which ESLint reads as an ES5-style
// constructor.
func allowsImplicitReturn(node *ast.Node) bool {
	return node.Kind == ast.KindConstructor || utils.IsES5Constructor(node)
}

func implicitReturnName(node *ast.Node) string {
	if node.Kind == ast.KindSourceFile {
		return "program"
	}
	return utils.GetFunctionNameWithKindCore(node)
}

func inconsistentReturnName(node *ast.Node) string {
	if node.Kind == ast.KindSourceFile {
		return "Program"
	}
	return utils.UpperCaseFirstASCII(utils.GetFunctionNameWithKindCore(node))
}

// implicitReturnRange is where ESLint points the "missing return" report: the
// `=>` of an arrow function, the key of a class member or of an object-literal
// shorthand method, and otherwise the function's own name or the keyword that
// opens it.
func implicitReturnRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	switch node.Kind {
	case ast.KindSourceFile:
		// The head of the program. ESLint reports a bare line/column there,
		// with no end position of its own.
		return core.NewTextRange(0, 0)

	case ast.KindArrowFunction:
		return scanner.GetRangeOfTokenAtPosition(sourceFile, node.AsArrowFunction().EqualsGreaterThanToken.Pos())

	case ast.KindMethodDeclaration, ast.KindConstructor:
		// A class method and an object-literal shorthand method are both a
		// keyed member upstream, reported on that key.
		return memberKeyRange(sourceFile, node)

	case ast.KindGetAccessor, ast.KindSetAccessor:
		if node.Parent != nil && ast.IsClassLike(node.Parent) {
			return memberKeyRange(sourceFile, node)
		}
		// An object-literal accessor is an ordinary property upstream, and the
		// function it holds starts at the `(` of the parameter list.
		if name := node.Name(); name != nil {
			return scanner.GetRangeOfTokenAtPosition(sourceFile, name.End())
		}
	}

	if name := ownFunctionName(node); name != nil {
		return utils.TrimNodeTextRange(sourceFile, name)
	}
	return openingKeywordRange(sourceFile, node)
}

func memberKeyRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	name := node.Name()
	if name == nil {
		return utils.TrimNodeTextRange(sourceFile, node)
	}
	if name.Kind == ast.KindComputedPropertyName {
		// Upstream reports the key expression itself, not the brackets.
		name = ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	}
	return utils.TrimNodeTextRange(sourceFile, name)
}

// ownFunctionName returns the name a function carries itself, which is the only
// name upstream's `node.id` holds — a method's key and a variable the function
// is assigned to are not it.
func ownFunctionName(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Name()
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Name()
	}
	return nil
}

// openingKeywordRange returns the token an anonymous function opens with:
// `async` when it is async, `function` otherwise. `export` and `default` belong
// to the surrounding declaration upstream, so they are stepped over.
func openingKeywordRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	pos := node.Pos()
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier.Kind == ast.KindExportKeyword || modifier.Kind == ast.KindDefaultKeyword {
				pos = modifier.End()
			}
		}
	}
	return scanner.GetRangeOfTokenAtPosition(sourceFile, pos)
}
