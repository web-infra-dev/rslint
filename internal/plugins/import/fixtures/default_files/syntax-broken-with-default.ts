// Intentional syntax errors — `let =` is malformed — but a default export
// is still parseable on its own line. Rule must not crash and must observe
// the default if TS can still bind it.
let = 1;
export default function brokenDefault() {
  return 1;
}
