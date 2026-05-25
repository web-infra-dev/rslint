package accessor_pairs

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type Options struct {
	GetWithoutSet          bool
	SetWithoutGet          bool
	EnforceForClassMembers bool
	EnforceForTSTypes      bool
}

func parseOptions(options any) Options {
	opts := Options{
		GetWithoutSet:          false,
		SetWithoutGet:          true,
		EnforceForClassMembers: true,
		EnforceForTSTypes:      false,
	}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["getWithoutSet"].(bool); ok {
			opts.GetWithoutSet = v
		}
		if v, ok := optsMap["setWithoutGet"].(bool); ok {
			opts.SetWithoutGet = v
		}
		if v, ok := optsMap["enforceForClassMembers"].(bool); ok {
			opts.EnforceForClassMembers = v
		}
		if v, ok := optsMap["enforceForTSTypes"].(bool); ok {
			opts.EnforceForTSTypes = v
		}
	}
	return opts
}

type containerKind int

const (
	containerObjectLiteral containerKind = iota
	containerClass
	containerType
)

// keyKind distinguishes how an accessor's key is compared. A static key never
// matches a private or dynamic key even when their textual forms coincide —
// mirroring ESLint's three-way split between `getStaticPropertyName`, the
// single-token list of a `PrivateIdentifier`, and the token-list comparison
// of an arbitrary computed expression.
type keyKind int

const (
	keyStatic  keyKind = iota // identifier / string / numeric / bigint / computed-with-literal
	keyPrivate                // `#name` — distinct equivalence class from string `'#name'`
	keyDynamic                // non-static computed expression — compared structurally
)

type accessorKey struct {
	kind keyKind
	// For keyStatic: the normalized name (e.g. `"100"` for `1e2`, `"a"` for `'a'`).
	// For keyPrivate: the private-identifier text (already includes `#`).
	// Unused for keyDynamic.
	text string
	// For keyDynamic: the computed expression, already unwrapped with
	// [ast.SkipParentheses]. nil for keyStatic / keyPrivate.
	expr *ast.Node
}

type accessorGroup struct {
	key      accessorKey
	isStatic bool
	getters  []*ast.Node
	setters  []*ast.Node
}

// makeKey derives a comparable key for an accessor.
//   - Static-resolvable names (identifier, string / numeric / bigint literal,
//     or a computed expression whose value is one of those literals) collapse
//     into `keyStatic` and are compared by their normalized text — so
//     `a`, `'a'`, `['a']`, and `` [`a`] `` all group together, and `1e2`
//     matches `100` / `'100'` / `['100']`.
//   - `PrivateIdentifier` keys (`#a`) form their own `keyPrivate` class,
//     separate from a string key `'#a'`.
//   - Everything else is `keyDynamic`, compared structurally via
//     [utils.AreNodesStructurallyEqual] — so `[a + b]` and `[a+b]` match
//     (whitespace is not part of the AST) while `[a + b]` and `[a - b]` do
//     not, and `[a.b]` and `[a?.b]` do not.
func makeKey(node *ast.Node) accessorKey {
	nameNode := node.Name()
	if nameNode == nil {
		return accessorKey{kind: keyDynamic}
	}
	if nameNode.Kind == ast.KindPrivateIdentifier {
		return accessorKey{kind: keyPrivate, text: nameNode.AsPrivateIdentifier().Text}
	}
	if name, ok := utils.GetStaticPropertyName(nameNode); ok {
		return accessorKey{kind: keyStatic, text: name}
	}
	expr := nameNode
	if nameNode.Kind == ast.KindComputedPropertyName {
		expr = nameNode.AsComputedPropertyName().Expression
	}
	return accessorKey{kind: keyDynamic, expr: ast.SkipParentheses(expr)}
}

func keysEqual(sf *ast.SourceFile, a, b accessorKey) bool {
	if a.kind != b.kind {
		return false
	}
	switch a.kind {
	case keyStatic, keyPrivate:
		return a.text == b.text
	case keyDynamic:
		return computedKeysEqual(sf, a.expr, b.expr)
	}
	return false
}

