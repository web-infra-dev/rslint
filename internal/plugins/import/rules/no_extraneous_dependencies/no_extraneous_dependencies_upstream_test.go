package no_extraneous_dependencies_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	. "github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_extraneous_dependencies"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func dependencyRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = "/fixture"
	archive := txtarfs.MustParseFile(t, "testdata/modules.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("empty fixture")
	}
	files := map[string]string{}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(root.Dir, name)] = string(data)
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	return root
}

// All semantic cases from v2.32.0, including TypeScript groups. Parser-version
// repetitions collapse into one native parse. Flow typeof and executable webpack
// resolver cases remain explained skips in upstream-cases.json. The docs label
// react, relative includeInternal imports and quoted @generated/foo as failures;
// those cases retain the pinned implementation's accepting behavior instead.
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-extraneous-dependencies.js
func TestUpstream(t *testing.T) {
	data, err := os.ReadFile("testdata/upstream-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var suites []struct {
		Name    string                        `json:"name"`
		Valid   []rule_tester.ValidTestCase   `json:"valid"`
		Invalid []rule_tester.InvalidTestCase `json:"invalid"`
	}
	if err = json.Unmarshal(data, &suites); err != nil {
		t.Fatal(err)
	}
	for i, suite := range suites {
		t.Run(fmt.Sprintf("group-%d-%s", i+1, suite.Name), func(t *testing.T) {
			rule_tester.RunRuleTester(dependencyRoot(t), "tsconfig.json", t, &NoExtraneousDependenciesRule, suite.Valid, suite.Invalid)
		})
	}
}
