package no_use_before_define

import (
	"cmp"
	_ "embed"
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

//go:embed no_use_before_define.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/no-use-before-define
//
// The rule walks every identifier reference recovered from the shared scope
// model in `internal/utils/scope` and reports the ones that read a binding
// declared later in the file, or that run while that binding is still being
// initialized.

type options struct {
	functions            bool
	classes              bool
	variables            bool
	allowNamedExports    bool
	enums                bool
	typedefs             bool
	ignoreTypeReferences bool
}

func defaultOptions() options {
	return options{
		functions:            true,
		classes:              true,
		variables:            true,
		allowNamedExports:    false,
		enums:                true,
		typedefs:             true,
		ignoreTypeReferences: true,
	}
}

func parseOptions(rawOptions []any) options {
	opts := defaultOptions()
	if len(rawOptions) == 0 {
		return opts
	}
	// The legacy string form only turns function declarations off.
	if s, ok := rawOptions[0].(string); ok {
		opts.functions = s != "nofunc"
		return opts
	}
	optsMap, _ := rawOptions[0].(map[string]any)
	readBool := func(key string, target *bool) {
		if v, ok := optsMap[key].(bool); ok {
			*target = v
		}
	}
	readBool("functions", &opts.functions)
	readBool("classes", &opts.classes)
	readBool("variables", &opts.variables)
	readBool("allowNamedExports", &opts.allowNamedExports)
	readBool("enums", &opts.enums)
	readBool("typedefs", &opts.typedefs)
	readBool("ignoreTypeReferences", &opts.ignoreTypeReferences)
	return opts
}

var NoUseBeforeDefineRule = rule.Rule{
	Name:   "no-use-before-define",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		if ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}

		manager := scope.Build(ctx.SourceFile, scope.Options{CollectReferences: true})

		// ESLint walks the scope tree depth-first; ordering the flat reference
		// list by source position gives the same sequence for every input the
		// rule can report on, without depending on scope-creation order.
		references := slices.Clone(manager.References)
		slices.SortStableFunc(references, func(a, b *scope.Reference) int {
			return cmp.Compare(a.Identifier.Pos(), b.Identifier.Pos())
		})

		for _, ref := range references {
			if !shouldCheck(ref, opts) {
				continue
			}
			declaration := ref.Resolved()
			if ref.Identifier.End() < declaration.ID.End() ||
				(isEvaluatedDuringInitialization(ref) && !isTypeReference(ref.Identifier)) {
				ctx.ReportNode(ref.Identifier, rule.RuleMessage{
					Id:          "usedBeforeDefined",
					Description: fmt.Sprintf("'%s' was used before it was defined.", ref.Identifier.Text()),
				})
			}
		}

		return rule.RuleListeners{}
	},
}

// shouldCheck filters out references the options exempt, references to
// bindings declared outside this file, and reference positions that don't
// carry a use-before-define hazard.
func shouldCheck(ref *scope.Reference, opts options) bool {
	identifier := ref.Identifier

	if opts.allowNamedExports && isExportSpecifierLocalName(identifier) {
		return false
	}

	declaration := ref.Resolved()
	if declaration == nil {
		return false
	}

	// A binding with no identifier — a string-literal enum member — has no
	// declaration site to point the reader at, so upstream leaves it alone.
	if declaration.Anonymous {
		return false
	}

	if !opts.functions && isFunctionNameDef(declaration) {
		return false
	}

	if ((!opts.variables && declaration.Kind == scope.DefVariable) ||
		(!opts.classes && isClassNameDef(declaration))) &&
		// Don't skip a reference in the same execution context: that one is a
		// temporal dead zone error regardless of the option.
		isFromSeparateExecutionContext(ref) {
		return false
	}

	if !opts.enums && declaration.Kind == scope.DefEnumName {
		return false
	}

	// Type parameters are `Type` definitions upstream, so `typedefs` covers
	// them alongside interfaces and type aliases.
	if !opts.typedefs &&
		(declaration.Kind == scope.DefType || declaration.Kind == scope.DefTypeParameter) {
		return false
	}

	if opts.ignoreTypeReferences &&
		(referenceContainsTypeQuery(identifier) || isTypeReference(identifier)) {
		return false
	}

	// `import Z = A.X.Y` and `typeof A.X.Y` only read the left-most name; the
	// rest are namespace members, not bindings.
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindQualifiedName {
		return leftMostQualifiedName(identifier.Parent) == identifier
	}

	// Decorators are evaluated after the class binding is initialized, so a
	// self-reference in one is safe.
	if isClassRefInClassDecorator(declaration, ref) {
		return false
	}

	if isFromFunctionTypeScope(ref) {
		return false
	}

	return true
}

