package id_length

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed id_length.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/id-length
//
// Upstream keys its SUPPORTED_EXPRESSIONS dispatch off ESTree parent types
// (MemberExpression, AssignmentPattern, VariableDeclarator, Property,
// ImportSpecifier, ImportDefaultSpecifier, ImportNamespaceSpecifier,
// RestElement, FunctionExpression, ArrowFunctionExpression, ClassDeclaration,
// FunctionDeclaration, MethodDefinition, PropertyDefinition, CatchClause,
// ArrayPattern). tsgo's AST does not carry that 1:1 grouping:
//
//   - ESTree gives plain function/arrow parameters (no default, no
//     destructuring, no rest) directly as children of the function node
//     itself, which is why FunctionDeclaration/FunctionExpression/
//     ArrowFunctionExpression are "always match" entries upstream. tsgo always
//     wraps every parameter — plain, defaulted, or rest — in a
//     ParameterDeclaration, so a single KindParameter case here subsumes
//     upstream's AssignmentPattern (defaults), RestElement (rest params), and
//     the parameter half of FunctionDeclaration/FunctionExpression/
//     ArrowFunctionExpression. Because a plain parameter is a direct child of
//     the function itself upstream, it is only checked when that function is
//     one of the supported nodes — see isCheckedParameterOwner. A rest or
//     defaulted parameter keeps its own supported wrapper and is checked
//     everywhere, TS signatures included.
//   - tsgo's KindArrowFunction case carries only the other half of upstream's
//     unconditional ArrowFunctionExpression entry: a concise body that is a
//     bare identifier (`(arg) => x`) is a direct child of the arrow just like
//     a plain parameter, so it matches too.
//   - ESTree's ObjectPattern/Property duplicates a shorthand identifier as two
//     nodes at the same range (`{ foo }` → `key: Identifier(foo), value:
//     Identifier(foo)`), which is why upstream deduplicates via a
//     `reportedNodes` range set. tsgo never duplicates nodes this way (a
//     shorthand binding/property is exactly one Identifier), so no dedup set
//     is needed here.
//   - ESTree represents both real bindings (var/let/const/params) and
//     assignment-expression destructuring (`({a} = {})`) as ObjectPattern/
//     ArrayPattern, decided once at parse time. tsgo only does this for real
//     bindings (ObjectBindingPattern/ArrayBindingPattern + BindingElement);
//     assignment-expression destructuring reuses plain
//     ObjectLiteralExpression/ArrayLiteralExpression (PropertyAssignment/
//     ShorthandPropertyAssignment/SpreadAssignment/SpreadElement), the same
//     nodes used for genuine object/array literals. Telling the two apart
//     needs a structural check — isPatternAssignmentTarget — mirroring
//     upstream's own parse-time pattern-conversion grammar (recursive through
//     parens/nested literals/spreads), for BOTH plain-identifier members
//     (Property's ObjectPattern branch) and rest members (RestElement).
//   - A computed key (`[x]() {}`, `{ [x]: 1 }`) is a single node with a
//     `.computed` flag in ESTree; tsgo wraps the key expression in a
//     dedicated ComputedPropertyName node one level below the member. This
//     rule dispatches on that wrapper's own parent to decide inclusion,
//     rather than reading a computed flag. Where the parent is a destructuring
//     property the wrapper is peeled off instead, since upstream still runs the
//     ordinary key/value comparison over the key expression itself.
//   - tsgo keeps parentheses as ParenthesizedExpression nodes and ESTree drops
//     them, so structural checks climb through them (skipParens). tsgo's
//     TS-only wrappers `x!`, `x as T` and `x satisfies T` are NOT transparent:
//     typescript-eslint keeps them as TSNonNullExpression / TSAsExpression /
//     TSSatisfiesExpression, none of those is a supported parent, and none of
//     them lets the pattern conversion reach the literal it wraps either —
//     see isPatternAssignmentTarget.
//
// Two upstream quirks are reproduced deliberately, verified against a local
// ESLint 10.8.0 install rather than assumed from reading the source:
//   - SUPPORTED_EXPRESSIONS.MemberExpression's checker function takes the
//     MemberExpression itself as its only parameter and never looks at which
//     side (object or property) the reported identifier occupies, so a
//     direct assignment target's object identifier is checked exactly like
//     its property identifier (`a.longName = 1` flags `a`, not just the
//     property). isValidMemberExpressionTarget below preserves this.
//   - SUPPORTED_EXPRESSIONS.PropertyDefinition/ClassDeclaration/
//     FunctionDeclaration/FunctionExpression/MethodDefinition are `true`
//     (unconditional) entries in upstream: any Identifier whose ESTree parent
//     is directly that node matches, not just the "name" slot. For
//     PropertyDeclaration this also matches the field initializer when it is
//     a bare identifier (`class C { x = q }` flags both `x` and `q`) because
//     ESTree's PropertyDefinition.value is a direct child just like .key.
//     tsgo's PropertyDeclaration has the same direct-child shape (Name and
//     Initializer both direct fields), so this rule matches either field
//     unconditionally, reproducing the quirk. ClassDeclaration's other
//     Identifier child is its superclass expression, which tsgo hides under a
//     HeritageClause / ExpressionWithTypeArguments pair — see
//     isClassExtendsExpression.
var IdLengthRule = rule.Rule{
	Name:   "id-length",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		// messageFor applies the length thresholds and the exception list to a
		// name, returning the diagnostic it earns (if any). It is shared by the
		// identifier listeners and by the constructor listener, which has no
		// identifier node to work from.
		messageFor := func(name string, isPrivate bool) (rule.RuleMessage, bool) {
			nameLength := utils.GraphemeCount(name)

			isShort := nameLength < opts.min
			isLong := opts.hasMax && nameLength > opts.max
			if !isShort && !isLong {
				return rule.RuleMessage{}, false
			}
			if opts.matchesException(name) {
				return rule.RuleMessage{}, false
			}

			if isShort {
				return messageTooShort(name, opts.min, isPrivate), true
			}
			return messageTooLong(name, opts.max, isPrivate), true
		}

		check := func(node *ast.Node) {
			name, isPrivate := identifierName(node)
			msg, ok := messageFor(name, isPrivate)
			if !ok {
				return
			}

			if !isValidPosition(opts, node) {
				return
			}

			ctx.ReportRange(declarationNameRange(ctx.SourceFile, node), msg)
		}

		checkConstructor := func(node *ast.Node) {
			// ESTree models a class constructor as a MethodDefinition whose key
			// is an ordinary Identifier named `constructor`, so upstream's
			// unconditional MethodDefinition entry measures it like any other
			// method name. tsgo's ConstructorDeclaration has no name node at
			// all — `constructor` is a keyword token — so the check is driven
			// from the declaration and reported on that token's range.
			msg, ok := messageFor("constructor", false)
			if !ok {
				return
			}
			if textRange, found := constructorKeywordRange(node, ctx.SourceFile); found {
				ctx.ReportRange(textRange, msg)
			}
		}

		return rule.RuleListeners{
			ast.KindIdentifier:        check,
			ast.KindPrivateIdentifier: check,
			ast.KindConstructor:       checkConstructor,
		}
	},
}

