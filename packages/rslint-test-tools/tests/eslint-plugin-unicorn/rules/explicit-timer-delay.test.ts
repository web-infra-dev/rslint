// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester, type ValidTestCase } from '../rule-tester';

const defaults = {
  filename: 'src/virtual.js',
  languageOptions: {
    sourceType: 'module',
    globals: {
      setTimeout: 'readonly',
      setInterval: 'readonly',
      window: 'readonly',
      globalThis: 'readonly',
      global: 'readonly',
      self: 'readonly',
    },
  },
} satisfies Pick<ValidTestCase, 'filename' | 'languageOptions'>;

const messages = {
  missingDelayTimeout: {
    messageId: 'missing-delay',
    message: '`setTimeout` should have an explicit delay argument.',
  },
  missingDelayInterval: {
    messageId: 'missing-delay',
    message: '`setInterval` should have an explicit delay argument.',
  },
  redundantDelayTimeout: {
    messageId: 'redundant-delay',
    message: '`setTimeout` should not have an explicit delay of `0`.',
  },
  redundantDelayInterval: {
    messageId: 'redundant-delay',
    message: '`setInterval` should not have an explicit delay of `0`.',
  },
};

new RuleTester().run('explicit-timer-delay', {} as never, {
  valid: [
    { ...defaults, code: 'setTimeout(() => console.log("Hello"), 0);' },
    { ...defaults, code: 'setInterval(callback, 0);' },
    { ...defaults, code: 'setTimeout(() => console.log("Hello"), 1000);' },
    { ...defaults, code: 'setInterval(callback, 100);' },
    { ...defaults, code: 'window.setTimeout(() => console.log("Hello"), 0);' },
    { ...defaults, code: 'globalThis.setInterval(callback, 0);' },
    { ...defaults, code: 'global.setTimeout(() => {}, 0);' },
    { ...defaults, code: 'self.setTimeout(() => {}, 0);' },
    { ...defaults, code: 'setTimeout(callback, 0, arg1, arg2);' },
    { ...defaults, code: 'setInterval(callback, 100, arg1);' },
    {
      ...defaults,
      code: "import {setTimeout as delay} from 'node:timers/promises';\n\nawait delay(100);",
    },
    { ...defaults, code: 'setTimeout?.(() => {});' },
    { ...defaults, code: 'window.setTimeout?.(callback);' },
    { ...defaults, code: 'setTimeout();' },
    { ...defaults, code: 'setTimeout(...args);' },
    { ...defaults, code: 'customSetTimeout(callback);' },
    { ...defaults, code: 'obj.customSetTimeout(callback);' },
    { ...defaults, code: 'globalThis["setTimeout"](callback);' },
    { ...defaults, code: 'Math.setTimeout(callback);' },
    { ...defaults, code: 'foo.setTimeout(callback);' },
    { ...defaults, code: 'const window = foo;\nwindow.setTimeout(callback);' },
    { ...defaults, code: 'const self = foo;\nself.setInterval(callback);' },
    {
      ...defaults,
      code: 'setTimeout(() => console.log("Hello"));',
      options: ['never'],
    },
    { ...defaults, code: 'setInterval(callback);', options: ['never'] },
    {
      ...defaults,
      code: 'window.setTimeout(() => console.log("Hello"));',
      options: ['never'],
    },
    {
      ...defaults,
      code: 'globalThis.setInterval(callback);',
      options: ['never'],
    },
    { ...defaults, code: 'setTimeout((callback), 0);' },
    {
      ...defaults,
      code: 'setTimeout(() => console.log("Hello"), 1000);',
      options: ['never'],
    },
    { ...defaults, code: 'setTimeout(callback, (1000));', options: ['never'] },
    { ...defaults, code: 'setInterval(callback, 100);', options: ['never'] },
    {
      ...defaults,
      code: 'setInterval(callback, 0, arg1);',
      options: ['never'],
    },
    {
      ...defaults,
      code: 'setTimeout(callback, 500, arg1);',
      options: ['never'],
    },
    {
      ...defaults,
      code: 'setTimeout(callback, 0, arg1, arg2);',
      options: ['never'],
    },
    {
      ...defaults,
      code: "// ✅\nsetTimeout(() => console.log('Hello'), 1000);\nsetInterval(callback, 100);",
    },
    {
      ...defaults,
      code: "// ✅\nsetTimeout(() => console.log('Hello'), 1000);\nglobalThis.setInterval(callback, 100);",
      options: ['never'],
    },
  ],
  invalid: [
    {
      ...defaults,
      code: 'setTimeout(() => console.log("Hello"));',
      errors: [messages.missingDelayTimeout],
    },
    {
      ...defaults,
      code: 'setInterval(callback);',
      errors: [messages.missingDelayInterval],
    },
    {
      ...defaults,
      code: 'window.setTimeout(() => console.log("Hello"));',
      errors: [messages.missingDelayTimeout],
    },
    {
      ...defaults,
      code: 'globalThis.setInterval(callback);',
      errors: [messages.missingDelayInterval],
    },
    {
      ...defaults,
      code: 'global.setTimeout(fn);',
      errors: [messages.missingDelayTimeout],
    },
    {
      ...defaults,
      code: 'self.setTimeout(fn);',
      errors: [messages.missingDelayTimeout],
    },
    {
      ...defaults,
      code: 'setTimeout(\n\t() => console.log("Hello")\n);',
      errors: [messages.missingDelayTimeout],
    },
    {
      ...defaults,
      code: 'setInterval(\n\tcallback\n);',
      errors: [messages.missingDelayInterval],
    },
    {
      ...defaults,
      code: 'setTimeout((callback));',
      errors: [messages.missingDelayTimeout],
    },
    {
      ...defaults,
      code: 'setTimeout(() => console.log("Hello"), 0);',
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: 'setInterval(callback, 0);',
      options: ['never'],
      errors: [messages.redundantDelayInterval],
    },
    {
      ...defaults,
      code: 'setTimeout((callback), 0);',
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: 'self.setTimeout(fn, 0);',
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: 'self.setInterval(fn, 0);',
      options: ['never'],
      errors: [messages.redundantDelayInterval],
    },
    {
      ...defaults,
      code: 'window.setTimeout(() => console.log("Hello"), 0);',
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: 'globalThis.setInterval(callback, 0);',
      options: ['never'],
      errors: [messages.redundantDelayInterval],
    },
    {
      ...defaults,
      code: 'global.setTimeout(fn, 0);',
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: 'setTimeout(() => console.log("Hello"), -0);',
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: 'setTimeout(callback, +0);',
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: 'setInterval(callback, +(0));',
      options: ['never'],
      errors: [messages.redundantDelayInterval],
    },
    {
      ...defaults,
      code: 'setTimeout(callback, (-0));',
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: 'setTimeout(\n\t() => console.log("Hello"),\n\t0\n);',
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: 'setInterval(\n\tcallback,\n\t0\n);',
      options: ['never'],
      errors: [messages.redundantDelayInterval],
    },
    {
      ...defaults,
      code: 'setTimeout(callback, (0));',
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: "// ❌\nsetTimeout(() => console.log('Hello'));\n\n// ✅\nsetTimeout(() => console.log('Hello'), 0);",
      errors: [messages.missingDelayTimeout],
    },
    {
      ...defaults,
      code: '// ❌\nsetInterval(callback);\n\n// ✅\nsetInterval(callback, 0);',
      errors: [messages.missingDelayInterval],
    },
    {
      ...defaults,
      code: "// ❌\nwindow.setTimeout(() => console.log('Hello'));\n\n// ✅\nwindow.setTimeout(() => console.log('Hello'), 0);",
      errors: [messages.missingDelayTimeout],
    },
    {
      ...defaults,
      code: '// ❌\nglobalThis.setInterval(callback);\n\n// ✅\nglobalThis.setInterval(callback, 0);',
      errors: [messages.missingDelayInterval],
    },
    {
      ...defaults,
      code: "// ❌\nsetTimeout(() => console.log('Hello'), 0);\n\n// ✅\nsetTimeout(() => console.log('Hello'));",
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: '// ❌\nsetInterval(callback, 0);\n\n// ✅\nsetInterval(callback);',
      options: ['never'],
      errors: [messages.redundantDelayInterval],
    },
    {
      ...defaults,
      code: "// ❌\nwindow.setTimeout(() => console.log('Hello'), 0);\n\n// ✅\nwindow.setTimeout(() => console.log('Hello'));",
      options: ['never'],
      errors: [messages.redundantDelayTimeout],
    },
    {
      ...defaults,
      code: '// ❌\nglobalThis.setInterval(callback, 0);\n\n// ✅\nglobalThis.setInterval(callback);',
      options: ['never'],
      errors: [messages.redundantDelayInterval],
    },
  ],
});
