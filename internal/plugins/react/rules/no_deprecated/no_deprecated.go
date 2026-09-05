package no_deprecated

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// deprecationInfo holds the triple upstream's `deprecated[method]` returns:
// the React version in which the method was deprecated, the replacement
// (`, use X instead` — empty when no replacement is suggested), and an
// optional external reference URL/notes (`, see Y`).
type deprecationInfo struct {
	version   string
	newMethod string
	refs      string
}

// modulesList mirrors upstream's MODULES map — maps each npm package source
// to the canonical binding name(s) that users typically give to it in a
// destructuring / import. The FIRST entry of `names` is the canonical name
// used to synthesize deprecation keys (e.g. `React.X`, `ReactDOM.X`).
var modulesList = []struct {
	source string
	names  []string
}{
	{"react", []string{"React"}},
	{"react-addons-perf", []string{"ReactPerf", "Perf"}},
	{"react-dom", []string{"ReactDOM"}},
	{"react-dom/server", []string{"ReactDOMServer"}},
}

// canonicalForModuleSource returns the canonical binding name (MODULES[key][0])
// for a module string like "react-dom", or "" when the module isn't a React
// npm package we track.
func canonicalForModuleSource(source string) string {
	for _, m := range modulesList {
		if m.source == source {
			return m.names[0]
		}
	}
	return ""
}

// canonicalForModuleIdentifier returns the canonical binding name for an
// Identifier whose text equals any known module's alias (e.g. `Perf` →
// `Perf`, since `ReactPerf` / `Perf` are both aliases of the
// `react-addons-perf` module). Mirrors upstream's second-arm match in
// `getReactModuleName` — the matched name itself becomes the module binding.
func canonicalForModuleIdentifier(name string) string {
	for _, m := range modulesList {
		for _, n := range m.names {
			if n == name {
				return n
			}
		}
	}
	return ""
}

