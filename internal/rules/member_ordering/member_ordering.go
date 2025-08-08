package member_ordering

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// MemberKind represents the type of class member
type MemberKind string

const (
	KindAccessor          MemberKind = "accessor"
	KindCallSignature     MemberKind = "call-signature"
	KindConstructor       MemberKind = "constructor"
	KindField             MemberKind = "field"
	KindGet               MemberKind = "get"
	KindMethod            MemberKind = "method"
	KindSet               MemberKind = "set"
	KindSignature         MemberKind = "signature"
	KindStaticInit        MemberKind = "static-initialization"
	KindReadonlyField     MemberKind = "readonly-field"
	KindReadonlySignature MemberKind = "readonly-signature"
)

// MemberScope represents the scope of a member
type MemberScope string

const (
	ScopeAbstract MemberScope = "abstract"
	ScopeInstance MemberScope = "instance"
	ScopeStatic   MemberScope = "static"
)

// Accessibility represents the accessibility modifier
type Accessibility string

const (
	AccessPrivateID Accessibility = "#private"
	AccessPublic    Accessibility = "public"
	AccessProtected Accessibility = "protected"
	AccessPrivate   Accessibility = "private"
)

// Order represents the sorting order
type Order string

const (
	OrderAsWritten                     Order = "as-written"
	OrderAlphabetically                Order = "alphabetically"
	OrderAlphabeticallyCaseInsensitive Order = "alphabetically-case-insensitive"
	OrderNatural                       Order = "natural"
	OrderNaturalCaseInsensitive        Order = "natural-case-insensitive"
)

// OptionalityOrder represents the order of optional members
type OptionalityOrder string

const (
	OptionalFirst OptionalityOrder = "optional-first"
	RequiredFirst OptionalityOrder = "required-first"
)

// MemberType represents a member type or group of member types
type MemberType string

// Options for the member-ordering rule
type Options struct {
	Classes          *OrderConfig `json:"classes,omitempty"`
	ClassExpressions *OrderConfig `json:"classExpressions,omitempty"`
	Default          *OrderConfig `json:"default,omitempty"`
	Interfaces       *OrderConfig `json:"interfaces,omitempty"`
	TypeLiterals     *OrderConfig `json:"typeLiterals,omitempty"`
}

// OrderConfig represents the configuration for member ordering
type OrderConfig struct {
	MemberTypes      interface{}       `json:"memberTypes,omitempty"`
	Order            Order             `json:"order,omitempty"`
	OptionalityOrder *OptionalityOrder `json:"optionalityOrder,omitempty"`
}

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

func parseOptions(options any) *Options {
	opts := &Options{
		Default: &OrderConfig{
			MemberTypes: defaultOrder,
		},
	}

	if options == nil {
		return opts
	}

	// Support ESLint-style array format: [{ ...options }]
	if optArray, isArray := options.([]interface{}); isArray {
		if len(optArray) > 0 {
			if first, ok := optArray[0].(map[string]interface{}); ok {
				options = first
			}
		}
	}

	if optsMap, ok := options.(map[string]interface{}); ok {
		// Parse classes
		if classes, ok := optsMap["classes"]; ok {
			opts.Classes = parseOrderConfig(classes)
		}

		// Parse classExpressions
		if classExpressions, ok := optsMap["classExpressions"]; ok {
			opts.ClassExpressions = parseOrderConfig(classExpressions)
		}

		// Parse default
		if defaultCfg, ok := optsMap["default"]; ok {
			opts.Default = parseOrderConfig(defaultCfg)
		}

		// Parse interfaces
		if interfaces, ok := optsMap["interfaces"]; ok {
			opts.Interfaces = parseOrderConfig(interfaces)
		}

		// Parse typeLiterals
		if typeLiterals, ok := optsMap["typeLiterals"]; ok {
			opts.TypeLiterals = parseOrderConfig(typeLiterals)
		}
	}

	return opts
}

