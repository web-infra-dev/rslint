# no-unused-private-class-members

## Rule Details

This rule reports private class members that are declared but never used.

A private field or method is considered unused if its value is never read. A
private accessor is considered unused if it is never accessed, either for a read
or a write.

Examples of **incorrect** code for this rule:

```javascript
class A {
    #unusedMember = 5;
}

class B {
    #usedOnlyInWrite = 5;
    method() {
        this.#usedOnlyInWrite = 42;
    }
}

class C {
    #usedOnlyToUpdateItself = 5;
    method() {
        this.#usedOnlyToUpdateItself++;
    }
}

class D {
    #unusedMethod() {}
}
```

Examples of **correct** code for this rule:

```javascript
class A {
    #usedMember = 42;
    method() {
        return this.#usedMember;
    }
}

class B {
    #usedMethod() {
        return 42;
    }
    anotherMethod() {
        return this.#usedMethod();
    }
}

class C {
    get #usedAccessor() {}
    set #usedAccessor(value) {}

    method() {
        this.#usedAccessor = 42;
    }
}
```

## Options

This rule has no options.

## Original Documentation

- [ESLint no-unused-private-class-members](https://eslint.org/docs/latest/rules/no-unused-private-class-members)
