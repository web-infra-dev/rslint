package media_has_caption

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError mirrors upstream's `expectedError` constant — single message,
// no positional / per-rule data; the message is hard-coded and reported on
// the opening element.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "mediaHasCaption",
	Message:   "Media elements such as <audio> and <video> must have a <track> for captions.",
}

// customSchema mirrors upstream's `customSchema` array — adds capital-A
// `Audio`, capital-V `Video`, capital-T `Track` as additional component names
// to the corresponding native element pools. Exercises the options-schema
// path of `isMediaType` / `isTrackType`.
var customSchema = map[string]interface{}{
	"audio": []interface{}{"Audio"},
	"video": []interface{}{"Video"},
	"track": []interface{}{"Track"},
}

// componentsSettings mirrors upstream's `componentsSettings` constant — the
// jsx-a11y settings shape that aliases capital-A `Audio` / `Video` / `Track`
// React components onto their lowercase intrinsic equivalents AND enables
// the `as`-prop polymorphic resolution. Exercises both
// `GetElementType.components` and `GetElementType.polymorphicPropName` paths
// in the same setting block.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
		"components": map[string]interface{}{
			"Audio": "audio",
			"Video": "video",
			"Track": "track",
		},
	},
}

// TestMediaHasCaptionUpstream covers the full valid / invalid suite migrated
// 1:1 from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/media-has-caption-test.js`. Order and grouping mirror
// the upstream file so a future audit can grep across both side-by-side.
//
// rslint-specific lock-ins (Dimension 1-4 edge shapes, options coverage,
// position assertions, polymorphic prop, listener boundary) live in
// media_has_caption_extras_test.go.
func TestMediaHasCaptionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MediaHasCaptionRule, []rule_tester.ValidTestCase{
		// ---- Non-media elements never trigger ----
		{Code: `<div />;`, Tsx: true},
		{Code: `<MyDiv />;`, Tsx: true},

		// ---- Native audio with captions (case-insensitive kind match) ----
		{Code: `<audio><track kind="captions" /></audio>`, Tsx: true},
		{Code: `<audio><track kind="Captions" /></audio>`, Tsx: true},
		{Code: `<audio><track kind="Captions" /><track kind="subtitles" /></audio>`, Tsx: true},

		// ---- Native video with captions ----
		{Code: `<video><track kind="captions" /></video>`, Tsx: true},
		{Code: `<video><track kind="Captions" /></video>`, Tsx: true},
		{Code: `<video><track kind="Captions" /><track kind="subtitles" /></video>`, Tsx: true},

		// ---- Muted exemption: explicit boolean and shorthand ----
		{Code: `<audio muted={true}></audio>`, Tsx: true},
		{Code: `<video muted={true}></video>`, Tsx: true},
		{Code: `<video muted></video>`, Tsx: true},

		// ---- Custom schema (`audio: ['Audio'], video: ['Video'], track: ['Track']`) ----
		{Code: `<Audio><track kind="captions" /></Audio>`, Tsx: true, Options: customSchema},
		{Code: `<audio><Track kind="captions" /></audio>`, Tsx: true, Options: customSchema},
		{Code: `<Video><track kind="captions" /></Video>`, Tsx: true, Options: customSchema},
		{Code: `<video><Track kind="captions" /></video>`, Tsx: true, Options: customSchema},
		{Code: `<Audio><Track kind="captions" /></Audio>`, Tsx: true, Options: customSchema},
		{Code: `<Video><Track kind="captions" /></Video>`, Tsx: true, Options: customSchema},
		{Code: `<Video muted></Video>`, Tsx: true, Options: customSchema},
		{Code: `<Video muted={true}></Video>`, Tsx: true, Options: customSchema},
		{Code: `<Audio muted></Audio>`, Tsx: true, Options: customSchema},
		{Code: `<Audio muted={true}></Audio>`, Tsx: true, Options: customSchema},

		// ---- componentsSettings (polymorphic + components map) ----
		{Code: `<Audio><track kind="captions" /></Audio>`, Tsx: true, Settings: componentsSettings},
		{Code: `<audio><Track kind="captions" /></audio>`, Tsx: true, Settings: componentsSettings},
		{Code: `<Video><track kind="captions" /></Video>`, Tsx: true, Settings: componentsSettings},
		{Code: `<video><Track kind="captions" /></video>`, Tsx: true, Settings: componentsSettings},
		{Code: `<Audio><Track kind="captions" /></Audio>`, Tsx: true, Settings: componentsSettings},
		{Code: `<Video><Track kind="captions" /></Video>`, Tsx: true, Settings: componentsSettings},
		{Code: `<Video muted></Video>`, Tsx: true, Settings: componentsSettings},
		{Code: `<Video muted={true}></Video>`, Tsx: true, Settings: componentsSettings},
		{Code: `<Audio muted></Audio>`, Tsx: true, Settings: componentsSettings},
		{Code: `<Audio muted={true}></Audio>`, Tsx: true, Settings: componentsSettings},

		// ---- Polymorphic-prop exemption: `<Box as="audio" muted={true}>` ----
		// `as` resolves Box → audio, then muted={true} silences the rule.
		{Code: `<Box as="audio" muted={true}></Box>`, Tsx: true, Settings: componentsSettings},
	}, []rule_tester.InvalidTestCase{
		// ---- Native audio missing captions ----
		{Code: `<audio><track /></audio>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<audio><track kind="subtitles" /></audio>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<audio />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Native video missing captions ----
		{Code: `<video><track /></video>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<video><track kind="subtitles" /></video>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Muted explicitly set to false (NOT exempt) ----
		{Code: `<Audio muted={false}></Audio>`, Tsx: true, Options: customSchema, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Video muted={false}></Video>`, Tsx: true, Options: customSchema, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Audio muted={false}></Audio>`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Video muted={false}></Video>`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Self-closing video with no captions ----
		{Code: `<video />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Text content children don't satisfy the track requirement ----
		{Code: `<audio>Foo</audio>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<video>Foo</video>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Custom-schema components without captions ----
		{Code: `<Audio />`, Tsx: true, Options: customSchema, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Video />`, Tsx: true, Options: customSchema, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Audio />`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Video />`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Track child without `kind` ----
		{Code: `<audio><Track /></audio>`, Tsx: true, Options: customSchema, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<video><Track /></video>`, Tsx: true, Options: customSchema, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Track child with kind != captions ----
		{Code: `<Audio><Track kind="subtitles" /></Audio>`, Tsx: true, Options: customSchema, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Video><Track kind="subtitles" /></Video>`, Tsx: true, Options: customSchema, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Audio><Track kind="subtitles" /></Audio>`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Video><Track kind="subtitles" /></Video>`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Polymorphic + components: `<Box as="audio">` resolves to audio,
		//      Track has kind="subtitles" so no caption ----
		{Code: `<Box as="audio"><Track kind="subtitles" /></Box>`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	})
}
