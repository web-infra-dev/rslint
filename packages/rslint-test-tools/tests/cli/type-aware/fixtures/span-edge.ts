// P0: drives the REAL no-undefined-union rule across identifier variants where
// oxc's decoded name / range diverges from tsgo's source span. Both Go
// (snapshot.go) and the worker key identifiers on the DECODED name length
// (range[0] + name.length), so the rule HITs every variant — including an
// escaped identifier whose source token is longer than its decoded name. This
// guards the bridge's span handling end to end, not just the unit FALLBACK path.
declare function maybeString(): string | undefined;

// plain annotated binding
const annotated: string | undefined = maybeString();

// annotated, no initializer
let noInit: string | undefined = maybeString();

// annotated reference variants (annotation shape is irrelevant once keyed on the
// decoded name length, but kept to cover several real-code spellings)
const spaced: string | undefined = maybeString();
const multiLine: string | undefined = maybeString();

// escaped identifier: oxc's decoded name is "esc" (length 3) while the source
// token `\u0065sc` is 8 chars. Keyed on the decoded length on both sides, so it
// HITs like any other binding. (Once a known MISS; decoded-length keying fixed it.)
const esc: string | undefined = maybeString();

export { annotated, noInit, spaced, multiLine, esc };