// isFromFunctionTypeScope reports whether the reference is evaluated in the
// scope a TS function type puts its type parameters and parameters in. Nothing
// in such a signature runs, so upstream never reports a reference from one.
func isFromFunctionTypeScope(ref *scope.Reference) bool {
	from := ref.From
	if from == nil || from.Kind != scope.KindFunction || from.Block == nil {
		return false
	}
	block := from.Block
	return ast.IsFunctionTypeNode(block) || ast.IsConstructorTypeNode(block) ||
		ast.IsCallSignatureDeclaration(block) || ast.IsConstructSignatureDeclaration(block) ||
		ast.IsMethodSignatureDeclaration(block)
}

func isFunctionNameDef(v *scope.Variable) bool {
	return v.Kind == scope.DefFunctionName || v.Kind == scope.DefFnExprName
}

func isClassNameDef(v *scope.Variable) bool {
	return v.Kind == scope.DefClassName || v.Kind == scope.DefClassInnerName
}

// isClassStaticInitializerScope reports whether a scope is a static block or
// the initializer of a static field. Both run while the class definition is
// being evaluated, so they belong to the enclosing execution context.
func isClassStaticInitializerScope(s *scope.Scope) bool {
	if s == nil {
		return false
	}
	switch s.Kind {
	case scope.KindClassStaticBlock:
		return true
	case scope.KindClassFieldInitializer:
		// Scope.Block is the initializer expression; its parent is the field.
		return s.Block != nil && s.Block.Parent != nil && ast.HasStaticModifier(s.Block.Parent)
	}
	return false
}

// isFromSeparateExecutionContext reports whether the reference is evaluated in
// a different execution context than the one its binding is declared in.
// Execution contexts are the top level, functions, class field initializers,
// and class static blocks — with static field initializers and static blocks
// folded into their parent context because they run during class definition.
func isFromSeparateExecutionContext(ref *scope.Reference) bool {
	declaration := ref.Resolved()
	if declaration == nil || declaration.Scope == nil {
		return false
	}
	declarationContext := declaration.Scope.VariableScope()
	for current := ref.From; current != nil; {
		context := current.VariableScope()
		if context == declarationContext {
			return false
		}
		if !isClassStaticInitializerScope(context) {
			return true
		}
		current = context.Parent
	}
	return true
}

// isEvaluatedDuringInitialization reports whether the reference runs while the
// binding it reads is still being initialized — `var a = a`, `class C extends
// C {}`, and friends.
func isEvaluatedDuringInitialization(ref *scope.Reference) bool {
	if isFromSeparateExecutionContext(ref) {
		// Even inside the initializer, the reference isn't evaluated during
		// initialization: `const x = () => x;` is fine.
		return false
	}

	location := ref.Identifier.End()
	declaration := ref.Resolved()

	if isClassNameDef(declaration) {
		classDefinition := declaration.DefNode
		// The class binding is initialized before its static initializers run,
		// so `class C { static foo = C; static { bar = C; } }` is valid.
		return isInRange(classDefinition, location) &&
			!isInClassStaticInitializerRange(classDefinition, location)
	}

	for node := declaration.ID.Parent; node != nil; node = node.Parent {
		switch node.Kind {
		case ast.KindVariableDeclaration:
			declarator := node.AsVariableDeclaration()
			if declarator != nil && isInRange(declarator.Initializer, location) {
				return true
			}
			// `for (var a in a)` / `for (var a of a)`: the right-hand side is
			// evaluated before the binding is initialized.
			if node.Parent != nil && node.Parent.Parent != nil {
				loop := node.Parent.Parent
				if loop.Kind == ast.KindForInStatement || loop.Kind == ast.KindForOfStatement {
					if forInOrOf := loop.AsForInOrOfStatement(); forInOrOf != nil &&
						isInRange(forInOrOf.Expression, location) {
						return true
					}
				}
			}
			return false
		case ast.KindBindingElement:
			// ESTree's AssignmentPattern inside a destructuring pattern.
			if element := node.AsBindingElement(); element != nil && isInRange(element.Initializer, location) {
				return true
			}
		case ast.KindParameter:
			// ESTree's AssignmentPattern in a parameter list.
			if isInRange(node.Initializer(), location) {
				return true
			}
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindClassDeclaration, ast.KindClassExpression,
			ast.KindArrowFunction, ast.KindCatchClause,
			ast.KindImportDeclaration, ast.KindExportDeclaration,
			// ESTree wraps every shorthand method and accessor body in a
			// FunctionExpression, which the walk stops at. tsgo has no such
			// wrapper — the member node *is* the function — so these kinds
			// stand in for it. Without them the walk escapes the method and
			// finds the enclosing `const x = { m(node) { node; } }`
			// initializer, reporting every parameter read inside.
			ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor:
			return false
		}
	}

	return false
}