func parseOrderConfig(cfg interface{}) *OrderConfig {
	if cfg == nil {
		return nil
	}

	// Handle "never" string
	if str, ok := cfg.(string); ok && str == "never" {
		return &OrderConfig{
			MemberTypes: "never",
		}
	}

	// Handle array of member types
	if arr, ok := cfg.([]interface{}); ok {
		return &OrderConfig{
			MemberTypes: arr,
		}
	}

	// Handle object config
	if obj, ok := cfg.(map[string]interface{}); ok {
		config := &OrderConfig{}

		if memberTypes, ok := obj["memberTypes"]; ok {
			config.MemberTypes = memberTypes
		}

		if order, ok := obj["order"].(string); ok {
			config.Order = Order(order)
		}

		if optionalityOrder, ok := obj["optionalityOrder"].(string); ok {
			o := OptionalityOrder(optionalityOrder)
			config.OptionalityOrder = &o
		}

		return config
	}

	return nil
}

func getNodeType(node *ast.Node) MemberKind {
	switch node.Kind {
	case ast.KindMethodDeclaration:
		return KindMethod

	case ast.KindMethodSignature:
		return KindMethod

	case ast.KindCallSignature:
		return KindCallSignature

	case ast.KindConstructSignature:
		return KindConstructor

	case ast.KindConstructor:
		return KindConstructor

	case ast.KindPropertyDeclaration:
		prop := node.AsPropertyDeclaration()
		// Check for accessor modifier
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsAccessor) {
			return KindAccessor
		}
		if prop.Initializer != nil && isFunctionExpression(prop.Initializer) {
			return KindMethod
		}
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly) {
			return KindReadonlyField
		}
		return KindField

	case ast.KindPropertySignature:
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly) {
			return KindReadonlyField
		}
		return KindField

	case ast.KindGetAccessor:
		return KindGet

	case ast.KindSetAccessor:
		return KindSet

	case ast.KindIndexSignature:
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly) {
			return KindReadonlySignature
		}
		return KindSignature

	case ast.KindClassStaticBlockDeclaration:
		return KindStaticInit
	}

	return ""
}

func isFunctionExpression(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionExpression ||
		node.Kind == ast.KindArrowFunction
}

