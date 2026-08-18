package id_length

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
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
//     ArrowFunctionExpression. ArrowFunctionExpression never appears in the
//     switch below: arrows can't be named, so upstream's entry for it only
//     ever matched a plain param, which is already covered by KindParameter.
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
//     needs a structural check — ast.IsAssignmentTarget — mirroring
//     upstream's own parse-time pattern-conversion grammar (recursive through
//     parens/nested literals/spreads), for BOTH plain-identifier members
//     (Property's ObjectPattern branch) and rest members (RestElement).
//   - A computed key (`[x]() {}`, `{ [x]: 1 }`) is a single node with a
//     `.computed` flag in ESTree; tsgo wraps the key expression in a
//     dedicated ComputedPropertyName node one level below the member. This
//     rule dispatches on that wrapper's own parent to decide inclusion,
//     rather than reading a computed flag.
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
//     unconditionally, reproducing the quirk. ClassDeclaration's superclass
//     expression, by contrast, sits under tsgo's HeritageClauseList /
//     ExpressionWithTypeArguments wrapper rather than as a direct child, so
//     `class C extends B {}` does not flag `B` here even though upstream's
//     ESTree-flat shape does — see id_length.md's "Differences from ESLint".
var IdLengthRule = rule.Rule{
	Name:   "id-length",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		check := func(node *ast.Node) {
			name, isPrivate := identifierName(node)
			nameLength := utils.GraphemeCount(name)

			isShort := nameLength < opts.min
			isLong := opts.hasMax && nameLength > opts.max
			if !isShort && !isLong {
				return
			}
			if opts.matchesException(name) {
				return
			}

			if !isValidPosition(opts, node) {
				return
			}

			if isShort {
				ctx.ReportNode(node, messageTooShort(name, opts.min, isPrivate))
			} else {
				ctx.ReportNode(node, messageTooLong(name, opts.max, isPrivate))
			}
		}

		return rule.RuleListeners{
			ast.KindIdentifier:        check,
			ast.KindPrivateIdentifier: check,
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
	// ESTree drops parentheses (and has no TS wrapper expressions) entirely,
	// so every SUPPORTED_EXPRESSIONS check upstream is implicitly transparent
	// to them. tsgo keeps them as real nodes, so climb through any receiver
	// wrapper around the identifier itself before dispatching on its
	// structural parent. `wrapped` (rather than the raw node) is what gets
	// compared against sibling expression-position fields below (Initializer,
	// Expression) — declaration-name positions (Name()) can never themselves
	// be wrapped, so using `wrapped` there too is equivalent, not just safe.
	wrapped := skipReceiverWrappers(node)
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
		return isValidBindingElement(opts, effectiveParent, wrapped)

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
		pa := effectiveParent.AsPropertyAssignment()
		container := effectiveParent.Parent
		if ast.IsAssignmentTarget(container) {
			return isValidPatternPropertyValue(opts, effectiveParent.Name(), pa.Initializer, wrapped)
		}
		return isValidPlainProperty(opts, effectiveParent.Name(), wrapped)

	case ast.KindShorthandPropertyAssignment:
		// Upstream's ObjectPattern and non-pattern branches converge on an
		// identical outcome for shorthand `{ a }`: key and value are the same
		// node, so both branches reduce to "properties enabled and this is
		// that node". See id_length.md analysis in the package doc.
		return opts.properties && effectiveParent.Name() == wrapped && !isImportAttributeKey(node)

	case ast.KindImportSpecifier:
		return isValidImportSpecifier(effectiveParent, wrapped)

	case ast.KindImportClause:
		return effectiveParent.Name() == wrapped

	case ast.KindNamespaceImport:
		return effectiveParent.Name() == wrapped

	case ast.KindParameter:
		return effectiveParent.Name() == wrapped

	case ast.KindClassDeclaration:
		return effectiveParent.Name() == wrapped

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		return effectiveParent.Name() == wrapped

	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return effectiveParent.Name() == wrapped

	case ast.KindPropertyDeclaration:
		pd := effectiveParent.AsPropertyDeclaration()
		return wrapped == effectiveParent.Name() || wrapped == pd.Initializer

	case ast.KindComputedPropertyName:
		return isValidComputedPropertyNameOwner(effectiveParent.Parent)

	case ast.KindPropertyAccessExpression:
		if !opts.properties {
			return false
		}
		return isValidMemberExpressionTarget(effectiveParent)

	case ast.KindSpreadElement:
		// ast.IsAssignmentTarget is called on wrapped (not effectiveParent):
		// GetAssignmentTarget decides how to continue climbing by switching
		// on node.Parent.Kind, so it must be invoked starting from the
		// identifier itself, not from the SpreadElement/SpreadAssignment
		// wrapper already one hop up. This also matters because
		// GetAssignmentTarget only special-cases ArrayLiteralExpression as a
		// "keep climbing" parent, not ObjectLiteralExpression — it is only
		// reachable by first landing on the PropertyAssignment/
		// ShorthandPropertyAssignment/SpreadAssignment case, which jumps
		// straight past it to its own parent.
		se := effectiveParent.AsSpreadElement()
		return se.Expression == wrapped && ast.IsAssignmentTarget(wrapped)

	case ast.KindSpreadAssignment:
		sa := effectiveParent.AsSpreadAssignment()
		return sa.Expression == wrapped && ast.IsAssignmentTarget(wrapped)
	}

	return false
}

