package switch_exhaustiveness_check

import (
	_ "embed"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed switch_exhaustiveness_check.schema.json
var schemaJSON []byte

type SwitchExhaustivenessCheckOptions struct {
	AllowDefaultCaseForExhaustiveSwitch bool
	RequireDefaultForNonUnion           bool
	ConsiderDefaultExhaustiveForUnions  bool
	DefaultCaseCommentPattern           *string
}

// SwitchExhaustivenessCheckRule implements the switch-exhaustiveness-check rule
// Require exhaustive switch statements
var SwitchExhaustivenessCheckRule = rule.CreateRule(rule.Rule{
	Name:             "switch-exhaustiveness-check",
	Schema:           rule.NewSchemaWithOptionsValidator(schemaJSON, validateSwitchExhaustivenessCheckOptions),
	RequiresTypeInfo: true,
	Run:              run,
})

func parseOptions(options []any) SwitchExhaustivenessCheckOptions {
	opts := SwitchExhaustivenessCheckOptions{
		AllowDefaultCaseForExhaustiveSwitch: true,
	}

	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if v, ok := optsMap["allowDefaultCaseForExhaustiveSwitch"].(bool); ok {
		opts.AllowDefaultCaseForExhaustiveSwitch = v
	}
	if v, ok := optsMap["requireDefaultForNonUnion"].(bool); ok {
		opts.RequireDefaultForNonUnion = v
	}
	if v, ok := optsMap["considerDefaultExhaustiveForUnions"].(bool); ok {
		opts.ConsiderDefaultExhaustiveForUnions = v
	}
	if v, ok := optsMap["defaultCaseCommentPattern"].(string); ok {
		opts.DefaultCaseCommentPattern = &v
	}
	return opts
}

// defaultCaseRef tracks a switch's "effective default": either a real
// `default:` clause, or (when none exists) a trailing comment after the last
// case that matches the configured defaultCaseCommentPattern.
type defaultCaseRef struct {
	clause  *ast.Node
	comment *ast.CommentRange
}

func (d defaultCaseRef) exists() bool {
	return d.clause != nil || d.comment != nil
}

// insertPos returns the position a "missing case" fix should be inserted
// before, when a default case/comment is present.
func (d defaultCaseRef) insertPos(sourceFile *ast.SourceFile) int {
	if d.clause != nil {
		return utils.TrimNodeTextRange(sourceFile, d.clause).Pos()
	}
	return d.comment.Pos()
}

func (d defaultCaseRef) reportRange(sourceFile *ast.SourceFile) core.TextRange {
	if d.clause != nil {
		return utils.TrimNodeTextRange(sourceFile, d.clause)
	}
	return core.NewTextRange(d.comment.Pos(), d.comment.End())
}

type switchMetadata struct {
	containsNonLiteralType    bool
	defaultCase               defaultCaseRef
	defaultCaseChecked        bool
	missingLiteralBranchTypes []*checker.Type
	discriminant              *ast.Node
}

// missingCase is either a concrete missing union constituent (typ != nil) or
// the sentinel "missing default clause" case (isDefault) used by
// checkSwitchNoUnionDefaultCase.
type missingCase struct {
	typ       *checker.Type
	typeText  string
	isDefault bool
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := parseOptions(options)
	commentPattern := newCommentPatternMatcher(opts.DefaultCaseCommentPattern)

	return rule.RuleListeners{
		ast.KindSwitchStatement: func(node *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}

			switchStmt := node.AsSwitchStatement()
			if switchStmt == nil || switchStmt.CaseBlock == nil {
				return
			}
			caseBlock := switchStmt.CaseBlock.AsCaseBlock()
			if caseBlock == nil || caseBlock.Clauses == nil {
				return
			}

			clauses := caseBlock.Clauses.Nodes
			meta := getSwitchMetadata(ctx, switchStmt, clauses, opts)
			if commentPattern.custom != nil || switchNeedsEffectiveDefaultCase(meta, opts) {
				resolveEffectiveDefaultCase(ctx, node, clauses, commentPattern, &meta)
			}

			checkSwitchExhaustive(ctx, node, switchStmt, caseBlock, meta, opts, commentPattern)
			checkSwitchUnnecessaryDefaultCase(ctx, meta, opts)
			checkSwitchNoUnionDefaultCase(ctx, node, switchStmt, caseBlock, meta, opts)
		},
	}
}