func getMemberName(node *ast.Node, sourceFile *ast.SourceFile) string {
	switch node.Kind {
	case ast.KindPropertySignature, ast.KindMethodSignature,
		ast.KindPropertyDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		// For signatures, the node itself is the name node
		var nameNode *ast.Node
		if node.Kind == ast.KindPropertyDeclaration {
			nameNode = node.AsPropertyDeclaration().Name()
		} else if node.Kind == ast.KindGetAccessor {
			nameNode = node.AsGetAccessorDeclaration().Name()
		} else if node.Kind == ast.KindSetAccessor {
			nameNode = node.AsSetAccessorDeclaration().Name()
		} else {
			nameNode = node
		}
		name, _ := utils.GetNameFromMember(sourceFile, nameNode)
		return name

	case ast.KindMethodDeclaration:
		if nameNode := node.AsMethodDeclaration().Name(); nameNode != nil {
			name, _ := utils.GetNameFromMember(sourceFile, nameNode)
			return name
		}
		return ""

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

// getMemberSortName returns a normalized name to be used for sorting comparisons.
// For quoted literal member names (e.g. "b.c"), this strips the surrounding quotes
// so comparisons are based on the actual text content rather than quote characters.
func getMemberSortName(node *ast.Node, sourceFile *ast.SourceFile) string {
	switch node.Kind {
	case ast.KindPropertySignature, ast.KindMethodSignature,
		ast.KindPropertyDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindMethodDeclaration:
		var nameNode *ast.Node
		if node.Kind == ast.KindPropertyDeclaration {
			nameNode = node.AsPropertyDeclaration().Name()
		} else if node.Kind == ast.KindGetAccessor {
			nameNode = node.AsGetAccessorDeclaration().Name()
		} else if node.Kind == ast.KindSetAccessor {
			nameNode = node.AsSetAccessorDeclaration().Name()
		} else if node.Kind == ast.KindMethodDeclaration {
			nameNode = node.AsMethodDeclaration().Name()
		} else {
			nameNode = node
		}
		name, nameType := utils.GetNameFromMember(sourceFile, nameNode)
		if nameType == utils.MemberNameTypeQuoted {
			// Strip surrounding quotes for sort comparisons
			if len(name) >= 2 && ((name[0] == '"' && name[len(name)-1] == '"') || (name[0] == '\'' && name[len(name)-1] == '\'')) {
				return name[1 : len(name)-1]
			}
		}
		return name
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

// getMemberDisplay returns a concise member name for messages (e.g., G, H, I, #I, a, 'b.c', call, new)
func getMemberDisplay(sourceFile *ast.SourceFile, node *ast.Node) string {
	switch node.Kind {
	case ast.KindMethodDeclaration:
		if name := node.AsMethodDeclaration().Name(); name != nil {
			n, _ := utils.GetNameFromMember(sourceFile, name)
			return n
		}
	case ast.KindGetAccessor:
		if name := node.AsGetAccessorDeclaration().Name(); name != nil {
			n, _ := utils.GetNameFromMember(sourceFile, name)
			return n
		}
	case ast.KindSetAccessor:
		if name := node.AsSetAccessorDeclaration().Name(); name != nil {
			n, _ := utils.GetNameFromMember(sourceFile, name)
			return n
		}
	case ast.KindPropertyDeclaration:
		if name := node.AsPropertyDeclaration().Name(); name != nil {
			n, _ := utils.GetNameFromMember(sourceFile, name)
			return n
		}
	case ast.KindPropertySignature, ast.KindMethodSignature:
		// use node directly
		n, _ := utils.GetNameFromMember(sourceFile, node)
		return n
	case ast.KindConstructSignature:
		return "new"
	case ast.KindCallSignature:
		return "call"
	case ast.KindClassStaticBlockDeclaration:
		return "static block"
	}
	return getMemberName(node, sourceFile)
}

// getMemberSnippet returns a compact source snippet for a member's head, used in messages
func getMemberSnippet(sourceFile *ast.SourceFile, node *ast.Node) string {
	r := utils.TrimNodeTextRange(sourceFile, node)
	text := sourceFile.Text()[r.Pos():r.End()]
	// Take only the first line and trim
	end := len(text)
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' || text[i] == '\r' {
			end = i
			break
		}
	}
	head := strings.TrimSpace(text[:end])
	// Avoid trailing bodies for methods: cut after name and parens or property name
	// leave as-is; snapshots show full heads like "public static G() {}" or "a: Foo;"
	return head
}

// getGroupOrderDisplay returns simplified member head text for group order messages
// e.g. "new", "call", "G();", "#I();", "public static a: string;" is simplified to name tokens when needed
func getGroupOrderDisplay(sourceFile *ast.SourceFile, node *ast.Node) string {
	switch node.Kind {
	case ast.KindConstructSignature:
		return "new"
	case ast.KindCallSignature:
		return "call"
	case ast.KindMethodDeclaration:
		name := getMemberName(node, sourceFile)
		if name == "" {
			return getMemberSnippet(sourceFile, node)
		}
		return name + "();"
	default:
		// fall back to concise name for fields/accessors
		n := getMemberName(node, sourceFile)
		if n != "" {
			return n
		}
		return getMemberSnippet(sourceFile, node)
	}
}

func getNameFromIndexSignature(node *ast.Node) string {
	sig := node.AsIndexSignatureDeclaration()
	if sig.Parameters != nil && len(sig.Parameters.Nodes) > 0 {
		param := sig.Parameters.Nodes[0]
		if param != nil && param.Name() != nil {
			if param.Name().Kind == ast.KindIdentifier {
				return param.Name().AsIdentifier().Text
			}
		}
	}
	return ""
}

func isMemberOptional(node *ast.Node) bool {
	return ast.HasQuestionToken(node)
}

func getAccessibility(node *ast.Node) Accessibility {
	// Check for private identifier (#private)
	var name *ast.Node
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		name = node.AsPropertyDeclaration().Name()
	case ast.KindMethodDeclaration:
		name = node.AsMethodDeclaration().Name()
	case ast.KindGetAccessor:
		name = node.AsGetAccessorDeclaration().Name()
	case ast.KindSetAccessor:
		name = node.AsSetAccessorDeclaration().Name()
	}

	if name != nil && name.Kind == ast.KindPrivateIdentifier {
		return AccessPrivateID
	}

	// Check accessibility modifiers
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate) {
		return AccessPrivate
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsProtected) {
		return AccessProtected
	}

	// Default to public
	return AccessPublic
}

func isAbstract(node *ast.Node) bool {
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract)
}

func isStatic(node *ast.Node) bool {
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic)
}

func hasDecorators(node *ast.Node) bool {
	// Check if the node has any decorators using combined modifier flags
	return (ast.GetCombinedModifierFlags(node) & ast.ModifierFlagsDecorator) != 0
}

