// cspell:ignore unshadowed logicals recognises pname
package no_implied_eval

import (
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// This file implements a self-recursive static fold that mirrors ESLint's
// `getStaticValue` from `@eslint-community/eslint-utils` for the node shapes
// the `no-implied-eval` rule relies on. It is intentionally NOT built on
// tsgo's `shim/evaluator`, because that evaluator's recursion is a private
// closure — supplementing it at the top level only would leave deeply nested
// kinds (e.g. ConditionalExpression inside BinaryExpression '+') unhandled.
// A self-recursive fold handles arbitrary nesting.

// nullVal / undefVal are sentinels distinguishing JS null / undefined values
// from an unresolved fold. `foldResult{}` (value=nil) means unresolved.
// arrayVal / objectVal mark resolved composite values — their AST nodes are
// kept for property lookup (objects) and for method-call gating (arrays).
type nullVal struct{}
type undefVal struct{}
type arrayVal struct{}
type objectVal struct{ node *ast.Node }

// foldResult is the outcome of a static fold. value==nil means unresolved.
type foldResult struct {
	value any
}

func (r foldResult) resolved() bool      { return r.value != nil }
func (r foldResult) isStringValue() bool { _, ok := r.value.(string); return ok }

func (r foldResult) truthy() bool {
	switch v := r.value.(type) {
	case string:
		return len(v) != 0
	case float64:
		return v != 0 && !math.IsNaN(v)
	case bool:
		return v
	case nullVal, undefVal:
		return false
	case arrayVal, objectVal:
		// Objects and arrays are truthy in JS (including empty ones).
		return true
	}
	return false
}

func (r foldResult) nullish() bool {
	switch r.value.(type) {
	case nullVal, undefVal:
		return true
	}
	return false
}

// asString implements JS ToString for the resolved value. Used when a String()
// call or binary '+' coerces a non-string operand to a string.
func (r foldResult) asString() string {
	switch v := r.value.(type) {
	case string:
		return v
	case float64:
		return jsNumberToString(v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nullVal:
		return "null"
	case undefVal:
		return "undefined"
	case arrayVal:
		// Exact join of element strings would require carrying the array's
		// content; we only classify, so any non-empty placeholder keeps the
		// binary '+' string-concat branch triggered.
		return "<array>"
	case objectVal:
		return "[object Object]"
	}
	return ""
}

// jsNumberToString produces a JS-like string for a float64. Exact parity with
// ECMAScript ToString(Number) isn't required — no-implied-eval only cares
// whether the overall expression classifies as string, not the specific digits.
func jsNumberToString(v float64) string {
	switch {
	case math.IsNaN(v):
		return "NaN"
	case math.IsInf(v, 1):
		return "Infinity"
	case math.IsInf(v, -1):
		return "-Infinity"
	case v == 0:
		return "0"
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// strCtx carries per-rule-invocation state for the static fold. A single
// instance is reused by all listener callbacks within one file.
type strCtx struct {
	ruleCtx rule.RuleContext

	writeRefsComputed bool
	writeRefsMap      map[*ast.Symbol]bool
}

func newStrCtx(ctx rule.RuleContext) *strCtx {
	return &strCtx{ruleCtx: ctx}
}

// isString reports whether `node` is known — syntactically or via static
// fold — to evaluate to a string. Union of ESLint's `isEvaluatedString`
// (syntactic shape) and `getStaticValue` with scope (semantic fold).
func (s *strCtx) isString(node *ast.Node) bool {
	if isEvaluatedString(node) {
		return true
	}
	return s.fold(node).isStringValue()
}

// isEvaluatedString is a purely syntactic check. Mirrors ESLint's
// `ast-utils.isEvaluatedString`: a StringLiteral / TemplateLiteral, or a '+'
// BinaryExpression where either side is syntactically a string. It fires on
// shapes the semantic fold may not resolve — e.g. `'x' + unknown`, where the
// overall concat is provably a string regardless of `unknown`.
func isEvaluatedString(node *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, argOuterKinds)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		return true
	case ast.KindBinaryExpression:
		be := node.AsBinaryExpression()
		if be != nil && be.OperatorToken != nil && be.OperatorToken.Kind == ast.KindPlusToken {
			return isEvaluatedString(be.Left) || isEvaluatedString(be.Right)
		}
	}
	return false
}

// fold recursively evaluates `node` to a static value, handling arbitrary
// nesting of the shapes ESLint's `getStaticValue` covers for this rule:
//
//   - Literals: string / template / numeric / true / false / null
//   - Identifier: const / let-no-writes / var-no-writes → initializer
//   - `undefined` global (when unshadowed) → undefined
//   - BinaryExpression: '+' (string or numeric concat), arithmetic, bitwise,
//     logical `||` `&&` `??` (short-circuit)
//   - PrefixUnaryExpression: '+' '-' '~' '!'
//   - ConditionalExpression: `c ? a : b` (picks the branch once `c` resolves)
//   - TypeOfExpression: always yields a string if the operand resolves
//   - VoidExpression: always undefined
//   - CallExpression: `String(x)` / `String()` with resolvable arg
//   - TaggedTemplateExpression: `String.raw`…``  with resolvable subs
//   - TemplateExpression: substitution folding
//
// Returns foldResult{} (value==nil) when the node can't be statically resolved.
func (s *strCtx) fold(node *ast.Node) foldResult {
	node = ast.SkipOuterExpressions(node, argOuterKinds)
	if node == nil {
		return foldResult{}
	}

	switch node.Kind {
	case ast.KindStringLiteral:
		return foldResult{node.AsStringLiteral().Text}
	case ast.KindNoSubstitutionTemplateLiteral:
		return foldResult{node.AsNoSubstitutionTemplateLiteral().Text}
	case ast.KindNumericLiteral:
		n, err := strconv.ParseFloat(node.AsNumericLiteral().Text, 64)
		if err != nil {
			return foldResult{}
		}
		return foldResult{n}
	case ast.KindTrueKeyword:
		return foldResult{true}
	case ast.KindFalseKeyword:
		return foldResult{false}
	case ast.KindNullKeyword:
		return foldResult{nullVal{}}
	case ast.KindIdentifier:
		id := node.AsIdentifier()
		if id.Text == "undefined" && !utils.IsShadowed(node, "undefined") {
			return foldResult{undefVal{}}
		}
		return s.foldIdentifier(node)
	case ast.KindTemplateExpression:
		return s.foldTemplate(node)
	case ast.KindBinaryExpression:
		return s.foldBinary(node)
	case ast.KindPrefixUnaryExpression:
		return s.foldPrefixUnary(node)
	case ast.KindConditionalExpression:
		return s.foldConditional(node)
	case ast.KindTypeOfExpression:
		return s.foldTypeOf(node)
	case ast.KindVoidExpression:
		// `void X` always evaluates to undefined regardless of X.
		return foldResult{undefVal{}}
	case ast.KindCallExpression:
		return s.foldCall(node)
	case ast.KindTaggedTemplateExpression:
		return s.foldStringRawTag(node)
	case ast.KindArrayLiteralExpression:
		return s.foldArrayLiteral(node)
	case ast.KindObjectLiteralExpression:
		return foldResult{objectVal{node: node}}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return s.foldMemberAccess(node)
	}
	return foldResult{}
}

// foldIdentifier resolves an Identifier to its initializer value, matching
// ESLint's scope-manager behavior: const is always eligible; let/var require
// no non-initial write references anywhere in the file.
func (s *strCtx) foldIdentifier(node *ast.Node) foldResult {
	if s.ruleCtx.TypeChecker == nil {
		return foldResult{}
	}
	// GetReferenceSymbol handles the shorthand-property edge case where the
	// identifier is the key of `{ foo }` — there the plain GetSymbolAtLocation
	// returns the property symbol, not the outer binding we want to resolve.
	sym := utils.GetReferenceSymbol(node, s.ruleCtx.TypeChecker)
	if sym == nil || len(sym.Declarations) != 1 {
		return foldResult{}
	}
	decl := sym.Declarations[0]
	if decl.Kind != ast.KindVariableDeclaration {
		return foldResult{}
	}
	varDecl := decl.AsVariableDeclaration()
	if varDecl == nil || varDecl.Initializer == nil {
		return foldResult{}
	}
	list := decl.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return foldResult{}
	}
	// const: always resolvable. let/var: require no writes after init.
	if list.Flags&ast.NodeFlagsConst == 0 && s.hasWrites(sym) {
		return foldResult{}
	}
	return s.fold(varDecl.Initializer)
}

// foldTemplate folds a TemplateExpression by concatenating head + each
// span's expression (recursively folded) + span's literal tail. Returns
// unresolved if any span can't be folded.
func (s *strCtx) foldTemplate(node *ast.Node) foldResult {
	te := node.AsTemplateExpression()
	if te == nil {
		return foldResult{}
	}
	var sb strings.Builder
	if te.Head != nil {
		sb.WriteString(te.Head.Text())
	}
	if te.TemplateSpans != nil {
		for _, span := range te.TemplateSpans.Nodes {
			sp := span.AsTemplateSpan()
			if sp == nil {
				return foldResult{}
			}
			sub := s.fold(sp.Expression)
			if !sub.resolved() {
				return foldResult{}
			}
			sb.WriteString(sub.asString())
			if sp.Literal != nil {
				sb.WriteString(sp.Literal.Text())
			}
		}
	}
	return foldResult{sb.String()}
}

// foldBinary handles both short-circuit logicals (with partial evaluation)
// and arithmetic / string-concat (requires both sides resolved).
func (s *strCtx) foldBinary(node *ast.Node) foldResult {
	be := node.AsBinaryExpression()
	if be == nil || be.OperatorToken == nil {
		return foldResult{}
	}
	op := be.OperatorToken.Kind

	// Short-circuit: we only need the left side to decide which operand the
	// overall expression evaluates to.
	switch op {
	case ast.KindBarBarToken:
		left := s.fold(be.Left)
		if !left.resolved() {
			return foldResult{}
		}
		if left.truthy() {
			return left
		}
		return s.fold(be.Right)
	case ast.KindAmpersandAmpersandToken:
		left := s.fold(be.Left)
		if !left.resolved() {
			return foldResult{}
		}
		if !left.truthy() {
			return left
		}
		return s.fold(be.Right)
	case ast.KindQuestionQuestionToken:
		left := s.fold(be.Left)
		if !left.resolved() {
			return foldResult{}
		}
		if left.nullish() {
			return s.fold(be.Right)
		}
		return left
	}

	// Arithmetic / concat: need both sides.
	left := s.fold(be.Left)
	right := s.fold(be.Right)
	if !left.resolved() || !right.resolved() {
		return foldResult{}
	}
	ln, lnIs := left.value.(float64)
	rn, rnIs := right.value.(float64)

	switch op {
	case ast.KindPlusToken:
		// Pure numeric add.
		if lnIs && rnIs {
			return foldResult{ln + rn}
		}
		// String concat: if either side is a string, stringify both.
		if _, lsIs := left.value.(string); lsIs {
			return foldResult{left.asString() + right.asString()}
		}
		if _, rsIs := right.value.(string); rsIs {
			return foldResult{left.asString() + right.asString()}
		}
	case ast.KindMinusToken:
		if lnIs && rnIs {
			return foldResult{ln - rn}
		}
	case ast.KindAsteriskToken:
		if lnIs && rnIs {
			return foldResult{ln * rn}
		}
	case ast.KindSlashToken:
		if lnIs && rnIs {
			return foldResult{ln / rn}
		}
	case ast.KindPercentToken:
		if lnIs && rnIs {
			return foldResult{math.Mod(ln, rn)}
		}
	case ast.KindAsteriskAsteriskToken:
		if lnIs && rnIs {
			return foldResult{math.Pow(ln, rn)}
		}
	case ast.KindBarToken:
		if lnIs && rnIs {
			return foldResult{float64(int32(ln) | int32(rn))}
		}
	case ast.KindAmpersandToken:
		if lnIs && rnIs {
			return foldResult{float64(int32(ln) & int32(rn))}
		}
	case ast.KindCaretToken:
		if lnIs && rnIs {
			return foldResult{float64(int32(ln) ^ int32(rn))}
		}
	case ast.KindLessThanLessThanToken:
		if lnIs && rnIs {
			return foldResult{float64(int32(ln) << (uint32(rn) & 31))}
		}
	case ast.KindGreaterThanGreaterThanToken:
		if lnIs && rnIs {
			return foldResult{float64(int32(ln) >> (uint32(rn) & 31))}
		}
	case ast.KindGreaterThanGreaterThanGreaterThanToken:
		if lnIs && rnIs {
			return foldResult{float64(uint32(ln) >> (uint32(rn) & 31))}
		}
	}
	return foldResult{}
}

func (s *strCtx) foldPrefixUnary(node *ast.Node) foldResult {
	pre := node.AsPrefixUnaryExpression()
	if pre == nil {
		return foldResult{}
	}
	operand := s.fold(pre.Operand)
	if !operand.resolved() {
		return foldResult{}
	}
	if n, ok := operand.value.(float64); ok {
		switch pre.Operator {
		case ast.KindPlusToken:
			return foldResult{n}
		case ast.KindMinusToken:
			return foldResult{-n}
		case ast.KindTildeToken:
			return foldResult{float64(^int32(n))}
		}
	}
	if pre.Operator == ast.KindExclamationToken {
		return foldResult{!operand.truthy()}
	}
	return foldResult{}
}

func (s *strCtx) foldConditional(node *ast.Node) foldResult {
	cond := node.AsConditionalExpression()
	if cond == nil {
		return foldResult{}
	}
	c := s.fold(cond.Condition)
	if !c.resolved() {
		return foldResult{}
	}
	if c.truthy() {
		return s.fold(cond.WhenTrue)
	}
	return s.fold(cond.WhenFalse)
}

// foldTypeOf matches ESLint's getStaticValue behavior: typeof is only folded
// when the operand itself resolves (otherwise the runtime type is
// unobservable — typeof of an undeclared identifier yields "undefined" but
// typeof of a resolvable binding yields its concrete type name).
func (s *strCtx) foldTypeOf(node *ast.Node) foldResult {
	tof := node.AsTypeOfExpression()
	if tof == nil {
		return foldResult{}
	}
	inner := s.fold(tof.Expression)
	if !inner.resolved() {
		return foldResult{}
	}
	switch inner.value.(type) {
	case string:
		return foldResult{"string"}
	case float64:
		return foldResult{"number"}
	case bool:
		return foldResult{"boolean"}
	case nullVal:
		return foldResult{"object"}
	case undefVal:
		return foldResult{"undefined"}
	}
	return foldResult{}
}

// foldCall dispatches CallExpressions to the handlers ESLint's getStaticValue
// recognises for this rule: `String(arg)` constructor, and prototype method
// calls in our whitelist. Everything else remains unresolved.
func (s *strCtx) foldCall(node *ast.Node) foldResult {
	call := node.AsCallExpression()
	if call == nil {
		return foldResult{}
	}
	callee := ast.SkipOuterExpressions(call.Expression, calleeOuterKinds)
	if callee == nil {
		return foldResult{}
	}
	// `String(arg)` — the global String constructor coerces any resolvable
	// value to its string form.
	if ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "String" && !utils.IsShadowed(callee, "String") {
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			return foldResult{""}
		}
		arg := s.fold(call.Arguments.Nodes[0])
		if !arg.resolved() {
			return foldResult{}
		}
		return foldResult{arg.asString()}
	}
	// `receiver.method(...)` — whitelisted prototype methods whose return type
	// is known (string or array), matching eslint-utils's getStaticValue.
	return s.foldMethodCall(call, callee)
}

