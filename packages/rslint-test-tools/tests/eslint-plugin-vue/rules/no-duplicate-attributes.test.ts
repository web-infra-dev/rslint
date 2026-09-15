import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-duplicate-attributes', null as never, {
  valid: [
    {
      filename: 'src/distinct.vue',
      code: `<template><div foo="a" bar="b"/></template>\n`,
    },
    {
      filename: 'src/class-coexists.vue',
      code: `<template><div class="a" :class="b"/></template>\n`,
    },
    {
      filename: 'src/style-coexists.vue',
      code: `<template><div style="color:red" :style="s"/></template>\n`,
    },
    {
      filename: 'src/object-spread.vue',
      code: `<template><div v-bind="attrs" foo="a"/></template>\n`,
    },
  ],
  invalid: [
    {
      filename: 'src/duplicate-plain.vue',
      code: `<template><div foo="a" foo="b"/></template>\n`,
      errors: [{ messageId: 'duplicateAttribute', line: 1 }],
    },
    {
      filename: 'src/plain-and-bound.vue',
      code: `<template><div foo="a" :foo="b"/></template>\n`,
      errors: [{ messageId: 'duplicateAttribute', line: 1 }],
    },
    {
      filename: 'src/duplicate-class.vue',
      code: `<template><div class="a" class="b"/></template>\n`,
      errors: [{ messageId: 'duplicateAttribute', line: 1 }],
    },
    {
      filename: 'src/class-not-allowed.vue',
      code: `<template><div class="a" :class="b"/></template>\n`,
      options: [{ allowCoexistClass: false }],
      errors: [{ messageId: 'duplicateAttribute', line: 1 }],
    },
  ],
});