func getSwitchMetadata(
	ctx rule.RuleContext,
	switchStmt *ast.SwitchStatement,
	clauses []*ast.Node,
	opts SwitchExhaustivenessCheckOptions,
) switchMetadata {
	var defaultClause *ast.Node
	discriminant := ast.SkipParentheses(switchStmt.Expression)
	discriminantType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, discriminant)

	containsNonLiteralType := false
	if !opts.AllowDefaultCaseForExhaustiveSwitch || opts.RequireDefaultForNonUnion {
		containsNonLiteralType = doesTypeContainNonLiteralType(discriminantType)
	}

	caseTypes := map[*checker.Type]bool{}
	hasUndefinedCase := false
	for _, clause := range clauses {
		if clause.Kind == ast.KindDefaultClause {
			if defaultClause == nil {
				defaultClause = clause
			}
			continue
		}
		clauseData := clause.AsCaseOrDefaultClause()
		if clauseData == nil || clauseData.Expression == nil {
			continue
		}
		caseType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, ast.SkipParentheses(clauseData.Expression))
		caseTypes[caseType] = true
		// "missing", "optional" and "undefined" are distinct Type objects
		// with the same flag, so any such case covers that constituent.
		if utils.IsTypeFlagSet(caseType, checker.TypeFlagsUndefined) {
			hasUndefinedCase = true
		}
	}

	unionParts := utils.UnionTypeParts(discriminantType)
	if len(unionParts) > 1 && !slices.IsSortedFunc(unionParts, compareTypeIDs) {
		// TypeScript stores union constituents in Type.Id order, whereas tsgo
		// canonicalizes them by value. Sort only this rule-owned outer slice to
		// restore upstream order. Intersection constituents intentionally retain
		// their checker-provided insertion order, matching TypeScript's map.
		unionParts = slices.Clone(unionParts)
		slices.SortFunc(unionParts, compareTypeIDs)
	}
	var missing []*checker.Type
	for _, unionPart := range unionParts {
		for _, part := range utils.IntersectionTypeParts(unionPart) {
			if caseTypes[part] || !isTypeLiteralLikeType(part) {
				continue
			}
			if hasUndefinedCase && utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
				continue
			}
			missing = append(missing, part)
		}
	}

	return switchMetadata{
		containsNonLiteralType:    containsNonLiteralType,
		defaultCase:               defaultCaseRef{clause: defaultClause},
		defaultCaseChecked:        defaultClause != nil,
		missingLiteralBranchTypes: missing,
		discriminant:              discriminant,
	}
}

func switchNeedsEffectiveDefaultCase(meta switchMetadata, opts SwitchExhaustivenessCheckOptions) bool {
	return len(meta.missingLiteralBranchTypes) != 0 && opts.ConsiderDefaultExhaustiveForUnions ||
		len(meta.missingLiteralBranchTypes) == 0 && !meta.containsNonLiteralType && !opts.AllowDefaultCaseForExhaustiveSwitch ||
		meta.containsNonLiteralType && opts.RequireDefaultForNonUnion
}

func resolveEffectiveDefaultCase(
	ctx rule.RuleContext,
	node *ast.Node,
	clauses []*ast.Node,
	commentPattern *commentPatternMatcher,
	meta *switchMetadata,
) {
	if meta.defaultCaseChecked {
		return
	}
	meta.defaultCase.comment = getCommentDefaultCase(ctx, node, clauses, commentPattern)
	meta.defaultCaseChecked = true
}

func compareTypeIDs(a, b *checker.Type) int {
	switch {
	case a.Id() < b.Id():
		return -1
	case a.Id() > b.Id():
		return 1
	default:
		return 0
	}
}

