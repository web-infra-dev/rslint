package typeservice

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/program"
)

type TypeDetails struct {
	pattern, node, rootNode           *ast.Node
	labels                            []*ast.Node
	patternRef, nodeRef, rootRef      *NodeRef
	labelRefs                         []*NodeRef
	fresh, regular, base, constraint  uint32
	outer, local, elements            Handles
	minPlusOne, fixedPlusOne          uint32
	texts                             []string
	rootCheck, rootExtends, rootAlias uint32
	distributive                      bool
	infer, rootOuter, rootArguments   Handles
}

func (*TypeDetails) result() {}

func (s *Session) typeDetails(c *checker.Checker, t *checker.Type) *TypeDetails {
	d := &TypeDetails{pattern: checker.Checker_patternForType(c)[t]}
	flags := t.Flags()
	if t.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		if ref := t.AsTypeReference(); ref != nil {
			d.node = checker.TypeReference_node(ref)
		}
	}
	if def := t.AsInterfaceType(); def != nil {
		d.outer = s.typeIDs(def.OuterTypeParameters())
		d.local = s.typeIDs(def.LocalTypeParameters())
		if t.ObjectFlags()&checker.ObjectFlagsTuple != 0 {
			tuple := t.AsTupleType()
			d.fixedPlusOne = uint32(tuple.FixedLength() + 1)
			d.minPlusOne = uint32(checker.TupleType_minLength(tuple) + 1)
			d.elements = make(Handles, len(tuple.ElementInfos()))
			d.labels = make([]*ast.Node, len(d.elements))
			for i, el := range tuple.ElementInfos() {
				d.elements[i] = uint32(el.TupleElementFlags())
				d.labels[i] = el.LabeledDeclaration()
			}
		}
	}
	switch {
	case flags&(checker.TypeFlagsStringLiteral|checker.TypeFlagsBooleanLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral|checker.TypeFlagsEnumLiteral) != 0 && flags&checker.TypeFlagsUnion == 0:
		literal := t.AsLiteralType()
		d.fresh, d.regular = s.typeID(literal.FreshType()), s.typeID(literal.RegularType())
	case flags&checker.TypeFlagsTemplateLiteral != 0:
		d.texts = t.AsTemplateLiteralType().Texts()
	case flags&checker.TypeFlagsSubstitution != 0:
		sub := t.AsSubstitutionType()
		d.base, d.constraint = s.typeID(sub.BaseType()), s.typeID(sub.SubstConstraint())
	case flags&checker.TypeFlagsConditional != 0:
		root := checker.ConditionalType_root(t.AsConditionalType())
		d.rootNode = checker.ConditionalRoot_node(root).AsNode()
		d.rootCheck = s.typeID(checker.ConditionalRoot_checkType(root))
		d.rootExtends = s.typeID(checker.ConditionalRoot_extendsType(root))
		d.infer = s.typeIDs(checker.ConditionalRoot_inferTypeParameters(root))
		d.rootOuter = s.typeIDs(checker.ConditionalRoot_outerTypeParameters(root))
		d.distributive = checker.ConditionalRoot_isDistributive(root)
		alias := checker.ConditionalRoot_alias(root)
		d.rootAlias, d.rootArguments = s.symbolID(alias.Symbol()), s.typeIDs(alias.TypeArguments())
	}
	return d
}
func nodeRefs(p *program.Program, nodes []*ast.Node) []*NodeRef {
	if nodes == nil {
		return nil
	}
	refs := make([]*NodeRef, len(nodes))
	for i, n := range nodes {
		refs[i] = nodeReference(p, n)
	}
	return refs
}
func globalExports(symbol *ast.Symbol) ast.SymbolTable {
	for _, declaration := range symbol.Declarations {
		if ast.IsSourceFile(declaration) {
			return declaration.AsSourceFile().GlobalExports
		}
	}
	return nil
}
