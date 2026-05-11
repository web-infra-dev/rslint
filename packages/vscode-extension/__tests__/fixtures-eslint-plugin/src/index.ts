// The identifier `forbidden` triggers the fixture plugin's
// `fx/no-forbidden` rule (defined in ../plugin.mjs). The other
// identifier (`safe`) is here as a negative control — if it ever
// flags, something in the worker is firing on identifiers it
// shouldn't see.
//
// `BANNED` triggers the `fx/rename-banned` rule which carries an
// autofix to rewrite it to `ALLOWED`. Used by the U10 fixAll test.
const forbidden = 1;
const safe = 2;
const BANNED = 3;

export { forbidden, safe, BANNED };
