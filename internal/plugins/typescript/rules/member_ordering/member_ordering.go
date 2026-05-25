package member_ordering

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// --- Message builders ---

func messageIncorrectGroupOrder(name, rank string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "incorrectGroupOrder",
		Description: fmt.Sprintf("Member %s should be declared before all %s definitions.", name, rank),
	}
}

func messageIncorrectOrder(member, beforeMember string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "incorrectOrder",
		Description: fmt.Sprintf("Member %s should be declared before member %s.", member, beforeMember),
	}
}

func messageIncorrectRequiredMembersOrder(member, optionalOrRequired string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "incorrectRequiredMembersOrder",
		Description: fmt.Sprintf("Member %s should be declared after all %s members.", member, optionalOrRequired),
	}
}

// --- Types ---

const (
	orderAsWritten              = "as-written"
	orderAlphabetically         = "alphabetically"
	orderAlphaCaseInsensitive   = "alphabetically-case-insensitive"
	orderNatural                = "natural"
	orderNaturalCaseInsensitive = "natural-case-insensitive"
	optionalityRequiredFirst    = "required-first"
)

// memberKind represents the kind of a class/interface member
type memberKind string

const (
	kindAccessor          memberKind = "accessor"
	kindCallSignature     memberKind = "call-signature"
	kindConstructor       memberKind = "constructor"
	kindField             memberKind = "field"
	kindReadonlyField     memberKind = "readonly-field"
	kindGet               memberKind = "get"
	kindMethod            memberKind = "method"
	kindSet               memberKind = "set"
	kindSignature         memberKind = "signature"
	kindReadonlySignature memberKind = "readonly-signature"
	kindStaticInit        memberKind = "static-initialization"
)

// parsedConfig holds the parsed configuration for a single context
type parsedConfig struct {
	memberTypes      []interface{} // each element is either string or []string
	order            string
	optionalityOrder string
	neverCheck       bool
}

// --- Default order ---

var defaultOrder = []interface{}{
	"signature",
	"call-signature",

	"public-static-field",
	"protected-static-field",
	"private-static-field",
	"#private-static-field",

	"public-decorated-field",
	"protected-decorated-field",
	"private-decorated-field",

	"public-instance-field",
	"protected-instance-field",
	"private-instance-field",
	"#private-instance-field",

	"public-abstract-field",
	"protected-abstract-field",

	"public-field",
	"protected-field",
	"private-field",
	"#private-field",

	"static-field",
	"instance-field",
	"abstract-field",

	"decorated-field",

	"field",

	"static-initialization",

	"public-constructor",
	"protected-constructor",
	"private-constructor",

	"constructor",

	"public-static-accessor",
	"protected-static-accessor",
	"private-static-accessor",
	"#private-static-accessor",

	"public-decorated-accessor",
	"protected-decorated-accessor",
	"private-decorated-accessor",

	"public-instance-accessor",
	"protected-instance-accessor",
	"private-instance-accessor",
	"#private-instance-accessor",

	"public-abstract-accessor",
	"protected-abstract-accessor",

	"public-accessor",
	"protected-accessor",
	"private-accessor",
	"#private-accessor",

	"static-accessor",
	"instance-accessor",
	"abstract-accessor",

	"decorated-accessor",

	"accessor",

	"public-static-get",
	"protected-static-get",
	"private-static-get",
	"#private-static-get",

	"public-decorated-get",
	"protected-decorated-get",
	"private-decorated-get",

	"public-instance-get",
	"protected-instance-get",
	"private-instance-get",
	"#private-instance-get",

	"public-abstract-get",
	"protected-abstract-get",

	"public-get",
	"protected-get",
	"private-get",
	"#private-get",

	"static-get",
	"instance-get",
	"abstract-get",

	"decorated-get",

	"get",

	"public-static-set",
	"protected-static-set",
	"private-static-set",
	"#private-static-set",

	"public-decorated-set",
	"protected-decorated-set",
	"private-decorated-set",

	"public-instance-set",
	"protected-instance-set",
	"private-instance-set",
	"#private-instance-set",

	"public-abstract-set",
	"protected-abstract-set",

	"public-set",
	"protected-set",
	"private-set",
	"#private-set",

	"static-set",
	"instance-set",
	"abstract-set",

	"decorated-set",

	"set",

	"public-static-method",
	"protected-static-method",
	"private-static-method",
	"#private-static-method",

	"public-decorated-method",
	"protected-decorated-method",
	"private-decorated-method",

	"public-instance-method",
	"protected-instance-method",
	"private-instance-method",
	"#private-instance-method",

	"public-abstract-method",
	"protected-abstract-method",

	"public-method",
	"protected-method",
	"private-method",
	"#private-method",

	"static-method",
	"instance-method",
	"abstract-method",

	"decorated-method",

	"method",
}

