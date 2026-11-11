// Every case from typescript-eslint v8.28.0, with messages and full ranges
// additionally pinned from the same upstream version. The receiver-union message difference is explained inline.
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.28.0/packages/eslint-plugin/tests/rules/unbound-method.test.ts
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  fixtureFiles: {
    'tsconfig.json':
      '{"compilerOptions":{"jsx":"preserve","target":"es2015","module":"commonjs","strict":true,"types":[],"lib":["es2015","es2017","esnext"],"experimentalDecorators":true},"include":["*.ts","*.tsx"]}',
    'class.ts': `// used by no-throw-literal test case to validate custom error
export class Error {}

// used by unbound-method test case to test imports
export const console = { log() {} };

// used by prefer-reduce-type-parameter to test native vs userland check
export class Reducable {
  reduce() {}
}

// used by no-implied-eval test function imports
export class Function {}
`,
  },
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
    },
  },
});

ruleTester.run('unbound-method', {
  valid: [
    {
      code: 'Promise.resolve().then(console.log);',
      filename: 'file.ts',
    },
    {
      code: "['1', '2', '3'].map(Number.parseInt);",
      filename: 'file.ts',
    },
    {
      code: '[5.2, 7.1, 3.6].map(Math.floor);',
      filename: 'file.ts',
    },
    {
      code: `
      const foo = Number;
      ['1', '2', '3'].map(foo.parseInt);
    `,
      filename: 'file.ts',
    },
    {
      code: `
      const foo = Math;
      [5.2, 7.1, 3.6].map(foo.floor);
    `,
      filename: 'file.ts',
    },
    {
      code: "['1', '2', '3'].map(Number['floor']);",
      filename: 'file.ts',
    },
    {
      code: 'const x = console.log;',
      filename: 'file.ts',
    },
    {
      code: 'const x = Object.defineProperty;',
      filename: 'file.ts',
    },
    {
      code: `
      const foo = Object;
      const x = foo.defineProperty;
    `,
      filename: 'file.ts',
    },
    {
      code: 'const x = String.fromCharCode;',
      filename: 'file.ts',
    },
    {
      code: `
      const foo = String;
      const x = foo.fromCharCode;
    `,
      filename: 'file.ts',
    },
    {
      code: 'const x = RegExp.prototype;',
      filename: 'file.ts',
    },
    {
      code: 'const x = Symbol.keyFor;',
      filename: 'file.ts',
    },
    {
      code: `
      const foo = Symbol;
      const x = foo.keyFor;
    `,
      filename: 'file.ts',
    },
    {
      code: 'const x = Array.isArray;',
      filename: 'file.ts',
    },
    {
      code: `
      const foo = Array;
      const x = foo.isArray;
    `,
      filename: 'file.ts',
    },
    {
      code: `
      class Foo extends Array {}
      const x = Foo.isArray;
    `,
      filename: 'file.ts',
    },
    {
      code: 'const x = Proxy.revocable;',
      filename: 'file.ts',
    },
    {
      code: `
      const foo = Proxy;
      const x = foo.revocable;
    `,
      filename: 'file.ts',
    },
    {
      code: 'const x = Date.parse;',
      filename: 'file.ts',
    },
    {
      code: `
      const foo = Date;
      const x = foo.parse;
    `,
      filename: 'file.ts',
    },
    {
      code: 'const x = Atomics.load;',
      filename: 'file.ts',
    },
    {
      code: `
      const foo = Atomics;
      const x = foo.load;
    `,
      filename: 'file.ts',
    },
    {
      code: 'const x = Reflect.deleteProperty;',
      filename: 'file.ts',
    },
    {
      code: 'const x = JSON.stringify;',
      filename: 'file.ts',
    },
    {
      code: `
      const foo = JSON;
      const x = foo.stringify;
    `,
      filename: 'file.ts',
    },
    {
      code: `
      const o = {
        f: function (this: void) {},
      };
      const f = o.f;
    `,
      filename: 'file.ts',
    },
    {
      code: `
      const { alert } = window;
    `,
      filename: 'file.ts',
    },
    {
      code: `
      let b = window.blur;
    `,
      filename: 'file.ts',
    },
    {
      code: `
      function foo() {}
      const fooObject = { foo };
      const { foo: bar } = fooObject;
    `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound();
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.unbound();
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.boundStatic();
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.unboundStatic();
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

const bound = instance.bound;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

const boundStatic = ContainsMethods;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

const { bound } = instance;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

const { boundStatic } = ContainsMethods;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

(instance.bound)();
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

(instance.unbound)();
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

(ContainsMethods.boundStatic)();
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

(ContainsMethods.unboundStatic)();
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound\`\`;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.unbound\`\`;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (instance.bound) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (instance.unbound) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (instance.bound !== undefined) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (instance.unbound !== undefined) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (ContainsMethods.boundStatic) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (ContainsMethods.unboundStatic) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (ContainsMethods.boundStatic !== undefined) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (ContainsMethods.unboundStatic !== undefined) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (ContainsMethods.boundStatic && instance) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (ContainsMethods.unboundStatic && instance) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (instance.bound || instance) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (instance.unbound || instance) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.unboundStatic && 0 || ContainsMethods;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

(instance.bound || instance) ? 1 : 0
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

(instance.unbound || instance) ? 1 : 0
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

while (instance.bound) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

while (instance.unbound) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

while (instance.bound !== undefined) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

while (instance.unbound !== undefined) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

while (ContainsMethods.boundStatic) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

while (ContainsMethods.unboundStatic) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

while (ContainsMethods.boundStatic !== undefined) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

while (ContainsMethods.unboundStatic !== undefined) { }
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound as any;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.boundStatic as any;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound++;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

+instance.bound;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

++instance.bound;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound--;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

-instance.bound;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

--instance.bound;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound += 1;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound -= 1;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound *= 1;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound /= 1;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound || 0;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound && 0;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.bound ? 1 : 0;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.unbound ? 1 : 0;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.boundStatic++;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

+ContainsMethods.boundStatic;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

++ContainsMethods.boundStatic;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.boundStatic--;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

-ContainsMethods.boundStatic;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

--ContainsMethods.boundStatic;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.boundStatic += 1;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.boundStatic -= 1;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.boundStatic *= 1;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.boundStatic /= 1;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.boundStatic || 0;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instane.boundStatic && 0;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.boundStatic ? 1 : 0;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.unboundStatic ? 1 : 0;
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

typeof instance.bound === 'function';
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

typeof instance.unbound === 'function';
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

typeof ContainsMethods.boundStatic === 'function';
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

typeof ContainsMethods.unboundStatic === 'function';
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.unbound = () => {};
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.unbound = instance.unbound.bind(instance);
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

if (!!instance.unbound) {}
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

void instance.unbound
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

delete instance.unbound
  `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

const { double } = arith;
  `,
      filename: 'file.ts',
    },
    {
      code: `
interface RecordA {
  readonly type: 'A';
  readonly a: {};
}
interface RecordB {
  readonly type: 'B';
  readonly b: {};
}
type AnyRecord = RecordA | RecordB;

function test(obj: AnyRecord) {
  switch (obj.type) {
  }
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class CommunicationError {
  constructor() {
    const x = CommunicationError.prototype;
  }
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class CommunicationError {}
const x = CommunicationError.prototype;
    `,
      filename: 'file.ts',
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

function foo(instance: ContainsMethods | null) {
  instance?.bound();
  instance?.unbound();

  if (instance?.bound) {
  }
  if (instance?.unbound) {
  }

  typeof instance?.bound === 'function';
  typeof instance?.unbound === 'function';
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
interface OptionalMethod {
  mightBeDefined?(): void;
}

const x: OptionalMethod = {};
declare const myCondition: boolean;
if (myCondition || x.mightBeDefined) {
  console.log('hello world');
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class A {
  unbound(): void {
    this.unbound = undefined;
    this.unbound = this.unbound.bind(this);
  }
}
    `,
      filename: 'file.ts',
    },
    {
      code: 'const { parseInt } = Number;',
      filename: 'file.ts',
    },
    {
      code: 'const { log } = console;',
      filename: 'file.ts',
    },
    {
      code: `
let parseInt;
({ parseInt } = Number);
    `,
      filename: 'file.ts',
    },
    {
      code: `
let log;
({ log } = console);
    `,
      filename: 'file.ts',
    },
    {
      code: `
const foo = {
  bar: 'bar',
};
const { bar } = foo;
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbnound() {}
  bar = 4;
}
const { bar } = new Foo();
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  bound = () => 'foo';
}
const { bound } = new Foo();
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  bound = () => 'foo';
}
function foo({ bound } = new Foo()) {}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  bound = () => 'foo';
}
declare const bar: Foo;
function foo({ bound }: Foo) {}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  bound = () => 'foo';
}
class Bar {
  bound = () => 'bar';
}
function foo({ bound }: Foo | Bar) {}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  bound = () => 'foo';
}
type foo = ({ bound }: Foo) => void;
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
type foo = ({ unbound }: Foo) => void;
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  bound = () => 'foo';
}
class Bar {
  bound = () => 'bar';
}
function foo({ bound }: Foo & Bar) {}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
declare const { unbound }: Foo;
    `,
      filename: 'file.ts',
    },
    {
      code: "declare const { unbound } = '***';",
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
type foo = (a: (b: (c: ({ unbound }: Foo) => void) => void) => void) => void;
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
class Bar {
  property: ({ unbound }: Foo) => void;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
function foo<T extends ({ unbound }: Foo) => void>() {}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
abstract class Bar {
  abstract foo({ unbound }: Foo);
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
declare class Bar {
  foo({ unbound }: Foo);
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
declare function foo({ unbound }: Foo);
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
interface Bar {
  foo: ({ unbound }: Foo) => void;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
interface Bar {
  foo({ unbound }: Foo): void;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
interface Bar {
  new ({ unbound }: Foo): Foo;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
type foo = new ({ unbound }: Foo) => void;
    `,
      filename: 'file.ts',
    },
    {
      code: 'const { unbound } = { unbound: () => {} };',
      filename: 'file.ts',
    },
    {
      code: 'function foo({ unbound }: { unbound: () => void } = { unbound: () => {} }) {}',
      filename: 'file.ts',
    },
    {
      code: `
class BaseClass {
  x: number = 42;
  logThis() {}
}
class OtherClass extends BaseClass {
  superLogThis: any;
  constructor() {
    super();
    this.superLogThis = super.logThis;
  }
}
const oc = new OtherClass();
oc.superLogThis();
    `,
      filename: 'file.ts',
    },
  ],
  invalid: [
    {
      code: `
class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();

Promise.resolve().then(console.log);
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 10,
          column: 24,
          endLine: 10,
          endColumn: 35,
        },
      ],
    },
    {
      code: `
import { console } from './class';
const x = console.log;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};


function foo(arg: ContainsMethods | null) {
  const unbound = arg?.unbound;
  arg.unbound += 1;
  arg?.unbound as any;
}
\x20\x20\x20\x20\x20\x20
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 20,
          column: 19,
          endLine: 20,
          endColumn: 31,
        },
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 21,
          column: 3,
          endLine: 21,
          endColumn: 14,
        },
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 22,
          column: 3,
          endLine: 22,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

const unbound = instance.unbound;
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 17,
          endLine: 18,
          endColumn: 33,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

const unboundStatic = ContainsMethods.unboundStatic;
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 23,
          endLine: 18,
          endColumn: 52,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

const { unbound } = instance;
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 9,
          endLine: 18,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

const { unboundStatic } = ContainsMethods;
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 9,
          endLine: 18,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

<any>instance.unbound;
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 6,
          endLine: 18,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.unbound as any;
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 1,
          endLine: 18,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

<any>ContainsMethods.unboundStatic;
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 6,
          endLine: 18,
          endColumn: 35,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.unboundStatic as any;
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 1,
          endLine: 18,
          endColumn: 30,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.unbound || 0;
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 1,
          endLine: 18,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

ContainsMethods.unboundStatic || 0;
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 1,
          endLine: 18,
          endColumn: 30,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

instance.unbound ? instance.unbound : null
  `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column: 20,
          endLine: 18,
          endColumn: 36,
        },
      ],
    },
    {
      code: `
class ContainsMethods {
  unbound?(): void;

  static unboundStatic?(): void;
}

new ContainsMethods().unbound;

ContainsMethods.unboundStatic;
      `,
      filename: 'file.ts',
      options: [
        {
          ignoreStatic: true,
        },
      ],
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 8,
          column: 1,
          endLine: 8,
          endColumn: 30,
        },
      ],
    },
    {
      code: `
class CommunicationError {
  foo() {}
}
const x = CommunicationError.prototype.foo;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'const x = Promise.all;',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound() {}
}
const instance = new Foo();

let x;

x = instance.unbound; // THIS SHOULD ERROR
instance.unbound = x; // THIS SHOULD NOT
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 9,
          column: 5,
          endLine: 9,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
class Foo extends Number {
  static parseInt = function (string: string, radix?: number): number {};
}
const foo = Foo;
['1', '2', '3'].map(foo.parseInt);
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 6,
          column: 21,
          endLine: 6,
          endColumn: 33,
        },
      ],
    },
    {
      code: `
declare const foo: Number;
const x = foo.toFixed;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
declare const foo: Object;
const x = foo.hasOwnProperty;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
declare const foo: String;
const x = foo.slice;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
declare const foo: Date;
const x = foo.getTime;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
class Foo extends Number {}
const x = Foo.parseInt;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo extends String {}
const x = Foo.fromCharCode;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
class Foo extends Object {}
const x = Foo.defineProperty;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
class Foo extends Date {}
const x = Foo.parse;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
const unbound = new Foo().unbound;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 5,
          column: 17,
          endLine: 5,
          endColumn: 34,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound() {}
}
const { unbound } = new Foo();
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
const { unbound } = new Foo();
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound() {}
}
let unbound;
({ unbound } = new Foo());
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 6,
          column: 4,
          endLine: 6,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
let unbound;
({ unbound } = new Foo());
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 6,
          column: 4,
          endLine: 6,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
function foo({ unbound }: Foo = new Foo()) {}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 5,
          column: 16,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
declare const bar: Foo;
function foo({ unbound }: Foo = bar) {}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
declare const bar: Foo;
function foo({ unbound }: Foo = { unbound: () => {} }) {}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
declare const bar: Foo;
function foo({ unbound }: Foo = { unbound: function () {} }) {}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
function foo({ unbound }: Foo) {}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 5,
          column: 16,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
function bar(cb: (arg: Foo) => void) {}
bar(({ unbound }) => {});
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 6,
          column: 8,
          endLine: 6,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
function bar(cb: (arg: { unbound: () => void }) => void) {}
bar(({ unbound } = new Foo()) => {});
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 6,
          column: 8,
          endLine: 6,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
for (const { unbound } of [new Foo(), new Foo()]) {
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};

  foo({ unbound }: Foo) {}
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = function () {};
}
function foo({ unbound }: Foo | Bar) {}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 8,
          column: 16,
          endLine: 8,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
function foo({ unbound }: { unbound: () => string } | Foo) {}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 5,
          column: 16,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = () => {};
}
function foo({ unbound }: Foo | Bar) {}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 8,
          column: 16,
          endLine: 8,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
const foo = ({ unbound }: Foo & { foo: () => 'bar' }) => {};
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 5,
          column: 16,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = () => {};
}
const foo = ({ unbound }: (Foo & { foo: () => 'bar' }) | Bar) => {};
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 8,
          column: 16,
          endLine: 8,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = () => {};
}
const foo = ({ unbound }: Foo & Bar) => {};
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 8,
          column: 16,
          endLine: 8,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class Foo {
  unbound = function () {};

  other = function () {};
}
class Bar {
  unbound = () => {};
}
const foo = ({ unbound, ...rest }: Foo & Bar) => {};
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 10,
          column: 16,
          endLine: 10,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'const { unbound } = { unbound: function () {} };',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
