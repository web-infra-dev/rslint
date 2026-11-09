package prefer_modern_math_apis

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/utils"
	"golang.org/x/text/unicode/norm"
)

// Keep ordinary keys on the shared evaluator. Lower the missing pure calls on
// a temporary tree so nested calls still use its coercion and indexing rules.
func staticKeyText(node *ast.Node) (string, bool) {
	if node == nil || ast.IsIdentifier(node) {
		return "", false
	}
	evaluator := utils.NewStaticStringEvaluatorWithoutScope()
	if text, ok := evaluator.EvalToString(node); ok {
		return text, true
	}
	folder := literalMethodFolder{evaluator: evaluator, factory: ast.NewNodeFactory(ast.NodeFactoryHooks{})}
	visitor := ast.NewNodeVisitor(nil, folder.factory, ast.NodeVisitorHooks{})
	visitor.Visit = func(node *ast.Node) *ast.Node {
		node = node.VisitEachChild(visitor)
		if ast.IsCallExpression(node) {
			if result := folder.foldCall(node); result != nil {
				return result
			}
		}
		return node
	}
	folded := visitor.VisitNode(node)
	if folded == node {
		return "", false
	}
	return evaluator.EvalToString(folded)
}

type literalMethodFolder struct {
	evaluator *utils.StaticStringEvaluator
	factory   *ast.NodeFactory
}

func (f literalMethodFolder) foldCall(node *ast.Node) *ast.Node {
	callee := utils.ESTreeCallCallee(node.Expression())
	if callee == nil || !ast.IsAccessExpression(callee) {
		return nil
	}
	method, ok := f.evaluator.EvalAccessExpressionName(callee)
	if !ok {
		return nil
	}
	switch method {
	case "at", "normalize", "concat", "slice", "flat", "includes", "indexOf", "lastIndexOf", "toString":
	default:
		return nil
	}
	args := node.Arguments()
	for _, arg := range args {
		if _, known := f.evaluator.EvalValue(arg); !known {
			return nil
		}
	}
	receiver := callee.Expression()
	if isArray, _ := f.evaluator.EvalArrayValue(receiver); isArray {
		return f.foldArrayCall(receiver, method, args)
	}
	if text, ok := f.evaluator.Eval(receiver); ok {
		switch method {
		case "at":
			return f.at(receiver, args)
		case "normalize":
			return f.normalize(text, args)
		}
	}
	return nil
}

func (f literalMethodFolder) at(receiver *ast.Node, args []*ast.Node) *ast.Node {
	lengthNode := f.factory.NewPropertyAccessExpression(receiver, nil, f.factory.NewIdentifier("length"), 0)
	length, ok := f.integer(lengthNode)
	if !ok {
		return nil
	}
	index, ok := f.argumentInteger(args, 0, 0)
	if !ok {
		return nil
	}
	index = relativeIndex(index, length)
	if index < 0 || index >= length {
		return f.undefined()
	}
	return f.factory.NewElementAccessExpression(receiver, nil, f.number(int(index)), 0)
}

func (f literalMethodFolder) normalize(text string, args []*ast.Node) *ast.Node {
	form := norm.NFC
	if len(args) != 0 && !f.isUndefined(args[0]) {
		name, ok := f.evaluator.EvalToString(args[0])
		if !ok {
			return nil
		}
		switch name {
		case "NFC":
		case "NFD":
			form = norm.NFD
		case "NFKC":
			form = norm.NFKC
		case "NFKD":
			form = norm.NFKD
		default:
			return nil
		}
	}
	result := form.String(text)
	// x/text can add stream-safety separators to long runs of combining
	// marks. JavaScript preserves those runs without inserting separators.
	if strings.Count(result, "\u034f") != strings.Count(text, "\u034f") {
		return nil
	}
	return f.factory.NewStringLiteral(result, 0)
}

func (f literalMethodFolder) foldArrayCall(receiver *ast.Node, method string, args []*ast.Node) *ast.Node {
	switch method {
	case "at":
		return f.at(receiver, args)
	case "toString":
		if text, ok := f.evaluator.EvalToString(receiver); ok {
			return f.factory.NewStringLiteral(text, 0)
		}
		return nil
	case "normalize":
		return nil
	}
	elements, ok := literalElements(receiver)
	if !ok {
		return nil
	}
	switch method {
	case "concat":
		result := slices.Clone(elements)
		for _, arg := range args {
			if isArray, _ := f.evaluator.EvalArrayValue(arg); isArray {
				part, ok := literalElements(arg)
				if !ok {
					return nil
				}
				result = append(result, part...)
			} else {
				result = append(result, arg)
			}
		}
		return f.array(result)
	case "slice":
		length := float64(len(elements))
		start, startOK := f.argumentInteger(args, 0, 0)
		end, endOK := f.argumentInteger(args, 1, length)
		if !startOK || !endOK {
			return nil
		}
		from := int(min(length, max(0, relativeIndex(start, length))))
		to := int(min(length, max(float64(from), relativeIndex(end, length))))
		return f.array(elements[from:to])
	case "flat":
		depth, ok := f.argumentInteger(args, 0, 1)
		var flattened []*ast.Node
		if !ok || !f.flatten(elements, depth, &flattened) {
			return nil
		}
		return f.array(flattened)
	case "includes", "indexOf", "lastIndexOf":
		return f.search(elements, method, args)
	}
	return nil
}

