// `export = function-with-namespace-merge` — function callable + members.
function fn(): number {
  return 1;
}
namespace fn {
  export const bar: string = 'bar';
}

export = fn;
