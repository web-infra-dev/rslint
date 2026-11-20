import { RuleTester, type ExpectedSuggestion } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectUseStateErrorMessageError(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'useStateErrorMessage',
    message: 'useState call is not destructured into value + setter pair',
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

function expectUseStateErrorMessageOrAddOptionError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'useStateErrorMessageOrAddOption',
    message:
      'useState call is not destructured into value + setter pair (you can allow destructuring by enabling "allowDestructuredState" option)',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);
const error = 'useState call is not destructured into value + setter pair';
const destructuredError =
  'useState call is not destructured into value + setter pair (you can allow destructuring by enabling "allowDestructuredState" option)';

ruleTester.run('hook-use-state', {
  valid: [
    {
      code: `import { useState } from 'react'; const [color, setColor] = useState()`,
    },
    {
      code: `import { useState } from 'react'; const [rgb, setRGB] = useState()`,
    },
    {
      code: `import React from 'react'; const [color, setColor] = React.useState()`,
    },
    {
      code: `import { useState as alternative } from 'react'; const [color, setColor] = alternative()`,
    },
    {
      code: `import { useState } from 'react'; function useColor() { return useState() }`,
    },
    {
      code: `import { useState } from 'react'; function useColor() { function useState() {} const result = useState() }`,
    },
    {
      code: `import { useState } from 'react'; const [{foo}, setFoo] = useState({foo: 1})`,
      options: [{ allowDestructuredState: true }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [color, setColor] = useState()
          return [color, setColor]
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useRGB() {
          const [rgb, setRGB] = useState()
          return [rgb, setRGB]
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useRGBValue() {
          const [rgbValue, setRGBValue] = useState()
          return [rgbValue, setRGBValue]
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useCustomColorValue() {
          const [customColorValue, setCustomColorValue] = useState()
          return [customColorValue, setCustomColorValue]
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [color, setColor] = useState('#ffffff')
          return [color, setColor]
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        import React from 'react'
        function useColor() {
          const [color, setColor] = React.useState()
          return [color, setColor]
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [color1, setColor1] = useState()
          return [color1, setColor1]
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `useState()`,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `const result = useState()`,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `const [color, setFlavor] = useState()`,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        import React from 'react'
        import useState from 'someOtherUseState'
        const [color, setFlavor] = useState()
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        import { useRef } from 'react'
        const result = useState()
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        import { useState as useStateAlternativeName } from 'react'
        function useColor() {
          const [color, setColor] = useStateAlternativeName()
          return [color, setColor]
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `const result = React.useState()`,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          return useState()
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          function useState() { // shadows React's useState
            return null
          }

          const result = useState()
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        import React from 'react'
        function useColor() {
          const React = {
            useState: () => {
              return null
            }
          }

          const result = React.useState()
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        import { useState } from 'react';
        function useColor() {
          const [color, setColor] = useState<string>()
          return [color, setColor]
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        import { useState } from 'react';
        function useColor() {
          const [color, setColor] = useState<string>('#ffffff')
          return [color, setColor]
        }
      `,
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        import { useState } from 'react';

        const [{foo, bar, baz}, setFooBarBaz] = useState({foo: "bbb", bar: "aaa", baz: "qqq"})
      `,
      options: [{ allowDestructuredState: true }],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        import { useState } from 'react';

        const [[index, value], setValueWithIndex] = useState([0, "hello"])
      `,
      options: [{ allowDestructuredState: true }],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
  ],
  invalid: [
    {
      code: `import { useState } from 'react'; const result = useState()`,
      errors: [expectUseStateErrorMessageError([1, 50, 1, 60])],
    },
    {
      code: `import React from 'react'; const result = React.useState()`,
      errors: [expectUseStateErrorMessageError([1, 43, 1, 59])],
    },
    {
      code: `import { useState } from 'react'; const [, setColor] = useState()`,
      errors: [expectUseStateErrorMessageError([1, 41, 1, 53])],
    },
    {
      code: `import { useState } from 'react'; const { color } = useState()`,
      errors: [expectUseStateErrorMessageError([1, 53, 1, 63])],
    },
    {
      code: `import { useState } from 'react'; const [color, setFlavor] = useState()`,
      errors: [
        expectUseStateErrorMessageError(
          [1, 41, 1, 59],
          [
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "import { useState } from 'react'; const [color, setColor] = useState()",
            },
          ],
        ),
      ],
    },
    {
      code: `import { useState } from 'react'; const [color] = useState(initialColor)`,
      errors: [
        expectUseStateErrorMessageError(
          [1, 41, 1, 48],
          [
            {
              messageId: 'suggestMemo',
              desc: 'Replace useState call with useMemo',
              output:
                "import { useState, useMemo } from 'react'; const color = useMemo(() => initialColor, [])",
            },
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "import { useState } from 'react'; const [color, setColor] = useState(initialColor)",
            },
          ],
        ),
      ],
    },
    {
      code: `import { useState } from 'react'; const [{foo}, setFoo] = useState({foo: 1})`,
      errors: [expectUseStateErrorMessageOrAddOptionError([1, 41, 1, 56])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        import { useState } from 'react';
        const result = useState()
      `,
      errors: [expectUseStateErrorMessageError([3, 24, 3, 34])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        import { useState } from 'react';
        function useColor() {
          const result = useState()
          return result
        }
      `,
      errors: [expectUseStateErrorMessageError([4, 26, 4, 36])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        import { useState as useStateAlternativeName } from 'react'
        function useColor() {
          const result = useStateAlternativeName()
          return result
        }
      `,
      errors: [expectUseStateErrorMessageError([4, 26, 4, 51])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        import React from 'react'
        function useColor() {
          const result = React.useState()
          return result
        }
      `,
      errors: [expectUseStateErrorMessageError([4, 26, 4, 42])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        import ReactAlternative from 'react'
        function useColor() {
          const result = ReactAlternative.useState()
          return result
        }
      `,
      errors: [expectUseStateErrorMessageError([4, 26, 4, 53])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const result = useState()
          return result
        }
      `,
      errors: [expectUseStateErrorMessageError([4, 26, 4, 36])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [, , extra1] = useState()
        }
      `,
      errors: [expectUseStateErrorMessageError([4, 17, 4, 29])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [, setColor] = useState()
        }
      `,
      errors: [expectUseStateErrorMessageError([4, 17, 4, 29])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const { color } = useState()
        }
      `,
      errors: [expectUseStateErrorMessageError([4, 29, 4, 39])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [] = useState()
        }
      `,
      errors: [expectUseStateErrorMessageError([4, 17, 4, 19])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [, , , ,] = useState()
        }
      `,
      errors: [expectUseStateErrorMessageError([4, 17, 4, 26])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [color] = useState()
        }
      `,
      errors: [
        expectUseStateErrorMessageError(
          [4, 17, 4, 24],
          [
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import { useState } from 'react'\n        function useColor() {\n          const [color, setColor] = useState()\n        }\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor(initialColor) {
          const [color] = useState(initialColor)
        }
      `,
      errors: [
        expectUseStateErrorMessageError(
          [4, 17, 4, 24],
          [
            {
              messageId: 'suggestMemo',
              desc: 'Replace useState call with useMemo',
              output:
                "\n        import { useState, useMemo } from 'react'\n        function useColor(initialColor) {\n          const color = useMemo(() => initialColor, [])\n        }\n      ",
            },
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import { useState } from 'react'\n        function useColor(initialColor) {\n          const [color, setColor] = useState(initialColor)\n        }\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        import { useState, useMemo as useMemoAlternative } from 'react'
        function useColor(initialColor) {
          const [color] = useState(initialColor)
        }
      `,
      errors: [
        expectUseStateErrorMessageError(
          [4, 17, 4, 24],
          [
            {
              messageId: 'suggestMemo',
              desc: 'Replace useState call with useMemo',
              output:
                "\n        import { useState, useMemo as useMemoAlternative } from 'react'\n        function useColor(initialColor) {\n          const color = useMemoAlternative(() => initialColor, [])\n        }\n      ",
            },
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import { useState, useMemo as useMemoAlternative } from 'react'\n        function useColor(initialColor) {\n          const [color, setColor] = useState(initialColor)\n        }\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        import React from 'react'
        function useColor(initialColor) {
          const [color] = React.useState(initialColor)
        }
      `,
      errors: [
        expectUseStateErrorMessageError(
          [4, 17, 4, 24],
          [
            {
              messageId: 'suggestMemo',
              desc: 'Replace useState call with useMemo',
              output:
                "\n        import React from 'react'\n        function useColor(initialColor) {\n          const color = React.useMemo(() => initialColor, [])\n        }\n      ",
            },
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import React from 'react'\n        function useColor(initialColor) {\n          const [color, setColor] = React.useState(initialColor)\n        }\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [color, , extra1] = useState()
          return [color]
        }
      `,
      errors: [
        expectUseStateErrorMessageError(
          [4, 17, 4, 34],
          [
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import { useState } from 'react'\n        function useColor() {\n          const [color, setColor] = useState()\n          return [color]\n        }\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [color, setColor, extra1, extra2, extra3] = useState()
          return [color, setColor]
        }
      `,
      errors: [
        expectUseStateErrorMessageError(
          [4, 17, 4, 58],
          [
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import { useState } from 'react'\n        function useColor() {\n          const [color, setColor] = useState()\n          return [color, setColor]\n        }\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        import { useState } from 'react'
        const [, makeColor] = useState()
      `,
      errors: [expectUseStateErrorMessageError([3, 15, 3, 28])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        import { useState } from 'react'
        const [color, setFlavor, extraneous] = useState()
      `,
      errors: [
        expectUseStateErrorMessageError(
          [3, 15, 3, 45],
          [
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import { useState } from 'react'\n        const [color, setColor] = useState()\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        import { useState } from 'react'
        const [color, setFlavor] = useState()
      `,
      errors: [
        expectUseStateErrorMessageError(
          [3, 15, 3, 33],
          [
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import { useState } from 'react'\n        const [color, setColor] = useState()\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        import { useState } from 'react';

        const [{foo, bar, baz}, setFooBarBaz] = useState({foo: "bbb", bar: "aaa", baz: "qqq"})
      `,
      errors: [expectUseStateErrorMessageOrAddOptionError([4, 15, 4, 46])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        import { useState } from 'react';

        const [[index, value], setValueWithIndex] = useState([0, "hello"])
      `,
      errors: [expectUseStateErrorMessageOrAddOptionError([4, 15, 4, 50])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        import { useState } from 'react';

        const [{foo, bar, baz}, {setFooBarBaz}] = useState({foo: "bbb", bar: "aaa", baz: "qqq"})
      `,
      options: [{ allowDestructuredState: true }],
      errors: [expectUseStateErrorMessageError([4, 15, 4, 48])],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        import { useState } from 'react'
        const [color, setFlavor] = useState<string>()
      `,
      errors: [
        expectUseStateErrorMessageError(
          [3, 15, 3, 33],
          [
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import { useState } from 'react'\n        const [color, setColor] = useState<string>()\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        import { useState } from 'react'
        function useColor() {
          const [color, setFlavor] = useState<string>('#ffffff')
          return [color, setFlavor]
        }
      `,
      errors: [
        expectUseStateErrorMessageError(
          [4, 17, 4, 35],
          [
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import { useState } from 'react'\n        function useColor() {\n          const [color, setColor] = useState<string>('#ffffff')\n          return [color, setFlavor]\n        }\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        import React from 'react'
        function useColor() {
          const [color, setFlavor] = React.useState<string>('#ffffff')
          return [color, setFlavor]
        }
      `,
      errors: [
        expectUseStateErrorMessageError(
          [4, 17, 4, 35],
          [
            {
              messageId: 'suggestPair',
              desc: 'Destructure useState call into value + setter pair',
              output:
                "\n        import React from 'react'\n        function useColor() {\n          const [color, setColor] = React.useState<string>('#ffffff')\n          return [color, setFlavor]\n        }\n      ",
            },
          ],
        ),
      ],
      languageOptions: { ecmaVersion: 2018, sourceType: 'module' },
    },
  ],
});