// --- Options parsing ---

type ruleOptions struct {
	defaultConfig          *parsedConfig
	classesConfig          *parsedConfig
	classExpressionsConfig *parsedConfig
	interfacesConfig       *parsedConfig
	typeLiteralsConfig     *parsedConfig
}

func parseConfig(val interface{}) *parsedConfig {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case string:
		if v == "never" {
			return &parsedConfig{neverCheck: true}
		}
		return nil
	case []interface{}:
		return &parsedConfig{
			memberTypes: convertToMemberTypes(v),
			order:       orderAsWritten,
		}
	case map[string]interface{}:
		cfg := &parsedConfig{order: orderAsWritten}

		if mt, ok := v["memberTypes"]; ok {
			switch mtv := mt.(type) {
			case string:
				if mtv == "never" {
					cfg.memberTypes = nil
				}
			case []interface{}:
				cfg.memberTypes = convertToMemberTypes(mtv)
			}
		} else {
			// If memberTypes not specified, use default
			cfg.memberTypes = defaultOrder
		}

		if order, ok := v["order"].(string); ok {
			cfg.order = order
		}
		if oo, ok := v["optionalityOrder"].(string); ok {
			cfg.optionalityOrder = oo
		}
		return cfg
	}

	return nil
}

func convertToMemberTypes(arr []interface{}) []interface{} {
	result := make([]interface{}, 0, len(arr))
	for _, item := range arr {
		switch v := item.(type) {
		case string:
			result = append(result, v)
		case []interface{}:
			group := make([]string, 0, len(v))
			for _, s := range v {
				if str, ok := s.(string); ok {
					group = append(group, str)
				}
			}
			result = append(result, group)
		}
	}
	return result
}

func parseOptions(options any) ruleOptions {
	opts := ruleOptions{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		opts.defaultConfig = &parsedConfig{
			memberTypes: defaultOrder,
			order:       orderAsWritten,
		}
		return opts
	}

	if def, ok := optsMap["default"]; ok {
		opts.defaultConfig = parseConfig(def)
	}
	if cl, ok := optsMap["classes"]; ok {
		opts.classesConfig = parseConfig(cl)
	}
	if ce, ok := optsMap["classExpressions"]; ok {
		opts.classExpressionsConfig = parseConfig(ce)
	}
	if itf, ok := optsMap["interfaces"]; ok {
		opts.interfacesConfig = parseConfig(itf)
	}
	if tl, ok := optsMap["typeLiterals"]; ok {
		opts.typeLiteralsConfig = parseConfig(tl)
	}

	// If no default config is set and no other config is set, use the default order
	if opts.defaultConfig == nil && opts.classesConfig == nil && opts.classExpressionsConfig == nil &&
		opts.interfacesConfig == nil && opts.typeLiteralsConfig == nil {
		opts.defaultConfig = &parsedConfig{
			memberTypes: defaultOrder,
			order:       orderAsWritten,
		}
	}

	return opts
}

// --- Node analysis helpers ---

