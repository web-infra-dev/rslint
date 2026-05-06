package boolean_prop_naming

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var BooleanPropNamingRule = rule.Rule{
	Name: "react/boolean-prop-naming",
	Run:  runRule,
}

type ruleOptions struct {
	rule           *regexp.Regexp
	rulePattern    string
	propTypeNames  map[string]bool
	customMessage  string
	validateNested bool
}

// parseOptions extracts the rule's options from rslint's weakly-typed input.
//
// Mirrors upstream's behavior at the boundary:
//   - missing / non-object options → all defaults; rule is a no-op (rule is
//     gated by an explicit `rule` regex below).
//   - missing `propTypeNames` → defaults to `["bool"]`.
//   - empty `propTypeNames` array (`[]`) → user explicitly cleared the
//     allow-list, so no PropTypes-style identifier counts as a boolean
//     marker. TS `: boolean` annotations remain checked because they go
//     through the `tsCheck` path independent of `propTypeNames`. Upstream
//     would have its loader reject this via `minItems: 1`; rslint can't
//     refuse the config so we honor user intent over loader behavior.
//   - `rule` is read as a string. An empty string ("") is preserved and
//     produces a no-op (matching upstream's `config.rule ? ... : null`).
func parseOptions(input any) ruleOptions {
	opts := ruleOptions{
		propTypeNames: map[string]bool{"bool": true},
	}
	optsMap := utils.GetOptionsMap(input)
	if optsMap == nil {
		return opts
	}
	if pat, ok := optsMap["rule"].(string); ok {
		opts.rulePattern = pat
	}
	if names, ok := optsMap["propTypeNames"].([]interface{}); ok {
		set := map[string]bool{}
		for _, n := range names {
			if s, ok := n.(string); ok && s != "" {
				set[s] = true
			}
		}
		// Even when set is empty, swap it in — caller cleared the list.
		opts.propTypeNames = set
	}
	if msg, ok := optsMap["message"].(string); ok {
		opts.customMessage = msg
	}
	if v, ok := optsMap["validateNested"].(bool); ok {
		opts.validateNested = v
	}
	return opts
}

// templatePlaceholder matches an ESLint-style `{{key}}` placeholder.
// Mirrors ESLint's `interpolate(message, data)` regex (`/{{([^{}]+?)}}/g`):
// any non-brace content between `{{` and `}}` is captured as the key, with
// surrounding whitespace stripped before lookup. Unknown keys (not present
// in `data`) are left as the literal `{{key}}`, matching upstream.
var templatePlaceholder = regexp.MustCompile(`\{\{([^{}]+?)\}\}`)

func renderTemplate(tmpl string, data map[string]string) string {
	return templatePlaceholder.ReplaceAllStringFunc(tmpl, func(match string) string {
		m := templatePlaceholder.FindStringSubmatch(match)
		key := strings.TrimSpace(m[1])
		if v, ok := data[key]; ok {
			return v
		}
		return match
	})
}

// resolveDepth caps `validateTypeNode`'s recursion so a pathological
// `type A = B; type B = A` (reachable upstream too, where the equivalent
// `objectTypeAnnotations.get(...)` returns nothing on cycles) can't recurse
// forever. The budget is consumed by every recursive hop —
// intersection / union branches, parenthesized wrappers, type-alias
// indirection, interface heritage extends — so a value of 8 covers
// realistic depth (e.g. `type Outer = Mid & (Inner1 | Inner2)` with two
// alias hops on each side) while still bounding pathological chains.
// Deeper chains degrade gracefully to "no validation" (same as upstream's
// single-level lookup for unresolvable references).
const resolveDepth = 8

