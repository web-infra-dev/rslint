import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectNoSpreadingError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic('noSpreading', 'Prop spreading is forbidden', range);
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-props-no-spreading', {
  valid: [
    { code: '<App one_prop={one_prop} two_prop={two_prop} />' },
    { code: '<div one_prop={one_prop} two_prop={two_prop}></div>' },
    {
      code: 'const newProps = {...props}; <App one_prop={newProps.one_prop} two_prop={newProps.two_prop} style={{...styles}} />',
    },
    {
      code: '<App><Image {...props} /><img {...props} /></App>',
      options: [{ exceptions: ['Image', 'img'] }],
    },
    {
      code: '<App><Image {...props} /><img src={src} alt={alt} /></App>',
      options: [{ custom: 'ignore' }],
    },
    {
      code: '<App><Image {...props} /><img {...props} /></App>',
      options: [{ custom: 'enforce', html: 'ignore', exceptions: ['Image'] }],
    },
    {
      code: '<App><img {...props} /><Image src={src} alt={alt} /><div {...someOtherProps} /></App>',
      options: [{ html: 'ignore' }],
    },
    {
      code: '<App><Foo {...{ prop1, prop2, prop3 }} /></App>',
      options: [{ explicitSpread: 'ignore' }],
    },
    {
      code: '<App><components.Group {...props} /><Nav.Item {...props} /></App>',
      options: [{ exceptions: ['components.Group', 'Nav.Item'] }],
    },
    {
      code: '<App><components.Group {...props} /><Nav.Item {...props} /></App>',
      options: [{ custom: 'ignore' }],
    },
    {
      code: '<App><components.Group {...props} /><Nav.Item {...props} /></App>',
      options: [
        {
          custom: 'enforce',
          html: 'ignore',
          exceptions: ['components.Group', 'Nav.Item'],
        },
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        const {one_prop, two_prop} = props;
        <App one_prop={one_prop} two_prop={two_prop}/>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        const {one_prop, two_prop} = props;
        <div one_prop={one_prop} two_prop={two_prop}></div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        const newProps = {...props};
        <App one_prop={newProps.one_prop} two_prop={newProps.two_prop} style={{...styles}}/>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        const props = {src: "dummy.jpg", alt: "dummy"};
        <App>
           <Image {...props}/>
           <img {...props}/>
        </App>
      `,
      options: [{ exceptions: ['Image', 'img'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        const props = {src: "dummy.jpg", alt: "dummy"};
        const { src, alt } = props;
        <App>
           <Image {...props}/>
           <img src={src} alt={alt}/>
        </App>
      `,
      options: [{ custom: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        const props = {src: "dummy.jpg", alt: "dummy"};
        const { src, alt } = props;
        <App>
           <Image {...props}/>
           <img {...props}/>
        </App>
      `,
      options: [{ custom: 'enforce', html: 'ignore', exceptions: ['Image'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        const props = {src: "dummy.jpg", alt: "dummy"};
        const { src, alt } = props;
        <App>
           <img {...props}/>
           <Image src={src} alt={alt}/>
           <div {...someOtherProps}/>
        </App>
      `,
      options: [{ html: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        <App>
          <Foo {...{ prop1, prop2, prop3 }} />
        </App>
      `,
      options: [{ explicitSpread: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        const props = {};
        <App>
           <components.Group {...props}/>
           <Nav.Item {...props}/>
        </App>
      `,
      options: [{ exceptions: ['components.Group', 'Nav.Item'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        const props = {};
        <App>
           <components.Group {...props}/>
           <Nav.Item {...props}/>
        </App>
      `,
      options: [{ custom: 'ignore' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        const props = {};
        <App>
           <components.Group {...props}/>
           <Nav.Item {...props}/>
        </App>
      `,
      options: [
        {
          custom: 'enforce',
          html: 'ignore',
          exceptions: ['components.Group', 'Nav.Item'],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: '<App {...props} />',
      errors: [expectNoSpreadingError([1, 6, 1, 16])],
    },
    {
      code: '<div {...props}></div>',
      errors: [expectNoSpreadingError([1, 6, 1, 16])],
    },
    {
      code: '<App {...props} some_other_prop={some_other_prop} />',
      errors: [expectNoSpreadingError([1, 6, 1, 16])],
    },
    {
      code: '<App><Image {...props} /><span {...props} /></App>',
      options: [{ exceptions: ['Image', 'img'] }],
      errors: [expectNoSpreadingError([1, 32, 1, 42])],
    },
    {
      code: '<App><Image {...props} /><img {...props} /></App>',
      options: [{ custom: 'ignore' }],
      errors: [expectNoSpreadingError([1, 31, 1, 41])],
    },
    {
      code: '<App><Image {...props} /><img {...props} /></App>',
      options: [{ html: 'ignore', exceptions: ['Image', 'img'] }],
      errors: [expectNoSpreadingError([1, 31, 1, 41])],
    },
    {
      code: '<App><Image {...props} /><img {...props} /><div {...props} /></App>',
      options: [
        { custom: 'ignore', html: 'ignore', exceptions: ['Image', 'img'] },
      ],
      errors: [
        expectNoSpreadingError([1, 13, 1, 23]),
        expectNoSpreadingError([1, 31, 1, 41]),
      ],
    },
    {
      code: '<App><img {...props} /><Image {...props} /></App>',
      options: [{ html: 'ignore' }],
      errors: [expectNoSpreadingError([1, 31, 1, 41])],
    },
    {
      code: '<App><Foo {...{ prop1, prop2, prop3 }} /></App>',
      errors: [expectNoSpreadingError([1, 11, 1, 39])],
    },
    {
      code: '<App><Foo {...{ prop1, ...rest }} /></App>',
      options: [{ explicitSpread: 'ignore' }],
      errors: [expectNoSpreadingError([1, 11, 1, 34])],
    },
    {
      code: '<App><Foo {...{ ...props }} /></App>',
      options: [{ explicitSpread: 'ignore' }],
      errors: [expectNoSpreadingError([1, 11, 1, 28])],
    },
    {
      code: '<App><Foo {...props} /></App>',
      options: [{ explicitSpread: 'ignore' }],
      errors: [expectNoSpreadingError([1, 11, 1, 21])],
    },
    {
      code: '<App><components.Group {...props} /><Nav.Item {...props} /></App>',
      options: [{ exceptions: ['components.DropdownIndicator', 'Nav.Item'] }],
      errors: [expectNoSpreadingError([1, 24, 1, 34])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        <App {...props}/>
      `,
      errors: [expectNoSpreadingError([2, 14, 2, 24])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        <div {...props}></div>
      `,
      errors: [expectNoSpreadingError([2, 14, 2, 24])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        <App {...props} some_other_prop={some_other_prop}/>
      `,
      errors: [expectNoSpreadingError([2, 14, 2, 24])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        const props = {src: "dummy.jpg", alt: "dummy"};
        <App>
           <Image {...props}/>
           <span {...props}/>
        </App>
      `,
      options: [{ exceptions: ['Image', 'img'] }],
      errors: [expectNoSpreadingError([5, 18, 5, 28])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        const props = {src: "dummy.jpg", alt: "dummy"};
        const { src, alt } = props;
        <App>
           <Image {...props}/>
           <img {...props}/>
        </App>
      `,
      options: [{ custom: 'ignore' }],
      errors: [expectNoSpreadingError([6, 17, 6, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        const props = {src: "dummy.jpg", alt: "dummy"};
        const { src, alt } = props;
        <App>
           <Image {...props}/>
           <img {...props}/>
        </App>
      `,
      options: [{ html: 'ignore', exceptions: ['Image', 'img'] }],
      errors: [expectNoSpreadingError([6, 17, 6, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        const props = {src: "dummy.jpg", alt: "dummy"};
        const { src, alt } = props;
        <App>
           <Image {...props}/>
           <img {...props}/>
           <div {...props}/>
        </App>
      `,
      options: [
        { custom: 'ignore', html: 'ignore', exceptions: ['Image', 'img'] },
      ],
      errors: [
        expectNoSpreadingError([5, 19, 5, 29]),
        expectNoSpreadingError([6, 17, 6, 27]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        const props = {src: "dummy.jpg", alt: "dummy"};
        const { src, alt } = props;
        <App>
           <img {...props}/>
           <Image {...props}/>
        </App>
      `,
      options: [{ html: 'ignore' }],
      errors: [expectNoSpreadingError([6, 19, 6, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        <App>
          <Foo {...{ prop1, prop2, prop3 }} />
        </App>
      `,
      errors: [expectNoSpreadingError([3, 16, 3, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        <App>
          <Foo {...{ prop1, ...rest }} />
        </App>
      `,
      options: [{ explicitSpread: 'ignore' }],
      errors: [expectNoSpreadingError([3, 16, 3, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        <App>
          <Foo {...{ ...props }} />
        </App>
      `,
      options: [{ explicitSpread: 'ignore' }],
      errors: [expectNoSpreadingError([3, 16, 3, 33])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        <App>
          <Foo {...props } />
        </App>
      `,
      options: [{ explicitSpread: 'ignore' }],
      errors: [expectNoSpreadingError([3, 16, 3, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        const props = {};
        <App>
           <components.Group {...props}/>
           <Nav.Item {...props}/>
        </App>
      `,
      options: [{ exceptions: ['components.DropdownIndicator', 'Nav.Item'] }],
      errors: [expectNoSpreadingError([4, 30, 4, 40])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
