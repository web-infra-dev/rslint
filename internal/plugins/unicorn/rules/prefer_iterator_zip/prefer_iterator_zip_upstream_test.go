package prefer_iterator_zip_test

// cspell:words metadatum

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_iterator_zip"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	loopHeader = `for (let index = 0; index < Math.min(names.length, scores.length); index++)`
	loopBody   = ` { save(names[index], scores[index]); }`
	zipHeader  = `for (const [name_, score] of Iterator.zip([names, scores]))`
	zipBody    = ` { save(name_, score); }`
)

func expectZip(filename, code, suggestion string, line, column, endLine, endColumn int) rule_tester.InvalidTestCase {
	diagnostic := rule_tester.InvalidTestCaseError{
		MessageId: "prefer-iterator-zip",
		Message:   "Use `Iterator.zip(…)` to iterate over these arrays together.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
	if suggestion != "" {
		diagnostic.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: "prefer-iterator-zip/suggestion", Output: suggestion,
		}}
	}
	return rule_tester.InvalidTestCase{FileName: filename, Code: code, Errors: []rule_tester.InvalidTestCaseError{diagnostic}}
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-iterator-zip.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-iterator-zip.md
// Ranges and applied suggestions come from ESLint 10.14.0 and Unicorn 79.0.0.
func TestPreferIteratorZipUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Unsupported loop headers and cached bounds.
		{FileName: "input.js", Code: `names.map((name, index) => scores[index])`},
		{FileName: "input.js", Code: `for (let index = 0; index < names.length; index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < length; index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.max(names.length, scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(names.length, names.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(names.length, scores.length, 5); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(...lengths); index++)` + loopBody},
		{FileName: "input.js", Code: `for (var index = 0; index < Math.min(names.length, scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (const index = 0; index < Math.min(names.length, scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 1; index < Math.min(names.length, scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0, length = 5; index < Math.min(names.length, scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(names.length, scores.length); index += 2)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(names.length, scores.length); index--)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(names.length, scores.length); index = index + 2)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(names.length, scores.length); index = otherIndex + 1)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index <= Math.min(names.length, scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < names.length || index < scores.length; index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < names.length && enabled && index < scores.length; index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < names.length && index < names.length; index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < object.names.length && index < scores.length; index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; ; index++)` + loopBody},
		{FileName: "input.js", Code: "const length = Math.min(names.length, scores.length);\nfor (let index = 0; index < length; index++)" + loopBody + ` use(length);`},
		{FileName: "input.js", Code: `const length = Math.min(names.length, scores.length); prepare(); for (let index = 0; index < length; index++)` + loopBody},
		{FileName: "input.js", Code: "const length = Math.min(names.length, scores.length), other = 1;\nfor (let index = 0; index < length; index++)" + loopBody},
		{FileName: "input.js", Code: `for (let index = 0, length = Math.min(names.length, scores.length); index < length; index++) { save(names[index], scores[index]); use(length); }`},
		{FileName: "input.js", Code: loopHeader + ` save(names[index], scores[index]);`},
		// Uses other than element reads, including mixed shadowing.
		{FileName: "input.js", Code: loopHeader + ` { save(names[index]); }`},
		{FileName: "input.js", Code: loopHeader + ` { save(names[index], scores[index], index); }`},
		{FileName: "input.js", Code: loopHeader + ` { save(names[index + 1], scores[index]); }`},
		{FileName: "input.js", Code: loopHeader + ` { save(names?.[index], scores[index]); }`},
		{FileName: "input.js", Code: loopHeader + ` { save(names[index], scores[index]); names.push("new"); }`},
		{FileName: "input.js", Code: loopHeader + ` { save(names[index], scores[index]); scores = replacement; }`},
		{FileName: "input.js", Code: loopHeader + ` { save(names[index], scores[index]); consume(scores); }`},
		{FileName: "input.js", Code: loopHeader + ` { save(names[index], scores[index]); index++; }`},
		{FileName: "input.js", Code: loopHeader + ` { names[index] = scores[index]; }`},
		{FileName: "input.js", Code: loopHeader + ` { names[index]++; save(scores[index]); }`},
		{FileName: "input.js", Code: loopHeader + ` { delete names[index]; save(scores[index]); }`},
		{FileName: "input.js", Code: loopHeader + ` { [names[index]] = scores[index]; }`},
		{FileName: "input.js", Code: loopHeader + ` { ({value: names[index]} = scores[index]); }`},
		{FileName: "input.js", Code: loopHeader + ` { for (names[index] of scores[index]) {} }`},
		{FileName: "input.js", Code: loopHeader + ` { for (names[index] in scores[index]) {} }`},
		{FileName: "input.js", Code: loopHeader + ` { names[index](scores[index]); }`},
		{FileName: "input.js", Code: loopHeader + ` { names[index]?.(scores[index]); }`},
		{FileName: "input.js", Code: loopHeader + " { names[index]`tag`; save(scores[index]); }"},
		{FileName: "input.js", Code: loopHeader + ` { callbacks.push(() => save(names[index], scores[index])); }`},
		{FileName: "input.js", Code: loopHeader + ` { class Pair { value = save(names[index], scores[index]); } }`},
		{FileName: "input.js", Code: loopHeader + ` { save(names[index], scores[index]); { const names = other; save(names[index]); } }`},
		{FileName: "input.js", Code: loopHeader + ` { save(names[index], scores[index]); { const index = 0; save(names[index]); } }`},
		// Receivers and TypeScript wrapper cases.
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(object.names.length, scores.length); index++) { save(object.names[index], scores[index]); }`},
		{FileName: "input.js", Code: `const names = 'abc'; ` + loopHeader + loopBody},
		{FileName: "input.js", Code: `const names = {0: 'a', length: 1}; ` + loopHeader + loopBody},
		{FileName: "input.js", Code: `const scores = new Set(); ` + loopHeader + loopBody},
		{FileName: "file.ts", Code: `declare const names: string; declare const scores: number[]; ` + loopHeader + loopBody},
		{FileName: "file.ts", Code: loopHeader + ` { names[index]!(); save(scores[index]); }`},
		{FileName: "file.ts", Code: loopHeader + ` { (names[index]<string>)(scores[index]); }`},
		{FileName: "file.ts", Code: loopHeader + " { (names[index]<string>)`tag`; save(scores[index]); }"},
		{FileName: "file.ts", Code: loopHeader + ` { (names[index] as number) = scores[index]; }`},
		{FileName: "file.ts", Code: `declare const names: {length: number; [index: number]: string}; declare const scores: number[]; ` + loopHeader + loopBody},
		// Documentation examples and the upstream empty-file check.
		{FileName: "input.js", Code: "for (const [name, score] of Iterator.zip([names, scores])) {\n\tsave(name, score);\n}"},
		{FileName: "input.js", Code: "index < Math.min(names.length, scores.length);\nindex < names.length && index < scores.length;"},
		{FileName: "input.js", Code: ""},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Math.min bounds, increments, and input types.
		expectZip("input.js", `for (let index = 0; index < Math.min(Iterators.length, scores.length); index++) { save(Iterators[index], scores[index]); }`,
			`for (const [Iterator_, score] of Iterator.zip([Iterators, scores])) { save(Iterator_, score); }`, 1, 1, 1, 80),
		expectZip("input.js", loopHeader+loopBody,
			zipHeader+zipBody, 1, 1, 1, 76),
		expectZip("input.js", `function pair(names, scores) { `+loopHeader+loopBody+` }`,
			`function pair(names, scores) { `+zipHeader+zipBody+` }`, 1, 32, 1, 107),
		expectZip("input.js", `const names = ['a', 'b']; const scores = [1]; `+loopHeader+loopBody,
			`const names = ['a', 'b']; const scores = [1]; `+zipHeader+zipBody, 1, 47, 1, 122),
		expectZip("input.js", `const names = []; const scores = [1]; `+loopHeader+loopBody,
			`const names = []; const scores = [1]; `+zipHeader+zipBody, 1, 39, 1, 114),
		expectZip("input.js", `const names = new Uint8Array(2); const scores = new Float64Array(3); `+loopHeader+loopBody,
			`const names = new Uint8Array(2); const scores = new Float64Array(3); `+zipHeader+zipBody, 1, 70, 1, 145),
		expectZip("input.js", `const names = new BigInt64Array(2); const scores = new BigUint64Array(3); `+loopHeader+loopBody,
			`const names = new BigInt64Array(2); const scores = new BigUint64Array(3); `+zipHeader+zipBody, 1, 75, 1, 150),
		expectZip("input.js", `for (let index = 0; index < Math.min(names.length, scores.length); ++index)`+loopBody,
			zipHeader+zipBody, 1, 1, 1, 76),
		expectZip("input.js", `for (let index = 0; index < Math.min(names.length, scores.length); index += 1)`+loopBody,
			zipHeader+zipBody, 1, 1, 1, 79),
		expectZip("input.js", `for (let index = 0; Math.min(names.length, scores.length) > index; index++)`+loopBody,
			zipHeader+zipBody, 1, 1, 1, 76),
		expectZip("input.js", `for (let index = 0; index < Math.min(names.length, scores.length, ranks.length); index++) { save(names[index], scores[index], ranks[index]); }`,
			`for (const [name_, score, rank] of Iterator.zip([names, scores, ranks])) { save(name_, score, rank); }`, 1, 1, 1, 90),
		// Element reads, shadowing, comments, and generated bindings.
		expectZip("input.js", loopHeader+` { save(names[index], scores[index], names[index]); }`,
			zipHeader+` { save(name_, score, name_); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { const name = names[index]; const score = scores[index]; save(name, score); }`,
			`for (const [name_, score_] of Iterator.zip([names, scores])) { const name = name_; const score = score_; save(name, score); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { if (names[index])`+loopBody+` }`,
			zipHeader+` { if (name_)`+zipBody+` }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { if (!names[index]) { continue; } save(names[index], scores[index]); break; }`,
			zipHeader+` { if (!name_) { continue; } save(name_, score); break; }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save((names)[(index)], ((scores[index]))); }`,
			zipHeader+` { save(name_, ((score))); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save(names[index].value, scores[index].value); }`,
			zipHeader+` { save(name_.value, score.value); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { names[index].save(scores[index]); }`,
			zipHeader+` { name_.save(score); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { names[index].value = scores[index]; }`,
			zipHeader+` { name_.value = score; }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save(name, names[index], scores[index]); }`,
			zipHeader+` { save(name, name_, score); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save(names[index], scores[index]); { const name = "inner"; use(name); } }`,
			zipHeader+` { save(name_, score); { const name = "inner"; use(name); } }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save(names[index], scores[index]); { const index = 1; use(index); } }`,
			zipHeader+` { save(name_, score); { const index = 1; use(index); } }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { /* keep body comment */ save(names[index], scores[index]); }`,
			zipHeader+` { /* keep body comment */ save(name_, score); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save(names[/* keep access comment */ index], scores[index]); }`,
			"", 1, 1, 1, 76),
		expectZip("input.js", `const name = 'existing'; `+loopHeader+` { save(name, names[index], scores[index]); }`,
			`const name = 'existing'; `+zipHeader+` { save(name, name_, score); }`, 1, 26, 1, 101),
		expectZip("input.js", `for (let index = 0; index < Math.min(items.length, item.length); index++) { save(items[index], item[index]); }`,
			`for (const [item_, element] of Iterator.zip([items, item])) { save(item_, element); }`, 1, 1, 1, 74),
		expectZip("input.js", `for (let index = 0; index < Math.min(data.length, metadata.length); index++) { save(data[index], metadata[index]); }`,
			`for (const [datum, metadatum] of Iterator.zip([data, metadata])) { save(datum, metadatum); }`, 1, 1, 1, 77),
		expectZip("input.js", `for (let index = 0; /* keep header comment */ index < Math.min(names.length, scores.length); index++)`+loopBody,
			"", 1, 1, 1, 102),
		expectZip("input.js", `for (/* before declaration */ let index = 0; index < Math.min(names.length, scores.length); index++)`+loopBody,
			"", 1, 1, 1, 101),
		expectZip("input.js", `for (let index = 0; index < Math.min(names.length, scores.length); index++ /* after update */)`+loopBody,
			"", 1, 1, 1, 95),
		// TypeScript wrappers and JSX.
		expectZip("file.ts", `function pair(names: readonly string[], scores: number[]) { `+loopHeader+loopBody+` }`,
			`function pair(names: readonly string[], scores: number[]) { `+zipHeader+zipBody+` }`, 1, 61, 1, 136),
		expectZip("file.ts", loopHeader+` { save(names[index]!, scores[index] as number); }`,
			zipHeader+` { save(name_!, score as number); }`, 1, 1, 1, 76),
		expectZip("file.ts", loopHeader+` { save(<string>names[index], scores[index] satisfies number); }`,
			zipHeader+` { save(<string>name_, score satisfies number); }`, 1, 1, 1, 76),
		expectZip("file.ts", loopHeader+` { save(names[index]<string>, scores[index]); }`,
			zipHeader+` { save(name_<string>, score); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save(names[index]in object, scores[index]instanceof Number); }`,
			zipHeader+` { save(name_ in object, score instanceof Number); }`, 1, 1, 1, 76),
		expectZip("file.ts", loopHeader+` { save(names[index]as string, scores[index]satisfies number); }`,
			zipHeader+` { save(name_ as string, score satisfies number); }`, 1, 1, 1, 76),
		expectZip("input.jsx", `for (let index = 0; index < Math.min(Cards.length, scores.length); index++) { render(<Card name={Cards[index]} score={scores[index]} />); }`,
			`for (const [Card_, score] of Iterator.zip([Cards, scores])) { render(<Card name={Card_} score={score} />); }`, 1, 1, 1, 76),
		expectZip("file.ts", `declare const names: string[]; declare const scores: readonly number[]; `+loopHeader+loopBody,
			`declare const names: string[]; declare const scores: readonly number[]; `+zipHeader+zipBody, 1, 73, 1, 148),
		expectZip("file.ts", `declare const names: Uint8Array; declare const scores: Float64Array; `+loopHeader+loopBody,
			`declare const names: Uint8Array; declare const scores: Float64Array; `+zipHeader+zipBody, 1, 70, 1, 145),
		expectZip("input.js", loopHeader+` { save(names[index], scores[index]); { const names = other; const index = 0; save(names[index]); } }`,
			zipHeader+` { save(name_, score); { const names = other; const index = 0; save(names[index]); } }`, 1, 1, 1, 76),
		expectZip("input.js", `for (let index = 0; index < Math.min(widgets.length, scores.length); index++) { save(widgets[index], scores[index]); function getWidget() { return widget; } }`,
			`for (const [widget_, score] of Iterator.zip([widgets, scores])) { save(widget_, score); function getWidget() { return widget; } }`, 1, 1, 1, 78),
		// Conjunctive conditions and cached bounds.
		expectZip("input.js", `for (let index = 0; index < names.length && index < scores.length; index++)`+loopBody,
			zipHeader+zipBody, 1, 1, 1, 76),
		expectZip("input.js", `for (let index = 0; names.length > index && index < scores.length; index++)`+loopBody,
			zipHeader+zipBody, 1, 1, 1, 76),
		expectZip("input.js", `for (let index = 0; index < names.length && (scores.length > index && index < ranks.length); index++) { save(names[index], scores[index], ranks[index]); }`,
			`for (const [name_, score, rank] of Iterator.zip([names, scores, ranks])) { save(name_, score, rank); }`, 1, 1, 1, 102),
		expectZip("input.js", `for (let index = 0; index < Math.min(names.length, scores.length); index = index + 1)`+loopBody,
			zipHeader+zipBody, 1, 1, 1, 86),
		expectZip("input.js", `for (let index = 0; index < scores.length && index < names.length; index++)`+loopBody,
			`for (const [score, name_] of Iterator.zip([scores, names]))`+zipBody, 1, 1, 1, 76),
		expectZip("input.js", "const length = Math.min(names.length, scores.length);\nfor (let index = 0; index < length; index++)"+loopBody,
			zipHeader+zipBody, 2, 1, 2, 45),
		expectZip("input.js", "const length = Math.min(names.length, scores.length);\nfor (let index = 0; length > index; index++)"+loopBody,
			zipHeader+zipBody, 2, 1, 2, 45),
		expectZip("input.js", `for (let index = 0, length = Math.min(names.length, scores.length); index < length; index++)`+loopBody,
			zipHeader+zipBody, 1, 1, 1, 93),
		expectZip("input.js", "const length = Math.min(names.length, scores.length); /* keep cached-bound comment */\nfor (let index = 0; index < length; index++)"+loopBody,
			"", 2, 1, 2, 45),
		expectZip("input.js", "let length = Math.min(names.length, scores.length);\nfor (let index = 0; index < length; index++)"+loopBody,
			zipHeader+zipBody, 2, 1, 2, 45),
		expectZip("file.ts", "const length: number = Math.min(names.length, scores.length);\nfor (let index = 0; index < length; index++)"+loopBody,
			zipHeader+zipBody, 2, 1, 2, 45),
		expectZip("input.js", loopHeader+` { for (const letter of names[index]) { save(letter, scores[index]); } }`,
			zipHeader+` { for (const letter of name_) { save(letter, score); } }`, 1, 1, 1, 76),
		// Documentation example.
		expectZip("input.js", loopHeader+" {\n\tsave(names[index], scores[index]);\n}",
			zipHeader+" {\n\tsave(name_, score);\n}", 1, 1, 1, 76),
	}
	// The upstream snapshot runner configures the browser global name.
	for i := range invalid {
		invalid[i].Globals = map[string]any{"name": "readonly"}
	}
	invalid[0].LanguageOptions = rule.LanguageOptions{ECMAVersion: 2022}
	invalid = append(invalid,
		// Standalone token-boundary and autofix cases.
		expectZip("input.js", `function pair(names, scores) { `+loopHeader+` { save(typeof(names)[index], scores[index]); return(names)[index]; } }`,
			`function pair(names, scores) { for (const [name, score] of Iterator.zip([names, scores])) { save(typeof name, score); return name; } }`, 1, 32, 1, 107),
		expectZip("input.js", loopHeader+loopBody,
			`for (const [name, score] of Iterator.zip([names, scores])) { save(name, score); }`, 1, 1, 1, 76),
	)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_iterator_zip.PreferIteratorZipRule, valid, invalid)
}

func TestUpstreamCompanionRulesAndPresets(t *testing.T) {
	t.Skip("The upstream integration case needs companion rules missing from rslint and Unicorn v79 presets. The target rule's diagnostic and edits are covered above.")
}
