package no_unnecessary_parameter_property_assignment

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoUnnecessaryParameterPropertyAssignmentRule mirrors
// @typescript-eslint/no-unnecessary-parameter-property-assignment.
//
// It flags `this.X = X` (and the `||= && ??= && &&=` variants) inside a
// constructor body whenever the constructor itself declares a parameter
// property named X — the parameter-property syntax already performs the
// assignment, so the explicit one is redundant.
//
// https://typescript-eslint.io/rules/no-unnecessary-parameter-property-assignment
// Upstream source: packages/eslint-plugin/src/rules/no-unnecessary-parameter-property-assignment.ts
var NoUnnecessaryParameterPropertyAssignmentRule = rule.CreateRule(rule.Rule{
	Name:             "no-unnecessary-parameter-property-assignment",
	RequiresTypeInfo: true,
	Run:              run,
})

// reportInfo mirrors upstream's per-ClassBody stack frame.
//
//   - assignedBeforeConstructor: property names that are assigned in a class
//     field initializer (directly, or via an arrow IIFE whose CallExpression
//     is the PropertyDeclaration.Initializer). If a name shows up here, any
//     `this.X = X` candidate that already landed in unnecessaryAssignments is
//     suppressed on ClassBody:exit.
//   - assignedBeforeUnnecessary: property names that received a non-unnecessary
//     assignment operator (`+=`, `-=`, …) earlier in the SAME constructor —
//     a later `this.X = X` is then treated as a re-assignment, not a redundant
//     one, and won't be pushed.
//   - unnecessaryAssignments: candidate `this.X = X` (or `||= && ??= && &&=`)
//     assignments that survived the assignedBeforeUnnecessary gate; reported on
//     ClassBody:exit unless suppressed by assignedBeforeConstructor.
type reportInfo struct {
	assignedBeforeConstructor map[string]bool
	assignedBeforeUnnecessary map[string]bool
	unnecessaryAssignments    []unnecessaryAssignment
}

type unnecessaryAssignment struct {
	name string
	node *ast.Node
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	var stack []*reportInfo
	// fileBuffer accumulates unsuppressed unnecessary assignments across all
	// classes visited so far. The linter doesn't sort diagnostics, and visit
	// order (inner class exit before outer class exit) doesn't match the
	// expected source-position order — so we hold reports until the
	// OUTERMOST class exit, sort by Pos, then emit. Reset after each flush
	// so multiple sibling top-level classes don't bleed into each other.
	var fileBuffer []unnecessaryAssignment

	push := func() {
		stack = append(stack, &reportInfo{
			assignedBeforeConstructor: map[string]bool{},
			assignedBeforeUnnecessary: map[string]bool{},
		})
	}

	pop := func() *reportInfo {
		if len(stack) == 0 {
			return nil
		}
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return top
	}

	top := func() *reportInfo {
		if len(stack) == 0 {
			return nil
		}
		return stack[len(stack)-1]
	}

	msg := rule.RuleMessage{
		Id:          "unnecessaryAssign",
		Description: "This assignment is unnecessary since it is already assigned by a parameter property.",
	}

	enterClass := func(*ast.Node) { push() }
	exitClass := func(*ast.Node) {
		info := pop()
		if info == nil {
			return
		}
		for _, ua := range info.unnecessaryAssignments {
			if info.assignedBeforeConstructor[ua.name] {
				continue
			}
			fileBuffer = append(fileBuffer, ua)
		}
		// Outermost class just exited — flush in source-position order.
		// Sibling top-level classes flush independently, so reports for
		// earlier classes are emitted before later ones (matching how
		// ESLint sorts its diagnostics).
		if len(stack) == 0 && len(fileBuffer) > 0 {
			sort.SliceStable(fileBuffer, func(i, j int) bool {
				return fileBuffer[i].node.Pos() < fileBuffer[j].node.Pos()
			})
			for _, ua := range fileBuffer {
				ctx.ReportNode(ua.node, msg)
			}
			fileBuffer = fileBuffer[:0]
		}
	}

	return rule.RuleListeners{
		// tsgo doesn't have a separate ClassBody node — the class declaration /
		// expression IS the body. Push/pop on the class node itself.
		ast.KindClassDeclaration:                      enterClass,
		rule.ListenerOnExit(ast.KindClassDeclaration): exitClass,
		ast.KindClassExpression:                       enterClass,
		rule.ListenerOnExit(ast.KindClassExpression):  exitClass,

		ast.KindBinaryExpression: func(node *ast.Node) {
			bin := node.AsBinaryExpression()
			op := bin.OperatorToken.Kind
			if !ast.IsAssignmentOperator(op) {
				return
			}
			if len(stack) == 0 {
				return
			}

			leftName, ok := getPropertyName(bin.Left)
			if !ok {
				return
			}

			funcNode := findParentFunction(node)
			info := top()

			// Constructor listener — mirrors upstream's
			// `MethodDefinition[kind='constructor'] > FunctionExpression
			// AssignmentExpression`. Arrow IIFE wrappers between the
			// assignment and the constructor are transparent.
			ctorFunc := funcNode
			if ctorFunc != nil && ast.IsArrowFunction(ctorFunc) {
				if call := iifeCallOfArrow(ctorFunc); call != nil {
					ctorFunc = findParentFunction(call)
				}
			}
			if ctorFunc != nil && ctorFunc.Kind == ast.KindConstructor {
				handleConstructor(ctx, info, bin, leftName, op, ctorFunc)
			}

			// PropertyDefinition listener — mirrors upstream's
			// `PropertyDefinition AssignmentExpression`. The two listeners
			// fire independently in upstream and may both apply to the same
			// assignment (e.g. a nested class field initializer inside an
			// outer constructor); we mirror that by running both arms here.
			handlePropertyDef(info, funcNode, node, leftName)
		},
	}
}

