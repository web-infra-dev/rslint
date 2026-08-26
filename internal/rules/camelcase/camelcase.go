package camelcase

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed camelcase.schema.json
var schemaJSON []byte

// CamelcaseRule enforces ESLint's deliberately narrow definition of camel
// case: internal underscores are rejected, while leading/trailing underscores
// and all-uppercase names are accepted.
var CamelcaseRule = rule.Rule{
	Name:   "camelcase",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		state := camelcaseState{
			ctx:      ctx,
			opts:     opts,
			reported: make(map[int]struct{}),
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				state.identifiers = append(state.identifiers, node)
			},
			ast.KindPrivateIdentifier: func(node *ast.Node) {
				state.privateIdentifiers = append(state.privateIdentifiers, node)
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(_ *ast.Node) {
				state.checkFile()
			},
		}
	},
}

type camelcaseOptions struct {
	ignoreDestructuring bool
	ignoreGlobals       bool
	ignoreImports       bool
	properties          bool
	allowedNames        map[string]struct{}
	allowedPatterns     []*esregexp.RegExp
}

func parseOptions(options []any) camelcaseOptions {
	opts := camelcaseOptions{properties: true}
	if len(options) == 0 {
		return opts
	}
	m, _ := options[0].(map[string]any)
	if m == nil {
		return opts
	}
	if value, ok := m["ignoreDestructuring"].(bool); ok {
		opts.ignoreDestructuring = value
	}
	if value, ok := m["ignoreGlobals"].(bool); ok {
		opts.ignoreGlobals = value
	}
	if value, ok := m["ignoreImports"].(bool); ok {
		opts.ignoreImports = value
	}
	if value, ok := m["properties"].(string); ok {
		opts.properties = value != "never"
	}
	if allow, ok := m["allow"].([]any); ok {
		opts.allowedNames = make(map[string]struct{}, len(allow))
		for _, raw := range allow {
			pattern, ok := raw.(string)
			if !ok {
				continue
			}
			opts.allowedNames[pattern] = struct{}{}
			if re, err := esregexp.Compile(pattern, "u"); err == nil {
				opts.allowedPatterns = append(opts.allowedPatterns, re)
			}
		}
	}
	return opts
}

func (o camelcaseOptions) isGoodName(name string) bool {
	body := strings.Trim(name, "_")
	if !strings.Contains(body, "_") || body == ecmascript.StringToUpperCase(body) {
		return true
	}
	if _, ok := o.allowedNames[name]; ok {
		return true
	}
	for _, pattern := range o.allowedPatterns {
		if pattern.Test(name) {
			return true
		}
	}
	return false
}

type bindingKind uint8

const (
	bindingNone bindingKind = iota
	bindingLocal
	bindingImport
)

type camelcaseState struct {
	ctx                rule.RuleContext
	opts               camelcaseOptions
	identifiers        []*ast.Node
	privateIdentifiers []*ast.Node
	reported           map[int]struct{}
}

