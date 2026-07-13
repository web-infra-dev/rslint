// TestNoUnusedPrivateClassMembersUpstream migrates the full valid/invalid suite from
// upstream eslint/tests/lib/rules/no-unused-private-class-members.js 1:1.
// Position assertions cover line/column/endLine/endColumn for every invalid
// case. rslint-specific lock-in cases live in the
// no_unused_private_class_members_extras_test.go file.
package no_unused_private_class_members

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedPrivateClassMembersUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedPrivateClassMembersRule,
		[]rule_tester.ValidTestCase{
			// ---- valid ----
			{Code: `class Foo {}`},
			{Code: `class Foo {
    publicMember = 42;
}`},
			{Code: `class Foo {
    #usedMember = 42;
    method() {
        return this.#usedMember;
    }
}`},
			{Code: `class Foo {
    #usedMember = 42;
    anotherMember = this.#usedMember;
}`},
			{Code: `class Foo {
    #usedMember = 42;
    foo() {
        anotherMember = this.#usedMember;
    }
}`},
			{Code: `class C {
    #usedMember;

    foo() {
        bar(this.#usedMember += 1);
    }
}`},
			{Code: `class Foo {
    #usedMember = 42;
    method() {
        return someGlobalMethod(this.#usedMember);
    }
}`},
			{Code: `class C {
    #usedInOuterClass;

    foo() {
        return class {};
    }

    bar() {
        return this.#usedInOuterClass;
    }
}`},
			{Code: `class Foo {
    #usedInForInLoop;
    method() {
        for (const bar in this.#usedInForInLoop) {

        }
    }
}`},
			{Code: `class Foo {
    #usedInForOfLoop;
    method() {
        for (const bar of this.#usedInForOfLoop) {

        }
    }
}`},
			{Code: `class Foo {
    #usedInAssignmentPattern;
    method() {
        [bar = 1] = this.#usedInAssignmentPattern;
    }
}`},
			{Code: `class Foo {
    #usedInArrayPattern;
    method() {
        [bar] = this.#usedInArrayPattern;
    }
}`},
			{Code: `class Foo {
    #usedInAssignmentPattern;
    method() {
        [bar] = this.#usedInAssignmentPattern;
    }
}`},
			{Code: `class C {
    #usedInObjectAssignment;

    method() {
        ({ [this.#usedInObjectAssignment]: a } = foo);
    }
}`},
			{Code: `class C {
    set #accessorWithSetterFirst(value) {
        doSomething(value);
    }
    get #accessorWithSetterFirst() {
        return something();
    }
    method() {
        this.#accessorWithSetterFirst += 1;
    }
}`},
			{Code: `class Foo {
    set #accessorUsedInMemberAccess(value) {}

    method(a) {
        [this.#accessorUsedInMemberAccess] = a;
    }
}`},
			{Code: `class C {
    get #accessorWithGetterFirst() {
        return something();
    }
    set #accessorWithGetterFirst(value) {
        doSomething(value);
    }
    method() {
        this.#accessorWithGetterFirst += 1;
    }
}`},
			{Code: `class C {
    #usedInInnerClass;

    method(a) {
        return class {
            foo = a.#usedInInnerClass;
        }
    }
}`},

			// ---- Method definitions ----
			{Code: `class Foo {
    #usedMethod() {
        return 42;
    }
    anotherMethod() {
        return this.#usedMethod();
    }
}`},
			{Code: `class C {
    set #x(value) {
        doSomething(value);
    }

    foo() {
        this.#x = 1;
    }
}`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid ----
			invalidUnusedPrivateCase(`class Foo {
    #unusedMember = 5;
}`,
				unusedPrivateError("#unusedMember", 2, 5, 2, 18, `class Foo {
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    /** docs */
    #unusedMember = 1;
}`,
				unusedPrivateError("#unusedMember", 3, 5, 3, 18, `class Foo {
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    // remove me
    #unusedMember = 1;
}`,
				unusedPrivateError("#unusedMember", 3, 5, 3, 18, `class Foo {
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    /* remove */ #unusedMember = 1;
}`,
				unusedPrivateError("#unusedMember", 2, 18, 2, 31, `class Foo {
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    /* keep */ #unusedMember = 1; foo = 1
}`,
				unusedPrivateError("#unusedMember", 2, 16, 2, 29, `class Foo {
    /* keep */ foo = 1
}`),
			),
			invalidUnusedPrivateCase(`class C {
    #unused1; /* keep */ foo;
}`,
				unusedPrivateError("#unused1", 2, 5, 2, 13, `class C {
    /* keep */ foo;
}`),
			),
			invalidUnusedPrivateCase(`class C {
    bar; #unused2; // keep
}`,
				unusedPrivateError("#unused2", 2, 10, 2, 18, `class C {
    bar; // keep
}`),
			),
			invalidUnusedPrivateCase(`class C {
    // comment
    #unused; foo;
}`,
				unusedPrivateError("#unused", 3, 5, 3, 12, `class C {
    // comment
    foo;
}`),
			),
			invalidUnusedPrivateCase(`class C {
    // comment
    #unused; /*
    */ foo;
}`,
				unusedPrivateError("#unused", 3, 5, 3, 12, `class C {
    foo;
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    #unusedMember = 1; // trailing
}`,
				unusedPrivateError("#unusedMember", 2, 5, 2, 18, `class Foo {
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    foo = 1; /*
    */ #unusedMember = 1;
}`,
				unusedPrivateError("#unusedMember", 3, 8, 3, 21, `class Foo {
    foo = 1; /*
    */ 
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    foo = 1; // keep this
    #unusedMember = 1;
}`,
				unusedPrivateError("#unusedMember", 3, 5, 3, 18, `class Foo {
    foo = 1; // keep this
}`),
			),
			invalidUnusedPrivateCase(`class First {}
class Second {
    #unusedMemberInSecondClass = 5;
}`,
				unusedPrivateError("#unusedMemberInSecondClass", 3, 5, 3, 31, `class First {}
class Second {
}`),
			),
			invalidUnusedPrivateCase(`class First {
    #unusedMemberInFirstClass = 5;
}
class Second {}`,
				unusedPrivateError("#unusedMemberInFirstClass", 2, 5, 2, 30, `class First {
}
class Second {}`),
			),
			invalidUnusedPrivateCase(`class First {
    #firstUnusedMemberInSameClass = 5;
    #secondUnusedMemberInSameClass = 5;
}`,
				unusedPrivateError("#firstUnusedMemberInSameClass", 2, 5, 2, 34, `class First {
    #secondUnusedMemberInSameClass = 5;
}`),
				unusedPrivateError("#secondUnusedMemberInSameClass", 3, 5, 3, 35, `class First {
    #firstUnusedMemberInSameClass = 5;
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    #usedOnlyInWrite = 5;
    method() {
        this.#usedOnlyInWrite = 42;
    }
}`,
				unusedPrivateError("#usedOnlyInWrite", 2, 5, 2, 21),
			),
			invalidUnusedPrivateCase(`class Foo {
    #usedOnlyInWriteStatement = 5;
    method() {
        this.#usedOnlyInWriteStatement += 42;
    }
}`,
				unusedPrivateError("#usedOnlyInWriteStatement", 2, 5, 2, 30),
			),
			invalidUnusedPrivateCase(`class C {
    #usedOnlyInIncrement;

    foo() {
        this.#usedOnlyInIncrement++;
    }
}`,
				unusedPrivateError("#usedOnlyInIncrement", 2, 5, 2, 25),
			),
			invalidUnusedPrivateCase(`class C {
    #unusedInOuterClass;

    foo() {
        return class {
            #unusedInOuterClass;

            bar() {
                return this.#unusedInOuterClass;
            }
        };
    }
}`,
				unusedPrivateError("#unusedInOuterClass", 2, 5, 2, 24, `class C {
    foo() {
        return class {
            #unusedInOuterClass;

            bar() {
                return this.#unusedInOuterClass;
            }
        };
    }
}`),
			),
			invalidUnusedPrivateCase(`class C {
    #unusedOnlyInSecondNestedClass;

    foo() {
        return class {
            #unusedOnlyInSecondNestedClass;

            bar() {
                return this.#unusedOnlyInSecondNestedClass;
            }
        };
    }

    baz() {
        return this.#unusedOnlyInSecondNestedClass;
    }

    bar() {
        return class {
            #unusedOnlyInSecondNestedClass;
        }
    }
}`,
				unusedPrivateError("#unusedOnlyInSecondNestedClass", 20, 13, 20, 43, `class C {
    #unusedOnlyInSecondNestedClass;

    foo() {
        return class {
            #unusedOnlyInSecondNestedClass;

            bar() {
                return this.#unusedOnlyInSecondNestedClass;
            }
        };
    }

    baz() {
        return this.#unusedOnlyInSecondNestedClass;
    }

    bar() {
        return class {
        }
    }
}`),
			),

			// ---- Unused method definitions ----
			invalidUnusedPrivateCase(`class Foo {
    #unusedMethod() {}
}`,
				unusedPrivateError("#unusedMethod", 2, 5, 2, 18, `class Foo {
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    #unusedMethod() {}
    #usedMethod() {
        return 42;
    }
    publicMethod() {
        return this.#usedMethod();
    }
}`,
				unusedPrivateError("#unusedMethod", 2, 5, 2, 18, `class Foo {
    #usedMethod() {
        return 42;
    }
    publicMethod() {
        return this.#usedMethod();
    }
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    set #unusedSetter(value) {}
}`,
				unusedPrivateError("#unusedSetter", 2, 9, 2, 22, `class Foo {
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    get #unusedAccessor() {
        return 1;
    }
    set #unusedAccessor(value) {}
}`,
				unusedPrivateError("#unusedAccessor", 5, 9, 5, 24, `class Foo {
    get #unusedAccessor() {
        return 1;
    }
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    #unusedForInLoop;
    method() {
        for (this.#unusedForInLoop in bar) {

        }
    }
}`,
				unusedPrivateError("#unusedForInLoop", 2, 5, 2, 21),
			),
			invalidUnusedPrivateCase(`class Foo {
    #unusedForOfLoop;
    method() {
        for (this.#unusedForOfLoop of bar) {

        }
    }
}`,
				unusedPrivateError("#unusedForOfLoop", 2, 5, 2, 21),
			),
			invalidUnusedPrivateCase(`class Foo {
    #unusedInDestructuring;
    method() {
        ({ x: this.#unusedInDestructuring } = bar);
    }
}`,
				unusedPrivateError("#unusedInDestructuring", 2, 5, 2, 27),
			),
			invalidUnusedPrivateCase(`class Foo {
    #unusedInRestPattern;
    method() {
        [...this.#unusedInRestPattern] = bar;
    }
}`,
				unusedPrivateError("#unusedInRestPattern", 2, 5, 2, 25),
			),
			invalidUnusedPrivateCase(`class Foo {
    #unusedInAssignmentPattern;
    method() {
        [this.#unusedInAssignmentPattern = 1] = bar;
    }
}`,
				unusedPrivateError("#unusedInAssignmentPattern", 2, 5, 2, 31),
			),
			invalidUnusedPrivateCase(`class Foo {
    #unusedInAssignmentPattern;
    method() {
        [this.#unusedInAssignmentPattern] = bar;
    }
}`,
				unusedPrivateError("#unusedInAssignmentPattern", 2, 5, 2, 31),
			),
			invalidUnusedPrivateCase(`class Foo {
    foo = 1
    #unusedMethod() {}
    [0]() {}
}`,
				unusedPrivateError("#unusedMethod", 3, 5, 3, 18, `class Foo {
    foo = 1;
    [0]() {}
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    foo = 1
    #unusedMethod() {}
    *generator() {}
}`,
				unusedPrivateError("#unusedMethod", 3, 5, 3, 18, `class Foo {
    foo = 1;
    *generator() {}
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    foo = 1
    #unusedMethod() {}
    in = 2
}`,
				unusedPrivateError("#unusedMethod", 3, 5, 3, 18, `class Foo {
    foo = 1;
    in = 2
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    foo = 1
    #unused
    instanceof() {}
}`,
				unusedPrivateError("#unused", 3, 5, 3, 12, `class Foo {
    foo = 1;
    instanceof() {}
}`),
			),
			invalidUnusedPrivateCase(`class C {
    foo = () => {}
    #unused
    [bar]
}`,
				unusedPrivateError("#unused", 3, 5, 3, 12, `class C {
    foo = () => {}
    [bar]
}`),
			),
			invalidUnusedPrivateCase(`class C {
    foo
    #unused
    [bar]
}`,
				unusedPrivateError("#unused", 3, 5, 3, 12, `class C {
    foo
    [bar]
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    foo = 1
    /** docs */
    #unusedMethod() {}
    [0]() {}
}`,
				unusedPrivateError("#unusedMethod", 4, 5, 4, 18, `class Foo {
    foo = 1;
    [0]() {}
}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    // keep

    /** remove */
    #unusedMember = 1;
}`,
				unusedPrivateError("#unusedMember", 5, 5, 5, 18, `class Foo {
    // keep

}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    // keep one
    // keep two

    /** remove */
    #unusedMember = 1;
}`,
				unusedPrivateError("#unusedMember", 6, 5, 6, 18, `class Foo {
    // keep one
    // keep two

}`),
			),
			invalidUnusedPrivateCase(`class Foo {
    // maybe unrelated

    #unusedMember = 1;
}`,
				unusedPrivateError("#unusedMember", 4, 5, 4, 18, `class Foo {
    // maybe unrelated

}`),
			),
			invalidUnusedPrivateCase(`class C {
    #usedOnlyInTheSecondInnerClass;

    method(a) {
        return class {
            #usedOnlyInTheSecondInnerClass;

            method2(b) {
                foo = b.#usedOnlyInTheSecondInnerClass;
            }

            method3(b) {
                foo = b.#usedOnlyInTheSecondInnerClass;
            }
        }
    }
}`,
				unusedPrivateError("#usedOnlyInTheSecondInnerClass", 2, 5, 2, 35, `class C {
    method(a) {
        return class {
            #usedOnlyInTheSecondInnerClass;

            method2(b) {
                foo = b.#usedOnlyInTheSecondInnerClass;
            }

            method3(b) {
                foo = b.#usedOnlyInTheSecondInnerClass;
            }
        }
    }
}`),
			),
		},
	)
}

func invalidUnusedPrivateCase(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:   code,
		Errors: errors,
	}
}

func unusedPrivateError(name string, line int, column int, endLine int, endColumn int, suggestionOutput ...string) rule_tester.InvalidTestCaseError {
	err := rule_tester.InvalidTestCaseError{
		MessageId: "unusedPrivateClassMember",
		Message:   fmt.Sprintf("'%s' is defined but never used.", name),
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
	if len(suggestionOutput) > 0 {
		err.Suggestions = []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "removeUnusedPrivateClassMember", Output: suggestionOutput[0]},
		}
	}
	return err
}
