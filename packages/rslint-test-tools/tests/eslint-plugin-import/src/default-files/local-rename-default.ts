// `export { foo as default }` — TS records `default` as an alias to `foo`.
function foo() {
  return 'foo';
}
export { foo as default };