func getMemberGroups(node *ast.Node, supportsModifiers bool) []string {
	nodeType := getNodeType(node)
	if nodeType == "" {
		return nil
	}

	groups := []string{}

	// Ignore overload-like method declarations without body (treat as transparent for ordering)
	if node.Kind == ast.KindMethodDeclaration {
		method := node.AsMethodDeclaration()
		if method.Body == nil && !isAbstract(node) {
			return groups
		}
	}

	if !supportsModifiers {
		groups = append(groups, string(nodeType))
		switch nt := nodeType; nt {
		case KindReadonlySignature:
			groups = append(groups, string(KindSignature))
		case KindReadonlyField:
			groups = append(groups, string(KindField))
		}
		return groups
	}

	abstract := isAbstract(node)
	static := isStatic(node)
	decorated := hasDecorators(node)
	accessibility := getAccessibility(node)

	scope := ScopeInstance
	if static {
		scope = ScopeStatic
	} else if abstract {
		scope = ScopeAbstract
	}

	// Add decorated member types
	if decorated && (nodeType == KindReadonlyField || nodeType == KindField ||
		nodeType == KindMethod || nodeType == KindAccessor ||
		nodeType == KindGet || nodeType == KindSet) {

		groups = append(groups, fmt.Sprintf("%s-decorated-%s", accessibility, nodeType))
		groups = append(groups, fmt.Sprintf("decorated-%s", nodeType))

		if nodeType == KindReadonlyField {
			groups = append(groups, fmt.Sprintf("%s-decorated-field", accessibility))
			groups = append(groups, "decorated-field")
		}
	}

	// Add scope-based member types
	if nodeType != KindReadonlySignature && nodeType != KindSignature &&
		nodeType != KindStaticInit && nodeType != KindConstructor {

		groups = append(groups, fmt.Sprintf("%s-%s-%s", accessibility, scope, nodeType))
		groups = append(groups, fmt.Sprintf("%s-%s", scope, nodeType))

		if nodeType == KindReadonlyField {
			groups = append(groups, fmt.Sprintf("%s-%s-field", accessibility, scope))
			groups = append(groups, fmt.Sprintf("%s-field", scope))
		}
	}

	// Add accessibility-based member types
	if nodeType != KindReadonlySignature && nodeType != KindSignature &&
		nodeType != KindStaticInit {
		groups = append(groups, fmt.Sprintf("%s-%s", accessibility, nodeType))
		if nodeType == KindReadonlyField {
			groups = append(groups, fmt.Sprintf("%s-field", accessibility))
		}
	}

	// Add base member type
	groups = append(groups, string(nodeType))
	switch nt := nodeType; nt {
	case KindReadonlySignature:
		groups = append(groups, string(KindSignature))
	case KindReadonlyField:
		groups = append(groups, string(KindField))
	}

	return groups
}

func getRank(node *ast.Node, memberTypes []interface{}, supportsModifiers bool) int {
	groups := getMemberGroups(node, supportsModifiers)
	if len(groups) == 0 {
		// No applicable groups (e.g., overload signature without body) -> ignore for ordering
		return -1
	}

	// First, try to find an exact match in member types
	for _, group := range groups {
		for i, memberType := range memberTypes {
			if arr, ok := memberType.([]interface{}); ok {
				// Check if group matches any in the array
				for _, item := range arr {
					if str, ok := item.(string); ok && str == group {
						return i
					}
				}
			} else if str, ok := memberType.(string); ok && str == group {
				return i
			}
		}
	}

	return -1
}

func getLowestRank(ranks []int, target int, order []interface{}) string {
	lowest := ranks[len(ranks)-1]

	for _, rank := range ranks {
		if rank > target && rank < lowest {
			lowest = rank
		}
	}

	lowestRank := order[lowest]
	var lowestRanks []string

	if arr, ok := lowestRank.([]interface{}); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok {
				lowestRanks = append(lowestRanks, str)
			}
		}
	} else if str, ok := lowestRank.(string); ok {
		lowestRanks = []string{str}
	}

	// Replace dashes with spaces
	for i, rank := range lowestRanks {
		lowestRanks[i] = strings.ReplaceAll(rank, "-", " ")
	}

	return strings.Join(lowestRanks, ", ")
}

