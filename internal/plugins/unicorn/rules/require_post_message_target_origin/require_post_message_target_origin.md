# require-post-message-target-origin

Require an explicit `targetOrigin` argument in `.postMessage()` calls so that
messages can be restricted to an intended origin.

## Examples

Incorrect:

```js
window.postMessage(sensitiveData);
window.postMessage({ token: authToken });
iframe.contentWindow.postMessage(data);
```

Correct:

```js
window.postMessage(sensitiveData, 'https://trusted-domain.com');
window.postMessage({ token: authToken }, 'https://api.example.com');
iframe.contentWindow.postMessage(data, 'https://expected-iframe-origin.com');

// Use a wildcard only for non-sensitive public data.
window.postMessage({ publicData: 'hello' }, '*');
```

The rule provides editor suggestions, not automatic fixes. Depending on the
receiver, suggestions use its `location.origin`, `self.location.origin`, or
`'*'`. Choose the intended recipient's origin; a wildcard does not restrict
which origin can receive the message.

## Options

This rule has no options.

## Limitations

Like upstream, this rule cannot distinguish a window from a `Worker`,
`MessagePort`, `Client`, or `BroadcastChannel`. Those APIs do not accept a
`targetOrigin` argument, so enable this rule only where it is appropriate.
The rule is not included in upstream's recommended configuration.

Computed property access, spread arguments, and optional calls are ignored.
Optional member access, such as `window?.postMessage(message)`, is checked.

## References

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/require-post-message-target-origin.md)
- [Upstream implementation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/require-post-message-target-origin.js)
