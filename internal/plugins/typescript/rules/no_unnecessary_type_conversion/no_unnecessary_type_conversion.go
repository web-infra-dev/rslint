package no_unnecessary_type_conversion

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUnnecessaryTypeConversionMessage(violation, typeName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryTypeConversion",
		Description: fmt.Sprintf("%s does not change the type or value of the %s.", violation, typeName),
		Data: map[string]string{
			"type":      typeName,
			"violation": violation,
		},
	}
}

func buildSuggestRemoveMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestRemove",
		Description: "Remove the type conversion.",
	}
}

func buildSuggestSatisfiesMessage(typeName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestSatisfies",
		Description: fmt.Sprintf("Instead, assert that the value satisfies the %s type.", typeName),
		Data: map[string]string{
			"type": typeName,
		},
	}
}

// doesUnderlyingTypeMatchFlag mirrors upstream's
// `unionConstituents(type).every(t => isTypeFlagSet(t, flag))` — only true when
// every union constituent already has the target type flag, so the conversion
// is provably a no-op.
func doesUnderlyingTypeMatchFlag(t *checker.Type, flag checker.TypeFlags) bool {
	if t == nil {
		return false
	}
	for _, part := range utils.UnionTypeParts(t) {
		if !utils.IsTypeFlagSet(part, flag) {
			return false
		}
	}
	return true
}

// wrappedInnerText returns innerNode's source text, parenthesized when the
// inner node does not have strong precedence (so inlining into a tighter
// context would silently rebind operators). Reuses the shared
// utils.IsStrongPrecedenceNode helper used by prefer_regexp_exec etc.
func wrappedInnerText(sourceFile *ast.SourceFile, innerNode *ast.Node) string {
	text := utils.TrimmedNodeText(sourceFile, innerNode)
	if !utils.IsStrongPrecedenceNode(innerNode) {
		return "(" + text + ")"
	}
	return text
}

// buildWrappingFix produces the replacement text for `outerNode` constructed
// from `innerNode`'s text. When `wrap` is nil the inner text replaces the
// outer node directly; when `wrap` is provided the result is additionally
// wrapped in parens if the surrounding context could rebind precedence.
func buildWrappingFix(sourceFile *ast.SourceFile, outerNode *ast.Node, innerNode *ast.Node, wrap func(string) string) rule.RuleFix {
	innerCode := wrappedInnerText(sourceFile, innerNode)
	var code string
	if wrap == nil {
		code = innerCode
	} else {
		code = wrap(innerCode)
		if typescriptutil.IsWeakPrecedenceParent(outerNode) {
			code = "(" + code + ")"
		}
	}
	return rule.RuleFixReplace(sourceFile, outerNode, code)
}

// argumentSkippingParens unwraps any ParenthesizedExpression chain so that the
// remaining node mirrors what ESLint's AST would expose as `.argument` /
// `.arguments[0]` / `.left` etc. tsgo keeps parens as explicit nodes; the
// upstream rule never sees them.
func argumentSkippingParens(node *ast.Node) *ast.Node {
	return ast.SkipParentheses(node)
}

// isEmptyStringLiteral mirrors upstream's `node.type === Literal && value === ”`.
// In ESTree a TemplateLiteral is NOT a Literal — even “ “ “. tsgo keeps
// NoSubstitutionTemplateLiteral as a literal-kind, but to stay aligned we
// match the upstream contract and only treat StringLiteral as empty here.
//
// SkipParentheses lets `(”)` and `((”))` qualify, matching upstream which
// never sees ParenthesizedExpression — ESTree treats parens transparently.
func isEmptyStringLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	stripped := ast.SkipParentheses(node)
	if stripped.Kind == ast.KindStringLiteral {
		return stripped.AsStringLiteral().Text == ""
	}
	return false
}