func (s *camelcaseState) checkFile() {
	badReferences := make(map[*ast.Node]struct{})
	for _, identifier := range s.identifiers {
		kind := runtimeBindingKind(identifier)
		if kind == bindingNone || s.opts.isGoodName(identifier.AsIdentifier().Text) {
			continue
		}
		if sym := utils.BindingNameSymbol(identifier); sym != nil && s.ctx.Refs != nil {
			for _, reference := range s.ctx.Refs.References(sym) {
				badReferences[reference] = struct{}{}
			}
		}
	}

	for _, identifier := range s.identifiers {
		name := identifier.AsIdentifier().Text
		if s.opts.isGoodName(name) || utils.IsImportAttributeKey(identifier) {
			continue
		}

		switch runtimeBindingKind(identifier) {
		case bindingLocal:
			if !s.opts.ignoreDestructuring || !equalsOriginalName(identifier) {
				s.reportBinding(identifier, name)
			}
			continue
		case bindingImport:
			if !s.opts.ignoreImports || !equalsImportedName(identifier) {
				s.reportBinding(identifier, name)
			}
			continue
		}

		if isPropertyName(identifier) {
			if s.opts.properties {
				s.report(identifier, name, false)
				continue
			}
			// A shorthand object member is both its property key and a
			// reference to the local variable. With properties:"never", only
			// the key half is exempt; keep evaluating the reference half.
			if utils.IsNonReferenceIdentifier(identifier) {
				continue
			}
		}
		if isAssignedPropertyAccessName(identifier) {
			if s.opts.properties {
				s.report(identifier, name, false)
			}
			continue
		}
		if isExportedName(identifier) || isLabel(identifier) {
			s.report(identifier, name, false)
			continue
		}

		if _, ok := badReferences[identifier]; ok {
			s.reportReference(identifier, name)
			continue
		}
		if !isRuntimeReference(identifier) {
			continue
		}
		if s.ctx.Refs != nil && s.ctx.Refs.IsDefinedInFile(identifier) {
			// The identifier resolves to a local declaration kind that ESLint's
			// camelcase visitors do not inspect (for example, a TS-only binding).
			continue
		}
		if s.opts.ignoreGlobals && s.ctx.Globals.Access(name).IsDeclared() {
			continue
		}
		s.reportReference(identifier, name)
	}

	if s.opts.properties {
		for _, identifier := range s.privateIdentifiers {
			if !isPrivatePropertyName(identifier) {
				continue
			}
			name := strings.TrimPrefix(identifier.AsPrivateIdentifier().Text, "#")
			if !s.opts.isGoodName(name) {
				s.report(identifier, name, true)
			}
		}
	}
}

func (s *camelcaseState) reportBinding(node *ast.Node, name string) {
	if _, exists := s.reported[node.Pos()]; exists {
		return
	}
	s.reported[node.Pos()] = struct{}{}
	s.ctx.ReportRange(utils.GetESTreeBindingIdentifierRange(s.ctx.SourceFile, node), camelcaseMessage(name, false))
}

func (s *camelcaseState) reportReference(node *ast.Node, name string) {
	outer := utils.OutermostParenthesizedExpression(node)
	parent := outer.Parent
	if parent != nil {
		if parent.Kind == ast.KindCallExpression || parent.Kind == ast.KindNewExpression {
			return
		}
		if isDefaultValue(parent, outer) {
			return
		}
	}
	if s.opts.ignoreDestructuring && equalsOriginalName(node) {
		return
	}
	s.report(node, name, false)
}

func (s *camelcaseState) report(node *ast.Node, name string, private bool) {
	if _, exists := s.reported[node.Pos()]; exists {
		return
	}
	s.reported[node.Pos()] = struct{}{}
	s.ctx.ReportNode(node, camelcaseMessage(name, private))
}

func camelcaseMessage(name string, private bool) rule.RuleMessage {
	if private {
		return rule.RuleMessage{
			Id:          "notCamelCasePrivate",
			Description: fmt.Sprintf("#%s is not in camel case.", name),
			Data:        map[string]string{"name": name},
		}
	}
	return rule.RuleMessage{
		Id:          "notCamelCase",
		Description: fmt.Sprintf("Identifier '%s' is not in camel case.", name),
		Data:        map[string]string{"name": name},
	}
}

func runtimeBindingKind(node *ast.Node) bindingKind {
	if node == nil || node.Parent == nil || node.Parent.Name() != node {
		return bindingNone
	}
	parent := node.Parent
	switch parent.Kind {
	case ast.KindImportClause, ast.KindImportSpecifier, ast.KindNamespaceImport:
		return bindingImport
	case ast.KindVariableDeclaration:
		return bindingLocal
	case ast.KindParameter:
		owner := parent.Parent
		if owner != nil && ast.IsFunctionLikeDeclaration(owner) && owner.Body() != nil {
			return bindingLocal
		}
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		if parent.Body() != nil {
			return bindingLocal
		}
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return bindingLocal
	case ast.KindBindingElement:
		owner := ast.WalkUpBindingElementsAndPatterns(parent)
		if owner == nil {
			return bindingNone
		}
		switch owner.Kind {
		case ast.KindVariableDeclaration:
			return bindingLocal
		case ast.KindParameter:
			function := owner.Parent
			if function != nil && ast.IsFunctionLikeDeclaration(function) && function.Body() != nil {
				return bindingLocal
			}
		}
	}
	return bindingNone
}

