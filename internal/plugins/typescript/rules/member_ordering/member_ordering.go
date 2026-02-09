package member_ordering

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	orderAsWritten                 = "as-written"
	orderAlphabetically            = "alphabetically"
	orderAlphabeticallyInsensitive = "alphabetically-case-insensitive"
	orderNatural                   = "natural"
	orderNaturalInsensitive        = "natural-case-insensitive"

	optionalityRequiredFirst = "required-first"
	optionalityOptionalFirst = "optional-first"
)

type memberTypeGroup struct {
	members []string
	isGroup bool
}

type parsedOrderConfig struct {
	disabled         bool
	memberTypes      []memberTypeGroup
	memberTypesNever bool
	order            string
	optionalityOrder string
}

type ruleOptions struct {
	defaultConfig    parsedOrderConfig
	classes          *parsedOrderConfig
	classExpressions *parsedOrderConfig
	interfaces       *parsedOrderConfig
	typeLiterals     *parsedOrderConfig
}

var defaultOrder = []string{
	// Index signature
	"signature",
	"call-signature",

	// Fields
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

	// Static initialization
	"static-initialization",

	// Constructors
	"public-constructor",
	"protected-constructor",
	"private-constructor",

	"constructor",

	// Accessors
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

	// Getters
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

	// Setters
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

	// Methods
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

var defaultMemberTypes = memberTypeGroupsFromStrings(defaultOrder)

var MemberOrderingRule = rule.CreateRule(rule.Rule{
	Name: "member-ordering",
	Run: func(ctx rule.RuleContext, rawOpts any) rule.RuleListeners {
		options := parseRuleOptions(rawOpts)

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				members := getClassMembers(node)
				if members == nil {
					return
				}
				validateMembersOrder(ctx, members, resolveOrderConfig(options.classes, options.defaultConfig), true)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				members := getClassMembers(node)
				if members == nil {
					return
				}
				validateMembersOrder(ctx, members, resolveOrderConfig(options.classExpressions, options.defaultConfig), true)
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				decl := node.AsInterfaceDeclaration()
				if decl == nil || decl.Members == nil {
					return
				}
				validateMembersOrder(ctx, decl.Members.Nodes, resolveOrderConfig(options.interfaces, options.defaultConfig), false)
			},
			ast.KindTypeLiteral: func(node *ast.Node) {
				literal := node.AsTypeLiteralNode()
				if literal == nil || literal.Members == nil {
					return
				}
				validateMembersOrder(ctx, literal.Members.Nodes, resolveOrderConfig(options.typeLiterals, options.defaultConfig), false)
			},
		}
	},
})

func resolveOrderConfig(config *parsedOrderConfig, fallback parsedOrderConfig) parsedOrderConfig {
	if config != nil {
		return *config
	}
	return fallback
}

func parseRuleOptions(raw any) ruleOptions {
	defaultConfig := parsedOrderConfig{
		memberTypes: defaultMemberTypes,
		order:       orderAsWritten,
	}
	result := ruleOptions{defaultConfig: defaultConfig}

	optsMap := extractOptionsMap(raw)
	if optsMap == nil {
		return result
	}

	if parsed := parseOrderConfig(optsMap["default"], defaultConfig.memberTypes); parsed != nil {
		result.defaultConfig = *parsed
	}
	if parsed := parseOrderConfig(optsMap["classes"], result.defaultConfig.memberTypes); parsed != nil {
		result.classes = parsed
	}
	if parsed := parseOrderConfig(optsMap["classExpressions"], result.defaultConfig.memberTypes); parsed != nil {
		result.classExpressions = parsed
	}
	if parsed := parseOrderConfig(optsMap["interfaces"], result.defaultConfig.memberTypes); parsed != nil {
		result.interfaces = parsed
	}
	if parsed := parseOrderConfig(optsMap["typeLiterals"], result.defaultConfig.memberTypes); parsed != nil {
		result.typeLiterals = parsed
	}

	return result
}

func extractOptionsMap(raw any) map[string]interface{} {
	if raw == nil {
		return nil
	}
	if arr, ok := raw.([]interface{}); ok {
		if len(arr) == 0 {
			return nil
		}
		opts, _ := arr[0].(map[string]interface{})
		return opts
	}
	opts, _ := raw.(map[string]interface{})
	return opts
}