// isLocallyShadowed mirrors upstream's `scope.set.get(name).defs.length > 0`
// check: it asks only the IMMEDIATELY enclosing scope (the nearest block,
// module body, function-like, or source file) whether it declares a binding
// with the given name. Unlike rslint's `utils.IsShadowed` it does NOT walk up
// the scope chain — upstream's ESLint scope.set is local-only, and matching
// that quirk is required for parity (verified via differential validation
// against typescript-eslint 8.x on nested function/block/arrow cases).
//
// Both value bindings (function/class/enum/var/import/namespace) and type
// bindings (type alias/interface) shadow, since ESLint's scope manager folds
// type bindings into the same scope.set.
func isLocallyShadowed(node *ast.Node, name string) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindSourceFile:
			sf := current.AsSourceFile()
			if sf != nil && sf.Statements != nil {
				return statementsDeclare(sf.Statements.Nodes, name)
			}
			return false
		case ast.KindModuleBlock:
			mb := current.AsModuleBlock()
			if mb != nil && mb.Statements != nil {
				return statementsDeclare(mb.Statements.Nodes, name)
			}
			return false
		case ast.KindBlock:
			block := current.AsBlock()
			if block != nil && block.Statements != nil {
				if statementsDeclare(block.Statements.Nodes, name) {
					return true
				}
			}
			// A Block that is a function-like body shares scope with the
			// function's name + parameters per ESLint's scope manager.
			parent := current.Parent
			if parent != nil && ast.IsFunctionLikeDeclaration(parent) {
				if functionHasOwnNameOrParam(parent, name) {
					return true
				}
			}
			return false
		}
		// Arrow function with expression body (no Block to land on), method
		// declarations without a separate visit-able body, etc.
		if ast.IsFunctionLikeDeclaration(current) {
			return functionHasOwnNameOrParam(current, name)
		}
	}
	return false
}

func functionHasOwnNameOrParam(fn *ast.Node, name string) bool {
	if fn.Kind == ast.KindFunctionDeclaration || fn.Kind == ast.KindFunctionExpression {
		n := fn.Name()
		if n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
			return true
		}
	}
	return utils.HasShadowingParameter(fn, name)
}

