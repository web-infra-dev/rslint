package no_deprecated

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
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

// jsxPragmaRe matches a `@jsx Foo` directive anywhere in source text. The
// comment form (block vs. line) isn't constrained — upstream's pragmaUtil
// accepts the pragma from any comment, so we scan the raw source.
var jsxPragmaRe = regexp.MustCompile(`@jsx\s+([A-Za-z_$][\w$.]*)`)

// detectJsxPragma returns the identifier following the first `@jsx` directive
// in the source, or "" when no directive is present.
func detectJsxPragma(sourceText string) string {
	m := jsxPragmaRe.FindStringSubmatch(sourceText)
	if m == nil {
		return ""
	}
	return m[1]
}

// parseVersion parses a leading "major[.minor[.patch]]" numeric triple and
// returns (M, m, p). Unparseable components become 0. Prerelease / build
// metadata tails are ignored — matches the lenient comparison used by
// eslint-plugin-react's version util for simple `>= X` checks.
func parseVersion(s string) (int, int, int) {
	var parts [3]int
	i := 0
	for _, seg := range strings.Split(s, ".") {
		if i >= 3 {
			break
		}
		// Strip a trailing non-digit tail (e.g. "-rc.1", "+build").
		cut := len(seg)
		for j, c := range seg {
			if c < '0' || c > '9' {
				cut = j
				break
			}
		}
		n, err := strconv.Atoi(seg[:cut])
		if err == nil {
			parts[i] = n
		}
		i++
	}
	return parts[0], parts[1], parts[2]
}

// versionActive reports whether the configured React version is greater-or-
// equal to `deprecVersion` — i.e. whether the deprecation is "in effect" for
// this project. An absent `settings.react.version` defaults to 999.999.999
// (matching upstream's "latest"), so every deprecation fires by default.
func versionActive(settings map[string]interface{}, deprecVersion string) bool {
	major, minor, patch := parseVersion(deprecVersion)
	return !reactutil.ReactVersionLessThan(settings, major, minor, patch)
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
	var segs []string
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

var NoDeprecatedRule = rule.Rule{
	Name: "react/no-deprecated",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Determine pragma: `@jsx` directive in source wins over
		// `settings.react.pragma`, matching upstream's `pragmaUtil.getFromContext`.
		pragma := detectJsxPragma(ctx.SourceFile.Text())
		if pragma == "" {
			pragma = reactutil.GetReactPragma(ctx.Settings)
		}
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		deprecated := buildDeprecated(pragma)

		report := func(node *ast.Node, methodName string, d deprecationInfo) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "deprecated",
				Description: formatMessage(methodName, d),
			})
		}

		// check is the common gate — present only when the key is both in the
		// table and active for the configured React version.
		check := func(node *ast.Node, methodName string) {
			d, ok := deprecated[methodName]
			if !ok {
				return
			}
			if !versionActive(ctx.Settings, d.version) {
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
