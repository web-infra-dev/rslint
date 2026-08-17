# block-scoped-var

## Rule Details

The `block-scoped-var` rule generates warnings when variables are used outside of the block in which they were defined. This emulates C-style block scope.

Examples of **incorrect** code for this rule:

```javascript
function doIf() {
  if (true) {
    var build = true;
  }

  console.log(build);
}

function doIfElse() {
  if (true) {
    var build = true;
  } else {
    var build = false;
  }
}

function doTryCatch() {
  try {
    var build = 1;
  } catch (e) {
    var f = build;
  }
}

function doFor() {
  for (var x = 1; x < 10; x++) {
    var y = f(x);
  }
  console.log(y);
}

class C {
  static {
    if (something) {
      var build = true;
    }
    build = false;
  }
}
```

Examples of **correct** code for this rule:

```javascript
function doIf() {
  var build;

  if (true) {
    build = true;
  }

  console.log(build);
}

function doIfElse() {
  var build;

  if (true) {
    build = true;
  } else {
    build = false;
  }
}

function doTryCatch() {
  var build;
  var f;

  try {
    build = 1;
  } catch (e) {
    f = build;
  }
}

function doFor() {
  for (var x = 1; x < 10; x++) {
    var y = f(x);
    console.log(y);
  }
}

class C {
  static {
    var build = false;
    if (something) {
      build = true;
    }
  }
}
```

## Options

This rule has no options.

## Original Documentation

- [ESLint: block-scoped-var](https://eslint.org/docs/latest/rules/block-scoped-var)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/block-scoped-var.js)
