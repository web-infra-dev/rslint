package class_methods_use_this

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ignoreInterfaceMode int

const (
	ignoreInterfaceNone ignoreInterfaceMode = iota
	ignoreInterfaceAll
	ignoreInterfacePublic
)

type classMethodsUseThisOptions struct {
	ExceptMethods                  []string
	EnforceForClassFields          bool
	IgnoreOverrideMethods          bool
	IgnoreClassesThatImplementMode ignoreInterfaceMode
}

type memberNameInfo struct {
	Name      string
	HasName   bool
	IsPrivate bool
}

var ClassMethodsUseThisRule = rule.CreateRule(rule.Rule{
	Name: "class-methods-use-this",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		exceptMethods := make(map[string]struct{}, len(opts.ExceptMethods))
		for _, method := range opts.ExceptMethods {
			exceptMethods[method] = struct{}{}
		}

		shouldIgnoreForInterface := func(classNode *ast.Node, member *ast.Node, nameInfo memberNameInfo) bool {
			if opts.IgnoreClassesThatImplementMode == ignoreInterfaceNone {
				return false
			}
			if !classImplementsInterface(classNode) {
				return false
			}
			if opts.IgnoreClassesThatImplementMode == ignoreInterfaceAll {
				return true
			}
			return isPublicMember(member, nameInfo)
		}

		reportMissingThis := func(member *ast.Node, startNode *ast.Node, nameNode *ast.Node, initializer *ast.Node, displayName string) {
			msg := buildMissingThisMessage(displayName)
			if startNode != nil {
				if reportRange, ok := buildMemberReportRange(ctx.SourceFile, member, startNode, nameNode, initializer); ok {
					ctx.ReportRange(reportRange, msg)
					return
				}
			}
			ctx.ReportNode(member, msg)
		}

		checkMethodLike := func(member *ast.Node, body *ast.Node, kind string, startNode *ast.Node) {
			if member == nil || body == nil {
				return
			}

			classNode := getParentClass(member)
			if classNode == nil {
				return
			}

			nameNode := getMemberNameNode(member)
			if startNode == nil {
				startNode = nameNode
			}
			nameInfo := getMemberNameInfo(ctx.SourceFile, nameNode)

			if opts.IgnoreOverrideMethods && ast.HasSyntacticModifier(member, ast.ModifierFlagsOverride) {
				return
			}

			if ast.IsStatic(member) {
				return
			}

			if shouldIgnoreForInterface(classNode, member, nameInfo) {
				return
			}

			if nameInfo.HasName {
				if _, ok := exceptMethods[nameInfo.Name]; ok {
					return
				}
			}

			if containsThisOrSuper(body) {
				return
			}

			displayName := buildMemberDisplayName(kind, nameInfo)
			reportMissingThis(member, startNode, nameNode, nil, displayName)
		}

		checkProperty := func(member *ast.Node) {
			if member == nil {
				return
			}

			if !opts.EnforceForClassFields {
				return
			}

			classNode := getParentClass(member)
			if classNode == nil {
				return
			}

			if ast.IsStatic(member) {
				return
			}

			nameNode := getMemberNameNode(member)
			nameInfo := getMemberNameInfo(ctx.SourceFile, nameNode)

			if opts.IgnoreOverrideMethods && ast.HasSyntacticModifier(member, ast.ModifierFlagsOverride) {
				return
			}

			if shouldIgnoreForInterface(classNode, member, nameInfo) {
				return
			}

			prop := member.AsPropertyDeclaration()
			if prop == nil || prop.Initializer == nil {
				return
			}

			if nameInfo.HasName {
				if _, ok := exceptMethods[nameInfo.Name]; ok {
					return
				}
			}

			init := prop.Initializer
			var body *ast.Node
			switch init.Kind {
			case ast.KindFunctionExpression:
				fn := init.AsFunctionExpression()
				if fn != nil {
					body = fn.Body
				}
			case ast.KindArrowFunction:
				arrow := init.AsArrowFunction()
				if arrow != nil {
					body = arrow.Body
				}
			default:
				return
			}

			if body == nil {
				return
			}

			if containsThisOrSuper(body) {
				return
			}

			displayName := buildMemberDisplayName("method", nameInfo)
			reportMissingThis(member, nameNode, nameNode, init, displayName)
		}

		return rule.RuleListeners{
			ast.KindMethodDeclaration: func(node *ast.Node) {
				method := node.AsMethodDeclaration()
				if method == nil || method.Body == nil {
					return
				}
				kind := "method"
				var startNode *ast.Node
				if method.AsteriskToken != nil {
					kind = "generator method"
					startNode = node
				}
				checkMethodLike(node, method.Body, kind, startNode)
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				accessor := node.AsGetAccessorDeclaration()
				if accessor == nil || accessor.Body == nil {
					return
				}
				checkMethodLike(node, accessor.Body, "getter", node)
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				accessor := node.AsSetAccessorDeclaration()
				if accessor == nil || accessor.Body == nil {
					return
				}
				checkMethodLike(node, accessor.Body, "setter", node)
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				checkProperty(node)
			},
		}
	},
})

