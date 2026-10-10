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
  interface PrototypeNode {
    child: PrototypeNode;
    label: string;
    run(): void;
  }
  var PrototypeNode: {
    prototype: PrototypeNode;
    new (): PrototypeNode;
    create(): PrototypeNode;
  };
  var PrototypeAlias: typeof PrototypeNode;
}

api.run();
export const value = Math.abs(-1);
export const text = 'hello'.slice(1);
