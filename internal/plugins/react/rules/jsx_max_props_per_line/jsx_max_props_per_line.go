package jsx_max_props_per_line

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JsxMaxPropsPerLineRule limits the maximum number of props on a single line in JSX.
var JsxMaxPropsPerLineRule = rule.Rule{
	Name: "react/jsx-max-props-per-line",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Defaults
		singleLimit := 0 // 0 means "not explicitly set"
		multiLimit := 0
		when := "always"
		maximumSet := false
		maximumIsObject := false

		// Parse options
		optsMap := utils.GetOptionsMap(options)
		if optsMap != nil {
			if w, ok := optsMap["when"].(string); ok {
				when = w
			}
			if maxVal, ok := optsMap["maximum"]; ok {
				maximumSet = true
				switch v := maxVal.(type) {
				case float64:
					singleLimit = int(v)
					multiLimit = int(v)
				case map[string]interface{}:
					maximumIsObject = true
					if s, ok := v["single"].(float64); ok {
						singleLimit = int(s)
					}
					if m, ok := v["multi"].(float64); ok {
						multiLimit = int(m)
					}
				}
			}
		}

		// Apply defaults if not explicitly set
		if !maximumSet {
			singleLimit = 1
			multiLimit = 1
		}

		// When maximum is a number and when="multiline", single-line tags have no limit
		if !maximumIsObject && when == "multiline" {
			singleLimit = 0
		}

		check := func(node *ast.Node) {
			var props []*ast.Node

			switch node.Kind {
			case ast.KindJsxOpeningElement:
				attrs := node.AsJsxOpeningElement().Attributes.AsJsxAttributes()
				if attrs.Properties != nil {
					props = attrs.Properties.Nodes
				}
			case ast.KindJsxSelfClosingElement:
				attrs := node.AsJsxSelfClosingElement().Attributes.AsJsxAttributes()
				if attrs.Properties != nil {
					props = attrs.Properties.Nodes
				}
			}

			if len(props) == 0 {
				return
			}

			lineMap := ctx.SourceFile.ECMALineMap()

			// Determine if the tag is single-line or multi-line
			openingTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
			firstLine := scanner.ComputeLineOfPosition(lineMap, openingTrimmed.Pos())
			lastLine := scanner.ComputeLineOfPosition(lineMap, openingTrimmed.End())
			isSingleLine := firstLine == lastLine

			var limit int
			if isSingleLine {
				limit = singleLimit
			} else {
				limit = multiLimit
			}

			// If limit is 0, no restriction
			if limit == 0 {
				return
			}

			// When "multiline", only check multi-line elements (only applies when maximum is a number, not an object)
			if when == "multiline" && !maximumIsObject && isSingleLine {
				return
			}

			// Group props by line — ESLint chains by checking if the previous prop's
			// END line equals the current prop's START line (handles multi-line props).
			type lineGroup struct {
				endLine int
				props   []*ast.Node
			}

			var groups []lineGroup
			for _, prop := range props {
				propTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, prop)
				propStartLine := scanner.ComputeLineOfPosition(lineMap, propTrimmed.Pos())
				propEndLine := scanner.ComputeLineOfPosition(lineMap, propTrimmed.End())
				if len(groups) == 0 || groups[len(groups)-1].endLine != propStartLine {
					groups = append(groups, lineGroup{endLine: propEndLine, props: []*ast.Node{prop}})
				} else {
					groups[len(groups)-1].endLine = propEndLine
					groups[len(groups)-1].props = append(groups[len(groups)-1].props, prop)
				}
			}

			// Check each group — report one error per line at the first excess prop
			for _, group := range groups {
				if len(group.props) > limit {
					prop := group.props[limit]
					propName := getPropName(prop)
					ctx.ReportNode(prop, rule.RuleMessage{
						Id:          "newLine",
						Description: fmt.Sprintf("Prop `%s` must be placed on a new line", propName),
					})
				}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

func getPropName(prop *ast.Node) string {
	if name := reactutil.GetJsxPropName(prop); name != "" {
		return name
	}
	return "prop"
}
