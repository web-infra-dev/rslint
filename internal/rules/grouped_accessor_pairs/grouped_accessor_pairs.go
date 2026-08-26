package grouped_accessor_pairs

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules/accessorutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed grouped_accessor_pairs.schema.json
var schemaJSON []byte

type Options struct {
	Order             string
	EnforceForTSTypes bool
}

func parseOptions(options []any) Options {
	opts := Options{Order: "anyOrder"}
	if len(options) > 0 {
		if order, ok := options[0].(string); ok {
			opts.Order = order
		}
	}
	if len(options) > 1 {
		if object, ok := options[1].(map[string]any); ok {
			if enforce, ok := object["enforceForTSTypes"].(bool); ok {
				opts.EnforceForTSTypes = enforce
			}
		}
	}
	return opts
}

type accessorGroup struct {
	key     accessorutil.Key
	getters []int
	setters []int
}

func accessorName(node *ast.Node) string {
	return utils.GetFunctionNameWithKindCore(node)
}

func reportPair(ctx rule.RuleContext, messageID string, former *ast.Node, latter *ast.Node) {
	formerName := accessorName(former)
	latterName := accessorName(latter)
	message := fmt.Sprintf("Accessor pair %s and %s should be grouped.", formerName, latterName)
	if messageID == "invalidOrder" {
		message = fmt.Sprintf("Expected %s to be before %s.", latterName, formerName)
	}
	ctx.ReportRange(utils.GetFunctionHeadLoc(ctx.SourceFile, latter), rule.RuleMessage{
		Id:          messageID,
		Description: message,
	})
}

func checkList(ctx rule.RuleContext, members []*ast.Node, opts Options, include func(*ast.Node) bool) {
	groups := make([]accessorGroup, 0)
	for index, member := range members {
		if !include(member) || (member.Kind != ast.KindGetAccessor && member.Kind != ast.KindSetAccessor) {
			continue
		}
		key := accessorutil.MakeKey(member)
		groupIndex := -1
		for index := range groups {
			if accessorutil.KeysEqual(ctx.SourceFile, groups[index].key, key) {
				groupIndex = index
				break
			}
		}
		if groupIndex < 0 {
			groups = append(groups, accessorGroup{key: key})
			groupIndex = len(groups) - 1
		}
		if member.Kind == ast.KindGetAccessor {
			groups[groupIndex].getters = append(groups[groupIndex].getters, index)
		} else {
			groups[groupIndex].setters = append(groups[groupIndex].setters, index)
		}
	}

	for _, group := range groups {
		if len(group.getters) != 1 || len(group.setters) != 1 {
			continue
		}
		getterIndex := group.getters[0]
		setterIndex := group.setters[0]
		formerIndex, latterIndex := getterIndex, setterIndex
		if setterIndex < getterIndex {
			formerIndex, latterIndex = setterIndex, getterIndex
		}
		if latterIndex-formerIndex > 1 {
			reportPair(ctx, "notGrouped", members[formerIndex], members[latterIndex])
			continue
		}
		if (opts.Order == "getBeforeSet" && getterIndex > setterIndex) ||
			(opts.Order == "setBeforeGet" && setterIndex > getterIndex) {
			reportPair(ctx, "invalidOrder", members[formerIndex], members[latterIndex])
		}
	}
}

func concreteAccessor(member *ast.Node) bool {
	return (member.Kind == ast.KindGetAccessor || member.Kind == ast.KindSetAccessor) &&
		!ast.HasSyntacticModifier(member, ast.ModifierFlagsAbstract)
}

func typeAccessor(member *ast.Node) bool {
	return member.Kind == ast.KindGetAccessor || member.Kind == ast.KindSetAccessor
}

var GroupedAccessorPairsRule = rule.Rule{
	Name:   "grouped-accessor-pairs",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		listeners := rule.RuleListeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				object := node.AsObjectLiteralExpression()
				if object != nil && object.Properties != nil {
					checkList(ctx, object.Properties.Nodes, opts, typeAccessor)
				}
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				members := node.Members()
				checkList(ctx, members, opts, func(member *ast.Node) bool {
					return concreteAccessor(member) && !ast.IsStatic(member)
				})
				checkList(ctx, members, opts, func(member *ast.Node) bool {
					return concreteAccessor(member) && ast.IsStatic(member)
				})
			},
		}
		listeners[ast.KindClassExpression] = listeners[ast.KindClassDeclaration]
		if opts.EnforceForTSTypes {
			listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
				checkList(ctx, node.Members(), opts, typeAccessor)
			}
			listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
				checkList(ctx, node.Members(), opts, typeAccessor)
			}
		}
		return listeners
	},
}