// getCommentDefaultCase looks for a trailing comment after the last case
// (e.g. `// no default`) matching commentPattern, standing in for a real
// `default:` clause. Mirrors upstream's getCommentDefaultCase.
func getCommentDefaultCase(
	ctx rule.RuleContext,
	node *ast.Node,
	clauses []*ast.Node,
	commentPattern *commentPatternMatcher,
) *ast.CommentRange {
	if len(clauses) == 0 {
		return nil
	}
	lastClause := clauses[len(clauses)-1]
	lastComment := utils.LastCommentInSpan(ctx.Comments.All(), lastClause.End(), node.End())
	if lastComment == nil {
		return nil
	}

	text := ctx.SourceFile.Text()[lastComment.Pos():lastComment.End()]
	switch {
	case strings.HasPrefix(text, "//"):
		text = text[2:]
	case strings.HasPrefix(text, "/*") && strings.HasSuffix(text, "*/"):
		text = text[2 : len(text)-2]
	}
	text = strings.TrimFunc(text, utils.IsStrWhiteSpace)

	if commentPattern.matches(text) {
		return lastComment
	}
	return nil
}

func isTypeLiteralLikeType(t *checker.Type) bool {
	return utils.IsTypeFlagSet(
		t,
		checker.TypeFlagsLiteral|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsUniqueESSymbol,
	)
}

// doesTypeContainNonLiteralType reports whether any union constituent of t
// has no literal-like intersection part, e.g. `"foo" | number` (true) versus
// `"foo" | "bar"` (false). Default cases are never superfluous in switches
// with non-literal types.
func doesTypeContainNonLiteralType(t *checker.Type) bool {
	for _, unionPart := range utils.UnionTypeParts(t) {
		allNonLiteral := true
		for _, part := range utils.IntersectionTypeParts(unionPart) {
			if isTypeLiteralLikeType(part) {
				allNonLiteral = false
				break
			}
		}
		if allNonLiteral {
			return true
		}
	}
	return false
}

func typeToString(typeChecker *checker.Checker, t *checker.Type) string {
	return typeChecker.TypeToStringEx(
		t,
		nil,
		checker.TypeFormatFlagsAllowUniqueESSymbolType|
			checker.TypeFormatFlagsUseAliasDefinedOutsideCurrentScope|
			checker.TypeFormatFlagsUseFullyQualifiedType,
		nil,
	)
}

func symbolToRuntimeExpression(typeChecker *checker.Checker, symbol *ast.Symbol, enclosing *ast.Node, target core.ScriptTarget) string {
	if symbol == nil {
		return ""
	}
	// Unlike GetAccessibleSymbolChain, this checker query recursively verifies
	// qualified containers. That identity check rejects a formatted path whose
	// root is shadowed by an unrelated value in the switch scope.
	if !typeChecker.IsValueSymbolAccessible(symbol, enclosing) {
		return ""
	}
	formatted := typeChecker.SymbolToStringEx(
		symbol,
		enclosing,
		ast.SymbolFlagsValue,
		checker.SymbolFormatFlagsUseAliasDefinedOutsideCurrentScope|
			checker.SymbolFormatFlagsAllowAnyNodeKind,
	)
	for candidate := symbol; candidate != nil; candidate = candidate.Parent {
		chain := typeChecker.GetAccessibleSymbolChain(candidate, enclosing, ast.SymbolFlagsValue, false)
		if len(chain) == 0 {
			// Member symbols aren't themselves eligible roots in tsgo's
			// accessibility query. Their containing enum/namespace can still
			// provide the runtime path that SymbolToStringEx prints.
			continue
		}
		usable := true
		for _, chainSymbol := range chain {
			if !hasRuntimeValueMeaning(typeChecker, chainSymbol) {
				usable = false
				break
			}
		}
		if usable {
			return formatted
		}
	}

	// tsgo currently picks the first alias chain for a target symbol even when
	// that chain is type-only. If a later ordinary import reaches the same value,
	// enumerate the lexical value aliases and rebuild the qualified path from the
	// first matching runtime ancestor. This fallback is cold, deterministic, and
	// still rejects aliases that become type-only through a barrel.
	if fallback := runtimeExpressionThroughValueAlias(typeChecker, symbol, enclosing, target); fallback != "" {
		return fallback
	}

	// Type-literal properties deliberately have no accessible symbol chain,
	// even when the checker can print them through an in-scope value such as
	// `values.member`. Validate that printed expression's root in the switch
	// scope, including transitive type-only aliases, before accepting it.
	separator := strings.IndexAny(formatted, ".[")
	if separator <= 0 {
		return ""
	}
	rootName := formatted[:separator]
	root := typeChecker.ResolveName(rootName, enclosing, ast.SymbolFlagsValue, false)
	if !hasRuntimeValueMeaning(typeChecker, root) {
		return ""
	}
	if isNamespaceImportSymbol(root) {
		paths := runtimeExpressionsFromValueRoot(typeChecker, root, symbol, target)
		if len(paths) == 0 {
			return ""
		}
		return paths[0]
	}
	return formatted
}