// computedKeysEqual reports whether two computed-key expressions are
// structurally equivalent, mirroring ESLint's token-level comparison
// (`sourceCode.getTokens(node.key)` with `areEqualTokenLists`). Unlike
// the general-purpose [utils.AreNodesStructurallyEqual] — which treats
// `0x1` and `1` as the same number — this walker goes through
// [scanner.GetSourceTextOfNodeFromSourceFile] for numeric and bigint
// literals so their raw source form is compared, keeping parity with
// ESLint even when tsgo's parser normalizes literal texts.
//
// Parentheses are transparent at every level (matching ESLint, whose
// ESTree parser elides parens entirely). All other kinds compare by
// their Kind plus, for leaves, the text stored on the AST node; for
// composite nodes, by the ordered list of non-nil children returned
// from [ast.Node.ForEachChild].
func computedKeysEqual(sf *ast.SourceFile, a, b *ast.Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	a = ast.SkipParentheses(a)
	b = ast.SkipParentheses(b)
	if a == nil || b == nil {
		return a == b
	}
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case ast.KindIdentifier:
		return a.AsIdentifier().Text == b.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return a.AsPrivateIdentifier().Text == b.AsPrivateIdentifier().Text
	case ast.KindStringLiteral:
		return a.AsStringLiteral().Text == b.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return a.AsNoSubstitutionTemplateLiteral().Text == b.AsNoSubstitutionTemplateLiteral().Text
	case ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail,
		ast.KindRegularExpressionLiteral:
		return a.Text() == b.Text()
	case ast.KindNumericLiteral, ast.KindBigIntLiteral:
		return scanner.GetSourceTextOfNodeFromSourceFile(sf, a, false) ==
			scanner.GetSourceTextOfNodeFromSourceFile(sf, b, false)
	}
	var aKids, bKids []*ast.Node
	a.ForEachChild(func(c *ast.Node) bool {
		aKids = append(aKids, c)
		return false
	})
	b.ForEachChild(func(c *ast.Node) bool {
		bKids = append(bKids, c)
		return false
	})
	if len(aKids) != len(bKids) {
		return false
	}
	for i := range aKids {
		if !computedKeysEqual(sf, aKids[i], bKids[i]) {
			return false
		}
	}
	return true
}

// checkList examines accessors in a list and reports any that lack a pair.
// For class containers, `distinguishStatic` should be true — static and
// instance members with the same name are independent pairs, and grouping
// by (name, isStatic) preserves source order of the final reports.
func checkList(ctx rule.RuleContext, members []*ast.Node, opts Options, kind containerKind, distinguishStatic bool) {
	var groups []*accessorGroup

	for _, m := range members {
		isGetter := m.Kind == ast.KindGetAccessor
		isSetter := m.Kind == ast.KindSetAccessor
		if !isGetter && !isSetter {
			continue
		}
		key := makeKey(m)
		isStatic := distinguishStatic && ast.IsStatic(m)
		var group *accessorGroup
		for _, g := range groups {
			if g.isStatic == isStatic && keysEqual(ctx.SourceFile, g.key, key) {
				group = g
				break
			}
		}
		if group == nil {
			group = &accessorGroup{key: key, isStatic: isStatic}
			groups = append(groups, group)
		}
		if isGetter {
			group.getters = append(group.getters, m)
		} else {
			group.setters = append(group.setters, m)
		}
	}

	for _, g := range groups {
		if opts.SetWithoutGet && len(g.setters) > 0 && len(g.getters) == 0 {
			for _, s := range g.setters {
				reportAccessor(ctx, s, kind, "Getter")
			}
		}
		if opts.GetWithoutSet && len(g.getters) > 0 && len(g.setters) == 0 {
			for _, g2 := range g.getters {
				reportAccessor(ctx, g2, kind, "Setter")
			}
		}
	}
}

// reportAccessor emits a diagnostic for an accessor that is missing its pair.
// missingKind is "Getter" or "Setter" (the accessor that is absent); node is
// the existing accessor whose counterpart is missing.
func reportAccessor(ctx rule.RuleContext, node *ast.Node, container containerKind, missingKind string) {
	existingKind := "setter"
	if node.Kind == ast.KindGetAccessor {
		existingKind = "getter"
	}

	var prefix string
	switch container {
	case containerClass:
		prefix = "class "
		if ast.IsStatic(node) {
			prefix += "static "
		}
		nameNode := node.Name()
		if nameNode != nil && nameNode.Kind == ast.KindPrivateIdentifier {
			prefix += "private "
		}
	case containerType:
		prefix = "type "
	}

	nameNode := node.Name()
	var namePart string
	if nameNode != nil {
		if nameNode.Kind == ast.KindPrivateIdentifier {
			// PrivateIdentifier.Text already includes the leading '#'; no quotes.
			namePart = " " + nameNode.AsPrivateIdentifier().Text
		} else if name, ok := utils.GetStaticPropertyName(nameNode); ok {
			namePart = fmt.Sprintf(" '%s'", name)
		}
	}

	var msgIdSuffix string
	switch container {
	case containerObjectLiteral:
		msgIdSuffix = "ObjectLiteral"
	case containerClass:
		msgIdSuffix = "Class"
	case containerType:
		msgIdSuffix = "Type"
	}

	ctx.ReportRange(
		utils.GetFunctionHeadLoc(ctx.SourceFile, node),
		rule.RuleMessage{
			Id:          fmt.Sprintf("missing%sIn%s", missingKind, msgIdSuffix),
			Description: fmt.Sprintf("%s is not present for %s%s%s.", missingKind, prefix, existingKind, namePart),
		},
	)
}

