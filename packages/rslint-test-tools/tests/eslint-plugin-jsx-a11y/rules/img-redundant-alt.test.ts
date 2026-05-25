import { RuleTester } from '../rule-tester';

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

const errorMessage =
  'Redundant alt attribute. Screen-readers already announce `img` tags as an image. ' +
  'You don’t need to use the words `image`, `photo,` or `picture` ' +
  '(or any specified custom words) in the alt prop.';

new RuleTester().run('img-redundant-alt', null as never, {
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
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="Picture of friend." />;',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="Image of friend." />;',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="PhOtO of friend." />;',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={"photo"} />;',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="piCTUre of friend." />;',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="imAGE of friend." />;',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="photo of cool person" aria-hidden={false} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="picture of cool person" aria-hidden={false} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="image of cool person" aria-hidden={false} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="photo" {...this.props} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="image" {...this.props} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="picture" {...this.props} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={`picture doing ${things}`} {...this.props} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={`photo doing ${things}`} {...this.props} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={`image doing ${things}`} {...this.props} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={`picture doing ${picture}`} {...this.props} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={`photo doing ${photo}`} {...this.props} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={`image doing ${image}`} {...this.props} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<Image alt="Photo of a friend" />',
      settings: componentsSettings,
      errors: [{ message: errorMessage }],
    },

    // ---- Array option tests ----
    {
      code: '<img alt="Word1" />;',
      options: arrayOpts,
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="Word2" />;',
      options: arrayOpts,
      errors: [{ message: errorMessage }],
    },
    {
      code: '<Image alt="Word1" />;',
      options: arrayOpts,
      errors: [{ message: errorMessage }],
    },
    {
      code: '<Image alt="Word2" />;',
      options: arrayOpts,
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="イメージ" />',
      options: [{ words: ['イメージ'] }],
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="イメージです" />',
      options: [{ words: ['イメージ'] }],
      errors: [{ message: errorMessage }],
    },

    // ---- rslint extras ----
    {
      code: '<img alt="    photo    " />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="画像 photo" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="photo image picture" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={("photo")} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={(("photo"))} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={`photo`} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img {...{alt: "photo"}} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img {...{alt: `photo of bar`}} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<div><img alt="photo of bar" /></div>',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'function L({xs}) { return xs.map(x => <img alt="photo" key={x} />); }',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'class C { render() { return <img alt="picture" />; } }',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'function f<T>(x: T) { return <img alt="image" />; }',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'function F() { return <div>{cond && <img alt="photo" />}</div>; }',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<>{cond && <img alt="photo" />}</>',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<Box as="img" alt="photo of cat" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [{ message: errorMessage }],
    },
    {
      code: '<MyImg alt="photo of cat" />',
      settings: { 'jsx-a11y': { components: { MyImg: 'img' } } },
      errors: [{ message: errorMessage }],
    },
    // TaggedTemplate with redundant content.
    { code: '<img alt={tag`photo`} />', errors: [{ message: errorMessage }] },
    {
      code: '<img alt={tag`picture of cat`} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={tag`image doing ${x}`} />',
      errors: [{ message: errorMessage }],
    },
    // Unicode whitespace classifications — split by JS \s, redundant token.
    { code: '<img alt={"photo bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo﻿bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo　bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo bar"} />', errors: [{ message: errorMessage }] },
    { code: '<img alt={"photo bar"} />', errors: [{ message: errorMessage }] },
    // Real-world React patterns with redundancy.
    {
      code: 'const Logo = () => <img alt="Photo of brand" />;',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'const X = forwardRef((p, r) => <img ref={r} alt="photo of x" />);',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'const Y = memo(() => <img alt="photo of memo" />);',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'function H() { useEffect(() => {}); return <img alt="picture hook" />; }',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'const j = (() => <img alt="photo iife" />)();',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'const o = { render() { return <img alt="photo method" />; } };',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'function G({c}) { return c ? <img alt="photo true" /> : <img alt="picture false" />; }',
      errors: [{ message: errorMessage }, { message: errorMessage }],
    },
    {
      code: 'const T = <section><header><img alt="photo header" /></header></section>;',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<><img alt="photo" /><img alt="picture" /></>',
      errors: [{ message: errorMessage }, { message: errorMessage }],
    },
    {
      code: '<img alt="photo cat" aria-label="Cat" aria-describedby="d1" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img aria-labelledby="x" alt="photo of y" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img src="x" srcSet="x@2x" sizes="100vw" alt="photo of bar" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img onClick={f} onLoad={g} alt="photo handler" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img data-test="x" data-foo="y" alt="photo data" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img role="presentation" alt="photo role" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img {...a} {...b} alt="photo" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img {...{src: "x", alt: "photo"}} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="🐱 photo of cat" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="A NICE PHOTO TODAY" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="    photo" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="photo    " />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt="lovely photo today" />',
      errors: [{ message: errorMessage }],
    },
    // AssignmentExpression — LITERAL_TYPES inherits TYPES, synthesizes
    // "x = photo" string → fires.
    {
      code: '<img alt={x = "photo"} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={x += "image"} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<img alt={x ||= "picture"} />',
      errors: [{ message: errorMessage }],
    },
    // Parenthesized ObjectLiteral spread — ESTree folds parens, ObjectExpression matches.
    {
      code: '<img {...({alt: "photo"})} />',
      errors: [{ message: errorMessage }],
    },
  ],
});
