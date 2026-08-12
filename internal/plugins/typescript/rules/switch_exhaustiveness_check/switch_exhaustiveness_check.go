package switch_exhaustiveness_check

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/dlclark/regexp2"
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
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run:              run,
})

// defaultCommentPattern mirrors upstream's DEFAULT_COMMENT_PATTERN
// (/^no default$/iu). A configured defaultCaseCommentPattern is compiled
// without the `i` flag, exactly like upstream's `new RegExp(pattern, "u")`.
var defaultCommentPattern = regexp2.MustCompile(`^no default$`, utils.JSUnicodeRegexOptions|regexp2.IgnoreCase)

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
	missingLiteralBranchTypes []*checker.Type
	symbolName                string
	discriminant              *ast.Node
}

// missingCase is either a concrete missing union constituent (typ != nil) or
// the sentinel "missing default clause" case (isDefault) used by
// checkSwitchNoUnionDefaultCase.
type missingCase struct {
	typ       *checker.Type
	isDefault bool
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := parseOptions(options)

	commentPattern := defaultCommentPattern
	if opts.DefaultCaseCommentPattern != nil {
		// An invalid pattern is rejected by the schema's `format: "regex"`
		// before linting starts, so this compile error is only defensive.
		if compiled, err := utils.CompileRegexp2(*opts.DefaultCaseCommentPattern, utils.JSUnicodeRegexOptions); err == nil {
			commentPattern = compiled
		}
	}

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

			meta := getSwitchMetadata(ctx, node, switchStmt, caseBlock.Clauses.Nodes, commentPattern)

			checkSwitchExhaustive(ctx, node, switchStmt, caseBlock, meta, opts)
			checkSwitchUnnecessaryDefaultCase(ctx, meta, opts)
			checkSwitchNoUnionDefaultCase(ctx, node, switchStmt, caseBlock, meta, opts)
		},
	}
}

func getSwitchMetadata(
	ctx rule.RuleContext,
	node *ast.Node,
	switchStmt *ast.SwitchStatement,
	clauses []*ast.Node,
	commentPattern *regexp2.Regexp,
) switchMetadata {
	var defaultClause *ast.Node
	for _, clause := range clauses {
		if clause.Kind == ast.KindDefaultClause {
			defaultClause = clause
			break
		}
	}

	discriminant := ast.SkipParentheses(switchStmt.Expression)
	discriminantType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, discriminant)

	symbolName := ""
	if sym := discriminantType.Symbol(); sym != nil {
		symbolName = sym.Name
	}

	containsNonLiteralType := doesTypeContainNonLiteralType(discriminantType)

	caseTypes := map[*checker.Type]bool{}
	for _, clause := range clauses {
		clauseData := clause.AsCaseOrDefaultClause()
		if clauseData == nil || clauseData.Expression == nil {
			// A `default:` clause has no test expression.
			continue
		}
		caseTypes[utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, ast.SkipParentheses(clauseData.Expression))] = true
	}

	// "missing", "optional" and "undefined" types are different runtime
	// objects, but all of them have TypeFlagsUndefined, so a case already
	// covering any of them is treated as covering the constituent below.
	hasUndefinedCase := false
	for t := range caseTypes {
		if utils.IsTypeFlagSet(t, checker.TypeFlagsUndefined) {
			hasUndefinedCase = true
			break
		}
	}

	var missing []*checker.Type
	for _, unionPart := range utils.UnionTypeParts(discriminantType) {
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

	defaultCase := defaultCaseRef{clause: defaultClause}
	if defaultClause == nil {
		defaultCase.comment = getCommentDefaultCase(ctx, node, clauses, commentPattern)
	}

	return switchMetadata{
		containsNonLiteralType:    containsNonLiteralType,
		defaultCase:               defaultCase,
		missingLiteralBranchTypes: missing,
		symbolName:                symbolName,
		discriminant:              discriminant,
	}
}

