# no-selector-as-dom-name

## Rule details

DOM APIs such as `getElementById()`, `getElementsByClassName()`, and `classList`
methods accept raw IDs and class names. This rule reports arguments that start
with a CSS selector prefix.

The rule can remove a leading `#` from a simple ID passed to `getElementById()`
or a leading `.` from a simple class name passed to a class-name API. It also
fixes templates with one substitution immediately after the prefix. It reports
complex selectors, mismatched prefixes, and escaped prefixes without an edit.

Examples of **incorrect** code for this rule

```javascript
document.getElementById('#app');
document.getElementsByClassName('.active');
element.classList.add('.active');
```

Examples of **correct** code for this rule

```javascript
document.getElementById('app');
document.getElementsByClassName('active');
element.classList.add('active');
document.querySelector('#app');
```

## Options

This rule has no options.

## References

- [Unicorn rule documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-selector-as-dom-name.md)
- [Upstream implementation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-selector-as-dom-name.js)
