package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// RefNeeds is a set of complete-file collections a rule may request from the
// per-file RefStore. Direct node queries such as Resolve and ReferenceFacts do
// not require a collection need.
type RefNeeds uint8

const (
	RefNeedReferences RefNeeds = 1 << iota
	RefNeedBindingDeclarations
	RefNeedImportBindings

	RefNeedsAll = RefNeedReferences | RefNeedBindingDeclarations | RefNeedImportBindings
)

// IsValid reports whether needs contains only supported RefStore features.
func (needs RefNeeds) IsValid() bool {
	return needs&^RefNeedsAll == 0
}

// RuleNeeds declares the shared-analysis capabilities a rule may dynamically
// request for a file. It is deliberately an extensible wrapper rather than a
// second analysis store.
type RuleNeeds struct {
	Refs RefNeeds
}

// ReferenceAccess describes the lexical access performed by a reference.
// Type-only references are reads in this lexical sense even though they do not
// read a runtime value.
type ReferenceAccess uint8

const (
	ReferenceAccessRead ReferenceAccess = 1 << iota
	ReferenceAccessWrite
)

// ReferenceSpace identifies the declaration spaces in which a reference may
// resolve. Flags may be combined for constructs such as a regular local export.
type ReferenceSpace uint8

const (
	ReferenceSpaceValue ReferenceSpace = 1 << iota
	ReferenceSpaceType
	ReferenceSpaceNamespace
)

// ReferenceSyntax records syntax distinctions that commonly affect rule
// policy without baking those policies (used, captured, shadowed, and so on)
// into RefStore.
type ReferenceSyntax uint16

const (
	ReferenceSyntaxTypePosition ReferenceSyntax = 1 << iota
	ReferenceSyntaxTypeQuery
	ReferenceSyntaxLocalExport
	ReferenceSyntaxShorthand
	ReferenceSyntaxJSXTag
)

// ReferenceFacts is a syntax-only description of one explicit lexical
// reference. It never performs checker resolution and contains no rule-level
// conclusion such as whether the reference makes a binding "used". The two
// Syntactic*Container fields are direct AST-helper results, not an ESLint scope
// or an execution-context claim. NearestFunction may therefore be a method
// whose computed name contains Node; OutsideFunctionBody records that exact
// syntax relationship without calling the reference a parameter/signature use.
type ReferenceFacts struct {
	Node                    *ast.Node
	Access                  ReferenceAccess
	Space                   ReferenceSpace
	Syntax                  ReferenceSyntax
	SyntacticBlockContainer *ast.Node
	SyntacticScopeContainer *ast.Node
	NearestFunction         *ast.Node
	OutsideFunctionBody     bool
}

// BindingDeclarationFacts describes one authored binding occurrence. Symbol is
// the lexical/type binder identity used by RefStore; declaration merging may
// therefore produce multiple facts that share a Symbol. A TypeScript parameter
// property has a second class-member identity in MemberSymbol; all other
// declarations leave MemberSymbol nil. The two Syntactic*Container fields are
// exact AST-helper results, not a claim that RefStore has constructed a
// complete ESLint scope or Variable object.
// NearestFunction includes a function declaration/expression that owns Name;
// OutsideFunctionBody says Name is outside that function's Body. Neither field
// identifies the ESLint scope that owns the binding.
type BindingDeclarationFacts struct {
	Name                    *ast.Node
	Declaration             *ast.Node
	RootDeclaration         *ast.Node
	Symbol                  *ast.Symbol
	MemberSymbol            *ast.Symbol
	SyntacticBlockContainer *ast.Node
	SyntacticScopeContainer *ast.Node
	NearestFunction         *ast.Node
	OutsideFunctionBody     bool
}

// ImportBindingKind identifies the syntax that declares a local import alias.
type ImportBindingKind uint8

const (
	ImportBindingDefault ImportBindingKind = iota + 1
	ImportBindingNamed
	ImportBindingNamespace
	ImportBindingEquals
)

// ImportBinding is syntax-only provenance for a local import binding. It never
// resolves the module, follows the alias target, or consults ModuleGraph.
type ImportBinding struct {
	Kind            ImportBindingKind
	Declaration     *ast.Node
	Specifier       *ast.Node
	LocalName       *ast.Node
	Symbol          *ast.Symbol
	ModuleSpecifier *ast.Node
	ImportedName    string
	TypeOnly        bool
}

