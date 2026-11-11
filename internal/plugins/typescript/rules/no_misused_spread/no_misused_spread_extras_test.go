package no_misused_spread

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMisusedSpreadSuggestionTokenBoundariesAndComments(t *testing.T) {
	const prefix = "const map = new Map<string, number>();\n"
	cases := []struct {
		code, output string
		tsx          bool
	}{
		// Whole-object replacement must keep the preceding keyword separate.
		{code: "function f(){return{ ...map };}", output: "function f(){return Object.fromEntries(map);}"},
		{code: "function f(){throw{ ...map };}", output: "function f(){throw Object.fromEntries(map);}"},
		{code: "function* f(){yield{ ...map };}", output: "function* f(){yield Object.fromEntries(map);}"},
		{code: "switch (1) { case{ ...map }: break; }", output: "switch (1) { case Object.fromEntries(map): break; }"},
		{code: "export default{ ...map };", output: "export default Object.fromEntries(map);"},
		{code: `function f(){return{
...map
};}`, output: "function f(){return Object.fromEntries(map);}"},
		// Preserve comments and line endings, including directives and ASI boundaries.
		{code: "const o={ /*leading*/...map };", output: "const o=( /*leading*/Object.fromEntries(map) );"},
		{code: "const o={ .../*argument*/map };", output: "const o=( Object.fromEntries(/*argument*/map) );"},
		{code: "const o={ ...( /*parentheses*/map) };", output: "const o=( Object.fromEntries(( /*parentheses*/map)) );"},
		{code: "const o={ ...map/*trailing*/ };", output: "const o=( Object.fromEntries(map)/*trailing*/ );"},
		{code: "const o={ ...map,/*trailing comma*/ };", output: "const o=( Object.fromEntries(map)/*trailing comma*/ );"},
		{code: `const o={
// leading
...map
};`, output: `const o=(
// leading
Object.fromEntries(map)
);`},
		{code: `const o={
...// argument
map
};`, output: `const o=(
Object.fromEntries(// argument
map)
);`},
		{code: `const o={ ...map // trailing
};`, output: `const o=( Object.fromEntries(map) // trailing
);`},
		{code: `function f(){return{ // keep return on this line
...map };}`, output: `function f(){return( // keep return on this line
Object.fromEntries(map) );}`},
		{code: `function f(){throw{ // keep throw on this line
...map };}`, output: `function f(){throw( // keep throw on this line
Object.fromEntries(map) );}`},
		{code: `function* f(){yield{ // keep yield on this line
...map };}`, output: `function* f(){yield( // keep yield on this line
Object.fromEntries(map) );}`},
		{code: "const o={...map/*before comma*/,};", output: "const o=(Object.fromEntries(map)/*before comma*/);"},
		{code: `const o={...map, // after comma
};`, output: `const o=(Object.fromEntries(map) // after comma
);`},
		{code: `const o={ /* eslint-disable no-debugger */ ...map };
debugger;`, output: `const o=( /* eslint-disable no-debugger */ Object.fromEntries(map) );
debugger;`},
		{code: "const o={\r\n// keep CRLF\r\n...map\r\n};", output: "const o=(\r\n// keep CRLF\r\nObject.fromEntries(map)\r\n);"},
		{code: "const emoji=\"😀\";const o={/*keep*/...map};", output: "const emoji=\"😀\";const o=(/*keep*/Object.fromEntries(map));"},
		// Controls: untouched trivia and comments already inside the expression.
		{code: "const o={a:1,.../*keep*/map};", output: "const o={a:1,.../*keep*/Object.fromEntries(map)};"},
		{code: "const c=<C {...(/*keep*/map)}/>;", output: "const c=<C {...(/*keep*/Object.fromEntries(map))}/>;", tsx: true},
		{code: "const o={...new Map(/*keep*/)};", output: "const o=Object.fromEntries(new Map(/*keep*/));"},
		{code: "const o={...(true /*condition*/ ? map : map)};", output: "const o=Object.fromEntries((true /*condition*/ ? map : map));"},
		{code: "const o={...new Map([[\"/*not a comment*/\",1]])};", output: "const o=Object.fromEntries(new Map([[\"/*not a comment*/\",1]]));"},
		{code: "const o={...new Map([[/\\/\\*/.source,1]])};", output: "const o=Object.fromEntries(new Map([[/\\/\\*/.source,1]]));"},
	}
	invalid := make([]rule_tester.InvalidTestCase, 0, len(cases))
	for _, test := range cases {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: prefix + test.code, Tsx: test.tsx,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId:   "noMapSpreadInObject",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceMapSpreadInObject", Output: prefix + test.output}},
			}},
		})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedSpreadRule, nil, invalid)
}
