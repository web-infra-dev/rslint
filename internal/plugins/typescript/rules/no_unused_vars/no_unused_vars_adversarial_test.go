package no_unused_vars

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnusedVarsAmbientFastPathMatchesLegacy attacks the NodeFlagsAmbient
// fast path with ambient, non-ambient, nested, global-augmentation, and .d.ts
// contexts. The optimized helper must remain equivalent to the former full
// ancestor/source-file walk for every node in the corpus.
func TestNoUnusedVarsAmbientFastPathMatchesLegacy(t *testing.T) {
	testCases := []struct {
		name     string
		fileName string
		code     string
	}{
		{
			name:     "non-ambient namespace with local declare",
			fileName: "/runtime.ts",
			code: `namespace Runtime {
  const ordinary = 1;
  declare const locallyDeclared: number;
  namespace Nested { const value = 1; }
}`,
		},
		{
			name:     "ambient namespace and explicit export boundary",
			fileName: "/ambient.ts",
			code: `declare namespace Ambient {
  const publicValue: number;
  namespace Nested { const nestedValue: number; }
}
export declare namespace PrivateAmbient {
  const privateValue: number;
  export {};
}`,
		},
		{
			name:     "global augmentation",
			fileName: "/global.ts",
			code: `export {};
declare global {
  const BUILD_HASH: string;
  namespace GlobalNested { interface Value<T> { value: T } }
}`,
		},
		{
			name:     "declaration file",
			fileName: "/types.d.ts",
			code: `declare namespace PublicTypes {
  type Name = string;
}
declare namespace PrivateTypes {
  type Hidden = string;
  export {};
}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: testCase.fileName,
				Path:     tspath.Path(testCase.fileName),
			}, testCase.code, core.ScriptKindTS)
			optimizedContext := &analysisContext{
				declarationFile: sourceFile.IsDeclarationFile,
				sourceFile:      sourceFile,
			}
			legacyContext := &analysisContext{}

			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				gotAmbient := isInsideAmbientModuleBlock(node, optimizedContext)
				wantAmbient := legacyIsInsideAmbientModuleBlock(node, legacyContext)
				if gotAmbient != wantAmbient {
					t.Fatalf(
						"isInsideAmbientModuleBlock(%v at %d, flags=%v) = %v, legacy = %v",
						node.Kind,
						node.Pos(),
						node.Flags,
						gotAmbient,
						wantAmbient,
					)
				}

				gotDts := isInDtsWithoutExplicitExports(node, optimizedContext)
				wantDts := legacyIsInDtsWithoutExplicitExports(node, legacyContext)
				if gotDts != wantDts {
					t.Fatalf(
						"isInDtsWithoutExplicitExports(%v at %d, flags=%v) = %v, legacy = %v",
						node.Kind,
						node.Pos(),
						node.Flags,
						gotDts,
						wantDts,
					)
				}

				node.ForEachChild(visit)
				return false
			}
			sourceFile.AsNode().ForEachChild(visit)
		})
	}
}

func legacyIsInsideAmbientModuleBlock(node *ast.Node, ac *analysisContext) bool {
	moduleBlock := ast.FindAncestorKind(node, ast.KindModuleBlock)
	if moduleBlock == nil {
		return false
	}
	moduleDecl := moduleBlock.Parent
	if moduleDecl == nil || moduleDecl.Kind != ast.KindModuleDeclaration {
		return false
	}

	isAmbient := ast.GetCombinedModifierFlags(moduleDecl)&ast.ModifierFlagsAmbient != 0
	if !isAmbient && ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		return ast.IsGlobalScopeAugmentation(n)
	}) != nil {
		isAmbient = true
	}
	if !isAmbient && ast.FindAncestor(moduleDecl.Parent, func(n *ast.Node) bool {
		return n.Kind == ast.KindModuleDeclaration &&
			ast.HasSyntacticModifier(n, ast.ModifierFlagsAmbient)
	}) != nil {
		isAmbient = true
	}
	if !isAmbient {
		sourceFile := ast.GetSourceFileOfNode(node)
		isAmbient = sourceFile != nil && sourceFile.IsDeclarationFile
	}
	return isAmbient && !containerHasExplicitExports(moduleBlock, ac)
}

func legacyIsInDtsWithoutExplicitExports(node *ast.Node, ac *analysisContext) bool {
	sourceFile := ast.GetSourceFileOfNode(node)
	if sourceFile == nil || !sourceFile.IsDeclarationFile {
		return false
	}
	if ast.FindAncestorKind(node, ast.KindModuleBlock) != nil {
		return false
	}
	return !containerHasExplicitExports(sourceFile.AsNode(), ac)
}

// TestNoUnusedVarsSelfModificationRequiresWriteReference verifies the
// invariant used to skip expensive self-modification analysis: every shape
// that the analyzer classifies as self-modifying also exposes at least one
// write reference for the same binding.
func TestNoUnusedVarsSelfModificationRequiresWriteReference(t *testing.T) {
	code := `export {};
let direct = 0;
direct = direct + 1;
let compound = 0;
(compound) += 1;
let update = 0;
(update)++;
let call: any = 0;
call = consume(call);
let method: any[] = [];
method = method.concat(method);
let iife: any;
iife = (function () { return iife(); })();
let sequence: any;
sequence = (0, function () { sequence(); });
function mutateParameter(parameter: number) {
  parameter = parameter + 1;
}
console.log(mutateParameter);
`
	program, sourceFile := createNoUnusedVarsProgram(t, "self-modification-invariant.ts", code)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()
	ctx := (rule.RuleContext{
		SourceFile:  sourceFile,
		TypeChecker: typeChecker,
		Refs:        rule.NewRefStore(sourceFile, program.Options(), typeChecker, rule.RefStoreInit{}),
	}).WithProgram(lintprogram.NewFromCompiler(program))
	globalSourceFile := ast.IsGlobalSourceFile(sourceFile.AsNode())
	selfModifyingCount := 0

	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration || node.Kind == ast.KindParameter {
			nameNode := node.Name()
			if nameNode != nil && ast.IsIdentifier(nameNode) {
				rawSymbol := binderSymbolForDefinition(nameNode, node)
				symbol := symbolForVariable(ctx, nameNode, node, rawSymbol, globalSourceFile)
				info := classifyReferenceNodes(ctx.Refs.References(rawSymbol))
				for _, usage := range info.usages {
					if symbol != nil && isSelfModifyingReference(usage, symbol, typeChecker, nameNode) {
						selfModifyingCount++
						if len(info.writeRefs) == 0 {
							t.Fatalf("self-modifying reference %q at %d has no write reference", nameNode.Text(), usage.Pos())
						}
					}
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	sourceFile.AsNode().ForEachChild(visit)
	if selfModifyingCount == 0 {
		t.Fatal("adversarial corpus did not exercise a self-modifying reference")
	}
}

// TestNoUnusedVarsSingleParameterBoundaryCache forces a nested function in a
// default initializer to replace the single-entry cache before traversal
// returns to later parameters of the outer function.
func TestNoUnusedVarsSingleParameterBoundaryCache(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{{
			Code: `function outer(
  first: number,
  interrupt = ((innerFirst: number, innerLast: number) => innerLast)(1, 2),
  resumed: number,
  finalUsed: number,
) {
  return finalUsed;
}
console.log(outer);`,
		}},
		nil,
	)
}

// TestNoUnusedVarsAfterUsedEarlyExitGuardrails attacks the early after-used
// exit with its important exception: a used prefix parameter that matches an
// ignore pattern must still be reported when reportUsedIgnorePattern is on.
func TestNoUnusedVarsAfterUsedEarlyExitGuardrails(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		nil,
		[]rule_tester.InvalidTestCase{{
			Code: `function consume(_prefix: number, finalUsed: number) {
  return _prefix + finalUsed;
}
console.log(consume);`,
			Options: map[string]interface{}{
				"argsIgnorePattern":       "^_",
				"reportUsedIgnorePattern": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "usedIgnoredVar",
				Line:      1,
				Column:    18,
			}},
		}},
	)
}