// getCommentDefaultCase looks for a trailing comment after the last case
// (e.g. `// no default`) matching commentPattern, standing in for a real
// `default:` clause. Mirrors upstream's getCommentDefaultCase.
func getCommentDefaultCase(
	ctx rule.RuleContext,
	node *ast.Node,
	clauses []*ast.Node,
	commentPattern *regexp2.Regexp,
) *ast.CommentRange {
	if len(clauses) == 0 {
		return nil
	}
	lastClause := clauses[len(clauses)-1]

	var lastComment *ast.CommentRange
	for c := range utils.GetCommentsInRange(ctx.SourceFile, core.NewTextRange(lastClause.End(), node.End())) {
		comment := c
		lastComment = &comment
	}
	if lastComment == nil {
		return nil
	}

	text := strings.TrimSpace(ctx.SourceFile.Text()[lastComment.Pos():lastComment.End()])
	switch {
	case strings.HasPrefix(text, "//"):
		text = strings.TrimSpace(text[2:])
	case strings.HasPrefix(text, "/*") && strings.HasSuffix(text, "*/"):
		text = strings.TrimSpace(text[2 : len(text)-2])
	}

	if utils.Regexp2MatchString(commentPattern, text) {
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

// requiresQuoting reports whether name needs bracket-quoting
// (`Enum['a-b']`) rather than dot access (`Enum.ab`) when read back as a
// property name. Like upstream, the name is inspected one UTF-16 code unit at
// a time, so a character outside the BMP is seen as a surrogate pair and
// requires quoting even though the code point itself is a letter.
func requiresQuoting(name string) bool {
	if name == "" {
		return true
	}
	for i, ch := range name {
		if ch > 0xFFFF {
			return true
		}
		if i == 0 {
			if !scanner.IsIdentifierStart(ch) {
				return true
			}
		} else if !scanner.IsIdentifierPart(ch) {
			return true
		}
	}
	return false
}

func checkSwitchExhaustive(
	ctx rule.RuleContext,
	node *ast.Node,
	switchStmt *ast.SwitchStatement,
	caseBlock *ast.CaseBlock,
	meta switchMetadata,
	opts SwitchExhaustivenessCheckOptions,
) {
	// If considerDefaultExhaustiveForUnions is enabled, the presence of a
	// default case always makes the switch exhaustive.
	if opts.ConsiderDefaultExhaustiveForUnions && meta.defaultCase.exists() {
		return
	}

	if len(meta.missingLiteralBranchTypes) == 0 {
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
	symbolName := meta.symbolName

	ctx.ReportNodeWithDeferredSuggestions(
		meta.discriminant,
		rule.RuleMessage{
			Id:          "switchIsNotExhaustive",
			Description: "Switch is not exhaustive. Cases not matched: " + missingBranches,
			Data:        map[string]string{"missingBranches": missingBranches},
		},
		func() []rule.RuleSuggestion {
			cases := make([]missingCase, len(missingTypes))
			for i, t := range missingTypes {
				cases[i] = missingCase{typ: t}
			}
			return []rule.RuleSuggestion{
				{
					Message: rule.RuleMessage{
						Id:          "addMissingCases",
						Description: "Add branches for missing cases.",
					},
					FixesArr: []rule.RuleFix{
						fixSwitch(ctx, node, switchStmt, caseBlock, cases, defaultCase, symbolName),
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

	if len(meta.missingLiteralBranchTypes) == 0 && meta.defaultCase.exists() && !meta.containsNonLiteralType {
		ctx.ReportRange(meta.defaultCase.reportRange(ctx.SourceFile), rule.RuleMessage{
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
			return []rule.RuleSuggestion{
				{
					Message: rule.RuleMessage{
						Id:          "addMissingCases",
						Description: "Add branches for missing cases.",
					},
					FixesArr: []rule.RuleFix{
						fixSwitch(ctx, node, switchStmt, caseBlock, []missingCase{{isDefault: true}}, defaultCase, ""),
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
	symbolName string,
) rule.RuleFix {
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

	missingCasesText := make([]string, len(missingBranchTypes))
	for i, mb := range missingBranchTypes {
		if mb.isDefault {
			missingCasesText[i] = `default: { throw new Error('default case') }`
			continue
		}

		missingBranchType := mb.typ
		missingBranchName := ""
		hasMissingBranchName := false
		if sym := missingBranchType.Symbol(); sym != nil {
			missingBranchName = sym.Name
			hasMissingBranchName = true
		}

		var caseTest string
		if utils.IsTypeFlagSet(missingBranchType, checker.TypeFlagsESSymbolLike) {
			caseTest = missingBranchName
		} else {
			caseTest = typeToString(ctx.TypeChecker, missingBranchType)
		}

		if symbolName != "" && hasMissingBranchName && requiresQuoting(missingBranchName) {
			escapedBranchName := strings.NewReplacer(
				`'`, `\'`,
				"\n", `\n`,
				"\r", `\r`,
			).Replace(missingBranchName)
			caseTest = fmt.Sprintf("%s['%s']", symbolName, escapedBranchName)
		}

		escapedForMessage := strings.NewReplacer(
			`\`, `\\`,
			`'`, `\'`,
		).Replace(caseTest)

		missingCasesText[i] = fmt.Sprintf(
			"case %s: { throw new Error('Not implemented yet: %s case') }",
			caseTest,
			escapedForMessage,
		)
	}

	fixString := caseIndent + strings.Join(missingCasesText, "\n"+caseIndent)

	if lastCase != nil {
		if defaultCase.exists() {
			var beforeFixString strings.Builder
			for _, text := range missingCasesText {
				beforeFixString.WriteString(text)
				beforeFixString.WriteString("\n")
				beforeFixString.WriteString(caseIndent)
			}
			pos := defaultCase.insertPos(ctx.SourceFile)
			return rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), beforeFixString.String())
		}
		return rule.RuleFixInsertAfter(lastCase, "\n"+fixString)
	}

	// There were no existing cases: replace the whole `{ ... }` block.
	trimmed := utils.TrimNodeTextRange(ctx.SourceFile, switchStmt.CaseBlock)
	newText := "{\n" + fixString + "\n" + caseIndent + "}"
	return rule.RuleFixReplaceRange(trimmed, newText)
}