func identifierName(node *ast.Node) (name string, isPrivate bool) {
	if node.Kind == ast.KindPrivateIdentifier {
		// tsgo's PrivateIdentifier.Text retains the leading '#'; ESTree's
		// PrivateIdentifier.name (what upstream measures and formats with)
		// does not.
		return strings.TrimPrefix(node.AsPrivateIdentifier().Text, "#"), true
	}
	return node.AsIdentifier().Text, false
}

func messageTooShort(name string, minLength int, private bool) rule.RuleMessage {
	minText := strconv.Itoa(minLength)
	if private {
		return rule.RuleMessage{
			Id:          "tooShortPrivate",
			Description: fmt.Sprintf("Identifier name '#%s' is too short (< %d).", name, minLength),
			Data:        map[string]string{"name": name, "min": minText},
		}
	}
	return rule.RuleMessage{
		Id:          "tooShort",
		Description: fmt.Sprintf("Identifier name '%s' is too short (< %d).", name, minLength),
		Data:        map[string]string{"name": name, "min": minText},
	}
}

// messageTooLong intentionally mirrors upstream's tooLongPrivate template
// verbatim, including its '#' placement outside the quotes (unlike
// tooShortPrivate's '#' inside the quotes) — this asymmetry is present in
// ESLint's own meta.messages and is not a typo introduced here.
func messageTooLong(name string, maxLength int, private bool) rule.RuleMessage {
	maxText := strconv.Itoa(maxLength)
	if private {
		return rule.RuleMessage{
			Id:          "tooLongPrivate",
			Description: fmt.Sprintf("Identifier name #'%s' is too long (> %d).", name, maxLength),
			Data:        map[string]string{"name": name, "max": maxText},
		}
	}
	return rule.RuleMessage{
		Id:          "tooLong",
		Description: fmt.Sprintf("Identifier name '%s' is too long (> %d).", name, maxLength),
		Data:        map[string]string{"name": name, "max": maxText},
	}
}

type idLengthOptions struct {
	min               int
	max               int
	hasMax            bool
	properties        bool
	exceptions        map[string]bool
	exceptionPatterns []*esregexp.RegExp
}

