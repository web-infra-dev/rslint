package accessor_pairs

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules/accessorutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed accessor_pairs.schema.json
var schemaJSON []byte

type Options struct {
	GetWithoutSet          bool
	SetWithoutGet          bool
	EnforceForClassMembers bool
	EnforceForTSTypes      bool
}

func parseOptions(options []any) Options {
	opts := Options{
		GetWithoutSet:          false,
		SetWithoutGet:          true,
		EnforceForClassMembers: true,
		EnforceForTSTypes:      false,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
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
	return opts
}

type containerKind int

const (
	containerObjectLiteral containerKind = iota
	containerClass
	containerType
)

// accessorGroup keeps the accessors belonging to one shared key. The key
// helper preserves ESLint's distinct static, private, and dynamic classes.
type accessorGroup struct {
	key      accessorutil.Key
	isStatic bool
	getters  []*ast.Node
	setters  []*ast.Node
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
		key := accessorutil.MakeKey(m)
		isStatic := distinguishStatic && ast.IsStatic(m)
		var group *accessorGroup
		for _, g := range groups {
			if g.isStatic == isStatic && accessorutil.KeysEqual(ctx.SourceFile, g.key, key) {
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
	Name:   "accessor-pairs",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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
