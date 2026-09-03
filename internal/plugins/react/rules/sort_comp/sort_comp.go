// Package sort_comp implements react/sort-comp.
package sort_comp

import (
	_ "embed"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed sort_comp.schema.json
var schemaJSON []byte

var defaultOrder = []string{"static-methods", "lifecycle", "everything-else", "render"}

var defaultLifecycle = []string{
	"displayName", "propTypes", "contextTypes", "childContextTypes", "mixins", "statics",
	"defaultProps", "constructor", "getDefaultProps", "state", "getInitialState", "getChildContext",
	"getDerivedStateFromProps", "componentWillMount", "UNSAFE_componentWillMount", "componentDidMount",
	"componentWillReceiveProps", "UNSAFE_componentWillReceiveProps", "shouldComponentUpdate",
	"componentWillUpdate", "UNSAFE_componentWillUpdate", "getSnapshotBeforeUpdate", "componentDidUpdate",
	"componentDidCatch", "componentWillUnmount",
}

type options struct {
	order  []string
	groups map[string][]string
}

func parseOptions(raw []any) options {
	opts := options{order: slices.Clone(defaultOrder), groups: map[string][]string{"lifecycle": slices.Clone(defaultLifecycle)}}
	if len(raw) == 0 {
		return opts
	}
	config, _ := raw[0].(map[string]any)
	if config == nil {
		return opts
	}
	if value, ok := config["order"].([]any); ok {
		opts.order = stringSlice(value)
	}
	if value, ok := config["groups"].(map[string]any); ok {
		for name, entries := range value {
			if values, ok := entries.([]any); ok {
				opts.groups[name] = stringSlice(values)
			}
		}
	}
	return opts
}

func stringSlice(values []any) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

type memberInfo struct {
	node             *ast.Node
	name             string
	displayName      string
	getter           bool
	setter           bool
	staticVariable   bool
	staticMethod     bool
	instanceVariable bool
	instanceMethod   bool
	typeAnnotation   bool
}

func isFunctionLikeExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction
}

func isConcreteMethod(member *ast.Node) bool {
	if member == nil {
		return false
	}
	switch member.Kind {
	case ast.KindMethodDeclaration:
		return member.AsMethodDeclaration().Body != nil
	case ast.KindGetAccessor:
		return member.AsGetAccessorDeclaration().Body != nil
	case ast.KindSetAccessor:
		return member.AsSetAccessorDeclaration().Body != nil
	default:
		return false
	}
}

func memberValue(member *ast.Node) *ast.Node {
	if member == nil {
		return nil
	}
	switch member.Kind {
	case ast.KindPropertyAssignment:
		return member.AsPropertyAssignment().Initializer
	case ast.KindPropertyDeclaration:
		return member.AsPropertyDeclaration().Initializer
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		// ESTree represents these as MethodDefinitions whose value is a
		// FunctionExpression. In tsgo the member itself is that function.
		return member
	default:
		return nil
	}
}

func getMembers(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return utils.ESTreeMembers(node.Members())
	case ast.KindObjectLiteralExpression:
		object := node.AsObjectLiteralExpression()
		if object == nil || object.Properties == nil {
			return nil
		}
		return object.Properties.Nodes
	default:
		return nil
	}
}

func getMemberNames(member *ast.Node) (string, string) {
	if member == nil {
		return "", ""
	}
	if member.Kind == ast.KindGetAccessor {
		return "getter functions", "getter functions"
	}
	if member.Kind == ast.KindSetAccessor {
		return "setter functions", "setter functions"
	}
	if member.Kind == ast.KindConstructor {
		return "constructor", "constructor"
	}
	name := member.Name()
	if name == nil {
		return "", ""
	}
	if name.Kind == ast.KindComputedPropertyName {
		name = ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	}
	if name == nil {
		return "undefined", "undefined"
	}
	switch name.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		value := reactutil.IdentifierOrPrivateName(name)
		return value, value
	default:
		// ESTree's property-name lookup has no name for literal and
		// computed expression keys; JavaScript interpolation renders it
		// as "undefined".
		return "undefined", "undefined"
	}
}

