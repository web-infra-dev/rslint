package reactutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// IsCreateClassCall reports whether the given CallExpression's callee is
// `<createClass>(...)` or `<pragma>.<createClass>(...)`. Parentheses are
// skipped on both the callee and the pragma identifier. Pass the empty string
// for pragma/createClass to fall back to `DefaultReactPragma` /
// `DefaultReactCreateClass`.
func IsCreateClassCall(call *ast.CallExpression, pragma, createClass string) bool {
	if call == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	if createClass == "" {
		createClass = DefaultReactCreateClass
	}
	callee := ast.SkipParentheses(call.Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.AsIdentifier().Text == createClass
	case ast.KindPropertyAccessExpression:
		pa := callee.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return name.AsIdentifier().Text == createClass
	}
	return false
}

// ExtendsReactComponent reports whether `classNode` (a ClassDeclaration or
// ClassExpression) has an `extends` clause referencing `Component` or
// `PureComponent` — either as a bare identifier or qualified by the
// configured pragma (e.g. `React.Component`). Parentheses are skipped. Pass
// the empty string for pragma to default to `DefaultReactPragma`.
//
// NOTE: Matches the name regex used by eslint-plugin-react's
// `componentUtil.isES6Component` (`/^(Pure)?Component$/`). Aliased imports
// (e.g. `import { Component as C }`) are not resolved — same as the upstream
// rule.
func ExtendsReactComponent(classNode *ast.Node, pragma string) bool {
	if classNode == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	heritage := ast.GetClassExtendsHeritageElement(classNode)
	if heritage == nil {
		return false
	}
	hc := heritage.AsExpressionWithTypeArguments()
	if hc == nil || hc.Expression == nil {
		return false
	}
	expr := ast.SkipParentheses(hc.Expression)
	// OptionalChain in extends (`extends React?.Component`) is parsed as a
	// `ChainExpression` upstream, which `componentUtil.isES6Component` does
	// NOT match (it only inspects `MemberExpression` / `Identifier`). tsgo
	// flags an OptionalChain via `QuestionDotToken` on the same
	// PropertyAccessExpression, so we must explicitly reject it here to
	// stay aligned with upstream's no-match behavior.
	if ast.IsOptionalChain(expr) {
		return false
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		return isComponentName(expr.AsIdentifier().Text)
	case ast.KindPropertyAccessExpression:
		pa := expr.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		nameNode := pa.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return false
		}
		return isComponentName(nameNode.AsIdentifier().Text)
	}
	return false
}

func isComponentName(name string) bool {
	return name == "Component" || name == "PureComponent"
}

// ExtendsReactPureComponent reports whether `classNode` (a ClassDeclaration
// or ClassExpression) has an `extends` clause referencing `PureComponent` —
// either as a bare identifier or qualified by the configured pragma (e.g.
// `React.PureComponent`). Parentheses are skipped. Pass the empty string for
// pragma to default to `DefaultReactPragma`.
//
// Mirrors eslint-plugin-react's `componentUtil.isPureComponent`, which uses
// the regex `/^(<pragma>\.)?PureComponent$/` over the rendered extends-clause
// text. Plain `Component` does NOT match (use ExtendsReactComponent for the
// broader detection).
func ExtendsReactPureComponent(classNode *ast.Node, pragma string) bool {
	if classNode == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	heritage := ast.GetClassExtendsHeritageElement(classNode)
	if heritage == nil {
		return false
	}
	hc := heritage.AsExpressionWithTypeArguments()
	if hc == nil || hc.Expression == nil {
		return false
	}
	expr := ast.SkipParentheses(hc.Expression)
	switch expr.Kind {
	case ast.KindIdentifier:
		return expr.AsIdentifier().Text == "PureComponent"
	case ast.KindPropertyAccessExpression:
		pa := expr.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return name.AsIdentifier().Text == "PureComponent"
	}
	return false
}

