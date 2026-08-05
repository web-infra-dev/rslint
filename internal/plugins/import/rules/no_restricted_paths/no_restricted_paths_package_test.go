// TestNoRestrictedPathsPackageSpecifiers exercises zones that reach into
// node_modules, which the embedded fixture tree cannot carry, so this file
// builds its own fixture root instead. It covers both halves of what a package
// specifier does here: that it reaches a zone at all, and which file inside the
// package it resolves to, since that differs from upstream's resolver.
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

// packageRoot builds a fixture root holding one installed package. `main` and
// `types` name different files on purpose: upstream's node resolver follows
// `main` to `index.js`, while TypeScript's resolution follows `types` to
// `index.d.ts`, and the rule matches zones against whichever file that is.
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

func TestNoRestrictedPathsPackageSpecifiers(t *testing.T) {
	rule_tester.RunRuleTester(
		packageRoot(),
		"tsconfig.json",
		t,
		&no_restricted_paths.NoRestrictedPathsRule,
		[]rule_tester.ValidTestCase{
			// ---- Differences from ESLint: the resolvers disagree on which file a
			// package specifier names. Upstream's node resolver follows `main`, so a
			// zone naming `index.js` restricts the import there; TypeScript follows
			// `types`, so the same zone matches nothing here ----
			{
				Code:     `import value from "some-package"`,
				FileName: "client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./client",
					"from":   "./node_modules/some-package/index.js",
				}),
			},
			{
				Code:     `const value = require("some-package")`,
				FileName: "client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./client",
					"from":   "./node_modules/some-package/index.js",
				}),
			},

			// A package outside every zone stays allowed.
			{
				Code:     `const value = require("some-package")`,
				FileName: "client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./client",
					"from":   "./node_modules/other-package",
				}),
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- A zone naming the package directory covers it whichever file the
			// specifier resolves to, which is the shape a real config uses ----
			{
				Code:     `import value from "some-package"`,
				FileName: "client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./client",
					"from":   "./node_modules/some-package",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("some-package", 1, 19)},
			},

			// ---- TypeScript records no resolution for a `require()` call in a
			// TypeScript file, and a package specifier carries no relative path to
			// probe, so this one reaches the package through the module resolver ----
			{
				Code:     `const value = require("some-package")`,
				FileName: "client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./client",
					"from":   "./node_modules/some-package",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("some-package", 1, 23)},
			},
			{
				Code:     `const value = (require)("some-package")`,
				FileName: "client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./client",
					"from":   "./node_modules/some-package",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("some-package", 1, 25)},
			},
			{
				Code:     `const value = require("some-package")`,
				FileName: "client/consumer.js",
				Options: zones(map[string]interface{}{
					"target": "./client",
					"from":   "./node_modules/some-package",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("some-package", 1, 23)},
			},

			// ---- Differences from ESLint, the other half: a zone naming the file
			// TypeScript resolves to does restrict the import, where upstream's
			// resolver never reaches `index.d.ts` and reports nothing ----
			{
				Code:     `import value from "some-package"`,
				FileName: "client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./client",
					"from":   "./node_modules/some-package/index.d.ts",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("some-package", 1, 19)},
			},
		},
	)
}
