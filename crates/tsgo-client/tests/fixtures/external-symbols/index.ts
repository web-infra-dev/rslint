import { api } from 'example-dependency';
import 'example-dependency/index.js';
import 'example-reexport';
import '@scope/pkg';
import '@scope/pkg/subpath';
import 'node:assert';
import 'react';
import 'react/jsx-runtime';
import './local';

declare global {
  var A: typeof Math;
}

api.run();
export const value = Math.abs(-1);