func runRule(ctx rule.RuleContext, input any) rule.RuleListeners {
	opts := parseOptions(input)
	// Upstream: `const rule = config.rule ? new RegExp(config.rule) : null;`
	// followed by `if (!rule) return;` in every listener.
	if opts.rulePattern == "" {
		return rule.RuleListeners{}
	}
	re, err := regexp.Compile(opts.rulePattern)
	if err != nil {
		// Upstream throws (`new RegExp(...)`) on a malformed pattern,
		// blowing up the whole lint run. We choose to silently degrade so
		// one bad project config doesn't make rslint unusable. This is a
		// documented divergence; tests lock the no-op behavior.
		return rule.RuleListeners{}
	}
	opts.rule = re

	pragma := reactutil.GetReactPragma(ctx.Settings)
	createClass := reactutil.GetReactCreateClass(ctx.Settings)
	propWrappers := reactutil.GetPropWrapperFunctions(ctx.Settings)
	componentWrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)

	// reportedSet protects against a single PropertySignature / property
	// being reported twice when both `static propTypes = {...}` and `props:
	// {...}` declarations exist on the same class, or both
	// `getComponentTypeAnnotation` paths (param[0] type + variable type
	// argument) resolve to the same node. Upstream's component-centric
	// accumulator is what implicitly dedupes; we dedupe by node pointer.
	reportedSet := map[*ast.Node]bool{}
	report := func(propNode *ast.Node, propName string) {
		if reportedSet[propNode] {
			return
		}
		reportedSet[propNode] = true
		data := map[string]string{
			"propName":  propName,
			"component": propName,
			"pattern":   opts.rulePattern,
		}
		var description string
		if opts.customMessage != "" {
			description = renderTemplate(opts.customMessage, data)
		} else {
			description = renderTemplate("Prop name `{{propName}}` doesn't match rule `{{pattern}}`", data)
		}
		ctx.ReportNode(propNode, rule.RuleMessage{
			Id:          "patternMismatch",
			Description: description,
			Data:        data,
		})
	}

	// First pass over the source file:
	//
	//  1. Build typeAliasMap: type-alias / interface name → its expanded
	//     set of TypeLiteral / InterfaceDeclaration nodes (intersection /
	//     union / parenthesized wrappers are flattened). Mirrors upstream's
	//     `findAllTypeAnnotations`. Interface declaration merging is
	//     supported because each declaration appends to the slice under
	//     the same name.
	//
	//  2. Populate componentsByName so `Foo.propTypes = ...` and
	//     `Foo['propTypes'] = ...` listeners can gate off non-component
	//     identifiers (mirrors upstream's `Components.detect` +
	//     `getRelatedComponent`).
	//
	//  3. Validate each function-component's TS type annotation directly —
	//     either the first parameter's type annotation, or the first type
	//     argument of the variable's type (for `React.FC<Props>` and
	//     friends). This is the rslint replacement for upstream's
	//     `Program:exit` traversal.
	//
	// rule.Run runs once per file before any listener fires, so the
	// alias map and component name set are fully populated by the time the
	// listeners need them.
	typeAliasMap := map[string][]*ast.Node{}
	componentsByName := map[string]bool{}
	type fnComponentEntry struct {
		paramType *ast.Node
		fcTypeArg *ast.Node
	}
	var fnComponents []fnComponentEntry

	var preWalk ast.Visitor
	preWalk = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		switch n.Kind {
		case ast.KindTypeAliasDeclaration:
			decl := n.AsTypeAliasDeclaration()
			if decl != nil && decl.Name() != nil && decl.Name().Kind == ast.KindIdentifier {
				flattenTypeIntoMap(typeAliasMap, decl.Name().AsIdentifier().Text, decl.Type)
			}
		case ast.KindInterfaceDeclaration:
			// Interface declaration merging: each `interface Foo {...}`
			// appends under the same name.
			decl := n.AsInterfaceDeclaration()
			if decl != nil && decl.Name() != nil && decl.Name().Kind == ast.KindIdentifier {
				name := decl.Name().AsIdentifier().Text
				typeAliasMap[name] = append(typeAliasMap[name], n)
			}
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if reactutil.ExtendsReactComponent(n, pragma) {
				if name := reactutil.BindingIdentifierName(n); name != "" {
					componentsByName[name] = true
				}
			}
		case ast.KindFunctionDeclaration:
			if reactutil.IsStatelessReactComponent(n, pragma) {
				if name := reactutil.BindingIdentifierName(n); name != "" {
					componentsByName[name] = true
				}
				if pt := reactutil.FirstParamType(n); pt != nil {
					fnComponents = append(fnComponents, fnComponentEntry{paramType: pt})
				}
			}
		case ast.KindVariableDeclaration:
			vd := n.AsVariableDeclaration()
			if vd == nil || vd.Initializer == nil {
				break
			}
			nameNode := vd.Name()
			if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
				break
			}
			init := reactutil.SkipExpressionWrappers(vd.Initializer)

			var fcArg *ast.Node
			if vd.Type != nil {
				fcArg = firstTypeArgumentOfType(vd.Type)
			}

			// `const X = class extends React.Component {}` registers `X` as
			// a component without further checks.
			if init.Kind == ast.KindClassExpression && reactutil.ExtendsReactComponent(init, pragma) {
				componentsByName[nameNode.AsIdentifier().Text] = true
				break
			}
			// `const X = memo(...)` / `forwardRef(...)` / nested
			// combinations like `memo(forwardRef(fn))`. Mirrors
			// upstream's `Components.detect` HOC recognition; honors
			// `settings.componentWrapperFunctions` via the shared
			// reactutil helper. We unwrap the chain and validate the
			// innermost function's first param type annotation.
			if init.Kind == ast.KindCallExpression {
				inner := unwrapComponentWrapperChain(init, componentWrappers)
				if inner != nil {
					componentsByName[nameNode.AsIdentifier().Text] = true
					if pt := reactutil.FirstParamType(inner); pt != nil {
						fnComponents = append(fnComponents, fnComponentEntry{paramType: pt})
					}
					if vd.Type != nil {
						if fcArg2 := firstTypeArgumentOfType(vd.Type); fcArg2 != nil {
							fnComponents = append(fnComponents, fnComponentEntry{fcTypeArg: fcArg2})
						}
					}
					break
				}
			}
			if init.Kind != ast.KindArrowFunction && init.Kind != ast.KindFunctionExpression {
				break
			}
			isComp := false
			if reactutil.IsStatelessReactComponent(init, pragma) {
				isComp = true
			} else if fcArg != nil {
				// Trust an explicit `React.FC<Props>` / `<Props>`
				// generic annotation even when the body wouldn't on its
				// own classify (e.g. `() => null`). Matches upstream:
				// `getComponentTypeAnnotation` for the
				// VariableDeclarator+typeAnnotation arm doesn't gate on
				// JSX-return-ness.
				isComp = true
			}
			if !isComp {
				break
			}
			componentsByName[nameNode.AsIdentifier().Text] = true

			pt := reactutil.FirstParamType(init)
			if pt != nil || fcArg != nil {
				fnComponents = append(fnComponents, fnComponentEntry{paramType: pt, fcTypeArg: fcArg})
			}
		}
		n.ForEachChild(preWalk)
		return false
	}
	ctx.SourceFile.Node.ForEachChild(preWalk)

	for _, c := range fnComponents {
		if c.paramType != nil {
			validateTypeNode(c.paramType, typeAliasMap, opts, report, resolveDepth)
		}
		if c.fcTypeArg != nil {
			validateTypeNode(c.fcTypeArg, typeAliasMap, opts, report, resolveDepth)
		}
	}

	return rule.RuleListeners{
		// Class fields: `static propTypes = {...}` and TS / Flow-style
		// `props: <Type>`. Both are PropertyDeclaration in tsgo.
		ast.KindPropertyDeclaration: func(node *ast.Node) {
			pd := node.AsPropertyDeclaration()
			if pd == nil {
				return
			}
			cls := reactutil.EnclosingClass(node)
			if cls == nil || !reactutil.ExtendsReactComponent(cls, pragma) {
				return
			}
			keyNode := pd.Name()
			if keyNode == nil {
				return
			}
			keyName, ok := utils.GetStaticPropertyName(keyNode)
			if !ok {
				return
			}
			isPropsField := keyName == "props" && pd.Type != nil
			if keyName != "propTypes" && !isPropsField {
				return
			}
			if pd.Initializer != nil {
				init := reactutil.SkipExpressionWrappers(pd.Initializer)
				if init.Kind == ast.KindCallExpression && reactutil.IsPropWrapperCall(init, propWrappers) {
					checkPropWrapperArguments(init, opts, report)
				} else if init.Kind == ast.KindObjectLiteralExpression {
					validateObjectLiteralProps(init, opts, report)
				}
			}
			if pd.Type != nil {
				validateTypeNode(pd.Type, typeAliasMap, opts, report, resolveDepth)
			}
		},

		// Assignment outside the class body:
		//   `Foo.propTypes = ...`,
		//   `Foo['propTypes'] = ...`,
		//   `ns.Foo.propTypes = ...`.
		// All three reach this listener as a BinaryExpression with `=`.
		ast.KindBinaryExpression: func(node *ast.Node) {
			bin := node.AsBinaryExpression()
			if bin == nil || bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindEqualsToken {
				return
			}
			left := reactutil.SkipExpressionWrappers(bin.Left)

			// LHS shapes accepted:
			//   - `<expr>.propTypes`        (PropertyAccessExpression)
			//   - `<expr>["propTypes"]`     (ElementAccessExpression with
			//     a static string-literal argument)
			var componentExpr *ast.Node
			var propertyName string
			switch left.Kind {
			case ast.KindPropertyAccessExpression:
				pa := left.AsPropertyAccessExpression()
				nameNode := pa.Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					return
				}
				componentExpr = pa.Expression
				propertyName = nameNode.AsIdentifier().Text
			case ast.KindElementAccessExpression:
				ea := left.AsElementAccessExpression()
				if ea == nil || ea.ArgumentExpression == nil {
					return
				}
				name, ok := utils.GetStaticPropertyName(ea.ArgumentExpression)
				if !ok {
					return
				}
				componentExpr = ea.Expression
				propertyName = name
			default:
				return
			}
			if propertyName != "propTypes" {
				return
			}
			if !isKnownComponentExpr(componentExpr, componentsByName) {
				return
			}
			right := reactutil.SkipExpressionWrappers(bin.Right)
			if right.Kind == ast.KindCallExpression && reactutil.IsPropWrapperCall(right, propWrappers) {
				checkPropWrapperArguments(right, opts, report)
				return
			}
			if right.Kind == ast.KindObjectLiteralExpression {
				validateObjectLiteralProps(right, opts, report)
			}
		},

		// `createReactClass({propTypes: {...}})` /
		// `<pragma>.createClass({...})`. Other listeners cover the class /
		// assignment forms.
		ast.KindObjectLiteralExpression: func(node *ast.Node) {
			if !reactutil.IsCreateReactClassObjectArg(node, pragma, createClass) {
				return
			}
			for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
				if prop.Kind != ast.KindPropertyAssignment {
					continue
				}
				pa := prop.AsPropertyAssignment()
				key := pa.Name()
				if key == nil {
					continue
				}
				keyName, ok := utils.GetStaticPropertyName(key)
				if !ok || keyName != "propTypes" {
					continue
				}
				value := reactutil.SkipExpressionWrappers(pa.Initializer)
				if value.Kind == ast.KindObjectLiteralExpression {
					validateObjectLiteralProps(value, opts, report)
				} else if value.Kind == ast.KindCallExpression && reactutil.IsPropWrapperCall(value, propWrappers) {
					// Upstream's `validatePropNaming` is invoked with
					// `property.value.properties` so a wrapped propTypes
					// inside a createReactClass arg is silently skipped.
					// We choose to honor propWrapperFunctions here too —
					// it's a strict superset of upstream's behavior, and
					// the equivalent class-form path below already does
					// the same. Document if a real conflict arises.
					checkPropWrapperArguments(value, opts, report)
				}
			}
		},
	}
}