const defaultMin = 2

func parseOptions(options []any) idLengthOptions {
	opts := idLengthOptions{min: defaultMin, properties: true}
	if len(options) == 0 {
		return opts
	}
	m, _ := options[0].(map[string]any)
	if m == nil {
		return opts
	}
	if v, ok := utils.CoerceInt(m["min"]); ok {
		opts.min = v
	}
	if v, ok := utils.CoerceInt(m["max"]); ok {
		opts.max = v
		opts.hasMax = true
	}
	if v, ok := m["properties"].(string); ok {
		opts.properties = v != "never"
	}
	if arr, ok := m["exceptions"].([]any); ok && len(arr) > 0 {
		opts.exceptions = make(map[string]bool, len(arr))
		for _, e := range arr {
			if s, ok := e.(string); ok {
				opts.exceptions[s] = true
			}
		}
	}
	if arr, ok := m["exceptionPatterns"].([]any); ok {
		for _, p := range arr {
			s, ok := p.(string)
			if !ok {
				continue
			}
			// Upstream compiles each pattern as `new RegExp(pattern, "u")`, so
			// lookaround, backreferences and `\p{...}` all behave as they do
			// there (Go's RE2 supports none of them). An unparsable pattern is
			// rejected up front by the schema's `format: "regex"`, so the
			// compile-error branch here is only defensive.
			if re, err := esregexp.Compile(s, "u"); err == nil {
				opts.exceptionPatterns = append(opts.exceptionPatterns, re)
			}
		}
	}
	return opts
}

func (o idLengthOptions) matchesException(name string) bool {
	if o.exceptions[name] {
		return true
	}
	for _, re := range o.exceptionPatterns {
		if re.Test(name) {
			return true
		}
	}
	return false
}

