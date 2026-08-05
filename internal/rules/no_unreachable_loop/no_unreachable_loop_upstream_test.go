package no_unreachable_loop

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// loopTemplates, validLoopBodies, and invalidLoopBodies mirror upstream's
// tests/lib/rules/no-unreachable-loop.js, which builds its basic cases as the
// cross product of every loop shape and every body.
var loopTemplates = []struct {
	kind      string
	templates []string
}{
	{"WhileStatement", []string{
		"while (a) <body>",
		"while (a && b) <body>",
	}},
	{"DoWhileStatement", []string{
		"do <body> while (a)",
		"do <body> while (a && b)",
	}},
	{"ForStatement", []string{
		"for (a; b; c) <body>",
		"for (var i = 0; i < a.length; i++) <body>",
		"for (; b; c) <body>",
		"for (; b < foo; c++) <body>",
		"for (a; ; c) <body>",
		"for (a = 0; ; c++) <body>",
		"for (a; b;) <body>",
		"for (a = 0; b < foo; ) <body>",
		"for (; ; c) <body>",
		"for (; ; c++) <body>",
		"for (; b;) <body>",
		"for (; b < foo; ) <body>",
		"for (a; ;) <body>",
		"for (a = 0; ;) <body>",
		"for (;;) <body>",
	}},
	{"ForInStatement", []string{
		"for (a in b) <body>",
		"for (a in f(b)) <body>",
		"for (var a in b) <body>",
		"for (let a in f(b)) <body>",
	}},
	{"ForOfStatement", []string{
		"for (a of b) <body>",
		"for (a of f(b)) <body>",
		"for ({ a, b } of c) <body>",
		"for (var a of f(b)) <body>",
		"async function foo() { for await (const a of b) <body> }",
	}},
}

var validLoopBodies = []string{
	";",
	"{}",
	"{ bar(); }",
	"continue;",
	"{ continue; }",
	"{ if (foo) break; }",
	"{ if (foo) { return; } bar(); }",
	"{ if (foo) { bar(); } else { break; } }",
	"{ if (foo) { continue; } return; }",
	"{ switch (foo) { case 1: return; } }",
	"{ switch (foo) { case 1: break; default: return; } }",
	"{ switch (foo) { case 1: continue; default: return; } throw err; }",
	"{ try { return bar(); } catch (e) {} }",

	// unreachable break
	"{ continue; break; }",

	// functions in loops
	"() => a;",
	"{ () => a }",
	"(() => a)();",
	"{ (() => a)() }",

	// loops in loops
	"while (a);",
	"do ; while (a)",
	"for (a; b; c);",
	"for (; b;);",
	"for (; ; c) if (foo) break;",
	"for (;;) if (foo) break;",
	"while (true) if (foo) break;",
	"while (foo) if (bar) return;",
	"for (a in b);",
	"for (a of b);",
}

var invalidLoopBodies = []string{
	"break;",
	"{ break; }",
	"return;",
	"{ return; }",
	"throw err;",
	"{ throw err; }",
	"{ foo(); break; }",
	"{ break; foo(); }",
	"if (foo) break; else return;",
	"{ if (foo) { return; } else { break; } bar(); }",
	"{ if (foo) { return; } throw err; }",
	"{ switch (foo) { default: throw err; } }",
	"{ switch (foo) { case 1: throw err; default: return; } }",
	"{ switch (foo) { case 1: something(); default: return; } }",
	"{ try { return bar(); } catch (e) { break; } }",

	// unreachable continue
	"{ break; continue; }",

	// functions in loops
	"{ () => a; break; }",
	"{ (() => a)(); break; }",

	// loops in loops
	"{ while (a); break; }",
	"{ do ; while (a) break; }",
	"{ for (a; b; c); break; }",
	"{ for (; b;); break; }",
	"{ for (; ; c) if (foo) break; break; }",
	"{ for(;;) if (foo) break; break; }",
	"{ for (a in b); break; }",
	"{ for (a of b); break; }",

	// Additional cases where code path analysis detects unreachable code: after
	// loops that don't have a test condition or have a constant truthy test
	// condition, and at the same time don't have any exit statements in the
	// body. These are special cases where this rule reports an error not
	// because the outer loop's body exits in all paths, but because it has an
	// infinite loop inside, thus it (the outer loop) cannot have more than one
	// iteration.
	"for (;;);",
	"{ for (var i = 0; ; i< 10) { foo(); } }",
	"while (true);",
}

// getSourceCode mirrors upstream's helper of the same name: a body that returns
// needs a function around it, unless the template already provides one.
func getSourceCode(template, body string) string {
	loop := strings.Replace(template, "<body>", body, 1)
	if strings.Contains(body, "return") && !strings.Contains(template, "function") {
		return "function someFunc() { " + loop + " }"
	}
	return loop
}

func basicValidTests() []rule_tester.ValidTestCase {
	var cases []rule_tester.ValidTestCase
	for _, group := range loopTemplates {
		for _, template := range group.templates {
			for _, body := range validLoopBodies {
				cases = append(cases, rule_tester.ValidTestCase{Code: getSourceCode(template, body)})
			}
		}
	}
	return cases
}

func basicInvalidTests() []rule_tester.InvalidTestCase {
	var cases []rule_tester.InvalidTestCase
	for _, group := range loopTemplates {
		for _, template := range group.templates {
			for _, body := range invalidLoopBodies {
				cases = append(cases, rule_tester.InvalidTestCase{
					Code:   getSourceCode(template, body),
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalid"}},
				})
			}
		}
	}
	return cases
}