func parseOptions(options any) classMethodsUseThisOptions {
	opts := classMethodsUseThisOptions{
		ExceptMethods:                  []string{},
		EnforceForClassFields:          true,
		IgnoreOverrideMethods:          false,
		IgnoreClassesThatImplementMode: ignoreInterfaceNone,
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	if arr, ok := options.([]interface{}); ok && len(arr) > 0 {
		optsMap, _ = arr[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap == nil {
		return opts
	}

	if enforce, ok := optsMap["enforceForClassFields"].(bool); ok {
		opts.EnforceForClassFields = enforce
	}

	if ignoreOverride, ok := optsMap["ignoreOverrideMethods"].(bool); ok {
		opts.IgnoreOverrideMethods = ignoreOverride
	}

	if except, ok := optsMap["exceptMethods"].([]interface{}); ok {
		for _, item := range except {
			if name, ok := item.(string); ok {
				opts.ExceptMethods = append(opts.ExceptMethods, name)
			}
		}
	}

	if ignoreInterfaces, ok := optsMap["ignoreClassesThatImplementAnInterface"]; ok {
		switch value := ignoreInterfaces.(type) {
		case bool:
			if value {
				opts.IgnoreClassesThatImplementMode = ignoreInterfaceAll
			} else {
				opts.IgnoreClassesThatImplementMode = ignoreInterfaceNone
			}
		case string:
			if value == "public-fields" {
				opts.IgnoreClassesThatImplementMode = ignoreInterfacePublic
			}
		}
	}

	return opts
}

func buildMissingThisMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingThis",
		Description: fmt.Sprintf("Expected 'this' to be used by class %s.", name),
	}
}

func buildMemberDisplayName(kind string, nameInfo memberNameInfo) string {
	if nameInfo.IsPrivate {
		return fmt.Sprintf("private %s %s", kind, nameInfo.Name)
	}
	if nameInfo.HasName {
		return fmt.Sprintf("%s '%s'", kind, nameInfo.Name)
	}
	return kind
}

func getParentClass(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	parent := node.Parent
	if parent == nil {
		return nil
	}
	if ast.IsClassDeclaration(parent) || ast.IsClassExpression(parent) {
		return parent
	}
	return nil
}

func isPublicMember(node *ast.Node, nameInfo memberNameInfo) bool {
	if nameInfo.IsPrivate {
		return false
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate) || ast.HasSyntacticModifier(node, ast.ModifierFlagsProtected) {
		return false
	}
	return true
}

func classImplementsInterface(classNode *ast.Node) bool {
	clauses := utils.GetHeritageClauses(classNode)
	if clauses == nil {
		return false
	}
	for _, clause := range clauses.Nodes {
		heritage := clause.AsHeritageClause()
		if heritage == nil {
			continue
		}
		if heritage.Token == ast.KindImplementsKeyword && heritage.Types != nil && len(heritage.Types.Nodes) > 0 {
			return true
		}
	}
	return false
}

func getMemberNameNode(member *ast.Node) *ast.Node {
	switch member.Kind {
	case ast.KindMethodDeclaration:
		method := member.AsMethodDeclaration()
		if method != nil {
			return method.Name()
		}
	case ast.KindGetAccessor:
		accessor := member.AsGetAccessorDeclaration()
		if accessor != nil {
			return accessor.Name()
		}
	case ast.KindSetAccessor:
		accessor := member.AsSetAccessorDeclaration()
		if accessor != nil {
			return accessor.Name()
		}
	case ast.KindPropertyDeclaration:
		prop := member.AsPropertyDeclaration()
		if prop != nil {
			return prop.Name()
		}
	}
	return nil
}

func getMemberNameInfo(sourceFile *ast.SourceFile, nameNode *ast.Node) memberNameInfo {
	if nameNode == nil {
		return memberNameInfo{}
	}

	switch nameNode.Kind {
	case ast.KindIdentifier:
		ident := nameNode.AsIdentifier()
		if ident != nil {
			return memberNameInfo{Name: ident.Text, HasName: true}
		}
	case ast.KindPrivateIdentifier:
		privateIdent := nameNode.AsPrivateIdentifier()
		if privateIdent != nil {
			name := privateIdent.Text
			if !strings.HasPrefix(name, "#") {
				name = "#" + name
			}
			return memberNameInfo{Name: name, HasName: true, IsPrivate: true}
		}
	case ast.KindComputedPropertyName:
		computed := nameNode.AsComputedPropertyName()
		if computed != nil {
			if name, ok := literalNameFromExpression(computed.Expression); ok {
				return memberNameInfo{Name: name, HasName: true}
			}
		}
	default:
		if ast.IsLiteralExpression(nameNode) {
			if name, ok := literalNameFromExpression(nameNode); ok {
				return memberNameInfo{Name: name, HasName: true}
			}
		}
	}

	return memberNameInfo{}
}

func literalNameFromExpression(expr *ast.Node) (string, bool) {
	if expr == nil {
		return "", false
	}

	if ast.IsLiteralExpression(expr) || expr.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return trimLiteralText(expr.Text()), true
	}

	if expr.Kind == ast.KindTemplateExpression {
		template := expr.AsTemplateExpression()
		if template != nil && len(template.TemplateSpans.Nodes) == 0 {
			return trimLiteralText(expr.Text()), true
		}
	}

	return "", false
}