func naturalCompare(a, b string) int {
	// Simple natural sort implementation
	if a == b {
		return 0
	}

	// For natural ordering, "a1" should come before "a10" which should come before "a2"
	// This is different from lexicographic ordering
	aRunes := []rune(a)
	bRunes := []rune(b)

	i, j := 0, 0
	for i < len(aRunes) && j < len(bRunes) {
		aChar := aRunes[i]
		bChar := bRunes[j]

		// If both are digits, compare numerically
		if aChar >= '0' && aChar <= '9' && bChar >= '0' && bChar <= '9' {
			aNum := 0
			for i < len(aRunes) && aRunes[i] >= '0' && aRunes[i] <= '9' {
				aNum = aNum*10 + int(aRunes[i]-'0')
				i++
			}

			bNum := 0
			for j < len(bRunes) && bRunes[j] >= '0' && bRunes[j] <= '9' {
				bNum = bNum*10 + int(bRunes[j]-'0')
				j++
			}

			if aNum != bNum {
				if aNum < bNum {
					return -1
				}
				return 1
			}
		} else {
			// Compare characters directly
			if aChar != bChar {
				if aChar < bChar {
					return -1
				}
				return 1
			}
			i++
			j++
		}
	}

	// If one string is shorter
	if len(aRunes) < len(bRunes) {
		return -1
	} else if len(aRunes) > len(bRunes) {
		return 1
	}

	return 0
}

func checkGroupSort(ctx rule.RuleContext, members []*ast.Node, groupOrder []interface{}, supportsModifiers bool) [][]*ast.Node {
	var previousRanks []int
	var memberGroups [][]*ast.Node
	// Defer reporting of group-order errors until after alpha-sort pass in caller
	isCorrectlySorted := true

	for _, member := range members {
		rank := getRank(member, groupOrder, supportsModifiers)
		// Use simplified display for group-order messages to match snapshots
		_ = getGroupOrderDisplay(ctx.SourceFile, member)

		if rank == -1 {
			continue
		}

		if len(previousRanks) > 0 {
			rankLastMember := previousRanks[len(previousRanks)-1]
			if rank < rankLastMember {
				// Mark as incorrect, but do not report yet
				isCorrectlySorted = false
			} else if rank == rankLastMember {
				// Same member group - add to existing group
				memberGroups[len(memberGroups)-1] = append(memberGroups[len(memberGroups)-1], member)
			} else {
				// New member group
				previousRanks = append(previousRanks, rank)
				memberGroups = append(memberGroups, []*ast.Node{member})
			}
		} else {
			// First member
			previousRanks = append(previousRanks, rank)
			memberGroups = append(memberGroups, []*ast.Node{member})
		}
	}

	if isCorrectlySorted {
		return memberGroups
	}
	// When group ordering is incorrect, return nil so caller can decide message precedence
	return nil
}

func checkAlphaSort(ctx rule.RuleContext, members []*ast.Node, order Order) bool {
	if len(members) == 0 {
		return true
	}

	previousDisplay := ""
	previousSortName := ""
	isCorrectlySorted := true

	for _, member := range members {
		// Use snippet for messages to match snapshots like "a: Foo;" and "public static G() {}"
		display := getMemberSnippet(ctx.SourceFile, member)
		sortName := getMemberSortName(member, ctx.SourceFile)

		if sortName != "" && previousSortName != "" {
			// For case-insensitive alphabetical/natural orders, compare using lowercase forms
			compareName := sortName
			comparePrev := previousSortName
			switch order {
			case OrderAlphabeticallyCaseInsensitive, OrderNaturalCaseInsensitive:
				compareName = strings.ToLower(compareName)
				comparePrev = strings.ToLower(comparePrev)
			}
			if naturalOutOfOrder(compareName, comparePrev, normalizeOrderForCompare(order)) {
				ctx.ReportNode(member, rule.RuleMessage{
					Id:          "incorrectOrder",
					Description: fmt.Sprintf("Member %s should be declared before member %s.", display, previousDisplay),
				})
				isCorrectlySorted = false
			}
		}

		if sortName != "" {
			previousDisplay = display
			previousSortName = sortName
		}
	}

	return isCorrectlySorted
}

func naturalOutOfOrder(name, previousName string, order Order) bool {
	if name == previousName {
		return false
	}

	switch order {
	case OrderAlphabetically:
		return name < previousName
	case OrderAlphabeticallyCaseInsensitive:
		return strings.ToLower(name) < strings.ToLower(previousName)
	case OrderNatural:
		return naturalCompare(name, previousName) < 0
	case OrderNaturalCaseInsensitive:
		return naturalCompare(strings.ToLower(name), strings.ToLower(previousName)) < 0
	}

	return false
}