// statementsDeclare reports whether any of the given top-level/block-level
// statements introduces a binding (value or type) with `name`.
func statementsDeclare(stmts []*ast.Node, name string) bool {
	for _, stmt := range stmts {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindClassDeclaration,
			ast.KindEnumDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindInterfaceDeclaration:
			n := stmt.Name()
			if n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		case ast.KindModuleDeclaration:
			md := stmt.AsModuleDeclaration()
			if md != nil && md.Name() != nil &&
				md.Name().Kind == ast.KindIdentifier && md.Name().Text() == name {
				return true
			}
		case ast.KindVariableStatement:
			vs := stmt.AsVariableStatement()
			if vs == nil || vs.DeclarationList == nil {
				continue
			}
			dl := vs.DeclarationList.AsVariableDeclarationList()
			if dl == nil || dl.Declarations == nil {
				continue
			}
			for _, decl := range dl.Declarations.Nodes {
				if decl == nil || decl.Kind != ast.KindVariableDeclaration {
					continue
				}
				vd := decl.AsVariableDeclaration()
				if vd != nil && vd.Name() != nil && utils.HasNameInBindingPattern(vd.Name(), name) {
					return true
				}
			}
		case ast.KindImportEqualsDeclaration:
			ie := stmt.AsImportEqualsDeclaration()
			if ie != nil && ie.Name() != nil && ie.Name().Text() == name {
				return true
			}
		case ast.KindImportDeclaration:
			id := stmt.AsImportDeclaration()
			if id == nil || id.ImportClause == nil {
				continue
			}
			ic := id.ImportClause.AsImportClause()
			if ic == nil {
				continue
			}
			if ic.Name() != nil && ic.Name().Text() == name {
				return true
			}
			if ic.NamedBindings != nil {
				switch ic.NamedBindings.Kind {
				case ast.KindNamespaceImport:
					ni := ic.NamedBindings.AsNamespaceImport()
					if ni != nil && ni.Name() != nil && ni.Name().Text() == name {
						return true
					}
				case ast.KindNamedImports:
					nis := ic.NamedBindings.AsNamedImports()
					if nis != nil && nis.Elements != nil {
						for _, elem := range nis.Elements.Nodes {
							if elem == nil {
								continue
							}
							is := elem.AsImportSpecifier()
							if is != nil && is.Name() != nil && is.Name().Text() == name {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

// callExprTypeFlags maps a built-in conversion function name to the TypeFlags
// its argument must already satisfy for the call to be a no-op.
var callExprTypeFlags = map[string]checker.TypeFlags{
	"BigInt":  checker.TypeFlagsBigIntLike,
	"Boolean": checker.TypeFlagsBooleanLike,
	"Number":  checker.TypeFlagsNumberLike,
	"String":  checker.TypeFlagsStringLike,
}

// isEnumOrEnumMemberType mirrors upstream's `isEnumType || isEnumMemberType`
// short-circuit on `.toString()` — calling `EnumMember.toString()` is a way
// to read the enum's underlying string/number, not a no-op type conversion.
func isEnumOrEnumMemberType(t *checker.Type) bool {
	if t == nil {
		return false
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsEnumLike) {
		return true
	}
	sym := checker.Type_symbol(t)
	if sym != nil && utils.IsSymbolFlagSet(sym, ast.SymbolFlagsEnumMember) {
		return true
	}
	return false
}

// isInteger mirrors upstream's `Number.isInteger((t as NumberLiteralType).value)`.
// tsgo stores NumberLiteral values as `jsnum.Number` (float64). Calling
// `checker.ValueToString` is unsafe for integrality checks because Go's
// `json.Marshal(float64)` emits scientific notation past ~1e21 (`1e+21`),
// matching JS but breaking a string-pattern integer test. Parse the resulting
// text back to a float and ask `math.Trunc(f) == f` to match
// `Number.isInteger` semantics exactly.
func isInteger(val interface{}) bool {
	if val == nil {
		return false
	}
	s := checker.ValueToString(val)
	if s == "" {
		return false
	}
	f, err := strconv.ParseFloat(strings.TrimSuffix(s, "n"), 64)
	if err != nil {
		return false
	}
	return !math.IsNaN(f) && !math.IsInf(f, 0) && f == math.Trunc(f)
}

// allUnionPartsAreIntegerNumberLiteral mirrors upstream's `~~` integer check:
// every union constituent must be a NumberLiteral whose value is an integer.
func allUnionPartsAreIntegerNumberLiteral(t *checker.Type) bool {
	if t == nil {
		return false
	}
	parts := utils.UnionTypeParts(t)
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if !utils.IsTypeFlagSet(part, checker.TypeFlagsNumberLiteral) {
			return false
		}
		if !isInteger(part.AsLiteralType().Value()) {
			return false
		}
	}
	return true
}

var NoUnnecessaryTypeConversionRule = rule.CreateRule(rule.Rule{
	Name:             "no-unnecessary-type-conversion",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		sourceFile := ctx.SourceFile

		// reportUnaryConversion reports unary-operator-style conversions (`+x`,
		// `!!x`, `~~x`). `outerNode` is the report anchor (the full `!!` /
		// `~~` for double-operator variants, the single operator otherwise),
		// `innerArgument` is the operand whose source text replaces it.
		reportUnaryConversion := func(outerNode, opStartNode, innerArgument *ast.Node, typeName, violation string) {
			outerRange := utils.TrimNodeTextRange(sourceFile, outerNode)
			opStart := utils.TrimNodeTextRange(sourceFile, opStartNode).Pos()
			reportRange := core.NewTextRange(outerRange.Pos(), opStart+1)
			ctx.ReportRangeWithSuggestions(reportRange,
				buildUnnecessaryTypeConversionMessage(violation, typeName),
				rule.RuleSuggestion{
					Message:  buildSuggestRemoveMessage(),
					FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, outerNode, innerArgument, nil)},
				},
				rule.RuleSuggestion{
					Message: buildSuggestSatisfiesMessage(typeName),
					FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, outerNode, innerArgument, func(code string) string {
						return code + " satisfies " + typeName
					})},
				},
			)
		}

		reportCallConversion := func(callNode, calleeIdentifier, innerArg *ast.Node, fnName string) {
			typeName := strings.ToLower(fnName)
			violation := "Passing a " + typeName + " to " + fnName + "()"
			calleeRange := utils.TrimNodeTextRange(sourceFile, calleeIdentifier)
			ctx.ReportRangeWithSuggestions(calleeRange,
				buildUnnecessaryTypeConversionMessage(violation, typeName),
				rule.RuleSuggestion{
					Message:  buildSuggestRemoveMessage(),
					FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, callNode, innerArg, nil)},
				},
				rule.RuleSuggestion{
					Message: buildSuggestSatisfiesMessage(typeName),
					FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, callNode, innerArg, func(code string) string {
						return code + " satisfies " + typeName
					})},
				},
			)
		}

		reportToString := func(callNode, propertyIdent, innerObject *ast.Node) {
			propertyRange := utils.TrimNodeTextRange(sourceFile, propertyIdent)
			callRange := utils.TrimNodeTextRange(sourceFile, callNode)
			reportRange := core.NewTextRange(propertyRange.Pos(), callRange.End())
			ctx.ReportRangeWithSuggestions(reportRange,
				buildUnnecessaryTypeConversionMessage("Calling a string's .toString() method", "string"),
				rule.RuleSuggestion{
					Message:  buildSuggestRemoveMessage(),
					FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, callNode, innerObject, nil)},
				},
				rule.RuleSuggestion{
					Message: buildSuggestSatisfiesMessage("string"),
					FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, callNode, innerObject, func(code string) string {
						return code + " satisfies string"
					})},
				},
			)
		}

		// handleBinaryPlus handles both `+` and `+=` BinaryExpressions, branching
		// on operator kind. tsgo collapses ESLint's separate AssignmentExpression
		// listener into BinaryExpression too, so a single visitor covers both.
		handleBinaryPlus := func(node *ast.Node) {
			bin := node.AsBinaryExpression()
			if bin == nil || bin.OperatorToken == nil {
				return
			}
			operator := bin.OperatorToken.Kind
			left := bin.Left
			right := bin.Right
			if left == nil || right == nil {
				return
			}

			switch operator {
			case ast.KindPlusEqualsToken:
				if !isEmptyStringLiteral(right) {
					return
				}
				leftType := ctx.TypeChecker.GetTypeAtLocation(left)
				if !doesUnderlyingTypeMatchFlag(leftType, checker.TypeFlagsStringLike) {
					return
				}

				inner := argumentSkippingParens(left)
				removeFix := buildWrappingFix(sourceFile, node, inner, nil)
				if node.Parent != nil && node.Parent.Kind == ast.KindExpressionStatement {
					removeFix = rule.RuleFixRemoveRange(utils.TrimNodeTextRange(sourceFile, node.Parent))
				}

				ctx.ReportNodeWithSuggestions(node,
					buildUnnecessaryTypeConversionMessage("Concatenating a string with ''", "string"),
					rule.RuleSuggestion{
						Message:  buildSuggestRemoveMessage(),
						FixesArr: []rule.RuleFix{removeFix},
					},
					rule.RuleSuggestion{
						Message: buildSuggestSatisfiesMessage("string"),
						FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, node, inner, func(code string) string {
							return code + " satisfies string"
						})},
					},
				)

			case ast.KindPlusToken:
				// case: <string-like> + ''
				if isEmptyStringLiteral(right) {
					leftType := ctx.TypeChecker.GetTypeAtLocation(left)
					if !doesUnderlyingTypeMatchFlag(leftType, checker.TypeFlagsStringLike) {
						return
					}
					inner := argumentSkippingParens(left)
					// Use the unwrapped inner's end for the start of the report,
					// mirroring upstream's ESTree view (which never sees parens).
					// `((('x'))) + ''` reports starting at the end of `'x'`, not
					// the end of the outermost paren.
					innerRange := utils.TrimNodeTextRange(sourceFile, inner)
					nodeRange := utils.TrimNodeTextRange(sourceFile, node)
					reportRange := core.NewTextRange(innerRange.End(), nodeRange.End())

					ctx.ReportRangeWithSuggestions(reportRange,
						buildUnnecessaryTypeConversionMessage("Concatenating a string with ''", "string"),
						rule.RuleSuggestion{
							Message:  buildSuggestRemoveMessage(),
							FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, node, inner, nil)},
						},
						rule.RuleSuggestion{
							Message: buildSuggestSatisfiesMessage("string"),
							FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, node, inner, func(code string) string {
								return code + " satisfies string"
							})},
						},
					)
					return
				}
				// case: '' + <string-like>
				if isEmptyStringLiteral(left) {
					rightType := ctx.TypeChecker.GetTypeAtLocation(right)
					if !doesUnderlyingTypeMatchFlag(rightType, checker.TypeFlagsStringLike) {
						return
					}
					inner := argumentSkippingParens(right)
					// Same paren-transparent rule as the right-empty branch: end
					// the report at the unwrapped inner's start, so `'' + ((s))`
					// matches upstream's columns instead of including parens.
					nodeRange := utils.TrimNodeTextRange(sourceFile, node)
					innerRange := utils.TrimNodeTextRange(sourceFile, inner)
					reportRange := core.NewTextRange(nodeRange.Pos(), innerRange.Pos())

					ctx.ReportRangeWithSuggestions(reportRange,
						buildUnnecessaryTypeConversionMessage("Concatenating '' with a string", "string"),
						rule.RuleSuggestion{
							Message:  buildSuggestRemoveMessage(),
							FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, node, inner, nil)},
						},
						rule.RuleSuggestion{
							Message: buildSuggestSatisfiesMessage("string"),
							FixesArr: []rule.RuleFix{buildWrappingFix(sourceFile, node, inner, func(code string) string {
								return code + " satisfies string"
							})},
						},
					)
				}
			}
		}

		handleCallExpression := func(node *ast.Node) {
			call := node.AsCallExpression()
			if call == nil || call.Expression == nil {
				return
			}

			// (1) String(arg) / Number(arg) / Boolean(arg) / BigInt(arg) —
			// tsgo keeps `(String)('x')` parens explicit as a
			// ParenthesizedExpression wrapping the Identifier; ESTree has
			// no paren node. Unwrap before matching the callee name and
			// report on the bare identifier so the location matches upstream.
			calleeExpr := ast.SkipParentheses(call.Expression)
			if calleeExpr.Kind == ast.KindIdentifier {
				name := calleeExpr.AsIdentifier().Text
				if flag, ok := callExprTypeFlags[name]; ok {
					if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
						return
					}
					if isLocallyShadowed(node, name) {
						return
					}
					arg0 := call.Arguments.Nodes[0]
					// `String(...arr)` is never a no-op: spread iterates the
					// rhs and forwards only the first element (or `undefined`
					// when the iterator is empty). tsgo unwraps SpreadElement
					// to its element type, which would otherwise let the rule
					// over-fire for `string[]` etc. — upstream stays silent
					// because TypeScript treats SpreadElement as the spreadable
					// (an array/iterable) type. Short-circuit explicitly.
					if arg0.Kind == ast.KindSpreadElement {
						return
					}
					argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, arg0)
					if !doesUnderlyingTypeMatchFlag(argType, flag) {
						return
					}
					inner := argumentSkippingParens(arg0)
					reportCallConversion(node, calleeExpr, inner, name)
					return
				}
			}

			// (2) <string-like>.toString()
			if call.Expression.Kind == ast.KindPropertyAccessExpression {
				member := call.Expression.AsPropertyAccessExpression()
				if member == nil || member.Name() == nil || member.Expression == nil {
					return
				}
				nameNode := member.Name()
				if nameNode.Kind != ast.KindIdentifier || nameNode.AsIdentifier().Text != "toString" {
					return
				}
				objectType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, member.Expression)
				if isEnumOrEnumMemberType(objectType) {
					return
				}
				// Treat each union part individually for the enum-member short-circuit;
				// upstream uses `isEnumType(type) || isEnumMemberType(type)` against
				// the constrained type as a whole, so a union containing a non-enum
				// constituent falls through to the StringLike-only check below.
				if !doesUnderlyingTypeMatchFlag(objectType, checker.TypeFlagsStringLike) {
					return
				}
				inner := argumentSkippingParens(member.Expression)
				reportToString(node, nameNode, inner)
			}
		}

		handlePrefixUnary := func(node *ast.Node) {
			pu := node.AsPrefixUnaryExpression()
			if pu == nil || pu.Operand == nil {
				return
			}

			switch pu.Operator {
			case ast.KindPlusToken:
				// `+x` where x is number-like.
				operand := pu.Operand
				argType := ctx.TypeChecker.GetTypeAtLocation(operand)
				if !doesUnderlyingTypeMatchFlag(argType, checker.TypeFlagsNumberLike) {
					return
				}
				inner := argumentSkippingParens(operand)
				reportUnaryConversion(node, node, inner, "number", "Using the unary + operator on a number")

			case ast.KindExclamationToken:
				// `!!x` — fire on the OUTER `!` whose operand (after stripping
				// any ParenthesizedExpression layers) is another `!`. Detecting
				// from the outer side lets `!(!x)` qualify too; ESTree has no
				// paren node so upstream sees both as identical.
				innerNode := ast.SkipParentheses(pu.Operand)
				if innerNode == nil || innerNode.Kind != ast.KindPrefixUnaryExpression {
					return
				}
				innerPu := innerNode.AsPrefixUnaryExpression()
				if innerPu == nil || innerPu.Operator != ast.KindExclamationToken || innerPu.Operand == nil {
					return
				}
				argType := ctx.TypeChecker.GetTypeAtLocation(innerPu.Operand)
				if !doesUnderlyingTypeMatchFlag(argType, checker.TypeFlagsBooleanLike) {
					return
				}
				inner := argumentSkippingParens(innerPu.Operand)
				reportUnaryConversion(node, innerNode, inner, "boolean", "Using !! on a boolean")

			case ast.KindTildeToken:
				// `~~x` — same outer-side detection so `~(~x)` qualifies.
				// The operand of the inner `~` must be an integer NumberLiteral
				// union; non-integers are rounded by `~~` and so the operator
				// is meaningful.
				innerNode := ast.SkipParentheses(pu.Operand)
				if innerNode == nil || innerNode.Kind != ast.KindPrefixUnaryExpression {
					return
				}
				innerPu := innerNode.AsPrefixUnaryExpression()
				if innerPu == nil || innerPu.Operator != ast.KindTildeToken || innerPu.Operand == nil {
					return
				}
				argType := ctx.TypeChecker.GetTypeAtLocation(innerPu.Operand)
				if !allUnionPartsAreIntegerNumberLiteral(argType) {
					return
				}
				inner := argumentSkippingParens(innerPu.Operand)
				reportUnaryConversion(node, innerNode, inner, "number", "Using ~~ on an integer")
			}
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression:      handleBinaryPlus,
			ast.KindCallExpression:        handleCallExpression,
			ast.KindPrefixUnaryExpression: handlePrefixUnary,
		}
	},
})