func isObjectArgumentOf(call *ast.CallExpression, obj *ast.Node) bool {
	if call.Arguments == nil {
		return false
	}
	for _, arg := range call.Arguments.Nodes {
		if arg == obj {
			return true
		}
	}
	return false
}

// IsCreateReactClassObjectArg reports whether `obj` (an ObjectLiteralExpression)
// is the FIRST argument of a `<createClass>(...)` / `<pragma>.<createClass>(...)`
// call. Parens wrapping `obj` before it reaches the call argument position are
// transparent — tsgo preserves them while ESTree flattens — so
// `createReactClass(({...}))` still matches.
//
// Pass the empty string for pragma / createClass to fall back to
// `DefaultReactPragma` / `DefaultReactCreateClass`. Returns false for any
// non-ObjectLiteralExpression input, for objects in non-argument positions,
// and for calls whose callee is not the configured createClass name.
func IsCreateReactClassObjectArg(obj *ast.Node, pragma, createClass string) bool {
	if obj == nil || obj.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	cur := obj
	for cur.Parent != nil && cur.Parent.Kind == ast.KindParenthesizedExpression {
		cur = cur.Parent
	}
	parent := cur.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 || call.Arguments.Nodes[0] != cur {
		return false
	}
	return IsCreateClassCall(call, pragma, createClass)
}

// EnclosingClass returns the nearest ClassDeclaration / ClassExpression
// ancestor of `node`, or nil when `node` is at the top level. Used by rules
// that need to test whether a class member belongs to a React component.
func EnclosingClass(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	for p := node.Parent; p != nil; p = p.Parent {
		switch p.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return p
		}
	}
	return nil
}

// ClassKeywordStart returns the report-anchor start position for a
// ClassDeclaration / ClassExpression aligned with ESLint's `node.loc.start`
// for that class.
//
// ESTree wraps `export class …` and `export default class …` in
// `ExportNamedDeclaration` / `ExportDefaultDeclaration` and exposes the
// inner `ClassDeclaration` starting at the `class` keyword. tsgo flattens
// these — `export` / `default` end up as Modifier kinds on the class node
// itself, shifting `node.Pos()` to the `export` token. We trim those two
// modifier kinds back out to recover the upstream-aligned start.
//
// Decorators (`@dec`), TS `abstract` / `declare` / accessibility modifiers
// are PART of the class's range upstream — TSESTree keeps them on the
// ClassDeclaration. We stop trimming the moment a non-export/default
// modifier appears so the report range still spans those.
//
// Pass `sourceFile.Text()` as `text` so trailing trivia after the
// modifier list is properly skipped.
func ClassKeywordStart(text string, node *ast.Node) int {
	mods := node.Modifiers()
	pos := node.Pos()
	if mods != nil {
		for _, mod := range mods.Nodes {
			switch mod.Kind {
			case ast.KindExportKeyword, ast.KindDefaultKeyword:
				pos = mod.End()
			default:
				return scanner.SkipTrivia(text, mod.Pos())
			}
		}
	}
	return scanner.SkipTrivia(text, pos)
}

// ClassHasMethodNamed reports whether `classNode` has a method (regular
// MethodDeclaration, GetAccessor, SetAccessor, or Constructor) whose key
// resolves via IdentifierOrPrivateName to `methodName`. PropertyDeclaration
// (class-field assignment like `name = () => {}`) does NOT count — upstream's
// MethodDefinition listeners fire on real methods only, not class fields.
//
// Mirrors upstream `astUtil.getComponentProperties(node).some(p => getPropertyName(p) === '<methodName>')`
// for the subset of properties that have method semantics.
func ClassHasMethodNamed(classNode *ast.Node, methodName string) bool {
	if classNode == nil {
		return false
	}
	for _, m := range classNode.Members() {
		if m == nil {
			continue
		}
		switch m.Kind {
		case ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor:
			// continue
		default:
			continue
		}
		if IdentifierOrPrivateName(m.Name()) == methodName {
			return true
		}
	}
	return false
}
