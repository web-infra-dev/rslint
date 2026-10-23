package require_artifact_dependency_entry

import (
	"fmt"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func artifactRoot() rule_tester.Root {
	base := fixtures.GetRootDir()
	files := map[string]string{}
	for name, content := range map[string]string{
		"tsconfig.rslim-artifact.json": `{"compilerOptions":{"target":"esnext","module":"esnext","moduleResolution":"bundler","allowJs":true,"noCheck":true,"types":[]},"include":["artifact/**/*"]}`,
		"tsconfig.rslim-source.json":   `{"compilerOptions":{"target":"esnext","module":"esnext","moduleResolution":"bundler","allowJs":true,"types":[]},"include":["artifact/**/*"]}`,
		"artifact/dep.ts":              `export function needed() { return 1; }`,
		"artifact/paired.d.ts":         `export declare function g(): number;`,
		"artifact/paired.js":           `export function g() { return 1; }`,
	} {
		files[tspath.ResolvePath(base.Dir, name)] = content
	}
	return rule_tester.Root{Dir: base.Dir, FS: utils.NewOverlayVFS(base.FS, files)}
}

func projectReferenceRoot() rule_tester.Root {
	base := artifactRoot()
	files := map[string]string{}
	for name, content := range map[string]string{
		"artifact/tsconfig.refs.json":  `{"compilerOptions":{"target":"esnext","module":"esnext","moduleResolution":"bundler","noCheck":true,"composite":true,"types":[]},"references":[{"path":"../source-built"},{"path":"../artifact-built"}],"include":["input.ts","dep.ts"]}`,
		"source-built/tsconfig.json":   `{"compilerOptions":{"target":"esnext","module":"esnext","moduleResolution":"bundler","composite":true,"declaration":true,"outDir":"dist","rootDir":".","types":[]},"include":["dep.ts"]}`,
		"source-built/dep.ts":          `export function needed() { return 1; }`,
		"source-built/dist/dep.d.ts":   `export declare function needed(): number;`,
		"artifact-built/tsconfig.json": `{"compilerOptions":{"target":"esnext","module":"esnext","moduleResolution":"bundler","noCheck":true,"composite":true,"declaration":true,"outDir":"dist","rootDir":".","types":[]},"include":["dep.ts"]}`,
		"artifact-built/dep.ts":        `export function needed() { return 1; }`,
		"artifact-built/dist/dep.d.ts": `export declare function needed(): number;`,
	} {
		files[tspath.ResolvePath(base.Dir, name)] = content
	}
	return rule_tester.Root{Dir: base.Dir, FS: utils.NewOverlayVFS(base.FS, files)}
}

func TestRequireArtifactDependencyEntryNoCheckLocal(t *testing.T) {
	rule_tester.RunRuleTester(artifactRoot(), "tsconfig.rslim-artifact.json", t, &RequireArtifactDependencyEntryRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { needed } from './dep';`, FileName: "artifact/input.ts"},
			{Code: `export { needed } from './dep';`, FileName: "artifact/input.ts"},
			{Code: `const dep = require('./dep');`, FileName: "artifact/input.ts"},
			{Code: `const dep = import('./dep');`, FileName: "artifact/input.ts"},
			{Code: `import { needed } from './dep'; export function g() { return needed(); }`, FileName: "artifact/paired.js"},
			{Code: `import './unresolved';`, FileName: "artifact/input.ts"},
		}, nil)
}

func TestRequireArtifactDependencyEntryProjectReference(t *testing.T) {
	root := projectReferenceRoot()
	rule_tester.RunRuleTester(root, "artifact/tsconfig.refs.json", t, &RequireArtifactDependencyEntryRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { needed } from './dep';`, FileName: "artifact/input.ts"},
			{Code: `import { needed } from '../artifact-built/dep';`, FileName: "artifact/input.ts"},
			{Code: `import type { needed } from '../source-built/dep';`, FileName: "artifact/input.ts"},
			{Code: `export type { needed } from '../source-built/dep';`, FileName: "artifact/input.ts"},
			{Code: `type Dep = typeof import('../source-built/dep');`, FileName: "artifact/input.ts"},
			{Code: `const dep = import('../source-built/dep');`, FileName: "artifact/input.ts"},
			{Code: `const dep = require('../source-built/dep');`, FileName: "artifact/input.ts"},
			{Code: `import dep = require('../source-built/dep');`, FileName: "artifact/input.ts"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `import { needed as use } from '../source-built/dep';`, FileName: "artifact/input.ts", Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "artifactDependency",
				Message:   fmt.Sprintf("This noCheck TypeScript module imports %q, which resolves to source-built module %q. Candidate runtime uses: needed. Rslim may not see references across this artifact boundary. Review the emitted JavaScript and dependent source modules; mark required symbols with @entry, or add source files to Rslim entries in Lib Mode.", "../source-built/dep", tspath.ResolvePath(root.Dir, "source-built/dep.ts")),
			}}},
			{Code: `export { needed } from '../source-built/dep';`, FileName: "artifact/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "artifactDependency"}}},
		},
	)
}

func TestRequireArtifactDependencyEntryDeclarationPair(t *testing.T) {
	rule_tester.RunRuleTester(artifactRoot(), "tsconfig.rslim-source.json", t, &RequireArtifactDependencyEntryRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { needed } from './dep';`, FileName: "artifact/input.ts"},
			{Code: `import { g } from './paired.js';`, FileName: "artifact/input.ts"},
			{Code: `export function g() { return 1; }`, FileName: "artifact/unpaired.js"},
			{Code: `export function g() { return 1; }`, FileName: "artifact/paired.js"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `import { needed } from './dep'; export function g() { return needed(); }`, FileName: "artifact/paired.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "artifactDependency"}}},
		},
	)
}
