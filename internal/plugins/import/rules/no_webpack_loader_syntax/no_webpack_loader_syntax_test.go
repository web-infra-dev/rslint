package no_webpack_loader_syntax_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_webpack_loader_syntax"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoWebpackLoaderSyntaxRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_webpack_loader_syntax.NoWebpackLoaderSyntax,
		[]rule_tester.ValidTestCase{
			{Code: `import _ from "lodash"`},
			{Code: `import find from "lodash.find"`},
			{Code: `import foo from "./foo.css"`},
			{Code: `import data from "@scope/my-package/data.json"`},
			{Code: `var _ = require("lodash")`},
			{Code: `var find = require("lodash.find")`},
			{Code: `var foo = require("./foo")`},
			{Code: `var foo = require("../foo")`},
			{Code: `var foo = require("foo")`},
			{Code: `var foo = require("./")`},
			{Code: `var foo = require("@scope/foo")`},
			{Code: `var foo = fn("babel!lodash")`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `import _ from "babel!lodash"`,
				FileName: "foo.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "import/no-webpack-loader-syntax",
					},
				},
			},
			{
				Code:     `import find from "-babel-loader!lodash.find"`,
				FileName: "foo.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "import/no-webpack-loader-syntax",
					},
				},
			},
			{
				Code:     `import foo from "style!css!./foo.css"`,
				FileName: "foo.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "import/no-webpack-loader-syntax",
					},
				},
			},
			{
				Code:     `import data from "json!@scope/my-package/data.json"`,
				FileName: "foo.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "import/no-webpack-loader-syntax",
					},
				},
			},
			{
				Code:     `var _ = require("babel!lodash")`,
				FileName: "foo.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "import/no-webpack-loader-syntax",
					},
				},
			},
			{
				Code:     `var find = require("-babel-loader!lodash.find")`,
				FileName: "foo.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "import/no-webpack-loader-syntax",
					},
				},
			},
			{
				Code:     `var foo = require("style!css!./foo.css")`,
				FileName: "foo.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "import/no-webpack-loader-syntax",
					},
				},
			},
			{
				Code:     `var data = require("json!@scope/my-package/data.json")`,
				FileName: "foo.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "import/no-webpack-loader-syntax",
					},
				},
			},
		},
	)
}
