import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('arrow-body-style', {
  valid: [
    'var foo = () => {};',
    'var foo = () => 0;',
    'var addToB = (a) => { b =  b + a };',
    'var foo = () => { /* do nothing */ };',
    'var foo = () => {\n /* do nothing */ \n};',
    'var foo = (retv, name) => {\nretv[name] = true;\nreturn retv;\n};',
    'var foo = () => ({});',
    'var foo = () => bar();',
    'var foo = () => { bar(); };',
    'var foo = () => { b = a };',
    'var foo = () => { bar: 1 };',
    { code: 'var foo = () => { return 0; };', options: ['always'] as any },
    { code: 'var foo = () => { return bar(); };', options: ['always'] as any },
    { code: 'var foo = () => 0;', options: ['never'] as any },
    { code: 'var foo = () => ({ foo: 0 });', options: ['never'] as any },
    {
      code: 'var foo = () => {};',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
    },
    {
      code: 'var foo = () => 0;',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
    },
    {
      code: 'var addToB = (a) => { b =  b + a };',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
    },
    {
      code: 'var foo = () => { /* do nothing */ };',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
    },
    {
      code: 'var foo = () => {\n /* do nothing */ \n};',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
    },
    {
      code: 'var foo = (retv, name) => {\nretv[name] = true;\nreturn retv;\n};',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
    },
    {
      code: 'var foo = () => bar();',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
    },
    {
      code: 'var foo = () => { bar(); };',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
    },
    {
      code: 'var foo = () => { return { bar: 0 }; };',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
    },
  ],
  invalid: [
    {
      code: 'for (var foo = () => { return a in b ? bar : () => {} } ;;);',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'a in b; for (var f = () => { return c };;);',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (a = b => { return c in d ? e : f } ;;);',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (var f = () => { return a };;);',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (var f;f = () => { return a };);',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (var f = () => { return a in c };;);',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (var f;f = () => { return a in c };);',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (;;){var f = () => { return a in c }}',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (a = b => { return c = d in e } ;;);',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (var a;;a = b => { return c = d in e } );',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (let a = (b, c, d) => { return vb && c in d; }; ;);',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (let a = (b, c, d) => { return v in b && c in d; }; ;);',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'function foo(){ for (let a = (b, c, d) => { return v in b && c in d; }; ;); }',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for ( a = (b, c, d) => { return v in b && c in d; }; ;);',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for ( a = (b) => { return (c in d) }; ;);',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (let a = (b, c, d) => { return vb in dd ; }; ;);',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'for (let a = (b, c, d) => { return vb in c in dd ; }; ;);',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'do{let a = () => {return f in ff}}while(true){}',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'do{for (let a = (b, c, d) => { return vb in c in dd ; }; ;);}while(true){}',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'scores.map(score => { return x in +(score / maxScore).toFixed(2)});',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'const fn = (a, b) => { return a + x in Number(b) };',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => 0',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: 'var foo = () => 0;',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: 'var foo = () => ({});',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: 'var foo = () => (  {});',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: '(() => ({}))',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: '(() => ( {}))',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: 'var foo = () => { return 0; };',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return 0 };',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return bar(); };',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => {};',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedEmptyBlock' }],
    },
    {
      code: 'var foo = () => {\nreturn 0;\n};',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return { bar: 0 }; };',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedObjectBlock' }],
    },
    {
      code: 'var foo = () => { return ({ bar: 0 }); };',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return a, b }',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return };',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return; };',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return ( /* a */ {ok: true} /* b */ ) };',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: `var foo = () => { return '{' };`,
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return { bar: 0 }.bar; };',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedObjectBlock' }],
    },
    {
      code: 'var foo = (retv, name) => {\nretv[name] = true;\nreturn retv;\n};',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedOtherBlock' }],
    },
    {
      code: 'var foo = () => { bar };',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedOtherBlock' }],
    },
    {
      code: 'var foo = () => { return 0; };',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return bar(); };',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => ({});',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: 'var foo = () => ({ bar: 0 });',
      options: ['as-needed', { requireReturnForObjectLiteral: true }] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: 'var foo = () => (((((((5)))))));',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: 'var foo = () => { return bar }\n[1, 2, 3].map(foo)',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return bar }\n(1).toString();',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => { return bar };\n[1, 2, 3].map(foo)',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = /* a */ ( /* b */ ) /* c */ => /* d */ { /* e */ return /* f */ 5 /* g */ ; /* h */ } /* i */ ;',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = /* a */ ( /* b */ ) /* c */ => /* d */ ( /* e */ 5 /* f */ ) /* g */ ;',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: 'var foo = () => {\nreturn bar;\n};',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => {\nreturn bar;};',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: 'var foo = () => {return bar;\n};',
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: `
              var foo = () => {
                return foo
                  .bar;
              };
            `,
      errors: [{ messageId: 'unexpectedSingleBlock' }],
    },
    {
      code: `
              var foo = () => {
                return {
                  bar: 1,
                  baz: 2
                };
              };
            `,
      errors: [{ messageId: 'unexpectedObjectBlock' }],
    },
    {
      code: 'var foo = () => ({foo: 1}).foo();',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: 'var foo = () => ({foo: 1}.foo());',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: 'var foo = () => ( {foo: 1} ).foo();',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: `
              var foo = () => ({
                  bar: 1,
                  baz: 2
                });
            `,
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    {
      code: `
              parsedYears = _map(years, (year) => (
                  {
                      index : year,
                      title : splitYear(year)
                  }
              ));
            `,
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
    // https://github.com/eslint/eslint/issues/14633
    {
      code: 'const createMarker = (color) => ({ latitude, longitude }, index) => {};',
      options: ['always'] as any,
      errors: [{ messageId: 'expectedBlock' }],
    },
  ],
});