// isValidPosition reports whether node sits in one of the identifier
// positions this rule inspects (declaration/binding names, not ordinary
// reads). It mirrors upstream's SUPPORTED_EXPRESSIONS dispatch — see the
// package doc comment above for the tsgo AST-shape translation.
func isValidPosition(opts idLengthOptions, node *ast.Node) bool {
	// ESTree drops parentheses entirely, so every SUPPORTED_EXPRESSIONS check
	// upstream is implicitly transparent to them. tsgo keeps them as real
	// nodes, so climb through any parentheses around the identifier itself
	// before dispatching on its structural parent. `wrapped` (rather than the
	// raw node) is what gets compared against sibling expression-position
	// fields below (Initializer, Expression) — declaration-name positions
	// (Name()) can never themselves be parenthesized, so using `wrapped` there
	// too is equivalent, not just safe. Name comparisons, by contrast, always
	// read through to the identifier itself, since ESTree has no wrapper for
	// them to hide behind.
	wrapped := skipParens(node)
	effectiveParent := wrapped.Parent
	if effectiveParent == nil {
		return false
	}

	switch effectiveParent.Kind {
	case ast.KindVariableDeclaration:
		// Covers var/let/const declarators AND catch-clause bindings: tsgo
		// parses a catch binding with the same parseVariableDeclaration used
		// for var/let/const, so `catch (e)` reaches this case too.
		return effectiveParent.Name() == wrapped

	case ast.KindBindingElement:
		return isValidBindingElement(opts, effectiveParent, node)

	case ast.KindPropertyAssignment:
		// NOTE: Unlike ESLint, a non-computed PropertyAssignment key inside an
		// object literal that is itself a destructuring-assignment target
		// (e.g. the first `a` in `({ a: a } = {})`) is never handed to this
		// listener at all: the linter's traversal, when propagating
		// "allow pattern" context into an assignment target (mirroring
		// ESLint's own internal pattern-conversion walk), only visits such a
		// key when it is computed. This only matters for the isSame branch
		// below (`{ a: a }` with identical key/value text) — the differing
		// case only ever needs the value side, which is always visited. See
		// id_length.md's "Differences from ESLint".
		return isValidPropertyAssignmentChild(opts, effectiveParent, node, wrapped)

	case ast.KindShorthandPropertyAssignment:
		if effectiveParent.Name() != wrapped {
			return false
		}
		// `({ a = 1 } = {})`: upstream's pattern conversion gives the binding
		// an AssignmentPattern parent rather than the Property itself, and
		// that entry is unconditional, so `properties` does not exempt it.
		if effectiveParent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer != nil {
			return true
		}
		// Upstream's ObjectPattern and non-pattern branches converge on an
		// identical outcome for shorthand `{ a }`: key and value are the same
		// node, so both branches reduce to "properties enabled and this is
		// that node". See id_length.md analysis in the package doc.
		return opts.properties && !isImportAttributeKey(node)

	case ast.KindImportSpecifier:
		return isValidImportSpecifier(effectiveParent, wrapped)

	case ast.KindImportClause:
		return effectiveParent.Name() == wrapped

	case ast.KindNamespaceImport:
		return effectiveParent.Name() == wrapped

	case ast.KindParameter:
		if effectiveParent.Name() != wrapped {
			return false
		}
		param := effectiveParent.AsParameterDeclaration()
		// A rest or defaulted parameter is a RestElement or an
		// AssignmentPattern in ESTree — both unconditional entries — no matter
		// what function-like node holds it, so neither of the two exclusions
		// below applies to it.
		if param.DotDotDotToken != nil || param.Initializer != nil {
			return true
		}
		// A plain parameter is a direct child of the function node itself, so
		// it is only checked when that node is one of the supported ones. TS's
		// signature-only function-likes (`type Fn = (x) => void`, `declare
		// function fn(x)`, `interface I { m(x) }`, an overload, an abstract or
		// otherwise body-less method) become TSFunctionType,
		// TSDeclareFunction, TSMethodSignature or TSEmptyBodyFunctionExpression
		// instead, none of which is in SUPPORTED_EXPRESSIONS.
		if !isCheckedParameterOwner(effectiveParent.Parent) {
			return false
		}
		// A TS parameter property (`constructor(private x: number)`) has no
		// supported ESTree parent either: typescript-eslint wraps the parameter
		// in a TSParameterProperty, which is absent from SUPPORTED_EXPRESSIONS,
		// so upstream never checks the binding name. A default value puts an
		// AssignmentPattern back between the two, which is why the rest/default
		// branch above returns before this check.
		return !ast.IsParameterPropertyDeclaration(effectiveParent, effectiveParent.Parent)

	case ast.KindArrowFunction:
		// SUPPORTED_EXPRESSIONS.ArrowFunctionExpression is unconditional, and
		// an arrow has no name, so the only identifier that can be a direct
		// child besides a plain parameter is a concise body that is a bare
		// identifier (`(arg) => x`).
		return effectiveParent.Body() == wrapped

	case ast.KindClassDeclaration:
		return effectiveParent.Name() == wrapped

	case ast.KindExpressionWithTypeArguments:
		// ESTree hangs a class's superclass expression off the
		// ClassDeclaration as directly as its name, and that entry is
		// unconditional, so a bare-identifier superclass is checked just like
		// the name. tsgo routes it through a HeritageClause /
		// ExpressionWithTypeArguments pair instead.
		return isClassExtendsExpression(effectiveParent, wrapped)

	case ast.KindFunctionDeclaration:
		// A body-less function declaration — an overload signature or an
		// ambient `declare function` — is a TSDeclareFunction in ESTree, not
		// a FunctionDeclaration, so its name is not checked. (A body-less
		// class method keeps its MethodDefinition shape and stays checked.)
		if effectiveParent.Body() == nil {
			return false
		}
		return effectiveParent.Name() == wrapped

	case ast.KindFunctionExpression:
		return effectiveParent.Name() == wrapped

	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if effectiveParent.Name() != wrapped {
			return false
		}
		// tsgo reuses these kinds for class members and object-literal members
		// alike; ESTree does not. A class member is a MethodDefinition
		// (unconditional upstream), while an object-literal member is a
		// Property with `method`/`kind` set, which goes through Property's
		// non-pattern branch and is therefore gated on `properties`.
		switch effectiveParent.Parent.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return isPlainClassMember(effectiveParent)
		case ast.KindObjectLiteralExpression:
			return isValidPlainProperty(opts, effectiveParent.Name(), node)
		}
		// An interface or type-literal member is a TSMethodSignature, which is
		// not a supported parent.
		return false

	case ast.KindPropertyDeclaration:
		if !isPlainClassMember(effectiveParent) {
			return false
		}
		pd := effectiveParent.AsPropertyDeclaration()
		return wrapped == effectiveParent.Name() || wrapped == pd.Initializer

	case ast.KindComputedPropertyName:
		// ESTree keeps a computed key as the Property's own `key`, so a
		// destructuring pattern's computed key still goes through the very same
		// key/value comparison as a plain one (`const {[x]: x} = obj` reports
		// the key, not the value). Route it there instead of treating the
		// wrapper as a position of its own.
		owner := effectiveParent.Parent
		if owner == nil {
			return false
		}
		switch owner.Kind {
		case ast.KindBindingElement:
			return isValidBindingElement(opts, owner, node)
		case ast.KindPropertyAssignment:
			return isValidPropertyAssignmentChild(opts, owner, node, wrapped)
		}
		return isValidComputedPropertyNameOwner(owner)

	case ast.KindPropertyAccessExpression:
		if !opts.properties {
			return false
		}
		return isValidMemberExpressionTarget(effectiveParent)

	case ast.KindArrayLiteralExpression:
		// An array literal that is a destructuring-assignment target
		// (`([x] = source)`, `for ([x] of source)`) is an ArrayPattern
		// upstream, whose entry is unconditional.
		return isPatternAssignmentTarget(wrapped)

	case ast.KindBinaryExpression:
		// `x = 0` inside a destructuring-assignment target (`({ key: x = 0 } =
		// source)`, `([x = 0] = source)`) is an AssignmentPattern upstream, not
		// an AssignmentExpression, and that entry is unconditional — so
		// `properties: "never"` does not exempt it either. A plain assignment
		// expression is not a supported parent, hence the assignment-target
		// test on the BinaryExpression itself.
		be := effectiveParent.AsBinaryExpression()
		return be.OperatorToken.Kind == ast.KindEqualsToken && be.Left == wrapped &&
			isPatternAssignmentTarget(effectiveParent)

	case ast.KindSpreadElement:
		// isPatternAssignmentTarget is called on wrapped (not
		// effectiveParent): its climb decides how to continue by switching on
		// node.Parent.Kind, so it must be invoked starting from the
		// identifier itself, not from the SpreadElement/SpreadAssignment
		// wrapper already one hop up. This also matters because the climb
		// only special-cases ArrayLiteralExpression as a "keep climbing"
		// parent, not ObjectLiteralExpression — it is only reachable by first
		// landing on the PropertyAssignment/ShorthandPropertyAssignment/
		// SpreadAssignment case, which jumps straight past it to its own
		// parent.
		se := effectiveParent.AsSpreadElement()
		return se.Expression == wrapped && isPatternAssignmentTarget(wrapped)

	case ast.KindSpreadAssignment:
		sa := effectiveParent.AsSpreadAssignment()
		return sa.Expression == wrapped && isPatternAssignmentTarget(wrapped)
	}

	return false
}

