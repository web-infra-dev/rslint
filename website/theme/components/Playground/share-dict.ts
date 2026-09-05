/**
 * DEFLATE preset dictionary for playground share links, version 1.
 *
 * A shared link is a DEFLATE stream whose back-references may point into this
 * dictionary, so the bytes below never travel in the URL. A config that is the
 * default with one rule flipped collapses to a couple of back-references plus
 * the handful of bytes that actually differ.
 *
 * DO NOT EDIT. This is a frozen snapshot of what the editor defaults happened
 * to be when share links were introduced, not a reference to them: the defaults
 * in `share-url.ts` are free to change, and this file must not follow. Changing
 * a single byte here makes every link ever produced decode to garbage.
 *
 * If the dictionary ever needs to be revised, add a `SHARE_DICT_V2` alongside
 * this one, bump `SHARE_VERSION`, and keep decoding v1 links with v1.
 *
 * The most valuable content sits at the end: DEFLATE encodes short distances in
 * fewer bits, so the text most likely to be matched belongs closest to the
 * start of the compressed input.
 */
export const SHARE_DICT_V1 = `{
  "compilerOptions": {
    "target": "ESNext",
    "module": "ESNext",
    "strict": true,
    "strictNullChecks": true
  }
}
import { defineConfig, js, ts } from '@rslint/core';

export default defineConfig([
  js.configs.recommended,
  ts.configs.recommendedTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      '@typescript-eslint/no-unsafe-member-access': 'error',
    },
  },
]);`;
