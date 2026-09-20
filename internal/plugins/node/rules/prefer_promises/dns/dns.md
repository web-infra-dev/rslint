# prefer-promises/dns

Prefer the promise API of Node.js's `dns` module.

## Rule Details

This rule reports calls to the callback DNS API and construction of `dns.Resolver`.
It follows CommonJS imports, ES module imports, `process.getBuiltinModule`, and
local aliases, including the `node:dns` module name.

Examples of **incorrect** code for this rule:

```javascript
const dns = require('node:dns');
dns.lookup(hostname, (error, address, family) => {});
new dns.Resolver();
```

Examples of **correct** code for this rule:

```javascript
const { promises: dns } = require('node:dns');
const { address, family } = await dns.lookup(hostname);
new dns.Resolver();
```

```javascript
import dns from 'node:dns/promises';
const addresses = await dns.resolve4(hostname);
```

The rule also reports `getServers()` and `setServers()` on the callback module,
matching upstream behavior. Reading a method without calling it is allowed.

## Options

This rule has no options and does not provide automatic fixes or suggestions.

## Original Documentation

- [eslint-plugin-n: prefer-promises/dns](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-promises/dns.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-promises/dns.js)
