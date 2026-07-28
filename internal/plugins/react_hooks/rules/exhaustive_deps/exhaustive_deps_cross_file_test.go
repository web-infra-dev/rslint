package exhaustive_deps

import (
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestContainsNodeRejectsCrossFileNodes pins the same-SourceFile
// requirement of containsNode. Node Pos()/End() are per-file byte
// offsets, so any two nodes whose per-file ranges happen to overlap
// numerically would falsely satisfy a raw-integer comparison even when
// they live in different files — and several call sites
// (computeExternalDeps, anyDeclWithinComponent, processIdentifier)
// rely on the inverse to classify a declaration as "in component scope".
func TestContainsNodeRejectsCrossFileNodes(t *testing.T) {
	t.Parallel()

	rootDir := "/virtual-contains"
	pathA := tspath.ResolvePath(rootDir, "a.ts")
	pathB := tspath.ResolvePath(rootDir, "b.ts")
	tsconfigPath := tspath.ResolvePath(rootDir, "tsconfig.json")

	// File A: `function outer() { let inner = 1; }` — `outer`'s range
	// spans most of the file.
	srcA := "function outer() { let inner = 1; }\n"
	// File B: a `function foo` declaration whose positions land
	// numerically inside `outer`'s range above.
	srcB := "function foo() {}\n"
	tsconfig := `{
  "compilerOptions": {
    "target": "esnext",
    "module": "esnext",
    "strict": false,
    "noEmit": true,
    "types": [],
    "lib": ["esnext"]
  },
  "include": ["a.ts", "b.ts"]
}
`

	virtual := map[string]string{
		pathA:        srcA,
		pathB:        srcB,
		tsconfigPath: tsconfig,
	}
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), virtual)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}

	fileA := program.GetSourceFile(pathA)
	fileB := program.GetSourceFile(pathB)
	if fileA == nil || fileB == nil {
		t.Fatalf("missing source files: a=%v b=%v", fileA, fileB)
	}

	var outerInA, fooInB *ast.Node
	walk := func(file *ast.SourceFile, name string, out **ast.Node) {
		var visit func(n *ast.Node) bool
		visit = func(n *ast.Node) bool {
			if n == nil || *out != nil {
				return false
			}
			if n.Kind == ast.KindFunctionDeclaration && n.Name() != nil &&
				n.Name().AsIdentifier().Text == name {
				*out = n
				return true
			}
			n.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
	}
	walk(fileA, "outer", &outerInA)
	walk(fileB, "foo", &fooInB)
	if outerInA == nil || fooInB == nil {
		t.Fatalf("missing test nodes: outer=%v foo=%v", outerInA, fooInB)
	}

	// Sanity: by construction the byte ranges must overlap so that the
	// raw-integer comparison would otherwise return true. Without this
	// invariant the assertion below would trivially pass.
	if fooInB.Pos() < outerInA.Pos() || fooInB.End() > outerInA.End() {
		t.Fatalf(
			"fixture sizing drift: foo[%d,%d] must lie numerically inside outer[%d,%d]",
			fooInB.Pos(), fooInB.End(), outerInA.Pos(), outerInA.End(),
		)
	}

	if containsNode(outerInA, fooInB) {
		t.Fatalf(
			"containsNode treated cross-file nodes as contained: outer in %s [%d,%d], foo in %s [%d,%d]",
			fileA.FileName(), outerInA.Pos(), outerInA.End(),
			fileB.FileName(), fooInB.Pos(), fooInB.End(),
		)
	}
}