// buildDeprecated returns the deprecation table for a given pragma. The
// pragma is substituted into every `<pragma>.X` key — so `settings.react.pragma`
// or a `@jsx` directive changes which bare-member accesses are flagged.
// Non-pragma keys (ReactDOM.*, ReactPerf.*, Perf.*, ReactDOMServer.*,
// this.transferPropsTo, lifecycle-method names) use literal binding names
// and are unaffected by pragma. Mirrors upstream's `getDeprecated(pragma)`.
func buildDeprecated(pragma string) map[string]deprecationInfo {
	m := make(map[string]deprecationInfo, 40)
	// 0.12.0
	m[pragma+".renderComponent"] = deprecationInfo{"0.12.0", pragma + ".render", ""}
	m[pragma+".renderComponentToString"] = deprecationInfo{"0.12.0", pragma + ".renderToString", ""}
	m[pragma+".renderComponentToStaticMarkup"] = deprecationInfo{"0.12.0", pragma + ".renderToStaticMarkup", ""}
	m[pragma+".isValidComponent"] = deprecationInfo{"0.12.0", pragma + ".isValidElement", ""}
	m[pragma+".PropTypes.component"] = deprecationInfo{"0.12.0", pragma + ".PropTypes.element", ""}
	m[pragma+".PropTypes.renderable"] = deprecationInfo{"0.12.0", pragma + ".PropTypes.node", ""}
	m[pragma+".isValidClass"] = deprecationInfo{"0.12.0", "", ""}
	m["this.transferPropsTo"] = deprecationInfo{"0.12.0", "spread operator ({...})", ""}
	// 0.13.0
	m[pragma+".addons.classSet"] = deprecationInfo{"0.13.0", "the npm module classnames", ""}
	m[pragma+".addons.cloneWithProps"] = deprecationInfo{"0.13.0", pragma + ".cloneElement", ""}
	// 0.14.0
	m[pragma+".render"] = deprecationInfo{"0.14.0", "ReactDOM.render", ""}
	m[pragma+".unmountComponentAtNode"] = deprecationInfo{"0.14.0", "ReactDOM.unmountComponentAtNode", ""}
	m[pragma+".findDOMNode"] = deprecationInfo{"0.14.0", "ReactDOM.findDOMNode", ""}
	m[pragma+".renderToString"] = deprecationInfo{"0.14.0", "ReactDOMServer.renderToString", ""}
	m[pragma+".renderToStaticMarkup"] = deprecationInfo{"0.14.0", "ReactDOMServer.renderToStaticMarkup", ""}
	// 15.0.0
	m[pragma+".addons.LinkedStateMixin"] = deprecationInfo{"15.0.0", "", ""}
	m["ReactPerf.printDOM"] = deprecationInfo{"15.0.0", "ReactPerf.printOperations", ""}
	m["Perf.printDOM"] = deprecationInfo{"15.0.0", "Perf.printOperations", ""}
	m["ReactPerf.getMeasurementsSummaryMap"] = deprecationInfo{"15.0.0", "ReactPerf.getWasted", ""}
	m["Perf.getMeasurementsSummaryMap"] = deprecationInfo{"15.0.0", "Perf.getWasted", ""}
	// 15.5.0
	m[pragma+".createClass"] = deprecationInfo{"15.5.0", "the npm module create-react-class", ""}
	m[pragma+".addons.TestUtils"] = deprecationInfo{"15.5.0", "ReactDOM.TestUtils", ""}
	m[pragma+".PropTypes"] = deprecationInfo{"15.5.0", "the npm module prop-types", ""}
	// 15.6.0
	m[pragma+".DOM"] = deprecationInfo{"15.6.0", "the npm module react-dom-factories", ""}
	// 16.9.0 — lifecycle methods (keys without pragma prefix; matched by bare member name).
	lifecycleRef := "https://reactjs.org/docs/react-component.html#"
	lifecycleTail := ". Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."
	m["componentWillMount"] = deprecationInfo{"16.9.0", "UNSAFE_componentWillMount", lifecycleRef + "unsafe_componentwillmount" + lifecycleTail}
	m["componentWillReceiveProps"] = deprecationInfo{"16.9.0", "UNSAFE_componentWillReceiveProps", lifecycleRef + "unsafe_componentwillreceiveprops" + lifecycleTail}
	m["componentWillUpdate"] = deprecationInfo{"16.9.0", "UNSAFE_componentWillUpdate", lifecycleRef + "unsafe_componentwillupdate" + lifecycleTail}
	// 18.0.0 — react-dom / react-dom/server deprecations (literal ReactDOM/ReactDOMServer, pragma-independent).
	m["ReactDOM.render"] = deprecationInfo{"18.0.0", "createRoot", "https://reactjs.org/link/switch-to-createroot"}
	m["ReactDOM.hydrate"] = deprecationInfo{"18.0.0", "hydrateRoot", "https://reactjs.org/link/switch-to-createroot"}
	m["ReactDOM.unmountComponentAtNode"] = deprecationInfo{"18.0.0", "root.unmount", "https://reactjs.org/link/switch-to-createroot"}
	m["ReactDOMServer.renderToNodeStream"] = deprecationInfo{"18.0.0", "renderToPipeableStream", "https://reactjs.org/docs/react-dom-server.html#rendertonodestream"}
	return m
}

// parseVersion parses a leading "major[.minor[.patch]]" numeric triple and
// returns (M, m, p). Unparseable components become 0. Prerelease / build
// metadata tails are ignored — matches the lenient comparison used by
// eslint-plugin-react's version util for simple `>= X` checks.
func parseVersion(s string) (int, int, int) {
	var parts [3]int
	segmentStart := 0
	for part := 0; part < len(parts) && segmentStart <= len(s); part++ {
		segmentEnd := len(s)
		hasNext := false
		if dot := strings.IndexByte(s[segmentStart:], '.'); dot >= 0 {
			segmentEnd = segmentStart + dot
			hasNext = true
		}

		// Strip a trailing non-digit tail (e.g. "-rc" or "+build") and
		// parse the numeric prefix without allocating a split slice.
		digitEnd := segmentStart
		for digitEnd < segmentEnd && s[digitEnd] >= '0' && s[digitEnd] <= '9' {
			digitEnd++
		}
		if digitEnd > segmentStart {
			value := 0
			overflow := false
			maxInt := int(^uint(0) >> 1)
			for i := segmentStart; i < digitEnd; i++ {
				digit := int(s[i] - '0')
				if value > (maxInt-digit)/10 {
					overflow = true
					break
				}
				value = value*10 + digit
			}
			if !overflow {
				parts[part] = value
			}
		}

		if !hasNext {
			break
		}
		segmentStart = segmentEnd + 1
	}
	return parts[0], parts[1], parts[2]
}

