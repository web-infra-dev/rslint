import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-import-assign', {
  valid: [
    "import mod from 'mod'; mod.prop = 0",
    "import {named} from 'mod'; named.prop = 0",
    "import * as mod from 'mod'; mod.named.prop = 0",
  ],
  invalid: [
    {
      code: "import mod from 'mod'; mod = 0",
      errors: [{ messageId: 'readonly' }],
    },
    {
      code: "import {named} from 'mod'; named = 0",
      errors: [{ messageId: 'readonly' }],
    },
    {
      code: "import * as mod from 'mod'; mod = 0",
      errors: [{ messageId: 'readonly' }],
    },
    {
      code: "import * as mod from 'mod'; mod.named = 0",
      errors: [{ messageId: 'readonlyMember' }],
    },
  ],
});
