// Package jsx_sort_props implements react/jsx-sort-props.
package jsx_sort_props

import (
	_ "embed"
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

//go:embed jsx_sort_props.schema.json
var schemaJSON []byte

var reservedProps = []string{"children", "dangerouslySetInnerHTML", "key", "ref"}

// JsxSortPropsRule enforces the configured ordering of JSX attributes.
var JsxSortPropsRule = rule.Rule{
	Name:   "react/jsx-sort-props",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		reported := map[*ast.Node]map[string]bool{}

		check := func(element *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(element)
			if len(attrs) == 0 {
				return
			}
			type pendingReport struct {
				attr           *ast.Node
				id             string
				description    string
				data           map[string]string
				wholeAttribute bool
			}
			var pending []pendingReport
			reserved := opts.reservedList
			if opts.reservedFirst && !reactutil.IsDOMComponent(element) {
				reserved = slices.DeleteFunc(slices.Clone(reserved), func(name string) bool {
					return name == "dangerouslySetInnerHTML"
				})
			}

			reportAttr := func(attr *ast.Node, id, description string, data map[string]string, wholeAttribute bool) {
				if attr == nil || attr.Kind != ast.KindJsxAttribute {
					return
				}
				ids := reported[attr]
				if ids == nil {
					ids = map[string]bool{}
					reported[attr] = ids
				}
				if ids[id] {
					return
				}
				ids[id] = true
				pending = append(pending, pendingReport{attr, id, description, data, wholeAttribute})
			}

			memo := attrs[0]
			for index, current := range attrs {
				if index == 0 {
					if opts.reservedFirst && opts.reservedError != "" && current.Kind == ast.KindJsxAttribute {
						reportAttr(current, opts.reservedError, opts.reservedDescription, opts.reservedData, true)
					}
					continue
				}
				if current.Kind == ast.KindJsxSpreadAttribute {
					if index+1 < len(attrs) {
						memo = attrs[index+1]
					} else {
						memo = nil
					}
					continue
				}
				if memo == nil || memo.Kind != ast.KindJsxAttribute || current.Kind != ast.KindJsxAttribute {
					memo = current
					continue
				}

				previousName := reactutil.GetJsxPropName(memo)
				currentName := reactutil.GetJsxPropName(current)
				previousValue := memo.AsJsxAttribute().Initializer != nil
				currentValue := current.AsJsxAttribute().Initializer != nil
				previousCallback := isCallback(previousName)
				currentCallback := isCallback(currentName)
				if opts.ignoreCase {
					previousName = ecmascript.StringToLowerCase(previousName)
					currentName = ecmascript.StringToLowerCase(currentName)
				}

				if opts.reservedFirst {
					if opts.reservedError != "" {
						reportAttr(current, opts.reservedError, opts.reservedDescription, opts.reservedData, true)
						continue
					}
					previousReserved := contains(reserved, previousName)
					currentReserved := contains(reserved, currentName)
					if previousReserved && !currentReserved {
						memo = current
						continue
					}
					if !previousReserved && currentReserved {
						reportAttr(current, "listReservedPropsFirst", "Reserved props must be listed before all other props", nil, false)
						continue
					}
				}
				if opts.callbacksLast {
					if !previousCallback && currentCallback {
						memo = current
						continue
					}
					if previousCallback && !currentCallback {
						reportAttr(memo, "listCallbacksLast", "Callbacks must be listed after all other props", nil, false)
						continue
					}
				}
				if opts.shorthandFirst {
					if currentValue && !previousValue {
						memo = current
						continue
					}
					if !currentValue && previousValue {
						reportAttr(current, "listShorthandFirst", "Shorthand props must be listed before all other props", nil, false)
						continue
					}
				}
				if opts.shorthandLast {
					if !currentValue && previousValue {
						memo = current
						continue
					}
					if currentValue && !previousValue {
						reportAttr(memo, "listShorthandLast", "Shorthand props must be listed after all other props", nil, false)
						continue
					}
				}
				previousMultiline := isMultiline(ctx, memo)
				currentMultiline := isMultiline(ctx, current)
				if opts.multiline == "first" {
					if previousMultiline && !currentMultiline {
						memo = current
						continue
					}
					if !previousMultiline && currentMultiline {
						reportAttr(current, "listMultilineFirst", "Multiline props must be listed before all other props", nil, false)
						continue
					}
				} else if opts.multiline == "last" {
					if !previousMultiline && currentMultiline {
						memo = current
						continue
					}
					if previousMultiline && !currentMultiline {
						reportAttr(memo, "listMultilineLast", "Multiline props must be listed after all other props", nil, false)
						continue
					}
				}
				if !opts.noSortAlphabetically && compareNames(previousName, currentName, opts) > 0 {
					reportAttr(current, "sortPropsByAlpha", "Props should be sorted alphabetically", nil, false)
					continue
				}
				memo = current
			}
			sort.SliceStable(pending, func(i, j int) bool {
				left := utils.TrimNodeTextRange(ctx.SourceFile, pending[i].attr)
				right := utils.TrimNodeTextRange(ctx.SourceFile, pending[j].attr)
				return left.Pos() < right.Pos()
		})
		for _, report := range pending {
			reportNode := report.attr.AsJsxAttribute().Name()
				if report.wholeAttribute {
					reportNode = report.attr
				}
				ctx.ReportNodeWithDeferredFixes(reportNode, rule.RuleMessage{Id: report.id, Description: report.description, Data: report.data}, func() []rule.RuleFix {
					return buildFixes(ctx, element, opts, reserved)
				})
			}
		}

		return rule.RuleListeners{ast.KindJsxOpeningElement: check, ast.KindJsxSelfClosingElement: check}
	},
}