func TestNoUnreachableLoopUpstream(t *testing.T) {
	valid := append(basicValidTests(), []rule_tester.ValidTestCase{
		// out of scope for the code path analysis and consequently out of scope for this rule
		{Code: `while (false) { foo(); }`},
		{Code: `while (bar) { foo(); if (true) { break; } }`},
		{Code: `do foo(); while (false)`},
		{Code: `for (x = 1; x < 10; i++) { if (x > 0) { foo(); throw err; } }`},
		{Code: `for (x of []);`},
		{Code: `for (x of [1]);`},

		// doesn't report unreachable loop statements, regardless of whether they would be valid or not in a reachable position
		{Code: `function foo() { return; while (a); }`},
		{Code: `function foo() { return; while (a) break; }`},
		{Code: `while(true); while(true);`},
		{Code: `while(true); while(true) break;`},

		// "ignore"
		{
			Code:    `while (a) break;`,
			Options: map[string]interface{}{"ignore": []interface{}{"WhileStatement"}},
		},
		{
			Code:    `do break; while (a)`,
			Options: map[string]interface{}{"ignore": []interface{}{"DoWhileStatement"}},
		},
		{
			Code:    `for (a; b; c) break;`,
			Options: map[string]interface{}{"ignore": []interface{}{"ForStatement"}},
		},
		{
			Code:    `for (a in b) break;`,
			Options: map[string]interface{}{"ignore": []interface{}{"ForInStatement"}},
		},
		{
			Code:    `for (a of b) break;`,
			Options: map[string]interface{}{"ignore": []interface{}{"ForOfStatement"}},
		},
		{
			Code:    `for (var key in obj) { hasEnumerableProperties = true; break; } for (const a of b) break;`,
			Options: map[string]interface{}{"ignore": []interface{}{"ForInStatement", "ForOfStatement"}},
		},
		{
			Code: `while (a) break;`,
			Options: map[string]interface{}{"ignore": []interface{}{
				"WhileStatement",
				"DoWhileStatement",
				"ForStatement",
				"ForInStatement",
				"ForOfStatement",
			}},
		},
	}...)

	invalid := append(basicInvalidTests(), []rule_tester.InvalidTestCase{
		// invalid loop nested in a valid loop (valid in valid, and valid in invalid are covered by basic tests)
		{
			Code: `while (foo) { for (a of b) { if (baz) { break; } else { throw err; } } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
			},
		},
		{
			Code: `lbl: for (var i = 0; i < 10; i++) { while (foo) break lbl; } /* outer is valid because inner can have 0 iterations */`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
			},
		},

		// invalid loop nested in another invalid loop
		{
			Code: `for (a in b) { while (foo) { if(baz) { break; } else { break; } } break; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
				{MessageId: "invalid"},
			},
		},

		// loop and nested loop both invalid because of the same exit statement
		{
			Code: `function foo() { for (var i = 0; i < 10; i++) { do { return; } while(i) } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
				{MessageId: "invalid"},
			},
		},
		{
			Code: `lbl: while(foo) { do { break lbl; } while(baz) }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
				{MessageId: "invalid"},
			},
		},

		// inner loop has continue, but to an outer loop
		{
			Code: `lbl: for (a in b) { while(foo) { continue lbl; } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
			},
		},

		// edge cases - inner loop has only one exit path, but at the same time it exits the outer loop in the first iteration
		{
			Code: `for (a of b) { for(;;) { if (foo) { throw err; } } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
			},
		},
		{
			Code: `function foo () { for (a in b) { while (true) { if (bar) { return; } } } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
			},
		},

		// edge cases where parts of the loops belong to the same code path segment, tests for false negatives
		{
			Code: `do for (var i = 1; i < 10; i++) break; while(foo)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
			},
		},
		{
			Code: `do { for (var i = 1; i < 10; i++) continue; break; } while(foo)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
			},
		},
		{
			Code: `for (;;) { for (var i = 1; i < 10; i ++) break; if (foo) break; continue; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid", Column: 12},
			},
		},

		// "ignore"
		{
			Code:    `while (a) break; do break; while (b); for (;;) break; for (c in d) break; for (e of f) break;`,
			Options: map[string]interface{}{"ignore": []interface{}{}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
				{MessageId: "invalid"},
				{MessageId: "invalid"},
				{MessageId: "invalid"},
				{MessageId: "invalid"},
			},
		},
		{
			Code:    `while (a) break;`,
			Options: map[string]interface{}{"ignore": []interface{}{"DoWhileStatement"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
			},
		},
		{
			Code:    `do break; while (a)`,
			Options: map[string]interface{}{"ignore": []interface{}{"WhileStatement"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
			},
		},
		{
			Code:    `for (a in b) break; for (c of d) break;`,
			Options: map[string]interface{}{"ignore": []interface{}{"ForStatement"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
				{MessageId: "invalid"},
			},
		},
		{
			Code:    `for (a in b) break; for (;;) break; for (c of d) break;`,
			Options: map[string]interface{}{"ignore": []interface{}{"ForInStatement", "ForOfStatement"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalid"},
			},
		},
	}...)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnreachableLoopRule,
		valid,
		invalid,
	)
}
