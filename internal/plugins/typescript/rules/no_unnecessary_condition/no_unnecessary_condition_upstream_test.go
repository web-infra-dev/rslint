package no_unnecessary_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Cases from typescript-eslint v8.74.0.
func TestNoUnnecessaryConditionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		{Code: `
const truthy = <T,>(value: Awaited<T>) => (value ? 1 : 2);
    `},
		{Code: `
const nullish = <T,>(value: Awaited<T>) => value ?? 1;
    `},
		{Code: `
const fromPromise = async <T,>(promise: Promise<T>) => {
  const value = await promise;
  return value ? 1 : 2;
};
    `},
		{Code: `
type Id<T> = T extends unknown ? T : never;
function test<T>(t: Id<T>) {
  return t ? 'yes' : 'no';
}
    `},
		{Code: `
const isUndefined = <T,>(value: Awaited<T>) => value === undefined;
    `},
		{Code: `
const isNull = <T,>(value: Awaited<T>) => value === null;
    `},
		{Code: `
function test<T, K extends keyof T>(value: Awaited<T[K]>) {
  return value ?? 1;
}
    `},
		{Code: `
function test(foo?: boolean, bar?: boolean) {
  return (foo && 'foo') || (bar && 'bar');
}
    `},
		{Code: `
declare const b1: boolean;
declare const b2: boolean;
const x = (b1 && 'b1') || b2;
    `},
		{Code: `
type Fields = {
  hello?: number;
  world?: boolean;
};

let fields: Fields = {};

for (const key of ['hello', 'world'] as const) {
  fields[key] ??= undefined;
}
    `},
	}, nil)
}