// isKnownComponentExpr reports whether `expr` (the receiver of a
// `.propTypes`-style assignment LHS) names something we previously
// classified as a React component.
//
// Accepted shapes:
//   - bare Identifier: `Foo`
//   - PropertyAccessExpression chain ending in an Identifier whose name is
//     in `componentsByName`: `ns.Foo`, `outer.inner.Foo`. Upstream uses
//     `utils.getRelatedComponent(node)` which walks the file's component
//     map; we approximate by checking the rightmost identifier of any
//     bare member chain. This is intentionally conservative — call
//     receivers (`getFoo().propTypes`) and computed access aren't
//     accepted, matching upstream's preference for static recognition.
func isKnownComponentExpr(expr *ast.Node, componentsByName map[string]bool) bool {
	expr = reactutil.SkipExpressionWrappers(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		return componentsByName[expr.AsIdentifier().Text]
	case ast.KindPropertyAccessExpression:
		pa := expr.AsPropertyAccessExpression()
		nameNode := pa.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return false
		}
		return componentsByName[nameNode.AsIdentifier().Text]
	}
	return false
}

// flattenTypeIntoMap registers each leaf TypeLiteral / InterfaceDeclaration
// reachable from `typeNode` under `name`. Intersection / union /
// parenthesized wrappers are descended into. Named TypeReferences are
// stored as forwarding entries so a chain like `type A = B; type B =
// {x:boolean}` resolves at validation time (rslint extends upstream here:
// upstream's `findAllTypeAnnotations` ignores TypeReferences entirely, so
// `type A = B` would silently lose its boolean members on the alias map).
//
// `validateTypeNode` is responsible for following the forwarding entries
// with a depth budget, so cycles can't recurse forever.
func flattenTypeIntoMap(m map[string][]*ast.Node, name string, typeNode *ast.Node) {
	if typeNode == nil {
		return
	}
	switch typeNode.Kind {
	case ast.KindTypeLiteral, ast.KindInterfaceDeclaration, ast.KindTypeReference:
		m[name] = append(m[name], typeNode)
	case ast.KindParenthesizedType:
		flattenTypeIntoMap(m, name, typeNode.AsParenthesizedTypeNode().Type)
	case ast.KindIntersectionType:
		for _, t := range typeNode.AsIntersectionTypeNode().Types.Nodes {
			flattenTypeIntoMap(m, name, t)
		}
	case ast.KindUnionType:
		for _, t := range typeNode.AsUnionTypeNode().Types.Nodes {
			flattenTypeIntoMap(m, name, t)
		}
	}
}

