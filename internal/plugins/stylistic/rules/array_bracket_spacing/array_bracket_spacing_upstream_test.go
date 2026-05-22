// TestArrayBracketSpacingUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/rules/array-bracket-spacing/array-bracket-spacing.test.ts
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in
// array_bracket_spacing_extras_test.go.
package array_bracket_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/array_bracket_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func optsAlways() []interface{}                    { return []interface{}{"always"} }
func optsNever() []interface{}                     { return []interface{}{"never"} }
func optsAlwaysObj(o map[string]interface{}) []any { return []interface{}{"always", o} }
func optsNeverObj(o map[string]interface{}) []any  { return []interface{}{"never", o} }

func TestArrayBracketSpacingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&array_bracket_spacing.ArrayBracketSpacingRule,
		[]rule_tester.ValidTestCase{
			// ---- Member access in always (upstream tests these for ArrayExpression visitor independence) ----
			{Code: `var foo = obj[ 1 ]`, Options: optsAlways()},
			{Code: `var foo = obj[ 'foo' ];`, Options: optsAlways()},
			{Code: `var foo = obj[ [ 1, 1 ] ];`, Options: optsAlways()},

			// ---- always - singleValue ----
			{Code: `var foo = ['foo']`, Options: optsAlwaysObj(map[string]interface{}{"singleValue": false})},
			{Code: `var foo = [2]`, Options: optsAlwaysObj(map[string]interface{}{"singleValue": false})},
			{Code: `var foo = [[ 1, 1 ]]`, Options: optsAlwaysObj(map[string]interface{}{"singleValue": false})},
			{Code: `var foo = [{ 'foo': 'bar' }]`, Options: optsAlwaysObj(map[string]interface{}{"singleValue": false})},
			{Code: `var foo = [bar]`, Options: optsAlwaysObj(map[string]interface{}{"singleValue": false})},

			// ---- always - objectsInArrays ----
			{Code: `var foo = [{ 'bar': 'baz' }, 1,  5 ];`, Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false})},
			{Code: `var foo = [ 1, 5, { 'bar': 'baz' }];`, Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false})},
			{Code: "var foo = [{\n'bar': 'baz', \n'qux': [{ 'bar': 'baz' }], \n'quxx': 1 \n}]", Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false})},
			{Code: `var foo = [{ 'bar': 'baz' }]`, Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false})},
			{Code: `var foo = [{ 'bar': 'baz' }, 1, { 'bar': 'baz' }];`, Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false})},
			{Code: `var foo = [ 1, { 'bar': 'baz' }, 5 ];`, Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false})},
			{Code: `var foo = [ 1, { 'bar': 'baz' }, [{ 'bar': 'baz' }] ];`, Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false})},
			{Code: `var foo = [ function(){} ];`, Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false})},

			// ---- always - arraysInArrays ----
			{Code: `var arr = [[ 1, 2 ], 2, 3, 4 ];`, Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false})},
			{Code: `var arr = [[ 1, 2 ], [[[ 1 ]]], 3, 4 ];`, Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false})},
			{Code: `var foo = [ arr[i], arr[j] ];`, Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false})},

			// ---- always - arraysInArrays, objectsInArrays ----
			{Code: `var arr = [[ 1, 2 ], 2, 3, { 'foo': 'bar' }];`, Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false, "objectsInArrays": false})},

			// ---- always - arraysInArrays, objectsInArrays, singleValue ----
			{Code: `var arr = [[ 1, 2 ], [2], 3, { 'foo': 'bar' }];`, Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false, "objectsInArrays": false, "singleValue": false})},

			// ---- always ----
			{Code: `obj[ foo ]`, Options: optsAlways()},
			{Code: "obj[\nfoo\n]", Options: optsAlways()},
			{Code: `obj[ 'foo' ]`, Options: optsAlways()},
			{Code: `obj[ 'foo' + 'bar' ]`, Options: optsAlways()},
			{Code: `obj[ obj2[ foo ] ]`, Options: optsAlways()},
			{Code: "obj.map(function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsAlways()},
			{Code: "obj[ 'map' ](function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsAlways()},
			{Code: "obj[ 'for' + 'Each' ](function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsAlways()},

			{Code: `var arr = [ 1, 2, 3, 4 ];`, Options: optsAlways()},
			{Code: `var arr = [ [ 1, 2 ], 2, 3, 4 ];`, Options: optsAlways()},
			{Code: "var arr = [\n1, 2, 3, 4\n];", Options: optsAlways()},
			{Code: `var foo = [];`, Options: optsAlways()},

			// ---- singleValue: false, objectsInArrays: true, arraysInArrays ----
			{Code: "this.db.mappings.insert([\n { alias: 'a', url: 'http://www.amazon.de' },\n { alias: 'g', url: 'http://www.google.de' }\n], function() {});", Options: optsAlwaysObj(map[string]interface{}{"singleValue": false, "objectsInArrays": true, "arraysInArrays": true})},

			// ---- always - destructuring assignment ----
			{Code: `var [ x, y ] = z`, Options: optsAlways()},
			{Code: `var [ x,y ] = z`, Options: optsAlways()},
			{Code: "var [ x, y\n] = z", Options: optsAlways()},
			{Code: "var [\nx, y ] = z", Options: optsAlways()},
			{Code: "var [\nx, y\n] = z", Options: optsAlways()},
			{Code: "var [\nx,,,\n] = z", Options: optsAlways()},
			{Code: `var [ ,x, ] = z`, Options: optsAlways()},
			{Code: "var [\nx, ...y\n] = z", Options: optsAlways()},
			{Code: "var [\nx, ...y ] = z", Options: optsAlways()},
			{Code: `var [[ x, y ], z ] = arr;`, Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false})},
			{Code: `var [ x, [ y, z ]] = arr;`, Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false})},
			{Code: `[{ x, y }, z ] = arr;`, Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false})},
			{Code: `[ x, { y, z }] = arr;`, Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false})},

			// ---- never ----
			{Code: `obj[foo]`, Options: optsNever()},
			{Code: `obj['foo']`, Options: optsNever()},
			{Code: `obj['foo' + 'bar']`, Options: optsNever()},
			{Code: `obj['foo'+'bar']`, Options: optsNever()},
			{Code: `obj[obj2[foo]]`, Options: optsNever()},
			{Code: "obj.map(function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsNever()},
			{Code: "obj['map'](function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsNever()},
			{Code: "obj['for' + 'Each'](function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsNever()},
			{Code: `var arr = [1, 2, 3, 4];`, Options: optsNever()},
			{Code: `var arr = [[1, 2], 2, 3, 4];`, Options: optsNever()},
			{Code: "var arr = [\n1, 2, 3, 4\n];", Options: optsNever()},
			{Code: "obj[\nfoo]", Options: optsNever()},
			{Code: "obj[foo\n]", Options: optsNever()},
			{Code: "var arr = [1,\n2,\n3,\n4\n];", Options: optsNever()},
			{Code: "var arr = [\n1,\n2,\n3,\n4];", Options: optsNever()},

			// ---- never - destructuring assignment ----
			{Code: `var [x, y] = z`, Options: optsNever()},
			{Code: `var [x,y] = z`, Options: optsNever()},
			{Code: "var [x, y\n] = z", Options: optsNever()},
			{Code: "var [\nx, y] = z", Options: optsNever()},
			{Code: "var [\nx, y\n] = z", Options: optsNever()},
			{Code: "var [\nx,,,\n] = z", Options: optsNever()},
			{Code: `var [,x,] = z`, Options: optsNever()},
			{Code: "var [\nx, ...y\n] = z", Options: optsNever()},
			{Code: "var [\nx, ...y] = z", Options: optsNever()},
			{Code: `var [ [x, y], z] = arr;`, Options: optsNeverObj(map[string]interface{}{"arraysInArrays": true})},
			{Code: `var [x, [y, z] ] = arr;`, Options: optsNeverObj(map[string]interface{}{"arraysInArrays": true})},
			{Code: `[ { x, y }, z] = arr;`, Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},
			{Code: `[x, { y, z } ] = arr;`, Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},

			// ---- never - singleValue ----
			{Code: `var foo = [ 'foo' ]`, Options: optsNeverObj(map[string]interface{}{"singleValue": true})},
			{Code: `var foo = [ 2 ]`, Options: optsNeverObj(map[string]interface{}{"singleValue": true})},
			{Code: `var foo = [ [1, 1] ]`, Options: optsNeverObj(map[string]interface{}{"singleValue": true})},
			{Code: `var foo = [ {'foo': 'bar'} ]`, Options: optsNeverObj(map[string]interface{}{"singleValue": true})},
			{Code: `var foo = [ bar ]`, Options: optsNeverObj(map[string]interface{}{"singleValue": true})},

			// ---- never - objectsInArrays ----
			{Code: `var foo = [ {'bar': 'baz'}, 1, 5];`, Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},
			{Code: `var foo = [1, 5, {'bar': 'baz'} ];`, Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},
			{Code: "var foo = [ {\n'bar': 'baz', \n'qux': [ {'bar': 'baz'} ], \n'quxx': 1 \n} ]", Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},
			{Code: `var foo = [ {'bar': 'baz'} ]`, Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},
			{Code: `var foo = [ {'bar': 'baz'}, 1, {'bar': 'baz'} ];`, Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},
			{Code: `var foo = [1, {'bar': 'baz'} , 5];`, Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},
			{Code: `var foo = [1, {'bar': 'baz'}, [ {'bar': 'baz'} ]];`, Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},
			{Code: `var foo = [function(){}];`, Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},
			{Code: `var foo = [];`, Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true})},

			// ---- never - arraysInArrays ----
			{Code: `var arr = [ [1, 2], 2, 3, 4];`, Options: optsNeverObj(map[string]interface{}{"arraysInArrays": true})},
			{Code: `var foo = [arr[i], arr[j]];`, Options: optsNeverObj(map[string]interface{}{"arraysInArrays": true})},
			{Code: `var foo = [];`, Options: optsNeverObj(map[string]interface{}{"arraysInArrays": true})},

			// ---- never - arraysInArrays, singleValue ----
			{Code: `var arr = [ [1, 2], [ [ [ 1 ] ] ], 3, 4];`, Options: optsNeverObj(map[string]interface{}{"arraysInArrays": true, "singleValue": true})},

			// ---- never - arraysInArrays, objectsInArrays ----
			{Code: `var arr = [ [1, 2], 2, 3, {'foo': 'bar'} ];`, Options: optsNeverObj(map[string]interface{}{"arraysInArrays": true, "objectsInArrays": true})},

			// ---- should not warn ----
			{Code: `var foo = {};`, Options: optsNever()},
			{Code: `var foo = [];`, Options: optsNever()},

			{Code: `var foo = [{'bar':'baz'}, 1, {'bar': 'baz'}];`, Options: optsNever()},
			{Code: `var foo = [{'bar': 'baz'}];`, Options: optsNever()},
			{Code: "var foo = [{\n'bar': 'baz', \n'qux': [{'bar': 'baz'}], \n'quxx': 1 \n}]", Options: optsNever()},
			{Code: `var foo = [1, {'bar': 'baz'}, 5];`, Options: optsNever()},
			{Code: `var foo = [{'bar': 'baz'}, 1,  5];`, Options: optsNever()},
			{Code: `var foo = [1, 5, {'bar': 'baz'}];`, Options: optsNever()},
			{Code: `var obj = {'foo': [1, 2]}`, Options: optsNever()},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `var foo = [ ]`,
				Output:  []string{`var foo = []`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- objectsInArrays ----
			{
				Code:    `var foo = [ { 'bar': 'baz' }, 1,  5];`,
				Output:  []string{`var foo = [{ 'bar': 'baz' }, 1,  5 ];`},
				Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    `var foo = [1, 5, { 'bar': 'baz' } ];`,
				Output:  []string{`var foo = [ 1, 5, { 'bar': 'baz' }];`},
				Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 34, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code:    `var foo = [ { 'bar':'baz' }, 1, { 'bar': 'baz' } ];`,
				Output:  []string{`var foo = [{ 'bar':'baz' }, 1, { 'bar': 'baz' }];`},
				Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 49, EndLine: 1, EndColumn: 50},
				},
			},

			// ---- singleValue ----
			{
				Code:    `var obj = [ 'foo' ];`,
				Output:  []string{`var obj = ['foo'];`},
				Options: optsAlwaysObj(map[string]interface{}{"singleValue": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `var obj = ['foo' ];`,
				Output:  []string{`var obj = ['foo'];`},
				Options: optsAlwaysObj(map[string]interface{}{"singleValue": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `var obj = ['foo'];`,
				Output:  []string{`var obj = [ 'foo' ];`},
				Options: optsNeverObj(map[string]interface{}{"singleValue": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},

			// ---- always - arraysInArrays ----
			{
				Code:    `var arr = [ [ 1, 2 ], 2, 3, 4 ];`,
				Output:  []string{`var arr = [[ 1, 2 ], 2, 3, 4 ];`},
				Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `var arr = [ 1, 2, 2, [ 3, 4 ] ];`,
				Output:  []string{`var arr = [ 1, 2, 2, [ 3, 4 ]];`},
				Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 30, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    `var arr = [[ 1, 2 ], 2, [ 3, 4 ] ];`,
				Output:  []string{`var arr = [[ 1, 2 ], 2, [ 3, 4 ]];`},
				Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 33, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    `var arr = [ [ 1, 2 ], 2, [ 3, 4 ]];`,
				Output:  []string{`var arr = [[ 1, 2 ], 2, [ 3, 4 ]];`},
				Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `var arr = [ [ 1, 2 ], 2, [ 3, 4 ] ];`,
				Output:  []string{`var arr = [[ 1, 2 ], 2, [ 3, 4 ]];`},
				Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 34, EndLine: 1, EndColumn: 35},
				},
			},

			// ---- always - destructuring ----
			{
				Code:    `var [x,y] = y`,
				Output:  []string{`var [ x,y ] = y`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    `var [x,y ] = y`,
				Output:  []string{`var [ x,y ] = y`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
				},
			},
			{
				Code:    `var [,,,x,,] = y`,
				Output:  []string{`var [ ,,,x,, ] = y`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `var [ ,,,x,,] = y`,
				Output:  []string{`var [ ,,,x,, ] = y`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceBefore", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    `var [...horse] = y`,
				Output:  []string{`var [ ...horse ] = y`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `var [...horse ] = y`,
				Output:  []string{`var [ ...horse ] = y`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
				},
			},
			{
				Code:    `var [ [ x, y ], z ] = arr;`,
				Output:  []string{`var [[ x, y ], z ] = arr;`},
				Options: optsAlwaysObj(map[string]interface{}{"arraysInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code:    `[ { x, y }, z ] = arr;`,
				Output:  []string{`[{ x, y }, z ] = arr;`},
				Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code:    `[ x, { y, z } ] = arr;`,
				Output:  []string{`[ x, { y, z }] = arr;`},
				Options: optsAlwaysObj(map[string]interface{}{"objectsInArrays": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},

			// ---- never - arraysInArrays ----
			{
				Code:    `var arr = [[1, 2], 2, [3, 4]];`,
				Output:  []string{`var arr = [ [1, 2], 2, [3, 4] ];`},
				Options: optsNeverObj(map[string]interface{}{"arraysInArrays": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    `var arr = [ ];`,
				Output:  []string{`var arr = [];`},
				Options: optsNeverObj(map[string]interface{}{"arraysInArrays": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- never - objectsInArrays ----
			{
				Code:    `var arr = [ ];`,
				Output:  []string{`var arr = [];`},
				Options: optsNeverObj(map[string]interface{}{"objectsInArrays": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- always ----
			{
				Code:    `var arr = [1, 2, 3, 4];`,
				Output:  []string{`var arr = [ 1, 2, 3, 4 ];`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `var arr = [1, 2, 3, 4 ];`,
				Output:  []string{`var arr = [ 1, 2, 3, 4 ];`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    `var arr = [ 1, 2, 3, 4];`,
				Output:  []string{`var arr = [ 1, 2, 3, 4 ];`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceBefore", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},

			// ---- never ----
			{
				Code:    `var arr = [ 1, 2, 3, 4 ];`,
				Output:  []string{`var arr = [1, 2, 3, 4];`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    `var arr = [1, 2, 3, 4 ];`,
				Output:  []string{`var arr = [1, 2, 3, 4];`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `var arr = [ 1, 2, 3, 4];`,
				Output:  []string{`var arr = [1, 2, 3, 4];`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `var arr = [ [ 1], 2, 3, 4];`,
				Output:  []string{`var arr = [[1], 2, 3, 4];`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `var arr = [[1 ], 2, 3, 4 ];`,
				Output:  []string{`var arr = [[1], 2, 3, 4];`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
				},
			},

			// ---- multiple spaces ----
			{
				Code:    `var arr = [  1, 2   ];`,
				Output:  []string{`var arr = [1, 2];`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 14},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 18, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `function f( [   a, b  ] ) {}`,
				Output:  []string{`function f( [a, b] ) {}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 14, EndLine: 1, EndColumn: 17},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 21, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var arr = [ 1,\n   2   ];",
				Output:  []string{"var arr = [1,\n   2];"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 2, Column: 5, EndLine: 2, EndColumn: 8},
				},
			},
			{
				Code:    `var arr = [  1, [ 2, 3  ] ];`,
				Output:  []string{`var arr = [1, [2, 3]];`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 14},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 23, EndLine: 1, EndColumn: 25},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
				},
			},
		},
	)
}
