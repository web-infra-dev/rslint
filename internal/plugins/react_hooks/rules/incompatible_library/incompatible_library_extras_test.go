package incompatible_library

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestIncompatibleLibraryExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
func TestIncompatibleLibraryExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 4: access/key forms; non-watch react-hook-form member is not incompatible. ----
		{Code: `
import * as RHF from 'react-hook-form';
function Form() {
  const form = RHF.useForm();
  form.register('name');
  return <form />;
}
		`, FileName: "form.tsx", Tsx: true},
		// ---- Dimension 4: declaration/container forms; a locally shadowed import alias is not reported. ----
		{Code: `
import { useReactTable } from '@tanstack/react-table';
function Table(useReactTable) {
  const table = useReactTable();
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: access/key forms; dynamic property names are not treated as known incompatible APIs. ----
		{Code: `
import { useForm } from 'react-hook-form';
function Form(name) {
  const form = useForm();
  const value = form[name]('field');
  return <div>{value}</div>;
}
		`, FileName: "form.tsx", Tsx: true},
		// ---- Real-user: TanStack Table issue #5141 stable options discussion; local non-imported hook name is ignored. ----
		{Code: `
function Table({ data, columns }) {
  const useReactTable = (config) => config;
  const table = useReactTable({ data, columns });
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Real-user: React issue #33057; non-hook table helpers are not reported. ----
		{Code: `
import { flexRender } from '@tanstack/react-table';
function Table({ cell }) {
  return <td>{flexRender(cell.column.columnDef.cell, cell.getContext())}</td>;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: nesting/traversal boundaries; later function declarations shadow imports for the whole block. ----
		{Code: `
import { useReactTable } from '@tanstack/react-table';
function Table({ data, columns }) {
  const table = useReactTable({ data, columns });
  function useReactTable(config) {
    return config;
  }
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: nesting/traversal boundaries; var declarations hoist from nested blocks. ----
		{Code: `
import { useReactTable } from '@tanstack/react-table';
function Table({ data, columns, condition }) {
  const table = useReactTable({ data, columns });
  if (condition) {
    var useReactTable = (config) => config;
  }
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: declaration/container forms; default imports do not imply named incompatible exports. ----
		{Code: `
import useReactTable from '@tanstack/react-table';
function Table({ data, columns }) {
  const table = useReactTable({ data, columns });
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: declaration/container forms; CommonJS require is not a source tracked by upstream. ----
		{Code: `
const TableLib = require('@tanstack/react-table');
function Table({ data, columns }) {
  const table = TableLib.useReactTable({ data, columns });
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: access/key forms; TypeScript import-equals mirrors upstream's non-reporting behavior. ----
		{Code: `
import TableLib = require('@tanstack/react-table');
function Table({ data, columns }) {
  const table = TableLib.useReactTable({ data, columns });
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: declaration/container forms; top-level uses are outside React Compiler targets. ----
		{Code: `
import { useForm } from 'react-hook-form';
const form = useForm();
form.watch('name');
		`, FileName: "form.tsx", Tsx: true},
		// ---- Dimension 4: nesting/traversal boundaries; top-level derived form bindings do not flow into components. ----
		{Code: `
import { useForm } from 'react-hook-form';
const form = useForm();
function Form() {
  return <div>{form.watch('name')}</div>;
}
		`, FileName: "form.tsx", Tsx: true},
		// ---- Dimension 4: nesting/traversal boundaries; top-level aliases do not flow into components. ----
		{Code: `
import { useReactTable } from '@tanstack/react-table';
const createTable = useReactTable;
function Table({ data, columns }) {
  const table = createTable({ data, columns });
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: nesting/traversal boundaries; top-level namespace destructuring does not flow into components. ----
		{Code: `
import * as TableLib from '@tanstack/react-table';
const { useReactTable } = TableLib;
function Table({ data, columns }) {
  const table = useReactTable({ data, columns });
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: nesting/traversal boundaries; top-level callbacks are outside compiler roots. ----
		{Code: `
import { useMemo } from 'react';
import { useForm } from 'react-hook-form';
const form = useForm();
const name = useMemo(() => form.watch('name'), [form]);
		`, FileName: "form.tsx", Tsx: true},
		// ---- Dimension 4: declaration/container forms; lowercase functions are not compiler roots even when returning JSX. ----
		{Code: `
import { useReactTable } from '@tanstack/react-table';
function table({ data, columns }) {
  const table = useReactTable({ data, columns });
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: declaration/container forms; top-level memo callbacks are not roots for this diagnostic. ----
		{Code: `
import { memo } from 'react';
import { useReactTable } from '@tanstack/react-table';
const Table = memo(function Table({ data, columns }) {
  const table = useReactTable({ data, columns });
  return <Grid table={table} />;
});
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: declaration/container forms; aliased incompatible calls do not make a PascalCase null-returning function a root. ----
		{Code: `
import { useReactTable as createTable } from '@tanstack/react-table';
function Table({ data, columns }) {
  const table = createTable({ data, columns });
  return null;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- Dimension 4: declaration/container forms; nested destructuring from watch does not preserve function identity. ----
		{Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const { watch: { current } } = useForm();
  return <div>{current()}</div>;
}
		`, FileName: "form.tsx", Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Dimension 4: receiver wrappers; parenthesized namespace receiver is transparent. ----
		// Locks in upstream defaultModuleTypeProvider arm 2: @tanstack/react-table namespace import.
		{
			Code: `
import * as TableLib from '@tanstack/react-table';
function Table({ data, columns }) {
  const table = (TableLib).useReactTable({ data, columns });
  return <Grid table={table} />;
}
			`,
			FileName: "table.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(tanStackTableMessage, 4, 18, 4, 26),
			},
		},
		// ---- Dimension 4: declaration/container forms; namespace destructuring inside a compiler root is tracked. ----
		// Locks in upstream defaultModuleTypeProvider arm 2 through local destructuring.
		{
			Code: `
import * as TableLib from '@tanstack/react-table';
function Table({ data, columns }) {
  const { useReactTable } = TableLib;
  const table = useReactTable({ data, columns });
  return <Grid table={table} />;
}
			`,
			FileName: "table.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(tanStackTableMessage, 5, 17, 5, 30),
			},
		},
		// ---- Dimension 4: receiver wrappers; TS assertions on the useForm return value are transparent. ----
		// Locks in upstream defaultModuleTypeProvider arm 1: react-hook-form watch return property.
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const form = useForm();
  const value = (form as any).watch('name');
  return <div>{value}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 5, 18, 5, 29),
			},
		},
		// ---- Dimension 4: receiver wrappers; TS non-null assertions are transparent for report ranges. ----
		// Locks in upstream report location for TSNonNullExpression receivers.
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const form = useForm();
  return <div>{form!.watch('name')}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 5, 16, 5, 20),
			},
		},
		// ---- Dimension 4: access/key forms; string-literal element access matches the static watch property. ----
		// Locks in upstream signature.knownIncompatible branch: known incompatible function property.
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const form = useForm();
  const value = form['watch']('name');
  return <div>{value}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 5, 17, 5, 21),
			},
		},
		// ---- Dimension 4: access/key forms; no-substitution template keys match static watch. ----
		// Locks in upstream signature.knownIncompatible branch for template-literal element access.
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const form = useForm();
  const value = form[` + "`watch`" + `]('name');
  return <div>{value}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 5, 17, 5, 21),
			},
		},
		// ---- Dimension 4: declaration/container forms; renamed destructuring tracks the watch binding. ----
		// Locks in upstream hook-return property arm: destructured incompatible function.
		{
			Code: `
import { useForm as useHookForm } from 'react-hook-form';
function Form() {
  const { watch: observe } = useHookForm();
  return <div>{observe('name')}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 5, 16, 5, 23),
			},
		},
		// ---- Real-user: TanStack Table issue #5141; useReactTable is reported at the call site. ----
		{
			Code: `
import { useReactTable as createTable } from '@tanstack/react-table';
function Table({ data, columns }) {
  const table = createTable({ data, columns, getCoreRowModel: () => null });
  return <Grid table={table} />;
}
			`,
			FileName: "table.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(tanStackTableMessage, 4, 17, 4, 28),
			},
		},
		// ---- Real-user: React issue #33057; the exact incompatible table hook is skipped by the compiler. ----
		{
			Code: `
import { useReactTable } from '@tanstack/react-table';
function Table({ data, columns }) {
  const table = useReactTable({ data, columns, state: {} });
  return <Grid table={table} />;
}
			`,
			FileName: "table.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(tanStackTableMessage, 4, 17, 4, 30),
			},
		},
		// ---- Dimension 4: declaration/container forms; assignment updates an existing form binding. ----
		// Locks in upstream hook-return property arm through assignment instead of declaration.
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  let form;
  form = useForm();
  return <div>{form.watch('name')}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 6, 16, 6, 20),
			},
		},
		// ---- Dimension 4: declaration/container forms; object destructuring assignment tracks watch. ----
		// Locks in upstream hook-return property arm for assignment patterns.
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  let watch;
  ({ watch } = useForm());
  return <div>{watch('name')}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 6, 16, 6, 21),
			},
		},
		// ---- Dimension 4: declaration/container forms; assigned aliases keep imported hook identity. ----
		// Locks in upstream defaultModuleTypeProvider arm 2 through alias assignment.
		{
			Code: `
import { useReactTable } from '@tanstack/react-table';
function Table({ data, columns }) {
  let createTable;
  createTable = useReactTable;
  const table = createTable({ data, columns });
  return <Grid table={table} />;
}
			`,
			FileName: "table.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(tanStackTableMessage, 6, 17, 6, 28),
			},
		},
		// ---- Dimension 4: receiver wrappers; direct useForm().watch reports the useForm() receiver. ----
		// Locks in upstream signature.knownIncompatible branch report location for call-expression receivers.
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  return <div>{useForm().watch('name')}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 4, 16, 4, 25),
			},
		},
		// ---- Dimension 4: nesting/traversal boundaries; nested callbacks inside compiler roots are still checked. ----
		// Locks in upstream compiler target traversal through nested function bodies.
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const form = useForm();
  const getName = () => form.watch('name');
  return <div>{getName()}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 5, 25, 5, 29),
			},
		},
		// ---- Dimension 4: optional chain; optional calls still use the same incompatible API. ----
		// Locks in upstream signature.knownIncompatible branch when tsgo marks the call optional.
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const form = useForm();
  return <div>{form.watch?.('name')}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 5, 16, 5, 20),
			},
		},
	}

	// N/A: PrivateIdentifier, numeric keys, class members, overload signatures,
	// abstract/declare members, and autofix boundaries do not apply because the
	// rule only checks imported call expressions and the static `watch` property.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &IncompatibleLibraryRule, valid, invalid)
}
