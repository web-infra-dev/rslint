import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

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

new RuleTester(config).run('media-has-caption', {
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
    {
      code: '<audio><track /></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio><track kind="subtitles" /></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<video><track /></video>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<video><track kind="subtitles" /></video>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<Audio muted={false}></Audio>',
      options: [customSchema],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<Video muted={false}></Video>',
      options: [customSchema],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<Audio muted={false}></Audio>',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<Video muted={false}></Video>',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<video />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<audio>Foo</audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<video>Foo</video>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<Audio />',
      options: [customSchema],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<Video />',
      options: [customSchema],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<Audio />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<Video />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<audio><Track /></audio>',
      options: [customSchema],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<video><Track /></video>',
      options: [customSchema],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<Audio><Track kind="subtitles" /></Audio>',
      options: [customSchema],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<Video><Track kind="subtitles" /></Video>',
      options: [customSchema],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<Audio><Track kind="subtitles" /></Audio>',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<Video><Track kind="subtitles" /></Video>',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<Box as="audio"><Track kind="subtitles" /></Box>',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },

    // ---- Paired form: report on the OPENING element only ----
    {
      code: '<audio>scrolling</audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },

    // ---- Same-kind / cross-kind nesting: each fires independently ----
    {
      code: '<video><audio /></video>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<audio><video><track kind="captions" /></video></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },

    // ---- Muted edge shapes — only strict literal `true` silences ----
    {
      code: '<audio muted={undefined} />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<audio muted={null} />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<audio muted={someVar} />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<audio muted={1} />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<audio muted="anything" />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<audio muted={cond ? true : false} />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<audio muted={!true} />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<audio muted={true as any} />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<audio muted={`true`} />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },

    // ---- Captions extraction edge shapes ----
    {
      code: '<audio><track kind={someVar} /></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio><track kind={cond ? "captions" : "subtitles"} /></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio><track kind="" /></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio><track kind={""} /></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio><track kind="true" /></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio><track kind={"captions" as string} /></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio><track kind /></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio><track kind="subtitles" /><track kind="metadata" /></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },

    // ---- Filter rejects JsxFragment / JsxExpression children ----
    {
      code: '<audio><><track kind="captions" /></></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<audio>{cond && <track kind="captions" />}</audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },

    // ---- Spread attributes don't carry muted / muted=false ----
    {
      code: '<audio {...props}></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<audio {...{muted: false}}></audio>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },

    // ---- Options extension via array-wrapped JSON shape ----
    {
      code: '<MyAudio />',
      options: [{ audio: ['MyAudio'] }],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '<audio><MyTrack kind="subtitles" /></audio>',
      options: [{ track: ['MyTrack'] }],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<MyAudio><MyTrack kind="subtitles" /></MyAudio>',
      options: [{ audio: ['MyAudio'], track: ['MyTrack'] }],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<MyVideo />',
      options: [{ video: ['MyVideo'] }],
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },

    // ---- Settings: components map ----
    {
      code: '<MyAudio />',
      settings: { 'jsx-a11y': { components: { MyAudio: 'audio' } } },
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '<MyVideo />',
      settings: { 'jsx-a11y': { components: { MyVideo: 'video' } } },
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '<audio />',
      settings: { 'jsx-a11y': { components: { audio: 'audio' } } },
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },

    // ---- Polymorphic without allow-list ----
    {
      code: '<Foo as="audio" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
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
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<Foo as="Bar" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          components: { Bar: 'audio' },
        },
      },
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },

    // ---- Real-world component patterns ----
    {
      code: 'function Banner() { return <video />; }',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'const x = items.map(item => <video key={item.id} />)',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: 'const x = cond ? <audio /> : <div />',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'class C { render() { return <div><audio /><video /></div>; } }',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 43,
        },
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },

    // ---- Listener boundary: wrapped in non-media ----
    {
      code: '<div><audio /></div>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<><audio /></>',
      errors: [
        {
          messageId: 'mediaHasCaption',
          message:
            'Media elements such as <audio> and <video> must have a <track> for captions.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
  ],
});