// versionAtLeast compares a React version parsed once per file with one of
// the fixed deprecation versions. Keeping the settings lookup and regexp-based
// version parsing out of each diagnostic is especially important for files
// that import or destructure several deprecated APIs.
func versionAtLeast(major, minor, patch int, deprecVersion string) bool {
	wantMajor, wantMinor, wantPatch := parseVersion(deprecVersion)
	if major != wantMajor {
		return major > wantMajor
	}
	if minor != wantMinor {
		return minor > wantMinor
	}
	return patch >= wantPatch
}

// buildDottedPath walks down the Expression chain of a PropertyAccessExpression
// and returns the dotted path "base.seg1.seg2…" when every segment is a
// bare-identifier member access through either an Identifier or `this` base.
// Returns "" for element-access (`foo['bar']`), computed bases, or any
// non-identifier reachable component — such shapes can't match an upstream
// deprecation key derived from source text.
//
// NOTE: Parentheses are transparently skipped at every step via
// `ast.SkipParentheses`. ESTree would preserve source-level parens when
// `getText(node)` reads the range, so `(React).createClass` would miss there.
// We flag it — a more permissive, rule-catches-more-cases divergence that
// we lock in via a dedicated test. See the rule's `.md` for details.
func buildDottedPath(node *ast.Node) string {
	// Deprecation keys are at most three members deep. Back the slice with a
	// small local array so all matching paths avoid a separate slice allocation;
	// append still preserves behavior for arbitrarily deep non-matching chains.
	var segmentBuffer [4]string
	segs := segmentBuffer[:0]
	cur := node
	for {
		cur = ast.SkipParentheses(cur)
		if cur.Kind != ast.KindPropertyAccessExpression {
			break
		}
		pa := cur.AsPropertyAccessExpression()
		nameNode := pa.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return ""
		}
		segs = append(segs, nameNode.AsIdentifier().Text)
		cur = pa.Expression
	}
	cur = ast.SkipParentheses(cur)
	var base string
	switch cur.Kind {
	case ast.KindIdentifier:
		base = cur.AsIdentifier().Text
	case ast.KindThisKeyword:
		base = "this"
	default:
		return ""
	}
	// segs was collected outer → inner (leaf first). Write base, then
	// append segments in reverse (inner first) to form "base.inner…leaf".
	var b strings.Builder
	b.WriteString(base)
	for i := len(segs) - 1; i >= 0; i-- {
		b.WriteByte('.')
		b.WriteString(segs[i])
	}
	return b.String()
}

// canBeDeprecatedPropertyName is a cheap allocation-free gate for the
// PropertyAccessExpression listener. Only these terminal names occur in the
// deprecation table; checking them first keeps dotted-path construction off
// unrelated property accesses such as response.data or props.children.
func canBeDeprecatedPropertyName(name string) bool {
	switch name {
	case "renderComponent",
		"renderComponentToString",
		"renderComponentToStaticMarkup",
		"isValidComponent",
		"component",
		"renderable",
		"isValidClass",
		"transferPropsTo",
		"classSet",
		"cloneWithProps",
		"render",
		"unmountComponentAtNode",
		"findDOMNode",
		"renderToString",
		"renderToStaticMarkup",
		"LinkedStateMixin",
		"printDOM",
		"getMeasurementsSummaryMap",
		"createClass",
		"TestUtils",
		"PropTypes",
		"DOM",
		"hydrate",
		"renderToNodeStream":
		return true
	default:
		return false
	}
}

// formatMessage builds the deprecation diagnostic string. Mirrors upstream's
// message template:
//
//	{{oldMethod}} is deprecated since React {{version}}{{newMethod}}{{refs}}
//
// where newMethod is `", use X instead"` when set, and refs is `", see Y"`
// when set — both empty otherwise.
func formatMessage(methodName string, d deprecationInfo) string {
	var b strings.Builder
	b.WriteString(methodName)
	b.WriteString(" is deprecated since React ")
	b.WriteString(d.version)
	if d.newMethod != "" {
		b.WriteString(", use ")
		b.WriteString(d.newMethod)
		b.WriteString(" instead")
	}
	if d.refs != "" {
		b.WriteString(", see ")
		b.WriteString(d.refs)
	}
	return b.String()
}