func runtimeExpressionThroughValueAlias(
	typeChecker *checker.Checker,
	symbol *ast.Symbol,
	enclosing *ast.Node,
	target core.ScriptTarget,
) string {
	var candidates []string
	for _, scoped := range typeChecker.GetSymbolsInScope(enclosing, ast.SymbolFlagsValue|ast.SymbolFlagsAlias) {
		if scoped == nil || scoped.Flags&ast.SymbolFlagsAlias == 0 ||
			!hasRuntimeValueMeaning(typeChecker, scoped) {
			continue
		}
		candidates = append(candidates, runtimeExpressionsFromValueRoot(typeChecker, scoped, symbol, target)...)
	}
	if len(candidates) == 0 {
		return ""
	}
	slices.Sort(candidates)
	return slices.Compact(candidates)[0]
}

func sameRuntimeSymbol(typeChecker *checker.Checker, left, right *ast.Symbol) bool {
	if left == nil || right == nil {
		return false
	}
	leftCandidates := []*ast.Symbol{left, left.ExportSymbol, typeChecker.GetMergedSymbol(left)}
	rightCandidates := []*ast.Symbol{right, right.ExportSymbol, typeChecker.GetMergedSymbol(right)}
	for _, leftCandidate := range leftCandidates {
		if leftCandidate == nil {
			continue
		}
		for _, rightCandidate := range rightCandidates {
			if leftCandidate == rightCandidate {
				return true
			}
		}
	}
	return false
}

func hasRuntimeValueMeaning(typeChecker *checker.Checker, symbol *ast.Symbol) bool {
	seen := make(map[*ast.Symbol]struct{}, 4)
	for symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		if _, exists := seen[symbol]; exists {
			return false
		}
		seen[symbol] = struct{}{}
		// This checker API is intentionally non-transitive. Walk immediate
		// targets so an ordinary import through one or more type-only barrel
		// exports cannot be mistaken for a runtime binding.
		if typeChecker.GetTypeOnlyAliasDeclaration(symbol) != nil {
			return false
		}
		symbol = typeChecker.GetImmediateAliasedSymbol(symbol)
	}
	return symbol != nil && symbol.Flags&ast.SymbolFlagsValue != 0
}

