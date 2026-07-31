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

type methodSignatureFixer struct {
	sourceFile *ast.SourceFile
	sourceText string
}

// getMethodKey returns the key text for a method or property signature,
// including optional marker and readonly prefix.
func (f *methodSignatureFixer) getMethodKey(node *ast.Node) string {
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
	key := utils.TrimmedNodeText(f.sourceFile, nameNode)

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
func (f *methodSignatureFixer) getMethodParams(node *ast.Node) string {
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
		firstRange := utils.TrimNodeTextRange(f.sourceFile, params.Nodes[0])
		lastRange := utils.TrimNodeTextRange(f.sourceFile, params.Nodes[len(params.Nodes)-1])
		// Widen the param-list span outward to the enclosing `(`/`)` pair
		// (comments may sit between a paren and the first/last param).
		paramsText = utils.SliceEnclosingDelimiters(
			f.sourceText, firstRange.Pos(), lastRange.End(), '(', ')',
		)
	}

	if typeParams != nil && len(typeParams.Nodes) > 0 {
		firstRange := utils.TrimNodeTextRange(f.sourceFile, typeParams.Nodes[0])
		lastRange := utils.TrimNodeTextRange(f.sourceFile, typeParams.Nodes[len(typeParams.Nodes)-1])
		// Widen the type-param-list span outward to the enclosing `<`/`>` pair.
		paramsText = utils.SliceEnclosingDelimiters(
			f.sourceText, firstRange.Pos(), lastRange.End(), '<', '>',
		) + paramsText
	}

	return paramsText
}

// getMethodReturnType returns the return type text, or "any" if omitted.
func (f *methodSignatureFixer) getMethodReturnType(node *ast.Node) string {
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
	return utils.TrimmedNodeText(f.sourceFile, typeNode)
}

// getDelimiter returns the trailing ';' or ',' of a member node, or "".
func (f *methodSignatureFixer) getDelimiter(node *ast.Node) string {
	end := utils.TrimNodeTextRange(f.sourceFile, node).End()
	if end > 0 && end <= len(f.sourceText) {
		if ch := f.sourceText[end-1]; ch == ';' || ch == ',' {
			return string(ch)
		}
	}
	return ""
}

func (f *methodSignatureFixer) buildMethodFixes(node *ast.Node) []rule.RuleFix {
	signature := node.AsMethodSignatureDeclaration()
	if containsThisType(signature.Type) ||
		ast.FindAncestorKind(node, ast.KindModuleDeclaration) != nil {
		return nil
	}

	thisKey := f.getMethodKey(node)
	members := node.Parent.Members()

	// Find overloaded method signatures with the same key.
	var duplicates []*ast.Node
	for _, member := range members {
		if member.Kind == ast.KindMethodSignature && member != node && f.getMethodKey(member) == thisKey {
			duplicates = append(duplicates, member)
		}
	}

	if len(duplicates) > 0 {
		// Sort all overloads by position; only the first provides the fix.
		allNodes := append([]*ast.Node{node}, duplicates...)
		sort.Slice(allNodes, func(i, j int) bool {
			return allNodes[i].Pos() < allNodes[j].Pos()
		})
		if allNodes[0] != node {
			return nil
		}

		// Merge overloads into an intersection of function types.
		typeParts := make([]string, 0, len(allNodes))
		for _, overload := range allNodes {
			typeParts = append(
				typeParts,
				"("+f.getMethodParams(overload)+" => "+f.getMethodReturnType(overload)+")",
			)
		}
		typeString := strings.Join(typeParts, " & ")
		replacement := f.getMethodKey(node) + ": " + typeString + f.getDelimiter(node)

		fixes := make([]rule.RuleFix, 0, len(duplicates)+1)
		fixes = append(fixes, rule.RuleFixReplace(f.sourceFile, node, replacement))

		// Remove each duplicate and its preceding whitespace.
		for _, duplicate := range duplicates {
			duplicateRange := utils.TrimNodeTextRange(f.sourceFile, duplicate)
			start := duplicateRange.Pos()
			// Consume leading whitespace/newlines back to the previous member's delimiter.
			for start > 0 &&
				(f.sourceText[start-1] == ' ' ||
					f.sourceText[start-1] == '\t' ||
					f.sourceText[start-1] == '\n' ||
					f.sourceText[start-1] == '\r') {
				start--
			}
			fixes = append(
				fixes,
				rule.RuleFixRemoveRange(duplicateRange.WithPos(start).WithEnd(duplicateRange.End())),
			)
		}

		return fixes
	}

	replacement := f.getMethodKey(node) +
		": " +
		f.getMethodParams(node) +
		" => " +
		f.getMethodReturnType(node) +
		f.getDelimiter(node)
	return []rule.RuleFix{rule.RuleFixReplace(f.sourceFile, node, replacement)}
}

func (f *methodSignatureFixer) buildPropertyFixes(
	node *ast.Node,
	functionType *ast.Node,
) []rule.RuleFix {
	replacement := f.getMethodKey(node) +
		f.getMethodParams(functionType) +
		": " +
		f.getMethodReturnType(functionType) +
		f.getDelimiter(node)
	return []rule.RuleFix{rule.RuleFixReplace(f.sourceFile, node, replacement)}
}

func run(ctx rule.RuleContext, _options []any) rule.RuleListeners {
	options := rule.LegacyUnwrapOptions(_options)
	opt := mode(utils.GetOptionsString(options))
	if opt == "" {
		opt = modeProperty
	}

	fixer := methodSignatureFixer{
		sourceFile: ctx.SourceFile,
		sourceText: ctx.SourceFile.Text(),
	}
	listeners := rule.RuleListeners{}

	if opt == modeProperty {
		listeners[ast.KindMethodSignature] = func(node *ast.Node) {
			// In typescript-go AST, get/set accessors use KindGetAccessor/KindSetAccessor,
			// so KindMethodSignature always represents an actual method — no kind check needed.
			ctx.ReportNodeWithDeferredFixes(node, messageErrorMethod(), func() []rule.RuleFix {
				return fixer.buildMethodFixes(node)
			})
		}
	}

	if opt == modeMethod {
		listeners[ast.KindPropertySignature] = func(node *ast.Node) {
			propSig := node.AsPropertySignatureDeclaration()
			if propSig == nil || propSig.Type == nil || propSig.Type.Kind != ast.KindFunctionType {
				return
			}

			ctx.ReportNodeWithDeferredFixes(node, messageErrorProperty(), func() []rule.RuleFix {
				return fixer.buildPropertyFixes(node, propSig.Type)
			})
		}
	}

	return listeners
}
