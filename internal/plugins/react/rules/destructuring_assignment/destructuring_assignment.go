package destructuring_assignment

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// sfcParam mirrors upstream's `evalParams` output: one entry per SFC parameter
// position. Either `destructuring` is true (param is an `ObjectPattern`,
// e.g. `({a, b})`) or `name` is non-empty (param is an `Identifier`,
// e.g. `props`). For other shapes (rest, default, identifier-with-default
// pattern) we leave both zero — upstream's `param.type === 'Identifier'`
// equality also misses those.
//
// `symbol` carries the TypeChecker-resolved Symbol for the parameter binding,
// when a TypeChecker is available. handleSFCUsage uses it to verify that an
// Identifier with a name-matching `obj.Text == propsName` actually refers to
// THIS SFC's parameter — without it, an inner `let props = …` shadow would
// be misreported as a missed destructure (upstream has this same false
// positive; rslint resolves it when a TypeChecker is present).
type sfcParam struct {
	destructuring bool
	name          string
	symbol        *ast.Symbol
}

// sfcParamsStack mirrors upstream's `createSFCParams()` queue: an
// unshift-on-enter / shift-on-exit stack of the current SFC chain's param
// shapes. `propsName` / `contextName` walk inner-to-outer for the first
// non-destructuring identifier-named param at position 0 / 1 respectively.
type sfcParamsStack struct {
	queue [][]sfcParam
}

func (s *sfcParamsStack) push(params []sfcParam) {
	s.queue = append([][]sfcParam{params}, s.queue...)
}

func (s *sfcParamsStack) pop() {
	if len(s.queue) > 0 {
		s.queue = s.queue[1:]
	}
}

func (s *sfcParamsStack) propsName() string {
	for _, p := range s.queue {
		if len(p) > 0 && !p[0].destructuring && p[0].name != "" {
			return p[0].name
		}
	}
	return ""
}

// propsSymbol returns the TypeChecker Symbol of the active props parameter,
// matching the same stack entry that `propsName` selects. Returns nil when no
// TypeChecker was available at push time, or when the entry's parameter shape
// doesn't carry a Symbol.
func (s *sfcParamsStack) propsSymbol() *ast.Symbol {
	for _, p := range s.queue {
		if len(p) > 0 && !p[0].destructuring && p[0].name != "" {
			return p[0].symbol
		}
	}
	return nil
}

func (s *sfcParamsStack) contextName() string {
	for _, p := range s.queue {
		if len(p) > 1 && !p[1].destructuring && p[1].name != "" {
			return p[1].name
		}
	}
	return ""
}

func (s *sfcParamsStack) contextSymbol() *ast.Symbol {
	for _, p := range s.queue {
		if len(p) > 1 && !p[1].destructuring && p[1].name != "" {
			return p[1].symbol
		}
	}
	return nil
}

// evalParams maps a function's parameter list to upstream's
// `evalParams(node.params)` output. Only top-level shapes the rule cares about
// are recorded — `KindObjectBindingPattern` for `{a, b}` and `KindIdentifier`
// for a named parameter. Other shapes (rest element, array binding, parameters
// with type-only annotations) leave the entry zero, matching upstream's
// `param.type === 'Identifier'` / `param.type === 'ObjectPattern'` strict
// equality on the raw param node.
//
// Default-valued parameters are skipped to mirror upstream. ESTree wraps
// `(props = {})` and `({id} = {})` in an `AssignmentPattern`, whose
// `param.type` matches neither `'Identifier'` nor `'ObjectPattern'`, so
// upstream silently leaves both `destructuring` and `name` falsy. tsgo
// represents the same shapes as a `ParameterDeclaration` with a non-nil
// `Initializer`, so we explicitly skip that case to stay aligned.
//
// Rest parameters are intentionally skipped — `function Foo(...rest)` binds
// an array, and ESTree's `RestElement.type` likewise fails the strict
// equality checks above.
//
// When `tc` is non-nil, each Identifier-named parameter additionally carries
// its TypeChecker-resolved Symbol so handleSFCUsage can verify that a
// name-matching reference actually resolves to *this* parameter and not an
// inner shadow.
func evalParams(params []*ast.Node, tc *checker.Checker) []sfcParam {
	out := make([]sfcParam, len(params))
	for i, p := range params {
		if p == nil || p.Kind != ast.KindParameter {
			continue
		}
		pd := p.AsParameterDeclaration()
		if pd.DotDotDotToken != nil || pd.Initializer != nil {
			continue
		}
		name := pd.Name()
		if name == nil {
			continue
		}
		switch name.Kind {
		case ast.KindObjectBindingPattern:
			out[i].destructuring = true
		case ast.KindIdentifier:
			out[i].name = name.AsIdentifier().Text
			if tc != nil {
				out[i].symbol = tc.GetSymbolAtLocation(name)
			}
		}
	}
	return out
}