// stringReturningMethods are the prototype methods that eslint-utils's
// getStaticValue folds to a string value. Discovered empirically against
// ESLint 10 (see /tmp/no-implied-eval-diff/method-scope.mjs and deep-scope.mjs).
// Excluded on purpose: repeat, replace, replaceAll, split, charCodeAt,
// codePointAt, indexOf, lastIndexOf, startsWith, endsWith, includes,
// toLocaleString — eslint-utils doesn't fold these and neither should we.
var stringReturningMethods = map[string]bool{
	// String.prototype
	"toString":    true,
	"toUpperCase": true,
	"toLowerCase": true,
	"trim":        true,
	"trimStart":   true,
	"trimEnd":     true,
	"concat":      true,
	"slice":       true,
	"substring":   true,
	"substr":      true,
	"padStart":    true,
	"padEnd":      true,
	"charAt":      true,
	"at":          true,
	"normalize":   true,
	// Number.prototype
	"toFixed":       true,
	"toExponential": true,
	"toPrecision":   true,
	// Array.prototype
	"join": true,
}

// arrayReturningMethods fold to an array value, enabling chains such as
// `[1,2].slice(0).toString()`.
var arrayReturningMethods = map[string]bool{
	"slice":   true,
	"concat":  true,
	"flat":    true,
	"flatMap": true,
	"reverse": true,
	"sort":    true,
}