// validateTypeNode resolves `typeNode` to a list of TypeLiteral / Interface
// member sets and validates each. TypeReferences are looked up in
// typeAliasMap (one hop per call; the depth budget allows us to follow
// `type A = B; type B = {x:boolean}` without recursing forever on cycles).
// Intersection / union / parenthesized wrappers are descended into. Inline
// TypeLiterals are validated directly. Anything else (primitive keyword,
// function type, generic with a non-resolvable name) is ignored.
func validateTypeNode(typeNode *ast.Node, typeAliasMap map[string][]*ast.Node, opts ruleOptions, report func(*ast.Node, string), depth int) {
	if typeNode == nil || depth <= 0 {
		return
	}
	switch typeNode.Kind {
	case ast.KindTypeLiteral:
		validateMembers(typeNode.AsTypeLiteralNode().Members.Nodes, opts, report)
	case ast.KindInterfaceDeclaration:
		decl := typeNode.AsInterfaceDeclaration()
		validateMembers(decl.Members.Nodes, opts, report)
		// Walk `extends` clauses so `interface A extends B { ... }`
		// validates B's members too. ESLint's plugin doesn't follow
		// heritage (it treats the interface as opaque outside the
		// component's own annotation chain), but real-world TS code
		// composes extensively via interface extension; following one
		// hop covers the common case without paying for arbitrary
		// recursion depth (the depth budget caps it anyway).
		if decl.HeritageClauses != nil {
			for _, hc := range decl.HeritageClauses.Nodes {
				if hc.Kind != ast.KindHeritageClause {
					continue
				}
				clause := hc.AsHeritageClause()
				if clause == nil || clause.Types == nil {
					continue
				}
				for _, t := range clause.Types.Nodes {
					// Each entry is ExpressionWithTypeArguments;
					// resolve via its expression as a type-name lookup.
					if t.Kind != ast.KindExpressionWithTypeArguments {
						continue
					}
					expr := t.AsExpressionWithTypeArguments()
					if expr == nil {
						continue
					}
					name := reactutil.EntityNameRightmost(expr.Expression)
					if name == nil {
						continue
					}
					for _, target := range typeAliasMap[name.AsIdentifier().Text] {
						validateTypeNode(target, typeAliasMap, opts, report, depth-1)
					}
				}
			}
		}
	case ast.KindTypeReference:
		ref := typeNode.AsTypeReferenceNode()
		nameIdent := reactutil.EntityNameRightmost(ref.TypeName)
		if nameIdent == nil {
			return
		}
		for _, t := range typeAliasMap[nameIdent.AsIdentifier().Text] {
			validateTypeNode(t, typeAliasMap, opts, report, depth-1)
		}
	case ast.KindParenthesizedType:
		validateTypeNode(typeNode.AsParenthesizedTypeNode().Type, typeAliasMap, opts, report, depth-1)
	case ast.KindIntersectionType:
		for _, t := range typeNode.AsIntersectionTypeNode().Types.Nodes {
			validateTypeNode(t, typeAliasMap, opts, report, depth-1)
		}
	case ast.KindUnionType:
		for _, t := range typeNode.AsUnionTypeNode().Types.Nodes {
			validateTypeNode(t, typeAliasMap, opts, report, depth-1)
		}
	}
}

