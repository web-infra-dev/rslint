/** Canonical child order and the independent frame-schema contract. */
import { describe, test, expect } from 'rstack/test';
import { KEYS as eslintCoreKeys } from 'eslint-visitor-keys';

import {
  VISITOR_KEYS,
  getVisitorKeys,
} from '../../../src/languages/js/visitor-keys.js';
import { FRAME_SCHEMAS } from '../../../src/languages/js/frame-schema.js';

describe('canonical visitor keys', () => {
  test('every frame type has all canonical child fields', () => {
    for (const [type, fields] of FRAME_SCHEMAS) {
      expect(VISITOR_KEYS[type], type).toBeDefined();
      for (const key of VISITOR_KEYS[type])
        expect(fields, `${type}.${key}`).toContain(key);
    }
  });

  test('JavaScript child order follows ESLint', () => {
    for (const [type, coreKeys] of Object.entries(eslintCoreKeys)) {
      expect(
        VISITOR_KEYS[type]?.filter((key) => coreKeys.includes(key)),
        type,
      ).toEqual(coreKeys);
    }
  });

  test('known types use canonical keys independently of the frame node set', () => {
    expect(
      getVisitorKeys({
        type: 'JSXElement',
        openingElement: {},
        children: [],
        closingElement: {},
        extra: {},
        parent: {},
      }),
    ).toEqual(['openingElement', 'children', 'closingElement']);
    expect(getVisitorKeys({ type: 'TSAbstractKeyword', child: {} })).toEqual(
      [],
    );
  });

  test('unknown synthetic nodes follow ESLint metadata exclusions', () => {
    expect(
      getVisitorKeys({
        type: 'WrapperSynthetic',
        child: {},
        parent: {},
        leadingComments: [],
        trailingComments: [],
        _metadata: {},
      }),
    ).toEqual(['type', 'child']);
  });
});
