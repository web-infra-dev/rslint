package forbid_prop_types

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const msgForbiddenPropType = `Prop type "{{target}}" is forbidden`

var defaultForbid = []string{"any", "array", "object"}

type ruleOptions struct {
	forbid                 []string
	checkContextTypes      bool
	checkChildContextTypes bool
}

// parseOptions reads `forbid`, `checkContextTypes`, and
// `checkChildContextTypes` from the rule options. Mirrors upstream's defaults:
//   - missing options object → forbid defaults to `['any', 'array', 'object']`,
//     checkContextTypes / checkChildContextTypes default to false
//   - explicit `forbid: []` → empty list (rule reports nothing)
//   - non-string entries inside `forbid` are silently dropped, matching
//     upstream's `forbid.indexOf(type) >= 0` membership check (which can't
//     match anything but strings anyway)
func parseOptions(input any) ruleOptions {
	opts := ruleOptions{forbid: defaultForbid}
	optsMap := utils.GetOptionsMap(input)
	if optsMap == nil {
		return opts
	}
	if raw, ok := optsMap["forbid"].([]interface{}); ok {
		list := make([]string, 0, len(raw))
		for _, x := range raw {
			if s, ok := x.(string); ok {
				list = append(list, s)
			}
		}
		opts.forbid = list
	}
	if v, ok := optsMap["checkContextTypes"].(bool); ok {
		opts.checkContextTypes = v
	}
	if v, ok := optsMap["checkChildContextTypes"].(bool); ok {
		opts.checkChildContextTypes = v
	}
	return opts
}

// getPropertyAccessName returns the name identifier text of a
// PropertyAccessExpression (`expr.name` → `"name"`); empty string for any
// other shape.
func getPropertyAccessName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return ""
	}
	pa := node.AsPropertyAccessExpression()
	nameNode := pa.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return ""
	}
	return nameNode.AsIdentifier().Text
}

