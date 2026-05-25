# no-useless-backreference

## Rule Details

This rule disallows useless backreferences in regular expressions — backreferences that can only ever match the empty string regardless of input. The rule recognizes five problematic patterns:

- A backreference to a group that contains the backreference itself (the group hasn't matched yet when the backreference starts).
- A backreference that appears before the group it refers to (forward reference).
- A backreference inside a lookbehind that refers to a group appearing before it in the same lookbehind (lookbehind matches right-to-left).
- A backreference and its referenced group are in different alternatives of the same disjunction.
- A backreference to a group inside a negative lookaround when the backreference itself is outside that lookaround (the group's match was discarded).

Examples of **incorrect** code for this rule:

```javascript
/^(?:(a)|\1b)$/;       // reference to a group in another alternative
/(?:(a)|b(?:c|\1))$/;  // reference to a group in another alternative
/\1(a)/;               // forward reference
RegExp('(a)\\2(b)');   // forward reference
/\k<foo>(?<foo>a)/;    // forward reference (named)
/(?<=(a)\1)b/;         // backward reference in lookbehind
new RegExp('(\\1)');   // nested reference
/^((a)\1)$/;           // nested reference
/a(?!(b)).\1/;         // reference into a negative lookahead
/(?<!(a))b\1/;         // reference into a negative lookbehind
```

Examples of **correct** code for this rule:

```javascript
/^(?:(a)|(b)\2)$/;
/(a)\1/;
RegExp('(a)\\1(b)');
/(?<foo>a)\k<foo>/;
/(?<=\1(a))b/;
/^(?:(a)\1)$/;
/a(?!(b|c)\1)./;
```

## Options

This rule has no options.

## Original Documentation

- [ESLint no-useless-backreference](https://eslint.org/docs/latest/rules/no-useless-backreference)
