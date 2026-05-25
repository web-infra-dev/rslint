import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('max-lines-per-function', {
  valid: [
    // Test code in global scope doesn't count
    {
      code: 'var x = 5;\nvar x = 2;\n',
      options: [1] as any,
    },

    // Test single line standalone function
    {
      code: 'function name() {}',
      options: [1] as any,
    },

    // Test standalone function with lines of code
    {
      code: 'function name() {\nvar x = 5;\nvar x = 2;\n}',
      options: [4] as any,
    },

    // Test inline arrow function
    {
      code: 'const bar = () => 2',
      options: [1] as any,
    },

    // Test arrow function
    {
      code: 'const bar = () => {\nconst x = 2 + 1;\nreturn x;\n}',
      options: [4] as any,
    },

    // skipBlankLines: false with simple standalone function
    {
      code: 'function name() {\nvar x = 5;\n\t\n \n\nvar x = 2;\n}',
      options: [{ max: 7, skipComments: false, skipBlankLines: false }] as any,
    },

    // skipBlankLines: true with simple standalone function
    {
      code: 'function name() {\nvar x = 5;\n\t\n \n\nvar x = 2;\n}',
      options: [{ max: 4, skipComments: false, skipBlankLines: true }] as any,
    },

    // skipComments: true with an individual single line comment
    {
      code: 'function name() {\nvar x = 5;\nvar x = 2; // end of line comment\n}',
      options: [{ max: 4, skipComments: true, skipBlankLines: false }] as any,
    },

    // skipComments: true with mixed line comments
    {
      code: "function name() {\nvar x = 5;\n// a comment on it's own line\nvar x = 2; // end of line comment\n}",
      options: [{ max: 4, skipComments: true, skipBlankLines: false }] as any,
    },

    // skipComments: true with multiple single line comments
    {
      code: "function name() {\nvar x = 5;\n// a comment on it's own line\n// and another line comment\nvar x = 2; // end of line comment\n}",
      options: [{ max: 4, skipComments: true, skipBlankLines: false }] as any,
    },

    // skipComments: true test with multiple different comment types
    {
      code: 'function name() {\nvar x = 5;\n/* a \n multi \n line \n comment \n*/\n\nvar x = 2; // end of line comment\n}',
      options: [{ max: 5, skipComments: true, skipBlankLines: false }] as any,
    },

    // skipComments: true with multiple different comment types, including trailing and leading whitespace
    {
      code: 'function name() {\nvar x = 5;\n\t/* a comment with leading whitespace */\n/* a comment with trailing whitespace */\t\t\n\t/* a comment with trailing and leading whitespace */\t\t\n/* a \n multi \n line \n comment \n*/\t\t\n\nvar x = 2; // end of line comment\n}',
      options: [{ max: 5, skipComments: true, skipBlankLines: false }] as any,
    },

    // Multiple params on separate lines test
    {
      code: 'function foo(\n    aaa = 1,\n    bbb = 2,\n    ccc = 3\n) {\n    return aaa + bbb + ccc\n}',
      options: [{ max: 7, skipComments: true, skipBlankLines: false }] as any,
    },

    // IIFE validity test (IIFEs: true; under limit)
    {
      code: '(\nfunction\n()\n{\n}\n)\n()',
      options: [
        { max: 4, skipComments: true, skipBlankLines: false, IIFEs: true },
      ] as any,
    },

    // Nested function validity test
    {
      code: 'function parent() {\nvar x = 0;\nfunction nested() {\n    var y = 0;\n    x = 2;\n}\nif ( x === y ) {\n    x++;\n}\n}',
      options: [{ max: 10, skipComments: true, skipBlankLines: false }] as any,
    },

    // Class method validity test
    {
      code: 'class foo {\n    method() {\n        let y = 10;\n        let x = 20;\n        return y + x;\n    }\n}',
      options: [{ max: 5, skipComments: true, skipBlankLines: false }] as any,
    },

    // IIFEs should be recognized if IIFEs: true
    {
      code: '(function(){\n    let x = 0;\n    let y = 0;\n    let z = x + y;\n    let foo = {};\n    return bar;\n}());',
      options: [
        { max: 7, skipComments: true, skipBlankLines: false, IIFEs: true },
      ] as any,
    },

    // IIFEs should not be recognized if IIFEs: false
    {
      code: '(function(){\n    let x = 0;\n    let y = 0;\n    let z = x + y;\n    let foo = {};\n    return bar;\n}());',
      options: [
        { max: 2, skipComments: true, skipBlankLines: false, IIFEs: false },
      ] as any,
    },

    // Arrow IIFEs should be recognized if IIFEs: true
    {
      code: '(() => {\n    let x = 0;\n    let y = 0;\n    let z = x + y;\n    let foo = {};\n    return bar;\n})();',
      options: [
        { max: 7, skipComments: true, skipBlankLines: false, IIFEs: true },
      ] as any,
    },

    // Arrow IIFEs should not be recognized if IIFEs: false
    {
      code: '(() => {\n    let x = 0;\n    let y = 0;\n    let z = x + y;\n    let foo = {};\n    return bar;\n})();',
      options: [
        { max: 2, skipComments: true, skipBlankLines: false, IIFEs: false },
      ] as any,
    },
  ],

  invalid: [
    // Test simple standalone function is recognized
    {
      code: 'function name() {\n}',
      options: [1] as any,
      errors: [
        {
          messageId: 'exceed',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 2,
        },
      ],
    },

    // Test anonymous function assigned to variable is recognized
    {
      code: 'var func = function() {\n}',
      options: [1] as any,
      errors: [
        {
          messageId: 'exceed',
          line: 1,
          column: 12,
          endLine: 2,
          endColumn: 2,
        },
      ],
    },

    // Test arrow functions are recognized
    {
      code: 'const bar = () => {\nconst x = 2 + 1;\nreturn x;\n}',
      options: [3] as any,
      errors: [
        {
          messageId: 'exceed',
          line: 1,
          column: 13,
          endLine: 4,
          endColumn: 2,
        },
      ],
    },

    // Test inline arrow functions are recognized
    {
      code: 'const bar = () =>\n 2',
      options: [1] as any,
      errors: [
        {
          messageId: 'exceed',
          line: 1,
          column: 13,
          endLine: 2,
          endColumn: 3,
        },
      ],
    },

    // Test that option defaults work as expected (51 lines arrow, default max=50)
    {
      code: '() => {' + 'foo\n'.repeat(60) + '}',
      options: [{}] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test skipBlankLines: false
    {
      code: 'function name() {\nvar x = 5;\n\t\n \n\nvar x = 2;\n}',
      options: [{ max: 6, skipComments: false, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test skipBlankLines: false with CRLF line endings
    {
      code: 'function name() {\r\nvar x = 5;\r\n\t\r\n \r\n\r\nvar x = 2;\r\n}',
      options: [{ max: 6, skipComments: true, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test skipBlankLines: true
    {
      code: 'function name() {\nvar x = 5;\n\t\n \n\nvar x = 2;\n}',
      options: [{ max: 2, skipComments: true, skipBlankLines: true }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test skipBlankLines: true with CRLF line endings
    {
      code: 'function name() {\r\nvar x = 5;\r\n\t\r\n \r\n\r\nvar x = 2;\r\n}',
      options: [{ max: 2, skipComments: true, skipBlankLines: true }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test skipComments: true and skipBlankLines: false for multiple types of comment
    {
      code: 'function name() { // end of line comment\nvar x = 5; /* mid line comment */\n\t// single line comment taking up whole line\n\t\n \n\nvar x = 2;\n}',
      options: [{ max: 6, skipComments: true, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test skipComments: true and skipBlankLines: true for multiple types of comment
    {
      code: 'function name() { // end of line comment\nvar x = 5; /* mid line comment */\n\t// single line comment taking up whole line\n\t\n \n\nvar x = 2;\n}',
      options: [{ max: 1, skipComments: true, skipBlankLines: true }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test skipComments: false and skipBlankLines: true for multiple types of comment
    {
      code: 'function name() { // end of line comment\nvar x = 5; /* mid line comment */\n\t// single line comment taking up whole line\n\t\n \n\nvar x = 2;\n}',
      options: [{ max: 1, skipComments: false, skipBlankLines: true }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test simple standalone function with params on separate lines
    {
      code: 'function foo(\n    aaa = 1,\n    bbb = 2,\n    ccc = 3\n) {\n    return aaa + bbb + ccc\n}',
      options: [{ max: 2, skipComments: true, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test IIFE "function" keyword is included in the count
    {
      code: '(\nfunction\n()\n{\n}\n)\n()',
      options: [
        { max: 2, skipComments: true, skipBlankLines: false, IIFEs: true },
      ] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test nested functions are included in their parent's function count.
    {
      code: 'function parent() {\nvar x = 0;\nfunction nested() {\n    var y = 0;\n    x = 2;\n}\nif ( x === y ) {\n    x++;\n}\n}',
      options: [{ max: 9, skipComments: true, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test parent + nested both reported when both exceed.
    {
      code: 'function parent() {\nvar x = 0;\nfunction nested() {\n    var y = 0;\n    x = 2;\n}\nif ( x === y ) {\n    x++;\n}\n}',
      options: [{ max: 2, skipComments: true, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test regular methods are recognized
    {
      code: 'class foo {\n    method() {\n        let y = 10;\n        let x = 20;\n        return y + x;\n    }\n}',
      options: [{ max: 2, skipComments: true, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test static methods are recognized
    {
      code: 'class A {\n    static\n    foo\n    (a) {\n        return a\n    }\n}',
      options: [{ max: 2, skipComments: true, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test getters are recognized as properties
    {
      code: 'var obj = {\n    get\n    foo\n    () {\n        return 1\n    }\n}',
      options: [{ max: 2, skipComments: true, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test setters are recognized as properties
    {
      code: 'var obj = {\n    set\n    foo\n    ( val ) {\n        this._foo = val;\n    }\n}',
      options: [{ max: 2, skipComments: true, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test computed property names
    {
      code: 'class A {\n    static\n    [\n        foo +\n            bar\n    ]\n    (a) {\n        return a\n    }\n}',
      options: [{ max: 2, skipComments: true, skipBlankLines: false }] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test the IIFEs option includes IIFEs
    {
      code: '(function(){\n    let x = 0;\n    let y = 0;\n    let z = x + y;\n    let foo = {};\n    return bar;\n}());',
      options: [
        { max: 2, skipComments: true, skipBlankLines: false, IIFEs: true },
      ] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },

    // Test the IIFEs option includes arrow IIFEs
    {
      code: '(() => {\n    let x = 0;\n    let y = 0;\n    let z = x + y;\n    let foo = {};\n    return bar;\n})();',
      options: [
        { max: 2, skipComments: true, skipBlankLines: false, IIFEs: true },
      ] as any,
      errors: [
        {
          messageId: 'exceed',
        },
      ],
    },
  ],
});
