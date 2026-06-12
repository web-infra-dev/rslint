package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// setupDualRootFixture writes a fixture where the same file is a root of two
// tsconfigs (the documented dual-root case) with parse-irrelevant
// compilerOptions differences (target), plus a shared d.ts.
func setupDualRootFixture(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	files := map[string]string{
		"shared.ts":       "import { tag } from './types';\nexport var shared = 1;\nvar unused_shared = tag;\n",
		"a.ts":            "import { shared } from './shared';\nexport var a = shared + 1;\n",
		"b.ts":            "import { shared } from './shared';\nexport var b = shared + 2;\n",
		"types.ts":        "export var tag = 'tag';\n",
		"tsconfig.a.json": `{"files":["shared.ts","a.ts","types.ts"],"compilerOptions":{"target":"esnext","strict":true}}`,
		"tsconfig.b.json": `{"files":["shared.ts","b.ts","types.ts"],"compilerOptions":{"target":"es2017","strict":true}}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return tmpDir
}

// varReportingRule reports every `var` keyword usage — enough to produce
// deterministic diagnostics on every fixture file, including dual-root ones.
func varReportingRule() []ConfiguredRule {
	return []ConfiguredRule{{
		Name:     "test/no-var",
		Severity: rule.SeverityError,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return rule.RuleListeners{
				ast.KindVariableStatement: func(node *ast.Node) {
					ctx.ReportNode(node, rule.RuleMessage{Id: "noVar", Description: "no var"})
				},
			}
		},
	}}
}

type diagKey struct {
	File string
	Pos  int
	End  int
	Rule string
	Msg  string
}

func collectDiags(t *testing.T, programs []*compiler.Program, singleThreaded bool, typeCheck bool, withRules bool) []diagKey {
	t.Helper()
	var mu sync.Mutex
	var got []diagKey
	opts := RunLinterOptions{
		Programs:       programs,
		SingleThreaded: singleThreaded,
		TypeCheck:      typeCheck,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			got = append(got, diagKey{
				File: tspath.NormalizePath(d.FilePath),
				Pos:  d.Range.Pos(),
				End:  d.Range.End(),
				Rule: d.RuleName,
				Msg:  d.Message.Description,
			})
		},
	}
	if withRules {
		opts.GetRulesForFile = func(*ast.SourceFile) []ConfiguredRule { return varReportingRule() }
	}
	if _, err := RunLinter(opts); err != nil {
		t.Fatal(err)
	}
	sort.Slice(got, func(i, j int) bool {
		a, b := got[i], got[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Pos != b.Pos {
			return a.Pos < b.Pos
		}
		if a.Rule != b.Rule {
			return a.Rule < b.Rule
		}
		return a.Msg < b.Msg
	})
	return got
}

func buildDualRootPrograms(t *testing.T, tmpDir string, cache *utils.ParseCache) []*compiler.Program {
	t.Helper()
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.WithParseCache(utils.CreateCompilerHost(tmpDir, fs), cache)
	var programs []*compiler.Program
	for _, tc := range []string{"tsconfig.a.json", "tsconfig.b.json"} {
		p, err := utils.CreateProgram(false, fs, tmpDir, tc, host)
		if err != nil {
			t.Fatalf("CreateProgram(%s): %v", tc, err)
		}
		programs = append(programs, p)
	}
	return programs
}

func findByName(t *testing.T, p *compiler.Program, match func(string) bool) *ast.SourceFile {
	t.Helper()
	for _, sf := range p.GetSourceFiles() {
		if match(sf.FileName()) {
			return sf
		}
	}
	t.Fatal("file not found in program")
	return nil
}

// T7: a dual-root file shared through the parse cache is concurrently linted
// and type-checked by both programs (run this package with -race); the
// diagnostics must match a serial, cache-free baseline exactly.
func TestParseCacheSharing_DualRootDiagnosticsMatchBaseline(t *testing.T) {
	tmpDir := setupDualRootFixture(t)

	cache := &utils.ParseCache{}
	cached := buildDualRootPrograms(t, tmpDir, cache)

	// Sharing preconditions: the dual-root source, the shared import and a
	// bundled lib d.ts are all one object across the two programs despite
	// differing (parse-irrelevant) compilerOptions.
	for _, suffix := range []string{"/shared.ts", "/types.ts", "lib.es5.d.ts"} {
		fa := findByName(t, cached[0], func(n string) bool { return strings.HasSuffix(n, suffix) })
		fb := findByName(t, cached[1], func(n string) bool { return strings.HasSuffix(n, suffix) })
		if fa != fb {
			t.Fatalf("%s: expected one shared object across programs", suffix)
		}
	}

	baseline := buildDualRootPrograms(t, tmpDir, nil) // nil cache: independent parses

	for _, tc := range []struct {
		name      string
		typeCheck bool
		withRules bool
	}{
		{"lint", false, true},
		{"type-check", true, true},
		{"type-check-only", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := collectDiags(t, cached, false, tc.typeCheck, tc.withRules)
			want := collectDiags(t, baseline, true, tc.typeCheck, tc.withRules)
			if len(got) == 0 && tc.withRules {
				t.Fatal("expected diagnostics from the fixture, got none")
			}
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
				t.Fatalf("diagnostics differ from serial cache-free baseline:\n got: %v\nwant: %v", got, want)
			}
		})
	}

	// Post-run sweep with both programs live must keep every entry usable.
	cache.RetainOnly(cached)
	again := collectDiags(t, cached, false, true, true)
	want := collectDiags(t, baseline, true, true, true)
	if fmt.Sprintf("%v", again) != fmt.Sprintf("%v", want) {
		t.Fatal("diagnostics changed after RetainOnly sweep")
	}
}