// skipParens walks up from node through any chain of parentheses and returns
// the outermost one — i.e. the node whose own .Parent is the first ancestor
// ESTree would also see. When node is not parenthesized it is returned
// unchanged.
//
// Only parentheses are transparent here. tsgo's TS-only wrapper expressions
// (`x!`, `x as T`, `x satisfies T`) do survive into typescript-eslint's AST as
// TSNonNullExpression / TSAsExpression / TSSatisfiesExpression, and none of
// those is a supported parent, so an identifier under one is never checked.
func skipParens(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}

// isPatternAssignmentTarget reports whether node sits inside the left-hand
// side of a destructuring assignment — the object/array literals upstream's
// parser has already turned into ObjectPattern/ArrayPattern by the time the
// rule runs.
//
// ast.GetAssignmentTarget answers almost the same question, but it climbs
// through a `!` non-null assertion (TS accepts `a! = b` as a target), and
// typescript-eslint does not: it keeps the wrapper as a TSNonNullExpression
// and hands its operand to the ordinary expression conversion, so the literal
// inside `([x]! = source)` stays an ArrayExpression and nothing within it is a
// pattern. Rejecting a path that crosses one restores that.
func isPatternAssignmentTarget(node *ast.Node) bool {
	target := ast.GetAssignmentTarget(node)
	if target == nil {
		return false
	}
	for n := node; n != nil && n != target; n = n.Parent {
		if n.Kind == ast.KindNonNullExpression {
			return false
		}
	}
	return true
}

// isClassExtendsExpression reports whether wrapped is the superclass
// expression of a class declaration — the one identifier position under
// tsgo's heritage-clause wrapper that ESTree exposes as a direct child of a
// supported node. An `implements` clause and an interface's own `extends`
// become TSClassImplements / TSInterfaceHeritage upstream, and a class
// expression's superclass hangs off ClassExpression; none of those three is a
// supported parent.
func isClassExtendsExpression(expressionWithTypeArguments, wrapped *ast.Node) bool {
	if expressionWithTypeArguments.AsExpressionWithTypeArguments().Expression != wrapped {
		return false
	}
	clause := expressionWithTypeArguments.Parent
	if clause == nil || clause.Kind != ast.KindHeritageClause ||
		clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
		return false
	}
	return clause.Parent != nil && clause.Parent.Kind == ast.KindClassDeclaration
}