// ReferenceFacts returns neutral syntax and container facts for an explicit
// lexical reference. node must belong to this store's source file. A
// syntactically valid but unresolved identifier still returns ok=true.
// Declaration names, property labels, re-export names, intrinsic JSX tags,
// and other non-reference positions return ok=false.
func (s *RefStore) ReferenceFacts(node *ast.Node) (facts ReferenceFacts, ok bool) {
	if s == nil || node == nil || node.Kind != ast.KindIdentifier || !isReferencePosition(node) {
		return ReferenceFacts{}, false
	}

	facts.Node = node
	facts.Access = ReferenceAccessRead
	if isWriteReference(node) {
		facts.Access = ReferenceAccessWrite
		if isReadWriteReference(node) {
			facts.Access |= ReferenceAccessRead
		}
	}

	facts.Space = referenceSpace(node)

	typeOnlyExport := node.Parent != nil && node.Parent.Kind == ast.KindExportSpecifier &&
		ast.IsTypeOnlyImportOrExportDeclaration(node.Parent)
	if ast.IsPartOfTypeNode(node) || ast.IsPartOfTypeQuery(node) ||
		isTypeOnlyPropertyAccessQualifier(node) || typeOnlyExport {
		facts.Syntax |= ReferenceSyntaxTypePosition
	}
	if ast.IsPartOfTypeQuery(node) {
		facts.Syntax |= ReferenceSyntaxTypeQuery
	}
	if parent := node.Parent; parent != nil {
		switch parent.Kind {
		case ast.KindExportSpecifier:
			if !utils.IsReExportSpecifier(parent) {
				facts.Syntax |= ReferenceSyntaxLocalExport
			}
		case ast.KindShorthandPropertyAssignment:
			facts.Syntax |= ReferenceSyntaxShorthand
		}
	}
	if isJSXTagReference(node) {
		facts.Syntax |= ReferenceSyntaxJSXTag
	}

	facts.SyntacticBlockContainer = ast.GetEnclosingBlockScopeContainer(node)
	facts.SyntacticScopeContainer = utils.FindEnclosingScope(node)
	facts.NearestFunction = ast.FindAncestor(node.Parent, ast.IsFunctionLike)
	if facts.NearestFunction != nil {
		facts.OutsideFunctionBody = isOutsideFunctionBody(node, facts.NearestFunction)
	}
	return facts, true
}

// referenceSpace describes the syntactic lookup spaces requested by a
// reference. It intentionally does not derive these flags by intersecting
// referenceMeaning with SymbolFlagsValue/Type/Namespace: TypeScript's symbol
// flag groups overlap (for example, class flags belong to both Value and Type)
// and every resolver meaning also carries SymbolFlagsAlias.
func referenceSpace(node *ast.Node) ReferenceSpace {
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindTypePredicate {
		return ReferenceSpaceValue
	}
	if parent != nil && parent.Kind == ast.KindExportAssignment {
		return ReferenceSpaceValue | ReferenceSpaceType | ReferenceSpaceNamespace
	}
	if parent != nil && parent.Kind == ast.KindExportSpecifier {
		if ast.IsTypeOnlyImportOrExportDeclaration(parent) {
			return ReferenceSpaceType | ReferenceSpaceNamespace
		}
		return ReferenceSpaceValue | ReferenceSpaceType | ReferenceSpaceNamespace
	}

	entity := node
	for entity.Parent != nil && entity.Parent.Kind == ast.KindQualifiedName {
		entity = entity.Parent
	}
	if entity.Parent != nil && entity.Parent.Kind == ast.KindImportEqualsDeclaration &&
		entity.Parent.AsImportEqualsDeclaration().ModuleReference == entity {
		return ReferenceSpaceValue | ReferenceSpaceType | ReferenceSpaceNamespace
	}
	if isTypeOnlyPropertyAccessQualifier(node) {
		return ReferenceSpaceNamespace
	}
	if ast.IsExpressionNode(node) {
		return ReferenceSpaceValue
	}
	if parent != nil && parent.Kind == ast.KindQualifiedName && parent.AsQualifiedName().Left == node {
		return ReferenceSpaceNamespace
	}
	if ast.IsPartOfTypeNode(node) {
		return ReferenceSpaceType
	}
	return ReferenceSpaceValue
}

// BindingDeclaration returns facts for one authored declaration identifier.
// name must belong to this store's source file. It is a direct query and never
// builds the complete-file declaration list.
func (s *RefStore) BindingDeclaration(name *ast.Node) (BindingDeclarationFacts, bool) {
	if s == nil {
		return BindingDeclarationFacts{}, false
	}
	return bindingDeclarationFacts(name)
}