func parseOrderConfig(raw any, defaultMemberTypes []memberTypeGroup) *parsedOrderConfig {
	if raw == nil {
		return nil
	}

	switch value := raw.(type) {
	case string:
		if value == "never" {
			return &parsedOrderConfig{disabled: true}
		}
	case []interface{}:
		return &parsedOrderConfig{
			memberTypes: parseMemberTypeGroups(value),
			order:       orderAsWritten,
		}
	case []string:
		return &parsedOrderConfig{
			memberTypes: memberTypeGroupsFromStrings(value),
			order:       orderAsWritten,
		}
	case map[string]interface{}:
		config := parsedOrderConfig{
			order: orderAsWritten,
		}

		if orderValue, ok := value["order"].(string); ok {
			config.order = orderValue
		}
		if optionality, ok := value["optionalityOrder"].(string); ok {
			config.optionalityOrder = optionality
		}

		if memberTypes, ok := value["memberTypes"]; ok {
			switch typed := memberTypes.(type) {
			case string:
				if typed == "never" {
					config.memberTypesNever = true
				} else {
					config.memberTypes = []memberTypeGroup{{members: []string{typed}}}
				}
			case []interface{}:
				config.memberTypes = parseMemberTypeGroups(typed)
			case []string:
				config.memberTypes = memberTypeGroupsFromStrings(typed)
			}
		} else {
			config.memberTypes = defaultMemberTypes
		}

		return &config
	}

	return nil
}

func memberTypeGroupsFromStrings(values []string) []memberTypeGroup {
	groups := make([]memberTypeGroup, 0, len(values))
	for _, value := range values {
		groups = append(groups, memberTypeGroup{members: []string{value}})
	}
	return groups
}

func parseMemberTypeGroups(values []interface{}) []memberTypeGroup {
	groups := make([]memberTypeGroup, 0, len(values))
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			groups = append(groups, memberTypeGroup{members: []string{typed}})
		case []interface{}:
			groupMembers := make([]string, 0, len(typed))
			for _, groupItem := range typed {
				if str, ok := groupItem.(string); ok {
					groupMembers = append(groupMembers, str)
				}
			}
			groups = append(groups, memberTypeGroup{members: groupMembers, isGroup: true})
		case []string:
			groupMembers := make([]string, 0, len(typed))
			groupMembers = append(groupMembers, typed...)
			groups = append(groups, memberTypeGroup{members: groupMembers, isGroup: true})
		}
	}
	return groups
}

func validateMembersOrder(ctx rule.RuleContext, members []*ast.Node, config parsedOrderConfig, supportsModifiers bool) {
	if config.disabled {
		return
	}

	order := config.order
	if order == "" {
		order = orderAsWritten
	}

	memberTypes := config.memberTypes
	if config.memberTypesNever {
		memberTypes = nil
	}

	checkAlphaSortForAllMembers := func(memberSet []*ast.Node) {
		hasAlphaSort := order != "" && order != orderAsWritten
		if !hasAlphaSort || len(memberTypes) == 0 {
			return
		}
		for _, group := range groupMembersByType(memberSet, memberTypes, supportsModifiers, ctx.SourceFile) {
			checkAlphaSort(ctx, group, order)
		}
	}

	checkOrder := func(memberSet []*ast.Node) bool {
		hasAlphaSort := order != "" && order != orderAsWritten
		if len(memberTypes) > 0 {
			grouped := checkGroupSort(ctx, memberSet, memberTypes, supportsModifiers)
			if grouped == nil {
				checkAlphaSortForAllMembers(memberSet)
				return false
			}
			if hasAlphaSort {
				for _, group := range grouped {
					checkAlphaSort(ctx, group, order)
				}
			}
		} else if hasAlphaSort {
			return checkAlphaSort(ctx, memberSet, order)
		}
		return false
	}

	if config.optionalityOrder == "" {
		checkOrder(members)
		return
	}

	switchIndex := firstOptionalitySwitchIndex(members)
	if switchIndex != -1 {
		if !checkRequiredOrder(ctx, members, config.optionalityOrder) {
			return
		}
		checkOrder(members[:switchIndex])
		checkOrder(members[switchIndex:])
	} else {
		checkOrder(members)
	}
}