// normalizeOrderForCompare maps case-insensitive variants to their base form for comparison
func normalizeOrderForCompare(order Order) Order {
	switch order {
	case OrderAlphabeticallyCaseInsensitive:
		return OrderAlphabetically
	case OrderNaturalCaseInsensitive:
		return OrderNatural
	default:
		return order
	}
}

// isMemberTypesNever returns true if the memberTypes config is set to "never"
func isMemberTypesNever(cfg *OrderConfig) bool {
	if cfg == nil {
		return false
	}
	if str, ok := cfg.MemberTypes.(string); ok && str == "never" {
		return true
	}
	return false
}

func checkRequiredOrder(ctx rule.RuleContext, members []*ast.Node, optionalityOrder OptionalityOrder) bool {
	switchIndex := -1
	for i := 1; i < len(members); i++ {
		if isMemberOptional(members[i]) != isMemberOptional(members[i-1]) {
			switchIndex = i
			break
		}
	}

	if switchIndex == -1 {
		return true
	}

	firstIsOptional := isMemberOptional(members[0])
	expectedFirstOptional := optionalityOrder == OptionalFirst

	if firstIsOptional != expectedFirstOptional {
		reportOptionalityError(ctx, members[0], optionalityOrder)
		return false
	}

	// Check remaining members after switch
	for i := switchIndex + 1; i < len(members); i++ {
		if isMemberOptional(members[i]) != isMemberOptional(members[switchIndex]) {
			reportOptionalityError(ctx, members[switchIndex], optionalityOrder)
			return false
		}
	}

	return true
}

func reportOptionalityError(ctx rule.RuleContext, member *ast.Node, optionalityOrder OptionalityOrder) {
	optionalOrRequired := "optional"
	if optionalityOrder == RequiredFirst {
		optionalOrRequired = "required"
	}

	ctx.ReportNode(member, rule.RuleMessage{
		Id: "incorrectRequiredMembersOrder",
		Description: fmt.Sprintf("Member %s should be declared after all %s members.",
			getMemberName(member, ctx.SourceFile), optionalOrRequired),
	})
}

func validateMembersOrder(ctx rule.RuleContext, members []*ast.Node, orderConfig *OrderConfig, supportsModifiers bool) {
	if orderConfig == nil {
		return
	}

	// If memberTypes is "never", skip group ordering but still apply alpha sort when requested
	if str, ok := orderConfig.MemberTypes.(string); ok && str == "never" {
		if orderConfig.Order != "" && orderConfig.Order != OrderAsWritten {
			memberPtrs := make([]*ast.Node, 0, len(members))
			for _, m := range members {
				// Ignore method declarations without body (overload-like) for sorting
				if m.Kind == ast.KindMethodDeclaration && m.AsMethodDeclaration().Body == nil && !isAbstract(m) {
					continue
				}
				memberPtrs = append(memberPtrs, m)
			}
			_ = checkAlphaSort(ctx, memberPtrs, orderConfig.Order)
		}
		return
	}

	// Parse member types
	var memberTypes []interface{}
	if arr, ok := orderConfig.MemberTypes.([]interface{}); ok {
		memberTypes = arr
	} else {
		// Use default order
		memberTypes = defaultOrder
	}

	// Convert ast.Node slice to pointer slice
	memberPtrs := make([]*ast.Node, len(members))
	copy(memberPtrs, members)

	// Handle optionality order
	if orderConfig.OptionalityOrder != nil {
		switchIndex := -1
		for i := 1; i < len(memberPtrs); i++ {
			if isMemberOptional(memberPtrs[i]) != isMemberOptional(memberPtrs[i-1]) {
				switchIndex = i
				break
			}
		}

		if switchIndex != -1 {
			if !checkRequiredOrder(ctx, memberPtrs, *orderConfig.OptionalityOrder) {
				return
			}

			// Check order for each group separately
			checkOrder(ctx, memberPtrs[:switchIndex], memberTypes, orderConfig.Order, supportsModifiers)
			checkOrder(ctx, memberPtrs[switchIndex:], memberTypes, orderConfig.Order, supportsModifiers)
			return
		}
	}

	// Check order for all members
	checkOrder(ctx, memberPtrs, memberTypes, orderConfig.Order, supportsModifiers)
}