// getEnclosingSFCComponent mirrors upstream's
// `components.get(getScope(context, node).block)` semantics — the
// **enclosing-only** check used in `VariableDeclarator` listeners.
//
// Unlike `reactutil.GetParentStatelessComponent` (which walks the entire
// ancestor chain looking for any SFC), this helper inspects only the
// nearest enclosing FunctionLike scope. If that scope is not classified
// as an SFC, returns nil — even if a further-out ancestor happens to be
// one. This matters for `const {x} = props` inside an inner non-SFC
// helper of an outer SFC: upstream's `scope.block` is the inner helper,
// `components.get(inner)` is undefined, and the rule stays silent.
func getEnclosingSFCComponent(node *ast.Node, pragma string, wrappers []reactutil.ComponentWrapperEntry) *ast.Node {
	for p := node.Parent; p != nil; p = p.Parent {
		if !ast.IsFunctionLike(p) {
			continue
		}
		if reactutil.IsStatelessReactComponentWithWrappers(p, pragma, nil, wrappers) {
			return p
		}
		return nil
	}
	return nil
}

// isAssignmentLHS mirrors upstream's `isAssignmentLHS(node)`: true when `node`
// is the left-hand side of a BinaryExpression with an assignment operator
// (`=`, `+=`, etc.). Used to suppress `useDestructAssignment` reports on
// `props.x = …` — the access there is a write target, not a read.
func isAssignmentLHS(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := parent.AsBinaryExpression()
	if bin.OperatorToken == nil || !ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
		return false
	}
	left := bin.Left
	for left != nil && left.Kind == ast.KindParenthesizedExpression {
		left = left.AsParenthesizedExpression().Expression
	}
	target := node
	for target != nil && target.Kind == ast.KindParenthesizedExpression {
		target = target.AsParenthesizedExpression().Expression
	}
	return left == target
}

// isOptionalMember mirrors ESTree's `node.optional` flag for member
// expressions. tsgo encodes `foo?.bar` via the `QuestionDotToken` field on
// PropertyAccessExpression / ElementAccessExpression — there is no
// `ChainExpression` wrapper.
func isOptionalMember(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().QuestionDotToken != nil
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().QuestionDotToken != nil
	}
	return false
}

// isInClassProperty walks up looking for a class field (PropertyDeclaration
// in tsgo, which collapses ESTree's `ClassProperty` and `PropertyDefinition`
// into one kind). Mirrors upstream's `isInClassProperty(node)`.
func isInClassProperty(node *ast.Node) bool {
	for p := node.Parent; p != nil; p = p.Parent {
		if p.Kind == ast.KindPropertyDeclaration {
			return true
		}
	}
	return false
}