// handlePropertyDef mirrors upstream's
// `PropertyDefinition AssignmentExpression` listener: marks `leftName` as
// assigned before constructor runs IF the assignment is either (a) directly
// inside a PropertyDeclaration's initializer with no intermediate function,
// or (b) inside an arrow IIFE whose CallExpression sits as the
// PropertyDeclaration's initializer (modulo wrapping parens).
//
// Any other function ancestor (non-arrow function, non-IIFE arrow, method,
// nested class's constructor) breaks the chain and we return.
func handlePropertyDef(info *reportInfo, funcNode, node *ast.Node, leftName string) {
	if info == nil {
		return
	}
	if funcNode != nil {
		if !ast.IsArrowFunction(funcNode) {
			return
		}
		call := iifeCallOfArrow(funcNode)
		if call == nil {
			return
		}
		propDef := ast.FindAncestorKind(node, ast.KindPropertyDeclaration)
		if propDef == nil {
			return
		}
		if ast.SkipParentheses(propDef.AsPropertyDeclaration().Initializer) != call {
			return
		}
	} else {
		if ast.FindAncestorKind(node, ast.KindPropertyDeclaration) == nil {
			return
		}
	}
	info.assignedBeforeConstructor[leftName] = true
}

// handleConstructor mirrors the `MethodDefinition[kind='constructor'] >
// FunctionExpression AssignmentExpression` listener: track `this.X`
// assignments inside the constructor body and either suppress later
// candidates (compound-op writes seed assignedBeforeUnnecessary) or push the
// candidate onto unnecessaryAssignments.
func handleConstructor(
	ctx rule.RuleContext,
	info *reportInfo,
	bin *ast.BinaryExpression,
	leftName string,
	op ast.Kind,
	funcNode *ast.Node,
) {
	if info == nil {
		return
	}

	if !isUnnecessaryOperator(op) {
		info.assignedBeforeUnnecessary[leftName] = true
		return
	}

	rightID := getIdentifier(bin.Right)
	if rightID == nil {
		return
	}
	rightName := rightID.AsIdentifier().Text
	if leftName != rightName {
		return
	}
	if !isReferenceFromParameter(ctx, rightID) {
		return
	}

	if !constructorHasParameterPropertyNamed(funcNode, rightName) {
		return
	}
	if info.assignedBeforeUnnecessary[leftName] {
		return
	}

	info.unnecessaryAssignments = append(info.unnecessaryAssignments, unnecessaryAssignment{
		name: leftName,
		node: bin.AsNode(),
	})
}

