package require_array_sort_compare

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireArraySortCompareExtras(t *testing.T) {
	for _, method := range []string{"sort", "toSorted"} {
		t.Run(method, func(t *testing.T) {
			var valid []rule_tester.ValidTestCase
			var invalid []rule_tester.InvalidTestCase
			errors := []rule_tester.InvalidTestCaseError{{MessageId: "requireCompare"}}

			// These element types are strings according to upstream getTypeName.
			for _, code := range []string{
				`enum Key { A = 'a', B = 'b' }
export const sorted = [Key.B, Key.A].filter(Boolean).%s();`,
				`enum Key { A = 'a', B = 'b' }
declare const values: Key.A[];
values.%s();`,
				`const enum Key { A = 'a', B = 'b' }
declare const values: Key[];
values.%s();`,
				`enum Key { A = 'a', B = 'b', Number = 0 }
declare const values: (Key.A | Key.B)[];
values.%s();`,
				`declare const values: ('a' | 'b')[];
values.%s();`,
				"declare const values: `key-${number}`[];\nvalues.%s();",
				`type Branded = string & { readonly brand: unique symbol };
declare const values: Branded[];
values.%s();`,
				`enum Key { A = 'a', B = 'b' }
type Branded = string & { readonly brand: unique symbol };
declare const values: (Key | Branded)[];
values.%s();`,
				`function sort<T extends string>(values: T[]) { values.%s(); }`,
				`function sort<T extends string, U extends T>(values: U[]) { values.%s(); }`,
				`enum Key { A = 'a', B = 'b' }
function sort<T extends Key>(values: T[]) { values.%s(); }`,
			} {
				code = fmt.Sprintf(code, method)
				valid = append(valid,
					rule_tester.ValidTestCase{Code: code},
					rule_tester.ValidTestCase{Code: code, Options: map[string]any{"ignoreStringArrays": true}},
				)
				// Disabling the exemption must still require a comparator.
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code: code, Options: map[string]any{"ignoreStringArrays": false}, Errors: errors,
				})
			}

			// Do not broaden the exemption to non-string elements or array unions.
			for _, code := range []string{
				`declare const values: 42[]; values.%s();`,
				`enum Key { A = 1, B = 2 }
declare const values: Key.A[];
values.%s();`,
				`enum Key { A = Math.random() }
declare const values: Key.A[];
values.%s();`,
				`enum Key { A, B }
declare const values: Key[];
values.%s();`,
				`enum Key { A = 'a', B = 0 }
declare const values: Key[];
values.%s();`,
				`enum Key { A = 'a', B = 'b' }
declare const values: (Key | number)[];
values.%s();`,
				`enum Key { A = 'a', B = 'b' }
declare const values: (Key | undefined)[];
values.%s();`,
				`declare const values: unknown[]; values.%s();`,
				`declare const values: never[]; values.%s();`,
				`declare const values: String[]; values.%s();`,
				`type Branded = number & { readonly brand: unique symbol };
declare const values: Branded[];
values.%s();`,
				`function sort<T>(values: T[]) { values.%s(); }`,
				`function sort<T extends string | number>(values: T[]) { values.%s(); }`,
				`enum Key { A = 'a', B = 'b' }
declare const values: Key[] | 'c'[];
values.%s();`,
			} {
				invalid = append(invalid, rule_tester.InvalidTestCase{Code: fmt.Sprintf(code, method), Errors: errors})
			}

			// Upstream only reports arrays and unions consisting entirely of arrays.
			// Tuples remain exempt even with ignoreStringArrays disabled.
			for _, code := range []string{
				`enum Key { A = 'a', B = 'b' }
declare const values: [Key.A, Key.B, 'c'];
values.%s();`,
				`declare const values: [number, string]; values.%s();`,
				`declare const values: [number, number?]; values.%s();`,
				`declare const values: [number, ...number[]]; values.%s();`,
				`declare const values: []; values.%s();`,
				`declare const values: [number] | [number, number]; values.%s();`,
				`declare const values: [number] | number[]; values.%s();`,
				`function sort<T extends number>(values: [T, T]) { values.%s(); }`,
			} {
				code = fmt.Sprintf(code, method)
				valid = append(valid,
					rule_tester.ValidTestCase{Code: code},
					rule_tester.ValidTestCase{Code: code, Options: map[string]any{"ignoreStringArrays": false}},
				)
			}

			if method == "toSorted" {
				valid = append(valid,
					rule_tester.ValidTestCase{Code: `enum Key { A = 'a', B = 'b' }
declare const values: readonly Key[]; values.toSorted();`},
					rule_tester.ValidTestCase{
						Code:    `declare const values: readonly [number, number]; values.toSorted();`,
						Options: map[string]any{"ignoreStringArrays": false},
					},
				)
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code: `declare const values: readonly number[]; values.toSorted();`, Errors: errors,
				})
			}

			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RequireArraySortCompareRule, valid, invalid)
		})
	}
}
