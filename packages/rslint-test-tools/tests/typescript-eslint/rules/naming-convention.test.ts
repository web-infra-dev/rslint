import { RuleTester } from '@typescript-eslint/rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('naming-convention', {
  valid: [
    // Default config: camelCase for most, PascalCase for types
    { code: 'const myVariable = 1;' },
    { code: 'let anotherVar = "hello";' },
    { code: 'function myFunction() {}' },
    { code: 'class MyClass {}' },
    { code: 'interface MyInterface {}' },
    { code: 'type MyType = string;' },
    { code: 'enum MyEnum { myValue = 1 }' },

    // Leading/trailing underscore allowed by default
    { code: 'const _privateVar = 1;' },
    { code: 'const trailingVar_ = 1;' },

    // Variables with UPPER_CASE (default allows camelCase + UPPER_CASE for variables)
    { code: 'const MY_CONSTANT = 1;' },

    // Imports (default: camelCase or PascalCase)
    { code: "import myModule from 'module';" },
    { code: "import MyModule from 'module';" },
    { code: "import { myExport } from 'module';" },
    { code: "import { MyExport } from 'module';" },

    // Custom: snake_case for variables
    {
      code: 'const my_variable = 1;',
      options: [
        {
          selector: 'variable',
          format: ['snake_case'],
        },
      ],
    },

    // Custom: PascalCase for variables
    {
      code: 'const MyVariable = 1;',
      options: [
        {
          selector: 'variable',
          format: ['PascalCase'],
        },
      ],
    },

    // Leading underscore required
    {
      code: 'const _myVariable = 1;',
      options: [
        {
          selector: 'variable',
          format: ['camelCase'],
          leadingUnderscore: 'require',
        },
      ],
    },

    // Trailing underscore required
    {
      code: 'const myVariable_ = 1;',
      options: [
        {
          selector: 'variable',
          format: ['camelCase'],
          trailingUnderscore: 'require',
        },
      ],
    },

    // Format null - skip format check
    {
      code: 'const ANY_NAME = 1;',
      options: [
        {
          selector: 'variable',
          format: null,
        },
      ],
    },

    // Class members with default
    { code: 'class MyClass { myProperty = 1; myMethod() {} }' },

    // Function expression assigned to variable
    { code: 'const myFunc = () => {};' },
    { code: 'const myFunc = function() {};' },

    // Object literal
    { code: 'const obj = { myProp: 1 };' },

    // Type parameter
    { code: 'function foo<T>() {}' },
    { code: 'function foo<TData>() {}' },

    // Multiple formats allowed
    {
      code: 'const myVar = 1; const MY_VAR = 2;',
      options: [
        {
          selector: 'variable',
          format: ['camelCase', 'UPPER_CASE'],
        },
      ],
    },

    // Parameter
    { code: 'function fn(myParam: string) {}' },

    // Destructured variable
    { code: 'const { myProp } = ({} as any);' },

    // Accessor
    { code: 'class Foo { get myProp() { return 1; } set myProp(v: number) {} }' },

    // Filter: skip names matching pattern
    {
      code: 'const __special__ = 1; const myNormal = 2;',
      options: [
        {
          selector: 'variable',
          format: ['camelCase'],
          filter: {
            regex: '^__.*__$',
            match: false,
          },
        },
      ],
    },
  ],

  invalid: [
    // Variable violating camelCase (default)
    {
      code: 'const my_variable = 1;',
      errors: [{ messageId: 'doesNotMatchFormat' }],
    },

    // Function violating camelCase (default)
    {
      code: 'function MyFunction() {}',
      errors: [{ messageId: 'doesNotMatchFormat' }],
    },

    // Class violating PascalCase (default)
    {
      code: 'class myClass {}',
      errors: [{ messageId: 'doesNotMatchFormat' }],
    },

    // Interface violating PascalCase (default)
    {
      code: 'interface myInterface {}',
      errors: [{ messageId: 'doesNotMatchFormat' }],
    },

    // Type alias violating PascalCase (default)
    {
      code: 'type myType = string;',
      errors: [{ messageId: 'doesNotMatchFormat' }],
    },

    // Enum violating PascalCase (default)
    {
      code: 'enum myEnum { a }',
      errors: [{ messageId: 'doesNotMatchFormat' }],
    },

    // Leading underscore forbidden
    {
      code: 'const _myVariable = 1;',
      options: [
        {
          selector: 'variable',
          format: ['camelCase'],
          leadingUnderscore: 'forbid',
        },
      ],
      errors: [{ messageId: 'unexpectedUnderscore' }],
    },

    // Trailing underscore forbidden
    {
      code: 'const myVariable_ = 1;',
      options: [
        {
          selector: 'variable',
          format: ['camelCase'],
          trailingUnderscore: 'forbid',
        },
      ],
      errors: [{ messageId: 'unexpectedUnderscore' }],
    },

    // Leading underscore required but missing
    {
      code: 'const myVariable = 1;',
      options: [
        {
          selector: 'variable',
          format: ['camelCase'],
          leadingUnderscore: 'require',
        },
      ],
      errors: [{ messageId: 'missingUnderscore' }],
    },

    // Missing prefix
    {
      code: 'const active = true;',
      options: [
        {
          selector: 'variable',
          format: ['camelCase'],
          prefix: ['is', 'has'],
        },
      ],
      errors: [{ messageId: 'missingAffix' }],
    },

    // Missing suffix
    {
      code: 'const name = "hello";',
      options: [
        {
          selector: 'variable',
          format: ['camelCase'],
          suffix: ['Str', 'Num'],
        },
      ],
      errors: [{ messageId: 'missingAffix' }],
    },

    // Custom regex not matching
    {
      code: 'const myVar = 1;',
      options: [
        {
          selector: 'variable',
          format: ['camelCase'],
          custom: {
            regex: '\\d+$',
            match: true,
          },
        },
      ],
      errors: [{ messageId: 'satisfyCustom' }],
    },

    // Parameter violating format
    {
      code: 'function fn(MY_PARAM: string) {}',
      errors: [{ messageId: 'doesNotMatchFormat' }],
    },

    // Strict camelCase violation
    {
      code: 'const myID = 1;',
      options: [
        {
          selector: 'variable',
          format: ['strictCamelCase'],
        },
      ],
      errors: [{ messageId: 'doesNotMatchFormat' }],
    },
  ],
});
