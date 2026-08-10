// TestOrderUpstreamTypescript01 migrates the upstream typescript cases from import-js/eslint-plugin-import at 1a39fb8672c83f8781c3a97f02eee00a7a2f0253.
// It preserves upstream messages, source positions, options, and iterative autofix outputs.
// Framework-only parser/resolver settings use rslint native equivalents; tsgo lock-ins live in order_extras_*_test.go.
package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderUpstreamTypescript01(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Upstream case 1:0.
			{Code: "\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 1:1.
			{Code: "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n\n              import index from './';\n            ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 1:2.
			{Code: "\n              import c from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ", Options: []any{map[string]any{"groups": []any{"external", "index", "type"}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 1:3.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { C } from 'dirA/Bar';\n              import type { A } from 'foo';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{"type"}}}},
			// Upstream case 1:4.
			{Code: "\n              import c from 'Bar';\n              import type { A } from 'foo';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n\n              import index from './';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}}}},
			// Upstream case 1:5.
			{Code: "\n              import a from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ", Options: []any{map[string]any{"groups": []any{"external", "index", "type"}, "alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 1:6.
			{Code: "\n              import { Partner } from '@models/partner/partner';\n              import { PartnerId } from '@models/partner/partner-id';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 1:7.
			{Code: "\n              import { serialize, parse, mapFieldErrors } from '@vtaits/form-schema';\n              import type { GetFieldSchema } from '@vtaits/form-schema';\n              import { useMemo, useCallback } from 'react';\n              import type { ReactElement, ReactNode } from 'react';\n              import { Form } from 'react-final-form';\n              import type { FormProps as FinalFormProps } from 'react-final-form';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 1:8.
			{Code: "\n              import type { CopyOptions } from 'fs';\n              import type { ParsedPath } from 'path';\n\n              declare module 'my-module' {\n                import type { CopyOptions } from 'fs';\n                import type { ParsedPath } from 'path';\n              }\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 1:9, 2:9.
			{Code: "\n              import { useLazyQuery, useQuery } from \"@apollo/client\";\n              import { useEffect } from \"react\";\n            ", Options: []any{map[string]any{"groups": []any{"builtin", "external", "internal", "parent", "sibling", "index", "object", "type"}, "pathGroups": []any{map[string]any{"pattern": "react", "group": "external", "position": "before"}}, "newlines-between": "always", "alphabetize": map[string]any{"order": "asc", "caseInsensitive": true}}}},
			// Upstream case 1:10, 2:10.
			{Code: "\n                import express from 'express';\n                import log4js from 'log4js';\n                import chpro from 'node:child_process';\n                // import fsp from 'node:fs/promises';\n              ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "external", "internal", "parent", "sibling", "index", "object", "type"}}}}},
			// Upstream case 1:11.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}}}},
			// Upstream case 1:12.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}, "sortTypesGroup": false}}},
			// Upstream case 1:13.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n              import type { A } from 'foo';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{"type"}, "sortTypesGroup": true}}},
			// Upstream case 1:14.
			{Code: "\n              import c from 'Bar';\n              import type { AA } from 'abc';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}}}},
			// Upstream case 1:15.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}, "sortTypesGroup": true}}},
			// Upstream case 1:16.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "never", "newlines-between-types": "always", "pathGroupsExcludedImportTypes": []any{}, "sortTypesGroup": true}}},
			// Upstream case 1:17.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "always", "newlines-between-types": "never", "pathGroupsExcludedImportTypes": []any{}, "sortTypesGroup": true}}},
			// Upstream case 1:18.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA } from 'abc';\n\n              import type { A } from 'foo';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "never", "newlines-between-types": "ignore", "pathGroupsExcludedImportTypes": []any{}, "sortTypesGroup": true}}},
			// Upstream case 1:19.
			{Code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n\n              import type { D } from 'dirA/bar';\n            ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}, "groups": []any{"external", "internal", "index", "type"}, "pathGroups": []any{map[string]any{"pattern": "dirA/**", "group": "internal"}}, "newlines-between": "never", "newlines-between-types": "always-and-inside-groups", "pathGroupsExcludedImportTypes": []any{}, "sortTypesGroup": true}}},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