// isValidBindingElement handles a destructuring binding element (var/let/
// const/parameter/catch destructuring). Its container distinguishes an
// object-pattern element (upstream's Property-in-ObjectPattern branch, with
// key/value semantics) from an array-pattern element (upstream's
// unconditional ArrayPattern entry).
func isValidBindingElement(opts idLengthOptions, bindingElement *ast.Node, node *ast.Node) bool {
	be := bindingElement.AsBindingElement()
	container := bindingElement.Parent
	switch container.Kind {
	case ast.KindObjectBindingPattern:
		// A rest (`{ ...a }`) or defaulted (`{ a = 1 }`, `{ a: b = 1 }`)
		// element is not an ESTree Property at all: the pattern conversion
		// hands the binding name to a RestElement or to an AssignmentPattern's
		// `left`, both unconditional entries upstream, so `properties: never`
		// does not exempt them. The property name of `{ a: b = 1 }` stays
		// unchecked, matching Property's ObjectPattern branch, whose
		// `parent.value === node` can never hold once the value is wrapped.
		if be.DotDotDotToken != nil || be.Initializer != nil {
			return bindingElement.Name() == node
		}
		keyNode := be.PropertyName
		valueNode := bindingElement.Name()
		if keyNode == nil {
			keyNode = valueNode
		}
		return isValidPatternPropertyValueNodes(opts, keyNode, valueNode, node, node)
	case ast.KindArrayBindingPattern:
		return bindingElement.Name() == node
	}
	return false
}

// isValidPropertyAssignmentChild dispatches a PropertyAssignment's key or
// value identifier to upstream's ObjectPattern or non-pattern Property branch,
// depending on whether the enclosing object literal is (transitively) a
// destructuring-assignment target. `node` is the identifier itself and
// `wrapped` the outermost pair of parentheses around it, if any.
func isValidPropertyAssignmentChild(opts idLengthOptions, propertyAssignment, node, wrapped *ast.Node) bool {
	pa := propertyAssignment.AsPropertyAssignment()
	if isPatternAssignmentTarget(propertyAssignment.Parent) {
		return isValidPatternPropertyValueNodes(opts, propertyAssignment.Name(), pa.Initializer, node, wrapped)
	}
	return isValidPlainProperty(opts, propertyAssignment.Name(), node)
}

// isValidPatternPropertyValueNodes ports upstream's Property/ObjectPattern
// branch: when key and value spell the same name only the key is checked (and
// only with `properties` enabled); otherwise only the value is.
func isValidPatternPropertyValueNodes(opts idLengthOptions, keyNode, valueNode, node, wrapped *ast.Node) bool {
	key := patternKeyNode(keyNode)
	if sameIdentifierText(key, unwrapParens(valueNode)) {
		return node == key && opts.properties
	}
	return wrapped == valueNode
}

// patternKeyNode returns the node ESTree exposes as a pattern Property's key:
// a computed key's own expression, with parentheses dropped.
func patternKeyNode(keyNode *ast.Node) *ast.Node {
	for keyNode != nil && keyNode.Kind == ast.KindComputedPropertyName {
		keyNode = keyNode.AsComputedPropertyName().Expression
	}
	return unwrapParens(keyNode)
}

func unwrapParens(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// isValidPlainProperty handles a PropertyAssignment used as a genuine object
// literal member (not a destructuring pattern). Upstream's non-pattern
// branch compares `parent.key.name === node.name` without regard to which of
// key/value node actually is — so a value identifier that happens to share
// the key's text is also flagged (`{ a: a }` flags both `a`s in an ordinary
// object literal). isValidPlainProperty preserves that.
func isValidPlainProperty(opts idLengthOptions, keyNode, node *ast.Node) bool {
	if !opts.properties || isImportAttributeKey(node) {
		return false
	}
	// Upstream's non-pattern branch guards on `!parent.computed`, so a computed
	// key exempts both halves of the property.
	if keyNode != nil && keyNode.Kind == ast.KindComputedPropertyName {
		return false
	}
	keyText, keyOk := plainIdentifierText(keyNode)
	if !keyOk {
		return false
	}
	nodeText, nodeOk := plainIdentifierText(node)
	if !nodeOk {
		return false
	}
	return keyText == nodeText
}

func plainIdentifierText(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindIdentifier {
		return "", false
	}
	return node.AsIdentifier().Text, true
}

func sameIdentifierText(a, b *ast.Node) bool {
	aText, aOk := plainIdentifierText(a)
	bText, bOk := plainIdentifierText(b)
	return aOk && bOk && aText == bText
}

// isValidImportSpecifier flags only a renamed named import's local binding
// (`import { x as z }` flags `z`, not `x`); an import that keeps the same
// local name as the exported one is exempt, since the developer didn't
// choose that name.
func isValidImportSpecifier(importSpecifier *ast.Node, node *ast.Node) bool {
	is := importSpecifier.AsImportSpecifier()
	if is.Name() != node {
		return false
	}
	if is.PropertyName == nil {
		return false
	}
	importedText, importedOk := moduleExportNameText(is.PropertyName)
	localText, localOk := moduleExportNameText(is.Name())
	if !importedOk || !localOk {
		return false
	}
	return importedText != localText
}

func moduleExportNameText(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text, true
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	}
	return "", false
}