func firstOptionalitySwitchIndex(members []*ast.Node) int {
	for i := 1; i < len(members); i++ {
		if isMemberOptional(members[i]) != isMemberOptional(members[i-1]) {
			return i
		}
	}
	return -1
}

func checkGroupSort(ctx rule.RuleContext, members []*ast.Node, memberTypes []memberTypeGroup, supportsModifiers bool) [][]*ast.Node {
	previousRanks := make([]int, 0, 4)
	memberGroups := make([][]*ast.Node, 0, 4)
	isCorrectlySorted := true

	for _, member := range members {
		rank := getRank(ctx.SourceFile, member, memberTypes, supportsModifiers)
		if rank == -1 {
			continue
		}

		if len(previousRanks) == 0 {
			previousRanks = append(previousRanks, rank)
			memberGroups = append(memberGroups, []*ast.Node{member})
			continue
		}

		lastRank := previousRanks[len(previousRanks)-1]
		if rank < lastRank {
			ctx.ReportNode(member, messageIncorrectGroupOrder(
				getMemberName(ctx, member),
				getLowestRank(previousRanks, rank, memberTypes),
			))
			isCorrectlySorted = false
		} else if rank == lastRank {
			memberGroups[len(memberGroups)-1] = append(memberGroups[len(memberGroups)-1], member)
		} else {
			previousRanks = append(previousRanks, rank)
			memberGroups = append(memberGroups, []*ast.Node{member})
		}
	}

	if !isCorrectlySorted {
		return nil
	}
	return memberGroups
}

func checkAlphaSort(ctx rule.RuleContext, members []*ast.Node, order string) bool {
	previousName := ""
	isCorrectlySorted := true

	for _, member := range members {
		name := getMemberName(ctx, member)
		if name == "" {
			continue
		}

		if naturalOutOfOrder(name, previousName, order) {
			ctx.ReportNode(member, messageIncorrectOrder(name, previousName))
			isCorrectlySorted = false
		}
		previousName = name
	}

	return isCorrectlySorted
}

func naturalOutOfOrder(name string, previousName string, order string) bool {
	if name == previousName {
		return false
	}

	switch order {
	case orderAlphabetically:
		return name < previousName
	case orderAlphabeticallyInsensitive:
		return strings.ToLower(name) < strings.ToLower(previousName)
	case orderNatural:
		return naturalCompare(name, previousName) != 1
	case orderNaturalInsensitive:
		return naturalCompare(strings.ToLower(name), strings.ToLower(previousName)) != 1
	default:
		return false
	}
}

func checkRequiredOrder(ctx rule.RuleContext, members []*ast.Node, optionalityOrder string) bool {
	switchIndex := firstOptionalitySwitchIndex(members)
	if switchIndex == -1 {
		return true
	}

	report := func(member *ast.Node) {
		requiredLabel := "required"
		if optionalityOrder == optionalityOptionalFirst {
			requiredLabel = "optional"
		}
		ctx.ReportNode(member, messageIncorrectRequiredMembersOrder(getMemberName(ctx, member), requiredLabel))
	}

	firstIsOptional := isMemberOptional(members[0])
	if firstIsOptional != (optionalityOrder == optionalityOptionalFirst) {
		report(members[0])
		return false
	}

	for i := switchIndex + 1; i < len(members); i++ {
		if isMemberOptional(members[i]) != isMemberOptional(members[switchIndex]) {
			report(members[switchIndex])
			return false
		}
	}

	return true
}

func groupMembersByType(members []*ast.Node, memberTypes []memberTypeGroup, supportsModifiers bool, sourceFile *ast.SourceFile) [][]*ast.Node {
	grouped := make([][]*ast.Node, 0, 4)
	if len(members) == 0 {
		return grouped
	}

	memberRanks := make([]int, len(members))
	for i, member := range members {
		memberRanks[i] = getRank(sourceFile, member, memberTypes, supportsModifiers)
	}

	previousRank := 0
	hasPrevious := false

	for i, member := range members {
		if i == len(members)-1 {
			return grouped
		}
		currentRank := memberRanks[i]
		nextRank := memberRanks[i+1]

		if hasPrevious && currentRank == previousRank {
			grouped[len(grouped)-1] = append(grouped[len(grouped)-1], member)
		} else if currentRank == nextRank {
			grouped = append(grouped, []*ast.Node{member})
			previousRank = currentRank
			hasPrevious = true
		}
	}

	return grouped
}

