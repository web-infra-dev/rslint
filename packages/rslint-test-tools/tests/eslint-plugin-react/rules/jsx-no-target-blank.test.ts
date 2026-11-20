import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectNoTargetBlankWithoutNoreferrerError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noTargetBlankWithoutNoreferrer',
    'Using target="_blank" without rel="noreferrer" (which implies rel="noopener") is a security risk in older browsers: see https://mathiasbynens.github.io/rel-noopener/#recommendations',
    range,
  );
}

function expectNoTargetBlankWithoutNoopenerError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noTargetBlankWithoutNoopener',
    'Using target="_blank" without rel="noreferrer" or rel="noopener" (the former implies the latter and is preferred due to wider support) is a security risk: see https://mathiasbynens.github.io/rel-noopener/#recommendations',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-no-target-blank', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<a href="foobar"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<a randomTag></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<a target />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `<a href="foobar" target="_blank" rel="noopener noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `<a href="foobar" target="_blank" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<a href="foobar" target="_blank" rel={"noopener noreferrer"}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `<a href="foobar" target="_blank" rel={"noreferrer"}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `<a href={"foobar"} target={"_blank"} rel={"noopener noreferrer"}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<a href={"foobar"} target={"_blank"} rel={"noreferrer"}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `<a href={'foobar'} target={'_blank'} rel={'noopener noreferrer'}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<a href={'foobar'} target={'_blank'} rel={'noreferrer'}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `<a href={\`foobar\`} target={\`_blank\`} rel={\`noopener noreferrer\`}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `<a href={\`foobar\`} target={\`_blank\`} rel={\`noreferrer\`}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `<a target="_blank" {...spreadProps} rel="noopener noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `<a target="_blank" {...spreadProps} rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `<a {...spreadProps} target="_blank" rel="noopener noreferrer" href="https://example.com">s</a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `<a {...spreadProps} target="_blank" rel="noreferrer" href="https://example.com">s</a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `<a target="_blank" rel="noopener noreferrer" {...spreadProps}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `<a target="_blank" rel="noreferrer" {...spreadProps}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `<p target="_blank"></p>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `<a href="foobar" target="_BLANK" rel="NOOPENER noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `<a href="foobar" target="_BLANK" rel="NOREFERRER"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `<a target="_blank" rel={relValue}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `<a target={targetValue} rel="noopener noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `<a target={targetValue} rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `<a target={targetValue} rel={"noopener noreferrer"}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `<a target={targetValue} rel={"noreferrer"}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `<a target={targetValue} href="relative/path"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `<a target={targetValue} href="/absolute/path"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `<a target={'targetValue'} href="/absolute/path"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `<a target={"targetValue"} href="/absolute/path"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `<a target={null} href="//example.com"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `<a {...someObject} href="/absolute/path"></a>`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `<a {...someObject} rel="noreferrer"></a>`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `<a {...someObject} rel="noreferrer" target="_blank"></a>`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `<a {...someObject} href="foobar" target="_blank"></a>`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `<a target="_blank" href={ dynamicLink }></a>`,
      options: [{ enforceDynamicLinks: 'never' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `<a target={"_blank"} href={ dynamicLink }></a>`,
      options: [{ enforceDynamicLinks: 'never' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `<a target={'_blank'} href={ dynamicLink }></a>`,
      options: [{ enforceDynamicLinks: 'never' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `<Link target="_blank" href={ dynamicLink }></Link>`,
      options: [{ enforceDynamicLinks: 'never' }],
      settings: { linkComponents: ['Link'] },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `<Link target="_blank" to={ dynamicLink }></Link>`,
      options: [{ enforceDynamicLinks: 'never' }],
      settings: { linkComponents: { name: 'Link', linkAttribute: 'to' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `<Link target="_blank" to={ dynamicLink }></Link>`,
      options: [{ enforceDynamicLinks: 'never' }],
      settings: { linkComponents: { name: 'Link', linkAttribute: ['to'] } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `<a href="foobar" target="_blank" rel="noopener"></a>`,
      options: [{ allowReferrer: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `<a href="foobar" target="_blank" rel="noreferrer"></a>`,
      options: [{ allowReferrer: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 44 (7.37.5)',
      code: `<a target={3} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 45 (7.37.5)',
      code: `<a href="some-link" {...otherProps} target="some-non-blank-target"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 46 (7.37.5)',
      code: `<a href="some-link" target="some-non-blank-target" {...otherProps}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
      code: `<a target="_blank" href="/absolute/path"></a>`,
      options: [{ forms: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
      code: `<a target="_blank" href="/absolute/path"></a>`,
      options: [{ forms: false, links: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `<form action="https://example.com" target="_blank"></form>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 50 (7.37.5)',
      code: `<form action="https://example.com" target="_blank" rel="noopener noreferrer"></form>`,
      options: [{ forms: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 51 (7.37.5)',
      code: `<form action="https://example.com" target="_blank" rel="noopener noreferrer"></form>`,
      options: [{ forms: true, links: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 52 (7.37.5)',
      code: `<a href target="_blank"/>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 53 (7.37.5)',
      code: `<a href={href} target={isExternal ? "_blank" : undefined} rel="noopener noreferrer" />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 54 (7.37.5)',
      code: `<a href={href} target={isExternal ? undefined : "_blank"} rel={isExternal ? "noreferrer" : "noopener noreferrer"} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 55 (7.37.5)',
      code: `<a href={href} target={isExternal ? undefined : "_blank"} rel={isExternal ? "noreferrer noopener" : "noreferrer"} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 56 (7.37.5)',
      code: `<a href={href} target="_blank" rel={isExternal ? "noreferrer" : "noopener"} />`,
      options: [{ allowReferrer: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 57 (7.37.5)',
      code: `<a href={href} target={isExternal ? "_blank" : undefined} rel={isExternal ? "noreferrer" : undefined} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 58 (7.37.5)',
      code: `<a href={href} target={isSelf ? "_self" : "_blank"} rel={isSelf ? undefined : "noreferrer"} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 59 (7.37.5)',
      code: `<a href={href} target={isSelf ? "_self" : ""} rel={isSelf ? undefined : ""} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 60 (7.37.5)',
      code: `<a href={href} target={isExternal ? "_blank" : undefined} rel={isExternal ? "noopener noreferrer" : undefined} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 61 (7.37.5)',
      code: `<form action={action} />`,
      options: [{ forms: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 62 (7.37.5)',
      code: `<form action={action} {...spread} />`,
      options: [{ forms: true }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<a target="_blank" href="https://example.com/1"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 49])],
      output: `<a target="_blank" href="https://example.com/1" rel="noreferrer"></a>;`,
    },
    {
      code: `<a target="_blank" rel="" href="https://example.com/2"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 56])],
      output: `<a target="_blank" rel="noreferrer" href="https://example.com/2"></a>;`,
    },
    {
      code: `<a target="_blank" rel={0} href="https://example.com/3"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 57])],
      output: `<a target="_blank" rel="noreferrer" href="https://example.com/3"></a>;`,
    },
    {
      code: `<a target="_blank" rel={false} href="https://example.com/4"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 61])],
      output: `<a target="_blank" rel="noreferrer" href="https://example.com/4"></a>;`,
    },
    {
      code: `<a target="_blank" rel={null} href="https://example.com/5"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 60])],
      output: `<a target="_blank" rel="noreferrer" href="https://example.com/5"></a>;`,
    },
    {
      code: `<a target="_blank" rel="noopenernoreferrer" href="https://example.com/6"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 74])],
      output: `<a target="_blank" rel="noopener noreferrer" href="https://example.com/6"></a>;`,
    },
    {
      code: `<a target="_blank" rel="no referrer" href="https://example.com/7"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 67])],
      output: `<a target="_blank" rel="no referrer noreferrer" href="https://example.com/7"></a>;`,
    },
    {
      code: `<a target="_BLANK" href="https://example.com/8"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 49])],
      output: `<a target="_BLANK" href="https://example.com/8" rel="noreferrer"></a>;`,
    },
    {
      code: `<a target="_blank" href="//example.com/9"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 43])],
      output: `<a target="_blank" href="//example.com/9" rel="noreferrer"></a>;`,
    },
    {
      code: `<a target="_blank" href="//example.com/10" rel={true}></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 55])],
      output: `<a target="_blank" href="//example.com/10" rel="noreferrer"></a>;`,
    },
    {
      code: `<a target="_blank" href="//example.com/13" rel={getRel()}></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 59])],
    },
    {
      code: `<a target="_blank" href="//example.com/14" rel={"noopenernoreferrer"}></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 71])],
      output: `<a target="_blank" href="//example.com/14" rel={"noopener noreferrer"}></a>;`,
    },
    {
      code: `<a target="_blank" href="//example.com/17" rel></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 48])],
      output: `<a target="_blank" href="//example.com/17" rel="noreferrer"></a>;`,
    },
    {
      code: `<a target="_blank" href={ dynamicLink }></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 41])],
      output: `<a target="_blank" href={ dynamicLink } rel="noreferrer"></a>;`,
    },
    {
      code: `<a target={'_blank'} href="//example.com/18"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 46])],
      output: `<a target={'_blank'} href="//example.com/18" rel="noreferrer"></a>;`,
    },
    {
      code: `<a target={"_blank"} href="//example.com/19"></a>;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 46])],
      output: `<a target={"_blank"} href="//example.com/19" rel="noreferrer"></a>;`,
    },
    {
      code: `<a href="https://example.com/20" target="_blank"></a>;`,
      options: [{ allowReferrer: true }],
      errors: [expectNoTargetBlankWithoutNoopenerError([1, 1, 1, 50])],
      output: `<a href="https://example.com/20" target="_blank" rel="noopener"></a>;`,
    },
    {
      code: `<a target="_blank" href={ dynamicLink }></a>;`,
      options: [{ enforceDynamicLinks: 'always' }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 41])],
      output: `<a target="_blank" href={ dynamicLink } rel="noreferrer"></a>;`,
    },
    {
      code: `<a {...someObject}></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 20])],
    },
    {
      code: `<a {...someObject} target="_blank"></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 36])],
    },
    {
      code: `<a href="foobar" {...someObject} target="_blank"></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 50])],
    },
    {
      code: `<a href="foobar" target="_blank" rel="noreferrer" {...someObject}></a>;`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 67])],
    },
    {
      code: `<a href="some-link" {...otherProps} target="some-non-blank-target"></a>;`,
      options: [{ warnOnSpreadAttributes: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 68])],
    },
    {
      code: `<a target="_blank" href="//example.com" rel></a>;`,
      options: [{ links: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 45])],
      output: `<a target="_blank" href="//example.com" rel="noreferrer"></a>;`,
    },
    {
      code: `<form method="POST" action="https://example.com" target="_blank"></form>;`,
      options: [{ forms: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 66])],
    },
    {
      code: `<form method="POST" action="https://example.com" rel="" target="_blank"></form>;`,
      options: [{ forms: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 73])],
    },
    {
      code: `<form method="POST" action="https://example.com" rel="noopenernoreferrer" target="_blank"></form>;`,
      options: [{ forms: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 91])],
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? "undefined" : "undefined"} />;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 79])],
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? "noopener" : undefined} />;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 76])],
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? 3 : "noopener noreferrer"} />;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 79])],
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? "noopener noreferrer" : "3"} />;`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 81])],
    },
    {
      code: `<a href={href} target="_blank" rel={isExternal ? "noopener" : "2"} />;`,
      options: [{ allowReferrer: true }],
      errors: [expectNoTargetBlankWithoutNoopenerError([1, 1, 1, 70])],
    },
    {
      code: `<form action={action} target="_blank" />;`,
      options: [{ allowReferrer: true, forms: true }],
      errors: [expectNoTargetBlankWithoutNoopenerError([1, 1, 1, 41])],
    },
    {
      code: `<form action={action} target="_blank" />;`,
      options: [{ forms: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 41])],
    },
    {
      code: `<form action={action} {...spread} />;`,
      options: [{ forms: true, warnOnSpreadAttributes: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 37])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<a target="_blank" href="https://example.com/1"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 49])],
      output: `<a target="_blank" href="https://example.com/1" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<a target="_blank" rel="" href="https://example.com/2"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 56])],
      output: `<a target="_blank" rel="noreferrer" href="https://example.com/2"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `<a target="_blank" rel={0} href="https://example.com/3"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 57])],
      output: `<a target="_blank" rel="noreferrer" href="https://example.com/3"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<a target="_blank" rel={1} href="https://example.com/3"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 57])],
      output: `<a target="_blank" rel="noreferrer" href="https://example.com/3"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `<a target="_blank" rel={false} href="https://example.com/4"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 61])],
      output: `<a target="_blank" rel="noreferrer" href="https://example.com/4"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `<a target="_blank" rel={null} href="https://example.com/5"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 60])],
      output: `<a target="_blank" rel="noreferrer" href="https://example.com/5"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<a target="_blank" rel="noopenernoreferrer" href="https://example.com/6"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 74])],
      output: `<a target="_blank" rel="noopener noreferrer" href="https://example.com/6"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `<a target="_blank" rel="no referrer" href="https://example.com/7"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 67])],
      output: `<a target="_blank" rel="no referrer noreferrer" href="https://example.com/7"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `<a target="_BLANK" href="https://example.com/8"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 49])],
      output: `<a target="_BLANK" href="https://example.com/8" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `<a target="_blank" href="//example.com/9"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 43])],
      output: `<a target="_blank" href="//example.com/9" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `<a target="_blank" href="//example.com/10" rel={true}></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 55])],
      output: `<a target="_blank" href="//example.com/10" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `<a target="_blank" href="//example.com/11" rel={3}></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 52])],
      output: `<a target="_blank" href="//example.com/11" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `<a target="_blank" href="//example.com/12" rel={null}></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 55])],
      output: `<a target="_blank" href="//example.com/12" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `<a target="_blank" href="//example.com/13" rel={getRel()}></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 59])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `<a target="_blank" href="//example.com/14" rel={"noopenernoreferrer"}></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 71])],
      output: `<a target="_blank" href="//example.com/14" rel={"noopener noreferrer"}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `<a target={"_blank"} href={"//example.com/15"} rel={"noopenernoreferrer"}></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 75])],
      output: `<a target={"_blank"} href={"//example.com/15"} rel={"noopener noreferrer"}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `<a target={"_blank"} href={"//example.com/16"} rel={"noopenernoreferrernoreferrernoreferrernoreferrernoreferrer"}></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 115])],
      output: `<a target={"_blank"} href={"//example.com/16"} rel={"noopener noreferrer"}></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `<a target="_blank" href="//example.com/17" rel></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 48])],
      output: `<a target="_blank" href="//example.com/17" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `<a target="_blank" href={ dynamicLink }></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 41])],
      output: `<a target="_blank" href={ dynamicLink } rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `<a target={'_blank'} href="//example.com/18"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 46])],
      output: `<a target={'_blank'} href="//example.com/18" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `<a target={"_blank"} href="//example.com/19"></a>`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 46])],
      output: `<a target={"_blank"} href="//example.com/19" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `<a href="https://example.com/20" target="_blank" rel></a>`,
      options: [{ allowReferrer: true }],
      errors: [expectNoTargetBlankWithoutNoopenerError([1, 1, 1, 54])],
      output: `<a href="https://example.com/20" target="_blank" rel="noopener"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `<a href="https://example.com/20" target="_blank"></a>`,
      options: [{ allowReferrer: true }],
      errors: [expectNoTargetBlankWithoutNoopenerError([1, 1, 1, 50])],
      output: `<a href="https://example.com/20" target="_blank" rel="noopener"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `<a target="_blank" href={ dynamicLink }></a>`,
      options: [{ enforceDynamicLinks: 'always' }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 41])],
      output: `<a target="_blank" href={ dynamicLink } rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `<a {...someObject}></a>`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 20])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `<a {...someObject} target="_blank"></a>`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `<a href="foobar" {...someObject} target="_blank"></a>`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 50])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `<a href="foobar" target="_blank" rel="noreferrer" {...someObject}></a>`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 67])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `<a href="foobar" target="_blank" {...someObject}></a>`,
      options: [
        { enforceDynamicLinks: 'always', warnOnSpreadAttributes: true },
      ],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 50])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `<Link target="_blank" href={ dynamicLink }></Link>`,
      options: [{ enforceDynamicLinks: 'always' }],
      settings: { linkComponents: ['Link'] },
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 44])],
      output: `<Link target="_blank" href={ dynamicLink } rel="noreferrer"></Link>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `<Link target="_blank" to={ dynamicLink }></Link>`,
      options: [{ enforceDynamicLinks: 'always' }],
      settings: { linkComponents: { name: 'Link', linkAttribute: 'to' } },
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 42])],
      output: `<Link target="_blank" to={ dynamicLink } rel="noreferrer"></Link>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `<a href="some-link" {...otherProps} target="some-non-blank-target"></a>`,
      options: [{ warnOnSpreadAttributes: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 68])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `<a href="some-link" target="some-non-blank-target" {...otherProps}></a>`,
      options: [{ warnOnSpreadAttributes: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 68])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `<a target="_blank" href="//example.com" rel></a>`,
      options: [{ links: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 45])],
      output: `<a target="_blank" href="//example.com" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `<a target="_blank" href="//example.com" rel></a>`,
      options: [{ links: true, forms: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 45])],
      output: `<a target="_blank" href="//example.com" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `<a target="_blank" href="//example.com" rel></a>`,
      options: [{ links: true, forms: false }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 45])],
      output: `<a target="_blank" href="//example.com" rel="noreferrer"></a>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `<form method="POST" action="https://example.com" target="_blank"></form>`,
      options: [{ forms: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 66])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `<form method="POST" action="https://example.com" rel="" target="_blank"></form>`,
      options: [{ forms: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 73])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `<form method="POST" action="https://example.com" rel="noopenernoreferrer" target="_blank"></form>`,
      options: [{ forms: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 91])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `<form method="POST" action="https://example.com" rel="noopenernoreferrer" target="_blank"></form>`,
      options: [{ forms: true, links: false }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 91])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `<a href={href} target="_blank" rel={isExternal ? "undefined" : "undefined"} />`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 79])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `<a href={href} target="_blank" rel={isExternal ? "noopener" : undefined} />`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 76])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `<a href={href} target="_blank" rel={isExternal ? "undefined" : "noopener"} />`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 78])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `<a href={href} target={isExternal ? "_blank" : undefined} rel={isExternal ? undefined : "noopener noreferrer"} />`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 114])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      code: `<a href={href} target="_blank" rel={isExternal ? 3 : "noopener noreferrer"} />`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 79])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      code: `<a href={href} target="_blank" rel={isExternal ? "noopener noreferrer" : "3"} />`,
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 81])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `<a href={href} target="_blank" rel={isExternal ? "noopener" : "2"} />`,
      options: [{ allowReferrer: true }],
      errors: [expectNoTargetBlankWithoutNoopenerError([1, 1, 1, 70])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `<form action={action} target="_blank" />`,
      options: [{ allowReferrer: true, forms: true }],
      errors: [expectNoTargetBlankWithoutNoopenerError([1, 1, 1, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      code: `<form action={action} target="_blank" />`,
      options: [{ forms: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `<form action={action} {...spread} />`,
      options: [{ forms: true, warnOnSpreadAttributes: true }],
      errors: [expectNoTargetBlankWithoutNoreferrerError([1, 1, 1, 37])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