// BindingDeclarations returns every authored lexical, type, import, and enum
// declaration identifier in source order. Class/object members are properties,
// not lexical bindings; `export as namespace` and the `global` token of a
// `declare global` block are augmentation syntax, so none is included. A
// parameter property contributes its parameter binding and exposes its second
// class-member identity on that fact. Treat the returned slice as read-only.
func (s *RefStore) BindingDeclarations() []BindingDeclarationFacts {
	if s == nil {
		return nil
	}
	s.ensureBuilt(RefNeedBindingDeclarations)
	if s.facts == nil {
		return nil
	}
	facts := s.facts
	if facts.materializedNeeds&RefNeedBindingDeclarations == 0 {
		if facts.materialized == nil {
			facts.materialized = &materializedRefFacts{}
		}
		declarations := make([]BindingDeclarationFacts, 0, len(facts.declarationNames))
		for _, name := range facts.declarationNames {
			if declaration, ok := bindingDeclarationFacts(name); ok {
				declarations = append(declarations, declaration)
			}
		}
		facts.materialized.declarations = declarations
		facts.declarationNames = nil
		facts.materializedNeeds |= RefNeedBindingDeclarations
	}
	return facts.materialized.declarations
}

// ImportBinding returns syntax-only import provenance for a local import
// declaration identifier or for an explicit reference to that binding. node
// must belong to this store's source file. Reference lookup is binder-only;
// this method never resolves a module target.
func (s *RefStore) ImportBinding(node *ast.Node) (ImportBinding, bool) {
	if s == nil || node == nil || node.Kind != ast.KindIdentifier {
		return ImportBinding{}, false
	}
	if binding, ok := importBindingFacts(node); ok {
		return binding, true
	}
	if !isReferencePosition(node) {
		return ImportBinding{}, false
	}
	sym := s.ResolveInFile(node)
	if sym == nil {
		return ImportBinding{}, false
	}
	for _, declaration := range sym.Declarations {
		if name := declaration.Name(); name != nil {
			if binding, ok := importBindingFacts(name); ok && binding.Symbol == sym {
				return binding, true
			}
		}
	}
	return ImportBinding{}, false
}

// ImportBindings returns every local binding from ECMAScript/TypeScript import
// declarations in source order, including the re-parsed form of JSDoc
// `@import`. Treat the returned slice as read-only. Module resolution is never
// performed.
func (s *RefStore) ImportBindings() []ImportBinding {
	if s == nil {
		return nil
	}
	s.ensureBuilt(RefNeedImportBindings)
	if s.facts == nil {
		return nil
	}
	facts := s.facts
	if facts.materializedNeeds&RefNeedImportBindings == 0 {
		if facts.materialized == nil {
			facts.materialized = &materializedRefFacts{}
		}
		imports := make([]ImportBinding, 0, len(facts.importNames))
		for _, name := range facts.importNames {
			if binding, ok := importBindingFacts(name); ok {
				imports = append(imports, binding)
			}
		}
		facts.materialized.imports = imports
		facts.importNames = nil
		facts.materializedNeeds |= RefNeedImportBindings
	}
	return facts.materialized.imports
}

func bindingDeclarationFacts(name *ast.Node) (BindingDeclarationFacts, bool) {
	if name == nil || name.Kind != ast.KindIdentifier || !isBindingDeclarationIdentifier(name) {
		return BindingDeclarationFacts{}, false
	}
	declaration := name.Parent
	root := ast.GetRootDeclaration(declaration)
	if root == nil {
		root = declaration
	}
	function := ast.FindAncestor(name.Parent, ast.IsFunctionLike)
	symbol := utils.BindingNameSymbol(name)
	var memberSymbol *ast.Symbol
	if declaration.Kind == ast.KindParameter && declaration.Parent != nil &&
		ast.IsParameterPropertyDeclaration(declaration, declaration.Parent) {
		memberSymbol = symbol
		symbol = declaration.Parent.Locals()[name.Text()]
	}
	return BindingDeclarationFacts{
		Name:                    name,
		Declaration:             declaration,
		RootDeclaration:         root,
		Symbol:                  symbol,
		MemberSymbol:            memberSymbol,
		SyntacticBlockContainer: ast.GetEnclosingBlockScopeContainer(declaration),
		SyntacticScopeContainer: utils.FindEnclosingScope(declaration),
		NearestFunction:         function,
		OutsideFunctionBody:     function != nil && isOutsideFunctionBody(name, function),
	}, true
}