// TestExhaustiveDepsRule_CrossFileGlobalInDeps exercises the end-to-end
// rule path when a deps array entry resolves to an ambient declaration
// in another SourceFile — the typical shape for globals like
// `setTimeout` once both DOM lib types and `@types/node` are loaded.
//
// The fixture pins the exact arithmetic that previously crashed the
// rule:
//
//   - A *small* ambient .d.ts declares `setTimeout` at byte offsets that
//     numerically fall inside the user component's function-body range
//     in the .tsx file. A purely positional containment check would
//     therefore classify that foreign declaration as "in scope" and stop
//     the rule from marking the dep as external — keeping
//     problemCount == 0 and routing the flow into
//     emitConstructionWarnings.
//
//   - A second, *much larger* ambient .d.ts also declares `setTimeout`,
//     with byte offsets well past the .tsx file's length. The
//     TypeChecker merges both into one symbol; resolveDeclaration in
//     scanForConstructions returns sym.Declarations[0], i.e. the
//     foreign FunctionDeclaration. When ReportNode then asks the
//     scanner to slice the .tsx text using that foreign position the
//     slice indices go out of bounds.
//
// The rule must finish without panicking and every diagnostic it emits
// must point at a range inside the user file.
func TestExhaustiveDepsRule_CrossFileGlobalInDeps(t *testing.T) {
	t.Parallel()

	rootDir := "/virtual-cross-file"
	tsxPath := tspath.ResolvePath(rootDir, "user.tsx")
	smallLibPath := tspath.ResolvePath(rootDir, "a-small.d.ts")
	bigLibPath := tspath.ResolvePath(rootDir, "b-big.d.ts")
	tsconfigPath := tspath.ResolvePath(rootDir, "tsconfig.json")

	// Pad b-big.d.ts so that `function setTimeout` lands well past the
	// user .tsx file's length. The padding is declaration noise the
	// linter never looks at; only the resulting byte offset matters.
	bigLib := strings.Repeat("declare const _b_pad: number;\n", 80) +
		`declare function setTimeout(handler: () => void, timeout: number): number;
declare function useEffect(cb: () => void, deps: unknown[]): void;
declare function useCallback<T>(cb: T, deps: unknown[]): T;
`

	// The small lib's `setTimeout` byte offsets fall inside the user
	// component's positional range; without the cross-file guard,
	// containsNode reports a false positive against it.
	smallLib := `declare const __a_pad1: number;
declare function setTimeout(handler: () => void, timeout: number): number;
`

	user := `function MyComponent() {
  const prefetchRoutes = useCallback(() => {
    setTimeout(() => {}, 2000);
  }, [setTimeout]);
  useEffect(() => {
    setTimeout(prefetchRoutes, 2000);
  }, [prefetchRoutes, setTimeout]);
  return null;
}
`

	tsconfig := `{
  "compilerOptions": {
    "jsx": "preserve",
    "target": "esnext",
    "module": "esnext",
    "strict": false,
    "noEmit": true,
    "types": [],
    "lib": ["esnext"]
  },
  "include": ["b-big.d.ts", "a-small.d.ts", "user.tsx"]
}
`

	virtual := map[string]string{
		tsxPath:      user,
		smallLibPath: smallLib,
		bigLibPath:   bigLib,
		tsconfigPath: tsconfig,
	}
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), virtual)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	tsxFile := program.GetSourceFile(tsxPath)
	if tsxFile == nil {
		t.Fatalf("missing user.tsx source file")
	}

	var mu sync.Mutex
	var diagnostics []rule.RuleDiagnostic

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("exhaustive-deps panicked on cross-file dep: %v", r)
		}
	}()

	if _, err := linter.RunLinter(linter.RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		Scope:          linter.FileScope{Files: []string{tsxPath}},
		ExcludePaths:   []string{},
		GetRulesForFile: func(_ *ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     "react-hooks/exhaustive-deps",
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return ExhaustiveDepsRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: func(d rule.RuleDiagnostic) {
				mu.Lock()
				defer mu.Unlock()
				diagnostics = append(diagnostics, d)
			},
		},
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}

	// Every diagnostic must point at a range that exists in the user
	// file. A cross-file leak shows up as out-of-range positions even
	// when the linter happens not to panic.
	tsxLen := len(tsxFile.Text())
	for i, d := range diagnostics {
		if d.Range.Pos() < 0 || d.Range.End() > tsxLen {
			t.Errorf("diagnostic %d has cross-file range [%d,%d] (user file length %d)",
				i, d.Range.Pos(), d.Range.End(), tsxLen)
		}
	}
}
