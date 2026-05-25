import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-this-before-super', {
  valid: [
    'class A { constructor() { this.b = 0; } }',
    'class A extends null { constructor() { } }',
    'class A extends B { constructor() { super(); this.c = 0; } }',
    'class A extends B { constructor() { super(); super.c(); } }',
    'class A extends B { constructor() { super(); this.c(); } }',
    'class A { b() { this.c = 0; } }',
    'class A extends B { c() { this.d = 0; } }',
    // do-while body always executes
    `class A extends B {
  constructor() {
    do { super(); } while (false);
    this.c = 0;
  }
}`,
    // Parenthesized ternary with super() in both branches
    `class A extends B {
  constructor(cond) {
    (cond ? super() : super());
    this.a = 0;
  }
}`,
    // this in object getter (scope boundary)
    `class A extends B {
  constructor() {
    const obj = { get foo() { return this.bar; } };
    super();
  }
}`,
    // this in object setter (scope boundary)
    `class A extends B {
  constructor() {
    const obj = { set foo(v) { this.bar = v; } };
    super();
  }
}`,
    // labeled statement with super
    `class A extends B {
  constructor() {
    label: super();
    this.a = 0;
  }
}`,
    // comma: super() first, then this
    `class A extends B {
  constructor() {
    super(), this.a = 0;
  }
}`,
    // for-loop initializer with super
    `class A extends B {
  constructor() {
    for (super();;) { break; }
    this.a = 0;
  }
}`,
    // variable declaration with super() initializer
    `class A extends B {
  constructor() {
    let x = super();
    this.a = 0;
  }
}`,
    // this in object method (scope boundary)
    `class A extends B {
  constructor() {
    const obj = { method() { this.x = 1; } };
    super();
  }
}`,
    // super() in finally block
    `class A extends B {
  constructor() {
    try { } catch(e) { } finally { super(); }
    this.a = 0;
  }
}`,
    // nested ternary with super in all leaves
    `class A extends B {
  constructor(a, b) {
    a ? (b ? super() : super()) : super();
    this.c = 0;
  }
}`,
    // this in async function (scope boundary)
    `class A extends B {
  constructor() {
    const fn = async function() { this.a = 0; };
    super();
  }
}`,
    // this in generator (scope boundary)
    `class A extends B {
  constructor() {
    const fn = function*() { this.a = 0; };
    super();
  }
}`,
    // arrow in expression (scope boundary)
    `class A extends B {
  constructor() {
    [1,2].forEach(x => { this.a = x; });
    super();
  }
}`,
    // return before this (no this reached)
    `class A extends B {
  constructor(a) {
    if (a) return;
    super();
    this.a = 0;
  }
}`,
    // throw before this (unreachable)
    `class A extends B {
  constructor() {
    throw new Error();
    this.a = 0;
  }
}`,
    // super() deep in assignment chain
    `class A extends B {
  constructor() {
    let a, b;
    a = b = super();
    this.x = 0;
  }
}`,
    // super in first var decl, this in second
    `class A extends B {
  constructor() {
    let a = super(), b = this.x;
  }
}`,
    // chained method on super() return value
    `class A extends B {
  constructor() {
    super().toString();
    this.a = 0;
  }
}`,
    // super() ?? fallback, this after
    `class A extends B {
  constructor() {
    super() ?? null;
    this.x = 0;
  }
}`,
    // this in async arrow (scope boundary)
    `class A extends B {
  constructor() {
    const fn = async () => { await this.a; };
    super();
  }
}`,
    // if-else if-else all with super (valid)
    `class A extends B {
  constructor(a) {
    if (a === 1) {
      super();
    } else if (a === 2) {
      super();
    } else {
      super();
    }
    this.x = 0;
  }
}`,
    // nested try with super in inner finally
    `class A extends B {
  constructor() {
    try {
      try { } finally { super(); }
    } catch(e) { super(); }
    this.a = 0;
  }
}`,
    // for-loop incrementor unreachable due to unconditional break
    `class A extends B {
  constructor() {
    for (let i = 0; i < 3; this.inc()) { break; }
    super();
  }
}`,
    // super() || fallback, this after (valid)
    `class A extends B {
  constructor() {
    super() || null;
    this.x = 0;
  }
}`,
    // super() && something, this after (valid)
    `class A extends B {
  constructor() {
    super() && this.init();
    this.x = 0;
  }
}`,
    // super() in object literal property value
    `class A extends B {
  constructor() {
    const obj = { key: super() };
    this.x = 0;
  }
}`,
    // super() as function argument
    `class A extends B {
  constructor() {
    foo(super());
    this.x = 0;
  }
}`,
    // super() in template literal
    `class A extends B {
  constructor() {
    const x = \`\${super()}\`;
    this.x = 0;
  }
}`,
    // super() in new expression argument
    `class A extends B {
  constructor() {
    new Foo(super());
    this.x = 0;
  }
}`,
    // switch with case fallthrough to super
    `class A extends B {
  constructor(x) {
    switch(x) {
      case 1:
      case 2: super(); break;
      default: super(); break;
    }
    this.a = 0;
  }
}`,
    // while condition with super
    `class A extends B {
  constructor() {
    while(super()) { break; }
    this.a = 0;
  }
}`,
    // try/catch both have super
    `class A extends B {
  constructor() {
    try { super(); } catch(e) { super(); }
    this.a = 0;
  }
}`,
    // [a] = [super()]; this after (valid)
    `class A extends B {
  constructor() {
    let a;
    [a] = [super()];
    this.x = 0;
  }
}`,
    // deeply nested ternary all leaves super (valid)
    `class A extends B {
  constructor(a, b, c) {
    a ? (b ? (c ? super() : super()) : super()) : super();
    this.x = 0;
  }
}`,
    // super() + this.a (valid - super on left of binary op)
    `class A extends B {
  constructor() {
    const x = super() + this.a;
  }
}`,
    // super() in ternary condition (valid)
    `class A extends B {
  constructor() {
    super() ? this.a : this.b;
  }
}`,
    // do {} while(super()); this after (valid)
    `class A extends B {
  constructor() {
    do {} while(super());
    this.a = 0;
  }
}`,
    // for-of with super in iterable (valid)
    `class A extends B {
  constructor() {
    for (const x of [super()]) {}
    this.a = 0;
  }
}`,
    // while(true) { super(); break; } this after (valid)
    `class A extends B {
  constructor() {
    while(true) { super(); break; }
    this.a = 0;
  }
}`,
    // +super() then this (unary plus)
    `class A extends B {
  constructor() {
    +super();
    this.a = 0;
  }
}`,
    // void super() then this
    `class A extends B {
  constructor() {
    void super();
    this.a = 0;
  }
}`,
    // typeof super() then this
    `class A extends B {
  constructor() {
    typeof super();
    this.a = 0;
  }
}`,
    // super() in array element with this after
    `class A extends B {
  constructor() {
    const arr = [super(), this.a];
  }
}`,
    // super() in nested function call
    `class A extends B {
  constructor() {
    a(b(c(super())));
    this.x = 0;
  }
}`,
    // switch with non-empty case fallthrough to super
    `class A extends B {
  constructor(x) {
    switch(x) {
      case 1: foo();
      case 2: super(); break;
      default: super(); break;
    }
    this.a = 0;
  }
}`,
    // super()?.foo?.bar then this
    `class A extends B {
  constructor() {
    super()?.foo?.bar;
    this.a = 0;
  }
}`,
    // super() with rest parameter
    `class A extends B {
  constructor(...args) {
    super(...args);
    this.a = 0;
  }
}`,
    // super()[key] then this
    `class A extends B {
  constructor(key) {
    super()[key];
    this.a = 0;
  }
}`,
    // super() in template with this after
    `class A extends B {
  constructor() {
    const x = \`\${super()} \${this.a}\`;
  }
}`,
    // arrow function as default parameter (scope boundary)
    `class A extends B {
  constructor(fn = () => this) {
    super();
  }
}`,
    // function expression as default parameter (scope boundary)
    `class A extends B {
  constructor(fn = function() { return this; }) {
    super();
  }
}`,
    // nested arrow in object method (scope boundaries)
    `class A extends B {
  constructor() {
    const obj = { method() { return () => this.a; } };
    super();
  }
}`,
    // destructuring from super() return
    `class A extends B {
  constructor() {
    const { a, b } = super();
    this.x = 0;
  }
}`,
    // empty try/catch/finally with super after
    `class A extends B {
  constructor() {
    try {} catch(e) {} finally {}
    super();
    this.a = 0;
  }
}`,
    // super() || super() then this
    `class A extends B {
  constructor() {
    super() || super();
    this.a = 0;
  }
}`,
    // super() in catch and also in finally
    `class A extends B {
  constructor() {
    try { } catch(e) { super(); } finally { super(); }
    this.a = 0;
  }
}`,
    // for(;;) { super(); break; } this after (infinite loop always enters body)
    `class A extends B {
  constructor() {
    for(;;) { super(); break; }
    this.a = 0;
  }
}`,
    // super before if block, this in block body
    `class A extends B {
  constructor(cond) {
    super();
    if (cond) {
      this.a = 0;
    }
  }
}`,
    // super before if/else, this in both branches
    `class A extends B {
  constructor(cond) {
    super();
    if (cond) {
      this.a = 0;
    } else {
      this.b = 0;
    }
  }
}`,
    // super before switch, this in case body
    `class A extends B {
  constructor(x) {
    super();
    switch(x) {
      case 1: this.a = 0; break;
      default: this.b = 0; break;
    }
  }
}`,
    // super before try, this in try/catch/finally
    `class A extends B {
  constructor() {
    super();
    try { this.a = 0; } catch(e) { this.b = 0; } finally { this.c = 0; }
  }
}`,
    // super() in if condition, this in then/else
    `class A extends B {
  constructor() {
    if (super()) {
      this.a = 0;
    } else {
      this.b = 0;
    }
  }
}`,
    // for(super();;) body uses this
    `class A extends B {
  constructor() {
    for(super();;) { this.a = 0; break; }
  }
}`,
    // super() in bare block
    `class A extends B {
  constructor() {
    super();
    {
      this.a = 0;
    }
  }
}`,
    // super() in while condition, this in body
    `class A extends B {
  constructor() {
    while(super()) { this.a = 0; break; }
  }
}`,
    // switch with default only
    `class A extends B {
  constructor(x) {
    switch(x) {
      default: super(); break;
    }
    this.a = 0;
  }
}`,
    // try { throw } catch { super() } this after
    `class A extends B {
  constructor() {
    try { throw 1; } catch(e) { super(); }
    this.a = 0;
  }
}`,
    // nested class then super then this
    `class A extends B {
  constructor() {
    class Inner extends Object { constructor() { super(); this.x = 0; } }
    super();
    this.a = 0;
  }
}`,
    // super() ?? this.fallback
    `class A extends B {
  constructor() {
    super() ?? this.fallback;
  }
}`,
    // switch(super()) case body with this
    `class A extends B {
  constructor() {
    switch(super()) {
      case 1: this.a = 0; break;
      default: this.b = 0; break;
    }
  }
}`,
    // switch(super()) then this after
    `class A extends B {
  constructor() {
    switch(super()) {}
    this.a = 0;
  }
}`,
    // return super() + this.a (super evaluated first)
    `class A extends B {
  constructor() {
    return super() + this.a;
  }
}`,
    // return (super(), this.a) (comma: super first)
    `class A extends B {
  constructor() {
    return (super(), this.a);
  }
}`,
    // throw (super(), new Error(this.a))
    `class A extends B {
  constructor() {
    throw (super(), new Error(this.a));
  }
}`,
    // for body: super then this (sequential in body)
    `class A extends B {
  constructor(cond) {
    for (; cond; ) { super(); this.a = 0; break; }
  }
}`,
    // for-in body: super then this
    `class A extends B {
  constructor() {
    for (const k in {a:1}) { super(); this.a = 0; }
  }
}`,
    // for-of body: super then this
    `class A extends B {
  constructor() {
    for (const x of [1]) { super(); this.a = 0; }
  }
}`,
    // while body: super then this
    `class A extends B {
  constructor(cond) {
    while(cond) { super(); this.a = 0; break; }
  }
}`,
    // return super().toString()
    `class A extends B {
  constructor() {
    return super().toString();
  }
}`,
    // return [super(), this.a] (array in return)
    `class A extends B {
  constructor() {
    return [super(), this.a];
  }
}`,
    // for body: if/else super then this
    `class A extends B {
  constructor(cond) {
    for (; cond; ) {
      if (true) { super(); } else { super(); }
      this.a = 0;
      break;
    }
  }
}`,
  ],
  invalid: [
    {
      code: 'class A extends B { constructor() { this.c = 0; } }',
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    {
      code: 'class A extends B { constructor() { this.c = 0; super(); } }',
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    {
      code: 'class A extends B { constructor() { super.c(); super(); } }',
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // comma: this before super
    {
      code: `class A extends B {
  constructor() {
    this.a = 0, super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // logical && super(), this after
    {
      code: `class A extends B {
  constructor(a) {
    a && super();
    this.a = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this in default parameter
    {
      code: `class A extends B {
  constructor(a = this.x) {
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this in destructuring default
    {
      code: `class A extends B {
  constructor() {
    const { a = this.b } = {};
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this in template literal before super
    {
      code: `class A extends B {
  constructor() {
    const x = \`\${this.a}\`;
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this with optional chaining before super
    {
      code: `class A extends B {
  constructor() {
    this?.a;
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this in for-of before super
    {
      code: `class A extends B {
  constructor() {
    for (const x of this.items) {}
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // super.prop in default parameter
    {
      code: `class A extends B {
  constructor(a = super.x) {
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // logical assignment: this.a &&= 0
    {
      code: `class A extends B {
  constructor() {
    this.a &&= 0;
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // logical assignment: this.a ||= 0
    {
      code: `class A extends B {
  constructor() {
    this.a ||= 0;
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // logical assignment: this.a ??= 0
    {
      code: `class A extends B {
  constructor() {
    this.a ??= 0;
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // nested destructuring default with this
    {
      code: `class A extends B {
  constructor() {
    const { a: { b = this.c } } = { a: {} };
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // array destructuring default with this
    {
      code: `class A extends B {
  constructor() {
    const [a = this.b] = [];
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this in for-loop condition
    {
      code: `class A extends B {
  constructor() {
    for (;this.check();) { break; }
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // new this.Foo() before super
    {
      code: `class A extends B {
  constructor() {
    new this.Foo();
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // delete this.foo before super
    {
      code: `class A extends B {
  constructor() {
    delete this.foo;
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // typeof this before super
    {
      code: `class A extends B {
  constructor() {
    const t = typeof this;
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // spread this before super
    {
      code: `class A extends B {
  constructor() {
    const arr = [...this.items];
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // tagged template with this before super
    {
      code: `class A extends B {
  constructor() {
    this.tag\`hello\`;
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // void this before super
    {
      code: `class A extends B {
  constructor() {
    void this.a;
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // ternary with super only in one branch, this after
    {
      code: `class A extends B {
  constructor(a) {
    a ? super() : null;
    this.x = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // if-else if without final else (conditional super)
    {
      code: `class A extends B {
  constructor(a) {
    if (a === 1) {
      super();
    } else if (a === 2) {
      super();
    }
    this.x = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // super in conditional for-loop, this after
    {
      code: `class A extends B {
  constructor(a) {
    for (;a;) { super(); break; }
    this.x = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // empty switch then this (no super)
    {
      code: `class A extends B {
  constructor(x) {
    switch(x) {}
    this.a = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // for-of destructuring with this in iterable
    {
      code: `class A extends B {
  constructor() {
    for (const {a} of this.items) {}
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // nested ternary with one leaf missing super
    {
      code: `class A extends B {
  constructor(a, b) {
    a ? (b ? super() : null) : super();
    this.c = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this in computed property of object literal
    {
      code: `class A extends B {
  constructor() {
    const obj = { [this.key]: 1 };
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // while with this in condition before super
    {
      code: `class A extends B {
  constructor() {
    while(this.check()) { break; }
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // for-in with this.obj
    {
      code: `class A extends B {
  constructor() {
    for (const k in this.obj) {}
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this in for-loop initializer
    {
      code: `class A extends B {
  constructor() {
    for (this.i = 0;;) { break; }
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // super in catch only (conditional)
    {
      code: `class A extends B {
  constructor() {
    try { } catch(e) { super(); }
    this.a = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this.a + 1 + super() (this before super in chain)
    {
      code: `class A extends B {
  constructor() {
    const x = this.a + 1 + super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // do {} while(this.a) before super (invalid)
    {
      code: `class A extends B {
  constructor() {
    do { } while(this.a);
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this in ternary condition before super (invalid)
    {
      code: `class A extends B {
  constructor() {
    this.a ? super() : super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this then super in array
    {
      code: `class A extends B {
  constructor() {
    const arr = [this.a, super()];
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // try { super(); } finally { this } — try might throw before super
    {
      code: `class A extends B {
  constructor() {
    try { super(); } finally { this.a = 0; }
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this in try, super in finally
    {
      code: `class A extends B {
  constructor() {
    try { this.a = 0; } finally { super(); }
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // nested for loops with super, this after
    {
      code: `class A extends B {
  constructor() {
    for (let i = 0; i < 1; i++) {
      for (let j = 0; j < 1; j++) {
        super();
      }
    }
    this.a = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // null || super(); this after (conditional)
    {
      code: `class A extends B {
  constructor() {
    null || super();
    this.a = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // multiple conditional super ifs, this after
    {
      code: `class A extends B {
  constructor(a) {
    if (a) { super(); }
    if (!a) { super(); }
    this.x = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // this before super in template literal
    {
      code: `class A extends B {
  constructor() {
    const x = \`\${this.a} \${super()}\`;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // while(true) { this; break; } super()
    {
      code: `class A extends B {
  constructor() {
    while(true) { this.a = 0; break; }
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // do { this } while(super())
    {
      code: `class A extends B {
  constructor() {
    do { this.a = 0; } while(super());
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // ternary with super in one branch only, this after binary
    {
      code: `class A extends B {
  constructor(c) {
    const x = (c ? super() : 0) + this.a;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // switch no default (conditional super)
    {
      code: `class A extends B {
  constructor(x) {
    switch(x) {
      case 1: super(); break;
      case 2: super(); break;
    }
    this.a = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // try super catch empty this after
    {
      code: `class A extends B {
  constructor() {
    try { super(); } catch(e) { }
    this.a = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // (a||b) && super(); this after (conditional)
    {
      code: `class A extends B {
  constructor(a, b) {
    (a || b) && super();
    this.x = 0;
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // super in catch, this in finally (no super in try)
    {
      code: `class A extends B {
  constructor() {
    try { } catch(e) { super(); } finally { this.a = 0; }
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // super() in try, this in catch (no finally)
    {
      code: `class A extends B {
  constructor() {
    try { super(); } catch(e) { this.a = 0; }
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // { ...this } before super
    {
      code: `class A extends B {
  constructor() {
    const obj = { ...this };
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // delete super.prop before super()
    {
      code: `class A extends B {
  constructor() {
    delete super.prop;
    super();
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
    // for-loop incrementor super, this in body (body runs before incrementor)
    {
      code: `class A extends B {
  constructor(cond) {
    for (; cond; super()) { this.a = 0; break; }
  }
}`,
      errors: [{ messageId: 'noBeforeSuper' }],
    },
  ],
});
