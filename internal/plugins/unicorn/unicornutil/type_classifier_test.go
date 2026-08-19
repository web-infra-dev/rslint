package unicornutil

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"gotest.tools/v3/assert"
)

func TestClassifyType(t *testing.T) {
	root := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(root.Dir, "type-classifier.ts")
	code := `
declare const stringValue: string;
declare const numberValue: number;
declare const unknownValue: unknown;
declare const mixedValue: string | number;
declare const nonTargetUnion: number | boolean;
declare const nullableString: string | null;
declare const stringIntersection: string & {brand: true};
function constrained<T extends string>(value: T) { return value; }
interface Text extends String {}
declare const inheritedString: Text;
class ClassText extends String {}
declare const inheritedClassString: ClassText;
void stringValue;
void numberValue;
void unknownValue;
void mixedValue;
void nonTargetUnion;
void nullableString;
void stringIntersection;
void inheritedString;
void inheritedClassString;
`
	fs := utils.NewOverlayVFS(root.FS, map[string]string{filePath: code})
	program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", utils.CreateCompilerHost(root.Dir, fs))
	assert.NilError(t, err)
	sourceFile := program.GetSourceFile(filePath)
	assert.Assert(t, sourceFile != nil)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()

	options := TypeClassifierOptions{
		TargetTypeNames: utils.NewSetFromItems("String"),
		IsTargetType: func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike)
		},
		HeritageSymbolFlags: ast.SymbolFlagsClass | ast.SymbolFlagsInterface,
	}
	ctx := rule.RuleContext{SourceFile: sourceFile, TypeChecker: typeChecker}
	tests := []struct {
		name string
		want TypeClass
	}{
		{name: "stringValue", want: TypeTarget},
		{name: "numberValue", want: TypeNonTarget},
		{name: "unknownValue", want: TypeUnknown},
		{name: "mixedValue", want: TypeUnknown},
		{name: "nonTargetUnion", want: TypeNonTarget},
		{name: "nullableString", want: TypeUnknown},
		{name: "stringIntersection", want: TypeTarget},
		{name: "value", want: TypeTarget},
		{name: "inheritedString", want: TypeTarget},
		{name: "inheritedClassString", want: TypeTarget},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := findReferenceIdentifier(t, sourceFile, test.name)
			got := ClassifyType(ctx, typeChecker.GetTypeAtLocation(node), options)
			if got != test.want {
				t.Fatalf("ClassifyType(%s) = %v, want %v", test.name, got, test.want)
			}
		})
	}
}

func findReferenceIdentifier(t *testing.T, sourceFile *ast.SourceFile, name string) *ast.Node {
	t.Helper()
	var found *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if ast.IsIdentifier(node) && node.AsIdentifier().Text == name && !ast.IsDeclarationName(node) {
			parent := node.Parent
			if parent != nil &&
				(parent.Kind == ast.KindVoidExpression || parent.Kind == ast.KindReturnStatement) {
				found = node
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	visit(sourceFile.AsNode())
	if found == nil {
		t.Fatalf("missing reference identifier %q", name)
	}
	return found
}