function foo(
  { unbound }: { unbound: () => void } = { unbound: function () {} },
) {}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
class Foo {
  floor = function () {};
}

const { floor } = Math.random() > 0.5 ? new Foo() : Math;
      `,
      filename: 'file.ts',
      errors: [
        {
          // Upstream visits Math's method first; tsgo visits Foo's function field.
          // Keep the field diagnostic: this annotation cannot exempt such fields.
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
class CommunicationError {
  foo() {}
}
const { foo } = CommunicationError.prototype;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
class CommunicationError {
  foo() {}
}
let foo;
({ foo } = CommunicationError.prototype);
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 6,
          column: 4,
          endLine: 6,
          endColumn: 7,
        },
      ],
    },
    {
      code: `
import { console } from './class';
const { log } = console;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'const { all } = Promise;',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
class BaseClass {
  logThis() {}
}
class OtherClass extends BaseClass {
  constructor() {
    super();
    const x = super.logThis;
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 8,
          column: 15,
          endLine: 8,
          endColumn: 28,
        },
      ],
    },
    {
      code: `
class BaseClass {
  logThis() {}
}
class OtherClass extends BaseClass {
  constructor() {
    super();
    let x;
    x = super.logThis;
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 9,
          column: 9,
          endLine: 9,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
const values = {
  a() {},
  b: () => {},
};

const { a, b } = values;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 7,
          column: 9,
          endLine: 7,
          endColumn: 10,
        },
      ],
    },
    {
      code: `
const values = {
  a() {},
  b: () => {},
};

const { a: c } = values;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 7,
          column: 9,
          endLine: 7,
          endColumn: 10,
        },
      ],
    },
    {
      code: `
const values = {
  a() {},
  b: () => {},
};

const { b, a } = values;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 7,
          column: 12,
          endLine: 7,
          endColumn: 13,
        },
      ],
    },
    {
      code: `
const objectLiteral = {
  f: function () {},
};
const f = objectLiteral.f;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 26,
        },
      ],
    },
  ],
});
