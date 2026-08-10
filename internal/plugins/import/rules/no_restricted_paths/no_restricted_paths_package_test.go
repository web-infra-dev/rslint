// TestNoRestrictedPathsPackageSpecifiers runs every module reference shape
// against every zone shape that can name an installed package. Zones reaching
// into node_modules need a fixture root the embedded tree cannot carry, so this
// file builds its own.
//
// The expectations mirror a run of eslint-plugin-import over the same tree,
// except for the one divergence the matrix is built to pin down: the resolvers
// disagree on which file inside a package a specifier names. Upstream's node
// resolver follows `main` to `index.js`; TypeScript follows `types` to
// `index.d.ts`. Zones naming the package, an ancestor of it or a glob over it
// cover the import either way, and only zones naming a single file inside the
// package can tell the two apart.
package no_restricted_paths_test

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_restricted_paths"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

const packageRootDir = "/restricted-package"

// packageRoot builds a fixture root holding one installed package whose `main`
// and `types` name different files, so a zone can name either one.
func packageRoot() rule_tester.Root {
	files := map[string]string{
		packageRootDir + "/tsconfig.json":                          `{"compilerOptions":{"module":"commonjs","target":"esnext","allowJs":true,"esModuleInterop":true}}`,
		packageRootDir + "/node_modules/some-package/package.json": `{"name":"some-package","version":"1.0.0","main":"index.js","types":"index.d.ts"}`,
		packageRootDir + "/node_modules/some-package/index.d.ts":   "declare const value: number;\nexport = value;\n",
		packageRootDir + "/node_modules/some-package/index.js":     "module.exports = 1;\n",
		packageRootDir + "/client/a.ts":                            "",
		packageRootDir + "/client/consumer.js":                     "",
	}

	return rule_tester.Root{
		Dir: packageRootDir,
		FS:  rslint_utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), files),
	}
}

// packageShape is one way of writing a module reference to the package.
type packageShape struct {
	name      string
	code      string
	fileName  string
	specifier string
	// column is where the specifier literal starts, and 0 marks a shape that is
	// not a module reference at all, so no zone can ever restrict it.
	column int
}

// reachingShapes name the package through a specifier the rule resolves.
var reachingShapes = []packageShape{
	{name: "default import", code: `import x from "some-package"`, fileName: "client/a.ts", specifier: "some-package", column: 15},
	{name: "side effect import", code: `import "some-package"`, fileName: "client/a.ts", specifier: "some-package", column: 8},
	{name: "star re-export", code: `export * from "some-package"`, fileName: "client/a.ts", specifier: "some-package", column: 15},
	{name: "dynamic import", code: `const p = import("some-package")`, fileName: "client/a.ts", specifier: "some-package", column: 18},
	{name: "require in TypeScript", code: `const x = require("some-package")`, fileName: "client/a.ts", specifier: "some-package", column: 19},
	{name: "parenthesized require", code: `const x = (require)("some-package")`, fileName: "client/a.ts", specifier: "some-package", column: 21},
	{name: "require of a subpath", code: `const x = require("some-package/index.js")`, fileName: "client/a.ts", specifier: "some-package/index.js", column: 19},
	{name: "require in JavaScript", code: `const x = require("some-package")`, fileName: "client/consumer.js", specifier: "some-package", column: 19},
}

// silentShapes never reach a zone, whichever files it names.
var silentShapes = []packageShape{
	// Neither upstream's module visitor nor this one reads an import-equals
	// declaration as a module reference.
	{name: "import equals require", code: `import x = require("some-package"); export { x };`, fileName: "client/a.ts"},
	// A `require` call is a module reference only with exactly one argument.
	{name: "require with an extra argument", code: `const x = require("some-package", 1)`, fileName: "client/a.ts"},
	// An unresolvable specifier is skipped before any zone is consulted.
	{name: "require of a missing package", code: `const x = require("does-not-exist")`, fileName: "client/a.ts"},
}

// coveringZones name the package in a way that covers it under either
// resolver, so every reaching shape is restricted.
var coveringZones = []struct {
	name string
	from string
}{
	{name: "the package directory", from: "./node_modules/some-package"},
	{name: "an ancestor of the package", from: "./node_modules"},
	{name: "a glob over the package", from: "./node_modules/some-package/*"},
	// Upstream's resolver never reaches `index.d.ts`, so upstream reports
	// nothing for this zone.
	{name: "the file TypeScript resolves to", from: "./node_modules/some-package/index.d.ts"},
}

// missingZones name a file no specifier resolves to here, so nothing is
// restricted.
var missingZones = []struct {
	name string
	from string
}{
	// Upstream's resolver follows `main` here and reports every reaching shape.
	{name: "the file upstream's resolver picks", from: "./node_modules/some-package/index.js"},
	{name: "a package that is not installed", from: "./node_modules/other-package"},
}

func TestNoRestrictedPathsPackageSpecifiers(t *testing.T) {
	valid := make([]rule_tester.ValidTestCase, 0, len(silentShapes)*(len(coveringZones)+len(missingZones))+len(reachingShapes)*len(missingZones)+len(reachingShapes))
	invalid := make([]rule_tester.InvalidTestCase, 0, len(reachingShapes)*len(coveringZones))

	zoneFor := func(from string) any {
		return zones(map[string]interface{}{"target": "./client", "from": from})
	}

	for _, zone := range coveringZones {
		for _, shape := range reachingShapes {
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code:     shape.code,
				FileName: shape.fileName,
				Options:  zoneFor(zone.from),
				Errors:   []rule_tester.InvalidTestCaseError{unexpectedPath(shape.specifier, 1, shape.column)},
			})
		}
		for _, shape := range silentShapes {
			valid = append(valid, rule_tester.ValidTestCase{
				Code:     shape.code,
				FileName: shape.fileName,
				Options:  zoneFor(zone.from),
			})
		}
	}

	for _, zone := range missingZones {
		for _, shape := range append(append([]packageShape{}, reachingShapes...), silentShapes...) {
			valid = append(valid, rule_tester.ValidTestCase{
				Code:     shape.code,
				FileName: shape.fileName,
				Options:  zoneFor(zone.from),
			})
		}
	}

	// An `except` naming the package lifts the restriction an ancestor zone
	// would otherwise impose, the same way upstream reports nothing here.
	for _, shape := range append(append([]packageShape{}, reachingShapes...), silentShapes...) {
		valid = append(valid, rule_tester.ValidTestCase{
			Code:     shape.code,
			FileName: shape.fileName,
			Options: zones(map[string]interface{}{
				"target": "./client",
				"from":   "./node_modules",
				"except": list("./some-package"),
			}),
		})
	}

	rule_tester.RunRuleTester(packageRoot(), "tsconfig.json", t, &no_restricted_paths.NoRestrictedPathsRule, valid, invalid)
}
