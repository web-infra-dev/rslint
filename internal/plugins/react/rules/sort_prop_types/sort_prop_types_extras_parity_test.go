// TestSortPropTypesParityRegressions covers additional ESTree/tsgo
// representation gaps and JavaScript coercion edge cases found by differential
// testing against eslint-plugin-react 7.37.5.
package sort_prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortPropTypesParityRegressions(t *testing.T) {
	shapeOptions := map[string]any{"sortShapeProp": true}
	requiredOptions := map[string]any{"requiredFirst": true}
	checkTypesOptions := map[string]any{"checkTypes": true}
	jsConfig := "tsconfig.allow-js.json"
	propsNotSorted := []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}
	requiredFirst := []rule_tester.InvalidTestCaseError{{MessageId: "requiredPropsFirst"}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SortPropTypesRule, []rule_tester.ValidTestCase{
		// Variable-like declarations without an initializer do not resolve to an object.
		{Code: `let props; C.propTypes = props;`, Tsx: true},
		{Code: `for (const props of values) { C.propTypes = props; }`, Tsx: true},
		{Code: `for (const props in values) { C.propTypes = props; }`, Tsx: true},
		{Code: `try {} catch (props) { C.propTypes = props; }`, Tsx: true},
		{Code: `function render(props) { C.propTypes = props; }`, Tsx: true},
		{Code: `var props; var props = { z: P, a: P }; C.propTypes = props;`, Tsx: true},
		{Code: `function render(props = { z: P, a: P }) { C.propTypes = props; }`, Tsx: true},
		// Binding-element resolution must stop at a parameter instead of escaping
		// into an enclosing variable declaration.
		{Code: `const holder = { z: P, a: P, method({ props }) { C.propTypes = props; } };`, Tsx: true},

		// A continuous optional chain is represented by ESTree as ChainExpression
		// and therefore is not a propTypes declaration.
		{Code: `C?.[propTypes] + ({ z: P, a: P });`, Tsx: true},
		{Code: `C?.member[propTypes] + ({ z: P, a: P });`, Tsx: true},
		{Code: `C.member?.[propTypes] + ({ z: P, a: P });`, Tsx: true},
		{Code: `const { x: value = C.propTypes } = source;`, Tsx: true},

		// Parentheses terminate an optional chain and make the callee a
		// ChainExpression, which is opaque to upstream's shape predicate.
		{Code: `(P?.shape)({ z: P, a: P });`, Options: shapeOptions, Tsx: true},
		{Code: `(P?.[shape])?.({ z: P, a: P });`, Options: shapeOptions, Tsx: true},
		{Code: `/** @type {any} */ (P?.shape)({ z: P, a: P });`, FileName: "optional-shape.js", TSConfig: jsConfig, Options: shapeOptions},
		// A JSDoc annotation is not an ESTree typeAnnotation on a class field.
		{Code: `class C { /** @type {any} */ props = ({ z: P, a: P }); }`, FileName: "jsdoc-props-field.js", TSConfig: jsConfig},

		// typescript-eslint exposes ambient declarations and overload signatures as
		// TSDeclareFunction, not FunctionDeclaration.
		{Code: `declare function C(props: { z: string; a: string }): void;`, Options: checkTypesOptions, Tsx: true},
		{Code: `export declare function C(props: { z: string; a: string }): void;`, Options: checkTypesOptions, Tsx: true},
		{Code: `function C(props: { z: string; a: string }): void;`, Options: checkTypesOptions, Tsx: true},

		// StringToBigInt rejects signs around a non-decimal prefix and doubled signs.
		{Code: `C.propTypes = { "+0x10": P, 15n: P };`, Tsx: true},
		{Code: `C.propTypes = { "0x+10": P, 15n: P };`, Tsx: true},
		{Code: `C.propTypes = { "0x-10": P, 15n: P };`, Tsx: true},
		{Code: `C.propTypes = { "++16": P, 15n: P };`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// Optional calls remain visible when the member itself is the direct callee.
		{Code: `P.shape?.({ z: P, a: P });`, Options: shapeOptions, Tsx: true, Errors: propsNotSorted},
		{Code: `P?.shape({ z: P, a: P });`, Options: shapeOptions, Tsx: true, Errors: propsNotSorted},
		{Code: `P?.shape?.({ z: P, a: P });`, Options: shapeOptions, Tsx: true, Errors: propsNotSorted},
		{Code: `(P.shape)?.({ z: P, a: P });`, Options: shapeOptions, Tsx: true, Errors: propsNotSorted},
		{Code: `P?.[shape]?.({ z: P, a: P });`, Options: shapeOptions, Tsx: true, Errors: propsNotSorted},
		{Code: `/** @satisfies {any} */ P?.shape({ z: P, a: P });`, FileName: "satisfies-shape.js", TSConfig: jsConfig, Options: shapeOptions, Errors: propsNotSorted},

		// Parentheses end the optional receiver chain before the computed access.
		{Code: `(C?.member)[propTypes] + ({ z: P, a: P });`, Tsx: true, Errors: propsNotSorted},

		// ESTree's MemberExpression listener observes for-in/of as well as binary parents.
		{Code: `for (C.propTypes in ({ z: P, a: P })) {}`, Tsx: true, Errors: propsNotSorted},
		{Code: `for (C[propTypes] of ({ z: P, a: P })) {}`, Tsx: true, Errors: propsNotSorted},
		{Code: `async function f() { for await (C.propTypes of ({ z: P, a: P })) {} }`, Tsx: true, Errors: propsNotSorted},
		// tsgo represents ESTree AssignmentPattern with a binary node; its
		// right-hand default is another valid MemberExpression parent value.
		{Code: `({ x: C.propTypes = ({ z: P, a: P }) } = source);`, Tsx: true, Errors: propsNotSorted},
		{Code: `([C.propTypes = ({ z: P, a: P })] = source);`, Tsx: true, Errors: propsNotSorted},

		// Only the overload implementation is represented as an ESTree function.
		{Code: `function C(props: { z: string; a: string }): void; function C(props: { z: string; a: string }) {}`, Options: checkTypesOptions, Tsx: true, Errors: propsNotSorted},

		// Valid decimal/radix strings still participate in BigInt comparison.
		{Code: `C.propTypes = { "0x10": P, 15n: P };`, Tsx: true, Errors: propsNotSorted},
		{Code: `C.propTypes = { "+16": P, 15n: P };`, Tsx: true, Errors: propsNotSorted},
		// The upstream variable utility reads the first declaration's initializer.
		{Code: `var props = { z: P, a: P }; var props; C.propTypes = props;`, Tsx: true, Errors: propsNotSorted},

		// JSDoc casts are absent from ESTree and must be transparent in every
		// runtime-expression position used by this rule.
		{Code: `C.propTypes = /** @type {any} */ ({ z: P, a: P });`, FileName: "jsdoc-value.js", TSConfig: jsConfig, Errors: propsNotSorted},
		{Code: `const props = /** @type {any} */ ({ z: P, a: P }); C.propTypes = props;`, FileName: "jsdoc-variable.js", TSConfig: jsConfig, Errors: propsNotSorted},
		{Code: `const C = { propTypes: /** @type {any} */ ({ z: P, a: P }) };`, FileName: "jsdoc-object-property.js", TSConfig: jsConfig, Errors: propsNotSorted},
		{Code: `const C = { [/** @type {any} */ (propTypes)]: ({ z: P, a: P }) };`, FileName: "jsdoc-object-property-name.js", TSConfig: jsConfig, Errors: propsNotSorted},
		{Code: `class C { static propTypes = /** @type {any} */ ({ z: P, a: P }); }`, FileName: "jsdoc-class-property.js", TSConfig: jsConfig, Errors: propsNotSorted},
		{Code: `/** @type {any} */ (C.propTypes) + ({ z: P, a: P });`, FileName: "jsdoc-member.js", TSConfig: jsConfig, Errors: propsNotSorted},
		{Code: `C[/** @type {any} */ (propTypes)] + ({ z: P, a: P });`, FileName: "jsdoc-prop-types-name.js", TSConfig: jsConfig, Errors: propsNotSorted},
		{Code: `C.propTypes = { a: P, b: /** @type {any} */ (P.isRequired) };`, FileName: "jsdoc-required.js", TSConfig: jsConfig, Options: requiredOptions, Errors: requiredFirst},
		{Code: `C.propTypes = { a: P, b: P[/** @type {any} */ (isRequired)] };`, FileName: "jsdoc-computed-required.js", TSConfig: jsConfig, Options: requiredOptions, Errors: requiredFirst},
		{Code: `/** @type {any} */ (P.shape)({ z: P, a: P });`, FileName: "jsdoc-shape-callee.js", TSConfig: jsConfig, Options: shapeOptions, Errors: propsNotSorted},
		{Code: `P.shape(/** @type {any} */ ({ z: P, a: P }));`, FileName: "jsdoc-shape-argument.js", TSConfig: jsConfig, Options: shapeOptions, Errors: propsNotSorted},
		{Code: `P[/** @type {any} */ (shape)]({ z: P, a: P });`, FileName: "jsdoc-shape-name.js", TSConfig: jsConfig, Options: shapeOptions, Errors: propsNotSorted},
		{Code: `C.propTypes = /** @type {any} */ (wrap)({ z: P, a: P });`, FileName: "jsdoc-wrapper.js", TSConfig: jsConfig, Settings: map[string]interface{}{"propWrapperFunctions": []any{"wrap"}}, Errors: propsNotSorted},
		{Code: `C.propTypes = { z: P, [/** @type {any} */ (a)]: P };`, FileName: "jsdoc-sort-key.js", TSConfig: jsConfig, Errors: propsNotSorted},
	})
}
