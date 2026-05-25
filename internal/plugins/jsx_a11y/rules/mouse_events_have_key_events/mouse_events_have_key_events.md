# mouse-events-have-key-events

## Rule Details

Enforce that hover-style mouse event handlers on HTML DOM elements are
accompanied by their keyboard-equivalent focus listeners. Coding for the
keyboard is important for users with physical disabilities, assistive
technology compatibility, and screen-reader navigation: a button or card
that only responds to `onMouseOver` is invisible to anyone who cannot
move a pointer.

By default, the rule checks two pairings:

- `onMouseOver` must be accompanied by `onFocus`
- `onMouseOut` must be accompanied by `onBlur`

For each configured hover-in handler whose value is not `null` /
`undefined`, the rule looks for an `onFocus` attribute whose value is
also not `null` / `undefined`. Same flow for hover-out handlers and
`onBlur`. The report sits on the offending mouse-handler attribute.

Custom React components (`<MyButton>`, `<Foo.Bar>`, `<svg:circle>`) are
skipped — the rule does not know which low-level element they render.

Examples of **incorrect** code for this rule:

```jsx
<div onMouseOver={() => void 0} />
<div onMouseOut={() => void 0} />
<div onMouseOver={() => void 0} onFocus={undefined} />
<div onMouseOut={() => void 0} onBlur={undefined} />
<div onMouseOver={() => void 0} {...otherProps} />
<div onMouseOut={() => void 0} {...otherProps} />
```

Examples of **correct** code for this rule:

```jsx
<div onMouseOver={() => void 0} onFocus={() => void 0} />
<div onMouseOut={() => void 0} onBlur={() => void 0} />
<div onMouseOver={() => void 0} onFocus={() => void 0} {...otherProps} />
<div onMouseOut={() => void 0} onBlur={() => void 0} {...otherProps} />
<MyComponent onMouseOver={() => void 0} />
```

## Rule Options

The rule accepts an options object with two array fields. Each field
lists the attribute names that should trigger the corresponding pairing
check. When provided, an explicit list **replaces** the defaults — the
canonical `onMouseOver` / `onMouseOut` names are not auto-included, so
add them yourself if you want both default behavior and additional
checks.

```json
{
  "jsx-a11y/mouse-events-have-key-events": [
    "error",
    {
      "hoverInHandlers": [
        "onMouseOver",
        "onMouseEnter",
        "onPointerOver",
        "onPointerEnter"
      ],
      "hoverOutHandlers": [
        "onMouseOut",
        "onMouseLeave",
        "onPointerOut",
        "onPointerLeave"
      ]
    }
  ]
}
```

Examples of **incorrect** code with the options above:

```jsx
<div onPointerEnter={() => void 0} />
<div onPointerLeave={() => void 0} />
```

Examples of **correct** code with the options above:

```jsx
<div onPointerEnter={() => void 0} onFocus={() => void 0} />
<div onPointerLeave={() => void 0} onBlur={() => void 0} />
```

An explicit empty array (`{ "hoverInHandlers": [] }`) disables the
corresponding pairing entirely.

## Resources

- [WCAG 2.1.1 — Keyboard](https://www.w3.org/WAI/WCAG21/Understanding/keyboard)

## Original Documentation

- [eslint-plugin-jsx-a11y/mouse-events-have-key-events](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/mouse-events-have-key-events.md)
