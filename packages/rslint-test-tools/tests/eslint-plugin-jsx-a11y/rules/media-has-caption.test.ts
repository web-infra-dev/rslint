import { RuleTester } from '../rule-tester';

const expectedError = {
  message:
    'Media elements such as <audio> and <video> must have a <track> for captions.',
};

const customSchema = {
  audio: ['Audio'],
  video: ['Video'],
  track: ['Track'],
};

const componentsSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
    components: {
      Audio: 'audio',
      Video: 'video',
      Track: 'track',
    },
  },
};

new RuleTester().run('media-has-caption', null as never, {
  valid: [
    // ---- Upstream valid ----
    { code: '<div />;' },
    { code: '<MyDiv />;' },
    { code: '<audio><track kind="captions" /></audio>' },
    { code: '<audio><track kind="Captions" /></audio>' },
    {
      code: '<audio><track kind="Captions" /><track kind="subtitles" /></audio>',
    },
    { code: '<video><track kind="captions" /></video>' },
    { code: '<video><track kind="Captions" /></video>' },
    {
      code: '<video><track kind="Captions" /><track kind="subtitles" /></video>',
    },
    { code: '<audio muted={true}></audio>' },
    { code: '<video muted={true}></video>' },
    { code: '<video muted></video>' },

    // ---- Custom schema (`audio: ['Audio'], video: ['Video'], track: ['Track']`) ----
    {
      code: '<Audio><track kind="captions" /></Audio>',
      options: [customSchema],
    },
    {
      code: '<audio><Track kind="captions" /></audio>',
      options: [customSchema],
    },
    {
      code: '<Video><track kind="captions" /></Video>',
      options: [customSchema],
    },
    {
      code: '<video><Track kind="captions" /></video>',
      options: [customSchema],
    },
    {
      code: '<Audio><Track kind="captions" /></Audio>',
      options: [customSchema],
    },
    {
      code: '<Video><Track kind="captions" /></Video>',
      options: [customSchema],
    },
    { code: '<Video muted></Video>', options: [customSchema] },
    { code: '<Video muted={true}></Video>', options: [customSchema] },
    { code: '<Audio muted></Audio>', options: [customSchema] },
    { code: '<Audio muted={true}></Audio>', options: [customSchema] },

    // ---- componentsSettings ----
    {
      code: '<Audio><track kind="captions" /></Audio>',
      settings: componentsSettings,
    },
    {
      code: '<audio><Track kind="captions" /></audio>',
      settings: componentsSettings,
    },
    {
      code: '<Video><track kind="captions" /></Video>',
      settings: componentsSettings,
    },
    {
      code: '<video><Track kind="captions" /></video>',
      settings: componentsSettings,
    },
    {
      code: '<Audio><Track kind="captions" /></Audio>',
      settings: componentsSettings,
    },
    {
      code: '<Video><Track kind="captions" /></Video>',
      settings: componentsSettings,
    },
    { code: '<Video muted></Video>', settings: componentsSettings },
    { code: '<Video muted={true}></Video>', settings: componentsSettings },
    { code: '<Audio muted></Audio>', settings: componentsSettings },
    { code: '<Audio muted={true}></Audio>', settings: componentsSettings },
    {
      code: '<Box as="audio" muted={true}></Box>',
      settings: componentsSettings,
    },

    // ---- Case sensitivity / property-access tags don't match ----
    { code: '<AUDIO />' },
    { code: '<Audio />' },
    { code: '<UI.audio />' },
    { code: '<svg:audio />' },

    // ---- Element kind survey: rule no-op for non-media tags ----
    { code: '<a />' },
    { code: '<input />' },
    { code: '<Component><track kind="captions" /></Component>' },
    { code: '<div><track kind="subtitles" /></div>' },
    { code: '<track />' },

    // ---- Muted exemption variants ----
    { code: '<audio muted />' },
    { code: '<audio muted={(true)}></audio>' },
    { code: '<audio muted="true"></audio>' },
    { code: '<audio muted={"true"}></audio>' },
    { code: '<audio {...{muted: true}}></audio>' },
    { code: '<audio muted={!false} />' },

    // ---- Captions extraction variants ----
    { code: '<audio><track kind={"captions"} /></audio>' },
    { code: '<audio><track kind={("captions")} /></audio>' },
    { code: '<audio><track kind="CAPTIONS" /></audio>' },
    { code: '<audio><track kind="cAPtIoNS" /></audio>' },
    {
      code: '<audio><track kind="subtitles" /><track kind="captions" /></audio>',
    },
    { code: '<audio>Hello <track kind="captions" /></audio>' },
    { code: '<audio>{header}<track kind="captions" /></audio>' },
    { code: '<audio><source src="a.mp3" /><track kind="captions" /></audio>' },
    { code: '<video><track kind="captions"></track></video>' },

    // ---- Settings: polymorphic / components reverse exemption ----
    {
      code: '<Audio />',
      settings: {
        'jsx-a11y': { components: { OtherTag: 'audio' } },
      },
    },
    {
      code: '<Foo as="audio" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['Bar'],
        },
      },
    },
    {
      code: '<audio />',
      settings: { 'jsx-a11y': { components: { audio: 'div' } } },
    },
    {
      code: '<audio as="div" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },

    // ---- Defensive type handling ----
    {
      code: '<MyAudio />',
      settings: { 'jsx-a11y': 'invalid' as unknown as object },
    },
    {
      code: '<MyAudio />',
      settings: { 'jsx-a11y': null as unknown as object },
    },
  ],
  invalid: [
    // ---- Upstream invalid ----
    { code: '<audio><track /></audio>', errors: [expectedError] },
    {
      code: '<audio><track kind="subtitles" /></audio>',
      errors: [expectedError],
    },
    { code: '<audio />', errors: [expectedError] },
    { code: '<video><track /></video>', errors: [expectedError] },
    {
      code: '<video><track kind="subtitles" /></video>',
      errors: [expectedError],
    },
    {
      code: '<Audio muted={false}></Audio>',
      options: [customSchema],
      errors: [expectedError],
    },
    {
      code: '<Video muted={false}></Video>',
      options: [customSchema],
      errors: [expectedError],
    },
    {
      code: '<Audio muted={false}></Audio>',
      settings: componentsSettings,
      errors: [expectedError],
    },
    {
      code: '<Video muted={false}></Video>',
      settings: componentsSettings,
      errors: [expectedError],
    },
    { code: '<video />', errors: [expectedError] },
    { code: '<audio>Foo</audio>', errors: [expectedError] },
    { code: '<video>Foo</video>', errors: [expectedError] },
    { code: '<Audio />', options: [customSchema], errors: [expectedError] },
    { code: '<Video />', options: [customSchema], errors: [expectedError] },
    {
      code: '<Audio />',
      settings: componentsSettings,
      errors: [expectedError],
    },
    {
      code: '<Video />',
      settings: componentsSettings,
      errors: [expectedError],
    },
    {
      code: '<audio><Track /></audio>',
      options: [customSchema],
      errors: [expectedError],
    },
    {
      code: '<video><Track /></video>',
      options: [customSchema],
      errors: [expectedError],
    },
    {
      code: '<Audio><Track kind="subtitles" /></Audio>',
      options: [customSchema],
      errors: [expectedError],
    },
    {
      code: '<Video><Track kind="subtitles" /></Video>',
      options: [customSchema],
      errors: [expectedError],
    },
    {
      code: '<Audio><Track kind="subtitles" /></Audio>',
      settings: componentsSettings,
      errors: [expectedError],
    },
    {
      code: '<Video><Track kind="subtitles" /></Video>',
      settings: componentsSettings,
      errors: [expectedError],
    },
    {
      code: '<Box as="audio"><Track kind="subtitles" /></Box>',
      settings: componentsSettings,
      errors: [expectedError],
    },

    // ---- Paired form: report on the OPENING element only ----
    { code: '<audio>scrolling</audio>', errors: [expectedError] },
    { code: '<audio></audio>', errors: [expectedError] },

    // ---- Same-kind / cross-kind nesting: each fires independently ----
    {
      code: '<video><audio /></video>',
      errors: [expectedError, expectedError],
    },
    {
      code: '<audio><video><track kind="captions" /></video></audio>',
      errors: [expectedError],
    },

    // ---- Muted edge shapes — only strict literal `true` silences ----
    { code: '<audio muted={undefined} />', errors: [expectedError] },
    { code: '<audio muted={null} />', errors: [expectedError] },
    { code: '<audio muted={someVar} />', errors: [expectedError] },
    { code: '<audio muted={1} />', errors: [expectedError] },
    { code: '<audio muted="anything" />', errors: [expectedError] },
    { code: '<audio muted={cond ? true : false} />', errors: [expectedError] },
    { code: '<audio muted={!true} />', errors: [expectedError] },
    { code: '<audio muted={true as any} />', errors: [expectedError] },
    { code: '<audio muted={`true`} />', errors: [expectedError] },

    // ---- Captions extraction edge shapes ----
    {
      code: '<audio><track kind={someVar} /></audio>',
      errors: [expectedError],
    },
    {
      code: '<audio><track kind={cond ? "captions" : "subtitles"} /></audio>',
      errors: [expectedError],
    },
    { code: '<audio><track kind="" /></audio>', errors: [expectedError] },
    { code: '<audio><track kind={""} /></audio>', errors: [expectedError] },
    { code: '<audio><track kind="true" /></audio>', errors: [expectedError] },
    {
      code: '<audio><track kind={"captions" as string} /></audio>',
      errors: [expectedError],
    },
    { code: '<audio><track kind /></audio>', errors: [expectedError] },
    {
      code: '<audio><track kind="subtitles" /><track kind="metadata" /></audio>',
      errors: [expectedError],
    },

    // ---- Filter rejects JsxFragment / JsxExpression children ----
    {
      code: '<audio><><track kind="captions" /></></audio>',
      errors: [expectedError],
    },
    {
      code: '<audio>{cond && <track kind="captions" />}</audio>',
      errors: [expectedError],
    },

    // ---- Spread attributes don't carry muted / muted=false ----
    { code: '<audio {...props}></audio>', errors: [expectedError] },
    { code: '<audio {...{muted: false}}></audio>', errors: [expectedError] },

    // ---- Options extension via array-wrapped JSON shape ----
    {
      code: '<MyAudio />',
      options: [{ audio: ['MyAudio'] }],
      errors: [expectedError],
    },
    {
      code: '<audio><MyTrack kind="subtitles" /></audio>',
      options: [{ track: ['MyTrack'] }],
      errors: [expectedError],
    },
    {
      code: '<MyAudio><MyTrack kind="subtitles" /></MyAudio>',
      options: [{ audio: ['MyAudio'], track: ['MyTrack'] }],
      errors: [expectedError],
    },
    {
      code: '<MyVideo />',
      options: [{ video: ['MyVideo'] }],
      errors: [expectedError],
    },

    // ---- Settings: components map ----
    {
      code: '<MyAudio />',
      settings: { 'jsx-a11y': { components: { MyAudio: 'audio' } } },
      errors: [expectedError],
    },
    {
      code: '<MyVideo />',
      settings: { 'jsx-a11y': { components: { MyVideo: 'video' } } },
      errors: [expectedError],
    },
    {
      code: '<audio />',
      settings: { 'jsx-a11y': { components: { audio: 'audio' } } },
      errors: [expectedError],
    },

    // ---- Polymorphic without allow-list ----
    {
      code: '<Foo as="audio" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [expectedError],
    },

    // ---- Polymorphic + components combo / chain ----
    {
      code: '<Foo as="audio" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          components: { Foo: 'div' },
        },
      },
      errors: [expectedError],
    },
    {
      code: '<Foo as="Bar" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          components: { Bar: 'audio' },
        },
      },
      errors: [expectedError],
    },

    // ---- Real-world component patterns ----
    {
      code: 'function Banner() { return <video />; }',
      errors: [expectedError],
    },
    {
      code: 'const x = items.map(item => <video key={item.id} />)',
      errors: [expectedError],
    },
    {
      code: 'const x = cond ? <audio /> : <div />',
      errors: [expectedError],
    },
    {
      code: 'class C { render() { return <div><audio /><video /></div>; } }',
      errors: [expectedError, expectedError],
    },

    // ---- Listener boundary: wrapped in non-media ----
    { code: '<div><audio /></div>', errors: [expectedError] },
    { code: '<><audio /></>', errors: [expectedError] },
  ],
});
