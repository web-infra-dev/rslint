import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-props-no-spreading', {} as never, {
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
  ],
  invalid: [
    { code: '<App {...props} />', errors: [{ messageId: 'noSpreading' }] },
    { code: '<div {...props}></div>', errors: [{ messageId: 'noSpreading' }] },
    {
      code: '<App {...props} some_other_prop={some_other_prop} />',
      errors: [{ messageId: 'noSpreading' }],
    },
    {
      code: '<App><Image {...props} /><span {...props} /></App>',
      options: [{ exceptions: ['Image', 'img'] }],
      errors: [{ messageId: 'noSpreading' }],
    },
    {
      code: '<App><Image {...props} /><img {...props} /></App>',
      options: [{ custom: 'ignore' }],
      errors: [{ messageId: 'noSpreading' }],
    },
    {
      code: '<App><Image {...props} /><img {...props} /></App>',
      options: [{ html: 'ignore', exceptions: ['Image', 'img'] }],
      errors: [{ messageId: 'noSpreading' }],
    },
    {
      code: '<App><Image {...props} /><img {...props} /><div {...props} /></App>',
      options: [
        { custom: 'ignore', html: 'ignore', exceptions: ['Image', 'img'] },
      ],
      errors: [{ messageId: 'noSpreading' }, { messageId: 'noSpreading' }],
    },
    {
      code: '<App><img {...props} /><Image {...props} /></App>',
      options: [{ html: 'ignore' }],
      errors: [{ messageId: 'noSpreading' }],
    },
    {
      code: '<App><Foo {...{ prop1, prop2, prop3 }} /></App>',
      errors: [{ messageId: 'noSpreading' }],
    },
    {
      code: '<App><Foo {...{ prop1, ...rest }} /></App>',
      options: [{ explicitSpread: 'ignore' }],
      errors: [{ messageId: 'noSpreading' }],
    },
    {
      code: '<App><Foo {...{ ...props }} /></App>',
      options: [{ explicitSpread: 'ignore' }],
      errors: [{ messageId: 'noSpreading' }],
    },
    {
      code: '<App><Foo {...props} /></App>',
      options: [{ explicitSpread: 'ignore' }],
      errors: [{ messageId: 'noSpreading' }],
    },
    {
      code: '<App><components.Group {...props} /><Nav.Item {...props} /></App>',
      options: [{ exceptions: ['components.DropdownIndicator', 'Nav.Item'] }],
      errors: [{ messageId: 'noSpreading' }],
    },
  ],
});