// validateMembers walks TypeElement members of a TypeLiteral / Interface and
// reports each PropertySignature whose type is `boolean` and whose name does
// not match the configured pattern. Other member kinds (index signatures,
// method signatures, call signatures) are ignored.
//
// `boolean` recognition includes parenthesized wrappers (`(boolean)`) so
// `prop: (boolean)` reports identically to `prop: boolean`. Type references
// to user types are not resolved here — `prop: MyBool` is treated as opaque,
// mirroring upstream's `tsCheck` which only matches `TSBooleanKeyword`.
func validateMembers(members []*ast.Node, opts ruleOptions, report func(*ast.Node, string)) {
	for _, m := range members {
		if m.Kind != ast.KindPropertySignature {
			continue
		}
		ps := m.AsPropertySignatureDeclaration()
		if !typeIsBooleanKeyword(ps.Type) {
			continue
		}
		nameNode := ps.Name()
		if nameNode == nil {
			continue
		}
		name, ok := utils.GetStaticPropertyName(nameNode)
		if !ok || name == "" {
			continue
		}
		if !opts.rule.MatchString(name) {
			report(m, name)
		}
	}
}

// typeIsBooleanKeyword reports whether `t` resolves to `boolean`, peeling
// any number of `ParenthesizedType` wrappers — `boolean` and `((boolean))`
// are equivalent.
func typeIsBooleanKeyword(t *ast.Node) bool {
	for t != nil {
		switch t.Kind {
		case ast.KindBooleanKeyword:
			return true
		case ast.KindParenthesizedType:
			t = t.AsParenthesizedTypeNode().Type
		default:
			return false
		}
	}
	return false
}