// The default pragma is used by virtually every file. Its lookup table and
// messages are immutable, so construct them once instead of rebuilding the
// same map and strings for every Rule.Run invocation. Custom pragmas still
// get a per-file table because their keys and replacement text are dynamic.
var defaultDeprecated = buildDeprecated(reactutil.DefaultReactPragma)

var defaultDeprecatedMessages = func() map[string]string {
	messages := make(map[string]string, len(defaultDeprecated))
	for methodName, d := range defaultDeprecated {
		messages[methodName] = formatMessage(methodName, d)
	}
	return messages
}()

var NoDeprecatedRule = rule.Rule{
	Name:   "react/no-deprecated",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// A `@jsx` annotation in the source wins over
		// `settings.react.pragma`, matching upstream's
		// `pragmaUtil.getFromContext`.
		pragma := reactutil.GetReactPragmaFromContext(ctx)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		usesDefaultPragma := pragma == reactutil.DefaultReactPragma
		deprecated := defaultDeprecated
		if !usesDefaultPragma {
			deprecated = buildDeprecated(pragma)
		}
		reactMajor, reactMinor, reactPatch := 0, 0, 0
		reactVersionParsed := false

		report := func(node *ast.Node, methodName string, d deprecationInfo) {
			description := ""
			if usesDefaultPragma {
				description = defaultDeprecatedMessages[methodName]
			}
			if description == "" {
				description = formatMessage(methodName, d)
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "deprecated",
				Description: description,
			})
		}

		// check is the common gate — present only when the key is both in the
		// table and active for the configured React version.
		check := func(node *ast.Node, methodName string) {
			d, ok := deprecated[methodName]
			if !ok {
				return
			}
			if !reactVersionParsed {
				reactMajor, reactMinor, reactPatch = reactutil.ParseReactVersion(ctx.Settings)
				reactVersionParsed = true
			}
			if !versionAtLeast(reactMajor, reactMinor, reactPatch, d.version) {
				return
			}
			report(node, methodName, d)
		}

		// checkComponentMembers inspects lifecycle-method keys on a class or
		// createReactClass object literal. Each member whose Identifier key is
		// a deprecated lifecycle name (`componentWillMount`, etc.) yields a
		// report at the key node — matching upstream's
		// `astUtil.getPropertyNameNode(property)` report target.
		checkComponentMembers := func(members []*ast.Node) {
			for _, m := range members {
				if m == nil {
					continue
				}
				// Skip SpreadAssignment / SemicolonClassElement / etc. that
				// don't carry a named key.
				key := m.Name()
				if key == nil || key.Kind != ast.KindIdentifier {
					continue
				}
				name := key.AsIdentifier().Text
				check(key, name)
			}
		}

		return rule.RuleListeners{
			// `React.createClass`, `React.addons.TestUtils`,
			// `this.transferPropsTo`, `ReactDOM.render`, … Each
			// PropertyAccessExpression level is checked independently;
			// `React.DOM.div` ⇒ inner `React.DOM` matches, outer doesn't.
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				name := node.AsPropertyAccessExpression().Name()
				if name == nil || name.Kind != ast.KindIdentifier || !canBeDeprecatedPropertyName(name.AsIdentifier().Text) {
					return
				}
				path := buildDottedPath(node)
				if path == "" {
					return
				}
				check(node, path)
			},

			// `import { createClass, PropTypes } from 'react'` → check each
			// named specifier as `<canonical>.<imported-name>`.
			ast.KindImportDeclaration: func(node *ast.Node) {
				decl := node.AsImportDeclaration()
				if decl == nil || decl.ModuleSpecifier == nil || decl.ModuleSpecifier.Kind != ast.KindStringLiteral {
					return
				}
				canonical := canonicalForModuleSource(decl.ModuleSpecifier.AsStringLiteral().Text)
				if canonical == "" {
					return
				}
				if decl.ImportClause == nil {
					return
				}
				clause := decl.ImportClause.AsImportClause()
				// ESLint's filter `'imported' in s && s.imported` excludes
				// default / namespace specifiers — only NamedImports carry
				// an `imported` identifier.
				if clause == nil || clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
					return
				}
				named := clause.NamedBindings.AsNamedImports()
				if named == nil || named.Elements == nil {
					return
				}
				for _, elem := range named.Elements.Nodes {
					spec := elem.AsImportSpecifier()
					if spec == nil {
						continue
					}
					// `PropertyName` holds the imported name when aliased
					// (`{ X as Y }`); otherwise `Name()` is the imported name.
					importedNode := spec.PropertyName
					if importedNode == nil {
						importedNode = spec.Name()
					}
					if importedNode == nil || importedNode.Kind != ast.KindIdentifier {
						continue
					}
					check(elem, canonical+"."+importedNode.AsIdentifier().Text)
				}
			},

			// Destructuring from a React-module call (`require('react')`,
			// `import('react-dom')`, etc.) or from an identifier whose name is
			// a module alias (`ReactPerf`). Mirrors upstream's
			// `VariableDeclarator` branch.
			ast.KindVariableDeclaration: func(node *ast.Node) {
				vd := node.AsVariableDeclaration()
				if vd == nil || vd.Initializer == nil {
					return
				}
				bindingName := vd.Name()
				if bindingName == nil || bindingName.Kind != ast.KindObjectBindingPattern {
					return
				}
				init := ast.SkipParentheses(vd.Initializer)

				// Arm 1 of `getReactModuleName`: init is a CallExpression and
				// the first argument is a string literal matching a module
				// source — `key === node.init.arguments[0].value`. Upstream
				// does not actually require the callee to be `require`, so
				// neither do we. (The proceed-condition's second arm
				// specifically checks `require`, but it's redundant: arm 1 of
				// getReactModuleName subsumes the require case.)
				canonical := ""
				if init.Kind == ast.KindCallExpression {
					call := init.AsCallExpression()
					if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
						arg0 := ast.SkipParentheses(call.Arguments.Nodes[0])
						if arg0.Kind == ast.KindStringLiteral {
							canonical = canonicalForModuleSource(arg0.AsStringLiteral().Text)
						}
					}
				}
				// Arm 2: init is a bare Identifier whose text matches one of
				// a module's alias names (e.g. `ReactPerf` / `Perf`).
				if canonical == "" && init.Kind == ast.KindIdentifier {
					canonical = canonicalForModuleIdentifier(init.AsIdentifier().Text)
				}
				if canonical == "" {
					return
				}

				obp := bindingName.AsBindingPattern()
				if obp == nil || obp.Elements == nil {
					return
				}
				for _, elem := range obp.Elements.Nodes {
					if elem == nil {
						continue
					}
					be := elem.AsBindingElement()
					if be == nil {
						continue
					}
					// `...rest` has no property-name; upstream filters it out
					// via `p.type !== 'RestElement' && p.key`.
					if be.DotDotDotToken != nil {
						continue
					}
					// ESLint's `property.key.name` is the imported/from-source
					// name. For an aliased binding `{ X: Y }` tsgo exposes
					// `PropertyName` (= X). For a shorthand `{ X }` the
					// `Name()` itself is the key.
					keyNode := be.PropertyName
					if keyNode == nil {
						keyNode = be.Name()
					}
					if keyNode == nil || keyNode.Kind != ast.KindIdentifier {
						continue
					}
					check(keyNode, canonical+"."+keyNode.AsIdentifier().Text)
				}
			},

			ast.KindClassDeclaration: func(node *ast.Node) {
				if !reactutil.ExtendsReactComponent(node, pragma) {
					return
				}
				checkComponentMembers(node.Members())
			},
			ast.KindClassExpression: func(node *ast.Node) {
				if !reactutil.ExtendsReactComponent(node, pragma) {
					return
				}
				checkComponentMembers(node.Members())
			},
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if !reactutil.IsCreateReactClassObjectArg(node, pragma, createClass) {
					return
				}
				ol := node.AsObjectLiteralExpression()
				if ol == nil || ol.Properties == nil {
					return
				}
				checkComponentMembers(ol.Properties.Nodes)
			},
		}
	},
}
