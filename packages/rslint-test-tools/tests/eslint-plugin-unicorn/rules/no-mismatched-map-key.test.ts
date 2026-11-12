// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-mismatched-map-key.js

import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

function js(code: string): ValidTestCase {
  return { code, filename: 'src/virtual.js' };
}

function ts(code: string): ValidTestCase {
  return {
    code,
    filename: 'src/virtual.ts',
    languageOptions: {
      parserOptions: { project: false, projectService: false },
    },
  };
}

function withError(testCase: ValidTestCase): InvalidTestCase {
  return {
    ...testCase,
    errors: [
      {
        messageId: 'no-mismatched-map-key',
        message: 'Do not check a Map key before accessing a different key.',
      },
    ],
  };
}

new RuleTester().run('no-mismatched-map-key', null as never, {
  valid: [
    // Guard and access ordering.
    js('const value = map.get(otherKey) && map.has(key);'),
    js('const value = map.get(otherKey) || map.has(key);'),
    js('if (map.has(key) && condition) { map.get(anotherKey); }'),
    js('do { map.get(anotherKey); } while (map.has(key));'),
    js('if (condition || map.has(key)) { map.get(anotherKey); }'),

    // Receiver writes, matching keys and an explicitly typed non-Map receiver.
    js(
      'class A { f() { if (this.map.has(key)) { (this.map = other, this.map.get(anotherKey)); } } }',
    ),
    js(
      'class A { f() { if (this.map.has(key)) { for (this.map of maps) { this.map.get(anotherKey); } } } }',
    ),
    js('const value = map.has(key) && map.get(key);'),
    js('if (map.has(key)) { (map++, map.get(anotherKey)); }'),
    ts(
      'function f(cache: {has(key: string): boolean; get(key: string): number}) { if (cache.has("a")) { return cache.get("b"); } }',
    ),
    js('map.has(key) ? map.get(key) : value;'),
    js('!map.has(key) ? map.set(key, value) : value;'),
    js('if (map.has(key)) { const value = map.get(key); }'),
    js('if (!map.has(key)) { map.set(key, value); }'),

    // Other receivers, calls in keys and unsupported method call forms.
    js('if (map.has(key)) { otherMap.get(anotherKey); }'),
    js('if (map.has(key)) { map.get(getKey()); }'),
    js('if (map.has(getKey())) { map.get(anotherKey); }'),
    js('if (map.has(...key)) { map.get(anotherKey); }'),
    js('if (map.has(key)) { map.set(anotherKey, ...values); }'),
    js('if (map?.has(key)) { map.get(anotherKey); }'),
    js('if (map.has?.(key)) { map.get(anotherKey); }'),
    js('if (map[has](key)) { map.get(anotherKey); }'),
    js('if (map.has(key)) { map[get](anotherKey); }'),
    js('if (map.has(key) && other) { map.get(anotherKey); }'),

    // Equal static keys and a call in a key.
    js('if (map.has(1 + 1)) { map.get(2); }'),
    js('if (map.has(NaN)) { map.get(NaN); }'),
    js('if (map.has(-0)) { map.get(0); }'),
    js('if (map.has(String("key"))) { map.get("anotherKey"); }'),

    // Receiver shadowing and writes.
    js(
      'if (map.has(key)) { const map = otherMap, value = map.get(anotherKey); }',
    ),
    js(
      'if (cache.map.has(key)) { const cache = otherCache, value = cache.map.get(anotherKey); }',
    ),
    js(
      'if (map.has(key)) { const {map} = cache, value = map.get(anotherKey); }',
    ),
    js('if (map.has(key)) { try {} catch (map) { map.get(anotherKey); } }'),
    js('if (map.has(key)) { for (map of maps) { map.get(anotherKey); } }'),
    js('if (map.has(key)) { map = otherMap; map.get(anotherKey); }'),
    js('if (map.has(key)) { var map = otherMap; map.get(anotherKey); }'),
    js('if (map.has(key)) { (map.get(anotherKey), map = otherMap); }'),
    js(
      'if (cache.map.has(key)) { (cache.map = otherMap, cache.map.get(anotherKey)); }',
    ),
    js(
      'if (cache.map.has(key)) { for (cache.map of maps) { cache.map.get(anotherKey); } }',
    ),
    js('if (map.has(key)) { ({map} = other); map.get(anotherKey); }'),
    js('if (map.has(key)) { var {map} = other; map.get(anotherKey); }'),
    js('if (map.has(key)) { for ({map} of maps) { map.get(anotherKey); } }'),

    // Nested guards on the same receiver.
    js(
      'if (map.has(key)) { if (map.has(anotherKey)) { map.get(anotherKey); } }',
    ),
    js(
      'if (map.has(key)) { map.has(anotherKey) ? map.get(anotherKey) : value; }',
    ),
    js(
      'if (map.has(key)) { if (!map.has(anotherKey)) { map.set(anotherKey, value); } }',
    ),
    js(
      'if (map.has(key)) { if (condition && map.has(anotherKey)) { map.get(anotherKey); } }',
    ),
    js(
      'if (map.has(key)) { condition && map.has(anotherKey) ? map.get(anotherKey) : value; }',
    ),

    // Conditional keys, multi-statement branches and nested scope boundaries.
    js('if (map.has(true ? String("key") : "x")) { map.get("anotherKey"); }'),
    js('if (map.has(key)) { map.property = value; map.get(anotherKey); }'),
    js(
      'if (map.has(key)) {\n\tmap.set(key, value);\n\tmap.set(anotherKey, value);\n}',
    ),
    js(
      'if (map.has(key)) {\n\tfor (const map of maps) {\n\t\tmap.get(anotherKey);\n\t}\n}',
    ),
    js(
      'if (map.has(key)) {\n\tfor (const item of items) {\n\t\tconst map = otherMap;\n\t\tmap.get(anotherKey);\n\t}\n}',
    ),
    js(
      'if (map.has(key)) {\n\tfor (const item of items) {\n\t\tclass map {}\n\t\tmap.get(anotherKey);\n\t}\n}',
    ),
    js(
      'if (map.has(key)) {\n\tfor (const item of items) {\n\t\tfunction map() {}\n\t\tmap.get(anotherKey);\n\t}\n}',
    ),
    js(
      'if (map.has(key)) {\n\tfunction getValue() {\n\t\treturn map.get(anotherKey);\n\t}\n}',
    ),
    js(
      'if (map.has(key)) {\n\tclass Value {\n\t\tstatic value = map.get(anotherKey);\n\t}\n}',
    ),

    // TypeScript assertions and mutable member reads.
    ts('if (map.has(key as string)) { map.get(key); }'),
    js(
      "const modes = new Set();\nmodes.add('foo');\nif (map.has(modes.size)) {\n\tmap.get(1);\n}",
    ),
    js(
      'const modes = new Set(["foo"]); modes.clear(); if (map.has((modes.size && 1) || key)) { map.get(1); }',
    ),
    js(
      'const object = {value: true}; Object.defineProperty(object, "value", {get() { return false; }}); if (map.has(object.value ? 2 : key)) { map.get(1); }',
    ),
    ts(
      'const modes = new Set(["foo"]); modes.clear(); if (map.has((modes.size ? 1 : key) as number)) { map.get(1); }',
    ),

    // Destructuring, non-primitive keys and branch boundaries.
    js('if (map.has(key)) { use([, map] = values, map.get(anotherKey)); }'),
    js('if (map.has(key)) { use({...map} = values, map.get(anotherKey)); }'),
    js(
      'if (map.has(key)) { use([map = fallback] = values, map.get(anotherKey)); }',
    ),
    js('if (map.has(key)) { use([...map] = values, map.get(anotherKey)); }'),
    js('if (map.has(key)) { map.get([anotherKey]); }'),
    js(
      'if (map.has(key)) { map.get(key); } else if (condition) { map.get(anotherKey); }',
    ),
    js('while (map.has(key)) { map.get(anotherKey); next(); }'),

    // Documentation examples.
    js('const value = map.has(key) ? map.get(key) : fallback;'),
    js('if (!map.has(key)) {\n\tmap.set(key, value);\n}'),
  ],
  invalid: [
    // An explicitly typed Map.
    withError(
      ts(
        'function f(map: Map<string, number>) { if (map.has("a")) { return map.get("b"); } }',
      ),
    ),

    // Conditional branches and WeakMap-shaped receivers.
    withError(js('map.has(key) ? map.get(anotherKey) : value;')),
    withError(js('map.has(key) ? map.set(anotherKey, value) : value;')),
    withError(js('if (weakMap.has(key)) { weakMap.get(anotherKey); }')),
    withError(js('weakMap.has(key) ? weakMap.get(anotherKey) : value;')),
    withError(js('if (map.has(key)) { const value = map.get(anotherKey); }')),
    withError(js('if (map.has(key)) { map.set(anotherKey, value); }')),
    withError(js('if (!map.has(key)) { const value = map.get(anotherKey); }')),
    withError(js('if (!map.has(key)) { map.set(anotherKey, value); }')),
    withError(js('map.has(key) ? value : map.get(anotherKey);')),
    withError(js('!map.has(key) ? value : map.get(anotherKey);')),
    withError(js('if (map.has(key)) {} else { map.get(anotherKey); }')),
    withError(js('if (!map.has(key)) {} else { map.get(anotherKey); }')),

    // Nested branches and shadowed key bindings.
    withError(
      js(
        'if (map.has(key)) { for (const item of items) { map.get(anotherKey); } }',
      ),
    ),
    withError(
      js(
        'map.has(key) ? (map.has(key) ? map.get(anotherKey) : value) : value;',
      ),
    ),
    withError(
      js('if (map.has(key)) { if (condition) { map.get(anotherKey); } }'),
    ),
    withError(
      js('if (map.has(key)) { condition ? map.get(anotherKey) : value; }'),
    ),
    withError(
      js('if (map.has(key)) { const key = anotherKey, value = map.get(key); }'),
    ),
    withError(
      js(
        'if (map.has(object.key)) { const object = other, value = map.get(object.key); }',
      ),
    ),

    // Literal and member keys, including a TypeScript assertion.
    withError(js('if (map.has("key")) { map.get("anotherKey"); }')),
    withError(js('if (map.has(1)) { map.get(2); }')),
    withError(js('if (map.has(object.key)) { map.get(object.anotherKey); }')),
    withError(
      js('if (this.map.has(this.key)) { this.map.get(this.anotherKey); }'),
    ),
    withError(
      ts('if (map.has(key as string)) { map.get(anotherKey as string); }'),
    ),

    // The final && operand, while loops and short-circuit expressions.
    withError(js('if (condition && map.has(key)) { map.get(anotherKey); }')),
    withError(js('while (map.has(key)) { map.get(anotherKey); }')),
    withError(js('while (!map.has(key)) { map.set(anotherKey, value); }')),
    withError(js('while (map.has(key)) { const x = map.get(anotherKey); }')),
    withError(js('const value = map.has(key) && map.get(otherKey);')),
    withError(js('if (!map.has(key) && map.get(otherKey)) {}')),
    withError(js('const value = map.has(key) || map.get(otherKey);')),

    // An unbraced branch, unrelated writes and a nested function.
    withError(js('if (map.has(key)) map.get(anotherKey);')),
    withError(
      js(
        'if (map.has(key)) { use([object.value] = values, map.get(anotherKey)); }',
      ),
    ),
    withError(js('if (map.has(key)) { (count = 1, map.get(anotherKey)); }')),
    withError(js('if (map.has(key)) { (count++, map.get(anotherKey)); }')),
    withError(
      js(
        'if (map.has(key)) { if (check(() => map.has(anotherKey))) { map.get(anotherKey); } }',
      ),
    ),

    // Documentation examples.
    withError(
      js('const value = map.has(key) ? map.get(anotherKey) : fallback;'),
    ),
    withError(js('if (!map.has(key)) {\n\tmap.set(anotherKey, value);\n}')),
  ],
});
