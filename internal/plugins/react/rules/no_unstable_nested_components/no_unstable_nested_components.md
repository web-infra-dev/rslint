# no-unstable-nested-components

## Rule Details

Creating components inside components causes React to see a new component type on every render. Because React identifies a component by its reference, a new reference forces React to unmount the existing subtree and mount a fresh one on every parent render — losing DOM state, triggering layout effects, and leaking memory. Move nested component definitions to the module scope (or accept data through props).

Examples of **incorrect** code for this rule:

```javascript
function ParentComponent() {
  function UnstableNestedComponent() {
    return <div />;
  }

  return (
    <div>
      <UnstableNestedComponent />
    </div>
  );
}
```

```javascript
class ParentComponent extends React.Component {
  render() {
    const UnstableNestedComponent = () => <div />;
    return <UnstableNestedComponent />;
  }
}
```

```javascript
function ParentComponent() {
  return <ComponentWithProps footer={() => <div />} />;
}
```

Examples of **correct** code for this rule:

```javascript
function OutsideDefinedComponent() {
  return <div />;
}

function ParentComponent() {
  return (
    <div>
      <OutsideDefinedComponent />
    </div>
  );
}
```

```javascript
function ParentComponent(props) {
  return (
    <ul>
      {props.items.map(item => (
        <li key={item.id}>{item.name}</li>
      ))}
    </ul>
  );
}
```

```javascript
function ParentComponent() {
  return <ComponentForProps renderFooter={() => <div />} />;
}
```

## Options

### `allowAsProps`

Allow components to be declared inside other components' props without reporting. Defaults to `false`.

```json
{ "react/no-unstable-nested-components": ["error", { "allowAsProps": true }] }
```

```javascript
function ParentComponent() {
  return <ComponentWithProps footer={() => <div />} />;
}
```

### `propNamePattern`

Glob pattern matched against JSX attribute / object-literal property names. When the pattern matches, a component defined in that position is treated as a render prop and is NOT reported. Defaults to `"render*"`.

```json
{ "react/no-unstable-nested-components": ["error", { "propNamePattern": "*Renderer" }] }
```

```javascript
function ParentComponent() {
  return <Table rowRenderer={(rowData) => <Row data={rowData} />} />;
}
```

## Original Documentation

- [eslint-plugin-react no-unstable-nested-components](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-unstable-nested-components.md)