// getNodeType returns the member kind for a given AST node
func getNodeType(node *ast.Node) memberKind {
	// Check for auto accessor (accessor keyword on property) before property declaration
	if ast.IsAutoAccessorPropertyDeclaration(node) {
		return kindAccessor
	}

	switch node.Kind {
	case ast.KindMethodDeclaration:
		return kindMethod

	case ast.KindMethodSignature:
		return kindMethod

	case ast.KindConstructor:
		return kindConstructor

	case ast.KindPropertyDeclaration:
		// Check if value is a function expression or arrow function → method
		propDecl := node.AsPropertyDeclaration()
		if propDecl.Initializer != nil {
			initKind := propDecl.Initializer.Kind
			if initKind == ast.KindFunctionExpression || initKind == ast.KindArrowFunction {
				return kindMethod
			}
		}
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly) {
			return kindReadonlyField
		}
		return kindField

	case ast.KindPropertySignature:
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly) {
			return kindReadonlyField
		}
		return kindField

	case ast.KindGetAccessor:
		return kindGet

	case ast.KindSetAccessor:
		return kindSet

	case ast.KindCallSignature:
		return kindCallSignature

	case ast.KindConstructSignature:
		return kindConstructor

	case ast.KindIndexSignature:
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly) {
			return kindReadonlySignature
		}
		return kindSignature

	case ast.KindClassStaticBlockDeclaration:
		return kindStaticInit

	}

	return ""
}

// getAccessibility returns the accessibility modifier: "public", "protected", "private",
// or "#private" for JS private fields (PrivateIdentifier). Defaults to "public".
func getAccessibility(node *ast.Node) string {
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate) {
		return "private"
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsProtected) {
		return "protected"
	}
	if name := node.Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
		return "#private"
	}
	return "public"
}

// getScope returns the scope of a member (static, abstract, instance)
func getScope(node *ast.Node) string {
	if ast.IsStatic(node) {
		return "static"
	}
	if ast.HasAbstractModifier(node) {
		return "abstract"
	}
	return "instance"
}

// canBeDecorated returns whether a member kind supports the "decorated" modifier group
func canBeDecorated(kind memberKind) bool {
	switch kind {
	case kindField, kindReadonlyField, kindMethod, kindAccessor, kindGet, kindSet:
		return true
	}
	return false
}

// getMemberName returns the name used for sorting and error messages.
// Returns "constructor", "new", "call", "static block" for special members.
func getMemberName(node *ast.Node) string {
	// Handle auto accessor property before switch
	if ast.IsAutoAccessorPropertyDeclaration(node) {
		nameNode := node.Name()
		if nameNode == nil {
			return ""
		}
		return getMemberRawName(nameNode)
	}

	switch node.Kind {
	case ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindMethodDeclaration,
		ast.KindMethodSignature, ast.KindGetAccessor, ast.KindSetAccessor:
		nameNode := node.Name()
		if nameNode == nil {
			return ""
		}
		return getMemberRawName(nameNode)

	case ast.KindConstructor:
		return "constructor"

	case ast.KindConstructSignature:
		return "new"

	case ast.KindCallSignature:
		return "call"

	case ast.KindIndexSignature:
		return getNameFromIndexSignature(node)

	case ast.KindClassStaticBlockDeclaration:
		return "static block"
	}

	return ""
}

// getMemberRawName extracts the sortable name from a member name node.
// Strips quotes from string literals and # from private identifiers.
func getMemberRawName(nameNode *ast.Node) string {
	// PrivateIdentifier: strip leading #
	if nameNode.Kind == ast.KindPrivateIdentifier {
		text := nameNode.AsPrivateIdentifier().Text
		if len(text) > 0 && text[0] == '#' {
			return text[1:]
		}
		return text
	}
	// GetStaticPropertyName handles Identifier, StringLiteral, NumericLiteral,
	// and ComputedPropertyName with static expressions — returns unquoted text.
	if name, ok := utils.GetStaticPropertyName(nameNode); ok {
		return name
	}
	return ""
}

// getNameFromIndexSignature returns the parameter name of an index signature.
// e.g., `[key: string]: any` → "key"
func getNameFromIndexSignature(node *ast.Node) string {
	indexSig := node.AsIndexSignatureDeclaration()
	if indexSig == nil || indexSig.Parameters == nil || len(indexSig.Parameters.Nodes) == 0 {
		return "index"
	}

	param := indexSig.Parameters.Nodes[0]
	paramName := param.Name()
	if paramName == nil {
		return "index"
	}

	return paramName.Text()
}

