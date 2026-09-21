import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-export-in-script-setup', null as never, {
  valid: [
    {
      filename: 'src/plain-script.vue',
      code: `<script>\nexport default { name: 'App' };\n</script>\n`,
    },
    {
      filename: 'src/no-export.vue',
      code: `<script setup>\nconst value = 1;\n</script>\n`,
    },
    {
      filename: 'src/exports-in-plain-block.vue',
      code:
        `<script>\nexport const shared = 1;\n</script>\n` +
        `<script setup>\nconst local = 2;\n</script>\n`,
    },
    {
      filename: 'src/type-only.vue',
      code: `<script setup lang="ts">\nexport type Value = number;\n</script>\n`,
    },
  ],
  invalid: [
    {
      filename: 'src/export-default.vue',
      code: `<script setup>\nexport default { name: 'App' };\n</script>\n`,
      errors: [{ messageId: 'forbidden', line: 2 }],
    },
    {
      filename: 'src/export-const.vue',
      code: `<script setup>\nexport const value = 1;\n</script>\n`,
      errors: [{ messageId: 'forbidden', line: 2 }],
    },
    {
      filename: 'src/export-named.vue',
      code: `<script setup>\nconst value = 1;\nexport { value };\n</script>\n`,
      errors: [{ messageId: 'forbidden', line: 3 }],
    },
    {
      filename: 'src/only-setup-block.vue',
      code:
        `<script>\nexport const shared = 1;\n</script>\n` +
        `<script setup>\nexport const local = 2;\n</script>\n`,
      errors: [{ messageId: 'forbidden', line: 5 }],
    },
  ],
});
