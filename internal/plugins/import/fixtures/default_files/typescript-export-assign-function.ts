// `export = function` — under esModuleInterop the function is the default.
function fn(): number {
  return 1;
}

export = fn;