// isMemberOptional returns whether a member is optional (has `?` token)
func isMemberOptional(node *ast.Node) bool {
	if ast.IsAutoAccessorPropertyDeclaration(node) {
		return ast.HasQuestionToken(node)
	}
	switch node.Kind {
	case ast.KindPropertyDeclaration, ast.KindPropertySignature,
		ast.KindMethodDeclaration, ast.KindMethodSignature,
		ast.KindGetAccessor, ast.KindSetAccessor:
		return ast.HasQuestionToken(node)
	}
	return false
}

// isOverloadSignature detects TS overload signatures (method declarations without body)
// Abstract methods also have no body but are NOT overload signatures
func isOverloadSignature(node *ast.Node) bool {
	if ast.HasAbstractModifier(node) {
		return false
	}
	switch node.Kind {
	case ast.KindMethodDeclaration:
		return node.Body() == nil
	case ast.KindConstructor:
		return node.Body() == nil
	}
	return false
}

// --- Ranking ---

// getRankOrder finds the rank (index) of memberGroups in orderConfig
func getRankOrder(memberGroups []string, orderConfig []interface{}) int {
	for _, group := range memberGroups {
		for i, configEntry := range orderConfig {
			switch entry := configEntry.(type) {
			case string:
				if entry == group {
					return i
				}
			case []string:
				for _, s := range entry {
					if s == group {
						return i
					}
				}
			}
		}
	}
	return -1
}

// getRank computes the rank (index in orderConfig) for a member.
// Builds a list of candidate group names from most specific to least specific:
//
//	accessibility-decorated-type → decorated-type →
//	accessibility-scope-type → scope-type → accessibility-type → type
//
// Returns -1 for overload signatures (skipped), or orderConfig length - 1 for unknown types.
func getRank(node *ast.Node, orderConfig []interface{}, supportsModifiers bool) int {
	nodeType := getNodeType(node)
	if nodeType == "" {
		return len(orderConfig) - 1
	}

	// Skip overload signatures
	if isOverloadSignature(node) {
		return -1
	}

	memberType := string(nodeType)
	accessibility := getAccessibility(node)
	scope := getScope(node)
	decorated := ast.HasDecorators(node)

	var memberGroups []string

	if supportsModifiers {
		// Decorated variants
		if decorated && canBeDecorated(nodeType) && accessibility != "#private" {
			memberGroups = append(memberGroups, accessibility+"-decorated-"+memberType)
			if nodeType == kindReadonlyField {
				memberGroups = append(memberGroups, accessibility+"-decorated-field")
			}
			memberGroups = append(memberGroups, "decorated-"+memberType)
			if nodeType == kindReadonlyField {
				memberGroups = append(memberGroups, "decorated-field")
			}
		}

		// Scope-based variants (not for constructor, signature, call-signature, static-initialization)
		switch nodeType {
		case kindConstructor, kindSignature, kindReadonlySignature, kindCallSignature, kindStaticInit:
			// These don't have scope variants (except accessibility for constructor)
			if nodeType == kindConstructor {
				memberGroups = append(memberGroups, accessibility+"-"+memberType)
			}
		default:
			memberGroups = append(memberGroups, accessibility+"-"+scope+"-"+memberType)
			if nodeType == kindReadonlyField {
				memberGroups = append(memberGroups, accessibility+"-"+scope+"-field")
			}
			memberGroups = append(memberGroups, scope+"-"+memberType)
			if nodeType == kindReadonlyField {
				memberGroups = append(memberGroups, scope+"-field")
			}
			memberGroups = append(memberGroups, accessibility+"-"+memberType)
			if nodeType == kindReadonlyField {
				memberGroups = append(memberGroups, accessibility+"-field")
			}
		}
	}

	// Always add the base type
	memberGroups = append(memberGroups, memberType)
	if nodeType == kindReadonlyField {
		memberGroups = append(memberGroups, "field")
	}
	if nodeType == kindReadonlySignature {
		memberGroups = append(memberGroups, "signature")
	}

	return getRankOrder(memberGroups, orderConfig)
}

