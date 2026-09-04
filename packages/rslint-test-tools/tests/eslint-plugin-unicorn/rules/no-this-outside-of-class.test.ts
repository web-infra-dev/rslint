import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const invalid = (code: string, errorCount = 1) => ({
  code,
  errors: Array.from({ length: errorCount }, () => ({
    messageId: 'no-this-outside-of-class',
    message: 'Do not use `this` outside of classes.',
  })),
});

ruleTester.run('no-this-outside-of-class', null as never, {
  valid: [
    {
      code: `
class Foo {
  constructor() {
    this.value = 1;
  }
  method() {
    return this.value;
  }
  methodWithDefault(value = this.value) {
    return value;
  }
  static method() {
    return this.value;
  }
  get value() {
    return this._value;
  }
  set value(value) {
    this._value = value;
  }
  #method() {
    return this.value;
  }
}
`,
    },
    {
      code: `
const Foo = class {
  value = this.defaultValue;
  static value = this.defaultValue;
};
`,
    },
    {
      code: `
class Foo {
  static {
    this.value = 1;
  }
}
`,
    },
    {
      code: `
class Foo {
  method() {
    const getValue = () => this.value;
    return getValue();
  }
}
`,
    },
    {
      code: `
class Foo {
  value = () => this.defaultValue;
  static {
    const update = () => this.value = 1;
    update();
  }
}
`,
    },
    {
      code: `
class Foo {
  method() {
    class Bar {
      method() {
        return this.value;
      }
    }
    return Bar;
  }
}
`,
    },
    {
      code: `
class Foo {
  method() {
    class Bar extends this.Base {
      [this.key]() {}
    }
    return Bar;
  }
}
`,
    },
    {
      code: `
class Foo {
  method() {
    class Bar {
      [this.key] = 1;
    }
    return Bar;
  }
}
`,
    },
    {
      code: `
class Foo {
  accessor value = this.defaultValue;
}
`,
    },
    {
      code: `
const foo = {
  validator(this: TrackedModel, value: Date | null) {
    const getValue = () => this.value;
    return getValue() === value;
  },
};
`,
    },
    {
      code: `
function validator(this: TrackedModel) {
  return this.value;
}
`,
    },
    {
      code: `
const validator = function (this: TrackedModel) {
  return this.value;
};
`,
    },
  ],
  invalid: [
    invalid('this.value;'),
    invalid('const getValue = () => this.value;'),
    invalid(`
function foo() {
  return this.value;
}
`),
    invalid(`
function Foo(value) {
  this.value = value;
}
`),
    invalid(`
const foo = function () {
  return this.value;
};
`),
    invalid(`
(function () {
  this.value = 1;
})();
`),
    invalid(`
function foo() {
  const getValue = () => this.value;
  return getValue();
}
`),
    invalid(`
class Foo {
  method() {
    function getValue() {
      return this.value;
    }
    return getValue();
  }
}
`),
    invalid(`
class Foo {
  value = function () {
    return this.value;
  };
}
`),
    invalid(`
class Foo {
  static {
    function update() {
      this.value = 1;
    }
    update();
  }
}
`),
    invalid(`
const foo = {
  method() {
    return this.value;
  }
};
`),
    invalid(
      `
const foo = {
  get value() {
    return this._value;
  },
  set value(value) {
    this._value = value;
  }
};
`,
      2,
    ),
    invalid(`
const foo = {
  method: function () {
    return this.value;
  }
};
`),
    invalid(`
Foo.prototype.method = function () {
  this.value();
};
`),
    invalid(`
new SDK({
  onReady: function () {
    this.value;
  }
});
`),
    invalid(`
export default {
  methods: {
    refresh() {
      this.list = getList();
    }
  }
};
`),
    invalid(`
class Foo {
  [this.key]() {}
}
`),
    invalid(`
class Foo {
  [this.key] = 1;
}
`),
    invalid('class Foo extends this.Base {}'),
    invalid(`
class Foo {
  accessor [this.key] = 1;
}
`),
    invalid(`
const foo = {
  validator(this: TrackedModel) {
    function getValue() {
      return this.value;
    }
    return getValue();
  },
};
`),
    invalid(`
const foo = {
  method() {
    return this.value;
  },
};
`),
    invalid('const validator = (this: TrackedModel) => this.value;'),
    invalid(`
function validator(value: TrackedModel) {
  return this.value;
}
`),
  ],
});