func getLowestRank(ranks []int, target int, order []memberTypeGroup) string {
	if len(ranks) == 0 {
		return ""
	}

	lowest := ranks[len(ranks)-1]
	for _, rank := range ranks {
		if rank > target {
			if rank < lowest {
				lowest = rank
			}
		}
	}

	if lowest < 0 || lowest >= len(order) {
		return ""
	}

	entries := order[lowest]
	var rankNames []string
	if entries.isGroup {
		for _, rank := range entries.members {
			rankNames = append(rankNames, strings.ReplaceAll(rank, "-", " "))
		}
	} else if len(entries.members) > 0 {
		rankNames = append(rankNames, strings.ReplaceAll(entries.members[0], "-", " "))
	}

	return strings.Join(rankNames, ", ")
}

func getRank(sourceFile *ast.SourceFile, node *ast.Node, orderConfig []memberTypeGroup, supportsModifiers bool) int {
	nodeType := getNodeType(node)
	if nodeType == "" {
		if len(orderConfig) == 0 {
			return -1
		}
		return len(orderConfig) - 1
	}

	abstract := ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract)
	if node.Kind == ast.KindMethodDeclaration {
		method := node.AsMethodDeclaration()
		if method != nil && method.Body == nil && !abstract {
			return -1
		}
	}
	if node.Kind == ast.KindConstructor {
		constructor := node.AsConstructorDeclaration()
		if constructor != nil && constructor.Body == nil {
			return -1
		}
	}
	scope := "instance"
	if ast.IsStatic(node) {
		scope = "static"
	} else if abstract {
		scope = "abstract"
	}

	accessibility := getAccessibility(node)

	memberGroups := make([]string, 0, 6)
	if supportsModifiers {
		decorated := ast.HasDecorators(node)
		if decorated && supportsDecorators(nodeType) {
			memberGroups = append(memberGroups, accessibility+"-decorated-"+nodeType)
			memberGroups = append(memberGroups, "decorated-"+nodeType)
			if nodeType == "readonly-field" {
				memberGroups = append(memberGroups, accessibility+"-decorated-field")
				memberGroups = append(memberGroups, "decorated-field")
			}
		}

		if nodeType != "readonly-signature" && nodeType != "signature" && nodeType != "static-initialization" {
			if nodeType != "constructor" {
				memberGroups = append(memberGroups, accessibility+"-"+scope+"-"+nodeType)
				memberGroups = append(memberGroups, scope+"-"+nodeType)
				if nodeType == "readonly-field" {
					memberGroups = append(memberGroups, accessibility+"-"+scope+"-field")
					memberGroups = append(memberGroups, scope+"-field")
				}
			}
			memberGroups = append(memberGroups, accessibility+"-"+nodeType)
			if nodeType == "readonly-field" {
				memberGroups = append(memberGroups, accessibility+"-field")
			}
		}
	}

	memberGroups = append(memberGroups, nodeType)
	switch nodeType {
	case "readonly-signature":
		memberGroups = append(memberGroups, "signature")
	case "readonly-field":
		memberGroups = append(memberGroups, "field")
	}

	return getRankOrder(memberGroups, orderConfig)
}

func supportsDecorators(nodeType string) bool {
	switch nodeType {
	case "readonly-field", "field", "method", "accessor", "get", "set":
		return true
	default:
		return false
	}
}

func getRankOrder(memberGroups []string, orderConfig []memberTypeGroup) int {
	for _, memberGroup := range memberGroups {
		for index, memberType := range orderConfig {
			if memberType.isGroup {
				for _, entry := range memberType.members {
					if entry == memberGroup {
						return index
					}
				}
				continue
			}
			if len(memberType.members) > 0 && memberType.members[0] == memberGroup {
				return index
			}
		}
	}
	return -1
}

