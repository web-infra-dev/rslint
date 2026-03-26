package utils

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
)

const (
	JestFnTypeExpect   JestFnType = "expect"
	JestFnTypeDescribe JestFnType = "describe"
	JestFnTypeHook     JestFnType = "hook"
	JestFnTypeJest     JestFnType = "jest"
	JestFnTypeTest     JestFnType = "test"
	JestFnTypeUnknown  JestFnType = "unknown"
)

var JEST_METHOD_NAMES = map[string]bool{
	"afterAll":   true,
	"afterEach":  true,
	"beforeAll":  true,
	"beforeEach": true,
	"describe":   true,
	"expect":     true,
	"fdescribe":  true,
	"fit":        true,
	"it":         true,
	"jest":       true,
	"test":       true,
	"xdescribe":  true,
	"xit":        true,
	"xtest":      true,
}

var EXPECT_MODIFIER_NAMES = map[string]bool{
	"not":      true,
	"rejects":  true,
	"resolves": true,
}

var VALID_JEST_FN_CALL_CHAINS = map[string]bool{
	"afterAll":                  true,
	"afterEach":                 true,
	"beforeAll":                 true,
	"beforeEach":                true,
	"describe":                  true,
	"describe.each":             true,
	"describe.only":             true,
	"describe.only.each":        true,
	"describe.skip":             true,
	"describe.skip.each":        true,
	"fdescribe":                 true,
	"fdescribe.each":            true,
	"fit":                       true,
	"fit.each":                  true,
	"fit.failing":               true,
	"fit.fails":                 true,
	"it":                        true,
	"it.concurrent":             true,
	"it.concurrent.each":        true,
	"it.concurrent.only.each":   true,
	"it.concurrent.skip.each":   true,
	"it.each":                   true,
	"it.failing":                true,
	"it.fails":                  true,
	"it.only":                   true,
	"it.only.each":              true,
	"it.only.failing":           true,
	"it.only.fails":             true,
	"it.skip":                   true,
	"it.skip.each":              true,
	"it.skip.failing":           true,
	"it.skip.fails":             true,
	"it.todo":                   true,
	"test":                      true,
	"test.concurrent":           true,
	"test.concurrent.each":      true,
	"test.concurrent.only.each": true,
	"test.concurrent.skip.each": true,
	"test.each":                 true,
	"test.failing":              true,
	"test.fails":                true,
	"test.only":                 true,
	"test.only.each":            true,
	"test.only.failing":         true,
	"test.only.fails":           true,
	"test.skip":                 true,
	"test.skip.each":            true,
	"test.skip.failing":         true,
	"test.skip.fails":           true,
	"test.todo":                 true,
	"xdescribe":                 true,
	"xdescribe.each":            true,
	"xit":                       true,
	"xit.each":                  true,
	"xit.failing":               true,
	"xit.fails":                 true,
	"xtest":                     true,
	"xtest.each":                true,
	"xtest.failing":             true,
	"xtest.fails":               true,
}

type JestFnType string

func getPropertyName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return node.AsPrivateIdentifier().Text
	}
	return ""
}

func GetJestKind(name string) JestFnType {
	switch name {
	case "describe", "fdescribe", "xdescribe":
		return JestFnTypeDescribe
	case "fit", "it", "test", "xit", "xtest":
		return JestFnTypeTest
	case "beforeEach", "afterEach", "beforeAll", "afterAll":
		return JestFnTypeHook
	case "jest":
		return JestFnTypeJest
	case "expect":
		return JestFnTypeExpect
	default:
		return JestFnTypeUnknown
	}
}

func GetMembersName(node *ast.Node) string {
	chain := GetMembersChain(node)
	if chain == nil {
		return ""
	}

	return strings.Join(chain, ".")
}

func GetMembersChain(node *ast.Node) []string {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return []string{node.AsIdentifier().Text}
	case ast.KindPropertyAccessExpression:
		p := node.AsPropertyAccessExpression()
		left := GetMembersChain(p.Expression)
		if name := getPropertyName(p.Name()); name != "" {
			return append(left, name)
		}
		return left
	case ast.KindElementAccessExpression:
		p := node.AsElementAccessExpression()
		left := GetMembersChain(p.Expression)
		if name := getElementAccessName(p.ArgumentExpression); name != "" {
			return append(left, name)
		}
		return left
	case ast.KindCallExpression:
		return GetMembersChain(node.AsCallExpression().Expression)
	case ast.KindTaggedTemplateExpression:
		return GetMembersChain(node.AsTaggedTemplateExpression().Tag)
	default:
		return nil
	}
}

func getElementAccessName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	node = ast.SkipParentheses(node)
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text
	default:
		return ""
	}
}