var ForbidPropTypesRule = rule.Rule{
	Name: "react/forbid-prop-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		propWrappers := reactutil.GetPropWrapperFunctions(ctx.Settings)

		// Per-file import tracking. Mirrors upstream's three module-state
		// variables verbatim — see the upstream `ImportDeclaration` handler
		// for the exact case-by-case logic.
		propTypesPackageName := ""
		reactPackageName := ""
		isForeignPropTypesPackage := false

		// Scope-aware Identifier resolution map: scope-node → name →
		// ObjectLiteralExpression initializer. Mirrors upstream's
		// `variableUtil.findVariableByName`, which walks the surrounding
		// scope chain via ESLint's scope manager and returns the
		// variable's `defs[0].node.init` (the FIRST declaration's
		// initializer — NOT the last assigned value). So:
		//   1. We pre-scan the file for every VariableDeclaration whose
		//      initializer is an object literal (TS / paren wrappers
		//      stripped) and bind it under its enclosing function-like /
		//      class-static-block / source-file scope (via
		//      `utils.FindEnclosingScope`).
		//   2. Same-scope re-declaration: first-write-wins (match
		//      upstream's `defs[0]` — first definition seen by the scope
		//      manager).
		//   3. Plain reassignments (`x = {...}` without `var`/`let`/
		//      `const`) are NOT tracked — upstream's lookup ignores them
		//      because they don't create new defs.
		//   4. At lookup time, we walk from the referencing identifier's
		//      enclosing scope upward — the closest binding wins, matching
		//      ESLint's `scope.upper` chain.
		scopeBindings := map[*ast.Node]map[string]*ast.Node{}

		var preWalk ast.Visitor
		preWalk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			if n.Kind == ast.KindVariableDeclaration {
				vd := n.AsVariableDeclaration()
				if vd != nil && vd.Initializer != nil {
					nameNode := vd.Name()
					if nameNode != nil && nameNode.Kind == ast.KindIdentifier {
						init := reactutil.SkipExpressionWrappers(vd.Initializer)
						if init.Kind == ast.KindObjectLiteralExpression {
							scope := utils.FindEnclosingScope(n)
							if scope == nil {
								scope = ctx.SourceFile.AsNode()
							}
							inner, ok := scopeBindings[scope]
							if !ok {
								inner = map[string]*ast.Node{}
								scopeBindings[scope] = inner
							}
							// First-write-wins: matches upstream's
							// `defs[0]` semantics. Subsequent `var x =
							// {...}` redeclarations within the same
							// scope are ignored.
							name := nameNode.AsIdentifier().Text
							if _, exists := inner[name]; !exists {
								inner[name] = init
							}
						}
					}
				}
			}
			n.ForEachChild(preWalk)
			return false
		}
		ctx.SourceFile.Node.ForEachChild(preWalk)

		// resolveIdentifier walks the scope chain from `ident` upward and
		// returns the closest object-literal binding for `name`, or nil.
		resolveIdentifier := func(ident *ast.Node, name string) *ast.Node {
			scope := utils.FindEnclosingScope(ident)
			for scope != nil {
				if inner, ok := scopeBindings[scope]; ok {
					if init, ok := inner[name]; ok {
						return init
					}
				}
				if scope.Kind == ast.KindSourceFile {
					return nil
				}
				scope = utils.FindEnclosingScope(scope)
			}
			// Fallback: the source-file scope is keyed under the
			// SourceFile node; if we walked past it via the parent chain
			// (scope.Parent == nil for SourceFile), check it explicitly.
			if inner, ok := scopeBindings[ctx.SourceFile.AsNode()]; ok {
				if init, ok := inner[name]; ok {
					return init
				}
			}
			return nil
		}

		isForbidden := func(t string) bool {
			return slices.Contains(opts.forbid, t)
		}

		report := func(reportNode *ast.Node, target string) {
			ctx.ReportNode(reportNode, rule.RuleMessage{
				Id:          "forbiddenPropType",
				Description: strings.ReplaceAll(msgForbiddenPropType, "{{target}}", target),
				Data:        map[string]string{"target": target},
			})
		}

		reportIfForbidden := func(t string, reportNode *ast.Node, target string) {
			if isForbidden(t) {
				report(reportNode, target)
			}
		}

		// isPropTypesPackage mirrors upstream's bare expression-classifier:
		//
		//   (Identifier && (
		//     name == null || name == propTypesPackageName || !isForeign
		//   ))
		//   || (MemberExpression && (
		//     object.name == null || object.name == reactPackageName || !isForeign
		//   ))
		//
		// `name == null` lines up with ESTree shapes where `.name` is not
		// always set on a sub-expression (e.g. a `MemberExpression` whose
		// `.object` is itself a `MemberExpression` has no `.name`). The
		// `!isForeign` arm is the catch-all that keeps the rule on by default
		// when no `import { PropTypes } from "<other>"` was seen — this is
		// what makes the rule report on bare `PropTypes.array` even when no
		// import was found at all.
		isPropTypesPackage := func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			switch node.Kind {
			case ast.KindIdentifier:
				name := node.AsIdentifier().Text
				return name == propTypesPackageName || !isForeignPropTypesPackage
			case ast.KindPropertyAccessExpression:
				pa := node.AsPropertyAccessExpression()
				obj := reactutil.SkipExpressionWrappers(pa.Expression)
				var objName string
				if obj.Kind == ast.KindIdentifier {
					objName = obj.AsIdentifier().Text
				}
				// objName == "" stands in for upstream's
				// `node.object.name === null` — both apply when the
				// receiver isn't a bare Identifier (e.g. nested member
				// chain `React.PropTypes.<foo>` whose `.object` is itself a
				// MemberExpression with no `.name`).
				return objName == "" || objName == reactPackageName || !isForeignPropTypesPackage
			}
			return false
		}

		// shouldCheckContextTypes / shouldCheckChildContextTypes mirror
		// upstream gates: only fire when the corresponding option is on AND
		// the property name matches.
		shouldCheckContextTypes := func(name string) bool {
			return opts.checkContextTypes && name == "contextTypes"
		}
		shouldCheckChildContextTypes := func(name string) bool {
			return opts.checkChildContextTypes && name == "childContextTypes"
		}

		// isRelevantName reports whether a property/member name should
		// trigger checking. Matches upstream's combined gate:
		//   isPropTypesDeclaration(node)
		//     || shouldCheckContextTypes(node)
		//     || shouldCheckChildContextTypes(node)
		// where isPropTypesPackage(node) is irrelevant for property nodes
		// (always false; see comment in upstream-listener `ObjectExpression`
		// section below).
		isRelevantName := func(name string) bool {
			return name == "propTypes" || shouldCheckContextTypes(name) || shouldCheckChildContextTypes(name)
		}

		var checkNode func(node *ast.Node)

		// checkProperties is the inner check applied to an object literal's
		// properties. Each entry is matched against the propTypes value
		// shape: identifier, member access, or PropTypes-call (with args).
		checkProperties := func(props []*ast.Node) {
			for _, prop := range props {
				if prop.Kind != ast.KindPropertyAssignment && prop.Kind != ast.KindShorthandPropertyAssignment {
					// SpreadAssignment, methods, accessors, etc. are
					// silently skipped — upstream's `if (declaration.type
					// !== 'Property') return;` does the same.
					continue
				}

				var value *ast.Node
				if prop.Kind == ast.KindPropertyAssignment {
					pa := prop.AsPropertyAssignment()
					if pa.Initializer == nil {
						continue
					}
					value = reactutil.SkipExpressionWrappers(pa.Initializer)
				} else {
					// Shorthand `{ object }` is parsed as a
					// ShorthandPropertyAssignment; ESLint represents it as
					// `Property` with the same `Identifier` for key and
					// value. Mirror that — value is the name identifier.
					value = prop.Name()
					if value == nil {
						continue
					}
				}

				// Strip a trailing `.isRequired`. Mirrors upstream's
				//   if (value.type === 'MemberExpression' &&
				//       value.property && value.property.name &&
				//       value.property.name === 'isRequired')
				//     value = value.object;
				if value.Kind == ast.KindPropertyAccessExpression {
					if getPropertyAccessName(value) == "isRequired" {
						value = reactutil.SkipExpressionWrappers(value.AsPropertyAccessExpression().Expression)
					}
				}

				// CallExpression: check arguments as named PropTypes (e.g.
				// `oneOf([...])` doesn't map here, but
				// `arrayOf(PropTypes.object)` reports `object` via the
				// argument-walk below). Then unwrap to the callee.
				if value.Kind == ast.KindCallExpression {
					call := value.AsCallExpression()
					callee := reactutil.SkipExpressionWrappers(call.Expression)
					if !isPropTypesPackage(callee) {
						continue
					}
					if call.Arguments != nil {
						for _, arg := range call.Arguments.Nodes {
							argU := reactutil.SkipExpressionWrappers(arg)
							var name string
							switch argU.Kind {
							case ast.KindPropertyAccessExpression:
								name = getPropertyAccessName(argU)
							case ast.KindIdentifier:
								name = argU.AsIdentifier().Text
							}
							if name != "" {
								reportIfForbidden(name, prop, name)
							}
						}
					}
					value = callee
				}

				if !isPropTypesPackage(value) {
					continue
				}
				var target string
				switch value.Kind {
				case ast.KindPropertyAccessExpression:
					target = getPropertyAccessName(value)
				case ast.KindIdentifier:
					target = value.AsIdentifier().Text
				}
				if target != "" {
					reportIfForbidden(target, prop, target)
				}
			}
		}

		checkNode = func(node *ast.Node) {
			if node == nil {
				return
			}
			node = reactutil.SkipExpressionWrappers(node)
			switch node.Kind {
			case ast.KindObjectLiteralExpression:
				checkProperties(node.AsObjectLiteralExpression().Properties.Nodes)
			case ast.KindIdentifier:
				name := node.AsIdentifier().Text
				if obj := resolveIdentifier(node, name); obj != nil {
					checkProperties(obj.AsObjectLiteralExpression().Properties.Nodes)
				}
			case ast.KindCallExpression:
				if !reactutil.IsPropWrapperCall(node, propWrappers) {
					return
				}
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				checkNode(call.Arguments.Nodes[0])
			}
		}

		// findLastReturn mirrors upstream's `loopNodes` (ast.js):
		// iterates statements from the end; first ReturnStatement wins;
		// if it encounters a SwitchStatement, recurses into the LAST
		// case clause's consequent. Returns the matching ReturnStatement
		// or nil.
		var findLastReturn func(stmts []*ast.Node) *ast.Node
		findLastReturn = func(stmts []*ast.Node) *ast.Node {
			for i := len(stmts) - 1; i >= 0; i-- {
				s := stmts[i]
				if s.Kind == ast.KindReturnStatement {
					return s
				}
				if s.Kind == ast.KindSwitchStatement {
					sw := s.AsSwitchStatement()
					if sw == nil || sw.CaseBlock == nil {
						continue
					}
					cb := sw.CaseBlock.AsCaseBlock()
					if cb == nil || cb.Clauses == nil || len(cb.Clauses.Nodes) == 0 {
						continue
					}
					last := cb.Clauses.Nodes[len(cb.Clauses.Nodes)-1]
					cc := last.AsCaseOrDefaultClause()
					if cc == nil || cc.Statements == nil {
						continue
					}
					if r := findLastReturn(cc.Statements.Nodes); r != nil {
						return r
					}
				}
			}
			return nil
		}

		// handleMethodLikeReturn locates the last top-level ReturnStatement
		// in the method/getter body (recursing into trailing SwitchStatement
		// case bodies) and recurses on its argument. Mirrors upstream's
		// `findReturnStatement` + `loopNodes`.
		handleMethodLikeReturn := func(name string, body *ast.Node) {
			if !isRelevantName(name) || body == nil || body.Kind != ast.KindBlock {
				return
			}
			block := body.AsBlock()
			if block == nil || block.Statements == nil {
				return
			}
			ret := findLastReturn(block.Statements.Nodes)
			if ret == nil {
				return
			}
			rs := ret.AsReturnStatement()
			if rs != nil && rs.Expression != nil {
				checkNode(rs.Expression)
			}
		}

		return rule.RuleListeners{
			// Track package bindings. Mirrors upstream's three branches:
			//   - source === 'prop-types' → propTypesPackageName = first
			//     specifier's local name
			//   - source === 'react' → reactPackageName = first specifier's
			//     local name; if any named specifier imports `PropTypes`,
			//     propTypesPackageName = that local name (i.e. the `as`-alias)
			//   - else (foreign module) → if any specifier's local name is
			//     literally `PropTypes`, mark isForeignPropTypesPackage
			ast.KindImportDeclaration: func(node *ast.Node) {
				decl := node.AsImportDeclaration()
				if decl == nil || decl.ModuleSpecifier == nil || decl.ModuleSpecifier.Kind != ast.KindStringLiteral {
					return
				}
				source := decl.ModuleSpecifier.AsStringLiteral().Text
				clause := decl.ImportClause
				switch source {
				case "prop-types":
					if name := firstSpecifierLocalName(clause); name != "" {
						propTypesPackageName = name
					}
				case "react":
					if name := firstSpecifierLocalName(clause); name != "" {
						reactPackageName = name
					}
					// Search named specifiers for one whose imported name
					// is `PropTypes` and capture its (possibly aliased)
					// local name.
					if local := findNamedImportLocal(clause, "PropTypes"); local != "" {
						propTypesPackageName = local
					}
				default:
					if hasLocalName(clause, "PropTypes") {
						isForeignPropTypesPackage = true
					}
				}
			},

			// Class fields: `static propTypes = {...}`,
			// `static contextTypes = {...}`, etc. Mirrors upstream's
			// `'ClassProperty, PropertyDefinition'` listener.
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				pd := node.AsPropertyDeclaration()
				if pd == nil {
					return
				}
				keyNode := pd.Name()
				if keyNode == nil {
					return
				}
				keyName, ok := utils.GetStaticPropertyName(keyNode)
				if !ok || !isRelevantName(keyName) {
					return
				}
				if pd.Initializer != nil {
					checkNode(pd.Initializer)
				}
			},

			// Assignment / mutation of `<expr>.propTypes` etc. Mirrors
			// upstream's `MemberExpression` listener: fires on every
			// MemberExpression whose name matches, then `checkNode('right'
			// in node.parent && node.parent.right)`. We collapse the
			// parent walk into a BinaryExpression listener — equivalent
			// because the `'right' in parent` test only succeeds for
			// BinaryExpression (assignment + compound + comparison +
			// arithmetic). No operator filter: upstream applies to `=`,
			// compound assigns (`||=`, `&&=`, `??=`, `+=`, …), and even
			// comparison-shape parents (`Foo.propTypes == X` is rare but
			// upstream walks parent.right anyway). Compound / comparison
			// cases rarely trigger real reports because checkNode falls
			// through on most non-ObjectLiteral right-hand sides; the
			// broad gate is what byte-for-byte alignment requires.
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin == nil {
					return
				}
				left := reactutil.SkipExpressionWrappers(bin.Left)
				if left.Kind != ast.KindPropertyAccessExpression {
					return
				}
				name := getPropertyAccessName(left)
				if !isRelevantName(name) {
					return
				}
				checkNode(bin.Right)
			},

			// Class methods and getters with body returning the propTypes
			// object: `static propTypes() { return ... }` or
			// `static get propTypes() { return ... }`.
			ast.KindMethodDeclaration: func(node *ast.Node) {
				md := node.AsMethodDeclaration()
				if md == nil || md.Name() == nil {
					return
				}
				name, ok := utils.GetStaticPropertyName(md.Name())
				if !ok {
					return
				}
				handleMethodLikeReturn(name, md.Body)
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				ga := node.AsGetAccessorDeclaration()
				if ga == nil || ga.Name() == nil {
					return
				}
				name, ok := utils.GetStaticPropertyName(ga.Name())
				if !ok {
					return
				}
				handleMethodLikeReturn(name, ga.Body)
			},

			// CallExpression: `PropTypes.shape({...})` or
			// `shape({...})` — recurse into the argument's properties.
			// Mirrors upstream's CallExpression listener literally,
			// including the gate that returns early when the callee is a
			// MemberExpression NOT receiving from a propTypes package and
			// NOT named `propTypes` itself.
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil {
					return
				}
				callee := reactutil.SkipExpressionWrappers(call.Expression)

				// Upstream gate:
				//   if (callee.type === 'MemberExpression'
				//       && callee.object
				//       && !isPropTypesPackage(callee.object)
				//       && !isPropTypesDeclaration(callee)) return;
				if callee.Kind == ast.KindPropertyAccessExpression {
					pa := callee.AsPropertyAccessExpression()
					obj := reactutil.SkipExpressionWrappers(pa.Expression)
					calleeName := getPropertyAccessName(callee)
					if obj != nil && !isPropTypesPackage(obj) && !isRelevantName(calleeName) {
						return
					}
				}

				// Upstream second condition:
				//   arguments.length > 0 && (
				//     ('name' in callee && callee.name === 'shape')
				//     || getPropertyName(callee) === 'shape'
				//   )
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				isShape := false
				switch callee.Kind {
				case ast.KindIdentifier:
					isShape = callee.AsIdentifier().Text == "shape"
				case ast.KindPropertyAccessExpression:
					isShape = getPropertyAccessName(callee) == "shape"
				}
				if !isShape {
					return
				}
				arg0 := reactutil.SkipExpressionWrappers(call.Arguments.Nodes[0])
				if arg0.Kind != ast.KindObjectLiteralExpression {
					return
				}
				checkProperties(arg0.AsObjectLiteralExpression().Properties.Nodes)
			},

			// `createReactClass({propTypes: {...}, ...})` — fires on every
			// ObjectLiteralExpression. Mirrors upstream's ObjectExpression
			// listener: for each property, check if the key is propTypes /
			// contextTypes / childContextTypes (gated) AND the value is a
			// nested ObjectLiteral. CallExpression-wrapped values (e.g.
			// `propTypes: forbidExtraProps({...})`) are NOT recursed into,
			// matching upstream — see PropertyDeclaration / BinaryExpression
			// listeners for those forms.
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				obj := node.AsObjectLiteralExpression()
				if obj == nil || obj.Properties == nil {
					return
				}
				for _, prop := range obj.Properties.Nodes {
					if prop.Kind != ast.KindPropertyAssignment {
						continue
					}
					pa := prop.AsPropertyAssignment()
					keyNode := pa.Name()
					if keyNode == nil {
						continue
					}
					keyName, ok := utils.GetStaticPropertyName(keyNode)
					if !ok || !isRelevantName(keyName) {
						continue
					}
					value := reactutil.SkipExpressionWrappers(pa.Initializer)
					if value != nil && value.Kind == ast.KindObjectLiteralExpression {
						checkProperties(value.AsObjectLiteralExpression().Properties.Nodes)
					}
				}
			},
		}
	},
}

