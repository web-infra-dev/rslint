package incompatible_library

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestIncompatibleLibraryUpstream migrates the public React Compiler
// IncompatibleLibrary module model from upstream
// compiler/packages/babel-plugin-react-compiler/src/HIR/DefaultModuleTypeProvider.ts
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in the incompatible_library_extras_test.go file.
func TestIncompatibleLibraryUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- DefaultModuleTypeProvider: react-hook-form useForm itself is allowed. ----
		{Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const form = useForm();
  return <form />;
}
		`, FileName: "form.tsx", Tsx: true},
		// ---- DefaultModuleTypeProvider: react-hook-form non-watch properties are allowed. ----
		{Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const { register } = useForm();
  return <input {...register('name')} />;
}
		`, FileName: "form.tsx", Tsx: true},
		// ---- DefaultModuleTypeProvider: unknown libraries are ignored. ----
		{Code: `
import { useReactTable } from '@example/react-table';
function Table() {
  const table = useReactTable();
  return <Grid table={table} />;
}
		`, FileName: "table.tsx", Tsx: true},
		// ---- react.dev docs: useWatch is the memoization-compatible React Hook Form API. ----
		{Code: `
import { useWatch } from 'react-hook-form';
function Form() {
  const name = useWatch({ name: 'name' });
  return <div>{name}</div>;
}
		`, FileName: "form.tsx", Tsx: true},
		// ---- react.dev docs pitfall: MobX observer is not detected by this rule yet. ----
		{Code: `
import { observer } from 'mobx-react-lite';
const TodoView = observer(function TodoView({ todo }) {
  return <div>{todo.title}</div>;
});
		`, FileName: "todo.tsx", Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- DefaultModuleTypeProvider: react-hook-form useForm().watch property. ----
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const form = useForm();
  const name = form.watch('name');
  return <div>{name}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 5, 16, 5, 20),
			},
		},
		// ---- DefaultModuleTypeProvider: react-hook-form destructured watch function. ----
		{
			Code: `
import { useForm } from 'react-hook-form';
function Form() {
  const { watch } = useForm();
  const name = watch('name');
  return <div>{name}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 5, 16, 5, 21),
			},
		},
		// ---- react.dev docs: react-hook-form watch inside useMemo is incompatible. ----
		{
			Code: `
import { useMemo } from 'react';
import { useForm } from 'react-hook-form';
function Form() {
  const { watch } = useForm();
  const name = useMemo(() => watch('name'), [watch]);
  return <div>{name}</div>;
}
			`,
			FileName: "form.tsx",
			Tsx:      true,
			Options:  map[string]interface{}{},
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(reactHookFormWatchMessage, 6, 30, 6, 35),
			},
		},
		// ---- react.dev docs: @tanstack/react-table useReactTable hook. ----
		{
			Code: `
import { useReactTable, getCoreRowModel } from '@tanstack/react-table';
function Table({ data, columns }) {
  const table = useReactTable({ data, columns, getCoreRowModel: getCoreRowModel() });
  return <Grid table={table} />;
}
			`,
			FileName: "table.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(tanStackTableMessage, 4, 17, 4, 30),
			},
		},
		// ---- DefaultModuleTypeProvider: @tanstack/react-virtual useVirtualizer hook. ----
		{
			Code: `
import { useVirtualizer } from '@tanstack/react-virtual';
function List({ count }) {
  const virtualizer = useVirtualizer({ count });
  return <Grid virtualizer={virtualizer} />;
}
			`,
			FileName: "list.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				incompatibleLibraryError(tanStackVirtualMessage, 4, 23, 4, 37),
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &IncompatibleLibraryRule, valid, invalid)
}

func incompatibleLibraryError(detail string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "incompatibleLibrary",
		Message:   buildIncompatibleLibraryMessage(detail).Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}