// isValidComputedPropertyNameOwner reports whether a computed key
// (`[x]() {}`, `[x] = 1`) belongs to a class member. Upstream's
// MethodDefinition/PropertyDefinition entries are unconditional (no
// `!computed` guard, unlike Property's non-pattern branch), so computed
// class-member keys are checked; computed object-literal/binding-pattern
// keys are not (matching upstream's explicit `!parent.computed` there).
func isValidComputedPropertyNameOwner(owner *ast.Node) bool {
	if owner == nil {
		return false
	}
	switch owner.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		// Only as a class member: an object literal's computed member is an
		// ESTree Property, whose non-pattern branch requires `!parent.computed`.
		switch owner.Parent.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return isPlainClassMember(owner)
		}
		return false
	case ast.KindPropertyDeclaration:
		return isPlainClassMember(owner)
	}
	return false
}

// isPlainClassMember reports whether a class member maps to ESTree's
// MethodDefinition/PropertyDefinition — the two unconditional entries in
// upstream's dispatch. An `abstract` or `accessor` member does not: it
// becomes a TSAbstractMethodDefinition, TSAbstractPropertyDefinition,
// TSAbstractAccessorProperty or AccessorProperty instead, none of which are
// supported parents, so their names go unchecked.
func isPlainClassMember(member *ast.Node) bool {
	return !ast.HasSyntacticModifier(member, ast.ModifierFlagsAbstract|ast.ModifierFlagsAccessor)
}

// isValidMemberExpressionTarget reports whether a non-computed
// PropertyAccessExpression (`a.b`) is a destructuring/assignment target,
// reproducing upstream's SUPPORTED_EXPRESSIONS.MemberExpression exactly:
//   - Regular assignment: the member expression (through any receiver
//     wrapper) is the left side of an assignment operator (any of `=`,
//     `+=`, `**=`, ...). Both the object and property identifiers of that
//     member expression match, not just the property — see the package doc
//     comment for why this asymmetric-looking behavior is intentional
//     upstream parity, not a bug introduced here.
//   - Destructuring-into-member-expression: the member expression is
//     (through parentheses) exactly the Initializer of a PropertyAssignment
//     whose enclosing object literal is itself the left side of an assignment
//     or the target of a `for-in`/`for-of` — upstream's check reads
//     `parent.parent.parent.parent.left`, and ForInStatement/ForOfStatement
//     carry a `left` just as AssignmentExpression does. This is a single hop
//     only — upstream does not recurse through further nesting for this
//     specific case (`({ a: { b: obj.z } } = {})` does not flag `z`), unlike
//     the plain-identifier Property case which upstream's own ObjectPattern
//     parse-time conversion makes recursive.
func isValidMemberExpressionTarget(pae *ast.Node) bool {
	outer := skipParens(pae)
	gp := outer.Parent
	if gp == nil {
		return false
	}
	switch gp.Kind {
	case ast.KindBinaryExpression:
		be := gp.AsBinaryExpression()
		if !ast.IsAssignmentOperator(be.OperatorToken.Kind) || be.Left != outer {
			return false
		}
		// Inside a destructuring target, `obj.x = 0` is an AssignmentPattern
		// holding a MemberExpression, not an AssignmentExpression, so upstream's
		// `parent.parent.type === "AssignmentExpression"` test fails and
		// `({ key: obj.x = 0 } = source)` reports nothing.
		return !isPatternAssignmentTarget(gp)

	case ast.KindPropertyAssignment:
		pa := gp.AsPropertyAssignment()
		if pa.Initializer != outer {
			return false
		}
		// The object literal must be the assignment target itself: parenthesizing
		// it (`(({ key: obj.x }) = source)`) stops upstream's parser from
		// converting it to an ObjectPattern at all.
		container := gp.Parent
		ggp := container.Parent
		if ggp == nil {
			return false
		}
		if ast.IsForInOrOfStatement(ggp) {
			return ggp.Initializer() == container
		}
		if ggp.Kind != ast.KindBinaryExpression {
			return false
		}
		be := ggp.AsBinaryExpression()
		return ast.IsAssignmentOperator(be.OperatorToken.Kind) && be.Left == container
	}
	return false
}

