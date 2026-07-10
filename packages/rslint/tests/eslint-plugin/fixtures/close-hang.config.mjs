import plugin from './close-hang-plugin.mjs';

export default [
  {
    plugins: { close: plugin },
    rules: { 'close/hang': 'error' },
  },
];