func getNodeType(node *ast.Node) string {
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindMethodSignature:
		return "method"
	case ast.KindConstructor, ast.KindConstructSignature:
		return "constructor"
	case ast.KindCallSignature:
		return "call-signature"
	case ast.KindGetAccessor:
		return "get"
	case ast.KindSetAccessor:
		return "set"
	case ast.KindPropertySignature:
		if isReadonly(node) {
			return "readonly-field"
		}
		return "field"
	case ast.KindPropertyDeclaration:
		if ast.IsAutoAccessorPropertyDeclaration(node) {
			return "accessor"
		}
		property := node.AsPropertyDeclaration()
		if property != nil && property.Initializer != nil {
			if property.Initializer.Kind == ast.KindFunctionExpression || property.Initializer.Kind == ast.KindArrowFunction {
				return "method"
			}
		}
		if isReadonly(node) {
			return "readonly-field"
		}
		return "field"
	case ast.KindIndexSignature:
		if isReadonly(node) {
			return "readonly-signature"
		}
		return "signature"
	case ast.KindClassStaticBlockDeclaration:
		return "static-initialization"
	}

	return ""
}

func getAccessibility(node *ast.Node) string {
	flags := ast.GetCombinedModifierFlags(node)
	if flags&ast.ModifierFlagsPrivate != 0 {
		return "private"
	}
	if flags&ast.ModifierFlagsProtected != 0 {
		return "protected"
	}
	if flags&ast.ModifierFlagsPublic != 0 {
		return "public"
	}

	nameNode := getNameNode(node)
	if nameNode != nil && nameNode.Kind == ast.KindPrivateIdentifier {
		return "#private"
	}

	return "public"
}

func getMemberName(ctx rule.RuleContext, node *ast.Node) string {
	switch node.Kind {
	case ast.KindPropertySignature, ast.KindMethodSignature, ast.KindPropertyDeclaration, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor:
		return getMemberRawName(ctx, getNameNode(node))
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

func getNameNode(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		if decl := node.AsPropertyDeclaration(); decl != nil {
			return decl.Name()
		}
	case ast.KindMethodDeclaration:
		if decl := node.AsMethodDeclaration(); decl != nil {
			return decl.Name()
		}
	case ast.KindGetAccessor:
		if decl := node.AsGetAccessorDeclaration(); decl != nil {
			return decl.Name()
		}
	case ast.KindSetAccessor:
		if decl := node.AsSetAccessorDeclaration(); decl != nil {
			return decl.Name()
		}
	case ast.KindPropertySignature:
		if decl := node.AsPropertySignatureDeclaration(); decl != nil {
			return decl.Name()
		}
	case ast.KindMethodSignature:
		if decl := node.AsMethodSignatureDeclaration(); decl != nil {
			return decl.Name()
		}
	}
	return nil
}

func getMemberRawName(ctx rule.RuleContext, nameNode *ast.Node) string {
	if nameNode == nil {
		return ""
	}

	switch nameNode.Kind {
	case ast.KindIdentifier:
		return nameNode.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return nameNode.AsPrivateIdentifier().Text
	case ast.KindStringLiteral:
		return nameNode.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		return nameNode.AsNumericLiteral().Text
	case ast.KindBigIntLiteral:
		return nameNode.AsBigIntLiteral().Text
	case ast.KindComputedPropertyName:
		expr := nameNode.AsComputedPropertyName().Expression
		if expr != nil && ast.IsLiteralExpression(expr) {
			switch expr.Kind {
			case ast.KindStringLiteral:
				return expr.AsStringLiteral().Text
			case ast.KindNumericLiteral:
				return expr.AsNumericLiteral().Text
			case ast.KindBigIntLiteral:
				return expr.AsBigIntLiteral().Text
			default:
				return expr.Text()
			}
		}
	}

	trimmed := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
	return ctx.SourceFile.Text()[trimmed.Pos():trimmed.End()]
}

func getNameFromIndexSignature(node *ast.Node) string {
	indexSig := node.AsIndexSignatureDeclaration()
	if indexSig == nil || indexSig.Parameters == nil {
		return "(index signature)"
	}

	for _, param := range indexSig.Parameters.Nodes {
		if param.Kind != ast.KindParameter {
			continue
		}
		name := param.AsParameterDeclaration().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text
		}
	}

	return "(index signature)"
}