// foldMethodCall handles `receiver.method(...)` for whitelisted prototype
// methods. The receiver must itself fold to a concrete value; otherwise we
// can't prove the call's result at compile time.
func (s *strCtx) foldMethodCall(call *ast.CallExpression, callee *ast.Node) foldResult {
	var method string
	var recvExpr *ast.Node
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		pae := callee.AsPropertyAccessExpression()
		if pae == nil || pae.Name() == nil || !ast.IsIdentifier(pae.Name()) {
			return foldResult{}
		}
		method = pae.Name().AsIdentifier().Text
		recvExpr = pae.Expression
	case ast.KindElementAccessExpression:
		eae := callee.AsElementAccessExpression()
		if eae == nil {
			return foldResult{}
		}
		name, ok := utils.GetStaticExpressionValue(eae.ArgumentExpression)
		if !ok {
			return foldResult{}
		}
		method = name
		recvExpr = eae.Expression
	default:
		return foldResult{}
	}

	returnsString := stringReturningMethods[method]
	returnsArray := arrayReturningMethods[method]
	if !returnsString && !returnsArray {
		return foldResult{}
	}
	recv := s.fold(recvExpr)
	if !recv.resolved() {
		return foldResult{}
	}
	// Array-only methods (join) require an array receiver.
	if method == "join" {
		if _, isArr := recv.value.(arrayVal); !isArr {
			return foldResult{}
		}
	}
	// Number-only methods require a numeric receiver.
	if method == "toFixed" || method == "toExponential" || method == "toPrecision" {
		if _, isNum := recv.value.(float64); !isNum {
			return foldResult{}
		}
	}
	if returnsString {
		return foldResult{""}
	}
	// returnsArray: propagate array-ness so the next method in a chain can fold.
	return foldResult{arrayVal{}}
}

