package method_signature_style

import (
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type mode string

const (
	modeProperty mode = "property"
	modeMethod   mode = "method"
)

func messageErrorMethod() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorMethod",
		Description: "Shorthand method signature is forbidden. Use a function property instead.",
	}
}

func messageErrorProperty() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorProperty",
		Description: "Function property signature is forbidden. Use a method shorthand instead.",
	}
}

var MethodSignatureStyleRule = rule.CreateRule(rule.Rule{
	Name: "method-signature-style",
	Run:  run,
})

// containsThisType recursively checks whether the given type node (or any of
// its descendants) is a `this` type reference.
func containsThisType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindThisType {
		return true
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if containsThisType(child) {
			found = true
		}
		return found
	})
	return found
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opt := mode(utils.GetOptionsString(options))
	if opt == "" {
		opt = modeProperty
	}

	sourceText := ctx.SourceFile.Text()

	// safeSlice returns sourceText[start:end] with bounds clamping.
	safeSlice := func(start, end int) string {
		if start < 0 {
			start = 0
		}
		if end > len(sourceText) {
			end = len(sourceText)
		}
		return sourceText[start:end]
	}

	// getMethodKey returns the key text for a method or property signature,
	// including optional marker and readonly prefix.
	getMethodKey := func(node *ast.Node) string {
		var nameNode *ast.Node
		switch node.Kind {
		case ast.KindMethodSignature:
			nameNode = node.AsMethodSignatureDeclaration().Name()
		case ast.KindPropertySignature:
			nameNode = node.AsPropertySignatureDeclaration().Name()
		default:
			return ""
		}

		// TrimmedNodeText handles computed property names correctly —
		// ComputedPropertyName nodes already include the surrounding brackets.
		key := utils.TrimmedNodeText(ctx.SourceFile, nameNode)

		if ast.HasQuestionToken(node) {
			key += "?"
		}
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly) {
			key = "readonly " + key
		}
		return key
	}

	// getMethodParams returns the type parameters + parameters text.
	// E.g. "<T>(a: T, b: T)" or "()" for no params.
	getMethodParams := func(node *ast.Node) string {
		var params *ast.NodeList
		var typeParams *ast.NodeList

		switch node.Kind {
		case ast.KindMethodSignature:
			sig := node.AsMethodSignatureDeclaration()
			params = sig.Parameters
			typeParams = sig.TypeParameters
		case ast.KindFunctionType:
			ft := node.AsFunctionTypeNode()
			params = ft.Parameters
			typeParams = ft.TypeParameters
		default:
			return "()"
		}

		paramsText := "()"
		if params != nil && len(params.Nodes) > 0 {
			firstRange := utils.TrimNodeTextRange(ctx.SourceFile, params.Nodes[0])
			lastRange := utils.TrimNodeTextRange(ctx.SourceFile, params.Nodes[len(params.Nodes)-1])
			// Scan backward to find '(' before first param
			openParen := firstRange.Pos() - 1
			for openParen > 0 && sourceText[openParen] != '(' {
				openParen--
			}
			// Scan forward to find ')' after last param
			closeParen := lastRange.End()
			for closeParen < len(sourceText) && sourceText[closeParen] != ')' {
				closeParen++
			}
			paramsText = safeSlice(openParen, closeParen+1)
		}

		if typeParams != nil && len(typeParams.Nodes) > 0 {
			firstRange := utils.TrimNodeTextRange(ctx.SourceFile, typeParams.Nodes[0])
			lastRange := utils.TrimNodeTextRange(ctx.SourceFile, typeParams.Nodes[len(typeParams.Nodes)-1])
			// Scan backward/forward to find '<'/'>' (comments may sit between the bracket and the first/last param)
			start := firstRange.Pos() - 1
			for start > 0 && sourceText[start] != '<' {
				start--
			}
			end := lastRange.End()
			for end < len(sourceText) && sourceText[end] != '>' {
				end++
			}
			paramsText = safeSlice(start, end+1) + paramsText
		}

		return paramsText
	}

	// getMethodReturnType returns the return type text, or "any" if omitted.
	getMethodReturnType := func(node *ast.Node) string {
		var typeNode *ast.Node
		switch node.Kind {
		case ast.KindMethodSignature:
			typeNode = node.AsMethodSignatureDeclaration().Type
		case ast.KindFunctionType:
			typeNode = node.AsFunctionTypeNode().Type
		}
		if typeNode == nil {
			return "any"
		}
		return utils.TrimmedNodeText(ctx.SourceFile, typeNode)
	}

	// getDelimiter returns the trailing ';' or ',' of a member node, or "".
	getDelimiter := func(node *ast.Node) string {
		end := utils.TrimNodeTextRange(ctx.SourceFile, node).End()
		if end > 0 && end <= len(sourceText) {
			if ch := sourceText[end-1]; ch == ';' || ch == ',' {
				return string(ch)
			}
		}
		return ""
	}

	listeners := rule.RuleListeners{}

	if opt == modeProperty {
		listeners[ast.KindMethodSignature] = func(node *ast.Node) {
			// In typescript-go AST, get/set accessors use KindGetAccessor/KindSetAccessor,
			// so KindMethodSignature always represents an actual method — no kind check needed.

			skipFix := containsThisType(node.AsMethodSignatureDeclaration().Type)
			isParentModule := ast.FindAncestorKind(node, ast.KindModuleDeclaration) != nil
			thisKey := getMethodKey(node)
			members := node.Parent.Members()

			// Find overloaded method signatures with the same key
			var duplicates []*ast.Node
			for _, member := range members {
				if member.Kind == ast.KindMethodSignature && member != node && getMethodKey(member) == thisKey {
					duplicates = append(duplicates, member)
				}
			}

			if len(duplicates) > 0 {
				if isParentModule || skipFix {
					ctx.ReportNode(node, messageErrorMethod())
					return
				}

				// Sort all overloads by position; only the first provides the fix
				allNodes := append([]*ast.Node{node}, duplicates...)
				sort.Slice(allNodes, func(i, j int) bool {
					return allNodes[i].Pos() < allNodes[j].Pos()
				})
				if allNodes[0] != node {
					ctx.ReportNode(node, messageErrorMethod())
					return
				}

				// Merge overloads into an intersection of function types
				var typeParts []string
				for _, n := range allNodes {
					typeParts = append(typeParts, "("+getMethodParams(n)+" => "+getMethodReturnType(n)+")")
				}
				typeString := strings.Join(typeParts, " & ")
				replacement := getMethodKey(node) + ": " + typeString + getDelimiter(node)

				var fixes []rule.RuleFix
				fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, node, replacement))

				// Remove each duplicate and its preceding whitespace
				for _, dup := range duplicates {
					dupRange := utils.TrimNodeTextRange(ctx.SourceFile, dup)
					start := dupRange.Pos()
					// Consume leading whitespace/newlines back to the previous member's delimiter
					for start > 0 && (sourceText[start-1] == ' ' || sourceText[start-1] == '\t' || sourceText[start-1] == '\n' || sourceText[start-1] == '\r') {
						start--
					}
					fixes = append(fixes, rule.RuleFixRemoveRange(dupRange.WithPos(start).WithEnd(dupRange.End())))
				}

				ctx.ReportNodeWithFixes(node, messageErrorMethod(), fixes...)
				return
			}

			// Single method signature (no overloads)
			if isParentModule || skipFix {
				ctx.ReportNode(node, messageErrorMethod())
			} else {
				replacement := getMethodKey(node) + ": " + getMethodParams(node) + " => " + getMethodReturnType(node) + getDelimiter(node)
				ctx.ReportNodeWithFixes(node, messageErrorMethod(), rule.RuleFixReplace(ctx.SourceFile, node, replacement))
			}
		}
	}

	if opt == modeMethod {
		listeners[ast.KindPropertySignature] = func(node *ast.Node) {
			propSig := node.AsPropertySignatureDeclaration()
			if propSig == nil || propSig.Type == nil || propSig.Type.Kind != ast.KindFunctionType {
				return
			}

			replacement := getMethodKey(node) + getMethodParams(propSig.Type) + ": " + getMethodReturnType(propSig.Type) + getDelimiter(node)
			ctx.ReportNodeWithFixes(node, messageErrorProperty(), rule.RuleFixReplace(ctx.SourceFile, node, replacement))
		}
	}

	return listeners
}
