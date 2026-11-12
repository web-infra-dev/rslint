import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const arrayOpts = [
  {
    components: ['Image'],
    words: ['Word1', 'Word2'],
  },
];

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Image: 'img',
    },
  },
};

new RuleTester(config).run('img-redundant-alt', {
  valid: [
    // ---- Upstream valid cases ----
    { code: '<img alt="foo" />;' },
    {
      code: '<img alt="picture of me taking a photo of an image" aria-hidden />',
    },
    { code: '<img aria-hidden alt="photo of image" />' },
    { code: '<img ALt="foo" />;' },
    { code: '<img {...this.props} alt="foo" />' },
    { code: '<img {...this.props} alt={"foo"} />' },
    { code: '<img {...this.props} alt={alt} />' },
    { code: '<a />' },
    { code: '<img />' },
    { code: '<IMG />' },
    { code: '<img alt={undefined} />' },
    { code: '<img alt={`this should pass for ${now}`} />' },
    { code: '<img alt={`this should pass for ${photo}`} />' },
    { code: '<img alt={`this should pass for ${image}`} />' },
    { code: '<img alt={`this should pass for ${picture}`} />' },
    { code: '<img alt={`${photo}`} />' },
    { code: '<img alt={`${image}`} />' },
    { code: '<img alt={`${picture}`} />' },
    { code: '<img alt={"undefined"} />' },
    { code: '<img alt={() => {}} />' },
    { code: '<img alt={function(e){}} />' },
    { code: '<img aria-hidden={false} alt="Doing cool things." />' },
    { code: '<UX.Layout>test</UX.Layout>' },
    { code: '<img alt />' },
    { code: '<img alt={imageAlt} />' },
    { code: '<img alt={imageAlt.name} />' },
    { code: '<img alt={imageAlt?.name} />' },
    { code: '<img alt="Doing cool things" aria-hidden={foo?.bar}/>' },
    { code: '<img alt="Photography" />;' },
    { code: '<img alt="ImageMagick" />;' },
    { code: '<Image alt="Photo of a friend" />' },
    { code: '<Image alt="Foo" />', settings: componentsSettings },
    {
      code: '<img alt="画像" />',
      options: [{ words: ['イメージ'] }],
    },

    // ---- rslint extras: edge-shape lock-ins ----
    // TS wrappers (`as`, `as const`, `!`, `satisfies`) — jsx-ast-utils'
    // LITERAL_TYPES treats these as noop → null, so the rule skips
    // regardless of whether the wrapped value would be redundant.
    { code: '<img alt={"foo" as string} />' },
    { code: '<img alt={"bar" as const} />' },
    { code: '<img alt={"baz"!} />' },
    { code: '<img alt={"photo" as string} />' },
    { code: '<img alt={"picture" as const} />' },
    { code: '<img alt={"image"!} />' },
    { code: '<img alt={"photo" satisfies string} />' },
    { code: '<img alt={("photo" as any)!} />' },
    { code: '<img alt={("foo")} />' },
    { code: '<img alt={(("foo"))} />' },
    { code: '<img alt="" />' },
    { code: '<img alt=" " />' },
    { code: '<img alt={true} />' },
    { code: '<img alt={false} />' },
    { code: '<img alt={42} />' },
    { code: '<img alt={null} />' },
    { code: '<img alt="true" />' },
    { code: '<img alt="false" />' },
    { code: '<img alt="foo"/>' },
    { code: '<img alt="foo"></img>' },
    { code: '<img alt="image-thing" />' },
    { code: '<img alt="photo,picture" />' },
    { code: '<img alt="photo.jpg" />' },
    { code: '<Foo.Image alt="photo" />' },
    { code: '<foo.img alt="photo" />' },
    { code: '<svg:image alt="photo" />' },
    { code: '<my-img alt="photo" />' },
    { code: '<img {...props} />' },
    { code: '<img {...{alt: "foo"}} />' },
    { code: '<img {...{alt: `foo bar`}} />' },
    { code: '<img aria-hidden="false" alt="foo" />' },
    { code: '<img aria-hidden="yes" alt="foo" />' },
    { code: '<img /* leading */ alt="foo" /* trailing */ />' },
    { code: 'function F() { return <div>{cond && <img alt="foo" />}</div>; }' },
    {
      code: 'function L({xs}) { return xs.map(x => <img alt="foo" key={x} />); }',
    },
    { code: '<>{cond && <img alt="foo" />}</>' },
    { code: '<img aria-hidden alt="photo of friend" />' },
    { code: '<img aria-hidden={true} alt="photo of friend" />' },
    { code: '<img aria-hidden="TRUE" alt="photo of friend" />' },
    { code: '<img alt="画像" />' },
    {
      code: '<img alt="画像写真" />',
      options: [{ words: ['イメージ'] }],
    },
    { code: '<img key="x" alt="foo" />' },
    {
      code: '<img alt="foo" />',
      settings: { 'jsx-a11y': {} },
    },
    // TaggedTemplateExpression with non-redundant content — LITERAL_TYPES
    // forwards to TemplateLiteral but the inner text doesn't match.
    { code: '<img alt={tag`foo`} />' },
    { code: '<img alt={tag`hello world`} />' },
    { code: '<img alt={tag`${x}`} />' },
    // NEL (U+0085) is NOT in JS \s — single token "photo<NEL>bar".
    { code: '<img alt="photobar" />' },
    { code: '<img alt={"photobar"} />' },
    // Real-world React patterns (no redundancy).
    { code: 'const Logo = () => <img alt="Brand badge" />;' },
    { code: 'function F() { return <img alt="A friendly mascot" />; }' },
    {
      code: 'class C extends Component { render() { return <img alt="cat sleeping" />; } }',
    },
    {
      code: 'function L({xs}) { return xs.map(x => <img key={x} alt="A drawing" />); }',
    },
    { code: '<><img alt="ok" /><img alt="fine" /></>' },
    {
      code: '<img src="x" srcSet="x@2x" sizes="100vw" alt="A drawing" />',
    },
    { code: '<img onClick={f} onLoad={g} alt="ok" />' },
    { code: '<img data-test="x" data-foo="y" alt="ok" />' },
    { code: '<img role="presentation" alt="A drawing" />' },
    { code: '<img aria-labelledby="x" alt="A drawing" />' },
    {
      code: '<Image alt="photo" />',
      settings: { 'jsx-a11y': { components: { Other: 'img' } } },
    },

    // LITERAL_TYPES inheritance edge cases (differential gap fill).
    { code: '<img alt={[]} />' },
    { code: '<img alt={["photo"]} />' },
    { code: '<img alt={{x: 1}} />' },
    { code: '<img alt={<span>photo</span>} />' },
    { code: '<img alt={<>photo</>} />' },
    { code: '<img alt={/photo/} />' },
    { code: '<img alt={42n} />' },
    { code: '<img alt={new String("photo")} />' },
    { code: '<img alt={(1, "photo")} />' },
    { code: 'async function ax() { return <img alt={await x} />; }' },
    { code: 'function* gx() { return <img alt={yield "photo"} />; }' },
    // Spread argument shape strictness — TS-wrapped object literal does NOT
    // match upstream's `argument.type === 'ObjectExpression'` check.
    { code: '<img {...({alt: "photo"} as any)} />' },
    { code: '<img {...({alt: "photo"})!} />' },
    { code: '<img {...["alt", "photo"]} />' },
    { code: "<img {...{['alt']: 'photo'}} />" },
    { code: "<img {...{0: 'photo'}} />" },
    { code: '<img {...{nested: {alt: "photo"}}} />' },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: '<img alt="Photo of friend." />;',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<img alt="Picture of friend." />;',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<img alt="Image of friend." />;',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<img alt="PhOtO of friend." />;',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<img alt={"photo"} />;',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<img alt="piCTUre of friend." />;',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<img alt="imAGE of friend." />;',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<img alt="photo of cool person" aria-hidden={false} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '<img alt="picture of cool person" aria-hidden={false} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: '<img alt="image of cool person" aria-hidden={false} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '<img alt="photo" {...this.props} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<img alt="image" {...this.props} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<img alt="picture" {...this.props} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<img alt={`picture doing ${things}`} {...this.props} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },
    {
      code: '<img alt={`photo doing ${things}`} {...this.props} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: '<img alt={`image doing ${things}`} {...this.props} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: '<img alt={`picture doing ${picture}`} {...this.props} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: '<img alt={`photo doing ${photo}`} {...this.props} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '<img alt={`image doing ${image}`} {...this.props} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '<Image alt="Photo of a friend" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },

    // ---- Array option tests ----
    {
      code: '<img alt="Word1" />;',
      options: arrayOpts,
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<img alt="Word2" />;',
      options: arrayOpts,
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<Image alt="Word1" />;',
      options: arrayOpts,
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<Image alt="Word2" />;',
      options: arrayOpts,
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<img alt="イメージ" />',
      options: [{ words: ['イメージ'] }],
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<img alt="イメージです" />',
      options: [{ words: ['イメージ'] }],
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },

    // ---- rslint extras ----
    {
      code: '<img alt="    photo    " />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<img alt="画像 photo" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<img alt="photo image picture" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<img alt={("photo")} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<img alt={(("photo"))} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={`photo`} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<img {...{alt: "photo"}} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<img {...{alt: `photo of bar`}} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '<div><img alt="photo of bar" /></div>',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'function L({xs}) { return xs.map(x => <img alt="photo" key={x} />); }',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 66,
        },
      ],
    },
    {
      code: 'class C { render() { return <img alt="picture" />; } }',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'function f<T>(x: T) { return <img alt="image" />; }',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: 'function F() { return <div>{cond && <img alt="photo" />}</div>; }',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },
    {
      code: '<>{cond && <img alt="photo" />}</>',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<Box as="img" alt="photo of cat" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<MyImg alt="photo of cat" />',
      settings: { 'jsx-a11y': { components: { MyImg: 'img' } } },
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    // TaggedTemplate with redundant content.
    {
      code: '<img alt={tag`photo`} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<img alt={tag`picture of cat`} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<img alt={tag`image doing ${x}`} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    // Unicode whitespace classifications — split by JS \s, redundant token.
    {
      code: '<img alt={"photo bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={"photo﻿bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={"photo　bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={"photo bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: '<img alt={"photo bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: '<img alt={"photo bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={"photo bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={"photo bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={"photo bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={"photo bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={"photo bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={"photo bar"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    // Real-world React patterns with redundancy.
    {
      code: 'const Logo = () => <img alt="Photo of brand" />;',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: 'const X = forwardRef((p, r) => <img ref={r} alt="photo of x" />);',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 64,
        },
      ],
    },
    {
      code: 'const Y = memo(() => <img alt="photo of memo" />);',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: 'function H() { useEffect(() => {}); return <img alt="picture hook" />; }',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 70,
        },
      ],
    },
    {
      code: 'const j = (() => <img alt="photo iife" />)();',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'const o = { render() { return <img alt="photo method" />; } };',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: 'function G({c}) { return c ? <img alt="photo true" /> : <img alt="picture false" />; }',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 54,
        },
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 57,
          endLine: 1,
          endColumn: 84,
        },
      ],
    },
    {
      code: 'const T = <section><header><img alt="photo header" /></header></section>;',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: '<><img alt="photo" /><img alt="picture" /></>',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: '<img alt="photo cat" aria-label="Cat" aria-describedby="d1" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
    {
      code: '<img aria-labelledby="x" alt="photo of y" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: '<img src="x" srcSet="x@2x" sizes="100vw" alt="photo of bar" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
    {
      code: '<img onClick={f} onLoad={g} alt="photo handler" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: '<img data-test="x" data-foo="y" alt="photo data" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<img role="presentation" alt="photo role" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: '<img {...a} {...b} alt="photo" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<img {...{src: "x", alt: "photo"}} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<img alt="🐱 photo of cat" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<img alt="A NICE PHOTO TODAY" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<img alt="    photo" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<img alt="photo    " />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<img alt="lovely photo today" />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    // AssignmentExpression — LITERAL_TYPES inherits TYPES, synthesizes
    // "x = photo" string → fires.
    {
      code: '<img alt={x = "photo"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<img alt={x += "image"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<img alt={x ||= "picture"} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    // Parenthesized ObjectLiteral spread — ESTree folds parens, ObjectExpression matches.
    {
      code: '<img {...({alt: "photo"})} />',
      errors: [
        {
          messageId: 'redundantAlt',
          message:
            'Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
  ],
});
