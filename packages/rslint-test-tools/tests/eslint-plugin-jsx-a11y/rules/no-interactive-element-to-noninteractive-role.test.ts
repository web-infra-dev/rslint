import { RuleTester } from '../rule-tester';

const errorMessage =
  'Interactive elements should not be assigned non-interactive roles.';
const expectedError = { message: errorMessage };

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Button: 'button',
      Link: 'a',
    },
  },
};

// Mirrors `configs.recommended.rules['jsx-a11y/no-interactive-element-to-noninteractive-role'][1]`.
const recommendedOptions = [
  {
    tr: ['none', 'presentation'],
    canvas: ['img'],
  },
];

new RuleTester().run(
  'no-interactive-element-to-noninteractive-role',
  null as never,
  {
    valid: [
      // ============================================================
      // Upstream alwaysValid (run under both :recommended and :strict)
      // ============================================================
      // Custom JSX components — element type isn't in aria-query's dom map.
      { code: '<TestComponent onClick={doFoo} />' },
      { code: '<Button onClick={doFoo} />' },

      // Interactive elements with an interactive role.
      { code: '<a href="http://x.y.z" role="button" />' },
      { code: '<a href="http://x.y.z" tabIndex="0" role="button" />' },
      { code: '<button className="foo" role="button" />' },

      // All flavors of input with role="button".
      { code: '<input role="button" />' },
      { code: '<input type="button" role="button" />' },
      { code: '<input type="checkbox" role="button" />' },
      { code: '<input type="color" role="button" />' },
      { code: '<input type="date" role="button" />' },
      { code: '<input type="datetime" role="button" />' },
      { code: '<input type="datetime-local" role="button" />' },
      { code: '<input type="email" role="button" />' },
      { code: '<input type="file" role="button" />' },
      { code: '<input type="image" role="button" />' },
      { code: '<input type="month" role="button" />' },
      { code: '<input type="number" role="button" />' },
      { code: '<input type="password" role="button" />' },
      { code: '<input type="radio" role="button" />' },
      { code: '<input type="range" role="button" />' },
      { code: '<input type="reset" role="button" />' },
      { code: '<input type="search" role="button" />' },
      { code: '<input type="submit" role="button" />' },
      { code: '<input type="tel" role="button" />' },
      { code: '<input type="text" role="button" />' },
      { code: '<input type="time" role="button" />' },
      { code: '<input type="url" role="button" />' },
      { code: '<input type="week" role="button" />' },
      { code: '<input type="hidden" role="button" />' },

      // Other inherently interactive controls.
      { code: '<menuitem role="button" />;' },
      { code: '<option className="foo" role="button" />' },
      { code: '<select className="foo" role="button" />' },
      { code: '<textarea className="foo" role="button" />' },
      { code: '<tr role="button" />;' },

      // HTML elements with neither interactive nor non-interactive valence.
      { code: '<a role="button" />' },
      { code: '<a role="img" />;' },
      { code: '<a tabIndex="0" role="button" />' },
      { code: '<a tabIndex="0" role="img" />' },
      { code: '<div role="button" />;' },
      { code: '<div role={undefined} role="button" />;' },
      { code: '<div role="img" />;' },
      { code: '<canvas role="button" />;' },

      // Non-DOM custom components are exempt.
      { code: '<MyButton role="img" />' },

      // Namespaced roles — propName isn't literally `role`.
      { code: '<div mynamespace:role="term" />' },
      { code: '<input mynamespace:role="img" />' },

      // Components with explicit aria mapping.
      { code: '<Link href="http://x.y.z" role="img" />' },
      { code: '<Link href="http://x.y.z" />', settings: componentsSettings },
      { code: '<Button onClick={doFoo} />', settings: componentsSettings },

      // ============================================================
      // Recommended-only valid (tr / canvas allow-list, non-DOM component)
      // ============================================================
      {
        code: '<tr role="presentation" />;',
        options: recommendedOptions,
      },
      {
        code: '<canvas role="img" />;',
        options: recommendedOptions,
      },
      {
        code: '<Component role="presentation" />;',
        options: recommendedOptions,
      },
    ],
    invalid: [
      // ============================================================
      // Upstream neverValid — invalid in both recommended and strict
      // ============================================================
      // Interactive elements + non-interactive role.
      {
        code: '<a href="http://x.y.z" role="img" />',
        errors: [expectedError],
      },
      {
        code: '<a href="http://x.y.z" tabIndex="0" role="img" />',
        errors: [expectedError],
      },

      // All flavors of input + role="img".
      { code: '<input role="img" />', errors: [expectedError] },
      { code: '<input type="img" role="img" />', errors: [expectedError] },
      { code: '<input type="checkbox" role="img" />', errors: [expectedError] },
      { code: '<input type="color" role="img" />', errors: [expectedError] },
      { code: '<input type="date" role="img" />', errors: [expectedError] },
      { code: '<input type="datetime" role="img" />', errors: [expectedError] },
      {
        code: '<input type="datetime-local" role="img" />',
        errors: [expectedError],
      },
      { code: '<input type="email" role="img" />', errors: [expectedError] },
      { code: '<input type="file" role="img" />', errors: [expectedError] },
      { code: '<input type="hidden" role="img" />', errors: [expectedError] },
      { code: '<input type="image" role="img" />', errors: [expectedError] },
      { code: '<input type="month" role="img" />', errors: [expectedError] },
      { code: '<input type="number" role="img" />', errors: [expectedError] },
      { code: '<input type="password" role="img" />', errors: [expectedError] },
      { code: '<input type="radio" role="img" />', errors: [expectedError] },
      { code: '<input type="range" role="img" />', errors: [expectedError] },
      { code: '<input type="reset" role="img" />', errors: [expectedError] },
      { code: '<input type="search" role="img" />', errors: [expectedError] },
      { code: '<input type="submit" role="img" />', errors: [expectedError] },
      { code: '<input type="tel" role="img" />', errors: [expectedError] },
      { code: '<input type="text" role="img" />', errors: [expectedError] },
      { code: '<input type="time" role="img" />', errors: [expectedError] },
      { code: '<input type="url" role="img" />', errors: [expectedError] },
      { code: '<input type="week" role="img" />', errors: [expectedError] },
      { code: '<menuitem role="img" />;', errors: [expectedError] },
      {
        code: '<option className="foo" role="img" />',
        errors: [expectedError],
      },
      {
        code: '<select className="foo" role="img" />',
        errors: [expectedError],
      },
      {
        code: '<textarea className="foo" role="img" />',
        errors: [expectedError],
      },
      { code: '<tr role="img" />;', errors: [expectedError] },

      // Interactive elements + role="listitem".
      {
        code: '<a href="http://x.y.z" role="listitem" />',
        errors: [expectedError],
      },
      {
        code: '<a href="http://x.y.z" tabIndex="0" role="listitem" />',
        errors: [expectedError],
      },
      { code: '<input role="listitem" />', errors: [expectedError] },
      {
        code: '<input type="listitem" role="listitem" />',
        errors: [expectedError],
      },
      {
        code: '<input type="checkbox" role="listitem" />',
        errors: [expectedError],
      },
      { code: '<menuitem role="listitem" />;', errors: [expectedError] },
      { code: '<tr role="listitem" />;', errors: [expectedError] },

      // Custom element resolved to <a> via settings.
      {
        code: '<Link href="http://x.y.z" role="img" />',
        settings: componentsSettings,
        errors: [expectedError],
      },

      // ============================================================
      // Strict-only invalid (no allowed-roles override for tr / canvas)
      // ============================================================
      { code: '<tr role="presentation" />;', errors: [expectedError] },
      { code: '<canvas role="img" />;', errors: [expectedError] },
    ],
  },
);