// foldArrayLiteral returns a resolved array value iff every element itself
// folds (mirrors ESLint's conservative array-literal evaluation — any
// unresolvable element taints the whole array).
func (s *strCtx) foldArrayLiteral(node *ast.Node) foldResult {
	arr := node.AsArrayLiteralExpression()
	if arr == nil {
		return foldResult{arrayVal{}}
	}
	if arr.Elements != nil {
		for _, el := range arr.Elements.Nodes {
			if el == nil || el.Kind == ast.KindOmittedExpression {
				continue
			}
			// Spread elements would require the spread target itself to be
			// iterable and fully resolved; we don't attempt to fold them.
			if el.Kind == ast.KindSpreadElement {
				return foldResult{}
			}
			if !s.fold(el).resolved() {
				return foldResult{}
			}
		}
	}
	return foldResult{arrayVal{}}
}

// foldMemberAccess handles `o.x` / `o['x']` / `o[0]` by resolving the receiver
// to an object literal (possibly via const / let-no-writes / var-no-writes
// identifier chains and nested property accesses) and then looking up the key.
// The key itself is resolved via the static fold — bracket access expressions
// that only become constant after const-folding (e.g. `o[k]` where
// `const k = 'x'`) are handled the same way ESLint's getStaticValue does.
func (s *strCtx) foldMemberAccess(node *ast.Node) foldResult {
	key, recvExpr, ok := s.memberAccessKey(node)
	if !ok {
		return foldResult{}
	}
	obj := s.resolveToObjectLiteral(recvExpr)
	if obj == nil {
		return foldResult{}
	}
	val := s.findPropertyValueNode(obj, key)
	if val == nil {
		return foldResult{}
	}
	return s.fold(val)
}

