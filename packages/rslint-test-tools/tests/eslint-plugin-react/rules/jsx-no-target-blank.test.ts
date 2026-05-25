import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-no-target-blank', {} as never, {
  valid: [
    { code: `<a href="foobar"></a>;` },
    { code: `<a randomTag></a>;` },
    { code: `<a target />;` },
    {
      code: `<a href="foobar" target="_blank" rel="noopener noreferrer"></a>;`,
    },
    { code: `<a href="foobar" target="_blank" rel="noreferrer"></a>;` },
    {
      code: `<a href="foobar" target="_blank" rel={"noopener noreferrer"}></a>;`,
    },
    { code: `<a href="foobar" target="_blank" rel={"noreferrer"}></a>;` },
    {
      code: `<a href={"foobar"} target={"_blank"} rel={"noopener noreferrer"}></a>;`,
    },
    { code: `<a href={"foobar"} target={"_blank"} rel={"noreferrer"}></a>;` },
    {
      code: `<a target="_blank" {...spreadProps} rel="noopener noreferrer"></a>;`,
    },
    { code: `<a target="_blank" {...spreadProps} rel="noreferrer"></a>;` },
    {
      code: `<a {...spreadProps} target="_blank" rel="noopener noreferrer" href="https://example.com">s</a>;`,
    },
    {
      code: `<a target="_blank" rel="noopener noreferrer" {...spreadProps}></a>;`,
    },
    { code: `<a target="_blank" rel="noreferrer" {...spreadProps}></a>;` },
    { code: `<p target="_blank"></p>;` },
    {
      code: `<a href="foobar" target="_BLANK" rel="NOOPENER noreferrer"></a>;`,
    },
    { code: `<a href="foobar" target="_BLANK" rel="NOREFERRER"></a>;` },
    { code: `<a target="_blank" rel={relValue}></a>;` },
    { code: `<a target={targetValue} rel="noopener noreferrer"></a>;` },
    { code: `<a target={targetValue} rel="noreferrer"></a>;` },
    { code: `<a target={targetValue} rel={"noopener noreferrer"}></a>;` },
    { code: `<a target={targetValue} rel={"noreferrer"}></a>;` },
    { code: `<a target={targetValue} href="relative/path"></a>;` },
    { code: `<a target={targetValue} href="/absolute/path"></a>;` },
    { code: `<a target={'targetValue'} href="/absolute/path"></a>;` },
    { code: `<a target={"targetValue"} href="/absolute/path"></a>;` },
    { code: `<a target={null} href="//example.com"></a>;` },
    {
      code: `<a {...someObject} href="/absolute/path"></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
    },
    {
      code: `<a {...someObject} rel="noreferrer"></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
    },
    {
      code: `<a {...someObject} rel="noreferrer" target="_blank"></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
    },
    {
      code: `<a {...someObject} href="foobar" target="_blank"></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
    },
    {
      code: `<a target="_blank" href={ dynamicLink }></a>;`,
      options: [{ enforceDynamicLinks: 'never' }],
    },
    {
      code: `<a target={"_blank"} href={ dynamicLink }></a>;`,
      options: [{ enforceDynamicLinks: 'never' }],
    },
    {
      code: `<a target={'_blank'} href={ dynamicLink }></a>;`,
      options: [{ enforceDynamicLinks: 'never' }],
    },
    {
      code: `<a href="foobar" target="_blank" rel="noopener"></a>;`,
      options: [{ allowReferrer: true }],
    },
    {
      code: `<a href="foobar" target="_blank" rel="noreferrer"></a>;`,
      options: [{ allowReferrer: true }],
    },
    { code: `<a target={3} />;` },
    {
      code: `<a href="some-link" {...otherProps} target="some-non-blank-target"></a>;`,
    },
    {
      code: `<a href="some-link" target="some-non-blank-target" {...otherProps}></a>;`,
    },
    {
      code: `<a target="_blank" href="/absolute/path"></a>;`,
      options: [{ forms: false }],
    },
    {
      code: `<a target="_blank" href="/absolute/path"></a>;`,
      options: [{ forms: false, links: true }],
    },
    { code: `<form action="https://example.com" target="_blank"></form>;` },
    {
      code: `<form action="https://example.com" target="_blank" rel="noopener noreferrer"></form>;`,
      options: [{ forms: true }],
    },
    {
      code: `<form action="https://example.com" target="_blank" rel="noopener noreferrer"></form>;`,
      options: [{ forms: true, links: false }],
    },
    { code: `<a href target="_blank"/>;` },
    {
      code: `<a href={href} target={isExternal ? "_blank" : undefined} rel="noopener noreferrer" />;`,
    },
    {
      code: `<a href={href} target={isExternal ? undefined : "_blank"} rel={isExternal ? "noreferrer" : "noopener noreferrer"} />;`,
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? "noreferrer" : "noopener"} />;`,
      options: [{ allowReferrer: true }],
    },
    {
      code: `<a href={href} target={isExternal ? "_blank" : undefined} rel={isExternal ? "noreferrer" : undefined} />;`,
    },
    {
      code: `<a href={href} target={isSelf ? "_self" : "_blank"} rel={isSelf ? undefined : "noreferrer"} />;`,
    },
    {
      code: `<form action={action} />;`,
      options: [{ forms: true }],
    },
    {
      code: `<form action={action} {...spread} />;`,
      options: [{ forms: true }],
    },
  ],
  invalid: [
    {
      code: `<a target="_blank" href="https://example.com/1"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" rel="" href="https://example.com/2"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" rel={0} href="https://example.com/3"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" rel={false} href="https://example.com/4"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" rel={null} href="https://example.com/5"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" rel="noopenernoreferrer" href="https://example.com/6"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" rel="no referrer" href="https://example.com/7"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_BLANK" href="https://example.com/8"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" href="//example.com/9"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" href="//example.com/10" rel={true}></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" href="//example.com/13" rel={getRel()}></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" href="//example.com/14" rel={"noopenernoreferrer"}></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" href="//example.com/17" rel></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" href={ dynamicLink }></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target={'_blank'} href="//example.com/18"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target={"_blank"} href="//example.com/19"></a>;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a href="https://example.com/20" target="_blank"></a>;`,
      options: [{ allowReferrer: true }],
      errors: [{ messageId: 'noTargetBlankWithoutNoopener' }],
    },
    {
      code: `<a target="_blank" href={ dynamicLink }></a>;`,
      options: [{ enforceDynamicLinks: 'always' }],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a {...someObject}></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a {...someObject} target="_blank"></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a href="foobar" {...someObject} target="_blank"></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a href="foobar" target="_blank" rel="noreferrer" {...someObject}></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a href="some-link" {...otherProps} target="some-non-blank-target"></a>;`,
      options: [{ warnOnSpreadAttributes: true }],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a target="_blank" href="//example.com" rel></a>;`,
      options: [{ links: true }],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<form method="POST" action="https://example.com" target="_blank"></form>;`,
      options: [{ forms: true }],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<form method="POST" action="https://example.com" rel="" target="_blank"></form>;`,
      options: [{ forms: true }],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<form method="POST" action="https://example.com" rel="noopenernoreferrer" target="_blank"></form>;`,
      options: [{ forms: true }],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? "undefined" : "undefined"} />;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? "noopener" : undefined} />;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? 3 : "noopener noreferrer"} />;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? "noopener noreferrer" : "3"} />;`,
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? "noopener" : "2"} />;`,
      options: [{ allowReferrer: true }],
      errors: [{ messageId: 'noTargetBlankWithoutNoopener' }],
    },
    {
      code: `<form action={action} target="_blank" />;`,
      options: [{ allowReferrer: true, forms: true }],
      errors: [{ messageId: 'noTargetBlankWithoutNoopener' }],
    },
    {
      code: `<form action={action} target="_blank" />;`,
      options: [{ forms: true }],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
    {
      code: `<form action={action} {...spread} />;`,
      options: [{ forms: true, warnOnSpreadAttributes: true }],
      errors: [{ messageId: 'noTargetBlankWithoutNoreferrer' }],
    },
  ],
});