// getLowestRank returns the human-readable name of the first group that should come
// after `target` rank, used in "should be declared before all X definitions" messages.
func getLowestRank(ranks []int, target int, orderConfig []interface{}) string {
	if len(ranks) == 0 {
		return ""
	}
	lowestRank := ranks[len(ranks)-1]
	for _, r := range ranks {
		if r > target && r < lowestRank {
			lowestRank = r
		}
	}

	if lowestRank < 0 || lowestRank >= len(orderConfig) {
		return ""
	}

	entry := orderConfig[lowestRank]
	switch v := entry.(type) {
	case string:
		return strings.ReplaceAll(v, "-", " ")
	case []string:
		parts := make([]string, len(v))
		for i, s := range v {
			parts[i] = strings.ReplaceAll(s, "-", " ")
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

// --- Sorting helpers ---

// isOutOfOrder returns true if name should have been declared before previousName
func isOutOfOrder(name, previousName, order string) bool {
	if name == previousName {
		return false
	}

	switch order {
	case orderAlphabetically:
		return name < previousName
	case orderAlphaCaseInsensitive:
		return strings.ToLower(name) < strings.ToLower(previousName)
	case orderNatural:
		return utils.NaturalCompare(name, previousName) == -1
	case orderNaturalCaseInsensitive:
		return utils.NaturalCompare(strings.ToLower(name), strings.ToLower(previousName)) == -1
	}
	return false
}

// --- Validation ---

type memberInfo struct {
	node *ast.Node
	name string
	rank int
}

// checkGroupSort validates that members appear in the correct group order.
// previousRanks only accumulates strictly increasing ranks (like ESLint).
// Returns groups of members (by rank) if valid, nil if errors were found.
func checkGroupSort(ctx rule.RuleContext, members []*ast.Node, orderConfig []interface{}, supportsModifiers bool) [][]memberInfo {
	var previousRanks []int
	var memberGroups [][]memberInfo
	isCorrectlySorted := true

	for _, member := range members {
		rank := getRank(member, orderConfig, supportsModifiers)
		if rank == -1 {
			continue
		}

		name := getMemberName(member)
		info := memberInfo{node: member, name: name, rank: rank}

		rankLastMember := -1
		if len(previousRanks) > 0 {
			rankLastMember = previousRanks[len(previousRanks)-1]
		}

		if rankLastMember >= 0 && rank < rankLastMember {
			// Out of order — report error but do NOT push this rank
			rankName := getLowestRank(previousRanks, rank, orderConfig)
			ctx.ReportNode(member, messageIncorrectGroupOrder(name, rankName))
			isCorrectlySorted = false
		} else if rank == rankLastMember {
			// Same rank as previous → append to current group
			memberGroups[len(memberGroups)-1] = append(memberGroups[len(memberGroups)-1], info)
		} else {
			// New (higher) rank → push rank + start new group
			previousRanks = append(previousRanks, rank)
			memberGroups = append(memberGroups, []memberInfo{info})
		}
	}

	if !isCorrectlySorted {
		return nil
	}
	return memberGroups
}

// groupMembersByType groups all members by their rank (not just consecutive)
func groupMembersByType(members []*ast.Node, orderConfig []interface{}, supportsModifiers bool) [][]memberInfo {
	rankMap := make(map[int][]memberInfo)
	var rankOrder []int

	for _, member := range members {
		rank := getRank(member, orderConfig, supportsModifiers)
		if rank == -1 {
			continue
		}
		name := getMemberName(member)
		info := memberInfo{node: member, name: name, rank: rank}

		if _, exists := rankMap[rank]; !exists {
			rankOrder = append(rankOrder, rank)
		}
		rankMap[rank] = append(rankMap[rank], info)
	}

	var groups [][]memberInfo
	for _, r := range rankOrder {
		groups = append(groups, rankMap[r])
	}
	return groups
}

// checkAlphaSort checks alphabetical ordering within a group of members
func checkAlphaSort(ctx rule.RuleContext, members []memberInfo, order string) {
	previousName := ""
	for _, info := range members {
		if info.name == "" {
			continue
		}
		if previousName != "" && isOutOfOrder(info.name, previousName, order) {
			ctx.ReportNode(info.node, messageIncorrectOrder(info.name, previousName))
		}
		previousName = info.name
	}
}

// checkAlphaSortForMembers creates memberInfo slice from nodes and checks alpha sort
func checkAlphaSortForMembers(ctx rule.RuleContext, members []*ast.Node, order string) {
	var infos []memberInfo
	for _, m := range members {
		name := getMemberName(m)
		infos = append(infos, memberInfo{node: m, name: name})
	}
	checkAlphaSort(ctx, infos, order)
}

// checkOrder validates member ordering (group + alpha).
// allMembers is the full member list (used for alpha fallback when group sort fails,
// matching ESLint's behavior of using the outer `members` variable).
func checkOrder(ctx rule.RuleContext, memberSet []*ast.Node, allMembers []*ast.Node, cfg *parsedConfig, supportsModifiers bool) {
	hasAlphaSort := cfg.order != "" && cfg.order != orderAsWritten

	if cfg.memberTypes != nil {
		groups := checkGroupSort(ctx, memberSet, cfg.memberTypes, supportsModifiers)
		if groups == nil {
			// Group sort failed — alpha sort on full members grouped by type (ESLint behavior)
			if hasAlphaSort {
				typeGroups := groupMembersByType(allMembers, cfg.memberTypes, supportsModifiers)
				for _, group := range typeGroups {
					checkAlphaSort(ctx, group, cfg.order)
				}
			}
		} else if hasAlphaSort {
			// Group sort succeeded — alpha sort within each group
			for _, group := range groups {
				checkAlphaSort(ctx, group, cfg.order)
			}
		}
	} else if hasAlphaSort {
		// No member types grouping, just check alpha sort on all
		checkAlphaSortForMembers(ctx, memberSet, cfg.order)
	}
}

// validateMembersOrder is the main entry point for validating member ordering
func validateMembersOrder(ctx rule.RuleContext, members []*ast.Node, cfg *parsedConfig, supportsModifiers bool) {
	if cfg == nil || cfg.neverCheck {
		return
	}

	if cfg.optionalityOrder != "" {
		validateWithOptionality(ctx, members, cfg, supportsModifiers)
		return
	}

	checkOrder(ctx, members, members, cfg, supportsModifiers)
}

// validateWithOptionality handles optionality ordering (required-first / optional-first)
func validateWithOptionality(ctx rule.RuleContext, members []*ast.Node, cfg *parsedConfig, supportsModifiers bool) {
	// Find the switch point where optionality changes
	switchIndex := -1
	for i := 1; i < len(members); i++ {
		prevOptional := isMemberOptional(members[i-1])
		currOptional := isMemberOptional(members[i])
		if prevOptional != currOptional {
			switchIndex = i
			break
		}
	}

	if switchIndex == -1 {
		// No switch - all members have same optionality, just check order
		checkOrder(ctx, members, members, cfg, supportsModifiers)
		return
	}

	// Check if the switch is in the right direction
	firstIsOptional := isMemberOptional(members[0])
	isRequiredFirst := cfg.optionalityOrder == optionalityRequiredFirst

	switchIsValid := true
	if isRequiredFirst && firstIsOptional {
		// First member is optional but required-first is expected
		switchIsValid = false
		name := getMemberName(members[0])
		ctx.ReportNode(members[0], messageIncorrectRequiredMembersOrder(name, "required"))
	} else if !isRequiredFirst && !firstIsOptional {
		// First member is required but optional-first is expected
		switchIsValid = false
		name := getMemberName(members[0])
		ctx.ReportNode(members[0], messageIncorrectRequiredMembersOrder(name, "optional"))
	}

	if switchIsValid {
		// Check for additional switches after the first valid one
		// If found, report on the member at the first switch (switchIndex), NOT the second switch
		hasSecondSwitch := false
		for i := switchIndex + 1; i < len(members); i++ {
			prevOptional := isMemberOptional(members[i-1])
			currOptional := isMemberOptional(members[i])
			if prevOptional != currOptional {
				hasSecondSwitch = true
				break
			}
		}

		if hasSecondSwitch {
			// Report on the member at the first switch point
			name := getMemberName(members[switchIndex])
			if isRequiredFirst {
				ctx.ReportNode(members[switchIndex], messageIncorrectRequiredMembersOrder(name, "required"))
			} else {
				ctx.ReportNode(members[switchIndex], messageIncorrectRequiredMembersOrder(name, "optional"))
			}
			switchIsValid = false
		}
	}

	if switchIsValid {
		// Valid switch - check order on each half separately
		checkOrder(ctx, members[:switchIndex], members, cfg, supportsModifiers)
		checkOrder(ctx, members[switchIndex:], members, cfg, supportsModifiers)
	} else {
		// Invalid switch - still check order on all members
		checkOrder(ctx, members, members, cfg, supportsModifiers)
	}
}

// isSyntacticMember returns whether a node is a real member that participates in ordering.
// TS-Go AST includes SemicolonClassElement (standalone `;`) in Members.Nodes,
// but ESTree parsers strip these. We filter them out to match ESLint behavior.
func isSyntacticMember(node *ast.Node) bool {
	return node.Kind != ast.KindSemicolonClassElement
}

// getClassMembers extracts member nodes from a class/interface/type literal,
// filtering out non-semantic nodes like SemicolonClassElement.
func getClassMembers(node *ast.Node) []*ast.Node {
	var raw []*ast.Node
	switch node.Kind {
	case ast.KindClassDeclaration:
		classDecl := node.AsClassDeclaration()
		if classDecl == nil || classDecl.Members == nil {
			return nil
		}
		raw = classDecl.Members.Nodes
	case ast.KindClassExpression:
		classExpr := node.AsClassExpression()
		if classExpr == nil || classExpr.Members == nil {
			return nil
		}
		raw = classExpr.Members.Nodes
	case ast.KindInterfaceDeclaration:
		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl == nil || interfaceDecl.Members == nil {
			return nil
		}
		raw = interfaceDecl.Members.Nodes
	case ast.KindTypeLiteral:
		typeLiteral := node.AsTypeLiteralNode()
		if typeLiteral == nil || typeLiteral.Members == nil {
			return nil
		}
		raw = typeLiteral.Members.Nodes
	default:
		return nil
	}
	return utils.Filter(raw, isSyntacticMember)
}

// --- Rule definition ---

var MemberOrderingRule = rule.CreateRule(rule.Rule{
	Name: "member-ordering",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		getConfig := func(nodeKind ast.Kind) *parsedConfig {
			switch nodeKind {
			case ast.KindClassDeclaration:
				if opts.classesConfig != nil {
					return opts.classesConfig
				}
				return opts.defaultConfig
			case ast.KindClassExpression:
				if opts.classExpressionsConfig != nil {
					return opts.classExpressionsConfig
				}
				return opts.defaultConfig
			case ast.KindInterfaceDeclaration:
				if opts.interfacesConfig != nil {
					return opts.interfacesConfig
				}
				return opts.defaultConfig
			case ast.KindTypeLiteral:
				if opts.typeLiteralsConfig != nil {
					return opts.typeLiteralsConfig
				}
				return opts.defaultConfig
			}
			return opts.defaultConfig
		}

		validate := func(node *ast.Node) {
			members := getClassMembers(node)
			if members == nil {
				return
			}
			cfg := getConfig(node.Kind)
			supportsModifiers := node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression
			validateMembersOrder(ctx, members, cfg, supportsModifiers)
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration:     validate,
			ast.KindClassExpression:      validate,
			ast.KindInterfaceDeclaration: validate,
			ast.KindTypeLiteral:          validate,
		}
	},
})