// isImportAttributeKey ports ESLint's ast-utils.isImportAttributeKey: is node
// used as an import attribute key, either in a static
// `import ... with { key: ... }` / `export ... from ... with { key: ... }`,
// or a dynamic `import(specifier, { with: { key: ... } })`. Import attribute
// keys are syntactic and not names the developer is free to choose, so
// id-length exempts them the same way ESLint does.
func isImportAttributeKey(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// static import/re-export: import ... with { key: value }
	if parent.Kind == ast.KindImportAttribute && parent.AsImportAttribute().Name() == node {
		return true
	}

	// dynamic import: import(specifier, { with: { key: value } })
	//
	// Methods and accessors are Property nodes in ESTree just like `key:
	// value` pairs, so their key counts here too.
	switch parent.Kind {
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
	default:
		return false
	}
	if parent.Name() != node {
		return false
	}
	objectExpression := parent.Parent
	if objectExpression == nil || objectExpression.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	// ESLint's AST has no parentheses, so an options object written as
	// `import("m", ({ with: ... }))` or `{ with: ({ ... }) }` still reaches the
	// call / the outer property directly.
	outer := skipParens(objectExpression)
	objectExpressionParent := outer.Parent
	if objectExpressionParent == nil {
		return false
	}

	if objectExpressionParent.Kind == ast.KindCallExpression {
		call := objectExpressionParent.AsCallExpression()
		if call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword &&
			call.Arguments != nil && len(call.Arguments.Nodes) > 1 && call.Arguments.Nodes[1] == outer {
			return true
		}
	}

	// nested key: import(specifier, { with: { key: value } }) — recurse on
	// the outer "with" key once we've confirmed this object is that
	// property's value.
	if objectExpressionParent.Kind == ast.KindPropertyAssignment {
		outerPa := objectExpressionParent.AsPropertyAssignment()
		if outerPa.Initializer == outer {
			return isImportAttributeKey(objectExpressionParent.Name())
		}
	}

	return false
}

// isCheckedParameterOwner reports whether a plain (non-rest, non-defaulted)
// parameter's owning function-like node is one of upstream's supported
// parents. TS's signature-only forms — function/constructor types, call,
// construct, method and index signatures, ambient or overload declarations,
// abstract methods — all become TS-specific ESTree nodes that
// SUPPORTED_EXPRESSIONS does not list, and a body-less function-like is
// exactly how tsgo spells every one of them that is not already a distinct
// kind.
func isCheckedParameterOwner(owner *ast.Node) bool {
	if owner == nil {
		return false
	}
	switch owner.Kind {
	case ast.KindArrowFunction:
		return true
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor:
		return owner.Body() != nil
	}
	return false
}

// declarationNameRange returns the span upstream reports for an identifier.
// typescript-eslint hangs a variable declarator's or a parameter's type
// annotation — and a parameter's `?` — off the name Identifier itself and
// widens that node to cover it (converter.ts's fixParentLocation), so
// `function fn(x: number) {}` is reported over `x: number` rather than over
// `x` alone. A rest parameter carries the annotation on its RestElement
// wrapper and a destructured one on the pattern, so neither widens an
// identifier; the same goes for a class field, whose annotation stays on the
// PropertyDefinition.
func declarationNameRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	textRange := utils.TrimNodeTextRange(sourceFile, node)
	parent := node.Parent
	if parent == nil || parent.Name() != node {
		return textRange
	}

	var last *ast.Node
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		// A definite-assignment `!` only ever appears alongside a type
		// annotation (`let x!: number`), which already covers it.
		last = parent.AsVariableDeclaration().Type
	case ast.KindParameter:
		param := parent.AsParameterDeclaration()
		if param.DotDotDotToken != nil {
			return textRange
		}
		last = param.Type
		if last == nil {
			last = param.QuestionToken
		}
	}
	if last == nil {
		return textRange
	}
	return textRange.WithEnd(utils.TrimNodeTextRange(sourceFile, last).End())
}

// constructorKeywordRange returns the range of the `constructor` token of a
// ConstructorDeclaration — the span ESLint reports, since its ESTree
// MethodDefinition carries an Identifier key there. Scanning forward from the
// declaration skips any modifiers (`public constructor() {}`) without having
// to reason about their order.
func constructorKeywordRange(node *ast.Node, sourceFile *ast.SourceFile) (core.TextRange, bool) {
	s := scanner.GetScannerForSourceFile(sourceFile, node.Pos())
	for {
		switch s.Token() {
		// A unicode-escaped spelling of the keyword still declares the
		// constructor, so accept whatever name token precedes the parameter
		// list rather than the keyword alone.
		case ast.KindConstructorKeyword, ast.KindIdentifier:
			return core.NewTextRange(s.TokenStart(), s.TokenEnd()), true
		case ast.KindOpenParenToken, ast.KindLessThanToken, ast.KindEndOfFile:
			return core.TextRange{}, false
		}
		s.Scan()
	}
}