type options struct {
	callbacksLast, shorthandFirst, shorthandLast, ignoreCase, noSortAlphabetically, reservedFirst bool
	multiline, locale                                                                             string
	reservedList                                                                                  []string
	reservedError, reservedDescription                                                            string
	reservedData                                                                                  map[string]string
}

func parseOptions(raw []any) options {
	opts := options{multiline: "ignore", locale: "auto", reservedList: reservedProps}
	if len(raw) == 0 {
		return opts
	}
	config, _ := raw[0].(map[string]any)
	if config == nil {
		return opts
	}
	opts.callbacksLast, _ = config["callbacksLast"].(bool)
	opts.shorthandFirst, _ = config["shorthandFirst"].(bool)
	opts.shorthandLast, _ = config["shorthandLast"].(bool)
	opts.ignoreCase, _ = config["ignoreCase"].(bool)
	opts.noSortAlphabetically, _ = config["noSortAlphabetically"].(bool)
	if value, ok := config["multiline"].(string); ok {
		opts.multiline = value
	}
	if value, ok := config["locale"].(string); ok {
		opts.locale = value
	}
	switch value := config["reservedFirst"].(type) {
	case bool:
		opts.reservedFirst = value
	case []any:
		opts.reservedFirst = true
		opts.reservedList = make([]string, 0, len(value))
		invalid := make([]string, 0)
		for _, item := range value {
			name, _ := item.(string)
			opts.reservedList = append(opts.reservedList, name)
			if !contains(reservedProps, name) {
				invalid = append(invalid, name)
			}
		}
		if len(value) == 0 {
			opts.reservedError = "listIsEmpty"
			opts.reservedDescription = "A customized reserved first list must not be empty"
		} else if len(invalid) != 0 {
			opts.reservedError = "noUnreservedProps"
			opts.reservedDescription = "A customized reserved first list must only contain a subset of React reserved props. Remove: " + joinComma(invalid)
			opts.reservedData = map[string]string{"unreservedWords": joinComma(invalid)}
		}
	}
	return opts
}

func joinComma(values []string) string {
	return strings.Join(values, ",")
}

func contains(values []string, target string) bool { return slices.Contains(values, target) }
func isCallback(name string) bool {
	return len(name) >= 3 && name[0] == 'o' && name[1] == 'n' && name[2] >= 'A' && name[2] <= 'Z'
}

func isMultiline(ctx rule.RuleContext, node *ast.Node) bool {
	r := utils.TrimNodeTextRange(ctx.SourceFile, node)
	lines := ctx.SourceFile.ECMALineMap()
	return scanner.ComputeLineOfPosition(lines, r.Pos()) != scanner.ComputeLineOfPosition(lines, r.End())
}

func compareNames(left, right string, opts options) int {
	if opts.ignoreCase || opts.locale != "auto" {
		return collate.New(language.Make(opts.locale)).CompareString(left, right)
	}
	return ecmascript.CompareStrings(left, right)
}

type sortableAttribute struct {
	node       *ast.Node
	start, end int
	pinned     bool
}

func buildFixes(ctx rule.RuleContext, element *ast.Node, opts options, reserved []string) []rule.RuleFix {
	attrs := reactutil.GetJsxElementAttributes(element)
	groups := sortableGroups(ctx, element, attrs)
	text := ctx.SourceFile.Text()
	var fixes []rule.RuleFix
	for _, group := range groups {
		sorted := slices.Clone(group)
		sort.SliceStable(sorted, func(i, j int) bool { return compareAttributes(ctx, sorted[i], sorted[j], opts, reserved) < 0 })
		for index, target := range group {
			source := sorted[index]
			if target.start == source.start && target.end == source.end {
				continue
			}
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(target.start, target.end), text[source.start:source.end]))
		}
	}
	return fixes
}