// isParentClassProperty mirrors upstream's
// `node.parent.type === 'ClassProperty' || 'PropertyDefinition'` test on the
// VariableDeclarator. For `node` = VariableDeclarator, `node.parent` is a
// VariableDeclaration, never a class field — so this guard is dead code in
// upstream and always evaluates to false. We preserve that exact behavior
// here so `ignoreClassFields` does NOT suppress `const {x} = this.props`
// reports under `'never'`. (Class-field-internal IIFE cases reach this via
// `isInClassProperty` from `handleClassUsage`, which IS a real ancestor
// walk; that path is unchanged.)
func isParentClassProperty(_ *ast.Node) bool {
	return false
}

// findEnclosingTypeQuery mirrors upstream's `findParent(n, n.type === 'TSTypeQuery')`.
// Returns true when any ancestor is a `typeof T` type query.
func findEnclosingTypeQuery(node *ast.Node) bool {
	for p := node.Parent; p != nil; p = p.Parent {
		if p.Kind == ast.KindTypeQuery {
			return true
		}
	}
	return false
}

// countPropsRefsExcludingDecl approximates upstream's
// `getScope(context, node).set.get('props').references.length` for the
// `destructureInSignature: 'always'` gate.
//
// One up-front decision selects the counting strategy:
//
//   - **Symbol path** (TypeChecker + parameter Symbol both resolved):
//     count Identifiers whose Symbol matches the SFC parameter. Inner
//     `let props = …` shadows have a different Symbol and are excluded —
//     scope-aware, matching ESLint's scopeManager exactly.
//
//   - **Name path** (anything required for the Symbol path missing):
//     count every `props` Identifier that isn't an obvious binding /
//     property name. Over-counts on shadows → conservative no-report.
//
// The decision happens once below; the walker calls `shouldCount` directly
// without re-testing the TypeChecker on each Identifier.
func countPropsRefsExcludingDecl(fn *ast.Node, exclude *ast.Node, tc *checker.Checker, paramName *ast.Node) int {
	body := reactutil.FunctionBody(fn)
	if body == nil {
		return 0
	}

	var paramSymbol *ast.Symbol
	if tc != nil && paramName != nil && paramName.Kind == ast.KindIdentifier {
		paramSymbol = tc.GetSymbolAtLocation(paramName)
	}

	var shouldCount func(n *ast.Node) bool
	if paramSymbol != nil {
		// Symbol path — `tc` is guaranteed non-nil because paramSymbol was
		// only assigned when `tc != nil`.
		shouldCount = func(n *ast.Node) bool {
			sym := tc.GetSymbolAtLocation(n)
			return sym != nil && sym == paramSymbol
		}
	} else {
		shouldCount = isPropsReference
	}

	count := 0
	var walk func(n *ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n == nil || n == exclude {
			return false
		}
		if n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == "props" && shouldCount(n) {
			count++
		}
		n.ForEachChild(walk)
		return false
	}
	body.ForEachChild(walk)
	return count
}

// isPropsReference is the no-checker fallback predicate for
// `countPropsRefsExcludingDecl`. Returns true when the Identifier `n` is
// *probably* a value reference (not a binding name, member key, label, etc.).
// Without a TypeChecker we can't tell whether the reference resolves to the
// outer SFC's `props` or an inner shadow — we conservatively count every
// suspect occurrence so the rule errs on "don't autofix" in ambiguous cases.
func isPropsReference(n *ast.Node) bool {
	parent := n.Parent
	if parent == nil {
		return true
	}
	switch parent.Kind {
	case ast.KindParameter:
		if parent.AsParameterDeclaration().Name() == n {
			return false
		}
	case ast.KindBindingElement:
		if parent.AsBindingElement().Name() == n {
			return false
		}
	case ast.KindVariableDeclaration:
		if parent.AsVariableDeclaration().Name() == n {
			return false
		}
	case ast.KindPropertyAccessExpression:
		if parent.AsPropertyAccessExpression().Name() == n {
			return false
		}
	case ast.KindQualifiedName:
		if parent.AsQualifiedName().Right == n {
			return false
		}
	case ast.KindPropertyAssignment:
		if parent.Name() == n {
			return false
		}
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindPropertyDeclaration:
		if parent.Name() == n {
			return false
		}
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindClassDeclaration, ast.KindClassExpression:
		if parent.Name() == n {
			return false
		}
	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement:
		return false
	}
	return true
}