// memberAccessKey extracts the static key and receiver expression of a
// Property/ElementAccess. For ElementAccess with a non-literal argument,
// it folds the argument as a last resort.
func (s *strCtx) memberAccessKey(node *ast.Node) (key string, receiver *ast.Node, ok bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		pae := node.AsPropertyAccessExpression()
		if pae == nil || pae.Name() == nil || !ast.IsIdentifier(pae.Name()) {
			return "", nil, false
		}
		return pae.Name().AsIdentifier().Text, pae.Expression, true
	case ast.KindElementAccessExpression:
		eae := node.AsElementAccessExpression()
		if eae == nil {
			return "", nil, false
		}
		if k, kok := utils.GetStaticExpressionValue(eae.ArgumentExpression); kok {
			return k, eae.Expression, true
		}
		// Fallback: fold the argument; covers `o[k]` where k resolves via
		// const / let-no-writes / var-no-writes to a string or number.
		r := s.fold(eae.ArgumentExpression)
		if !r.resolved() {
			return "", nil, false
		}
		return r.asString(), eae.Expression, true
	}
	return "", nil, false
}

// resolveToObjectLiteral walks through const / let-no-writes / var-no-writes
// identifier initializers and nested property accesses to locate the
// underlying ObjectLiteralExpression node. Returns nil if the receiver
// can't be resolved to a single object literal at compile time.
func (s *strCtx) resolveToObjectLiteral(node *ast.Node) *ast.Node {
	node = ast.SkipOuterExpressions(node, argOuterKinds)
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression:
		return node
	case ast.KindIdentifier:
		if s.ruleCtx.TypeChecker == nil {
			return nil
		}
		sym := utils.GetReferenceSymbol(node, s.ruleCtx.TypeChecker)
		if sym == nil || len(sym.Declarations) != 1 {
			return nil
		}
		decl := sym.Declarations[0]
		if decl.Kind != ast.KindVariableDeclaration {
			return nil
		}
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil || varDecl.Initializer == nil {
			return nil
		}
		list := decl.Parent
		if list == nil || list.Kind != ast.KindVariableDeclarationList {
			return nil
		}
		if list.Flags&ast.NodeFlagsConst == 0 && s.hasWrites(sym) {
			return nil
		}
		return s.resolveToObjectLiteral(varDecl.Initializer)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		key, recvExpr, ok := s.memberAccessKey(node)
		if !ok {
			return nil
		}
		parent := s.resolveToObjectLiteral(recvExpr)
		if parent == nil {
			return nil
		}
		sub := s.findPropertyValueNode(parent, key)
		if sub == nil {
			return nil
		}
		return s.resolveToObjectLiteral(sub)
	}
	return nil
}