// firstSpecifierLocalName returns the local name of the first import
// specifier — default, namespace, or first named — in source order. Mirrors
// `node.specifiers[0].local.name` in ESTree, which is the bound local
// identifier regardless of specifier kind.
func firstSpecifierLocalName(clause *ast.Node) string {
	if clause == nil {
		return ""
	}
	c := clause.AsImportClause()
	if c == nil {
		return ""
	}
	// Default import (`import Foo from "x"`) sits at position 0 in ESTree
	// when present.
	if c.Name() != nil && c.Name().Kind == ast.KindIdentifier {
		return c.Name().AsIdentifier().Text
	}
	if c.NamedBindings == nil {
		return ""
	}
	switch c.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		ns := c.NamedBindings.AsNamespaceImport()
		if ns != nil && ns.Name() != nil && ns.Name().Kind == ast.KindIdentifier {
			return ns.Name().AsIdentifier().Text
		}
	case ast.KindNamedImports:
		named := c.NamedBindings.AsNamedImports()
		if named == nil || named.Elements == nil || len(named.Elements.Nodes) == 0 {
			return ""
		}
		first := named.Elements.Nodes[0].AsImportSpecifier()
		if first != nil && first.Name() != nil && first.Name().Kind == ast.KindIdentifier {
			return first.Name().AsIdentifier().Text
		}
	}
	return ""
}