type ruleOptions struct {
	configuration          string
	ignoreClassFields      bool
	destructureInSignature string
}

// buildSFCMatcher returns a `(receiver, paramSymbol) → bool` closure used by
// handleSFCUsage. The closure is constructed ONCE per rule invocation based
// on whether a TypeChecker is available:
//
//   - `tc == nil` → returns a closure that always answers true. The
//     caller's name comparison alone gates the report — equivalent to
//     upstream's name-only behavior. The closure has no reference to `tc`,
//     so a nil checker can never be dereferenced downstream.
//
//   - `tc != nil` → returns a closure that compares Symbols. A
//     name-matching but scope-shadowed reference is correctly rejected.
//     `paramSymbol == nil` (Symbol resolution failed at push time) falls
//     back to "accept" so a transient resolver miss never suppresses a
//     legitimate report.
func buildSFCMatcher(tc *checker.Checker) func(obj *ast.Node, paramSymbol *ast.Symbol) bool {
	if tc == nil {
		return func(_ *ast.Node, _ *ast.Symbol) bool { return true }
	}
	return func(obj *ast.Node, paramSymbol *ast.Symbol) bool {
		if paramSymbol == nil {
			return true
		}
		sym := tc.GetSymbolAtLocation(obj)
		return sym != nil && sym == paramSymbol
	}
}

// parseOptions handles both the array-wrapped shape the rule_tester emits
// (`["always", {...}]`) and the bare-string / bare-object shapes the CLI
// can deliver after `internal/config/config.go` unwraps a single-option
// array. Defaults: `configuration="always"`, `ignoreClassFields=false`,
// `destructureInSignature="ignore"` — matching upstream.
func parseOptions(options any) ruleOptions {
	opts := ruleOptions{
		configuration:          "always",
		ignoreClassFields:      false,
		destructureInSignature: "ignore",
	}
	if options == nil {
		return opts
	}
	var arr []interface{}
	switch v := options.(type) {
	case []interface{}:
		arr = v
	case string:
		opts.configuration = v
		return opts
	default:
		return opts
	}
	if len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			opts.configuration = s
		}
	}
	if len(arr) > 1 {
		if m, ok := arr[1].(map[string]interface{}); ok {
			if v, ok := m["ignoreClassFields"].(bool); ok {
				opts.ignoreClassFields = v
			}
			if v, ok := m["destructureInSignature"].(string); ok {
				opts.destructureInSignature = v
			}
		}
	}
	return opts
}