func getMemberInfo(member *ast.Node) memberInfo {
	info := memberInfo{node: member}
	if member == nil {
		return info
	}
	info.name, info.displayName = getMemberNames(member)
	plainClassMember := utils.IsPlainClassMember(member)
	// Upstream exposes abstract accessors as TSAbstractMethodDefinition nodes,
	// but they still retain the getter/setter kind. Keep these categories
	// independent from the plain-member check below, which only controls the
	// static/instance variable and method groups.
	info.getter = member.Kind == ast.KindGetAccessor
	info.setter = member.Kind == ast.KindSetAccessor
	value := memberValue(member)
	propertyDeclaration := member.Kind == ast.KindPropertyDeclaration
	// tsgo represents abstract properties and auto-accessors as
	// PropertyDeclaration, but typescript-eslint exposes them as
	// TSAbstractPropertyDefinition / AccessorProperty. Only plain class
	// fields participate in the static/instance variable or method groups.
	classField := propertyDeclaration && plainClassMember
	static := ast.IsStatic(member)
	functionValue := isFunctionLikeExpression(value) || isConcreteMethod(member) || member.Kind == ast.KindConstructor
	if propertyDeclaration && member.AsPropertyDeclaration().Type != nil && value == nil {
		info.typeAnnotation = true
	}
	if classField {
		if static {
			info.staticVariable = value == nil || !functionValue
			info.staticMethod = value != nil && functionValue
		} else {
			info.instanceVariable = value == nil || !functionValue
			info.instanceMethod = value != nil && functionValue
		}
	} else if plainClassMember && static && isConcreteMethod(member) {
		info.staticMethod = true
	}
	return info
}

func expandedOrder(opts options) []string {
	result := make([]string, 0, len(opts.order))
	for _, entry := range opts.order {
		if group, ok := opts.groups[entry]; ok {
			result = append(result, group...)
		} else {
			result = append(result, entry)
		}
	}
	return result
}

var namedGroups = map[string]bool{
	"displayName": true, "propTypes": true, "contextTypes": true, "childContextTypes": true,
	"mixins": true, "statics": true, "defaultProps": true, "constructor": true,
	"getDefaultProps": true, "state": true, "getInitialState": true, "getChildContext": true,
	"getDerivedStateFromProps": true, "componentWillMount": true, "UNSAFE_componentWillMount": true,
	"componentDidMount": true, "componentWillReceiveProps": true, "UNSAFE_componentWillReceiveProps": true,
	"shouldComponentUpdate": true, "componentWillUpdate": true, "UNSAFE_componentWillUpdate": true,
	"getSnapshotBeforeUpdate": true, "componentDidUpdate": true, "componentDidCatch": true,
	"componentWillUnmount": true, "render": true,
}

type pattern struct {
	text string
	re   *esregexp.RegExp
}

func parsePattern(text string) pattern {
	first := strings.IndexByte(text, '/')
	last := strings.LastIndexByte(text, '/')
	if first < 0 || first == last || last == 0 {
		return pattern{text: text}
	}
	flags := text[last+1:]
	flagEnd := 0
	for flagEnd < len(flags) && strings.ContainsRune("gimsuy", rune(flags[flagEnd])) {
		flagEnd++
	}
	flags = flags[:flagEnd]
	re, err := esregexp.Compile(text[first+1:last], flags)
	if err != nil {
		return pattern{text: text}
	}
	return pattern{re: re}
}

func referenceIndexes(info memberInfo, order []pattern) []int {
	indexes := make([]int, 0, 1)
	for index, entry := range order {
		switch entry.text {
		case "getters":
			if info.getter {
				indexes = append(indexes, index)
			}
		case "setters":
			if info.setter {
				indexes = append(indexes, index)
			}
		case "type-annotations":
			if info.typeAnnotation {
				indexes = append(indexes, index)
			}
		case "static-variables":
			if info.staticVariable {
				indexes = append(indexes, index)
			}
		case "static-methods":
			if info.staticMethod {
				indexes = append(indexes, index)
			}
		case "instance-variables":
			if info.instanceVariable {
				indexes = append(indexes, index)
			}
		case "instance-methods":
			if info.instanceMethod {
				indexes = append(indexes, index)
			}
		default:
			if namedGroups[entry.text] {
				if entry.text == info.name {
					indexes = append(indexes, index)
				}
			} else if entry.re != nil && entry.re.Test(info.name) {
				indexes = append(indexes, index)
			} else if entry.re == nil && entry.text == info.name {
				indexes = append(indexes, index)
			}
		}
	}
	if len(indexes) == 0 {
		for index, entry := range order {
			if entry.text == "everything-else" {
				return []int{index}
			}
		}
		return []int{math.MaxInt}
	}
	return indexes
}

type comparison struct {
	correct bool
	indexA  int
	indexB  int
}

func compareProps(infos []memberInfo, a, b memberInfo, order []pattern) comparison {
	aIndexes := referenceIndexes(a, order)
	bIndexes := referenceIndexes(b, order)
	classIndexA := indexOfMember(infos, a.node)
	classIndexB := indexOfMember(infos, b.node)
	lastA, lastB := 0, 0
	for _, indexA := range aIndexes {
		lastA = indexA
		for _, indexB := range bIndexes {
			lastB = indexB
			if indexA == indexB || (indexA < indexB && classIndexA < classIndexB) || (indexA > indexB && classIndexA > classIndexB) {
				return comparison{correct: true, indexA: classIndexA, indexB: classIndexB}
			}
		}
	}
	return comparison{indexA: lastA, indexB: lastB}
}

func indexOfMember(infos []memberInfo, node *ast.Node) int {
	for index, info := range infos {
		if info.node == node {
			return index
		}
	}
	return -1
}