func runtimeExpressionsFromValueRoot(
	typeChecker *checker.Checker,
	root *ast.Symbol,
	target *ast.Symbol,
	scriptTarget core.ScriptTarget,
) []string {
	rootName := declaredRuntimeSymbolName(root)
	if rootName == "" || strings.Contains(rootName, ast.InternalSymbolNamePrefix) {
		return nil
	}
	current := root
	if root.Flags&ast.SymbolFlagsAlias != 0 {
		current = typeChecker.GetAliasedSymbol(root)
	}
	if current == nil {
		return nil
	}
	if sameRuntimeSymbol(typeChecker, current, target) {
		return []string{rootName}
	}
	var candidates []string
	currentType := typeChecker.GetTypeOfSymbol(current)
	if currentType == nil {
		return candidates
	}
	ancestors := make([]*ast.Symbol, 0, 4)
	for symbol := target; symbol != nil; symbol = symbol.Parent {
		ancestors = append(ancestors, symbol)
	}

	// Try every suffix of the target's symbol ancestry. Resolving properties
	// from the root's runtime type expands value export-star barrels while
	// omitting type-only exports. Matching by symbol identity also preserves a
	// renamed export's actual property spelling.
	for highest := range ancestors {
		typ := currentType
		var expression strings.Builder
		expression.WriteString(rootName)
		valid := true
		for i := highest; i >= 0; i-- {
			member := runtimePropertyForSymbol(typeChecker, typ, ancestors[i])
			if member == nil {
				valid = false
				break
			}
			name := declaredRuntimeSymbolName(member)
			if name == "" || strings.Contains(name, ast.InternalSymbolNamePrefix) {
				valid = false
				break
			}
			if utils.RequiresQuoting(name, scriptTarget) {
				expression.WriteString("['")
				expression.WriteString(utils.EscapeJSSingleQuotedString(name))
				expression.WriteString("']")
			} else {
				expression.WriteByte('.')
				expression.WriteString(name)
			}
			if i != 0 {
				resolved := member
				if member.Flags&ast.SymbolFlagsAlias != 0 {
					resolved = typeChecker.GetAliasedSymbol(member)
				}
				typ = typeChecker.GetTypeOfSymbol(resolved)
				if typ == nil {
					valid = false
					break
				}
			}
		}
		if valid {
			candidates = append(candidates, expression.String())
		}
	}
	slices.Sort(candidates)
	return slices.Compact(candidates)
}

func runtimePropertyForSymbol(typeChecker *checker.Checker, typ *checker.Type, target *ast.Symbol) *ast.Symbol {
	properties := slices.Clone(typeChecker.GetPropertiesOfType(typ))
	slices.SortFunc(properties, func(a, b *ast.Symbol) int {
		return strings.Compare(declaredRuntimeSymbolName(a), declaredRuntimeSymbolName(b))
	})
	for _, property := range properties {
		// GetPropertiesOfType includes symbols contributed only by a type-only
		// export-star. Re-resolve the candidate as a runtime property; this query
		// consults the module's typeOnlyExportStarMap and filters them out.
		property = typeChecker.GetPropertyOfType(typ, property.Name)
		if !hasRuntimeValueMeaning(typeChecker, property) {
			continue
		}
		resolved := property
		if property.Flags&ast.SymbolFlagsAlias != 0 {
			resolved = typeChecker.GetAliasedSymbol(property)
		}
		if sameRuntimeSymbol(typeChecker, resolved, target) {
			return property
		}
	}
	return nil
}

func isNamespaceImportSymbol(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if ast.IsNamespaceImport(declaration) {
			return true
		}
	}
	return false
}

func declaredRuntimeSymbolName(symbol *ast.Symbol) string {
	if symbol == nil {
		return ""
	}
	if symbol.ValueDeclaration != nil {
		if name := runtimeDeclarationName(symbol.ValueDeclaration); name != "" {
			return name
		}
	}
	for _, declaration := range symbol.Declarations {
		if name := runtimeDeclarationName(declaration); name != "" {
			return name
		}
	}
	if strings.Contains(symbol.Name, ast.InternalSymbolNamePrefix) {
		return ""
	}
	return symbol.Name
}

func runtimeDeclarationName(declaration *ast.Node) string {
	if declaration == nil {
		return ""
	}
	name := declaration.Name()
	if name == nil {
		return ""
	}
	switch name.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return name.Text()
	default:
		return ""
	}
}