var DestructuringAssignmentRule = rule.Rule{
	Name: "react/destructuring-assignment",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
		stack := &sfcParamsStack{}

		reportUseDestruct := func(node *ast.Node, t string) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "useDestructAssignment",
				Description: "Must use destructuring " + t + " assignment",
				Data:        map[string]string{"type": t},
			})
		}

		reportNoDestruct := func(node *ast.Node, t string) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noDestructAssignment",
				Description: "Must never use destructuring " + t + " assignment",
				Data:        map[string]string{"type": t},
			})
		}

		// handleStatelessComponent enters an SFC: push its params onto the
		// stack so nested member-expression listeners can see the active
		// `props` / `context` names, then in `never` mode emit the
		// destructured-arg diagnostic.
		handleStatelessComponent := func(node *ast.Node) {
			if !reactutil.IsStatelessReactComponentWithWrappers(node, pragma, nil, wrappers) {
				return
			}
			params := evalParams(reactutil.FunctionParameters(node), ctx.TypeChecker)
			stack.push(params)
			if opts.configuration != "never" {
				return
			}
			if len(params) > 0 && params[0].destructuring {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noDestructPropsInSFCArg",
					Description: "Must never use destructuring props assignment in SFC argument",
				})
			} else if len(params) > 1 && params[1].destructuring {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noDestructContextInSFCArg",
					Description: "Must never use destructuring context assignment in SFC argument",
				})
			}
		}

		handleStatelessComponentExit := func(node *ast.Node) {
			if !reactutil.IsStatelessReactComponentWithWrappers(node, pragma, nil, wrappers) {
				return
			}
			stack.pop()
		}

		// matchesSFCParam returns true when the receiver Identifier `obj`
		// (name-matched by the caller) actually refers to the SFC parameter
		// recorded with `paramSymbol`. The closure is built ONCE based on
		// whether a TypeChecker is wired up:
		//
		//   - **No TypeChecker**: closure always returns true. Only the
		//     caller's name comparison gates the report — equivalent to
		//     upstream's name-only behavior. The closure never touches
		//     `ctx.TypeChecker`, so nil-deref is impossible.
		//
		//   - **TypeChecker present**: closure compares Symbols. A
		//     name-matching but scope-shadowed reference is correctly
		//     rejected. `paramSymbol == nil` (Symbol resolution failed at
		//     push time) still falls back to "accept" so the rule never
		//     loses a legitimate report due to a transient resolver miss.
		matchesSFCParam := buildSFCMatcher(ctx.TypeChecker)

		// handleSFCUsage reports `props.X` / `context.X` accesses on `always`
		// mode. The receiver must be a bare Identifier matching the active
		// `propsName` / `contextName`; `(props).x` is unwrapped via
		// SkipParentheses (tsgo preserves what ESTree flattens).
		// Optional-chain accesses (`props?.x`) do NOT trigger — upstream's
		// `!node.optional` guard.
		handleSFCUsage := func(node *ast.Node) {
			propsName := stack.propsName()
			contextName := stack.contextName()
			var objNode *ast.Node
			switch node.Kind {
			case ast.KindPropertyAccessExpression:
				objNode = node.AsPropertyAccessExpression().Expression
			case ast.KindElementAccessExpression:
				objNode = node.AsElementAccessExpression().Expression
			default:
				return
			}
			obj := ast.SkipParentheses(objNode)
			if obj.Kind != ast.KindIdentifier {
				return
			}
			objName := obj.AsIdentifier().Text

			matched := false
			if propsName != "" && objName == propsName &&
				matchesSFCParam(obj, stack.propsSymbol()) {
				matched = true
			} else if contextName != "" && objName == contextName &&
				matchesSFCParam(obj, stack.contextSymbol()) {
				matched = true
			}
			if !matched {
				return
			}
			if isAssignmentLHS(node) {
				return
			}
			if opts.configuration != "always" || isOptionalMember(node) {
				return
			}
			reportUseDestruct(node, objName)
		}

		// handleClassUsage reports `this.props.X` / `this.state.X` /
		// `this.context.X` accesses on `always` mode. The receiver chain must
		// match exactly: outer member's object is itself a member with `this`
		// as its receiver and one of the three known property names. Other
		// `this.X.Y` shapes are left alone.
		handleClassUsage := func(node *ast.Node) {
			var objNode *ast.Node
			switch node.Kind {
			case ast.KindPropertyAccessExpression:
				objNode = node.AsPropertyAccessExpression().Expression
			case ast.KindElementAccessExpression:
				objNode = node.AsElementAccessExpression().Expression
			default:
				return
			}
			obj := ast.SkipParentheses(objNode)
			if obj.Kind != ast.KindPropertyAccessExpression {
				return
			}
			inner := obj.AsPropertyAccessExpression()
			if ast.SkipParentheses(inner.Expression).Kind != ast.KindThisKeyword {
				return
			}
			name := reactutil.EsTreeName(inner.Name())
			if name != "props" && name != "state" && name != "context" {
				return
			}
			if isAssignmentLHS(node) {
				return
			}
			if opts.configuration != "always" {
				return
			}
			if opts.ignoreClassFields && isInClassProperty(node) {
				return
			}
			reportUseDestruct(node, name)
		}

		memberExprListener := func(node *ast.Node) {
			if reactutil.GetParentStatelessComponent(node, pragma, wrappers) != nil {
				handleSFCUsage(node)
			}
			if reactutil.GetParentReactComponentScopeBasedOrStateless(node, pragma, createClass, wrappers) != nil {
				handleClassUsage(node)
			}
		}

		return rule.RuleListeners{
			// MethodDeclaration is included so object-literal shorthand
			// methods (`{ Foo(props) {...} }`) participate in the SFC stack
			// — ESTree wraps these as FunctionExpression, tsgo doesn't.
			// `IsStatelessReactComponent`'s MethodDeclaration arm then
			// classifies them when the parent is ObjectLiteralExpression and
			// the key is a capitalized Identifier returning JSX.
			ast.KindFunctionDeclaration:                      handleStatelessComponent,
			ast.KindFunctionExpression:                       handleStatelessComponent,
			ast.KindArrowFunction:                            handleStatelessComponent,
			ast.KindMethodDeclaration:                        handleStatelessComponent,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): handleStatelessComponentExit,
			rule.ListenerOnExit(ast.KindFunctionExpression):  handleStatelessComponentExit,
			rule.ListenerOnExit(ast.KindArrowFunction):       handleStatelessComponentExit,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   handleStatelessComponentExit,

			ast.KindPropertyAccessExpression: memberExprListener,
			ast.KindElementAccessExpression:  memberExprListener,

			// Upstream `TSQualifiedName` listener: report `typeof props.X` in
			// SFC `always` mode. Only the leftmost QualifiedName in a chain
			// matches (its `Left` is an Identifier); outer wrappers in
			// `props.a.b` have a QualifiedName as `Left` and are skipped.
			ast.KindQualifiedName: func(node *ast.Node) {
				if opts.configuration != "always" {
					return
				}
				qn := node.AsQualifiedName()
				if qn.Left == nil || qn.Left.Kind != ast.KindIdentifier {
					return
				}
				propsName := stack.propsName()
				if propsName == "" || qn.Left.AsIdentifier().Text != propsName {
					return
				}
				if !findEnclosingTypeQuery(node) {
					return
				}
				if reactutil.GetParentStatelessComponent(node, pragma, wrappers) == nil {
					return
				}
				reportUseDestruct(node, "props")
			},

			// Upstream `VariableDeclarator` listener: covers three reports:
			//   - `never` mode SFC `const {a} = props|context`
			//   - `never` mode class `const {a} = this.props|state|context`
			//   - `always` + `destructureInSignature: 'always'` SFC props
			ast.KindVariableDeclaration: func(node *ast.Node) {
				vd := node.AsVariableDeclaration()
				if vd.Initializer == nil {
					return
				}
				name := vd.Name()
				if name == nil || name.Kind != ast.KindObjectBindingPattern {
					return
				}
				init := ast.SkipParentheses(vd.Initializer)

				var sfcType string
				var classType string
				switch init.Kind {
				case ast.KindIdentifier:
					nm := init.AsIdentifier().Text
					if nm == "props" || nm == "context" {
						sfcType = nm
					}
				case ast.KindPropertyAccessExpression:
					pa := init.AsPropertyAccessExpression()
					if ast.SkipParentheses(pa.Expression).Kind == ast.KindThisKeyword {
						propName := reactutil.EsTreeName(pa.Name())
						if propName == "props" || propName == "context" || propName == "state" {
							classType = propName
						}
					}
				}

				// Mirror upstream's `components.get(getScope(context, node).block)`:
				// enclosing-only semantics. A `const {x} = props` inside an
				// inner non-SFC helper of an outer SFC must NOT report —
				// upstream's `scope.block` is the inner helper, `components.
				// get(inner) === undefined`, and the rule stays silent.
				// `GetParentStatelessComponent` (ancestor walk) would
				// over-report here.
				sfcComp := getEnclosingSFCComponent(node, pragma, wrappers)
				// Mirror upstream's `utils.getParentComponent(node)` =
				// `getParentES6Component || getParentES5Component
				// || getParentStatelessComponent`. The OrStateless tail
				// matters: when a SFC body contains `const {x} = this.props`
				// (semantically nonsensical but legal syntax), upstream's
				// `classComponent` falls back to the SFC and reports under
				// `'never'`. Using only `GetEnclosingReactComponent` (class-
				// only) would silently drop that report.
				classComp := reactutil.GetParentReactComponentScopeBasedOrStateless(node, pragma, createClass, wrappers)

				if opts.configuration == "never" {
					if sfcComp != nil && sfcType != "" {
						reportNoDestruct(node, sfcType)
					}
					if classComp != nil && classType != "" {
						if !opts.ignoreClassFields || !isParentClassProperty(node) {
							reportNoDestruct(node, classType)
						}
					}
				}

				if sfcComp != nil &&
					sfcType == "props" &&
					opts.configuration == "always" &&
					opts.destructureInSignature == "always" {
					params := reactutil.FunctionParameters(sfcComp)
					if len(params) == 0 {
						return
					}
					param := params[0]
					if param == nil || param.Kind != ast.KindParameter {
						return
					}
					pd := param.AsParameterDeclaration()
					paramName := pd.Name()
					// SFC's first parameter must be an Identifier literally
					// named `props` for this gate to apply. Mirrors upstream's
					// `getScope(node).set.get('props')` lookup, which only
					// succeeds when `props` is bound in the enclosing function
					// scope. Without this guard, an SFC whose parameter is
					// renamed (e.g. `function Foo(myProps)`) but whose body
					// references an outer `props` would trigger an autofix
					// that rewrites the parameter and silently changes the
					// reference target.
					if paramName == nil || paramName.Kind != ast.KindIdentifier ||
						paramName.AsIdentifier().Text != "props" {
						return
					}
					// Use TypeChecker Symbol comparison when available so
					// inner `let props = …` shadows are correctly excluded;
					// otherwise fall back to a name+parent-shape walk.
					if countPropsRefsExcludingDecl(sfcComp, vd.Initializer, ctx.TypeChecker, paramName) > 0 {
						return
					}

					// Replace the parameter binding name span with the
					// destructure pattern. The trimmed range is identical to
					// upstream's `[param.range[0], param.typeAnnotation
					// ? param.typeAnnotation.range[0] : param.range[1]]` —
					// when a type annotation exists, ESTree's `param.range[1]`
					// includes it, while tsgo's `paramName.End()` is the bare
					// identifier end, so this branch falls out naturally.
					replaceRange := utils.TrimNodeTextRange(ctx.SourceFile, paramName)
					patternText := utils.TrimmedNodeText(ctx.SourceFile, name)

					// Remove the surrounding VariableStatement so the body
					// loses the `const {a} = props;` line. tsgo's
					// VariableDeclaration parent chain is
					// VariableDeclarationList → VariableStatement, while
					// ESTree's VariableDeclarator parent is VariableDeclaration
					// (which carries the trailing semicolon directly).
					removeTarget := node
					if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclarationList {
						list := node.Parent
						if list.Parent != nil && list.Parent.Kind == ast.KindVariableStatement {
							removeTarget = list.Parent
						}
					}
					ctx.ReportNodeWithFixes(node, rule.RuleMessage{
						Id:          "destructureInSignature",
						Description: "Must destructure props in the function signature.",
					},
						rule.RuleFixReplaceRange(replaceRange, patternText),
						rule.RuleFixRemove(ctx.SourceFile, removeTarget),
					)
				}
			},
		}
	},
}