func referenceDistance(indexA, indexB int) int {
	// math.MaxInt represents JavaScript's Infinity for an unmatched member.
	// Keep that distance infinite instead of making it appear closer to one
	// finite group than another.
	if indexA == math.MaxInt || indexB == math.MaxInt {
		return math.MaxInt
	}
	distance := indexA - indexB
	if distance < 0 {
		return -distance
	}
	return distance
}

type storedError struct {
	node     *ast.Node
	key      int
	score    int
	distance int
	refNode  *ast.Node
	refIndex int
}

func propertyName(info memberInfo) string {
	if info.getter {
		return "getter functions"
	}
	if info.setter {
		return "setter functions"
	}
	return info.displayName
}

func isES5ComponentObject(node *ast.Node, pragma, createClass string) bool {
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil {
		return false
	}
	var callee *ast.Node
	var args *ast.NodeList
	switch parent.Kind {
	case ast.KindCallExpression:
		call := parent.AsCallExpression()
		callee, args = call.Expression, call.Arguments
	case ast.KindNewExpression:
		newExpr := parent.AsNewExpression()
		callee, args = newExpr.Expression, newExpr.Arguments
	default:
		return false
	}
	if args == nil {
		return false
	}
	inArgs := false
	for _, arg := range args.Nodes {
		if arg == current {
			inArgs = true
			break
		}
	}
	if !inArgs {
		return false
	}
	if parent.Kind == ast.KindCallExpression {
		return reactutil.IsCreateClassCall(parent.AsCallExpression(), pragma, createClass)
	}
	callee = ast.SkipParentheses(callee)
	if callee.Kind == ast.KindIdentifier {
		return callee.AsIdentifier().Text == createClass
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	object := ast.SkipParentheses(access.Expression)
	name := access.Name()
	return object.Kind == ast.KindIdentifier && object.AsIdentifier().Text == pragma && name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == createClass
}

func checkComponent(ctx rule.RuleContext, node *ast.Node, order []pattern, errors map[int]*storedError) {
	members := getMembers(node)
	infos := make([]memberInfo, len(members))
	for index, member := range members {
		infos[index] = getMemberInfo(member)
	}
	for i, a := range infos {
		for k, b := range infos {
			if i == k {
				continue
			}
			result := compareProps(infos, a, b, order)
			if result.correct {
				continue
			}
			stored := errors[result.indexA]
			if stored == nil {
				stored = &storedError{node: a.node, key: result.indexA, distance: math.MaxInt, refIndex: 0}
				errors[result.indexA] = stored
			}
			stored.score++
			if propertyName(getMemberInfo(stored.node)) != propertyName(a) {
				continue
			}
			distance := referenceDistance(result.indexA, result.indexB)
			if distance > stored.distance {
				continue
			}
			stored.distance, stored.refNode, stored.refIndex = distance, b.node, result.indexB
		}
	}
}

func reportErrors(ctx rule.RuleContext, errors map[int]*storedError) {
	keys := make([]int, 0, len(errors))
	for key := range errors {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	for _, key := range keys {
		stored := errors[key]
		if stored == nil {
			continue
		}
		if other := errors[stored.refIndex]; other != nil {
			if stored.score > other.score {
				delete(errors, stored.refIndex)
			} else {
				delete(errors, stored.key)
			}
		}
	}
	ordered := make([]*storedError, 0, len(errors))
	for _, error := range errors {
		ordered = append(ordered, error)
	}
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].node.Pos() < ordered[j].node.Pos() })
	for _, error := range ordered {
		if error.node == nil || error.refNode == nil {
			continue
		}
		a, b := getMemberInfo(error.node), getMemberInfo(error.refNode)
		position := "after"
		if error.key < error.refIndex {
			position = "before"
		}
		ctx.ReportNode(error.node, rule.RuleMessage{Id: "unsortedProps", Description: propertyName(a) + " should be placed " + position + " " + propertyName(b), Data: map[string]string{"propA": propertyName(a), "propB": propertyName(b), "position": position}})
	}
}

var SortCompRule = rule.Rule{
	Name:   "react/sort-comp",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		expanded := expandedOrder(opts)
		order := make([]pattern, len(expanded))
		for index, entry := range expanded {
			order[index] = parsePattern(entry)
		}
		pragma := reactutil.GetReactPragmaFromContext(ctx)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		errors := map[int]*storedError{}
		checkClass := func(node *ast.Node) {
			if reactutil.ExtendsReactComponent(node, pragma) || reactutil.IsExplicitReactComponent(node) {
				checkComponent(ctx, node, order, errors)
			}
		}
		return rule.RuleListeners{
			ast.KindClassDeclaration: checkClass,
			ast.KindClassExpression:  checkClass,
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if isES5ComponentObject(node, pragma, createClass) {
					checkComponent(ctx, node, order, errors)
				}
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(node *ast.Node) { reportErrors(ctx, errors) },
		}
	},
}
