// TestNoThisAssignmentUpstream migrates the full valid/invalid suite from
// upstream test/no-this-assignment.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// no_this_assignment_extras_test.go.
package no_this_assignment_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_this_assignment"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "no-this-assignment"

func TestNoThisAssignmentUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_this_assignment.NoThisAssignmentRule,
		[]rule_tester.ValidTestCase{
			{Code: `const {property} = this;`},
			{Code: `const property = this.property;`},
			{Code: `const [element] = this;`},
			{Code: `const element = this[0];`},
			{Code: `([element] = this);`},
			{Code: `element = this[0];`},
			{Code: `property = this.property;`},
			{Code: `const [element] = [this];`},
			{Code: `([element] = [this]);`},
			{Code: `const {property} = {property: this};`},
			{Code: `({property} = {property: this});`},
			{Code: `const self = true && this;`},
			{Code: `const self = false || this;`},
			{Code: `const self = false ?? this;`},
			{Code: `foo.bar = this;`},
			{Code: `function foo(a = this) {}`},
			{Code: `function foo({a = this}) {}`},
			{Code: `function foo([a = this]) {}`},
			{Code: "class A {\n\tfoo = this;\n}"},
			{Code: "class A {\n\tstatic foo = this;\n}"},
		},
		[]rule_tester.InvalidTestCase{
			invalid(`const foo = this;`, `foo = this`, "foo"),
			invalid(`let foo;foo = this;`, `foo = this`, "foo"),
			invalid(`var foo = bar, baz = this;`, `baz = this`, "baz"),
		},
	)
}

func invalid(code, target, name string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, name),
		},
	}
}

func expectedError(code, target, name string) rule_tester.InvalidTestCaseError {
	offset := strings.Index(code, target)
	if offset < 0 {
		panic("target not found in no-this-assignment test: " + target)
	}

	line, column := lineColumn(code, offset)
	endLine, endColumn := lineColumn(code, offset+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   fmt.Sprintf("Do not assign `this` to `%s`.", name),
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func lineColumn(code string, offset int) (int, int) {
	line := 1
	column := 1
	for index, character := range code {
		if index >= offset {
			break
		}
		if character == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