func compareAttributes(ctx rule.RuleContext, left, right sortableAttribute, opts options, reserved []string) int {
	if left.pinned != right.pinned {
		if left.pinned {
			return 1
		}
		return -1
	}
	leftName, rightName := reactutil.GetJsxPropName(left.node), reactutil.GetJsxPropName(right.node)
	if opts.reservedFirst {
		if contains(reserved, leftName) != contains(reserved, rightName) {
			if contains(reserved, leftName) {
				return -1
			}
			return 1
		}
	}
	if opts.callbacksLast {
		if isCallback(leftName) != isCallback(rightName) {
			if isCallback(leftName) {
				return 1
			}
			return -1
		}
	}
	leftValue := left.node.AsJsxAttribute().Initializer != nil
	rightValue := right.node.AsJsxAttribute().Initializer != nil
	if opts.shorthandFirst || opts.shorthandLast {
		if leftValue != rightValue {
			if opts.shorthandFirst {
				if !leftValue {
					return -1
				}
				return 1
			}
			if !leftValue {
				return 1
			}
			return -1
		}
	}
	if opts.multiline != "ignore" {
		leftMulti, rightMulti := isMultiline(ctx, left.node), isMultiline(ctx, right.node)
		if leftMulti != rightMulti {
			if opts.multiline == "first" {
				if leftMulti {
					return -1
				}
				return 1
			}
			if leftMulti {
				return 1
			}
			return -1
		}
	}
	if opts.noSortAlphabetically {
		return 0
	}
	if opts.ignoreCase {
		leftName, rightName = ecmascript.StringToLowerCase(leftName), ecmascript.StringToLowerCase(rightName)
	}
	return compareNames(leftName, rightName, opts)
}

func sortableGroups(ctx rule.RuleContext, element *ast.Node, attrs []*ast.Node) [][]sortableAttribute {
	var groups [][]sortableAttribute
	var group []sortableAttribute
	flush := func() {
		if len(group) > 0 {
			groups = append(groups, group)
			group = nil
		}
	}
	comments := ctx.Comments.All()
	for index := 0; index < len(attrs); index++ {
		attr := attrs[index]
		if attr.Kind == ast.KindJsxSpreadAttribute {
			flush()
			continue
		}
		if attr.Kind != ast.KindJsxAttribute {
			continue
		}
		r := utils.TrimNodeTextRange(ctx.SourceFile, attr)
		item := sortableAttribute{node: attr, start: r.Pos(), end: r.End()}
		var next *ast.Node
		if index+1 < len(attrs) {
			next = attrs[index+1]
		}
		nextIsAttribute := next != nil && next.Kind == ast.KindJsxAttribute
		limit := utils.TrimNodeTextRange(ctx.SourceFile, element).End()
		if next != nil {
			limit = utils.TrimNodeTextRange(ctx.SourceFile, next).Pos()
		}
		between := commentsInSpan(comments, r.End(), limit)
		line := scanner.ComputeLineOfPosition(ctx.SourceFile.ECMALineMap(), r.Pos())
		if len(between) == 1 {
			commentLine := scanner.ComputeLineOfPosition(ctx.SourceFile.ECMALineMap(), between[0].Pos())
			if nextIsAttribute && line+1 == commentLine {
				item.end = utils.TrimNodeTextRange(ctx.SourceFile, next).End()
				item.pinned = true
				index++
			} else if line == commentLine {
				if between[0].Kind == ast.KindMultiLineCommentTrivia && nextIsAttribute {
					item.end = utils.TrimNodeTextRange(ctx.SourceFile, next).End()
					item.pinned = true
					index++
				} else {
					item.end = between[0].End()
					item.pinned = between[0].Kind == ast.KindMultiLineCommentTrivia
				}
			} else {
				continue
			}
		} else if len(between) > 1 {
			if !nextIsAttribute || line+1 != scanner.ComputeLineOfPosition(ctx.SourceFile.ECMALineMap(), between[1].Pos()) {
				continue
			}
			nextRange := utils.TrimNodeTextRange(ctx.SourceFile, next)
			item.end = nextRange.End()
			afterNextLimit := utils.TrimNodeTextRange(ctx.SourceFile, element).End()
			if index+2 < len(attrs) {
				afterNextLimit = utils.TrimNodeTextRange(ctx.SourceFile, attrs[index+2]).Pos()
			}
			nextComments := commentsInSpan(comments, item.end, afterNextLimit)
			if len(nextComments) == 1 && scanner.ComputeLineOfPosition(ctx.SourceFile.ECMALineMap(), nextRange.Pos()) == scanner.ComputeLineOfPosition(ctx.SourceFile.ECMALineMap(), nextComments[0].Pos()) {
				item.end = nextComments[0].End()
			}
			item.pinned = true
			index++
		}
		group = append(group, item)
	}
	flush()
	return groups
}

func commentsInSpan(comments []*ast.CommentRange, start, end int) []*ast.CommentRange {
	index := sort.Search(len(comments), func(i int) bool { return comments[i].Pos() >= start })
	var result []*ast.CommentRange
	for ; index < len(comments) && comments[index].Pos() < end; index++ {
		result = append(result, comments[index])
	}
	return result
}