func trimLiteralText(text string) string {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) >= 2 {
		first := trimmed[0]
		last := trimmed[len(trimmed)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') || (first == '`' && last == '`') {
			return trimmed[1 : len(trimmed)-1]
		}
	}
	return trimmed
}

func containsThisOrSuper(node *ast.Node) bool {
	found := false
	var visit func(n *ast.Node)
	visitClass := func(n *ast.Node) {
		if n == nil || found {
			return
		}

		if clauses := utils.GetHeritageClauses(n); clauses != nil {
			for _, clause := range clauses.Nodes {
				visit(clause)
				if found {
					return
				}
			}
		}

		if modifiers := n.Modifiers(); modifiers != nil {
			for _, mod := range modifiers.Nodes {
				if mod.Kind == ast.KindDecorator {
					visit(mod)
					if found {
						return
					}
				}
			}
		}

		var members *ast.NodeList
		if ast.IsClassDeclaration(n) {
			members = n.AsClassDeclaration().Members
		} else if ast.IsClassExpression(n) {
			members = n.AsClassExpression().Members
		}

		if members == nil {
			return
		}

		for _, member := range members.Nodes {
			if nameNode := getMemberNameNode(member); nameNode != nil && nameNode.Kind == ast.KindComputedPropertyName {
				computed := nameNode.AsComputedPropertyName()
				if computed != nil {
					visit(computed.Expression)
					if found {
						return
					}
				}
			}

			if modifiers := member.Modifiers(); modifiers != nil {
				for _, mod := range modifiers.Nodes {
					if mod.Kind == ast.KindDecorator {
						visit(mod)
						if found {
							return
						}
					}
				}
			}
		}
	}

	visit = func(n *ast.Node) {
		if n == nil || found {
			return
		}
		switch n.Kind {
		case ast.KindThisKeyword, ast.KindSuperKeyword:
			found = true
			return
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			return
		case ast.KindClassDeclaration, ast.KindClassExpression:
			visitClass(n)
			return
		}

		n.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}

	visit(node)
	return found
}

func buildMemberReportRange(sourceFile *ast.SourceFile, member *ast.Node, startNode *ast.Node, nameNode *ast.Node, initializer *ast.Node) (core.TextRange, bool) {
	if sourceFile == nil || startNode == nil {
		return core.TextRange{}, false
	}

	startRange := utils.TrimNodeTextRange(sourceFile, startNode)
	start := startRange.Pos()

	searchStart := startNode.End()
	if nameNode != nil {
		searchStart = nameNode.End()
	}
	searchEnd := member.End()
	if initializer != nil {
		searchStart = initializer.Pos()
		searchEnd = initializer.End()
	}

	parenPos := findNextParen(sourceFile.Text(), searchStart, searchEnd)
	if parenPos == -1 || parenPos <= start {
		return core.TextRange{}, false
	}

	return core.NewTextRange(start, parenPos), true
}

func findNextParen(text string, start int, end int) int {
	if start < 0 {
		start = 0
	}
	if end <= 0 || end > len(text) {
		end = len(text)
	}
	for i := start; i < end; i++ {
		if text[i] == '(' {
			return i
		}
	}
	return -1
}