// skipReceiverWrappers walks up from node through any chain of transparent
// wrapper expressions (parentheses, and the TS-only non-null/as/satisfies
// wrappers with no ESTree equivalent) and returns the outermost one — i.e.
// the node whose own .Parent is the first structurally meaningful ancestor.
// When node has no such wrapper immediately above it, it is returned
// unchanged.
func skipReceiverWrappers(node *ast.Node) *ast.Node {
	for node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression, ast.KindSatisfiesExpression:
			node = node.Parent
			continue
		}
		break
	}
	return node
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
		keyNode := be.PropertyName
		valueNode := bindingElement.Name()
		if keyNode == nil {
			keyNode = valueNode
		}
		return isValidPatternPropertyValueNodes(opts, keyNode, valueNode, node)
	case ast.KindArrayBindingPattern:
		return bindingElement.Name() == node
	}
	return false
}

// isValidPatternPropertyValue handles a PropertyAssignment (`key: value`)
// that upstream would see as ObjectPattern's Property branch — i.e. the
// enclosing object literal is (transitively) a destructuring-assignment
// target.
func isValidPatternPropertyValue(opts idLengthOptions, keyNode, valueNode *ast.Node, node *ast.Node) bool {
	return isValidPatternPropertyValueNodes(opts, keyNode, valueNode, node)
}

func isValidPatternPropertyValueNodes(opts idLengthOptions, keyNode, valueNode *ast.Node, node *ast.Node) bool {
	if sameIdentifierText(keyNode, valueNode) {
		return node == keyNode && opts.properties
	}
	return node == valueNode
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
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindPropertyDeclaration:
		return true
	}
	return false
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
//     (through a wrapper) exactly the Initializer of a PropertyAssignment
//     whose enclosing object literal is itself directly the left side of an
//     assignment. This is a single hop only — upstream does not recurse
//     through further nesting for this specific case (`({ a: { b: obj.z } }
//     = {})` does not flag `z`), unlike the plain-identifier Property case
//     which upstream's own ObjectPattern parse-time conversion makes
//     recursive.
func isValidMemberExpressionTarget(pae *ast.Node) bool {
	outer := skipReceiverWrappers(pae)
	gp := outer.Parent
	if gp == nil {
		return false
	}
	switch gp.Kind {
	case ast.KindBinaryExpression:
		be := gp.AsBinaryExpression()
		return ast.IsAssignmentOperator(be.OperatorToken.Kind) && be.Left == outer

	case ast.KindPropertyAssignment:
		pa := gp.AsPropertyAssignment()
		if pa.Initializer != outer {
			return false
		}
		container := skipReceiverWrappers(gp.Parent)
		ggp := container.Parent
		if ggp == nil || ggp.Kind != ast.KindBinaryExpression {
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
	switch parent.Kind {
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
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
	objectExpressionParent := objectExpression.Parent
	if objectExpressionParent == nil {
		return false
	}

	if objectExpressionParent.Kind == ast.KindCallExpression {
		call := objectExpressionParent.AsCallExpression()
		if call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword &&
			call.Arguments != nil && len(call.Arguments.Nodes) > 1 && call.Arguments.Nodes[1] == objectExpression {
			return true
		}
	}

	// nested key: import(specifier, { with: { key: value } }) — recurse on
	// the outer "with" key once we've confirmed this object is that
	// property's value.
	if objectExpressionParent.Kind == ast.KindPropertyAssignment {
		outerPa := objectExpressionParent.AsPropertyAssignment()
		if outerPa.Initializer == objectExpression {
			return isImportAttributeKey(objectExpressionParent.Name())
		}
	}

	return false
}