// isUnnecessaryOperator mirrors upstream's UNNECESSARY_OPERATORS set:
// the bare assignment plus the logical-assignment family that produces an
// identical effect to the parameter-property assignment when the receiver
// has just been initialized.
func isUnnecessaryOperator(op ast.Kind) bool {
	switch op {
	case ast.KindEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// getPropertyName extracts the property name from a `this.X` / `this[X]`
// access on the assignment's left-hand side. Returns ("", false) if `left`
// is not a `this`-rooted member access or the property is not a statically
// known name.
//
// Upstream's logic (verbatim):
//
//	if (node.property.type === Identifier) return node.property.name
//	if (node.computed) return getStaticStringValue(node.property)
//	return null
//
// Note the first branch fires for BOTH `this.foo` (non-computed Identifier)
// AND `this[foo]` (computed Identifier) — the same name string is returned
// in both cases. We mirror that, including the case where a `this[ident]`
// is treated as if it directly named `ident`; the downstream leftName ==
// rightId.name + isReferenceFromParameter checks act as the safety net.
func getPropertyName(left *ast.Node) (string, bool) {
	left = ast.SkipParentheses(left)
	if left == nil {
		return "", false
	}
	switch left.Kind {
	case ast.KindPropertyAccessExpression:
		pa := left.AsPropertyAccessExpression()
		// Mirror ESTree's paren-elision: `(this).foo` is identical to
		// `this.foo` in upstream's AST, so we strip the receiver's parens
		// before the ThisKeyword check.
		receiver := ast.SkipParentheses(pa.Expression)
		if receiver == nil || receiver.Kind != ast.KindThisKeyword {
			return "", false
		}
		name := pa.Name()
		if name == nil || !ast.IsIdentifier(name) {
			return "", false
		}
		return name.AsIdentifier().Text, true
	case ast.KindElementAccessExpression:
		ea := left.AsElementAccessExpression()
		receiver := ast.SkipParentheses(ea.Expression)
		if receiver == nil || receiver.Kind != ast.KindThisKeyword {
			return "", false
		}
		// Index argument is also paren-elided in ESTree, so unwrap before
		// kind dispatch (`this[(foo)]` should behave like `this[foo]`).
		arg := ast.SkipParentheses(ea.ArgumentExpression)
		if arg == nil {
			return "", false
		}
		if arg.Kind == ast.KindIdentifier {
			return arg.AsIdentifier().Text, true
		}
		if v := utils.GetStaticStringValue(arg); v != "" {
			return v, true
		}
		return "", false
	}
	return "", false
}

// findParentFunction walks up the parent chain to the nearest function-like
// declaration — `ast.IsFunctionLikeDeclaration` matches the same set upstream
// stops at, once the ESTree → tsgo collapse is accounted for:
// FunctionDeclaration / FunctionExpression / ArrowFunction (upstream stops
// here) PLUS Constructor / MethodDeclaration / Get/SetAccessor (tsgo's
// counterpart of ESTree's MethodDefinition-wrapping-FunctionExpression form).
//
// Signature kinds (CallSignature, FunctionType, …) are intentionally
// excluded — they have no executable body and so cannot contain a
// BinaryExpression.
//
// ClassStaticBlockDeclaration is also NOT a stop point: upstream's walker
// doesn't list StaticBlock either, so an assignment in a static block walks
// past it to find any enclosing function. (None of the upstream tests
// exercise this combination; we mirror the behavior for parity.)
func findParentFunction(node *ast.Node) *ast.Node {
	return ast.FindAncestor(node.Parent, ast.IsFunctionLikeDeclaration)
}

// iifeCallOfArrow returns the CallExpression that immediately invokes
// `arrow` if it does, walking through any wrapping ParenthesizedExpressions
// that tsgo preserves (ESTree elides them, which is why upstream can match
// `arrow.parent.type === CallExpression` directly). Returns nil when the
// arrow is not the (possibly parenthesized) operand of a CallExpression.
func iifeCallOfArrow(arrow *ast.Node) *ast.Node {
	parent := ast.WalkUpParenthesizedExpressions(arrow.Parent)
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return nil
	}
	return parent
}

// getIdentifier unwraps the right-hand side through `as` / `!` (TSAsExpression
// / TSNonNullExpression in upstream) to find a bare Identifier. Parens ARE
// stripped via ast.SkipParentheses so the function tracks upstream's behavior
// on the ESTree AST (which elides ParenthesizedExpression entirely) — e.g.
// `this.foo = (foo)` is reported by upstream because in ESTree the right-hand
// side IS the Identifier; tsgo preserves the wrapping ParenthesizedExpression,
// so we must skip it explicitly to align. SatisfiesExpression and other
// wrappers are intentionally left alone — upstream's switch lists only
// TSAsExpression and TSNonNullExpression, so anything else falls through to
// the default `return null` arm.
func getIdentifier(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node
	case ast.KindAsExpression:
		return getIdentifier(node.AsAsExpression().Expression)
	case ast.KindNonNullExpression:
		return getIdentifier(node.AsNonNullExpression().Expression)
	}
	return nil
}

// isReferenceFromParameter reports whether `id` resolves to a parameter
// binding. Upstream uses ESLint's scope manager
// (`resolved.defs.at(0)?.type === DefinitionType.Parameter`); tsgo exposes
// the same fact via TypeChecker symbol resolution. We mirror upstream's
// "first definition only" check by inspecting only Declarations[0], so
// pathological cases involving declaration-merging or overload chains line
// up byte-for-byte with upstream's decision.
func isReferenceFromParameter(ctx rule.RuleContext, id *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	sym := ctx.TypeChecker.GetSymbolAtLocation(id)
	if sym == nil || len(sym.Declarations) == 0 {
		return false
	}
	return sym.Declarations[0].Kind == ast.KindParameter
}

// constructorHasParameterPropertyNamed reports whether the constructor's
// parameter list contains a parameter-property declaration whose binding
// identifier matches `name`. tsgo encodes parameter properties as
// KindParameter nodes carrying a public/private/protected/readonly modifier
// (recognized via ast.IsParameterPropertyDeclaration); the binding name is
// the parameter's `Name()`. Default-valued parameter properties
// (`public foo = 1`) still have an Identifier as Name (the default lives in
// Initializer), so a single Identifier check covers both shapes ESLint
// distinguishes via Identifier vs AssignmentPattern.
func constructorHasParameterPropertyNamed(constructorNode *ast.Node, name string) bool {
	for _, param := range constructorNode.Parameters() {
		if !ast.IsParameterPropertyDeclaration(param, constructorNode) {
			continue
		}
		nameNode := param.Name()
		if nameNode == nil || !ast.IsIdentifier(nameNode) {
			continue
		}
		if nameNode.AsIdentifier().Text == name {
			return true
		}
	}
	return false
}