// findNamedImportLocal looks for a named specifier whose imported (original)
// name matches `imported` and returns its local (possibly aliased) name.
// Mirrors upstream:
//
//	specifier.imported && specifier.imported.name === imported
//	  → return specifier.local.name
func findNamedImportLocal(clause *ast.Node, imported string) string {
	if clause == nil {
		return ""
	}
	c := clause.AsImportClause()
	if c == nil || c.NamedBindings == nil || c.NamedBindings.Kind != ast.KindNamedImports {
		return ""
	}
	named := c.NamedBindings.AsNamedImports()
	if named == nil || named.Elements == nil {
		return ""
	}
	for _, elem := range named.Elements.Nodes {
		spec := elem.AsImportSpecifier()
		if spec == nil {
			continue
		}
		// `PropertyName` carries the imported name in `{X as Y}`;
		// otherwise `Name()` is both imported and local.
		var importedName string
		if spec.PropertyName != nil && spec.PropertyName.Kind == ast.KindIdentifier {
			importedName = spec.PropertyName.AsIdentifier().Text
		} else if spec.Name() != nil && spec.Name().Kind == ast.KindIdentifier {
			importedName = spec.Name().AsIdentifier().Text
		}
		if importedName != imported {
			continue
		}
		if spec.Name() != nil && spec.Name().Kind == ast.KindIdentifier {
			return spec.Name().AsIdentifier().Text
		}
	}
	return ""
}

