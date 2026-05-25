package reactutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// SkipExpressionWrappers is a paren-and-TS-type-wrapper-transparent variant
// of `ast.SkipParentheses`. It additionally peels back tsgo's TS-only
// expression wrappers that ESLint's ESTree never produces: `as`-expressions,
// `satisfies`-expressions, `<T>x` type assertions, and `x!` non-null
// assertions. Use it whenever a rule must reach the underlying expression
// regardless of whether the source uses any of those wrappers — e.g. when
// matching a callee identifier, a JSX tag base, or a return-statement
// argument that may sit behind a `(x as Foo)`.
func SkipExpressionWrappers(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	for {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		default:
			return node
		}
	}
}

// SkipExpressionWrappersUp is the parent-walk equivalent of
// `SkipExpressionWrappers`: starting from `node.Parent`, walks up while the
// current parent is a transparent expression wrapper (`()` / `as` /
// `satisfies` / `<T>x` / `x!`) and returns the first non-wrapper ancestor,
// or nil when no such ancestor exists. Mirrors what ESTree implicitly does
// by flattening these wrappers — three sites in this rule used to inline
// the loop; one helper keeps them in lockstep.
func SkipExpressionWrappersUp(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	parent := node.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindNonNullExpression,
			ast.KindTypeAssertionExpression:
			parent = parent.Parent
			continue
		}
		break
	}
	return parent
}

// ParamListOpenParenPos returns the source position of the `(` that opens
// `node`'s parameter list, or -1 when the position cannot be located.
// Walks tokens after `node.Name().End()` via the scanner — robust against
// type-parameter lists (`<T>(p: T)`) where the `(` is not contiguous with
// the name. Use this when narrowing a diagnostic range on an
// object-literal shorthand method / getter / setter so the report site
// aligns with ESTree's `Property { value: FunctionExpression }` shape
// (FE.loc.start at `(`).
//
// `node` must be a MethodDeclaration / GetAccessor / SetAccessor (or
// anything with a non-nil `Name()`); other inputs return -1.
func ParamListOpenParenPos(sf *ast.SourceFile, node *ast.Node) int {
	if sf == nil || node == nil {
		return -1
	}
	name := node.Name()
	if name == nil {
		return -1
	}
	sc := scanner.GetScannerForSourceFile(sf, name.End())
	for {
		tok := sc.Token()
		if tok == ast.KindEndOfFile {
			return -1
		}
		if tok == ast.KindOpenParenToken {
			return sc.TokenStart()
		}
		sc.Scan()
	}
}

// IsObjectLiteralShorthandFunction reports whether `node` is a
// FunctionLike that, in ESTree, would be the inner FunctionExpression of
// a `Property { value: FunctionExpression }` — i.e. an object-literal
// shorthand method / getter / setter. Useful for callers that want to
// narrow diagnostic ranges to the parameter-list `(` (see
// ParamListOpenParenPos) so positions align with ESTree's reporting shape.
func IsObjectLiteralShorthandFunction(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return node.Parent.Kind == ast.KindObjectLiteralExpression
	}
	return false
}

// IdentifierOrPrivateName mirrors ESLint's `node.name` lookup over a key /
// member-name node: returns the Identifier text directly, or the
// PrivateIdentifier text with the leading `#` stripped (per ESTree spec
// `PrivateIdentifier.name` excludes the `#`). Returns "" for any other Kind
// (StringLiteral, NumericLiteral, ComputedPropertyName, …) — those never
// populate `.name` upstream and never match upstream's `key.name === 'X'`
// gate.
//
// Use this when a rule's upstream form is `X.name === 'Y'`. tsgo's
// PrivateIdentifier.Text retains the `#`, so a naive `.Text` compare would
// require the caller to know which input is private — this helper hides
// that.
func IdentifierOrPrivateName(name *ast.Node) string {
	if name == nil {
		return ""
	}
	switch name.Kind {
	case ast.KindIdentifier:
		return name.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(name.AsPrivateIdentifier().Text, "#")
	}
	return ""
}

