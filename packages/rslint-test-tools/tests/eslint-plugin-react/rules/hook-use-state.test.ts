import { RuleTester } from "../rule-tester";

const ruleTester = new RuleTester();
const error = "useState call is not destructured into value + setter pair";
const destructuredError =
  'useState call is not destructured into value + setter pair (you can allow destructuring by enabling "allowDestructuredState" option)';

ruleTester.run("hook-use-state", {} as never, {
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
  ],
  invalid: [
    {
      code: `import { useState } from 'react'; const result = useState()`,
      errors: [{ message: error }],
    },
    {
      code: `import React from 'react'; const result = React.useState()`,
      errors: [{ message: error }],
    },
    {
      code: `import { useState } from 'react'; const [, setColor] = useState()`,
      errors: [{ message: error }],
    },
    {
      code: `import { useState } from 'react'; const { color } = useState()`,
      errors: [{ message: error }],
    },
    {
      code: `import { useState } from 'react'; const [color, setFlavor] = useState()`,
      errors: [
        {
          message: error,
          suggestions: [
            {
              messageId: "suggestPair",
              output: `import { useState } from 'react'; const [color, setColor] = useState()`,
            },
          ],
        },
      ],
    },
    {
      code: `import { useState } from 'react'; const [color] = useState(initialColor)`,
      errors: [
        {
          message: error,
          suggestions: [
            {
              messageId: "suggestMemo",
              output: `import { useState, useMemo } from 'react'; const color = useMemo(() => initialColor, [])`,
            },
            {
              messageId: "suggestPair",
              output: `import { useState } from 'react'; const [color, setColor] = useState(initialColor)`,
            },
          ],
        },
      ],
    },
    {
      code: `import { useState } from 'react'; const [{foo}, setFoo] = useState({foo: 1})`,
      errors: [{ message: destructuredError }],
    },
  ],
});
