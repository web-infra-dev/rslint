package config

import (
	"strings"
	"testing"

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

// --- Converted gitignore globs, exact classification ---
//
// Conversion itself is covered in the gitignore package. These assertions pin
// the config-side Kind boundary that drives directory pruning.
func TestConvertedGitignoreGlob_Classify(t *testing.T) {
	cases := []struct {
		glob    string
		negated bool
		kind    dirKind
	}{
		{"**/node_modules/**/*", false, dirFileLevelCover},
		{"**/dist/**/*", false, dirFileLevelCover},
		{"**/dist-*", false, dirAbsoluteBlock},
		{"**/*.log*", false, dirAbsoluteBlock},
		{"**/*.css.d.ts", false, dirAbsoluteBlock},
		{".vscode/**/*", false, dirFileLevelCover},
		{"!.vscode/settings.json", true, dirAbsoluteBlock},
		{"**/test-results/**/*", false, dirFileLevelCover},
		{"**/output/**/*", false, dirFileLevelCover},
		{"**/*.rs.bk", false, dirAbsoluteBlock},
		{"**/report.[0-9]*.[0-9]*.[0-9]*.[0-9]*.json", false, dirAbsoluteBlock},
		{"!scripts/node_modules/**/*", true, dirFileLevelCover},
		{"!tests/fixtures/*/**/node_modules", true, dirAbsoluteBlock},
		{"!packages/tool/tests/**/node_modules", true, dirAbsoluteBlock},
		{"!tests/fixtures/cases/output", true, dirAbsoluteBlock},
		{"packages/*/tests/js", false, dirAbsoluteBlock},
		{"github/**/*", false, dirFileLevelCover},
		{"artifacts/**/*", false, dirFileLevelCover},
		{"npm/**/*.node", false, dirAbsoluteBlock},
		{"npm/*", false, dirNone},
		{"!npm/darwin-arm64/**/*", true, dirFileLevelCover},
		{"packages/*/temp", false, dirAbsoluteBlock},
		{"build/Release", false, dirAbsoluteBlock},
		{"**/.env.*", false, dirAbsoluteBlock},
		{"!**/.env.example", true, dirAbsoluteBlock},
		{".vscode/*", false, dirNone},
	}
	for _, c := range cases {
		// ParseIgnorePattern strips the leading `!` into Negated; its Glob is the
		// `!`-free matcher. Derive the expected stripped glob from c.glob.
		wantStrippedGlob := strings.TrimPrefix(c.glob, "!")
		p := ParseIgnorePattern(c.glob)
		if p.Glob != wantStrippedGlob || p.Negated != c.negated || p.Kind != c.kind {
			t.Errorf("ParseIgnorePattern(%q) = {Glob:%q Negated:%v Kind:%d}, want {Glob:%q Negated:%v Kind:%d}",
				c.glob, p.Glob, p.Negated, p.Kind, wantStrippedGlob, c.negated, c.kind)
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

	// Gitignore matching is case-insensitive on the default macOS/Windows
	// filesystems. The anchored re-includes must still protect their own paths
	// without globally disabling pruning for every unrelated node_modules tree.
	caseInsensitive := append([]IgnorePattern(nil), pats...)
	for i := range caseInsensitive {
		caseInsensitive[i].CaseInsensitive = true
	}
	caseInsensitiveNeg := buildNegReach(caseInsensitive)
	assert.Assert(t, !canPruneDir("TESTS/RSPACK-TEST", caseInsensitive, caseInsensitiveNeg),
		"case-folded protected subtree must not be pruned")
	assert.Assert(t, !canPruneDir("SCRIPTS", caseInsensitive, caseInsensitiveNeg),
		"case-folded scripts subtree must not be pruned")
	assert.Assert(t, canPruneDir("PACKAGES/CORE/NODE_MODULES", caseInsensitive, caseInsensitiveNeg),
		"case-insensitive unrelated node_modules must remain prunable")
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