func checkOrder(ctx rule.RuleContext, members []*ast.Node, memberTypes []interface{}, order Order, supportsModifiers bool) {
	hasAlphaSort := order != "" && order != OrderAsWritten
	// Detect when memberTypes is "never" to skip group ordering entirely (alpha-sort only)
	memberTypesNever := len(memberTypes) == 1 && memberTypes[0] == "never"
	var grouped [][]*ast.Node
	if !memberTypesNever {
		grouped = checkGroupSort(ctx, members, memberTypes, supportsModifiers)
	} else {
		grouped = [][]*ast.Node{members}
	}

	if grouped == nil {
		// Group order incorrect; emit group-order diagnostics to match TS-ESLint
		// Walk members and report the first offending member for each out-of-order transition
		prevRank := -1
		for _, member := range members {
			rank := getRank(member, memberTypes, supportsModifiers)
			if rank == -1 {
				continue
			}
			if prevRank != -1 && rank < prevRank {
				ctx.ReportNode(member, rule.RuleMessage{
					Id:          "incorrectGroupOrder",
					Description: fmt.Sprintf("Member %s should be declared before all %s definitions.", getMemberSnippet(ctx.SourceFile, member), getLowestRank([]int{prevRank}, rank, memberTypes)),
				})
				// continue scanning to allow further alpha checks below
			}
			prevRank = rank
		}
		// Still check alpha sort within same-rank groups
		if hasAlphaSort {
			groupMembersByType(members, memberTypes, supportsModifiers, func(group []*ast.Node) {
				checkAlphaSort(ctx, group, order)
			})
		}
		return
	}

	// Check alpha sort within groups
	if hasAlphaSort {
		for _, group := range grouped {
			checkAlphaSort(ctx, group, order)
		}
	}
}

func groupMembersByType(members []*ast.Node, memberTypes []interface{}, supportsModifiers bool, callback func([]*ast.Node)) {
	var groupedMembers [][]*ast.Node
	memberRanks := make([]int, len(members))

	for i, member := range members {
		memberRanks[i] = getRank(member, memberTypes, supportsModifiers)
	}

	previousRank := -2 // Different from any possible rank
	for i, member := range members {
		if i == len(members)-1 {
			continue
		}

		rankOfCurrentMember := memberRanks[i]
		rankOfNextMember := memberRanks[i+1]

		// Group members with same rank
		if rankOfCurrentMember == previousRank { //nolint:staticcheck // False positive: runtime values, not suitable for tagged switch
			// Add to existing group
			groupedMembers[len(groupedMembers)-1] = append(groupedMembers[len(groupedMembers)-1], member)
		} else if rankOfCurrentMember == rankOfNextMember {
			// Start new group
			groupedMembers = append(groupedMembers, []*ast.Node{member})
			previousRank = rankOfCurrentMember
		}
	}

	for _, group := range groupedMembers {
		callback(group)
	}
}

var MemberOrderingRule = rule.Rule{
	Name: "member-ordering",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				class := node.AsClassDeclaration()
				config := opts.Classes
				if config == nil {
					config = opts.Default
				}
				if config != nil {
					members := make([]*ast.Node, len(class.Members.Nodes))
					copy(members, class.Members.Nodes)
					validateMembersOrder(ctx, members, config, true)
				}
			},

			ast.KindClassExpression: func(node *ast.Node) {
				class := node.AsClassExpression()
				config := opts.ClassExpressions
				if config == nil {
					config = opts.Default
				}
				if config != nil {
					members := make([]*ast.Node, len(class.Members.Nodes))
					copy(members, class.Members.Nodes)
					validateMembersOrder(ctx, members, config, true)
				}
			},

			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				iface := node.AsInterfaceDeclaration()
				config := opts.Interfaces
				if config == nil {
					config = opts.Default
				}
				if config != nil {
					members := make([]*ast.Node, len(iface.Members.Nodes))
					copy(members, iface.Members.Nodes)
					validateMembersOrder(ctx, members, config, false)
				}
			},

			ast.KindTypeLiteral: func(node *ast.Node) {
				typeLit := node.AsTypeLiteralNode()
				config := opts.TypeLiterals
				if config == nil {
					config = opts.Default
				}
				if config != nil {
					members := make([]*ast.Node, len(typeLit.Members.Nodes))
					copy(members, typeLit.Members.Nodes)
					validateMembersOrder(ctx, members, config, false)
				}
			},
		}
	},
}