// validateObjectLiteralProps walks an object literal of PropTypes-style
// declarations and reports each entry whose value resolves to one of the
// configured propTypeNames AND whose key does not match the regex.
//
// Recurses into `PropTypes.shape({...})` arguments when `validateNested` is
// enabled, mirroring upstream's `runCheck`. The recursion accepts
// CallExpression unconditionally — upstream gates on `Property` (non-spread,
// non-shorthand) only, which we cover by the early continue above.
func validateObjectLiteralProps(obj *ast.Node, opts ruleOptions, report func(*ast.Node, string)) {
	for _, prop := range obj.AsObjectLiteralExpression().Properties.Nodes {
		if prop.Kind != ast.KindPropertyAssignment {
			// SpreadAssignment, ShorthandPropertyAssignment, methods,
			// accessors — none have a `key: PropTypes.X` shape we can
			// examine, so all are skipped (matches upstream's
			// `getPropKey` which returns null for these).
			continue
		}
		pa := prop.AsPropertyAssignment()
		if pa.Initializer == nil {
			continue
		}
		valueUnwrapped := reactutil.SkipExpressionWrappers(pa.Initializer)

		if opts.validateNested && valueUnwrapped.Kind == ast.KindCallExpression {
			call := valueUnwrapped.AsCallExpression()
			if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
				inner := reactutil.SkipExpressionWrappers(call.Arguments.Nodes[0])
				if inner.Kind == ast.KindObjectLiteralExpression {
					validateObjectLiteralProps(inner, opts, report)
					continue
				}
			}
		}

		propKey := getPropTypeKey(valueUnwrapped)
		if propKey == "" || !opts.propTypeNames[propKey] {
			continue
		}
		nameNode := pa.Name()
		if nameNode == nil {
			continue
		}
		name, ok := utils.GetStaticPropertyName(nameNode)
		if !ok || name == "" {
			continue
		}
		if !opts.rule.MatchString(name) {
			report(prop, name)
		}
	}
}

