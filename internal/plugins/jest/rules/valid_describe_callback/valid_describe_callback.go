package valid_describe_callback

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builder
func buildErrorNameAndCallbackMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "nameAndCallback",
		Description: "Describe requires name and callback arguments",
	}
}

func buildErrorSecondArgumentMustBeFunctionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "secondArgumentMustBeFunction",
		Description: "Second argument must be function",
	}
}

func buildErrorNoAsyncDescribeCallbackMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noAsyncDescribeCallback",
		Description: "No async describe callback",
	}
}

func buildErrorUnexpectedDescribeArgumentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedDescribeArgument",
		Description: "Unexpected argument(s) in describe callback",
	}
}

func buildErrorUnexpectedReturnInDescribeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedReturnInDescribe",
		Description: "Unexpected return statement in describe callback",
	}
}

func findFirstReturnStatement(nodes []*ast.Node) *ast.Node {
	for _, statement := range nodes {
		if statement.Kind == ast.KindReturnStatement {
			return statement
		}
	}
	return nil
}

var ValidDescribeCallbackRule = rule.Rule{
	Name: "jest/valid-describe-callback",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeDescribe {
					return
				}

				argumentLength := len(node.AsCallExpression().Arguments.Nodes)

				switch argumentLength {
				case 0:
					ctx.ReportNode(node, buildErrorNameAndCallbackMessage())
				case 1:
					ctx.ReportNode(node.AsCallExpression().Arguments.Nodes[0], buildErrorNameAndCallbackMessage())
				default:
					switch node.AsCallExpression().Arguments.Nodes[1].Kind {
					case ast.KindArrowFunction:
						{
							arrowFuncExpression := node.AsCallExpression().Arguments.Nodes[1]
							if ast.IsAsyncFunction(arrowFuncExpression) {
								ctx.ReportNode(arrowFuncExpression, buildErrorNoAsyncDescribeCallbackMessage())
							}

							arrowFn := arrowFuncExpression.AsArrowFunction()
							if !slices.Contains(jestFnCall.Members, "each") && len(arrowFn.Parameters.Nodes) > 0 {
								ctx.ReportNode(arrowFn.Parameters.Nodes[0], buildErrorUnexpectedDescribeArgumentMessage())
							}

							body := arrowFn.Body
							if body != nil {
								if body.Kind == ast.KindBlock {
									if ret := findFirstReturnStatement(body.AsBlock().Statements.Nodes); ret != nil {
										ctx.ReportNode(ret, buildErrorUnexpectedReturnInDescribeMessage())
									}
								} else {
									// Concise arrow body: implicit return value (not a block).
									ctx.ReportNode(body, buildErrorUnexpectedReturnInDescribeMessage())
								}
							}
						}
					case ast.KindFunctionExpression:
						{
							funcExpression := node.AsCallExpression().Arguments.Nodes[1]
							if ast.IsAsyncFunction(funcExpression) {
								ctx.ReportNode(funcExpression, buildErrorNoAsyncDescribeCallbackMessage())
							}

							if !slices.Contains(jestFnCall.Members, "each") && len(funcExpression.AsFunctionExpression().Parameters.Nodes) > 0 {
								ctx.ReportNode(funcExpression.AsFunctionExpression().Parameters.Nodes[0], buildErrorUnexpectedDescribeArgumentMessage())
							}

							body := funcExpression.AsFunctionExpression().Body
							if body != nil {
								returnStatement := findFirstReturnStatement(body.AsBlock().Statements.Nodes)
								if returnStatement != nil {
									ctx.ReportNode(returnStatement, buildErrorUnexpectedReturnInDescribeMessage())
								}
							}
						}
					default:
						ctx.ReportNode(node.AsCallExpression().Arguments.Nodes[1], buildErrorSecondArgumentMustBeFunctionMessage())
					}
				}
			},
		}
	},
}