func literalElements(node *ast.Node) ([]*ast.Node, bool) {
	node = ast.SkipOuterExpressions(node, numberWrappers)
	if !ast.IsArrayLiteralExpression(node) {
		return nil, false
	}
	elements := node.AsArrayLiteralExpression().Elements.Nodes
	return elements, !slices.ContainsFunc(elements, ast.IsSpreadElement)
}

func (f literalMethodFolder) flatten(elements []*ast.Node, depth float64, result *[]*ast.Node) bool {
	for _, element := range elements {
		if ast.IsOmittedExpression(element) {
			continue
		}
		if depth > 0 {
			isArray, _ := f.evaluator.EvalArrayValue(element)
			if !isArray {
				*result = append(*result, element)
				continue
			}
			inner, ok := literalElements(element)
			if !ok || !f.flatten(inner, depth-1, result) {
				return false
			}
			continue
		}
		*result = append(*result, element)
	}
	return true
}

func (f literalMethodFolder) search(elements []*ast.Node, method string, args []*ast.Node) *ast.Node {
	needle := f.undefined()
	if len(args) > 0 {
		needle = args[0]
	}
	length := float64(len(elements))
	start, step := 0.0, 1
	if method == "lastIndexOf" {
		start, step = length-1, -1
	}
	if len(args) > 1 {
		index, ok := f.integer(args[1])
		if !ok {
			return nil
		}
		start = relativeIndex(index, length)
	}
	if step > 0 {
		start = min(length, max(0, start))
	} else {
		start = max(-1, min(length-1, start))
	}
	needleValue, _ := f.evaluator.EvalValue(needle)
	number, numericNeedle := needleValue.(jsnum.Number)
	includes := method == "includes"
	for i := int(start); i >= 0 && i < len(elements); i += step {
		element := elements[i]
		if ast.IsOmittedExpression(element) {
			if !includes {
				continue
			}
			element = f.undefined()
		}
		equal, ok := f.strictEqual(element, needle)
		if !ok {
			return nil
		}
		if includes && numericNeedle && number.IsNaN() {
			value, _ := f.evaluator.EvalValue(element)
			num, ok := value.(jsnum.Number)
			equal = ok && num.IsNaN()
		}
		if equal {
			if includes {
				return f.factory.NewToken(ast.KindTrueKeyword)
			}
			return f.number(i)
		}
	}
	if includes {
		return f.factory.NewToken(ast.KindFalseKeyword)
	}
	return f.number(-1)
}

func (f literalMethodFolder) argumentInteger(args []*ast.Node, index int, defaultValue float64) (float64, bool) {
	if index >= len(args) || defaultValue != 0 && f.isUndefined(args[index]) {
		return defaultValue, true
	}
	return f.integer(args[index])
}

func (f literalMethodFolder) integer(node *ast.Node) (float64, bool) {
	value, ok := f.evaluator.EvalValue(f.factory.NewPrefixUnaryExpression(ast.KindPlusToken, node))
	if number, isNumber := value.(jsnum.Number); ok && isNumber {
		if number.IsNaN() {
			return 0, true
		}
		return math.Trunc(float64(number)), true
	}
	return 0, false
}

func relativeIndex(index, length float64) float64 {
	if index < 0 {
		return length + index
	}
	return index
}

func (f literalMethodFolder) strictEqual(left, right *ast.Node) (bool, bool) {
	comparison := f.factory.NewBinaryExpression(nil, left, nil, f.factory.NewToken(ast.KindEqualsEqualsEqualsToken), right)
	value, ok := f.evaluator.EvalValue(comparison)
	equal, isBoolean := value.(bool)
	return equal, ok && isBoolean
}

func (f literalMethodFolder) isUndefined(node *ast.Node) bool {
	equal, _ := f.strictEqual(node, f.undefined())
	return equal
}

func (f literalMethodFolder) number(value int) *ast.Node {
	return f.factory.NewNumericLiteral(strconv.Itoa(value), 0)
}

func (f literalMethodFolder) undefined() *ast.Node {
	return f.factory.NewVoidExpression(f.number(0))
}

func (f literalMethodFolder) array(elements []*ast.Node) *ast.Node {
	return f.factory.NewArrayLiteralExpression(f.factory.NewNodeList(elements), false)
}