func equalsImportedName(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindImportSpecifier {
		return false
	}
	specifier := node.Parent.AsImportSpecifier()
	if specifier.Name() != node {
		return false
	}
	imported := specifier.PropertyName
	if imported == nil {
		imported = specifier.Name()
	}
	name, ok := utils.GetStaticPropertyName(imported)
	return ok && name == node.AsIdentifier().Text
}

func equalsOriginalName(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	name := node.AsIdentifier().Text
	if node.Parent.Kind == ast.KindBindingElement {
		binding := node.Parent.AsBindingElement()
		if binding.Name() != node || node.Parent.Parent == nil || node.Parent.Parent.Kind != ast.KindObjectBindingPattern {
			return false
		}
		if binding.DotDotDotToken != nil {
			return false
		}
		propertyName := binding.PropertyName
		if propertyName == nil {
			return true
		}
		property, ok := plainIdentifierName(propertyName)
		return ok && property == name
	}

	outer := utils.OutermostParenthesizedExpression(node)
	parent := outer.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindShorthandPropertyAssignment:
		return parent.Name() == node
	case ast.KindPropertyAssignment:
		assignment := parent.AsPropertyAssignment()
		if assignment.Initializer != outer {
			return false
		}
		property, ok := plainIdentifierName(parent.Name())
		return ok && property == name
	}
	return false
}

func plainIdentifierName(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindIdentifier {
		return "", false
	}
	return node.AsIdentifier().Text, true
}

func isDefaultValue(parent, child *ast.Node) bool {
	switch parent.Kind {
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Initializer == child
	case ast.KindBindingElement:
		return parent.AsBindingElement().Initializer == child
	case ast.KindShorthandPropertyAssignment:
		return parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer == child
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == child &&
			utils.IsInDestructuringAssignment(parent)
	}
	return false
}

func isPropertyName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Name() != node {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if parent.Parent == nil || parent.Parent.Kind != ast.KindObjectLiteralExpression {
			return parent.Parent != nil &&
				(parent.Parent.Kind == ast.KindClassDeclaration || parent.Parent.Kind == ast.KindClassExpression) &&
				utils.IsPlainClassMember(parent)
		}
		return !utils.IsInDestructuringAssignment(parent)
	case ast.KindPropertyDeclaration:
		return parent.Parent != nil &&
			(parent.Parent.Kind == ast.KindClassDeclaration || parent.Parent.Kind == ast.KindClassExpression) &&
			utils.IsPlainClassMember(parent)
	}
	return false
}

func isPrivatePropertyName(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Name() != node {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindPropertyDeclaration:
		owner := node.Parent.Parent
		return owner != nil && (owner.Kind == ast.KindClassDeclaration || owner.Kind == ast.KindClassExpression) &&
			utils.IsPlainClassMember(node.Parent)
	}
	return false
}

func isAssignedPropertyAccessName(node *ast.Node) bool {
	parent := node.Parent
	return parent != nil && parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Name() == node && utils.IsWriteReference(parent)
}

func isExportedName(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindExportSpecifier, ast.KindNamespaceExport:
		return node.Parent.Name() == node
	}
	return false
}

func isLabel(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement:
		return true
	}
	return false
}

func isRuntimeReference(node *ast.Node) bool {
	return node != nil && !ast.IsPartOfTypeNode(node) && !ast.IsPartOfTypeQuery(node) &&
		!utils.IsNonReferenceIdentifier(node)
}