// isBindingDeclarationIdentifier is the exact declaration taxonomy exposed by
// BindingDeclarations. The shared utility covers regular JavaScript and
// TypeScript declarations; JSDoc typedef/callback syntax is re-parsed as a
// JSTypeAliasDeclaration and needs the explicit final case. Property/member
// names and NamespaceExportDeclaration aliases deliberately remain excluded.
func isBindingDeclarationIdentifier(node *ast.Node) bool {
	if node != nil && node.Parent != nil && node.Parent.Kind == ast.KindModuleDeclaration &&
		ast.IsGlobalScopeAugmentation(node.Parent) {
		return false
	}
	if node != nil && node.Parent != nil && node.Parent.Kind == ast.KindFunctionDeclaration &&
		node.Parent.Flags&ast.NodeFlagsReparsed != 0 && node.Parent.Name() == node {
		// A JSDoc @overload signature clones the implementation's name into a
		// synthetic FunctionDeclaration even though the tag did not author a
		// second name occurrence.
		return false
	}
	if utils.IsDeclarationIdentifier(node) {
		return true
	}
	return node != nil && node.Parent != nil && node.Parent.Kind == ast.KindJSTypeAliasDeclaration &&
		node.Parent.Name() == node
}

func importBindingFacts(name *ast.Node) (ImportBinding, bool) {
	if name == nil || name.Kind != ast.KindIdentifier || name.Parent == nil || name.Parent.Name() != name {
		return ImportBinding{}, false
	}

	specifier := name.Parent
	binding := ImportBinding{
		Specifier: specifier,
		LocalName: name,
		Symbol:    utils.BindingNameSymbol(name),
	}

	switch specifier.Kind {
	case ast.KindImportClause:
		binding.Kind = ImportBindingDefault
		binding.ImportedName = "default"
	case ast.KindImportSpecifier:
		binding.Kind = ImportBindingNamed
		importSpecifier := specifier.AsImportSpecifier()
		imported := importSpecifier.PropertyName
		if imported == nil {
			imported = importSpecifier.Name()
		}
		if imported != nil {
			binding.ImportedName = imported.Text()
		}
	case ast.KindNamespaceImport:
		binding.Kind = ImportBindingNamespace
		binding.ImportedName = "*"
	case ast.KindImportEqualsDeclaration:
		binding.Kind = ImportBindingEquals
		declaration := specifier.AsImportEqualsDeclaration()
		binding.Declaration = specifier
		binding.TypeOnly = declaration.IsTypeOnly
		if moduleReference := declaration.ModuleReference; moduleReference != nil && moduleReference.Kind == ast.KindExternalModuleReference {
			binding.ModuleSpecifier = moduleReference.AsExternalModuleReference().Expression
		}
		return binding, true
	default:
		return ImportBinding{}, false
	}

	binding.TypeOnly = ast.IsTypeOnlyImportOrExportDeclaration(specifier)
	for current := specifier.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindImportDeclaration || current.Kind == ast.KindJSImportDeclaration {
			binding.Declaration = current
			binding.ModuleSpecifier = current.AsImportDeclaration().ModuleSpecifier
			return binding, true
		}
		if current.Kind == ast.KindSourceFile {
			break
		}
	}
	return ImportBinding{}, false
}

func isReadWriteReference(node *ast.Node) bool {
	current := transparentReferenceTarget(node)
	parent := current.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.Left == current && binary.OperatorToken.Kind != ast.KindEqualsToken &&
			ast.IsAssignmentOperator(binary.OperatorToken.Kind)
	case ast.KindPrefixUnaryExpression:
		unary := parent.AsPrefixUnaryExpression()
		return unary.Operand == current &&
			(unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken)
	case ast.KindPostfixUnaryExpression:
		unary := parent.AsPostfixUnaryExpression()
		return unary.Operand == current &&
			(unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken)
	}
	return false
}

func isWriteReference(node *ast.Node) bool {
	return utils.IsWriteReference(transparentReferenceTarget(node))
}

// transparentReferenceTarget returns the outer expression that remains the
// same assignment target as node. The shared IsWriteReference already handles
// most wrappers recursively; normalizing here also covers SatisfiesExpression,
// which TypeScript accepts and erases around assignment targets.
func transparentReferenceTarget(node *ast.Node) *ast.Node {
	current := node
	for current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression,
			ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
			if parent.Expression() != current {
				return current
			}
			current = parent
		default:
			return current
		}
	}
	return current
}

func isOutsideFunctionBody(node *ast.Node, function *ast.Node) bool {
	body := function.Body()
	return body == nil || !ast.IsNodeDescendantOf(node, body)
}

func isJSXTagReference(node *ast.Node) bool {
	entity := node
	for entity.Parent != nil && entity.Parent.Kind == ast.KindPropertyAccessExpression &&
		entity.Parent.AsPropertyAccessExpression().Expression == entity {
		entity = entity.Parent
	}
	return ast.IsJsxTagName(entity)
}
