package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestAccessExpressionParts(t *testing.T) {
	t.Parallel()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/access.ts",
		Path:     "/access.ts",
	}, `value.member; value[key]; value["literal"]; value;`, core.ScriptKindTS)

	tests := []struct {
		statement    int
		propertyKind ast.Kind
		propertyText string
	}{
		{statement: 0, propertyKind: ast.KindIdentifier, propertyText: "member"},
		{statement: 1, propertyKind: ast.KindIdentifier, propertyText: "key"},
		{statement: 2, propertyKind: ast.KindStringLiteral, propertyText: "literal"},
	}
	for _, test := range tests {
		expression := sourceFile.Statements.Nodes[test.statement].AsExpressionStatement().Expression
		object := AccessExpressionObject(expression)
		if object == nil || object.Text() != "value" {
			t.Errorf("statement %d object = %v, want value", test.statement, object)
		}
		property := AccessExpressionProperty(expression)
		if property == nil || property.Kind != test.propertyKind || property.Text() != test.propertyText {
			t.Errorf("statement %d property = %v, want %s %q", test.statement, property, test.propertyKind, test.propertyText)
		}
	}

	nonAccess := sourceFile.Statements.Nodes[3].AsExpressionStatement().Expression
	if AccessExpressionObject(nonAccess) != nil || AccessExpressionProperty(nonAccess) != nil {
		t.Fatal("non-access expression returned an access part")
	}
	if AccessExpressionObject(nil) != nil || AccessExpressionProperty(nil) != nil {
		t.Fatal("nil expression returned an access part")
	}
}

func TestESTreePropertyKey(t *testing.T) {
	t.Parallel()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/properties.ts",
		Path:     "/properties.ts",
	}, `const object = { plain: 1, [computed]: 2, [("literal")]: 3 };`, core.ScriptKindTS)
	declaration := sourceFile.Statements.Nodes[0].AsVariableStatement().
		DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].AsVariableDeclaration()
	properties := declaration.Initializer.AsObjectLiteralExpression().Properties.Nodes
	tests := []struct {
		index int
		kind  ast.Kind
		text  string
	}{
		{index: 0, kind: ast.KindIdentifier, text: "plain"},
		{index: 1, kind: ast.KindIdentifier, text: "computed"},
		{index: 2, kind: ast.KindStringLiteral, text: "literal"},
	}
	for _, test := range tests {
		key := ESTreePropertyKey(properties[test.index].Name())
		if key == nil || key.Kind != test.kind || key.Text() != test.text {
			t.Errorf("property %d key = %v, want %s %q", test.index, key, test.kind, test.text)
		}
	}
	if ESTreePropertyKey(nil) != nil {
		t.Fatal("nil property produced an ESTree key")
	}
}