func checkSwitchExhaustive(
	ctx rule.RuleContext,
	node *ast.Node,
	switchStmt *ast.SwitchStatement,
	caseBlock *ast.CaseBlock,
	meta switchMetadata,
	opts SwitchExhaustivenessCheckOptions,
	commentPattern *commentPatternMatcher,
) {
	if len(meta.missingLiteralBranchTypes) == 0 {
		return
	}

	// If considerDefaultExhaustiveForUnions is enabled, the presence of a
	// default case always makes the switch exhaustive.
	if opts.ConsiderDefaultExhaustiveForUnions && meta.defaultCase.exists() {
		return
	}

	parts := make([]string, len(meta.missingLiteralBranchTypes))
	for i, t := range meta.missingLiteralBranchTypes {
		if utils.IsTypeFlagSet(t, checker.TypeFlagsESSymbolLike) {
			name := ""
			if sym := t.Symbol(); sym != nil {
				name = sym.Name
			}
			parts[i] = "typeof " + name
		} else {
			parts[i] = typeToString(ctx.TypeChecker, t)
		}
	}
	missingBranches := strings.Join(parts, " | ")

	missingTypes := meta.missingLiteralBranchTypes
	defaultCase := meta.defaultCase
	defaultCaseChecked := meta.defaultCaseChecked

	ctx.ReportNodeWithDeferredSuggestions(
		meta.discriminant,
		rule.RuleMessage{
			Id:          "switchIsNotExhaustive",
			Description: "Switch is not exhaustive. Cases not matched: " + missingBranches,
			Data:        map[string]string{"missingBranches": missingBranches},
		},
		func() []rule.RuleSuggestion {
			if !defaultCaseChecked {
				defaultCase.comment = getCommentDefaultCase(ctx, node, caseBlock.Clauses.Nodes, commentPattern)
			}
			cases := make([]missingCase, len(missingTypes))
			for i, t := range missingTypes {
				cases[i] = missingCase{typ: t, typeText: parts[i]}
			}
			fix, ok := fixSwitch(ctx, node, switchStmt, caseBlock, cases, defaultCase)
			if !ok {
				return nil
			}
			return []rule.RuleSuggestion{
				{
					Message: rule.RuleMessage{
						Id:          "addMissingCases",
						Description: "Add branches for missing cases.",
					},
					FixesArr: []rule.RuleFix{
						fix,
					},
				},
			}
		},
	)
}

func checkSwitchUnnecessaryDefaultCase(ctx rule.RuleContext, meta switchMetadata, opts SwitchExhaustivenessCheckOptions) {
	if opts.AllowDefaultCaseForExhaustiveSwitch {
		return
	}

	if len(meta.missingLiteralBranchTypes) != 0 || meta.containsNonLiteralType {
		return
	}
	defaultCase := meta.defaultCase
	if defaultCase.exists() {
		ctx.ReportRange(defaultCase.reportRange(ctx.SourceFile), rule.RuleMessage{
			Id:          "dangerousDefaultCase",
			Description: "The switch statement is exhaustive, so the default case is unnecessary.",
		})
	}
}

func checkSwitchNoUnionDefaultCase(
	ctx rule.RuleContext,
	node *ast.Node,
	switchStmt *ast.SwitchStatement,
	caseBlock *ast.CaseBlock,
	meta switchMetadata,
	opts SwitchExhaustivenessCheckOptions,
) {
	if !opts.RequireDefaultForNonUnion {
		return
	}

	if !meta.containsNonLiteralType || meta.defaultCase.exists() {
		return
	}

	defaultCase := meta.defaultCase

	ctx.ReportNodeWithDeferredSuggestions(
		meta.discriminant,
		rule.RuleMessage{
			Id:          "switchIsNotExhaustive",
			Description: "Switch is not exhaustive. Cases not matched: default",
			Data:        map[string]string{"missingBranches": "default"},
		},
		func() []rule.RuleSuggestion {
			fix, _ := fixSwitch(ctx, node, switchStmt, caseBlock, []missingCase{{isDefault: true}}, defaultCase)
			return []rule.RuleSuggestion{
				{
					Message: rule.RuleMessage{
						Id:          "addMissingCases",
						Description: "Add branches for missing cases.",
					},
					FixesArr: []rule.RuleFix{
						fix,
					},
				},
			}
		},
	)
}

