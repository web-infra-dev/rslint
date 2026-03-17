package no_this_before_super

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoThisBeforeSuperRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoThisBeforeSuperRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Non-derived class
			{Code: `class A { constructor() { this.b = 0; } }`},
			// Class with no constructor
			{Code: `class A extends B { }`},
			// Class extends null
			{Code: `class A extends null { constructor() { } }`},
			// super() before this
			{Code: `class A extends B { constructor() { super(); this.c = 0; } }`},
			// super() before super.c()
			{Code: `class A extends B { constructor() { super(); super.c(); } }`},
			// super() in both branches of if/else
			{Code: `class A extends B { constructor() { if (true) { super(); } else { super(); } this.c(); } }`},
			// Nested class has its own scope
			{Code: `class A extends B { constructor() { class C extends D { constructor() { super(); this.d = 0; } } super(); } }`},
			// this in nested function
			{Code: `class A extends B { constructor() { function c() { this.d(); } super(); } }`},
			// this in nested arrow function
			{Code: `class A extends B { constructor() { var c = () => this.d; super(); } }`},
			// super() in ternary on both branches
			{Code: `class A extends B { constructor() { a ? super() : super(); this.c = 0; } }`},
			// super() before this in nested expression
			{Code: `class A extends B { constructor() { super(); this.a = [this.b, this.c]; } }`},
			// Multiple this and super member access after super()
			{Code: `class A extends B { constructor() { super(); this.a = 1; super.b(); } }`},
			// super in all branches with variable condition
			{Code: `class A extends B { constructor() { if (foo) { super(); } else { super(); } this.c = 0; } }`},
			// Non-derived class with empty constructor
			{Code: `class A { constructor() { } }`},
			// Basic super call without this
			{Code: `class A extends B { constructor() { super(); } }`},
			// Non-constructor method
			{Code: `class A extends B { method() { this.x = 1; } }`},
			// Arrow function (no this usage) before super
			{Code: `class A extends B { constructor() { (() => { })(); super(); } }`},
			// do-while body always executes
			{Code: "class A extends B {\n  constructor() {\n    do { super(); } while (false);\n    this.c = 0;\n  }\n}"},
			// Parenthesized ternary with super() in both branches
			{Code: "class A extends B {\n  constructor(cond) {\n    (cond ? super() : super());\n    this.a = 0;\n  }\n}"},
			// this in object getter before super (getter creates scope boundary)
			{Code: "class A extends B {\n  constructor() {\n    const obj = { get foo() { return this.bar; } };\n    super();\n  }\n}"},
			// this in object setter before super (setter creates scope boundary)
			{Code: "class A extends B {\n  constructor() {\n    const obj = { set foo(v) { this.bar = v; } };\n    super();\n  }\n}"},
			// labeled statement with super
			{Code: "class A extends B {\n  constructor() {\n    label: super();\n    this.a = 0;\n  }\n}"},
			// comma expression: super() first, then this
			{Code: "class A extends B {\n  constructor() {\n    super(), this.a = 0;\n  }\n}"},
			// for-loop initializer with super
			{Code: "class A extends B {\n  constructor() {\n    for (super();;) { break; }\n    this.a = 0;\n  }\n}"},
			// variable declaration with super() initializer
			{Code: "class A extends B {\n  constructor() {\n    let x = super();\n    this.a = 0;\n  }\n}"},
			// this inside object method before super (method is scope boundary)
			{Code: "class A extends B {\n  constructor() {\n    const obj = { method() { this.x = 1; } };\n    super();\n  }\n}"},
			// super() in finally block
			{Code: "class A extends B {\n  constructor() {\n    try { } catch(e) { } finally { super(); }\n    this.a = 0;\n  }\n}"},
			// nested ternary with super in all leaves
			{Code: "class A extends B {\n  constructor(a, b) {\n    a ? (b ? super() : super()) : super();\n    this.c = 0;\n  }\n}"},
			// this in async function before super (scope boundary)
			{Code: "class A extends B {\n  constructor() {\n    const fn = async function() { this.a = 0; };\n    super();\n  }\n}"},
			// this in generator before super (scope boundary)
			{Code: "class A extends B {\n  constructor() {\n    const fn = function*() { this.a = 0; };\n    super();\n  }\n}"},
			// arrow function in expression (arrow is boundary)
			{Code: "class A extends B {\n  constructor() {\n    [1,2].forEach(x => { this.a = x; });\n    super();\n  }\n}"},
			// return before this (no this reached)
			{Code: "class A extends B {\n  constructor(a) {\n    if (a) return;\n    super();\n    this.a = 0;\n  }\n}"},
			// throw before this (unreachable)
			{Code: "class A extends B {\n  constructor() {\n    throw new Error();\n    this.a = 0;\n  }\n}"},
			// super() deep in assignment chain
			{Code: "class A extends B {\n  constructor() {\n    let a, b;\n    a = b = super();\n    this.x = 0;\n  }\n}"},
			// super() in first var decl, this in second
			{Code: "class A extends B {\n  constructor() {\n    let a = super(), b = this.x;\n  }\n}"},
			// super in if/else both branches, this after
			{Code: "class A extends B {\n  constructor(a) {\n    if (a) {\n      super();\n    } else {\n      super();\n    }\n    this.a = 0;\n  }\n}"},
			// chained method on super() return value
			{Code: "class A extends B {\n  constructor() {\n    super().toString();\n    this.a = 0;\n  }\n}"},
			// super() ?? fallback, this after
			{Code: "class A extends B {\n  constructor() {\n    super() ?? null;\n    this.x = 0;\n  }\n}"},
			// this in async arrow (scope boundary)
			{Code: "class A extends B {\n  constructor() {\n    const fn = async () => { await this.a; };\n    super();\n  }\n}"},
			// if-else if-else all with super (valid)
			{Code: "class A extends B {\n  constructor(a) {\n    if (a === 1) {\n      super();\n    } else if (a === 2) {\n      super();\n    } else {\n      super();\n    }\n    this.x = 0;\n  }\n}"},
			// nested try with super in inner finally
			{Code: "class A extends B {\n  constructor() {\n    try {\n      try { } finally { super(); }\n    } catch(e) { super(); }\n    this.a = 0;\n  }\n}"},
			// for-loop incrementor unreachable due to unconditional break
			{Code: "class A extends B {\n  constructor() {\n    for (let i = 0; i < 3; this.inc()) { break; }\n    super();\n  }\n}"},
			// super() || fallback, this after (valid - super always called on left)
			{Code: "class A extends B {\n  constructor() {\n    super() || null;\n    this.x = 0;\n  }\n}"},
			// super() && something, this after (valid)
			{Code: "class A extends B {\n  constructor() {\n    super() && this.init();\n    this.x = 0;\n  }\n}"},
			// super() in object literal property value
			{Code: "class A extends B {\n  constructor() {\n    const obj = { key: super() };\n    this.x = 0;\n  }\n}"},
			// super() as function argument
			{Code: "class A extends B {\n  constructor() {\n    foo(super());\n    this.x = 0;\n  }\n}"},
			// super() in template literal
			{Code: "class A extends B {\n  constructor() {\n    const x = `${super()}`;\n    this.x = 0;\n  }\n}"},
			// super() in new expression argument
			{Code: "class A extends B {\n  constructor() {\n    new Foo(super());\n    this.x = 0;\n  }\n}"},
			// comma in for-loop initializer with super
			{Code: "class A extends B {\n  constructor() {\n    let i;\n    for (i = 0, super();;) { break; }\n    this.x = 0;\n  }\n}"},
			// super() chained deeply: super().foo.bar
			{Code: "class A extends B {\n  constructor() {\n    super().foo.bar;\n    this.a = 0;\n  }\n}"},
			// super() chained: super().foo()
			{Code: "class A extends B {\n  constructor() {\n    super().foo();\n    this.a = 0;\n  }\n}"},
			// switch with case fallthrough to super
			{Code: "class A extends B {\n  constructor(x) {\n    switch(x) {\n      case 1:\n      case 2: super(); break;\n      default: super(); break;\n    }\n    this.a = 0;\n  }\n}"},
			// while condition with super
			{Code: "class A extends B {\n  constructor() {\n    while(super()) { break; }\n    this.a = 0;\n  }\n}"},
			// try/catch both have super
			{Code: "class A extends B {\n  constructor() {\n    try { super(); } catch(e) { super(); }\n    this.a = 0;\n  }\n}"},
			// spread with super in array
			{Code: "class A extends B {\n  constructor() {\n    foo(...[super()]);\n    this.x = 0;\n  }\n}"},
			// optional chaining on super() return
			{Code: "class A extends B {\n  constructor() {\n    super()?.foo;\n    this.x = 0;\n  }\n}"},
			// nested class expression (scope boundary)
			{Code: "class A extends B {\n  constructor() {\n    const C = class { constructor() { this.a = 0; } };\n    super();\n  }\n}"},
			// arrow returns this (scope boundary)
			{Code: "class A extends B {\n  constructor() {\n    const fn = () => this;\n    super();\n  }\n}"},
			// [a] = [super()]; this after (valid)
			{Code: "class A extends B {\n  constructor() {\n    let a;\n    [a] = [super()];\n    this.x = 0;\n  }\n}"},
			// deeply nested if-else all branches super (valid)
			{Code: "class A extends B {\n  constructor(a, b) {\n    if (a) {\n      if (b) {\n        super();\n      } else {\n        super();\n      }\n    } else {\n      super();\n    }\n    this.x = 0;\n  }\n}"},
			// deeply nested ternary all leaves super (valid)
			{Code: "class A extends B {\n  constructor(a, b, c) {\n    a ? (b ? (c ? super() : super()) : super()) : super();\n    this.x = 0;\n  }\n}"},
			// super() + this.a (valid - super on left of binary op)
			{Code: "class A extends B {\n  constructor() {\n    const x = super() + this.a;\n  }\n}"},
			// super() in ternary condition (valid - always evaluated before branches)
			{Code: "class A extends B {\n  constructor() {\n    super() ? this.a : this.b;\n  }\n}"},
			// do {} while(super()); this after (valid - condition always runs)
			{Code: "class A extends B {\n  constructor() {\n    do {} while(super());\n    this.a = 0;\n  }\n}"},
			// for-of with super in iterable: for (const x of [super()]) {} this after (valid)
			{Code: "class A extends B {\n  constructor() {\n    for (const x of [super()]) {}\n    this.a = 0;\n  }\n}"},
			// while(true) { super(); break; } this after (valid)
			{Code: "class A extends B {\n  constructor() {\n    while(true) { super(); break; }\n    this.a = 0;\n  }\n}"},
			// super() * this.a + this.b (valid - super on far left of chain)
			{Code: "class A extends B {\n  constructor() {\n    const x = super() * this.a + this.b;\n  }\n}"},
			// for-in with super in expression
			{Code: "class A extends B {\n  constructor() {\n    for (const k in super()) {}\n    this.a = 0;\n  }\n}"},
			// compound assignment with super: x += super()
			{Code: "class A extends B {\n  constructor() {\n    let x = 0;\n    x += super();\n    this.a = 0;\n  }\n}"},
			// multiple var decls: super in 2nd, this in 3rd
			{Code: "class A extends B {\n  constructor() {\n    let a = 1, b = super(), c = this.x;\n  }\n}"},
			// +super() then this (unary plus)
			{Code: "class A extends B {\n  constructor() {\n    +super();\n    this.a = 0;\n  }\n}"},
			// !super() then this (unary not)
			{Code: "class A extends B {\n  constructor() {\n    !super();\n    this.a = 0;\n  }\n}"},
			// void super() then this
			{Code: "class A extends B {\n  constructor() {\n    void super();\n    this.a = 0;\n  }\n}"},
			// typeof super() then this
			{Code: "class A extends B {\n  constructor() {\n    typeof super();\n    this.a = 0;\n  }\n}"},
			// ~super() then this
			{Code: "class A extends B {\n  constructor() {\n    ~super();\n    this.a = 0;\n  }\n}"},
			// -super() then this
			{Code: "class A extends B {\n  constructor() {\n    -super();\n    this.a = 0;\n  }\n}"},
			// super() in array element with this after
			{Code: "class A extends B {\n  constructor() {\n    const arr = [super(), this.a];\n  }\n}"},
			// Object.assign with super()
			{Code: "class A extends B {\n  constructor() {\n    Object.assign(super(), { a: 1 });\n    this.x = 0;\n  }\n}"},
			// Immediately invoked class expression (scope boundary)
			{Code: "class A extends B {\n  constructor() {\n    new (class { constructor() { } })();\n    super();\n    this.a = 0;\n  }\n}"},
			// super() in nested function call a(b(c(super())))
			{Code: "class A extends B {\n  constructor() {\n    a(b(c(super())));\n    this.x = 0;\n  }\n}"},
			// switch with non-empty case fallthrough to super
			{Code: "class A extends B {\n  constructor(x) {\n    switch(x) {\n      case 1: foo();\n      case 2: super(); break;\n      default: super(); break;\n    }\n    this.a = 0;\n  }\n}"},
			// super()?.foo?.bar then this
			{Code: "class A extends B {\n  constructor() {\n    super()?.foo?.bar;\n    this.a = 0;\n  }\n}"},
			// super() with rest parameter
			{Code: "class A extends B {\n  constructor(...args) {\n    super(...args);\n    this.a = 0;\n  }\n}"},
			// super()[key] then this
			{Code: "class A extends B {\n  constructor(key) {\n    super()[key];\n    this.a = 0;\n  }\n}"},
			// super() in destructuring assignment RHS
			{Code: "class A extends B {\n  constructor() {\n    let a;\n    ({ a = 1 } = super());\n    this.x = 0;\n  }\n}"},
			// super() in template literal with this after: `${super()} ${this.a}`
			{Code: "class A extends B {\n  constructor() {\n    const x = `${super()} ${this.a}`;\n  }\n}"},
			// Arrow function as default parameter (scope boundary)
			{Code: "class A extends B {\n  constructor(fn = () => this) {\n    super();\n  }\n}"},
			// Function expression as default parameter (scope boundary)
			{Code: "class A extends B {\n  constructor(fn = function() { return this; }) {\n    super();\n  }\n}"},
			// Nested arrow in object method (scope boundaries)
			{Code: "class A extends B {\n  constructor() {\n    const obj = { method() { return () => this.a; } };\n    super();\n  }\n}"},
			// Destructuring from super() return value
			{Code: "class A extends B {\n  constructor() {\n    const { a, b } = super();\n    this.x = 0;\n  }\n}"},
			// Class extends ternary
			{Code: "class A extends (true ? Object : Object) {\n  constructor() {\n    super();\n    this.a = 0;\n  }\n}"},
			// Empty try/catch/finally with super after
			{Code: "class A extends B {\n  constructor() {\n    try {} catch(e) {} finally {}\n    super();\n    this.a = 0;\n  }\n}"},
			// super() || super() then this
			{Code: "class A extends B {\n  constructor() {\n    super() || super();\n    this.a = 0;\n  }\n}"},
			// super() in catch also super() in finally
			{Code: "class A extends B {\n  constructor() {\n    try { } catch(e) { super(); } finally { super(); }\n    this.a = 0;\n  }\n}"},
			// for(;;) { super(); break; } this after (infinite loop always enters body)
			{Code: "class A extends B {\n  constructor() {\n    for(;;) { super(); break; }\n    this.a = 0;\n  }\n}"},
			// super before if block, this in block body
			{Code: "class A extends B {\n  constructor(cond) {\n    super();\n    if (cond) {\n      this.a = 0;\n    }\n  }\n}"},
			// super before if/else, this in both branches
			{Code: "class A extends B {\n  constructor(cond) {\n    super();\n    if (cond) {\n      this.a = 0;\n    } else {\n      this.b = 0;\n    }\n  }\n}"},
			// super before switch, this in case body
			{Code: "class A extends B {\n  constructor(x) {\n    super();\n    switch(x) {\n      case 1: this.a = 0; break;\n      default: this.b = 0; break;\n    }\n  }\n}"},
			// super before try, this in try/catch/finally
			{Code: "class A extends B {\n  constructor() {\n    super();\n    try { this.a = 0; } catch(e) { this.b = 0; } finally { this.c = 0; }\n  }\n}"},
			// super() in if condition, this in then/else
			{Code: "class A extends B {\n  constructor() {\n    if (super()) {\n      this.a = 0;\n    } else {\n      this.b = 0;\n    }\n  }\n}"},
			// super() in if condition, this in nested block inside then
			{Code: "class A extends B {\n  constructor() {\n    if (super()) {\n      { this.a = 0; }\n    }\n  }\n}"},
			// for(super();;) body 中使用 this
			{Code: "class A extends B {\n  constructor() {\n    for(super();;) { this.a = 0; break; }\n  }\n}"},
			// for(let x = super();;) body 中使用 this
			{Code: "class A extends B {\n  constructor() {\n    for(let x = super();;) { this.a = 0; break; }\n  }\n}"},
			// super() in bare block
			{Code: "class A extends B {\n  constructor() {\n    super();\n    {\n      this.a = 0;\n    }\n  }\n}"},
			// nested bare blocks after super
			{Code: "class A extends B {\n  constructor() {\n    super();\n    { { this.a = 0; } }\n  }\n}"},
			// switch(super()) case body with this
			{Code: "class A extends B {\n  constructor() {\n    switch(super()) {\n      case 1: this.a = 0; break;\n      default: this.b = 0; break;\n    }\n  }\n}"},
			// switch(super()) then this after
			{Code: "class A extends B {\n  constructor() {\n    switch(super()) {}\n    this.a = 0;\n  }\n}"},
			// return super() + this.a (super evaluated first)
			{Code: "class A extends B {\n  constructor() {\n    return super() + this.a;\n  }\n}"},
			// return (super(), this.a) (comma: super first)
			{Code: "class A extends B {\n  constructor() {\n    return (super(), this.a);\n  }\n}"},
			// throw (super(), new Error(this.a))
			{Code: "class A extends B {\n  constructor() {\n    throw (super(), new Error(this.a));\n  }\n}"},
			// for body: super then this (sequential in body)
			{Code: "class A extends B {\n  constructor(cond) {\n    for (; cond; ) { super(); this.a = 0; break; }\n  }\n}"},
			// for-in body: super then this
			{Code: "class A extends B {\n  constructor() {\n    for (const k in {a:1}) { super(); this.a = 0; }\n  }\n}"},
			// for-of body: super then this
			{Code: "class A extends B {\n  constructor() {\n    for (const x of [1]) { super(); this.a = 0; }\n  }\n}"},
			// while body: super then this
			{Code: "class A extends B {\n  constructor(cond) {\n    while(cond) { super(); this.a = 0; break; }\n  }\n}"},
			// return super().toString()
			{Code: "class A extends B {\n  constructor() {\n    return super().toString();\n  }\n}"},
			// return [super(), this.a] (array in return)
			{Code: "class A extends B {\n  constructor() {\n    return [super(), this.a];\n  }\n}"},
			// for body: if/else super then this
			{Code: "class A extends B {\n  constructor(cond) {\n    for (; cond; ) {\n      if (true) { super(); } else { super(); }\n      this.a = 0;\n      break;\n    }\n  }\n}"},
			// super() in while condition, this in body (condition always evaluated)
			{Code: "class A extends B {\n  constructor() {\n    while(super()) { this.a = 0; break; }\n  }\n}"},
			// switch with default only, super in default
			{Code: "class A extends B {\n  constructor(x) {\n    switch(x) {\n      default: super(); break;\n    }\n    this.a = 0;\n  }\n}"},
			// try { throw 1; } catch(e) { super(); } this after
			{Code: "class A extends B {\n  constructor() {\n    try { throw 1; } catch(e) { super(); }\n    this.a = 0;\n  }\n}"},
			// nested class then super then this (outer scope)
			{Code: "class A extends B {\n  constructor() {\n    class Inner extends Object { constructor() { super(); this.x = 0; } }\n    super();\n    this.a = 0;\n  }\n}"},
			// chained assignment a = b = c = super(); this after
			{Code: "class A extends B {\n  constructor() {\n    let a, b, c;\n    a = b = c = super();\n    this.x = 0;\n  }\n}"},
			// super() in ternary condition, this in both branches
			{Code: "class A extends B {\n  constructor() {\n    (super() ? this.a : this.b);\n  }\n}"},
			// super() ?? this.fallback (left always evaluated, right conditional)
			{Code: "class A extends B {\n  constructor() {\n    super() ?? this.fallback;\n  }\n}"},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// this before super (no super at all)
			{
				Code: `class A extends B { constructor() { this.c = 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 1, Column: 37},
				},
			},
			// this before super (super comes later)
			{
				Code: `class A extends B { constructor() { this.c = 0; super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 1, Column: 37},
				},
			},
			// super.c() before super()
			{
				Code: `class A extends B { constructor() { super.c(); super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 1, Column: 37},
				},
			},
			// this in super() arguments
			{
				Code: `class A extends B { constructor() { super(this.c); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 1, Column: 43},
				},
			},
			// super() only in if (no else), this after
			{
				Code: `class A extends B { constructor() { if (a) { super(); } this.c = 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 1, Column: 57},
				},
			},
			// this in if condition, super in body
			{
				Code: `class A extends B { constructor() { if (this.a) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 1, Column: 41},
				},
			},
			// this method call without super
			{
				Code: `class A extends B { constructor() { this.c(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 1, Column: 37},
				},
			},
			// this in super arguments (method call)
			{
				Code: `class A extends B { constructor() { super(this.c()); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 1, Column: 43},
				},
			},
			// comma: this before super
			{
				Code: "class A extends B {\n  constructor() {\n    this.a = 0, super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
				},
			},
			// logical && super(), this after (conditional super)
			{
				Code: "class A extends B {\n  constructor(a) {\n    a && super();\n    this.a = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// this in default parameter
			{
				Code: "class A extends B {\n  constructor(a = this.x) {\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 2, Column: 19},
				},
			},
			// this in destructuring default before super
			{
				Code: "class A extends B {\n  constructor() {\n    const { a = this.b } = {};\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 17},
				},
			},
			// this in template literal before super
			{
				Code: "class A extends B {\n  constructor() {\n    const x = `${this.a}`;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 18},
				},
			},
			// this with optional chaining before super
			{
				Code: "class A extends B {\n  constructor() {\n    this?.a;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
				},
			},
			// this in for-of before super
			{
				Code: "class A extends B {\n  constructor() {\n    for (const x of this.items) {}\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 21},
				},
			},
			// super.prop in default parameter
			{
				Code: "class A extends B {\n  constructor(a = super.x) {\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 2, Column: 19},
				},
			},
			// logical assignment: this.a &&= 0
			{
				Code: "class A extends B {\n  constructor() {\n    this.a &&= 0;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
				},
			},
			// logical assignment: this.a ||= 0
			{
				Code: "class A extends B {\n  constructor() {\n    this.a ||= 0;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
				},
			},
			// logical assignment: this.a ??= 0
			{
				Code: "class A extends B {\n  constructor() {\n    this.a ??= 0;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
				},
			},
			// nested destructuring default with this
			{
				Code: "class A extends B {\n  constructor() {\n    const { a: { b = this.c } } = { a: {} };\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 22},
				},
			},
			// array destructuring default with this
			{
				Code: "class A extends B {\n  constructor() {\n    const [a = this.b] = [];\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 16},
				},
			},
			// this in for-loop condition
			{
				Code: "class A extends B {\n  constructor() {\n    for (;this.check();) { break; }\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 11},
				},
			},
			// new this.Foo() before super
			{
				Code: "class A extends B {\n  constructor() {\n    new this.Foo();\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 9},
				},
			},
			// delete this.foo before super
			{
				Code: "class A extends B {\n  constructor() {\n    delete this.foo;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 12},
				},
			},
			// typeof this before super
			{
				Code: "class A extends B {\n  constructor() {\n    const t = typeof this;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 22},
				},
			},
			// spread this before super
			{
				Code: "class A extends B {\n  constructor() {\n    const arr = [...this.items];\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 21},
				},
			},
			// tagged template with this before super
			{
				Code: "class A extends B {\n  constructor() {\n    this.tag`hello`;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
				},
			},
			// void this before super
			{
				Code: "class A extends B {\n  constructor() {\n    void this.a;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 10},
				},
			},
			// ternary with super only in one branch, this after
			{
				Code: "class A extends B {\n  constructor(a) {\n    a ? super() : null;\n    this.x = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// if-else if without final else (conditional super)
			{
				Code: "class A extends B {\n  constructor(a) {\n    if (a === 1) {\n      super();\n    } else if (a === 2) {\n      super();\n    }\n    this.x = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 8, Column: 5},
				},
			},
			// super in conditional for-loop, this after
			{
				Code: "class A extends B {\n  constructor(a) {\n    for (;a;) { super(); break; }\n    this.x = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// break before super in loop, this after
			{
				Code: "class A extends B {\n  constructor() {\n    for (let i = 0; i < 3; i++) {\n      if (i === 1) break;\n      super();\n    }\n    this.x = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 7, Column: 5},
				},
			},
			// empty switch then this (no super)
			{
				Code: "class A extends B {\n  constructor(x) {\n    switch(x) {}\n    this.a = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// for-of destructuring with this in iterable
			{
				Code: "class A extends B {\n  constructor() {\n    for (const {a} of this.items) {}\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 23},
				},
			},
			// this on left of logical with super on right, this after
			{
				Code: "class A extends B {\n  constructor() {\n    this.a || super();\n    this.x = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// this.x = super() (this target evaluated before super value)
			{
				Code: "class A extends B {\n  constructor() {\n    this.x = super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
				},
			},
			// this in computed property of object literal
			{
				Code: "class A extends B {\n  constructor() {\n    const obj = { [this.key]: 1 };\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 20},
				},
			},
			// nested ternary with one leaf missing super
			{
				Code: "class A extends B {\n  constructor(a, b) {\n    a ? (b ? super() : null) : super();\n    this.c = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// inner if without else → conditional super
			{
				Code: "class A extends B {\n  constructor(a, b) {\n    if (a) {\n      if (b) {\n        super();\n      }\n    } else {\n      super();\n    }\n    this.x = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 10, Column: 5},
				},
			},
			// this in instanceof before super
			{
				Code: "class A extends B {\n  constructor() {\n    const x = this instanceof Object;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 15},
				},
			},
			// while with this in condition before super
			{
				Code: "class A extends B {\n  constructor() {\n    while(this.check()) { break; }\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 11},
				},
			},
			// for-in with this.obj
			{
				Code: "class A extends B {\n  constructor() {\n    for (const k in this.obj) {}\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 21},
				},
			},
			// this in switch discriminant
			{
				Code: "class A extends B {\n  constructor() {\n    switch(this.type) {\n      case 'a': super(); break;\n    }\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 12},
				},
			},
			// two this usages before super (both reported)
			{
				Code: "class A extends B {\n  constructor() {\n    this.a = 0;\n    this.b = 1;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// this[key] before super
			{
				Code: "class A extends B {\n  constructor() {\n    this['foo'];\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
				},
			},
			// this in for-loop initializer
			{
				Code: "class A extends B {\n  constructor() {\n    for (this.i = 0;;) { break; }\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 10},
				},
			},
			// super in catch only (conditional)
			{
				Code: "class A extends B {\n  constructor() {\n    try { } catch(e) { super(); }\n    this.a = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// this.a + 1 + super() (invalid - this before super in chain)
			{
				Code: "class A extends B {\n  constructor() {\n    const x = this.a + 1 + super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 15},
				},
			},
			// (this.a + this.b) * super() (both this before super)
			{
				Code: "class A extends B {\n  constructor() {\n    const x = (this.a + this.b) * super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 16},
					{MessageId: "noBeforeSuper", Line: 3, Column: 25},
				},
			},
			// for-of body with super is conditional, this after
			{
				Code: "class A extends B {\n  constructor() {\n    for (const x of [1,2,3]) { super(); }\n    this.a = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// do {} while(this.a) before super (invalid)
			{
				Code: "class A extends B {\n  constructor() {\n    do { } while(this.a);\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 18},
				},
			},
			// this in ternary condition before super branches (invalid)
			{
				Code: "class A extends B {\n  constructor() {\n    this.a ? super() : super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
				},
			},
			// this then super in array (invalid)
			{
				Code: "class A extends B {\n  constructor() {\n    const arr = [this.a, super()];\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 18},
				},
			},
			// try { super(); } finally { this.a = 0; } — try might throw before super
			{
				Code: "class A extends B {\n  constructor() {\n    try { super(); } finally { this.a = 0; }\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 32},
				},
			},
			// this in try, super in finally
			{
				Code: "class A extends B {\n  constructor() {\n    try { this.a = 0; } finally { super(); }\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 11},
				},
			},
			// nested for loops with super, this after (conditional)
			{
				Code: "class A extends B {\n  constructor() {\n    for (let i = 0; i < 1; i++) {\n      for (let j = 0; j < 1; j++) {\n        super();\n      }\n    }\n    this.a = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 8, Column: 5},
				},
			},
			// null || super(); this after (conditional — left side may be truthy)
			{
				Code: "class A extends B {\n  constructor() {\n    null || super();\n    this.a = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// Multiple conditional super() calls in separate ifs, this after
			{
				Code: "class A extends B {\n  constructor(a) {\n    if (a) { super(); }\n    if (!a) { super(); }\n    this.x = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 5, Column: 5},
				},
			},
			// this before super in template literal
			{
				Code: "class A extends B {\n  constructor() {\n    const x = `${this.a} ${super()}`;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 18},
				},
			},
			// while(true) { this; break; } super()
			{
				Code: "class A extends B {\n  constructor() {\n    while(true) { this.a = 0; break; }\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 19},
				},
			},
			// do { this } while(super())
			{
				Code: "class A extends B {\n  constructor() {\n    do { this.a = 0; } while(super());\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 10},
				},
			},
			// ternary with super in one branch only, this in expression after
			{
				Code: "class A extends B {\n  constructor(c) {\n    const x = (c ? super() : 0) + this.a;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 35},
				},
			},
			// comma: this before super in for-loop initializer
			{
				Code: "class A extends B {\n  constructor() {\n    let i;\n    for (i = 0, this.a = 1, super();;) { break; }\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 17},
				},
			},
			// nested assignment: this.a = this.b = super()
			{
				Code: "class A extends B {\n  constructor() {\n    this.a = this.b = super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 5},
					{MessageId: "noBeforeSuper", Line: 3, Column: 14},
				},
			},
			// switch no default, all cases have super (conditional)
			{
				Code: "class A extends B {\n  constructor(x) {\n    switch(x) {\n      case 1: super(); break;\n      case 2: super(); break;\n    }\n    this.a = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 7, Column: 5},
				},
			},
			// for-in body with super (conditional)
			{
				Code: "class A extends B {\n  constructor() {\n    for (const k in {a:1}) { super(); }\n    this.a = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// try { super(); } catch(e) { } this after (conditional - catch may skip super)
			{
				Code: "class A extends B {\n  constructor() {\n    try { super(); } catch(e) { }\n    this.a = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// (a || b) && super(); this after (conditional)
			{
				Code: "class A extends B {\n  constructor(a, b) {\n    (a || b) && super();\n    this.x = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// super() in catch, this in finally (no super in try)
			{
				Code: "class A extends B {\n  constructor() {\n    try { } catch(e) { super(); } finally { this.a = 0; }\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 45},
				},
			},
			// chained ternary missing super in one leaf
			{
				Code: "class A extends B {\n  constructor(a, b, c) {\n    a ? (b ? super() : super()) : (c ? super() : 0);\n    this.x = 0;\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 4, Column: 5},
				},
			},
			// super() in try, this in catch (no finally)
			{
				Code: "class A extends B {\n  constructor() {\n    try { super(); } catch(e) { this.a = 0; }\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 33},
				},
			},
			// this as argument to super()
			{
				Code: "class A extends B {\n  constructor() {\n    super(this);\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 11},
				},
			},
			// super.method() as argument to super()
			{
				Code: "class A extends B {\n  constructor() {\n    super(super.method());\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 11},
				},
			},
			// { ...this } before super
			{
				Code: "class A extends B {\n  constructor() {\n    const obj = { ...this };\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 22},
				},
			},
			// delete super.prop before super()
			{
				Code: "class A extends B {\n  constructor() {\n    delete super.prop;\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 12},
				},
			},
			// destructuring with computed key using this
			{
				Code: "class A extends B {\n  constructor() {\n    const { [this.key]: val } = {};\n    super();\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 14},
				},
			},
			// for-loop incrementor super, this in body (body runs before incrementor)
			{
				Code: "class A extends B {\n  constructor(cond) {\n    for (; cond; super()) { this.a = 0; break; }\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noBeforeSuper", Line: 3, Column: 29},
				},
			},
		},
	)
}
