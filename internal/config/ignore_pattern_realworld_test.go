package config

import (
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

// Tests anchored to REAL ignore configurations harvested from production repos
// (rsbuild / rspack / create-rstack / agent-skills .gitignore + rslint.config.ts)
// and the well-known ESLint flat-config idioms. The synthetic suites in
// ignore_pattern_test.go / file_discovery_dir_prune_test.go already pin the
// classification switch and the prune predicate on hand-built inputs; these add
// the realistic patterns those suites do not exercise — specifically the
// gitignore-converted forms and the wildcard-mid-path `!` re-includes that
// rspack actually ships. Every assertion is exact; no loose matches.

// --- Real gitignore lines, exact converted glob + classification ---
//
// convertSinglePattern is the production gitignore→glob converter. Pinning its
// output for real lines guards the Kind boundary that drives directory pruning:
// a `dir/` (→ fileLevelCover, prunable) vs a rooted nested path `a/b` (→
// absoluteBlock) vs a single-level `dir/*` (→ dirNone, NOT prunable).
func TestRealGitignoreLine_ConvertAndClassify(t *testing.T) {
	cases := []struct {
		line    string // raw .gitignore line (root .gitignore, baseDir="")
		glob    string // expected convertSinglePattern output
		negated bool
		kind    dirKind
	}{
		// rsbuild
		{"node_modules/", "**/node_modules/**/*", false, dirFileLevelCover},
		{"dist/", "**/dist/**/*", false, dirFileLevelCover},
		{"dist-*", "**/dist-*", false, dirAbsoluteBlock}, // glob name, not a dir cover
		{"*.log*", "**/*.log*", false, dirAbsoluteBlock}, // file pattern
		{"*.css.d.ts", "**/*.css.d.ts", false, dirAbsoluteBlock},
		{".vscode/**/*", ".vscode/**/*", false, dirFileLevelCover},                   // contains / → rooted (no **/ prefix)
		{"!.vscode/settings.json", "!.vscode/settings.json", true, dirAbsoluteBlock}, // rooted negation
		{"test-results/", "**/test-results/**/*", false, dirFileLevelCover},
		// rspack — rooted + nested + bracket + the wildcard-mid-path negations
		{"target/", "**/target/**/*", false, dirFileLevelCover},
		{"**/*.rs.bk", "**/*.rs.bk", false, dirAbsoluteBlock},
		{"report.[0-9]*.[0-9]*.[0-9]*.[0-9]*.json", "**/report.[0-9]*.[0-9]*.[0-9]*.[0-9]*.json", false, dirAbsoluteBlock},
		{"!scripts/node_modules/", "!scripts/node_modules/**/*", true, dirFileLevelCover}, // rooted (has /) dir negation
		{"!tests/rspack-test/*/**/node_modules", "!tests/rspack-test/*/**/node_modules", true, dirAbsoluteBlock},
		{"!packages/rspack-cli/tests/**/node_modules", "!packages/rspack-cli/tests/**/node_modules", true, dirAbsoluteBlock},
		{"!tests/rspack-test/configCases/target", "!tests/rspack-test/configCases/target", true, dirAbsoluteBlock},
		{"packages/*/tests/js", "packages/*/tests/js", false, dirAbsoluteBlock}, // rooted nested, no dir suffix
		{"/github/", "github/**/*", false, dirFileLevelCover},                   // rooted dir
		{"/artifacts/", "artifacts/**/*", false, dirFileLevelCover},
		{"npm/**/*.node", "npm/**/*.node", false, dirAbsoluteBlock}, // rooted ext filter under dir
		{"/npm/*", "npm/*", false, dirNone},                         // rooted single-level → NOT a subtree cover
		{"!npm/darwin-arm64/", "!npm/darwin-arm64/**/*", true, dirFileLevelCover},
		{"/packages/*/temp", "packages/*/temp", false, dirAbsoluteBlock},
		// rslint self — nested-rooted dir name (build/Release) stays absolute
		{"build/Release", "build/Release", false, dirAbsoluteBlock},
		{".env.*", "**/.env.*", false, dirAbsoluteBlock},
		{"!.env.example", "!**/.env.example", true, dirAbsoluteBlock},
		// create-rstack
		{".vscode/*", ".vscode/*", false, dirNone}, // contains / → rooted; single-level tail → dirNone
	}
	for _, c := range cases {
		gotGlob := convertSinglePattern(c.line, "")
		if gotGlob != c.glob {
			t.Errorf("convertSinglePattern(%q) = %q, want %q", c.line, gotGlob, c.glob)
			continue
		}
		// ParseIgnorePattern strips the leading `!` into Negated; its Glob is the
		// `!`-free matcher. Derive the expected stripped glob from c.glob.
		wantStrippedGlob := strings.TrimPrefix(c.glob, "!")
		p := ParseIgnorePattern(gotGlob)
		if p.Glob != wantStrippedGlob || p.Negated != c.negated || p.Kind != c.kind {
			t.Errorf("ParseIgnorePattern(%q) = {Glob:%q Negated:%v Kind:%d}, want {Glob:%q Negated:%v Kind:%d}",
				gotGlob, p.Glob, p.Negated, p.Kind, wantStrippedGlob, c.negated, c.kind)
		}
	}
}

// --- Real ESLint flat-config `ignores` strings, exact classification ---
//
// These are taken verbatim from the harvested rslint.config.ts / create-rstack /
// agent-skills configs and the documented idioms. They bypass the gitignore
// converter (ESLint configs author globs directly), so they exercise different
// raw forms than the gitignore lines above.
func TestRealConfigIgnore_Classify(t *testing.T) {
	cases := []struct {
		raw     string
		negated bool
		kind    dirKind
	}{
		{"**/dist/**", false, dirAbsoluteBlock},  // create-rstack
		{"**/tests/**", false, dirAbsoluteBlock}, // rspack config
		{"skills/**/scripts/*", false, dirNone},  // agent-skills: single-level tail → dirNone
		{"packages/rspack/src/runtime/moduleFederationDefaultRuntime.js", false, dirAbsoluteBlock},
		{"xtask/benchmark/benches/fixtures/rspack_sources/**", false, dirAbsoluteBlock},
		// documented idioms
		{"**/node_modules/**", false, dirAbsoluteBlock},
		{"**/*.test.ts", false, dirAbsoluteBlock},
		{"!**/*.integration.test.ts", true, dirAbsoluteBlock},
		{"packages/*/dist", false, dirAbsoluteBlock},
		{"packages/*/lib/**", false, dirAbsoluteBlock},
		{"**/__tests__/**", false, dirAbsoluteBlock},
		{"!.storybook", true, dirAbsoluteBlock},
		{"/public", false, dirAbsoluteBlock},
		{"!/public/keep", true, dirAbsoluteBlock},
		{"*.generated.*", false, dirAbsoluteBlock},
		{"apps/*/.next", false, dirAbsoluteBlock},
		{"{dist,build}/", false, dirAbsoluteBlock}, // brace + trailing slash → "/" tail, NOT "/*" nor "/**/*"
	}
	for _, c := range cases {
		p := ParseIgnorePattern(c.raw)
		if p.Negated != c.negated || p.Kind != c.kind {
			t.Errorf("ParseIgnorePattern(%q) = {Negated:%v Kind:%d}, want {Negated:%v Kind:%d}",
				c.raw, p.Negated, p.Kind, c.negated, c.kind)
		}
	}
}

// --- agent-skills `skills/**/scripts/*`: dirNone must NOT prune, but direct
// children ARE file-ignored (the `./*` regression class, deep variant) ---
//
// This is the realistic analogue of the fixed `./*` bug: a single-level tail on
// a deep pattern. canPruneDir must keep the scripts dir (a nested file under it
// is lintable), while isFileIgnored hides only the direct children.
func TestRealConfigIgnore_SkillsScriptsSingleLevel(t *testing.T) {
	pats := ParseIgnorePatterns([]string{"skills/**/scripts/*"})
	neg := buildNegReach(pats)

	// scripts dir must NOT be pruned (deep file under it stays lintable).
	if canPruneDir("skills/a/scripts", pats, neg) {
		t.Error(`"skills/**/scripts/*" must not prune dir skills/a/scripts (single-level tail)`)
	}
	if isDirAbsolutelyBlocked("skills/a/scripts", pats) {
		t.Error(`"skills/**/scripts/*" must not absolutely-block skills/a/scripts`)
	}

	// File-level: direct child ignored; nested-deeper child NOT ignored.
	cwd := "/repo"
	directChild := cwd + "/skills/a/scripts/build.ts"
	deeperChild := cwd + "/skills/a/scripts/sub/util.ts"
	assert.Assert(t, isFileIgnored(directChild, pats, cwd), "direct child of scripts/* must be ignored")
	assert.Assert(t, !isFileIgnored(deeperChild, pats, cwd), "nested-deeper child must NOT be ignored")

	// Linter authority agrees, end to end.
	cfg := RslintConfig{
		{Ignores: []string{"skills/**/scripts/*"}},
		{Files: []string{"**/*.ts"}, Rules: Rules{"r": "error"}},
	}
	assert.Assert(t, cfg.GetConfigForFile(directChild, cwd) == nil, "direct child not linted")
	assert.Assert(t, cfg.GetConfigForFile(deeperChild, cwd) != nil, "nested-deeper child must still be linted")
}

// --- rspack wildcard-mid-path negation: `!tests/rspack-test/*/**/node_modules`
// re-includes node_modules anywhere under tests/rspack-test/<pkg>/, so the
// node_modules cover must NOT prune the protected subtree, but a node_modules
// outside it stays pruned. literalSegmentPrefix truncates the negation reach at
// the first wildcard (`tests/rspack-test`). ---
func TestRealConfig_RspackNodeModulesReinclude(t *testing.T) {
	// Production-shaped set: node_modules cover + the two real re-includes.
	pats := ParseIgnorePatterns([]string{
		"**/node_modules/**/*",
		"!tests/rspack-test/*/**/node_modules", // wildcard mid-path
		"!scripts/node_modules/**/*",           // rooted literal
	})
	neg := buildNegReach(pats)

	// negReach literals: the wildcard one collapses to "tests/rspack-test".
	gotLits := neg.prefixes
	wantLits := []negPrefix{
		{literal: "tests/rspack-test"},
		{literal: "scripts/node_modules"},
	}
	assert.Equal(t, len(gotLits), len(wantLits), "negReach prefixes: %+v", gotLits)
	for i := range wantLits {
		assert.Equal(t, gotLits[i], wantLits[i], "negReach entry %d", i)
	}

	// A node_modules under the protected tests/rspack-test tree: parent must not
	// be pruned (segment-overlap with the negation literal).
	assert.Assert(t, !canPruneDir("tests/rspack-test", pats, neg),
		"tests/rspack-test must not be pruned (protected node_modules inside)")
	assert.Assert(t, !canPruneDir("scripts", pats, neg),
		"scripts must not be pruned (protected node_modules inside)")
	// An unrelated node_modules cover IS prunable (no negation reaches it).
	assert.Assert(t, canPruneDir("packages/core/node_modules", pats, neg),
		"unprotected node_modules must be pruned")
}

// --- rspack rooted ext-filter `npm/**/*.node` + `/npm/*` single-level +
// `!npm/<triple>/` re-include: the npm dir is kept (only .node files and direct
// children are ignored, re-included triples survive) ---
func TestRealConfig_RspackNpmArtifacts(t *testing.T) {
	pats := ParseIgnorePatterns([]string{
		"npm/**/*.node", // ext filter under dir → absoluteBlock, but does NOT block npm dir
		"npm/*",         // single-level → dirNone
		"!npm/darwin-arm64/**/*",
		"!npm/linux-x64-gnu/**/*",
	})
	neg := buildNegReach(pats)

	// npm dir must NOT be pruned: re-included triples + non-.node files survive.
	assert.Assert(t, !isDirAbsolutelyBlocked("npm", pats), "npm dir must not be absolutely blocked")
	assert.Assert(t, !canPruneDir("npm", pats, neg), "npm dir must not be pruned")

	cwd := "/repo"
	// .node artifact ignored; .js under a re-included triple is linted.
	assert.Assert(t, isFileIgnored(cwd+"/npm/foo/bar.node", pats, cwd), "*.node artifact ignored")
	assert.Assert(t, !isFileIgnored(cwd+"/npm/darwin-arm64/index.js", pats, cwd),
		"re-included triple .js must not be ignored")
	// Direct child of npm (npm/*) ignored; nested deeper not (unless re-included).
	assert.Assert(t, isFileIgnored(cwd+"/npm/index.js", pats, cwd), "npm/* direct child ignored")
}

// --- End-to-end on a real-shaped layout: a multi-source ignore set (rspack
// config `**/tests/**` + gitignore-converted dir covers + the wildcard-mid-path
// `!` re-include) must produce a gap-file set IDENTICAL to the linter's own
// per-file decision (GetConfigForFile != nil), regardless of directory pruning.
// Oracle = { f : f matches a files pattern ∧ GetConfigForFile(f) != nil }, the
// same contract as TestDiscoverGapFiles_PruningPreservesGapFiles but on a
// realistic merged ignore set rather than synthetic patterns. ---
func TestRealWorld_DiscoverGapFiles_MatchesLinterOracle(t *testing.T) {
	layout := []string{
		"packages/core/src/index.ts",                          // lintable
		"packages/core/dist/bundle.ts",                        // dist cover → ignored
		"packages/core/node_modules/dep/i.ts",                 // node_modules cover → ignored
		"target/build/a.ts",                                   // target cover → ignored
		"tests/rspack-test/configCases/pkg/node_modules/d.ts", // node_modules re-included, BUT under **/tests/** → still ignored
		"tests/rspack-test/configCases/c.ts",                  // under **/tests/** → ignored
		"scripts/build.ts",                                    // lintable
		"npm/darwin-arm64/index.ts",                           // re-included triple → lintable
		"npm/win32-x64-msvc/index.ts",                         // npm/* direct? no, nested → lintable (not covered)
		"npm/util.ts",                                         // npm/* direct child → ignored
		"src/app/main.ts",                                     // lintable
		"src/util.tsx",                                        // lintable (.tsx)
	}
	configDir, paths := setupDiscoveryFixture(t, layout)

	// Realistic merged ignore set: config global ignores + gitignore-converted forms.
	ignores := []string{
		"**/tests/**",                          // rspack rslint.config.ts (absolute dir block)
		"**/dist/**/*",                         // gitignore dist/
		"**/node_modules/**/*",                 // gitignore node_modules/
		"!tests/rspack-test/*/**/node_modules", // rspack gitignore re-include (wildcard mid-path)
		"**/target/**/*",                       // gitignore target/
		"npm/**/*.node",                        // gitignore npm/**/*.node
		"npm/*",                                // gitignore /npm/*
		"!npm/darwin-arm64/**/*",               // gitignore !npm/darwin-arm64/
	}
	filesPatterns := []string{"**/*.ts", "**/*.tsx"}
	config := RslintConfig{
		{Ignores: ignores},
		{Files: filesPatterns, Rules: Rules{"test-rule": "error"}},
	}

	// Oracle = the linter's own per-file decision over the fixture's files.
	var oracle []string
	for _, abs := range paths {
		if !isFileMatched(abs, filesPatterns, configDir) {
			continue
		}
		if config.GetConfigForFile(abs, configDir) != nil {
			oracle = append(oracle, abs)
		}
	}
	sort.Strings(oracle)

	got := DiscoverGapFiles(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, false)
	sort.Strings(got)

	// Walker must equal the linter oracle exactly (no over-prune, no over-include).
	assert.DeepEqual(t, got, oracle)

	// Pin the expected lintable set explicitly so a silent oracle shift can't
	// make this pass vacuously.
	want := []string{
		paths["npm/darwin-arm64/index.ts"],
		paths["npm/win32-x64-msvc/index.ts"],
		paths["packages/core/src/index.ts"],
		paths["scripts/build.ts"],
		paths["src/app/main.ts"],
		paths["src/util.tsx"],
	}
	sort.Strings(want)
	assert.DeepEqual(t, got, want)
}
