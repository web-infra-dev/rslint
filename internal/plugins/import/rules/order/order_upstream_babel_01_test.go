// TestOrderUpstreamBabel01 migrates the upstream babel cases from import-js/eslint-plugin-import at 1a39fb8672c83f8781c3a97f02eee00a7a2f0253.
// It preserves upstream messages, source positions, options, and iterative autofix outputs.
// Framework-only parser/resolver settings use rslint native equivalents; tsgo lock-ins live in order_extras_*_test.go.
package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderUpstreamBabel01(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Upstream case 1:9, 2:9.
			{Code: "\n              import { useLazyQuery, useQuery } from \"@apollo/client\";\n              import { useEffect } from \"react\";\n            ", Options: []any{map[string]any{"groups": []any{"builtin", "external", "internal", "parent", "sibling", "index", "object", "type"}, "pathGroups": []any{map[string]any{"pattern": "react", "group": "external", "position": "before"}}, "newlines-between": "always", "alphabetize": map[string]any{"order": "asc", "caseInsensitive": true}}}},
			// Upstream case 1:10, 2:10.
			{Code: "\n                import express from 'express';\n                import log4js from 'log4js';\n                import chpro from 'node:child_process';\n                // import fsp from 'node:fs/promises';\n              ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "external", "internal", "parent", "sibling", "index", "object", "type"}}}}},
			// Upstream case 2:0.
			{Code: "\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 2:1.
			{Code: "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n\n              import index from './';\n            ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 2:2.
			{Code: "\n              import c from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ", Options: []any{map[string]any{"groups": []any{"external", "index", "type"}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 2:3.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { C } from 'dirA/Bar';\n              import type { A } from 'foo';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{"type"}}}},
			// Upstream case 2:4.
			{Code: "\n              import c from 'Bar';\n              import type { A } from 'foo';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n\n              import index from './';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}}}},
			// Upstream case 2:5.
			{Code: "\n              import a from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ", Options: []any{map[string]any{"groups": []any{"external", "index", "type"}, "alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 2:6.
			{Code: "\n              import { Partner } from '@models/partner/partner';\n              import { PartnerId } from '@models/partner/partner-id';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 2:7.
			{Code: "\n              import { serialize, parse, mapFieldErrors } from '@vtaits/form-schema';\n              import type { GetFieldSchema } from '@vtaits/form-schema';\n              import { useMemo, useCallback } from 'react';\n              import type { ReactElement, ReactNode } from 'react';\n              import { Form } from 'react-final-form';\n              import type { FormProps as FinalFormProps } from 'react-final-form';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 2:8.
			{Code: "\n              import type { CopyOptions } from 'fs';\n              import type { ParsedPath } from 'path';\n\n              declare module 'my-module' {\n                import type { CopyOptions } from 'fs';\n                import type { ParsedPath } from 'path';\n              }\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 2:11.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}}}},
			// Upstream case 2:12.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}, "sortTypesGroup": false}}},
			// Upstream case 2:13.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n              import type { A } from 'foo';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{"type"}, "sortTypesGroup": true}}},
			// Upstream case 2:14.
			{Code: "\n              import c from 'Bar';\n              import type { AA } from 'abc';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}}}},
			// Upstream case 2:15.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}, "sortTypesGroup": true}}},
			// Upstream case 2:16.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "never", "newlines-between-types": "always", "pathGroupsExcludedImportTypes": []any{}, "sortTypesGroup": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream case 1:4, 2:4.
			{Code: "\n              import './local1';\n              import global from 'global1';\n              import local from './local2';\n              import 'global2';\n            ", Options: []any{map[string]any{"warnOnUnassignedImports": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`global1` import should occur before import of `./local1`", Line: 3, Column: 15, EndLine: 3, EndColumn: 44}, {MessageId: "order", Message: "`global2` import should occur before import of `./local1`", Line: 5, Column: 15, EndLine: 5, EndColumn: 32}}},
			// Upstream case 1:5, 2:5.
			{Code: "\n              import local from './local';\n\n              import 'global1';\n\n              import global2 from 'global2';\n              import global3 from 'global3';\n            ", Options: []any{map[string]any{"warnOnUnassignedImports": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./local` import should occur after import of `global3`", Line: 2, Column: 15, EndLine: 2, EndColumn: 43}}},
			// Upstream case 1:27, 2:24.
			{Code: "\n                import express from 'express';\n                import log4js from 'log4js';\n                import chpro from 'node:child_process';\n                // import fsp from 'node:fs/promises';\n              ", Options: []any{map[string]any{"groups": []any{"builtin", "external", "internal", "parent", "sibling", "index", "object", "type"}}}, Output: []string{"\n                import chpro from 'node:child_process';\n                import express from 'express';\n                import log4js from 'log4js';\n                // import fsp from 'node:fs/promises';\n              "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`node:child_process` import should occur before import of `express`", Line: 4, Column: 17, EndLine: 4, EndColumn: 56}}},
		},
	)
}
