import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('init-declarations', {
  valid: [
    'var foo = null;',
    'foo = true;',
    'var foo = 1, bar = false, baz = {};',
    'function foo() { var foo = 0; var bar = []; }',
    'var fn = function() {};',
    'var foo = bar = 2;',
    'for (var i = 0; i < 1; i++) {}',
    'for (var foo in []) {}',
    'for (var foo of []) {}',
    { code: 'let a = true;', options: ['always'] as any },
    { code: 'const a = {};', options: ['always'] as any },
    { code: 'using a = foo();', options: ['always'] as any },
    { code: 'await using a = foo();', options: ['always'] as any },
    {
      code: 'function foo() { let a = 1, b = false; if (a) { let c = 3, d = null; } }',
      options: ['always'] as any,
    },
    {
      code: 'function foo() { const a = 1, b = true; if (a) { const c = 3, d = null; } }',
      options: ['always'] as any,
    },
    {
      code: 'function foo() { let a = 1; const b = false; var c = true; }',
      options: ['always'] as any,
    },
    { code: 'var foo;', options: ['never'] as any },
    { code: 'var foo, bar, baz;', options: ['never'] as any },
    { code: 'function foo() { var foo; var bar; }', options: ['never'] as any },
    { code: 'let a;', options: ['never'] as any },
    { code: 'const a = 1;', options: ['never'] as any },
    { code: 'using a = foo();', options: ['never'] as any },
    { code: 'await using a = foo();', options: ['never'] as any },
    {
      code: 'function foo() { let a, b; if (a) { let c, d; } }',
      options: ['never'] as any,
    },
    {
      code: 'function foo() { const a = 1, b = true; if (a) { const c = 3, d = null; } }',
      options: ['never'] as any,
    },
    {
      code: 'function foo() { let a; const b = false; var c; }',
      options: ['never'] as any,
    },
    {
      code: 'for(var i = 0; i < 1; i++){}',
      options: ['never', { ignoreForLoopInit: true }] as any,
    },
    {
      code: 'for (var foo in []) {}',
      options: ['never', { ignoreForLoopInit: true }] as any,
    },
    {
      code: 'for (var foo of []) {}',
      options: ['never', { ignoreForLoopInit: true }] as any,
    },

    // ---- ruleTesterTypeScript (@typescript-eslint/parser) suite ----
    { code: 'declare const foo: number;', options: ['always'] as any },
    { code: 'declare const foo: number;', options: ['never'] as any },
    {
      code: `
	  declare namespace myLib {
		let numberOfGreetings: number;
	  }
			`,
      options: ['always'] as any,
    },
    {
      code: `
	  declare namespace myLib {
		let numberOfGreetings: number;
	  }
			`,
      options: ['never'] as any,
    },
    {
      code: `
	  declare namespace myLib {
		let valueInside: number;
	  }
		let valueOutside: number;
			`,
      options: ['never'] as any,
    },
    `
	  interface GreetingSettings {
		greeting: string;
		duration?: number;
		color?: string;
	  }
			`,
    {
      code: `
	  interface GreetingSettings {
		greeting: string;
		duration?: number;
		color?: string;
	  }
			`,
      options: ['never'] as any,
    },
    'type GreetingLike = string | (() => string) | Greeter;',
    {
      code: 'type GreetingLike = string | (() => string) | Greeter;',
      options: ['never'] as any,
    },
    {
      code: `
	  function foo() {
		var bar: string;
	  }
			`,
      options: ['never'] as any,
    },
    { code: 'var bar: string;', options: ['never'] as any },
    {
      code: `
	  var bar: string = function (): string {
		return 'string';
	  };
			`,
      options: ['always'] as any,
    },
    {
      code: `
	  var bar: string = function (arg1: string): string {
		return 'string';
	  };
			`,
      options: ['always'] as any,
    },
    {
      code: "function foo(arg1: string = 'string'): void {}",
      options: ['never'] as any,
    },
    { code: "const foo: string = 'hello';", options: ['never'] as any },
    `
	  const class1 = class NAME {
		constructor() {
		  var name1: string = 'hello';
		}
	  };
			`,
    `
	  const class1 = class NAME {
		static pi: number = 3.14;
	  };
			`,
    {
      code: `
	  const class1 = class NAME {
		static pi: number = 3.14;
	  };
			`,
      options: ['never'] as any,
    },
    `
	  interface IEmployee {
		empCode: number;
		empName: string;
		getSalary: (number) => number; // arrow function
		getManagerName(number): string;
	  }
			`,
    {
      code: `
	  interface IEmployee {
		empCode: number;
		empName: string;
		getSalary: (number) => number; // arrow function
		getManagerName(number): string;
	  }
			`,
      options: ['never'] as any,
    },
    { code: "const foo: number = 'asd';", options: ['always'] as any },
    { code: 'const foo: number;', options: ['never'] as any },
    {
      code: `
	  namespace myLib {
		let numberOfGreetings: number;
	  }
			`,
      options: ['never'] as any,
    },
    {
      code: `
	  namespace myLib {
		let numberOfGreetings: number = 2;
	  }
			`,
      options: ['always'] as any,
    },
    {
      code: `
	  declare namespace myLib1 {
		const foo: number;
		namespace myLib2 {
		  let bar: string;
		  namespace myLib3 {
			let baz: object;
		  }
		}
	  }
			`,
      options: ['always'] as any,
    },
    {
      code: `
	  declare namespace myLib1 {
		const foo: number;
		namespace myLib2 {
		  let bar: string;
		  namespace myLib3 {
			let baz: object;
		  }
		}
	  }
			`,
      options: ['never'] as any,
    },
  ],
  invalid: [
    {
      code: 'var foo;',
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }],
    },
    {
      code: 'for (var a in []) var foo;',
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }],
    },
    {
      code: 'var foo, bar = false, baz;',
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }, { messageId: 'initialized' }],
    },
    {
      code: 'function foo() { var foo = 0; var bar; }',
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }],
    },
    {
      code: 'function foo() { var foo; var bar = foo; }',
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }],
    },
    {
      code: 'let a;',
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }],
    },
    {
      code: 'function foo() { let a = 1, b; if (a) { let c = 3, d = null; } }',
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }],
    },
    {
      code: 'function foo() { let a; const b = false; var c; }',
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }, { messageId: 'initialized' }],
    },
    {
      code: 'var foo = bar = 2;',
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: 'var foo = true;',
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: 'var foo, bar = 5, baz = 3;',
      options: ['never'] as any,
      errors: [
        { messageId: 'notInitialized' },
        { messageId: 'notInitialized' },
      ],
    },
    {
      code: 'function foo() { var foo; var bar = foo; }',
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: 'let a = 1;',
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: "function foo() { let a = 'foo', b; if (a) { let c, d; } }",
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: 'function foo() { let a; const b = false; var c = 1; }',
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: 'for(var i = 0; i < 1; i++){}',
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: 'for (var foo in []) {}',
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: 'for (var foo of []) {}',
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },

    // ---- ruleTesterTypeScript (@typescript-eslint/parser) suite ----
    {
      code: "let arr: string[] = ['arr', 'ar'];",
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: 'let arr: string = function () {};',
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: `
	  const class1 = class NAME {
		constructor() {
		  var name1: string = 'hello';
		}
	  };
			`,
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: 'let arr: string;',
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }],
    },
    {
      code: `
	  namespace myLib {
		let numberOfGreetings: number;
	  }
			`,
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }],
    },
    {
      code: `
	  namespace myLib {
		let numberOfGreetings: number = 2;
	  }
			`,
      options: ['never'] as any,
      errors: [{ messageId: 'notInitialized' }],
    },
    {
      code: `
		namespace myLib1 {
		  const foo: number;
			namespace myLib2 {
			  let bar: string;
			  namespace myLib3 {
				let baz: object;
			  }
		  }
		}
			`,
      options: ['always'] as any,
      errors: [
        { messageId: 'initialized' },
        { messageId: 'initialized' },
        { messageId: 'initialized' },
      ],
    },
    {
      code: `
	  declare namespace myLib {
		let valueInside: number;
	  }
		let valueOutside: number;
			`,
      options: ['always'] as any,
      errors: [{ messageId: 'initialized' }],
    },
  ],
});