// findPropertyValueNode returns the value expression associated with `key`
// inside an ObjectLiteralExpression. Handles regular, shorthand, and computed
// property names — including computed keys whose expression only becomes
// constant after folding (e.g. `const k = 'x'; {[k]: 'y'}.x`). Skips spread,
// accessors, and computed keys whose expression can't be resolved.
func (s *strCtx) findPropertyValueNode(obj *ast.Node, key string) *ast.Node {
	if obj == nil || obj.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	lit := obj.AsObjectLiteralExpression()
	if lit == nil || lit.Properties == nil {
		return nil
	}
	for _, p := range lit.Properties.Nodes {
		switch p.Kind {
		case ast.KindPropertyAssignment:
			pa := p.AsPropertyAssignment()
			if pa == nil || pa.Name() == nil {
				continue
			}
			if !s.propertyKeyMatches(pa.Name(), key) {
				continue
			}
			return pa.Initializer
		case ast.KindShorthandPropertyAssignment:
			spa := p.AsShorthandPropertyAssignment()
			if spa == nil || spa.Name() == nil {
				continue
			}
			n := spa.Name()
			if !ast.IsIdentifier(n) || n.AsIdentifier().Text != key {
				continue
			}
			// `{x}` resolves to the outer binding `x` — fold it via the
			// same identifier path the rest of the fold uses.
			return n
		}
	}
	return nil
}