func isMemberOptional(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindPropertySignature:
		if decl := node.AsPropertySignatureDeclaration(); decl != nil {
			return decl.PostfixToken != nil && decl.PostfixToken.Kind == ast.KindQuestionToken
		}
	case ast.KindMethodSignature:
		if decl := node.AsMethodSignatureDeclaration(); decl != nil {
			return decl.PostfixToken != nil && decl.PostfixToken.Kind == ast.KindQuestionToken
		}
	case ast.KindPropertyDeclaration:
		if decl := node.AsPropertyDeclaration(); decl != nil {
			return decl.PostfixToken != nil && decl.PostfixToken.Kind == ast.KindQuestionToken
		}
	case ast.KindMethodDeclaration:
		if decl := node.AsMethodDeclaration(); decl != nil {
			return decl.PostfixToken != nil && decl.PostfixToken.Kind == ast.KindQuestionToken
		}
	}
	return false
}

func isReadonly(node *ast.Node) bool {
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly)
}

func getClassMembers(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindClassDeclaration:
		classDecl := node.AsClassDeclaration()
		if classDecl == nil || classDecl.Members == nil {
			return nil
		}
		return classDecl.Members.Nodes
	case ast.KindClassExpression:
		classExpr := node.AsClassExpression()
		if classExpr == nil || classExpr.Members == nil {
			return nil
		}
		return classExpr.Members.Nodes
	default:
		return nil
	}
}

func naturalCompare(first string, second string) int {
	if first == second {
		return 0
	}

	firstIndex := 0
	secondIndex := 0

	for firstIndex < len(first) && secondIndex < len(second) {
		firstRune, firstSize := utf8.DecodeRuneInString(first[firstIndex:])
		secondRune, secondSize := utf8.DecodeRuneInString(second[secondIndex:])

		if unicode.IsDigit(firstRune) && unicode.IsDigit(secondRune) {
			firstStart := firstIndex
			for firstIndex < len(first) {
				r, size := utf8.DecodeRuneInString(first[firstIndex:])
				if !unicode.IsDigit(r) {
					break
				}
				firstIndex += size
			}

			secondStart := secondIndex
			for secondIndex < len(second) {
				r, size := utf8.DecodeRuneInString(second[secondIndex:])
				if !unicode.IsDigit(r) {
					break
				}
				secondIndex += size
			}

			firstNumber := first[firstStart:firstIndex]
			secondNumber := second[secondStart:secondIndex]

			firstTrimmed := strings.TrimLeft(firstNumber, "0")
			secondTrimmed := strings.TrimLeft(secondNumber, "0")
			if firstTrimmed == "" {
				firstTrimmed = "0"
			}
			if secondTrimmed == "" {
				secondTrimmed = "0"
			}

			if len(firstTrimmed) != len(secondTrimmed) {
				if len(firstTrimmed) < len(secondTrimmed) {
					return -1
				}
				return 1
			}
			if firstTrimmed != secondTrimmed {
				if firstTrimmed < secondTrimmed {
					return -1
				}
				return 1
			}
			if len(firstNumber) != len(secondNumber) {
				if len(firstNumber) < len(secondNumber) {
					return -1
				}
				return 1
			}

			continue
		}

		if firstRune != secondRune {
			if firstRune < secondRune {
				return -1
			}
			return 1
		}

		firstIndex += firstSize
		secondIndex += secondSize
	}

	if firstIndex == len(first) && secondIndex == len(second) {
		return 0
	}
	if firstIndex == len(first) {
		return -1
	}
	return 1
}

func messageIncorrectGroupOrder(name string, rank string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "incorrectGroupOrder",
		Description: fmt.Sprintf("Member %s should be declared before all %s definitions.", name, rank),
	}
}

func messageIncorrectOrder(member string, before string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "incorrectOrder",
		Description: fmt.Sprintf("Member %s should be declared before member %s.", member, before),
	}
}

func messageIncorrectRequiredMembersOrder(member string, optionalOrRequired string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "incorrectRequiredMembersOrder",
		Description: fmt.Sprintf("Member %s should be declared after all %s members.", member, optionalOrRequired),
	}
}