func isInRange(node *ast.Node, location int) bool {
	return node != nil && node.Pos() <= location && location <= node.End()
}

// isInClassStaticInitializerRange reports whether a location falls inside one
// of the class's static initializers — a static block or the value of a static
// field.
func isInClassStaticInitializerRange(classNode *ast.Node, location int) bool {
	if classNode == nil {
		return false
	}
	for _, member := range classNode.Members() {
		if member == nil {
			continue
		}
		switch member.Kind {
		case ast.KindClassStaticBlockDeclaration:
			if isInRange(member, location) {
				return true
			}
		case ast.KindPropertyDeclaration:
			if ast.HasStaticModifier(member) && isInRange(member.Initializer(), location) {
				return true
			}
		}
	}
	return false
}

// referenceContainsTypeQuery reports whether the identifier sits inside a
// `typeof X` type position, possibly behind a qualified name.
func referenceContainsTypeQuery(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindTypeQuery:
			return true
		case ast.KindIdentifier, ast.KindQualifiedName:
			continue
		default:
			return false
		}
	}
	return false
}

// isTypeReference reports whether the identifier names a type — every type
// position, plus `export = X`, which exports whichever space `X` lives in.
// eslint-scope carries the answer on the reference; here it is read back off
// the syntax, so that the option filters stay independent of how the scope
// model resolves a name.
func isTypeReference(node *ast.Node) bool {
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindExportAssignment {
		if assignment := parent.AsExportAssignment(); assignment != nil && assignment.IsExportEquals {
			return true
		}
	}
	return scope.InTypePosition(node)
}

func leftMostQualifiedName(node *ast.Node) *ast.Node {
	current := node
	for current.Kind == ast.KindQualifiedName {
		qualified := current.AsQualifiedName()
		if qualified == nil || qualified.Left == nil {
			break
		}
		current = qualified.Left
	}
	return current
}

// isExportSpecifierLocalName reports whether the identifier is the local half
// of an `export { local as exported }` specifier.
func isExportSpecifierLocalName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindExportSpecifier {
		return false
	}
	specifier := parent.AsExportSpecifier()
	if specifier == nil {
		return false
	}
	local := specifier.PropertyName
	if local == nil {
		local = specifier.Name()
	}
	return local == node
}

// isClassRefInClassDecorator reports whether a reference to a class binding
// appears in one of that class's own decorators. Decorators are applied after
// the class is defined, so the binding is already initialized.
func isClassRefInClassDecorator(declaration *scope.Variable, ref *scope.Reference) bool {
	if !isClassNameDef(declaration) || declaration.DefNode == nil {
		return false
	}
	for _, decorator := range declaration.DefNode.Decorators() {
		if decorator == nil {
			continue
		}
		if ref.Identifier.Pos() >= decorator.Pos() && ref.Identifier.End() <= decorator.End() {
			return true
		}
	}
	return false
}