// getPropTypeKey extracts the rightmost type-name identifier of a prop
// value expression — `PropTypes.bool`, `React.PropTypes.bool`, `bool`, and
// `PropTypes.bool.isRequired` all yield `"bool"`. Returns "" for shapes
// upstream's `getPropKey` rejects (e.g. the `.isRequired` of a
// `shape({...}).isRequired`, where the inner is a CallExpression).
//
// Receivers in the chain are unwrapped via `SkipExpressionWrappers`, so
// `(PropTypes.bool as any)` and `PropTypes.bool!` resolve identically to
// `PropTypes.bool`.
func getPropTypeKey(value *ast.Node) string {
	value = reactutil.SkipExpressionWrappers(value)
	switch value.Kind {
	case ast.KindIdentifier:
		return value.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		pa := value.AsPropertyAccessExpression()
		nameNode := pa.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return ""
		}
		propName := nameNode.AsIdentifier().Text
		if propName == "isRequired" {
			obj := reactutil.SkipExpressionWrappers(pa.Expression)
			if obj.Kind != ast.KindPropertyAccessExpression {
				return ""
			}
			objPa := obj.AsPropertyAccessExpression()
			objName := objPa.Name()
			if objName == nil || objName.Kind != ast.KindIdentifier {
				return ""
			}
			return objName.AsIdentifier().Text
		}
		return propName
	}
	return ""
}

// firstTypeArgumentOfType returns the FIRST type argument of a TypeReference
// (e.g. `React.FC<Props>` → `Props`). Returns nil if the type is not a
// TypeReference, or has no type arguments.
//
// Mirrors upstream's `annotationTypeArguments.params.find(...)` — but unlike
// upstream we don't filter by node kind, since `validateTypeNode`
// gracefully ignores primitive-keyword and other non-object kinds at the
// leaf.
func firstTypeArgumentOfType(typeNode *ast.Node) *ast.Node {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return nil
	}
	ref := typeNode.AsTypeReferenceNode()
	if ref.TypeArguments == nil || len(ref.TypeArguments.Nodes) == 0 {
		return nil
	}
	return ref.TypeArguments.Nodes[0]
}

// unwrapComponentWrapperChain peels off any number of nested known
// component-wrapper calls (`memo(...)`, `forwardRef(...)`,
// `React.memo(...)`, `React.forwardRef(...)`, plus any user-configured
// `settings.componentWrapperFunctions` entries) and returns the innermost
// arrow / function expression — or nil if no wrapper at the head matches,
// or if the innermost argument is not a function expression.
//
// Examples (assume default wrappers):
//
//	memo(fn)                       → fn
//	memo(forwardRef(fn))           → fn
//	React.memo(forwardRef(fn))     → fn
//	connect(s,d)(memo(fn))         → nil (outer call's callee isn't a
//	                                  recognized wrapper, even though the
//	                                  inner is — upstream's
//	                                  `isPragmaComponentWrapper` makes the
//	                                  same conservative call)
//	memo(someIdentifier)           → nil (innermost arg is not a function
//	                                  literal; we can't validate its props)
//
// The chain depth is implicitly bounded by the source file's nesting
// depth, but we cap at 16 hops as a defensive guard.
func unwrapComponentWrapperChain(call *ast.Node, wrappers []reactutil.ComponentWrapperEntry) *ast.Node {
	for range 16 {
		if call == nil || call.Kind != ast.KindCallExpression {
			return nil
		}
		args := call.AsCallExpression().Arguments
		if args == nil || len(args.Nodes) == 0 {
			return nil
		}
		first := reactutil.SkipExpressionWrappers(args.Nodes[0])
		// `MatchesAnyComponentWrapper` requires the inner `fn` for its
		// first-arg sanity check; we already have it as `first`.
		if !reactutil.MatchesAnyComponentWrapper(call, first, wrappers) {
			return nil
		}
		switch first.Kind {
		case ast.KindArrowFunction, ast.KindFunctionExpression:
			return first
		case ast.KindCallExpression:
			call = first
			continue
		}
		return nil
	}
	return nil
}

// checkPropWrapperArguments validates each ObjectLiteralExpression argument
// of `call`. Non-object arguments (e.g. the `{}` target and
// `Card.propTypes` reference of `Object.assign({}, Card.propTypes, {...})`)
// are ignored. TS-wrapped object arguments (`(arg as Foo)`) are accepted.
func checkPropWrapperArguments(call *ast.Node, opts ruleOptions, report func(*ast.Node, string)) {
	args := call.AsCallExpression().Arguments
	if args == nil {
		return
	}
	for _, arg := range args.Nodes {
		obj := reactutil.SkipExpressionWrappers(arg)
		if obj.Kind == ast.KindObjectLiteralExpression {
			validateObjectLiteralProps(obj, opts, report)
		}
	}
}