// BindingIdentifierName returns the identifier text of a named declaration's
// binding, or "" when the declaration is anonymous, the binding is a pattern
// rather than a bare Identifier, or `n` is nil.
func BindingIdentifierName(n *ast.Node) string {
	if n == nil {
		return ""
	}
	name := n.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

// FunctionParameters returns the parameter list of a function-like node
// (FunctionDeclaration / FunctionExpression / ArrowFunction). Returns nil
// for nil input or any other node kind. Methods / accessors / constructors
// are intentionally not covered — callers that need them should add the
// kind explicitly to keep this helper a thin shim over the common shapes.
func FunctionParameters(fn *ast.Node) []*ast.Node {
	if fn == nil {
		return nil
	}
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		fd := fn.AsFunctionDeclaration()
		if fd.Parameters == nil {
			return nil
		}
		return fd.Parameters.Nodes
	case ast.KindFunctionExpression:
		fe := fn.AsFunctionExpression()
		if fe.Parameters == nil {
			return nil
		}
		return fe.Parameters.Nodes
	case ast.KindArrowFunction:
		af := fn.AsArrowFunction()
		if af.Parameters == nil {
			return nil
		}
		return af.Parameters.Nodes
	case ast.KindMethodDeclaration:
		// Object-literal shorthand methods (`{ Foo(props) {...} }`) and class
		// methods both use this kind in tsgo. ESTree wraps the former as a
		// `FunctionExpression`, so callers iterating "function-like
		// components" need MethodDeclaration to participate too.
		md := fn.AsMethodDeclaration()
		if md.Parameters == nil {
			return nil
		}
		return md.Parameters.Nodes
	case ast.KindGetAccessor:
		ga := fn.AsGetAccessorDeclaration()
		if ga.Parameters == nil {
			return nil
		}
		return ga.Parameters.Nodes
	case ast.KindSetAccessor:
		sa := fn.AsSetAccessorDeclaration()
		if sa.Parameters == nil {
			return nil
		}
		return sa.Parameters.Nodes
	case ast.KindConstructor:
		c := fn.AsConstructorDeclaration()
		if c.Parameters == nil {
			return nil
		}
		return c.Parameters.Nodes
	}
	return nil
}

// FunctionBody returns the body of a function-like node, mirroring
// `FunctionParameters`'s coverage. Returns nil when the kind is not
// function-like, or when the body is absent (overload signatures, abstract
// methods, ambient declarations).
//
// Arrow expression bodies (`() => expr`) are returned as the expression node
// itself — callers that distinguish block bodies from expression bodies
// should branch on `body.Kind == ast.KindBlock`.
func FunctionBody(fn *ast.Node) *ast.Node {
	if fn == nil {
		return nil
	}
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		return fn.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		return fn.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		return fn.AsArrowFunction().Body
	case ast.KindMethodDeclaration:
		return fn.AsMethodDeclaration().Body
	case ast.KindGetAccessor:
		return fn.AsGetAccessorDeclaration().Body
	case ast.KindSetAccessor:
		return fn.AsSetAccessorDeclaration().Body
	case ast.KindConstructor:
		return fn.AsConstructorDeclaration().Body
	}
	return nil
}

// FirstParamType returns the type annotation of the first parameter of `fn`
// (a FunctionDeclaration / FunctionExpression / ArrowFunction), or nil when
// the function has no parameters or the first parameter is untyped.
func FirstParamType(fn *ast.Node) *ast.Node {
	params := FunctionParameters(fn)
	if len(params) == 0 {
		return nil
	}
	pd := params[0].AsParameterDeclaration()
	if pd == nil {
		return nil
	}
	return pd.Type
}

// IsFunctionLikeForComponent reports whether `node` is a function-shaped node
// the React component-detection pipeline classifies as a "potential
// component" candidate. Covers FunctionDeclaration / FunctionExpression /
// ArrowFunction and the object-literal shorthand MethodDeclaration /
// GetAccessor / SetAccessor (upstream's ESTree shape exposes these as a
// `Property { method: true, value: FunctionExpression }`). Class methods
// share the same Kind values but are not function-shaped *components*; rule
// callers gate by parent / context where that matters.
func IsFunctionLikeForComponent(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return true
	}
	return false
}

// EntityNameRightmost returns the rightmost Identifier of a TypeReference's
// EntityName. For a bare `Foo`, returns `Foo`. For `A.B.C`, returns `C`.
// Returns nil if no identifier can be extracted.
func EntityNameRightmost(name *ast.Node) *ast.Node {
	for name != nil {
		switch name.Kind {
		case ast.KindIdentifier:
			return name
		case ast.KindQualifiedName:
			qn := name.AsQualifiedName()
			if qn == nil || qn.Right == nil {
				return nil
			}
			name = qn.Right
		default:
			return nil
		}
	}
	return nil
}