// reportPropertyDescriptor emits a diagnostic for a property descriptor
// object literal that declares only one of `get`/`set`.
func reportPropertyDescriptor(ctx rule.RuleContext, node *ast.Node, missingKind string) {
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          fmt.Sprintf("missing%sInPropertyDescriptor", missingKind),
		Description: missingKind + " is not present in property descriptor.",
	})
}

// isPropertyDescriptor reports whether an ObjectLiteralExpression sits in a
// position that makes it a property descriptor.
//
// Recognized shapes:
//   - `Object.defineProperty(obj, key, <here>)`
//   - `Reflect.defineProperty(obj, key, <here>)`
//   - `Object.defineProperties(obj, { foo: <here> })`
//   - `Object.create(proto,           { foo: <here> })`
func isPropertyDescriptor(node *ast.Node) bool {
	if utils.IsArgumentOfSpecificCall(node, 2, "Object", "defineProperty") ||
		utils.IsArgumentOfSpecificCall(node, 2, "Reflect", "defineProperty") {
		return true
	}
	// Inner `{get/set: ...}` of a descriptor map: walk up to the outer
	// ObjectLiteralExpression and check if IT is the arg[1] of create /
	// defineProperties.
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindPropertyAssignment {
		return false
	}
	grandparent := parent.Parent
	if grandparent == nil || grandparent.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	return utils.IsArgumentOfSpecificCall(grandparent, 1, "Object", "create") ||
		utils.IsArgumentOfSpecificCall(grandparent, 1, "Object", "defineProperties")
}

// checkObjectLiteral collects accessor properties and checks pairs.
func checkObjectLiteral(ctx rule.RuleContext, node *ast.Node, opts Options) {
	obj := node.AsObjectLiteralExpression()
	if obj == nil || obj.Properties == nil {
		return
	}
	checkList(ctx, obj.Properties.Nodes, opts, containerObjectLiteral, false)
}

// checkDescriptorObject reports when a property descriptor declares only
// `get` or only `set`. Descriptors use regular value-position properties
// (NOT getter / setter syntax), so accessor-kind entries are ignored. All
// three "init" shapes with a non-computed identifier key contribute:
//   - `{ set: fn }`        — PropertyAssignment
//   - `{ set(v) {} }`      — MethodDeclaration (method shorthand)
//   - `{ set }`            — ShorthandPropertyAssignment
//
// `{ "set": fn }` and `{ [set]: fn }` are NOT recognized (matches ESLint:
// the string-keyed and computed forms require `!p.computed && key.name`).
func checkDescriptorObject(ctx rule.RuleContext, node *ast.Node, opts Options) {
	obj := node.AsObjectLiteralExpression()
	if obj == nil || obj.Properties == nil {
		return
	}
	names := map[string]bool{}
	for _, p := range obj.Properties.Nodes {
		switch p.Kind {
		case ast.KindPropertyAssignment,
			ast.KindMethodDeclaration,
			ast.KindShorthandPropertyAssignment:
			// fall through
		default:
			continue
		}
		nameNode := p.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			continue
		}
		names[nameNode.AsIdentifier().Text] = true
	}
	hasGet := names["get"]
	hasSet := names["set"]
	if opts.SetWithoutGet && hasSet && !hasGet {
		reportPropertyDescriptor(ctx, node, "Getter")
	}
	if opts.GetWithoutSet && hasGet && !hasSet {
		reportPropertyDescriptor(ctx, node, "Setter")
	}
}

// checkClassBody examines class members. Static and instance members with
// the same name form independent pairs, but we iterate in source order and
// distinguish groups by (name, isStatic) so reports land in source order.
func checkClassBody(ctx rule.RuleContext, node *ast.Node, opts Options) {
	checkList(ctx, node.Members(), opts, containerClass, true)
}

// checkTypeMembers handles InterfaceDeclaration and TypeLiteral members.
func checkTypeMembers(ctx rule.RuleContext, members []*ast.Node, opts Options) {
	checkList(ctx, members, opts, containerType, false)
}

// https://eslint.org/docs/latest/rules/accessor-pairs
var AccessorPairsRule = rule.Rule{
	Name: "accessor-pairs",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		if !opts.SetWithoutGet && !opts.GetWithoutSet {
			return rule.RuleListeners{}
		}

		listeners := rule.RuleListeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				checkObjectLiteral(ctx, node, opts)
				if isPropertyDescriptor(node) {
					checkDescriptorObject(ctx, node, opts)
				}
			},
		}

		if opts.EnforceForClassMembers {
			listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
				checkClassBody(ctx, node, opts)
			}
			listeners[ast.KindClassExpression] = func(node *ast.Node) {
				checkClassBody(ctx, node, opts)
			}
		}

		if opts.EnforceForTSTypes {
			listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
				checkTypeMembers(ctx, node.Members(), opts)
			}
			listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
				checkTypeMembers(ctx, node.Members(), opts)
			}
		}

		return listeners
	},
}
