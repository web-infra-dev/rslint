import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string) => ({ code });
const invalid = (code: string) => ({ code, errors: 1 });

ruleTester.run('prefer-array-flat-map', null as never, {
  valid: [
    valid('const bar = [1,2,3].map()'),
    valid('const bar = [1,2,3].map(i => i)'),
    valid('const bar = [1,2,3].map((i) => i)'),
    valid('const bar = [1,2,3].map((i) => { return i; })'),
    valid('const bar = foo.map(i => i)'),
    valid('const bar = foo.map?.(i => [i]).flat()'),
    valid('const bar = foo.map(i => [i])?.flat()'),
    valid('const bar = foo.map(i => [i]).flat?.()'),
    valid('const bar = [[1],[2],[3]].flat()'),
    valid('const bar = [1,2,3].map(i => [i]).sort().flat()'),
    valid('let bar = [1,2,3].map(i => [i]);\n' + 'bar = bar.flat();'),
    valid('const bar = [[1],[2],[3]].map(i => [i]).flat(2)'),
    valid('const bar = [[1],[2],[3]].map(i => [i]).flat(1, null)'),
    valid('const bar = [[1],[2],[3]].map(i => [i]).flat(Infinity)'),
    valid(
      'const bar = [[1],[2],[3]].map(i => [i]).flat(Number.POSITIVE_INFINITY)',
    ),
    valid('const bar = [[1],[2],[3]].map(i => [i]).flat(Number.MAX_VALUE)'),
    valid(
      'const bar = [[1],[2],[3]].map(i => [i]).flat(Number.MAX_SAFE_INTEGER)',
    ),
    valid('const bar = [[1],[2],[3]].map(i => [i]).flat(...[1])'),
    valid('const bar = [[1],[2],[3]].map(i => [i]).flat(0.4 +.6)'),
    valid('const bar = [[1],[2],[3]].map(i => [i]).flat(+1)'),
    valid('const bar = [[1],[2],[3]].map(i => [i]).flat(foo)'),
    valid('const bar = [[1],[2],[3]].map(i => [i]).flat(foo.bar)'),
    valid('const bar = [[1],[2],[3]].map(i => [i]).flat(1.00)'),
    valid('Children.map(children, fn).flat()'),
    valid('React.Children.map(children, fn).flat()'),
  ],
  invalid: [
    invalid('const bar = [[1],[2],[3]].map(i => [i]).flat()'),
    invalid('const bar = [[1],[2],[3]].map(i => [i]).flat(1,)'),
    invalid('const bar = [1,2,3].map(i => [i]).flat()'),
    invalid('const bar = [1,2,3].map((i) => [i]).flat()'),
    invalid('const bar = [1,2,3].map((i) => { return [i]; }).flat()'),
    invalid('const bar = [1,2,3].map(foo).flat()'),
    invalid('const bar = foo.map(i => [i]).flat()'),
    invalid('const bar = foo?.map(i => [i]).flat()'),
    invalid('const bar = { map: () => {} }.map(i => [i]).flat()'),
    invalid('const bar = [1,2,3].map(i => i).map(i => [i]).flat()'),
    invalid('const bar = [1,2,3].sort().map(i => [i]).flat()'),
    invalid('const bar = (([1,2,3].map(i => [i]))).flat()'),
    invalid(
      'let bar = [1,2,3].map(i => {\n' + '\treturn [i];\n' + '}).flat();',
    ),
    invalid(
      'let bar = [1,2,3].map(i => {\n' +
        '\treturn [i];\n' +
        '})\n' +
        '.flat();',
    ),
    invalid(
      'let bar = [1,2,3].map(i => {\n' +
        '\treturn [i];\n' +
        '}) // comment\n' +
        '.flat();',
    ),
    invalid(
      'let bar = [1,2,3].map(i => {\n' +
        '\treturn [i];\n' +
        '}) // comment\n' +
        '.flat(); // other',
    ),
    invalid(
      'let bar = [1,2,3]\n' + '\t.map(i => { return [i]; })\n' + '\t.flat();',
    ),
    invalid('let bar = [1,2,3].map(i => { return [i]; })\n' + '\t.flat();'),
    invalid('let bar = [1,2,3] . map( x => y ) . flat () // 🤪'),
    invalid('const bar = [1,2,3].map(i => [i]).flat(1);'),
    invalid(
      'const foo = bars\n' +
        '\t.filter(foo => !!foo.zaz)\n' +
        '\t.map(foo => doFoo(foo))\n' +
        '\t.flat();',
    ),
  ],
});