// hasLocalName reports whether any specifier in `clause` (default,
// namespace, or named) binds the local name `name`. Mirrors upstream's
// `node.specifiers.some((x) => x.local.name === 'PropTypes')` test for
// foreign-package classification.
func hasLocalName(clause *ast.Node, name string) bool {
	if clause == nil {
		return false
	}
	c := clause.AsImportClause()
	if c == nil {
		return false
	}
	if c.Name() != nil && c.Name().Kind == ast.KindIdentifier && c.Name().AsIdentifier().Text == name {
		return true
	}
	if c.NamedBindings == nil {
		return false
	}
	switch c.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		ns := c.NamedBindings.AsNamespaceImport()
		if ns != nil && ns.Name() != nil && ns.Name().Kind == ast.KindIdentifier && ns.Name().AsIdentifier().Text == name {
			return true
		}
	case ast.KindNamedImports:
		named := c.NamedBindings.AsNamedImports()
		if named == nil || named.Elements == nil {
			return false
		}
		for _, elem := range named.Elements.Nodes {
			spec := elem.AsImportSpecifier()
			if spec != nil && spec.Name() != nil && spec.Name().Kind == ast.KindIdentifier && spec.Name().AsIdentifier().Text == name {
				return true
			}
		}
	}
	return false
}
