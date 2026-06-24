declare function maybeString(): string | undefined;
declare function makeInter(): { a: number } & { b: string };

// Each binding exercises a distinct snapshot type-block layout, so the
// report-type-shape rule reading it round-trips Go encode ↔ Node decode for
// that kind.
const unionVar = maybeString(); // union:2 (string | undefined)
const interVar = makeInter(); // intersection:2
const arrayVar: number[] = []; // array:1 (one type arg: number)
const fnVar: () => number = () => 1; // callable:1
const plainVar = 'hi'; // no shape tags → no report

export { unionVar, interVar, arrayVar, fnVar, plainVar };