// fixSwitch builds the single autofix-suggestion edit that appends the
// missing branches (and/or a `default:` sentinel) to the switch. Mirrors
// upstream's fixSwitch.
func fixSwitch(
	ctx rule.RuleContext,
	node *ast.Node,
	switchStmt *ast.SwitchStatement,
	caseBlock *ast.CaseBlock,
	missingBranchTypes []missingCase,
	defaultCase defaultCaseRef,
) (rule.RuleFix, bool) {
	clauses := caseBlock.Clauses.Nodes
	var lastCase *ast.Node
	if len(clauses) > 0 {
		lastCase = clauses[len(clauses)-1]
	}

	// If there are no existing cases, use the indentation of the switch
	// statement and leave it to the user to format it correctly.
	indentAnchor := node
	if lastCase != nil {
		indentAnchor = lastCase
	}
	_, col := scanner.GetECMALineAndUTF16CharacterOfPosition(
		ctx.SourceFile,
		utils.TrimNodeTextRange(ctx.SourceFile, indentAnchor).Pos(),
	)
	caseIndent := strings.Repeat(" ", int(col))
	target := core.ScriptTargetNone
	if sourceProgram := ctx.Program(); sourceProgram != nil && sourceProgram.Options() != nil {
		target = sourceProgram.Options().Target
	}
	enclosing := ast.SkipParentheses(switchStmt.Expression)
	var cachedContainerSymbol *ast.Symbol
	cachedContainerName := ""
	resolveContainer := func(symbol *ast.Symbol) string {
		if symbol != cachedContainerSymbol {
			cachedContainerSymbol = symbol
			cachedContainerName = symbolToRuntimeExpression(ctx.TypeChecker, symbol, enclosing, target)
		}
		return cachedContainerName
	}

	missingCasesText := make([]string, len(missingBranchTypes))
	for i, mb := range missingBranchTypes {
		if mb.isDefault {
			missingCasesText[i] = `default: { throw new Error('default case') }`
			continue
		}

		missingBranchType := mb.typ
		missingBranchName := ""
		hasMissingBranchName := false
		missingBranchSymbol := missingBranchType.Symbol()
		if missingBranchSymbol != nil {
			missingBranchName = missingBranchSymbol.Name
			hasMissingBranchName = true
		}

		caseTest := mb.typeText
		if utils.IsTypeFlagSet(missingBranchType, checker.TypeFlagsESSymbolLike) {
			caseTest = symbolToRuntimeExpression(ctx.TypeChecker, missingBranchSymbol, enclosing, target)
			if caseTest == "" {
				return rule.RuleFix{}, false
			}
		} else if hasMissingBranchName && utils.RequiresQuoting(missingBranchName, target) {
			containerName := resolveContainer(missingBranchSymbol.Parent)
			if containerName == "" {
				return rule.RuleFix{}, false
			}
			escapedBranchName := utils.EscapeJSSingleQuotedString(missingBranchName)
			caseTest = containerName + "['" + escapedBranchName + "']"
		} else if hasMissingBranchName {
			containerName := resolveContainer(missingBranchSymbol.Parent)
			if containerName == "" {
				return rule.RuleFix{}, false
			}
			caseTest = containerName + "." + missingBranchName
		}

		escapedForMessage := strings.ReplaceAll(caseTest, `\`, `\\`)
		escapedForMessage = strings.ReplaceAll(escapedForMessage, `'`, `\'`)

		missingCasesText[i] = "case " + caseTest + ": { throw new Error('Not implemented yet: " + escapedForMessage + " case') }"
	}

	if lastCase != nil {
		if defaultCase.exists() {
			var beforeFixString strings.Builder
			for _, text := range missingCasesText {
				beforeFixString.WriteString(text)
				beforeFixString.WriteString("\n")
				beforeFixString.WriteString(caseIndent)
			}
			pos := defaultCase.insertPos(ctx.SourceFile)
			return rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), beforeFixString.String()), true
		}
		fixString := caseIndent + strings.Join(missingCasesText, "\n"+caseIndent)
		return rule.RuleFixInsertAfter(lastCase, "\n"+fixString), true
	}

	// There were no existing cases: replace the whole `{ ... }` block.
	fixString := caseIndent + strings.Join(missingCasesText, "\n"+caseIndent)
	trimmed := utils.TrimNodeTextRange(ctx.SourceFile, switchStmt.CaseBlock)
	newText := "{\n" + fixString + "\n" + caseIndent + "}"
	return rule.RuleFixReplaceRange(trimmed, newText), true
}
