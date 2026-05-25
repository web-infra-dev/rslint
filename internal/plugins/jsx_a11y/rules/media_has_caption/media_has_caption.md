# media-has-caption

## Rule Details

Enforces that `<audio>` and `<video>` elements include a `<track>` child with
`kind="captions"`. Captions are essential for deaf users to follow along, and
should contain dialogue, sound effects, musical cues, and other relevant audio
information. Captions are **not** required for media elements that carry the
`muted` attribute set to `true`.

The `kind` value comparison is case-insensitive — `kind="captions"`,
`kind="Captions"`, and `kind="CAPTIONS"` all satisfy the rule. The `muted`
exemption is the strict literal `true` (boolean form `<audio muted />`,
`muted={true}`, and the `"true"` string that coerces to boolean true).
`muted={false}`, `muted={null}`, identifiers, conditionals, and other
non-literal expressions do **not** silence the rule.

Examples of **incorrect** code for this rule:

```jsx
<audio />
<video />
<audio><track /></audio>
<video><track kind="subtitles" /></video>
<audio>Foo</audio>
<audio muted={false}></audio>
```

Examples of **correct** code for this rule:

```jsx
<audio><track kind="captions" /></audio>
<video><track kind="captions" /></video>
<audio><track kind="Captions" /><track kind="subtitles" /></audio>
<audio muted></audio>
<video muted={true}></video>
```

## Rule Options

```json
{
  "jsx-a11y/media-has-caption": [
    "error",
    {
      "audio": ["Audio"],
      "video": ["Video"],
      "track": ["Track"]
    }
  ]
}
```

### `audio`

Type: `string[]`. Default: `[]`.

Additional component names that should be treated as `<audio>` for purposes
of the caption requirement. The native `audio` tag is always checked.

### `video`

Type: `string[]`. Default: `[]`.

Additional component names that should be treated as `<video>`. The native
`video` tag is always checked.

### `track`

Type: `string[]`. Default: `[]`.

Additional component names that should be treated as `<track>` when looking
for caption children. The native `track` tag is always recognized.

Examples of **correct** code with
`{ "audio": ["Audio"], "video": ["Video"], "track": ["Track"] }`:

```json
{
  "jsx-a11y/media-has-caption": [
    "error",
    { "audio": ["Audio"], "video": ["Video"], "track": ["Track"] }
  ]
}
```

```jsx
<Audio><Track kind="captions" /></Audio>
<Video><Track kind="captions" /></Video>
<Audio muted></Audio>
```

## Settings

The rule reads `settings['jsx-a11y']` to resolve the effective element type
for a JSX tag. Resolution is two-step: the polymorphic prop runs first,
then the components map looks up the (possibly already-replaced) name.

- `polymorphicPropName` — name of a polymorphic prop (e.g. `"as"`) that
  remaps the element type. With `polymorphicPropName: "as"`,
  `<Box as="audio" muted={true}>` is recognized as `<audio muted={true}>`
  and silenced. The reverse direction is supported too —
  `<audio as="div" />` resolves to `"div"` and is not checked.
- `polymorphicAllowList` — `string[]` restricting which raw component names
  may be remapped via the polymorphic prop. When omitted, every component
  may be remapped.
- `components` — a `{ ComponentName: "html-element" }` map. With
  `{ "Audio": "audio", "Track": "track" }`, `<Audio><Track kind="captions" /></Audio>`
  is recognized as `<audio><track kind="captions" /></audio>` and accepted.

These mirror the upstream `eslint-plugin-jsx-a11y` settings exactly.

## Accessibility guidelines

- [WCAG 1.2.2 — Captions (Prerecorded)](https://www.w3.org/WAI/WCAG21/Understanding/captions-prerecorded.html)
- [WCAG 1.2.3 — Audio Description or Media Alternative (Prerecorded)](https://www.w3.org/WAI/WCAG21/Understanding/audio-description-or-media-alternative-prerecorded.html)

### Resources

- [axe-core, audio-caption](https://dequeuniversity.com/rules/axe/2.1/audio-caption)
- [axe-core, video-caption](https://dequeuniversity.com/rules/axe/2.1/video-caption)

## Original Documentation

- [eslint-plugin-jsx-a11y/media-has-caption](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/media-has-caption.md)