// propertyKeyMatches compares a property-name node against a target key,
// resolving computed keys through the static fold so that
// `const k = 'x'; {[k]: ...}` matches the key "x".
func (s *strCtx) propertyKeyMatches(nameNode *ast.Node, key string) bool {
	if pname, ok := utils.GetStaticPropertyName(nameNode); ok {
		return pname == key
	}
	if nameNode.Kind == ast.KindComputedPropertyName {
		c := nameNode.AsComputedPropertyName()
		if c == nil {
			return false
		}
		r := s.fold(c.Expression)
		if !r.resolved() {
			return false
		}
		return r.asString() == key
	}
	return false
}

// foldStringRawTag handles `` String.raw`...` `` — ESLint's getStaticValue
// recognises the built-in tag and folds the template with each substitution
// recursively evaluated.
func (s *strCtx) foldStringRawTag(node *ast.Node) foldResult {
	tt := node.AsTaggedTemplateExpression()
	if tt == nil || tt.Tag == nil || tt.Template == nil {
		return foldResult{}
	}
	if !s.isStringRawTag(tt.Tag) {
		return foldResult{}
	}
	switch tt.Template.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		return foldResult{tt.Template.Text()}
	case ast.KindTemplateExpression:
		te := tt.Template.AsTemplateExpression()
		if te == nil {
			return foldResult{}
		}
		var sb strings.Builder
		if te.Head != nil {
			sb.WriteString(te.Head.Text())
		}
		if te.TemplateSpans != nil {
			for _, span := range te.TemplateSpans.Nodes {
				sp := span.AsTemplateSpan()
				if sp == nil {
					return foldResult{}
				}
				sub := s.fold(sp.Expression)
				if !sub.resolved() {
					return foldResult{}
				}
				sb.WriteString(sub.asString())
				if sp.Literal != nil {
					sb.WriteString(sp.Literal.Text())
				}
			}
		}
		return foldResult{sb.String()}
	}
	return foldResult{}
}

// isStringRawTag reports whether `tag` refers to the built-in `String.raw`.
func (s *strCtx) isStringRawTag(tag *ast.Node) bool {
	tag = ast.SkipOuterExpressions(tag, calleeOuterKinds)
	if tag == nil || tag.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	pae := tag.AsPropertyAccessExpression()
	if pae == nil {
		return false
	}
	name := pae.Name()
	if name == nil || !ast.IsIdentifier(name) || name.AsIdentifier().Text != "raw" {
		return false
	}
	obj := ast.SkipOuterExpressions(pae.Expression, calleeOuterKinds)
	if obj == nil || !ast.IsIdentifier(obj) {
		return false
	}
	if obj.AsIdentifier().Text != "String" {
		return false
	}
	return !utils.IsShadowed(obj, "String")
}

// hasWrites reports whether `sym` has any write reference in the source file.
// For let / var, a single write after initialization disqualifies the binding
// from static resolution. Lazily computed on first call.
func (s *strCtx) hasWrites(sym *ast.Symbol) bool {
	if !s.writeRefsComputed {
		s.computeWriteRefs()
	}
	return s.writeRefsMap[sym]
}

func (s *strCtx) computeWriteRefs() {
	s.writeRefsComputed = true
	if s.ruleCtx.TypeChecker == nil || s.ruleCtx.SourceFile == nil {
		return
	}
	m := map[*ast.Symbol]bool{}
	var visit func(n *ast.Node)
	visit = func(n *ast.Node) {
		if n == nil {
			return
		}
		if n.Kind == ast.KindIdentifier && utils.IsWriteReference(n) {
			if sym := s.ruleCtx.TypeChecker.GetSymbolAtLocation(n); sym != nil {
				m[sym] = true
			}
		}
		n.ForEachChild(func(c *ast.Node) bool {
			visit(c)
			return false
		})
	}
	visit(&s.ruleCtx.SourceFile.Node)
	s.writeRefsMap = m
}
