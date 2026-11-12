// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-at.js
import { mkdtemp, realpath, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { runInNewContext } from 'node:vm';

import { afterAll, beforeAll, describe, expect, test } from 'rstack/test';
import {
  NodeRslintService,
  type Diagnostic,
  type LintOptions,
} from '@rslint/core/internal';
import { RSLintService } from '@rslint/core/service';

interface ExpectedDiagnostic {
  id: string;
  range: [number, number, number, number];
  description?: string;
  fixed?: string;
  suggested?: string;
}

interface Case {
  code: string;
  mode?: 'ts' | 'typed';
  options?: {
    checkAllIndexAccess?: boolean;
    getLastElementFunctions?: string[];
  }[];
  errors?: ExpectedDiagnostic[];
}

interface Group {
  name: string;
  valid?: (string | Case)[];
  invalid?: Case[];
}

const checkAll = [{ checkAllIndexAccess: true }];
const groups: Group[] = [
  {
    name: 'Keyword spacing',
    invalid: [
      {
        code: 'function f() {return(_).last(getItems(/* keep */));}',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 21, 1, 29],
            description: '_.last',
            fixed: 'function f() {return getItems(/* keep */).at(-1);}',
          },
        ],
      },
      {
        code: 'function f() {throw(_).last(getItems(/* keep */));}',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 20, 1, 28],
            description: '_.last',
            fixed: 'function f() {throw getItems(/* keep */).at(-1);}',
          },
        ],
      },
      {
        code: 'if (flag) {}else(_.last)(getItems(/* keep */));',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 18, 1, 24],
            description: '_.last',
            fixed: 'if (flag) {}else getItems(/* keep */).at(-1);',
          },
        ],
      },
      {
        code: 'do(_).last(getItems(/* keep */));while (flag);',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 3, 1, 11],
            description: '_.last',
            fixed: 'do getItems(/* keep */).at(-1);while (flag);',
          },
        ],
      },
    ],
  },
  {
    name: 'Last-element calls and retained comments',
    valid: [
      '_.last(document.querySelectorAll(/* keep */ "li"));',
      'function f() { return _.last((/* keep */ arguments)); }',
    ],
    invalid: [
      {
        code: 'lodash.last(((items /* keep */)));',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 12],
            description: 'lodash.last',
            fixed: '((items /* keep */)).at(-1);',
          },
        ],
      },
      {
        code: 'underscore.last((items // keep\r\n));',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 16],
            description: 'underscore.last',
            fixed: '(items // keep\r\n).at(-1);',
          },
        ],
      },
      {
        code: 'before()\n_.last([1, /* keep */ 2]);',
        errors: [
          {
            id: 'get-last-function',
            range: [2, 1, 2, 7],
            description: '_.last',
            fixed: 'before()\n;[1, /* keep */ 2].at(-1);',
          },
        ],
      },
      {
        code: '_.last(items as /* keep */ number[]);',
        mode: 'ts',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: '(items as /* keep */ number[]).at(-1);',
          },
        ],
      },
      {
        code: 'utils.lastElement(getItems(/* keep */));',
        options: [{ getLastElementFunctions: ['utils.lastElement'] }],
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 18],
            description: 'utils.lastElement',
            fixed: 'getItems(/* keep */).at(-1);',
          },
        ],
      },
      {
        code: '_./* discard */last(getItems(/* keep */));',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 20],
            description: '_.last',
          },
        ],
      },
      {
        code: '_.last(/* discard */ getItems(/* keep */));',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
          },
        ],
      },
      {
        code: '_.last(getItems(/* keep */) /* discard */);',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
          },
        ],
      },
      {
        code: '_.last /* discard */ ((/* keep */ items));',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
          },
        ],
      },
    ],
  },
  {
    name: 'Last-element call comment placement',
    invalid: [
      {
        code: '_.last(getItems(/* keep */));',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: 'getItems(/* keep */).at(-1);',
          },
        ],
      },
      {
        code: '_.last((/* keep */ items));',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: '(/* keep */ items).at(-1);',
          },
        ],
      },
    ],
  },
  {
    name: 'JSDoc DOM types in JavaScript',
    valid: [
      '/** @param {NodeList | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }',
      '/** @returns {HTMLCollection | (Null | Undefined)} */ function read() { return source; } const nodes = read(); nodes[nodes.length - 1];',
      '/** @param {() => (NodeList | (null | undefined))} read */ function run(read) { const nodes = read(); return _.last(nodes); }',
      {
        code: '/** @param {(null | undefined) | HTMLCollectionOf<Element>} nodes */ function run(nodes) { return nodes?.[0]; }',
        options: checkAll,
      },
    ],
    invalid: [
      {
        code: '/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }',
        errors: [
          {
            id: 'negative-index',
            range: [1, 105, 1, 121],
            fixed:
              '/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes.at(-1); }',
          },
        ],
      },
      {
        code: '/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }',
        errors: [
          {
            id: 'negative-index',
            range: [1, 89, 1, 105],
            fixed:
              '/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes.at(-1); }',
          },
        ],
      },
    ],
  },
  {
    name: 'JSDoc DOM types with the TypeScript parser',
    valid: [
      {
        code: '/** @param {NodeList | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }',
        mode: 'ts',
      },
      {
        code: '/** @returns {HTMLCollection | (Null | Undefined)} */ function read() { return source; } const nodes = read(); nodes[nodes.length - 1];',
        mode: 'ts',
      },
      {
        code: '/** @param {() => (NodeList | (null | undefined))} read */ function run(read) { const nodes = read(); return _.last(nodes); }',
        mode: 'ts',
      },
      {
        code: '/** @param {(null | undefined) | HTMLCollectionOf<Element>} nodes */ function run(nodes) { return nodes?.[0]; }',
        mode: 'ts',
        options: checkAll,
      },
    ],
    invalid: [
      {
        code: '/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 105, 1, 121],
            fixed:
              '/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes.at(-1); }',
          },
        ],
      },
      {
        code: '/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 89, 1, 105],
            fixed:
              '/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes.at(-1); }',
          },
        ],
      },
    ],
  },
  {
    name: 'JSDoc DOM types with project service',
    valid: [
      {
        code: '/** @param {NodeList | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }',
        mode: 'typed',
      },
      {
        code: '/** @returns {HTMLCollection | (Null | Undefined)} */ function read() { return source; } const nodes = read(); nodes[nodes.length - 1];',
        mode: 'typed',
      },
      {
        code: '/** @param {() => (NodeList | (null | undefined))} read */ function run(read) { const nodes = read(); return _.last(nodes); }',
        mode: 'typed',
      },
      {
        code: '/** @param {(null | undefined) | HTMLCollectionOf<Element>} nodes */ function run(nodes) { return nodes?.[0]; }',
        mode: 'typed',
        options: checkAll,
      },
    ],
    invalid: [
      {
        code: '/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 105, 1, 121],
            fixed:
              '/** @param {NodeList | Array<Element> | (null | undefined)} nodes */ function run(nodes) { return nodes.at(-1); }',
          },
        ],
      },
      {
        code: '/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes[nodes.length - 1]; }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 89, 1, 105],
            fixed:
              '/** @param {NodeList | (Unresolved | null)} nodes */ function run(nodes) { return nodes.at(-1); }',
          },
        ],
      },
    ],
  },
  {
    name: 'Assignment and update targets',
    valid: [
      { code: '(array[array.length - 1] as number) = 3;', mode: 'ts' },
      { code: '(array[array.length - 1] as number) += 3;', mode: 'ts' },
      { code: '(array[array.length - 1] as number)++;', mode: 'ts' },
      { code: '--(array[array.length - 1] as number);', mode: 'ts' },
      { code: 'delete (array[array.length - 1] as number);', mode: 'ts' },
      { code: 'delete (array.slice(-1)[0] as number);', mode: 'ts' },
      { code: '(<number>array[array.length - 1]) = 3;', mode: 'ts' },
      { code: '(<number>array[array.length - 1]) += 3;', mode: 'ts' },
      { code: '(<number>array[array.length - 1])++;', mode: 'ts' },
      { code: '--(<number>array[array.length - 1]);', mode: 'ts' },
      { code: 'delete (<number>array[array.length - 1]);', mode: 'ts' },
      { code: 'delete (<number>array.slice(-1)[0]);', mode: 'ts' },
      { code: 'array[array.length - 1]! = 3;', mode: 'ts' },
      { code: 'array[array.length - 1]! += 3;', mode: 'ts' },
      { code: 'array[array.length - 1]!++;', mode: 'ts' },
      { code: '--array[array.length - 1]!;', mode: 'ts' },
      { code: 'delete array[array.length - 1]!;', mode: 'ts' },
      { code: 'delete array.slice(-1)[0]!;', mode: 'ts' },
      { code: '(array[array.length - 1] satisfies number) = 3;', mode: 'ts' },
      { code: '(array[array.length - 1] satisfies number) += 3;', mode: 'ts' },
      { code: '(array[array.length - 1] satisfies number)++;', mode: 'ts' },
      { code: '--(array[array.length - 1] satisfies number);', mode: 'ts' },
      {
        code: 'delete (array[array.length - 1] satisfies number);',
        mode: 'ts',
      },
      { code: 'delete (array.slice(-1)[0] satisfies number);', mode: 'ts' },
      { code: '((array[array.length - 1] as number)!) = 3;', mode: 'ts' },
      { code: '((array[array.length - 1] as number)!) += 3;', mode: 'ts' },
      { code: '((array[array.length - 1] as number)!)++;', mode: 'ts' },
      { code: '--((array[array.length - 1] as number)!);', mode: 'ts' },
      { code: 'delete ((array[array.length - 1] as number)!);', mode: 'ts' },
      { code: 'delete ((array.slice(-1)[0] as number)!);', mode: 'ts' },
      'delete array?.[array.length - 1];',
      'delete array?.slice(-1)[0];',
      { code: 'delete ((array?.[array.length - 1] as number)!);', mode: 'ts' },
      { code: '([array[array.length - 1]!] = values);', mode: 'ts' },
      {
        code: '({value: (array[array.length - 1] as number)} = object);',
        mode: 'ts',
      },
      { code: 'for (array[array.length - 1]! of values) {}', mode: 'ts' },
      { code: '(array[0] as number) = 3;', mode: 'ts', options: checkAll },
    ],
    invalid: [
      {
        code: 'const value = (array[array.length - 1] as number);',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 22, 1, 38],
            fixed: 'const value = (array.at(-1) as number);',
          },
        ],
      },
    ],
  },
  {
    name: 'Method, tag, and constructor contexts',
    valid: [
      'array[array.length - 1]()',
      'array[array.length - 1]?.()',
      'array[array.length - 1]`template`',
      'array?.[array.length - 1]()',
      '(array?.[array.length - 1])()',
      'array.slice(-1)[0]()',
      'array.slice(-1)[0]`template`',
      '(array?.slice(-1)[0])()',
      { code: '(array[array.length - 1] as number)()', mode: 'ts' },
      { code: '(array.slice(-1)[0] as number)`template`', mode: 'ts' },
      { code: '(<number>array[array.length - 1])()', mode: 'ts' },
      { code: '(<number>array.slice(-1)[0])`template`', mode: 'ts' },
      { code: 'array[array.length - 1]!()', mode: 'ts' },
      { code: 'array.slice(-1)[0]!`template`', mode: 'ts' },
      { code: '(array[array.length - 1] satisfies number)()', mode: 'ts' },
      { code: '(array.slice(-1)[0] satisfies number)`template`', mode: 'ts' },
      { code: '((array[array.length - 1] as number)!)()', mode: 'ts' },
      { code: '((array.slice(-1)[0] as number)!)`template`', mode: 'ts' },
      { code: 'array[0]()', options: checkAll },
    ],
    invalid: [
      {
        code: 'new array[array.length - 1]()',
        errors: [
          {
            id: 'negative-index',
            range: [1, 11, 1, 27],
            fixed: 'new (array.at(-1))()',
          },
        ],
      },
      {
        code: 'new array[array.length - 1].Constructor()',
        errors: [
          {
            id: 'negative-index',
            range: [1, 11, 1, 27],
            fixed: 'new (array.at(-1)).Constructor()',
          },
        ],
      },
      {
        code: 'new array[array.length - 1].tag`template`()',
        errors: [
          {
            id: 'negative-index',
            range: [1, 11, 1, 27],
            fixed: 'new (array.at(-1)).tag`template`()',
          },
        ],
      },
      {
        code: 'new (array[array.length - 1])()',
        errors: [
          {
            id: 'negative-index',
            range: [1, 12, 1, 28],
            fixed: 'new (array.at(-1))()',
          },
        ],
      },
      {
        code: 'new (array[array.length - 1].Constructor)()',
        errors: [
          {
            id: 'negative-index',
            range: [1, 12, 1, 28],
            fixed: 'new (array.at(-1).Constructor)()',
          },
        ],
      },
      {
        code: 'new array[array.length - /* keep */ 1]()',
        errors: [
          {
            id: 'negative-index',
            range: [1, 11, 1, 38],
            fixed: 'new (array.at(- /* keep */ 1))()',
          },
        ],
      },
      {
        code: 'new array[array.length - 1]!.Constructor()',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 11, 1, 27],
            fixed: 'new (array.at(-1))!.Constructor()',
          },
        ],
      },
      {
        code: 'new array[0]()',
        options: checkAll,
        errors: [
          { id: 'index', range: [1, 11, 1, 12], fixed: 'new (array.at(0))()' },
        ],
      },
    ],
  },
  {
    name: 'DOM collection receivers',
    valid: [
      'const nodes = element.children; nodes[nodes.length - 1];',
      'const nodes = element.children; nodes.slice(-1)[0];',
      'const nodes = element.children; _.last(nodes);',
      '_.last(element.children);',
      { code: 'const nodes = element.children; nodes[0];', options: checkAll },
      'const nodes = parent.childNodes; nodes[nodes.length - 1];',
      'const nodes = parent.childNodes; nodes.slice(-1)[0];',
      'const nodes = parent.childNodes; _.last(nodes);',
      '_.last(parent.childNodes);',
      { code: 'const nodes = parent.childNodes; nodes[0];', options: checkAll },
      'const nodes = document.querySelectorAll("li"); nodes[nodes.length - 1];',
      'const nodes = document.querySelectorAll("li"); nodes.slice(-1)[0];',
      'const nodes = document.querySelectorAll("li"); _.last(nodes);',
      '_.last(document.querySelectorAll("li"));',
      {
        code: 'const nodes = document.querySelectorAll("li"); nodes[0];',
        options: checkAll,
      },
      'const nodes = document.getElementsByClassName("item"); nodes[nodes.length - 1];',
      'const nodes = document.getElementsByClassName("item"); nodes.slice(-1)[0];',
      'const nodes = document.getElementsByClassName("item"); _.last(nodes);',
      '_.last(document.getElementsByClassName("item"));',
      {
        code: 'const nodes = document.getElementsByClassName("item"); nodes[0];',
        options: checkAll,
      },
      'const nodes = document.getElementsByName("item"); nodes[nodes.length - 1];',
      'const nodes = document.getElementsByName("item"); nodes.slice(-1)[0];',
      'const nodes = document.getElementsByName("item"); _.last(nodes);',
      '_.last(document.getElementsByName("item"));',
      {
        code: 'const nodes = document.getElementsByName("item"); nodes[0];',
        options: checkAll,
      },
      'const nodes = document.getElementsByTagName("li"); nodes[nodes.length - 1];',
      'const nodes = document.getElementsByTagName("li"); nodes.slice(-1)[0];',
      'const nodes = document.getElementsByTagName("li"); _.last(nodes);',
      '_.last(document.getElementsByTagName("li"));',
      {
        code: 'const nodes = document.getElementsByTagName("li"); nodes[0];',
        options: checkAll,
      },
      'const nodes = document.getElementsByTagNameNS(namespace, "li"); nodes[nodes.length - 1];',
      'const nodes = document.getElementsByTagNameNS(namespace, "li"); nodes.slice(-1)[0];',
      'const nodes = document.getElementsByTagNameNS(namespace, "li"); _.last(nodes);',
      '_.last(document.getElementsByTagNameNS(namespace, "li"));',
      {
        code: 'const nodes = document.getElementsByTagNameNS(namespace, "li"); nodes[0];',
        options: checkAll,
      },
      'const nodes = document.querySelectorAll("li"); const alias = nodes; alias[alias.length - 1];',
      'const nodes = parent?.childNodes; nodes?.[nodes.length - 1];',
      'const nodes = document?.querySelectorAll?.("li"); _.last(nodes);',
      'const nodes = element.children; nodes.slice(-1).pop();',
      'const nodes = element.children; nodes.slice(-1).shift();',
      'const nodes = element.children; nodes.slice(-2, -1)[0];',
      'element.children.slice(-1)[0];',
      'lodash.last(element.children);',
      'underscore.last(element.childNodes);',
      {
        code: 'const nodes = document.querySelectorAll("li"); utils.lastElement(nodes);',
        options: [{ getLastElementFunctions: ['utils.lastElement'] }],
      },
      {
        code: 'function foo(nodes: NodeList) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'ts',
      },
      {
        code: 'function foo(nodes: NodeList) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'typed',
      },
      {
        code: 'function foo(nodes: NodeListOf<Element>) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'ts',
      },
      {
        code: 'function foo(nodes: NodeListOf<Element>) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'typed',
      },
      {
        code: 'function foo(nodes: HTMLCollection) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'ts',
      },
      {
        code: 'function foo(nodes: HTMLCollection) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'typed',
      },
      {
        code: 'function foo(nodes: HTMLCollectionOf<Element>) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'ts',
      },
      {
        code: 'function foo(nodes: HTMLCollectionOf<Element>) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'typed',
      },
      {
        code: 'function foo<T extends NodeList>(nodes: T) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'ts',
      },
      {
        code: 'function foo<T extends NodeList>(nodes: T) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'typed',
      },
      {
        code: 'const nodes = document.querySelectorAll("li") as NodeListOf<Element>; (nodes as NodeListOf<Element>)[nodes.length - 1];',
        mode: 'ts',
      },
      {
        code: 'const nodes = <NodeList>element.childNodes; _.last(<NodeList>nodes);',
        mode: 'ts',
      },
      {
        code: 'const nodes = element.children satisfies HTMLCollection; nodes![nodes!.length - 1];',
        mode: 'ts',
      },
      {
        code: '(element.children as string[])[element.children.length - 1];\n(element.children as string[]).slice(-1)[0];\n_.last(element.children as string[]);',
        mode: 'ts',
      },
      {
        code: 'type Nodes = NodeListOf<Element>; function foo(nodes: Nodes) { _.last(nodes); }',
        mode: 'ts',
      },
      {
        code: 'function foo(nodes: NodeList | HTMLCollection) { _.last(nodes); }',
        mode: 'ts',
      },
      {
        code: 'function foo(nodes: NodeList | undefined) { return nodes?.[nodes.length - 1]; }',
        mode: 'ts',
      },
      {
        code: 'function foo(nodes: HTMLCollection | null) { return nodes?.[0]; }',
        mode: 'ts',
        options: checkAll,
      },
      {
        code: 'function foo(nodes: NodeList) { return nodes[0]; }',
        mode: 'ts',
        options: checkAll,
      },
      {
        code: 'declare function getNodes(): NodeListOf<Element>; const nodes = getNodes(); nodes[nodes.length - 1];',
        mode: 'typed',
      },
      {
        code: 'declare function getNodes(): NodeListOf<Element>; const nodes = getNodes(); _.last(nodes);',
        mode: 'ts',
      },
      {
        code: 'declare function getNodes(): NodeList; const nodes = getNodes?.(); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);',
        mode: 'ts',
      },
      {
        code: 'declare function getNodes(): NodeList; const nodes = getNodes?.(); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);',
        mode: 'typed',
      },
      {
        code: 'declare function getNodes(): HTMLCollection; _.last(getNodes?.());',
        mode: 'ts',
      },
      {
        code: 'function getNodes() { return document.querySelectorAll("li"); } const nodes = getNodes(); nodes[nodes.length - 1];',
        mode: 'typed',
      },
      {
        code: 'let nodes = document.querySelectorAll("li"); _.last(nodes);',
        mode: 'typed',
      },
      {
        code: 'const {children: nodes} = document.body; _.last(nodes);',
        mode: 'typed',
      },
      {
        code: 'function foo(value: unknown) { if (value instanceof NodeList) { return value[value.length - 1]; } }',
        mode: 'typed',
      },
      {
        code: 'const nodes: unknown = element.children; _.last(nodes);',
        mode: 'ts',
      },
      {
        code: 'const nodes: NodeList | string[] = element.children; _.last(nodes);',
        mode: 'ts',
      },
      {
        code: 'declare const source: {nodes: HTMLCollection}; _.last(source.nodes);',
        mode: 'typed',
      },
      {
        code: 'type Nodes = ReturnType<() => NodeListOf<Element>>; declare const nodes: Nodes; _.last(nodes);',
        mode: 'typed',
      },
      {
        code: 'function foo(nodes: NodeList | undefined) { return nodes?.[nodes.length - 1]; }',
        mode: 'typed',
      },
    ],
  },
  {
    name: 'Negative index access',
    valid: [
      'array.at(-1)',
      'array[array.length - 0];',
      'array[array.length + 1]',
      'array[array.length + -1]',
      "const modes = new Set(['foo']); modes.clear(); array[modes.size ? 1 : 'x'];",
      "const modes = new Set(['foo']); modes.clear(); array[(modes.size && 1) || index];",
      {
        code: "const modes = new Set(['foo']); modes.clear(); array[(modes.size ? 1 : index) as number];",
        mode: 'typed',
      },
      'foo[bar.length - 1]',
      'array[array.length - 0.5]',
      'array[array.length - 1.5]',
      'array[array.length - 1] = 1',
      'array[array.length - 1] %= 1',
      '++ array[array.length - 1]',
      'array[array.length - 1] --',
      'delete array[array.length - 1]',
      'class Foo {bar; #bar; baz() {return this.#bar[this.bar.length - 1]}}',
      '([array[array.length - 1]] = [])',
      '({foo: array[array.length - 1] = 9} = {})',
      'function foo() {return arguments[arguments.length - 1]}',
      'parent?.childNodes[parent.childNodes.length - 1];',
      'element.children[element.children.length - 1];',
      'document.querySelectorAll("li")[document.querySelectorAll("li").length - 1];',
    ],
    invalid: [
      {
        code: 'array[array.length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 7, 1, 23],
            fixed: 'array.at(-1);',
          },
        ],
      },
      {
        code: 'array?.[array.length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 9, 1, 25],
            fixed: 'array?.at(-1);',
          },
        ],
      },
      {
        code: 'array[array.length -1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 7, 1, 22],
            fixed: 'array.at(-1);',
          },
        ],
      },
      {
        code: 'array[array.length - /* comment */ 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 7, 1, 37],
            fixed: 'array.at(- /* comment */ 1);',
          },
        ],
      },
      {
        code: 'array[array.length - 1.];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 7, 1, 24],
            fixed: 'array.at(- 1.);',
          },
        ],
      },
      {
        code: 'array[array.length - 0b1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 7, 1, 25],
            fixed: 'array.at(- 0b1);',
          },
        ],
      },
      {
        code: 'array[array.length - 9];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 7, 1, 23],
            fixed: 'array.at(-9);',
          },
        ],
      },
      {
        code: 'array[0][array[0].length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 10, 1, 29],
            fixed: 'array[0].at(-1);',
          },
        ],
      },
      {
        code: 'array[(( array.length )) - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 7, 1, 29],
            fixed: 'array.at(-1);',
          },
        ],
      },
      {
        code: 'array[array.length - (( 1 ))];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 7, 1, 29],
            fixed: 'array.at(- (( 1 )));',
          },
        ],
      },
      {
        code: 'array[(( array.length - 1 ))];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 10, 1, 26],
            fixed: 'array.at((( -1 )));',
          },
        ],
      },
      {
        code: '(( array ))[array.length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 13, 1, 29],
            fixed: '(( array )).at(-1);',
          },
        ],
      },
      {
        code: '(( array[array.length - 1] ));',
        errors: [
          {
            id: 'negative-index',
            range: [1, 10, 1, 26],
            fixed: '(( array.at(-1) ));',
          },
        ],
      },
      {
        code: 'array[array.length - 1].pop().shift()[0];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 7, 1, 23],
            fixed: 'array.at(-1).pop().shift()[0];',
          },
        ],
      },
      {
        code: 'a = array[array.length - 1]',
        errors: [
          {
            id: 'negative-index',
            range: [1, 11, 1, 27],
            fixed: 'a = array.at(-1)',
          },
        ],
      },
      {
        code: 'const a = array[array.length - 1]',
        errors: [
          {
            id: 'negative-index',
            range: [1, 17, 1, 33],
            fixed: 'const a = array.at(-1)',
          },
        ],
      },
      {
        code: 'const {a = array[array.length - 1]} = {}',
        errors: [
          {
            id: 'negative-index',
            range: [1, 18, 1, 34],
            fixed: 'const {a = array.at(-1)} = {}',
          },
        ],
      },
      {
        code: 'typeof array[array.length - 1]',
        errors: [
          {
            id: 'negative-index',
            range: [1, 14, 1, 30],
            fixed: 'typeof array.at(-1)',
          },
        ],
      },
      {
        code: 'class Foo {bar; baz() {return this.bar[this.bar.length - 1]}}',
        errors: [
          {
            id: 'negative-index',
            range: [1, 40, 1, 59],
            fixed: 'class Foo {bar; baz() {return this.bar.at(-1)}}',
          },
        ],
      },
      {
        code: 'class Foo {#bar; baz() {return this.#bar[this.#bar.length - 1]}}',
        errors: [
          {
            id: 'negative-index',
            range: [1, 42, 1, 62],
            fixed: 'class Foo {#bar; baz() {return this.#bar.at(-1)}}',
          },
        ],
      },
      {
        code: '(array as Foo[])[array.length - 1]',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 18, 1, 34],
            fixed: '(array as Foo[]).at(-1)',
          },
        ],
      },
      {
        code: '(array as Foo[])[(array as Foo[]).length - 1]',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 18, 1, 45],
            fixed: '(array as Foo[]).at(-1)',
          },
        ],
      },
      {
        code: '<Foo[]>array[array.length - 1]',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 14, 1, 30],
            fixed: '<Foo[]>array.at(-1)',
          },
        ],
      },
      {
        code: '<Foo[]>array[(<Foo[]>array).length - 1]',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 14, 1, 39],
            fixed: '<Foo[]>array.at(-1)',
          },
        ],
      },
      {
        code: 'array![array.length - 1]',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 8, 1, 24],
            fixed: 'array!.at(-1)',
          },
        ],
      },
      {
        code: 'array![array!.length - 1]',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 8, 1, 25],
            fixed: 'array!.at(-1)',
          },
        ],
      },
    ],
  },
  {
    name: 'String index access',
    valid: [
      'array[0]',
      'unknown[0]',
      '1[0]',
      '({0: "a"})[0]',
      'const object = {0: "a"}; object[0]',
      'string[unknown]',
      'string[-1]',
      'string[1.5]',
      'string[1n]',
      'string[0] = value',
      '"string"[1]',
      '`string`[1]',
      'String(value)[0]',
      'String.fromCodePoint(65)[0]',
      '(typeof value)[0]',
      'const string = "string"; string[1]',
      'String.fromCharCode(65)[0]',
      {
        code: 'function foo(value: string | number) { return value[0]; }',
        mode: 'ts',
      },
      { code: 'function foo(value: number) { return value[0]; }', mode: 'ts' },
      'const value = other; const other = value; value[0];',
      {
        code: 'type Value = Other; type Other = Value; function foo(value: Value) { return value[0]; }',
        mode: 'ts',
      },
      { code: '(value as string)[0]', mode: 'ts' },
      { code: '(value satisfies string)[0]', mode: 'ts' },
      { code: '(<string>value)[0]', mode: 'ts' },
      { code: 'function foo(value: string) { return value![0]; }', mode: 'ts' },
      { code: 'function foo(value: string) { return value[0]; }', mode: 'ts' },
      {
        code: 'const value: string | number = "string"; value[0];',
        mode: 'ts',
      },
      { code: 'const value: unknown = "string"; value[0];', mode: 'ts' },
      {
        code: 'type StringAlias = string; function foo(value: StringAlias) { return value[0]; }',
        mode: 'ts',
      },
      {
        code: 'type StringUnion = "a" | "b"; function foo(value: StringUnion) { return value[0]; }',
        mode: 'ts',
      },
      {
        code: 'declare function getValue(): string; getValue()[0];',
        mode: 'typed',
      },
      {
        code: 'type Value = ReturnType<() => string>; declare const value: Value; value[0];',
        mode: 'typed',
      },
      {
        code: 'function foo(value: string | number) { if (typeof value === "string") { return value[0]; } }',
        mode: 'typed',
      },
    ],
    invalid: [
      {
        code: '"string"["string".length - 1]',
        errors: [
          {
            id: 'negative-index',
            range: [1, 10, 1, 29],
            fixed: '"string".at(-1)',
          },
        ],
      },
      {
        code: 'const string = "string"; string[string.length - 1]',
        errors: [
          {
            id: 'negative-index',
            range: [1, 33, 1, 50],
            fixed: 'const string = "string"; string.at(-1)',
          },
        ],
      },
    ],
  },
  {
    name: 'charAt',
    valid: [
      'string.charAt(string.length - 0);',
      'string.charAt(string.length + 1)',
      'string.charAt(string.length + -1)',
      'foo.charAt(bar.length - 1)',
      'string?.charAt?.(string.length - 1);',
      'string.charAt(9);',
    ],
    invalid: [
      {
        code: 'string.charAt(string.length - 1);',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 15, 1, 32],
            suggested: 'string.at(- 1);',
          },
        ],
      },
      {
        code: 'string?.charAt(string.length - 1);',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 16, 1, 33],
            suggested: 'string?.at(- 1);',
          },
        ],
      },
      {
        code: 'string.charAt(string.length - 0o11);',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 15, 1, 35],
            suggested: 'string.at(- 0o11);',
          },
        ],
      },
      {
        code: 'some.string.charAt(some.string.length - 1);',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 20, 1, 42],
            suggested: 'some.string.at(- 1);',
          },
        ],
      },
      {
        code: 'string.charAt((( string.length )) - 0xFF);',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 15, 1, 41],
            suggested: 'string.at(- 0xFF);',
          },
        ],
      },
      {
        code: 'string.charAt(string.length - (( 1 )));',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 15, 1, 38],
            suggested: 'string.at(- (( 1 )));',
          },
        ],
      },
      {
        code: 'string.charAt((( string.length - 1 )));',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 18, 1, 35],
            suggested: 'string.at((( - 1 )));',
          },
        ],
      },
      {
        code: '(( string )).charAt(string.length - 1);',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 21, 1, 38],
            suggested: '(( string )).at(- 1);',
          },
        ],
      },
      {
        code: '(( string.charAt ))(string.length - 1);',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 21, 1, 38],
            suggested: '(( string.at ))(- 1);',
          },
        ],
      },
      {
        code: '(( string.charAt(string.length - 1) ));',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 18, 1, 35],
            suggested: '(( string.at(- 1) ));',
          },
        ],
      },
      {
        code: '(string as string).charAt((string as string).length - 1);',
        mode: 'ts',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 27, 1, 56],
            suggested: '(string as string).at(- 1);',
          },
        ],
      },
      {
        code: '<string>string.charAt((<string>string).length - 1);',
        mode: 'ts',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 23, 1, 50],
            suggested: '<string>string.at(- 1);',
          },
        ],
      },
      {
        code: 'string!.charAt(string!.length - 1);',
        mode: 'ts',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 16, 1, 34],
            suggested: 'string!.at(- 1);',
          },
        ],
      },
      {
        code: 'string.charAt(string /* keep */ .length - 1);',
        errors: [{ id: 'string-char-at-negative', range: [1, 15, 1, 44] }],
      },
    ],
  },
  {
    name: 'Single-character substring',
    valid: [
      'string.substring()',
      'string.substring(0)',
      'string.substring(index)',
      'string.substring(0, 2)',
      'string.substring(-1, 0)',
      'string.substring(1.5, 2.5)',
      'string.substring(index, index + 2)',
      'string.substring(index, other + 1)',
      'string.substring(other, index + 1)',
      'string.substring(index, index + length)',
      'string.substring(index, length + index)',
      'string.substring(...arguments)',
      'string.substring(index, index + 1, extraArgument)',
    ],
    invalid: [
      {
        code: 'string.substring(0, 1)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 23],
            suggested: 'string.at(0)',
          },
        ],
      },
      {
        code: 'string.substring(1, 2)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 23],
            suggested: 'string.at(1)',
          },
        ],
      },
      {
        code: 'string.substring(2, 1)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 23],
            suggested: 'string.at(1)',
          },
        ],
      },
      {
        code: 'string.substring(index, index + 1)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 35],
            suggested: 'string.at(index)',
          },
        ],
      },
      {
        code: 'string.substring(index, 1 + index)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 35],
            suggested: 'string.at(index)',
          },
        ],
      },
      {
        code: 'string.substring(index - 1, index)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 35],
            suggested: 'string.at(index - 1)',
          },
        ],
      },
      {
        code: 'string.substring(index + 1, index)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 35],
            suggested: 'string.at(index)',
          },
        ],
      },
      {
        code: 'string.substring(1 + index, index)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 35],
            suggested: 'string.at(index)',
          },
        ],
      },
      {
        code: 'string.substring(index, /* comment */ index + 1)',
        errors: [{ id: 'string-substring', range: [1, 1, 1, 49] }],
      },
      {
        code: 'string.substring(index, index + /* comment */ 1)',
        errors: [{ id: 'string-substring', range: [1, 1, 1, 49] }],
      },
      {
        code: 'string.substring(0, /* between */ 1)',
        errors: [{ id: 'string-substring', range: [1, 1, 1, 37] }],
      },
      {
        code: 'string.substring((( index )), (( index )) + 1)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 47],
            suggested: 'string.at((( index )))',
          },
        ],
      },
      {
        code: 'string.substring((( index )) - 1, (( index )))',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 47],
            suggested: 'string.at((( index )) - 1)',
          },
        ],
      },
      {
        code: '(( string )).substring(index, index + 1)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 41],
            suggested: '(( string )).at(index)',
          },
        ],
      },
      {
        code: 'string?.substring(index, index + 1)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 36],
            suggested: 'string?.at(index)',
          },
        ],
      },
      {
        code: 'string.substring?.(index, index + 1)',
        errors: [
          {
            id: 'string-substring',
            range: [1, 1, 1, 37],
            suggested: 'string.at?.(index)',
          },
        ],
      },
    ],
  },
  {
    name: 'One-argument slice',
    valid: [
      'array.slice(-1)',
      'new array.slice(-1)',
      'array.slice(-0)[0]',
      'array.slice(-9).pop()',
      'array.slice(-1.1)[0]',
      'array.slice(-1)?.[0]',
      'array.slice?.(-1)[0]',
      'array.notSlice(-1)[0]',
      'array.slice()[0]',
      'array.slice(...[-1])[0]',
      'array.slice(-1).shift?.()',
      'array.slice(-1)?.shift()',
      'array.slice(-1).shift(...[])',
      'new array.slice(-1).shift()',
      'array.slice(-1)[0] += 1',
      '++ array.slice(-1)[0]',
      'array.slice(-1)[0] --',
      'delete array.slice(-1)[0]',
      'array.slice(-9)[0]',
      'array.slice(-9).shift()',
      'array.slice(-0xA)[0b000]',
    ],
    invalid: [
      {
        code: 'array.slice(-1)[0]',
        errors: [{ id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at(-1)' }],
      },
      {
        code: 'array?.slice(-1)[0]',
        errors: [{ id: 'slice', range: [1, 8, 1, 13], fixed: 'array?.at(-1)' }],
      },
      {
        code: 'array.slice(-1).pop()',
        errors: [{ id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at(-1)' }],
      },
      {
        code: 'array.slice(-1.0).shift()',
        errors: [
          { id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at(-1.0)' },
        ],
      },
      {
        code: 'array.slice(-1)[(( 0 ))];',
        errors: [{ id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at(-1);' }],
      },
      {
        code: 'array.slice(-(( 1 )))[0];',
        errors: [
          { id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at(-(( 1 )));' },
        ],
      },
      {
        code: 'array.slice((( -1 )))[0];',
        errors: [
          { id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at((( -1 )));' },
        ],
      },
      {
        code: '(( array.slice(-1) ))[0];',
        errors: [
          { id: 'slice', range: [1, 10, 1, 15], fixed: '(( array.at(-1) ));' },
        ],
      },
      {
        code: '(( array )).slice(-1)[0];',
        errors: [
          { id: 'slice', range: [1, 13, 1, 18], fixed: '(( array )).at(-1);' },
        ],
      },
      {
        code: '(( array.slice(-1)[0] ));',
        errors: [
          { id: 'slice', range: [1, 10, 1, 15], fixed: '(( array.at(-1) ));' },
        ],
      },
      {
        code: '(( array.slice(-1) )).pop();',
        errors: [
          { id: 'slice', range: [1, 10, 1, 15], fixed: '(( array.at(-1) ));' },
        ],
      },
      {
        code: '(( array.slice(-1).pop ))();',
        errors: [
          { id: 'slice', range: [1, 10, 1, 15], fixed: '(( array.at(-1) ));' },
        ],
      },
      {
        code: '(( array.slice(-1).pop() ));',
        errors: [
          { id: 'slice', range: [1, 10, 1, 15], fixed: '(( array.at(-1) ));' },
        ],
      },
      {
        code: 'array.slice(-1)[0].pop().shift().slice(-1)',
        errors: [
          {
            id: 'slice',
            range: [1, 7, 1, 12],
            fixed: 'array.at(-1).pop().shift().slice(-1)',
          },
        ],
      },
      {
        code: '(array as Foo[]).slice(-1)[0]',
        mode: 'ts',
        errors: [
          {
            id: 'slice',
            range: [1, 18, 1, 23],
            fixed: '(array as Foo[]).at(-1)',
          },
        ],
      },
      {
        code: '(array as Foo[]).slice(-1).pop()',
        mode: 'ts',
        errors: [
          {
            id: 'slice',
            range: [1, 18, 1, 23],
            fixed: '(array as Foo[]).at(-1)',
          },
        ],
      },
      {
        code: '(<Foo[]>array).slice(-1).pop()',
        mode: 'ts',
        errors: [
          {
            id: 'slice',
            range: [1, 16, 1, 21],
            fixed: '(<Foo[]>array).at(-1)',
          },
        ],
      },
      {
        code: '<Foo[]>array.slice(-1)[0]',
        mode: 'ts',
        errors: [
          { id: 'slice', range: [1, 14, 1, 19], fixed: '<Foo[]>array.at(-1)' },
        ],
      },
      {
        code: '<Foo[]>array.slice(-1).shift()',
        mode: 'ts',
        errors: [
          { id: 'slice', range: [1, 14, 1, 19], fixed: '<Foo[]>array.at(-1)' },
        ],
      },
      {
        code: 'array!.slice(-1)[0]',
        mode: 'ts',
        errors: [{ id: 'slice', range: [1, 8, 1, 13], fixed: 'array!.at(-1)' }],
      },
      {
        code: 'array!.slice(-1).pop()',
        mode: 'ts',
        errors: [{ id: 'slice', range: [1, 8, 1, 13], fixed: 'array!.at(-1)' }],
      },
    ],
  },
  {
    name: 'Two-argument slice with adjacent bounds',
    valid: ['array.slice(-9.1, -8.1)[0]'],
    invalid: [
      {
        code: 'array.slice(-9, -8)[0]',
        errors: [{ id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at(-9)' }],
      },
      {
        code: 'array.slice(-9, -0o10)[0]',
        errors: [{ id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at(-9)' }],
      },
      {
        code: 'array.slice(-9, -8).pop()',
        errors: [{ id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at(-9)' }],
      },
      {
        code: 'array.slice(-9, -8).shift()',
        errors: [{ id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at(-9)' }],
      },
      {
        code: 'array.slice((( -9 )), (( -8 )), ).shift()',
        errors: [
          { id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at((( -9 )))' },
        ],
      },
      {
        code: '(( array.slice(-9, -8).shift ))()',
        errors: [
          { id: 'slice', range: [1, 10, 1, 15], fixed: '(( array.at(-9) ))' },
        ],
      },
    ],
  },
  {
    name: 'Other two-argument slices',
    valid: [
      'array.slice(-unknown, -unknown2)[0]',
      'array.slice(-9.1, unknown)[0]',
      'array.slice(-9, unknown).pop()',
      'array.slice(-9, ...unknown)[0]',
      'array.slice(...[-9], unknown)[0]',
      'array.slice(-9, unknown)[0]',
      'array.slice(-0o11, -7)[0]',
      'array.slice(-9, unknown).shift()',
      'const KNOWN = -8; array.slice(-9, KNOWN).shift()',
      'array.slice(-9, 0)[0]',
      '(( (( array.slice( ((-9)), ((unknown)), ).shift ))() ));',
      'array.slice(-9, (a, really, _really, complicated, second) => argument)[0]',
    ],
  },
  {
    name: 'Last-element functions',
    valid: [
      'new _.last(array)',
      '_.last(array, 2)',
      '_.last(...array)',
      'function foo() {return _.last(arguments)}',
      'function foo() {return lodash.last(arguments)}',
      'function foo() {return underscore.last(arguments)}',
      {
        code: 'function foo() {return getLast(arguments)}',
        options: [{ getLastElementFunctions: ['getLast'] }],
      },
    ],
    invalid: [
      {
        code: '_.last(array)',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: 'array.at(-1)',
          },
        ],
      },
      {
        code: 'lodash.last(array)',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 12],
            description: 'lodash.last',
            fixed: 'array.at(-1)',
          },
        ],
      },
      {
        code: 'underscore.last(array)',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 16],
            description: 'underscore.last',
            fixed: 'array.at(-1)',
          },
        ],
      },
      {
        code: '_.last(foo.bar)',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: 'foo.bar.at(-1)',
          },
        ],
      },
      {
        code: '_.last(new Array)',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: '(new Array).at(-1)',
          },
        ],
      },
      {
        code: 'const foo = []\n_.last([bar])',
        errors: [
          {
            id: 'get-last-function',
            range: [2, 1, 2, 7],
            description: '_.last',
            fixed: 'const foo = []\n;[bar].at(-1)',
          },
        ],
      },
      {
        code: 'const foo = []\n_.last( new Array )',
        errors: [
          {
            id: 'get-last-function',
            range: [2, 1, 2, 7],
            description: '_.last',
            fixed: 'const foo = []\n;(new Array).at(-1)',
          },
        ],
      },
      {
        code: 'const foo = []\n_.last( (( new Array )) )',
        errors: [
          {
            id: 'get-last-function',
            range: [2, 1, 2, 7],
            description: '_.last',
            fixed: 'const foo = []\n;(( new Array )).at(-1)',
          },
        ],
      },
      {
        code: 'if (foo) _.last([bar])',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 10, 1, 16],
            description: '_.last',
            fixed: 'if (foo) [bar].at(-1)',
          },
        ],
      },
      {
        code: '_.last(getLast(utils.lastOne(array)))',
        options: [
          { getLastElementFunctions: ['getLast', '  utils.lastOne  '] },
        ],
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: 'getLast(utils.lastOne(array)).at(-1)',
          },
          {
            id: 'get-last-function',
            range: [1, 8, 1, 15],
            description: 'getLast',
            fixed: '_.last(utils.lastOne(array).at(-1))',
          },
          {
            id: 'get-last-function',
            range: [1, 16, 1, 29],
            description: 'utils.lastOne',
            fixed: '_.last(getLast(array.at(-1)))',
          },
        ],
      },
    ],
  },
  {
    name: 'checkAllIndexAccess',
    valid: [
      {
        code: "const object = {value: true}; Object.defineProperty(object, 'value', {get() { return false; }}); array[object.value ? 1 : index];",
        options: checkAll,
      },
      { code: '++array[1]', options: checkAll },
      {
        code: 'const offset = 5;const extraArgument = 6;string.charAt(offset + 9, extraArgument)',
        options: checkAll,
      },
      { code: 'array[unknown]', options: checkAll },
      { code: 'array[-1]', options: checkAll },
      { code: 'array[1.5]', options: checkAll },
      { code: 'array[1n]', options: checkAll },
      { code: 'function foo() {return arguments[0]}', options: checkAll },
      { code: 'const object = {1: 1, a: 2}; object[1]', options: checkAll },
      { code: '({1: 1})[1]', options: checkAll },
      {
        code: 'const object = {1: 1} as const; object[1]',
        mode: 'ts',
        options: checkAll,
      },
      {
        code: 'const object = {1: 1} satisfies Record<number, number>; object[1]',
        mode: 'ts',
        options: checkAll,
      },
      {
        code: 'const object = {1: 1}; (object as Record<number, number>)[1]',
        mode: 'ts',
        options: checkAll,
      },
      {
        code: 'const object = {1: 1}; object![1]',
        mode: 'ts',
        options: checkAll,
      },
      { code: 'parent.childNodes[0]', options: checkAll },
      { code: 'element.children[0]', options: checkAll },
      { code: 'document.querySelectorAll("li")[0]', options: checkAll },
      { code: 'document.getElementsByClassName("item")[0]', options: checkAll },
      { code: 'document.getElementsByTagName("li")[0]', options: checkAll },
      {
        code: 'document.getElementsByTagNameNS(namespace, "li")[0]',
        options: checkAll,
      },
      { code: 'document.getElementsByName("item")[0]', options: checkAll },
    ],
    invalid: [
      {
        code: 'array[0]',
        options: checkAll,
        errors: [{ id: 'index', range: [1, 7, 1, 8], fixed: 'array.at(0)' }],
      },
      {
        code: 'array[1]',
        options: checkAll,
        errors: [{ id: 'index', range: [1, 7, 1, 8], fixed: 'array.at(1)' }],
      },
      {
        code: 'array[5 + 9]',
        options: checkAll,
        errors: [
          { id: 'index', range: [1, 7, 1, 12], fixed: 'array.at(5 + 9)' },
        ],
      },
      {
        code: 'const offset = 5;array[offset + 9]',
        options: checkAll,
        errors: [
          {
            id: 'index',
            range: [1, 24, 1, 34],
            fixed: 'const offset = 5;array.at(offset + 9)',
          },
        ],
      },
      {
        code: 'let object = {1: 1, a: 2}; object[1]',
        options: checkAll,
        errors: [
          {
            id: 'index',
            range: [1, 35, 1, 36],
            fixed: 'let object = {1: 1, a: 2}; object.at(1)',
          },
        ],
      },
      {
        code: '"string"[1]',
        options: checkAll,
        errors: [
          { id: 'index', range: [1, 10, 1, 11], fixed: '"string".at(1)' },
        ],
      },
      {
        code: '`string`[1]',
        options: checkAll,
        errors: [
          { id: 'index', range: [1, 10, 1, 11], fixed: '`string`.at(1)' },
        ],
      },
      {
        code: 'const string = "string"; string[1]',
        options: checkAll,
        errors: [
          {
            id: 'index',
            range: [1, 33, 1, 34],
            fixed: 'const string = "string"; string.at(1)',
          },
        ],
      },
      {
        code: 'new Uint8Array([1])[0]',
        options: checkAll,
        errors: [
          {
            id: 'index',
            range: [1, 21, 1, 22],
            fixed: 'new Uint8Array([1]).at(0)',
          },
        ],
      },
      {
        code: 'array[array.length - 1]',
        options: checkAll,
        errors: [
          { id: 'negative-index', range: [1, 7, 1, 23], fixed: 'array.at(-1)' },
        ],
      },
      {
        code: 'string.charAt(9)',
        options: checkAll,
        errors: [
          {
            id: 'string-char-at',
            range: [1, 15, 1, 16],
            suggested: 'string.at(9)',
          },
        ],
      },
      {
        code: 'string.charAt(5 + 9)',
        options: checkAll,
        errors: [
          {
            id: 'string-char-at',
            range: [1, 15, 1, 20],
            suggested: 'string.at(5 + 9)',
          },
        ],
      },
      {
        code: 'const offset = 5;string.charAt(offset + 9)',
        options: checkAll,
        errors: [
          {
            id: 'string-char-at',
            range: [1, 32, 1, 42],
            suggested: 'const offset = 5;string.at(offset + 9)',
          },
        ],
      },
      {
        code: 'string.charAt(unknown)',
        options: checkAll,
        errors: [
          {
            id: 'string-char-at',
            range: [1, 15, 1, 22],
            suggested: 'string.at(unknown)',
          },
        ],
      },
      {
        code: 'string.charAt(1.5)',
        options: checkAll,
        errors: [
          {
            id: 'string-char-at',
            range: [1, 15, 1, 18],
            suggested: 'string.at(1.5)',
          },
        ],
      },
      {
        code: 'string.charAt(1n)',
        options: checkAll,
        errors: [
          {
            id: 'string-char-at',
            range: [1, 15, 1, 17],
            suggested: 'string.at(1n)',
          },
        ],
      },
      {
        code: 'string.charAt(string.length - 1)',
        options: checkAll,
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 15, 1, 32],
            suggested: 'string.at(- 1)',
          },
        ],
      },
      {
        code: 'foo.charAt(bar.length - 1)',
        options: checkAll,
        errors: [
          {
            id: 'string-char-at',
            range: [1, 12, 1, 26],
            suggested: 'foo.at(bar.length - 1)',
          },
        ],
      },
    ],
  },
  {
    name: 'Negative charAt and last-element comments',
    valid: [
      { code: 'string.charAt(-1)', options: checkAll },
      { code: 'string.charAt(-1.5)', options: checkAll },
      { code: 'string.charAt(-0x1)', options: checkAll },
      { code: 'string.charAt(-0b1)', options: checkAll },
      { code: 'string.charAt(1 - 2)', options: checkAll },
      { code: 'const index = -1; string.charAt(index)', options: checkAll },
    ],
    invalid: [
      {
        code: 'string.charAt(Symbol.iterator)',
        options: checkAll,
        errors: [
          {
            id: 'string-char-at',
            range: [1, 15, 1, 30],
            suggested: 'string.at(Symbol.iterator)',
          },
        ],
      },
      {
        code: 'const a = _.last(/* keep */ foo);',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 11, 1, 17],
            description: '_.last',
          },
        ],
      },
      {
        code: 'const a = _.last(foo);',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 11, 1, 17],
            description: '_.last',
            fixed: 'const a = foo.at(-1);',
          },
        ],
      },
    ],
  },
  {
    name: 'Comments in the length access',
    invalid: [
      {
        code: 'const a = array[array /* keep */ .length - 1];',
        errors: [{ id: 'negative-index', range: [1, 17, 1, 45] }],
      },
      {
        code: 'const a = array[/* keep */ array.length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 28, 1, 44],
            fixed: 'const a = array.at(/* keep */ -1);',
          },
        ],
      },
      {
        code: 'const a = array[(/* keep */ array.length) - 1];',
        errors: [{ id: 'negative-index', range: [1, 17, 1, 46] }],
      },
      {
        code: 'const a = array[array.length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 17, 1, 33],
            fixed: 'const a = array.at(-1);',
          },
        ],
      },
      {
        code: 'const a = array[1 /* keep */ + 1];',
        options: checkAll,
        errors: [
          {
            id: 'index',
            range: [1, 17, 1, 33],
            fixed: 'const a = array.at(1 /* keep */ + 1);',
          },
        ],
      },
    ],
  },
  {
    name: 'Comments in the slice call',
    invalid: [
      {
        code: 'array.slice(-1).pop(/* keep */)',
        errors: [{ id: 'slice', range: [1, 7, 1, 12] }],
      },
      {
        code: 'array.slice(/* keep */ -1).pop()',
        errors: [{ id: 'slice', range: [1, 7, 1, 12] }],
      },
      {
        code: 'array.slice(-1).shift()',
        errors: [{ id: 'slice', range: [1, 7, 1, 12], fixed: 'array.at(-1)' }],
      },
      {
        code: 'array.slice(-1)[0 /* keep */]',
        errors: [{ id: 'slice', range: [1, 7, 1, 12] }],
      },
      {
        code: 'foo(array.slice(-1)[0], /* keep */ bar)',
        errors: [
          {
            id: 'slice',
            range: [1, 11, 1, 16],
            fixed: 'foo(array.at(-1), /* keep */ bar)',
          },
        ],
      },
    ],
  },
  {
    name: 'Unknown and non-DOM receivers',
    invalid: [
      {
        code: 'const nodes = getNodes(); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);',
        errors: [
          {
            id: 'negative-index',
            range: [1, 33, 1, 49],
            fixed:
              'const nodes = getNodes(); nodes.at(-1); nodes.slice(-1)[0]; _.last(nodes);',
          },
          {
            id: 'slice',
            range: [1, 58, 1, 63],
            fixed:
              'const nodes = getNodes(); nodes[nodes.length - 1]; nodes.at(-1); _.last(nodes);',
          },
          {
            id: 'get-last-function',
            range: [1, 72, 1, 78],
            description: '_.last',
            fixed:
              'const nodes = getNodes(); nodes[nodes.length - 1]; nodes.slice(-1)[0]; nodes.at(-1);',
          },
        ],
      },
      {
        code: '_.last(getNodes());',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: 'getNodes().at(-1);',
          },
        ],
      },
      {
        code: 'const nodes = Array.from(document.querySelectorAll("li")); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);',
        errors: [
          {
            id: 'negative-index',
            range: [1, 66, 1, 82],
            fixed:
              'const nodes = Array.from(document.querySelectorAll("li")); nodes.at(-1); nodes.slice(-1)[0]; _.last(nodes);',
          },
          {
            id: 'slice',
            range: [1, 91, 1, 96],
            fixed:
              'const nodes = Array.from(document.querySelectorAll("li")); nodes[nodes.length - 1]; nodes.at(-1); _.last(nodes);',
          },
          {
            id: 'get-last-function',
            range: [1, 105, 1, 111],
            description: '_.last',
            fixed:
              'const nodes = Array.from(document.querySelectorAll("li")); nodes[nodes.length - 1]; nodes.slice(-1)[0]; nodes.at(-1);',
          },
        ],
      },
      {
        code: 'const nodes = [...element.children]; nodes[nodes.length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 44, 1, 60],
            fixed: 'const nodes = [...element.children]; nodes.at(-1);',
          },
        ],
      },
      {
        code: '_.last(Array.from(element.children));',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: 'Array.from(element.children).at(-1);',
          },
        ],
      },
      {
        code: '_.last([...element.childNodes]);',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: '[...element.childNodes].at(-1);',
          },
        ],
      },
      {
        code: 'const nodes = [1, 2]; nodes[nodes.length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 29, 1, 45],
            fixed: 'const nodes = [1, 2]; nodes.at(-1);',
          },
        ],
      },
      {
        code: 'const string = "abc"; string[string.length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 30, 1, 47],
            fixed: 'const string = "abc"; string.at(-1);',
          },
        ],
      },
      {
        code: 'const bytes = new Uint8Array([1, 2]); bytes[bytes.length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 45, 1, 61],
            fixed: 'const bytes = new Uint8Array([1, 2]); bytes.at(-1);',
          },
        ],
      },
      {
        code: 'const nodes = alias; const alias = nodes; _.last(nodes);',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 43, 1, 49],
            description: '_.last',
            fixed: 'const nodes = alias; const alias = nodes; nodes.at(-1);',
          },
        ],
      },
      {
        code: 'let nodes = element.children; nodes = []; _.last(nodes);',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 43, 1, 49],
            description: '_.last',
            fixed: 'let nodes = element.children; nodes = []; nodes.at(-1);',
          },
        ],
      },
      {
        code: 'const {children: nodes} = element; _.last(nodes);',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 36, 1, 42],
            description: '_.last',
            fixed: 'const {children: nodes} = element; nodes.at(-1);',
          },
        ],
      },
      {
        code: 'const nodes = element.children; function foo(nodes) { return _.last(nodes); }',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 62, 1, 68],
            description: '_.last',
            fixed:
              'const nodes = element.children; function foo(nodes) { return nodes.at(-1); }',
          },
        ],
      },
      {
        code: '_.last(document["querySelectorAll"]("li"));',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 1, 1, 7],
            description: '_.last',
            fixed: 'document["querySelectorAll"]("li").at(-1);',
          },
        ],
      },
      {
        code: 'function foo(value: any) { return value[value.length - 1]; }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 41, 1, 57],
            fixed: 'function foo(value: any) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: any) { return value[value.length - 1]; }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 41, 1, 57],
            fixed: 'function foo(value: any) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: unknown) { return value[value.length - 1]; }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 45, 1, 61],
            fixed: 'function foo(value: unknown) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: unknown) { return value[value.length - 1]; }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 45, 1, 61],
            fixed: 'function foo(value: unknown) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: NodeList | string[]) { return value[value.length - 1]; }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 57, 1, 73],
            fixed:
              'function foo(value: NodeList | string[]) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: NodeList | string[]) { return value[value.length - 1]; }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 57, 1, 73],
            fixed:
              'function foo(value: NodeList | string[]) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: string[]) { return value[value.length - 1]; }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 46, 1, 62],
            fixed: 'function foo(value: string[]) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: string[]) { return value[value.length - 1]; }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 46, 1, 62],
            fixed: 'function foo(value: string[]) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: readonly number[]) { return value[value.length - 1]; }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 55, 1, 71],
            fixed:
              'function foo(value: readonly number[]) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: readonly number[]) { return value[value.length - 1]; }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 55, 1, 71],
            fixed:
              'function foo(value: readonly number[]) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: string) { return value[value.length - 1]; }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 44, 1, 60],
            fixed: 'function foo(value: string) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: string) { return value[value.length - 1]; }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 44, 1, 60],
            fixed: 'function foo(value: string) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: Uint8Array) { return value[value.length - 1]; }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 48, 1, 64],
            fixed: 'function foo(value: Uint8Array) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: Uint8Array) { return value[value.length - 1]; }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 48, 1, 64],
            fixed: 'function foo(value: Uint8Array) { return value.at(-1); }',
          },
        ],
      },
      {
        code: 'const nodes = Array.from(element.children); nodes[0];',
        options: checkAll,
        errors: [
          {
            id: 'index',
            range: [1, 51, 1, 52],
            fixed: 'const nodes = Array.from(element.children); nodes.at(0);',
          },
        ],
      },
      {
        code: 'function foo(value: any) { return value[0]; }',
        mode: 'typed',
        options: checkAll,
        errors: [
          {
            id: 'index',
            range: [1, 41, 1, 42],
            fixed: 'function foo(value: any) { return value.at(0); }',
          },
        ],
      },
      {
        code: 'const nodes = Array.from(document.querySelectorAll("li")); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 66, 1, 82],
            fixed:
              'const nodes = Array.from(document.querySelectorAll("li")); nodes.at(-1); nodes.slice(-1)[0]; _.last(nodes);',
          },
          {
            id: 'slice',
            range: [1, 91, 1, 96],
            fixed:
              'const nodes = Array.from(document.querySelectorAll("li")); nodes[nodes.length - 1]; nodes.at(-1); _.last(nodes);',
          },
          {
            id: 'get-last-function',
            range: [1, 105, 1, 111],
            description: '_.last',
            fixed:
              'const nodes = Array.from(document.querySelectorAll("li")); nodes[nodes.length - 1]; nodes.slice(-1)[0]; nodes.at(-1);',
          },
        ],
      },
      {
        code: 'const nodes = condition ? element.children : []; nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes);',
        errors: [
          {
            id: 'negative-index',
            range: [1, 56, 1, 72],
            fixed:
              'const nodes = condition ? element.children : []; nodes.at(-1); nodes.slice(-1)[0]; _.last(nodes);',
          },
          {
            id: 'slice',
            range: [1, 81, 1, 86],
            fixed:
              'const nodes = condition ? element.children : []; nodes[nodes.length - 1]; nodes.at(-1); _.last(nodes);',
          },
          {
            id: 'get-last-function',
            range: [1, 95, 1, 101],
            description: '_.last',
            fixed:
              'const nodes = condition ? element.children : []; nodes[nodes.length - 1]; nodes.slice(-1)[0]; nodes.at(-1);',
          },
        ],
      },
      {
        code: 'const nodes = element.children; _.last(nodes as string[]);',
        mode: 'ts',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 33, 1, 39],
            description: '_.last',
            fixed:
              'const nodes = element.children; (nodes as string[]).at(-1);',
          },
        ],
      },
      {
        code: 'function foo(value: unknown) { if (value instanceof NodeList) { return value[value.length - 1]; } }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 78, 1, 94],
            fixed:
              'function foo(value: unknown) { if (value instanceof NodeList) { return value.at(-1); } }',
          },
        ],
      },
      {
        code: 'function foo(value: any) { value.slice(-1)[0]; _.last(value); }',
        mode: 'ts',
        errors: [
          {
            id: 'slice',
            range: [1, 34, 1, 39],
            fixed: 'function foo(value: any) { value.at(-1); _.last(value); }',
          },
          {
            id: 'get-last-function',
            range: [1, 48, 1, 54],
            description: '_.last',
            fixed:
              'function foo(value: any) { value.slice(-1)[0]; value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: any) { value.slice(-1)[0]; _.last(value); }',
        mode: 'typed',
        errors: [
          {
            id: 'slice',
            range: [1, 34, 1, 39],
            fixed: 'function foo(value: any) { value.at(-1); _.last(value); }',
          },
          {
            id: 'get-last-function',
            range: [1, 48, 1, 54],
            description: '_.last',
            fixed:
              'function foo(value: any) { value.slice(-1)[0]; value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: unknown) { value.slice(-1)[0]; _.last(value); }',
        mode: 'ts',
        errors: [
          {
            id: 'slice',
            range: [1, 38, 1, 43],
            fixed:
              'function foo(value: unknown) { value.at(-1); _.last(value); }',
          },
          {
            id: 'get-last-function',
            range: [1, 52, 1, 58],
            description: '_.last',
            fixed:
              'function foo(value: unknown) { value.slice(-1)[0]; value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: unknown) { value.slice(-1)[0]; _.last(value); }',
        mode: 'typed',
        errors: [
          {
            id: 'slice',
            range: [1, 38, 1, 43],
            fixed:
              'function foo(value: unknown) { value.at(-1); _.last(value); }',
          },
          {
            id: 'get-last-function',
            range: [1, 52, 1, 58],
            description: '_.last',
            fixed:
              'function foo(value: unknown) { value.slice(-1)[0]; value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: NodeList | string[]) { value.slice(-1)[0]; _.last(value); }',
        mode: 'ts',
        errors: [
          {
            id: 'slice',
            range: [1, 50, 1, 55],
            fixed:
              'function foo(value: NodeList | string[]) { value.at(-1); _.last(value); }',
          },
          {
            id: 'get-last-function',
            range: [1, 64, 1, 70],
            description: '_.last',
            fixed:
              'function foo(value: NodeList | string[]) { value.slice(-1)[0]; value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(value: NodeList | string[]) { value.slice(-1)[0]; _.last(value); }',
        mode: 'typed',
        errors: [
          {
            id: 'slice',
            range: [1, 50, 1, 55],
            fixed:
              'function foo(value: NodeList | string[]) { value.at(-1); _.last(value); }',
          },
          {
            id: 'get-last-function',
            range: [1, 64, 1, 70],
            description: '_.last',
            fixed:
              'function foo(value: NodeList | string[]) { value.slice(-1)[0]; value.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo<T extends NodeList | string[]>(nodes: T) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 63, 1, 79],
            fixed:
              'function foo<T extends NodeList | string[]>(nodes: T) { nodes.at(-1); nodes.slice(-1)[0]; _.last(nodes); }',
          },
          {
            id: 'slice',
            range: [1, 88, 1, 93],
            fixed:
              'function foo<T extends NodeList | string[]>(nodes: T) { nodes[nodes.length - 1]; nodes.at(-1); _.last(nodes); }',
          },
          {
            id: 'get-last-function',
            range: [1, 102, 1, 108],
            description: '_.last',
            fixed:
              'function foo<T extends NodeList | string[]>(nodes: T) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; nodes.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo<T extends NodeList | string[]>(nodes: T) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 63, 1, 79],
            fixed:
              'function foo<T extends NodeList | string[]>(nodes: T) { nodes.at(-1); nodes.slice(-1)[0]; _.last(nodes); }',
          },
          {
            id: 'slice',
            range: [1, 88, 1, 93],
            fixed:
              'function foo<T extends NodeList | string[]>(nodes: T) { nodes[nodes.length - 1]; nodes.at(-1); _.last(nodes); }',
          },
          {
            id: 'get-last-function',
            range: [1, 102, 1, 108],
            description: '_.last',
            fixed:
              'function foo<T extends NodeList | string[]>(nodes: T) { nodes[nodes.length - 1]; nodes.slice(-1)[0]; nodes.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(nodes: NodeList | string[] | undefined) { nodes?.[nodes.length - 1]; nodes?.slice(-1)[0]; _.last(nodes); }',
        mode: 'ts',
        errors: [
          {
            id: 'negative-index',
            range: [1, 64, 1, 80],
            fixed:
              'function foo(nodes: NodeList | string[] | undefined) { nodes?.at(-1); nodes?.slice(-1)[0]; _.last(nodes); }',
          },
          {
            id: 'slice',
            range: [1, 90, 1, 95],
            fixed:
              'function foo(nodes: NodeList | string[] | undefined) { nodes?.[nodes.length - 1]; nodes?.at(-1); _.last(nodes); }',
          },
          {
            id: 'get-last-function',
            range: [1, 104, 1, 110],
            description: '_.last',
            fixed:
              'function foo(nodes: NodeList | string[] | undefined) { nodes?.[nodes.length - 1]; nodes?.slice(-1)[0]; nodes.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(nodes: NodeList | string[] | undefined) { nodes?.[nodes.length - 1]; nodes?.slice(-1)[0]; _.last(nodes); }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 64, 1, 80],
            fixed:
              'function foo(nodes: NodeList | string[] | undefined) { nodes?.at(-1); nodes?.slice(-1)[0]; _.last(nodes); }',
          },
          {
            id: 'slice',
            range: [1, 90, 1, 95],
            fixed:
              'function foo(nodes: NodeList | string[] | undefined) { nodes?.[nodes.length - 1]; nodes?.at(-1); _.last(nodes); }',
          },
          {
            id: 'get-last-function',
            range: [1, 104, 1, 110],
            description: '_.last',
            fixed:
              'function foo(nodes: NodeList | string[] | undefined) { nodes?.[nodes.length - 1]; nodes?.slice(-1)[0]; nodes.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(nodes: NodeList | string[]) { if (nodes instanceof NodeList) { return nodes[nodes.length - 1]; } return nodes[nodes.length - 1]; }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 124, 1, 140],
            fixed:
              'function foo(nodes: NodeList | string[]) { if (nodes instanceof NodeList) { return nodes[nodes.length - 1]; } return nodes.at(-1); }',
          },
        ],
      },
      {
        code: 'declare function getNodes(): any; const nodes = getNodes?.(); _.last(nodes);',
        mode: 'ts',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 63, 1, 69],
            description: '_.last',
            fixed:
              'declare function getNodes(): any; const nodes = getNodes?.(); nodes.at(-1);',
          },
        ],
      },
      {
        code: 'declare function getNodes(): unknown; const nodes = getNodes?.(); _.last(nodes);',
        mode: 'ts',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 67, 1, 73],
            description: '_.last',
            fixed:
              'declare function getNodes(): unknown; const nodes = getNodes?.(); nodes.at(-1);',
          },
        ],
      },
      {
        code: 'declare function getNodes(): NodeList | string[]; const nodes = getNodes?.(); _.last(nodes);',
        mode: 'ts',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 79, 1, 85],
            description: '_.last',
            fixed:
              'declare function getNodes(): NodeList | string[]; const nodes = getNodes?.(); nodes.at(-1);',
          },
        ],
      },
      {
        code: 'declare function getNodes(): string[]; const nodes = getNodes?.(); _.last(nodes);',
        mode: 'ts',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 68, 1, 74],
            description: '_.last',
            fixed:
              'declare function getNodes(): string[]; const nodes = getNodes?.(); nodes.at(-1);',
          },
        ],
      },
      {
        code: 'declare function getNodes(): NodeList; const nodes = getNodes?.() as string[]; _.last(nodes);',
        mode: 'ts',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 80, 1, 86],
            description: '_.last',
            fixed:
              'declare function getNodes(): NodeList; const nodes = getNodes?.() as string[]; nodes.at(-1);',
          },
        ],
      },
      {
        code: 'function foo(slot: HTMLSlotElement) { const nodes = slot.assignedElements(); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 84, 1, 100],
            fixed:
              'function foo(slot: HTMLSlotElement) { const nodes = slot.assignedElements(); nodes.at(-1); nodes.slice(-1)[0]; _.last(nodes); }',
          },
          {
            id: 'slice',
            range: [1, 109, 1, 114],
            fixed:
              'function foo(slot: HTMLSlotElement) { const nodes = slot.assignedElements(); nodes[nodes.length - 1]; nodes.at(-1); _.last(nodes); }',
          },
          {
            id: 'get-last-function',
            range: [1, 123, 1, 129],
            description: '_.last',
            fixed:
              'function foo(slot: HTMLSlotElement) { const nodes = slot.assignedElements(); nodes[nodes.length - 1]; nodes.slice(-1)[0]; nodes.at(-1); }',
          },
        ],
      },
      {
        code: 'function foo(form: HTMLFormElement) { const nodes = Array.from(form.elements); nodes[nodes.length - 1]; nodes.slice(-1)[0]; _.last(nodes); }',
        mode: 'typed',
        errors: [
          {
            id: 'negative-index',
            range: [1, 86, 1, 102],
            fixed:
              'function foo(form: HTMLFormElement) { const nodes = Array.from(form.elements); nodes.at(-1); nodes.slice(-1)[0]; _.last(nodes); }',
          },
          {
            id: 'slice',
            range: [1, 111, 1, 116],
            fixed:
              'function foo(form: HTMLFormElement) { const nodes = Array.from(form.elements); nodes[nodes.length - 1]; nodes.at(-1); _.last(nodes); }',
          },
          {
            id: 'get-last-function',
            range: [1, 125, 1, 131],
            description: '_.last',
            fixed:
              'function foo(form: HTMLFormElement) { const nodes = Array.from(form.elements); nodes[nodes.length - 1]; nodes.slice(-1)[0]; nodes.at(-1); }',
          },
        ],
      },
    ],
  },
  {
    name: 'Upstream runtime edit checks',
    valid: [
      'const array = [function () { return this === array; }]; array[array.length - 1]();',
      'const array = [function () { return this.length; }]; array.slice(-1)[0]();',
    ],
    invalid: [
      {
        code: 'function read() {return(_).last(getItems(/* keep */));} read();',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 24, 1, 32],
            description: '_.last',
            fixed:
              'function read() {return getItems(/* keep */).at(-1);} read();',
          },
        ],
      },
      {
        code: 'const array = [class { value = 1; }]; new array[array.length - 1]().value;',
        errors: [
          {
            id: 'negative-index',
            range: [1, 49, 1, 65],
            fixed:
              'const array = [class { value = 1; }]; new (array.at(-1))().value;',
          },
        ],
      },
      {
        code: 'const array = [{Constructor: class { value = 1; }}]; new array[array.length - 1].Constructor().value;',
        errors: [
          {
            id: 'negative-index',
            range: [1, 64, 1, 80],
            fixed:
              'const array = [{Constructor: class { value = 1; }}]; new (array.at(-1)).Constructor().value;',
          },
        ],
      },
      {
        code: 'const array = [{tag() { return class { value = 1; }; }}]; new array[array.length - 1].tag`template`().value;',
        errors: [
          {
            id: 'negative-index',
            range: [1, 69, 1, 85],
            fixed:
              'const array = [{tag() { return class { value = 1; }; }}]; new (array.at(-1)).tag`template`().value;',
          },
        ],
      },
    ],
  },
  {
    name: 'Documentation examples',
    valid: [
      'const foo = array.at(-1);',
      'const foo = array.at(-5);',
      'const foo = string.at(-5);',
      'const foo = string.at(index);',
      'const foo = array[100];',
      'const foo = array.at(array.length - 1);',
      'array[array.length - 1] = foo;',
      'function foo() {\n\treturn arguments[arguments.length - 1];\n}',
      'const foo = element.children[element.children.length - 1];',
      "const elements = document.querySelectorAll('div');\nconst lastElement = elements[elements.length - 1];",
      'const lastElement = Array.from(elements).at(-1);',
      { code: 'const foo = bar[unknownProperty];', options: checkAll },
      { code: 'const foo = string.charAt(-1);', options: checkAll },
      {
        code: 'const foo = bar.at(-1);',
        options: [{ getLastElementFunctions: ['utils.lastElement'] }],
      },
    ],
    invalid: [
      {
        code: 'const foo = array[array.length - 1];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 19, 1, 35],
            fixed: 'const foo = array.at(-1);',
          },
        ],
      },
      {
        code: 'const foo = array.slice(-1)[0];',
        errors: [
          {
            id: 'slice',
            range: [1, 19, 1, 24],
            fixed: 'const foo = array.at(-1);',
          },
        ],
      },
      {
        code: 'const foo = array.slice(-1).pop();',
        errors: [
          {
            id: 'slice',
            range: [1, 19, 1, 24],
            fixed: 'const foo = array.at(-1);',
          },
        ],
      },
      {
        code: 'const foo = array.slice(-1).shift();',
        errors: [
          {
            id: 'slice',
            range: [1, 19, 1, 24],
            fixed: 'const foo = array.at(-1);',
          },
        ],
      },
      {
        code: 'const foo = lodash.last(array);',
        errors: [
          {
            id: 'get-last-function',
            range: [1, 13, 1, 24],
            description: 'lodash.last',
            fixed: 'const foo = array.at(-1);',
          },
        ],
      },
      {
        code: 'const foo = array[array.length - 5];',
        errors: [
          {
            id: 'negative-index',
            range: [1, 19, 1, 35],
            fixed: 'const foo = array.at(-5);',
          },
        ],
      },
      {
        code: 'const foo = string.charAt(string.length - 5);',
        errors: [
          {
            id: 'string-char-at-negative',
            range: [1, 27, 1, 44],
            suggested: 'const foo = string.at(- 5);',
          },
        ],
      },
      {
        code: 'const foo = string.substring(index, index + 1);',
        errors: [
          {
            id: 'string-substring',
            range: [1, 13, 1, 47],
            suggested: 'const foo = string.at(index);',
          },
        ],
      },
      {
        code: 'const lastElement = Array.from(elements).slice(-1)[0];',
        errors: [
          {
            id: 'slice',
            range: [1, 42, 1, 47],
            fixed: 'const lastElement = Array.from(elements).at(-1);',
          },
        ],
      },
      {
        code: 'const foo = bar[10];',
        options: checkAll,
        errors: [
          {
            id: 'index',
            range: [1, 17, 1, 19],
            fixed: 'const foo = bar.at(10);',
          },
        ],
      },
      {
        code: 'const foo = string.charAt(unknownIndex);',
        options: checkAll,
        errors: [
          {
            id: 'string-char-at',
            range: [1, 27, 1, 39],
            suggested: 'const foo = string.at(unknownIndex);',
          },
        ],
      },
      {
        code: 'const foo = utils.lastElement(bar);',
        options: [{ getLastElementFunctions: ['utils.lastElement'] }],
        errors: [
          {
            id: 'get-last-function',
            range: [1, 13, 1, 30],
            description: 'utils.lastElement',
            fixed: 'const foo = bar.at(-1);',
          },
        ],
      },
    ],
  },
];

const messages: Record<string, string> = {
  'negative-index': 'Prefer `.at(…)` over `[….length - index]`.',
  index: 'Prefer `.at(…)` over index access.',
  'string-char-at-negative':
    'Prefer `String#at(…)` over `String#charAt(….length - index)`.',
  'string-substring':
    'Prefer `String#at(…)` over `String#substring(…)` when getting one character.',
  slice: 'Prefer `.at(…)` over the first element from `.slice(…)`.',
  'string-char-at': 'Prefer `String#at(…)` over `String#charAt(…)`.',
};

const ruleName = 'unicorn/prefer-at';
let root: string;
let service: RSLintService;
beforeAll(async () => {
  root = await realpath(
    await mkdtemp(path.join(tmpdir(), 'rslint-prefer-at-')),
  );
  await writeFile(
    path.join(root, 'tsconfig.json'),
    JSON.stringify({
      compilerOptions: {
        target: 'ESNext',
        lib: ['ESNext', 'DOM'],
        strict: true,
      },
      files: ['typed.ts'],
    }),
  );
  service = new RSLintService(
    new NodeRslintService({ workingDirectory: root }),
  );
});
afterAll(async () => {
  await service?.close();
  if (root) {
    await rm(root, {
      recursive: true,
      force: true,
      maxRetries: 20,
      retryDelay: 100,
    });
  }
});

function request(item: Case): LintOptions {
  const filename = path.join(
    root,
    item.mode === 'typed'
      ? 'typed.ts'
      : item.mode === 'ts'
        ? 'file.ts'
        : 'file.js',
  );
  return {
    configDirectory: root,
    workingDirectory: root,
    files: [filename],
    config: [
      {
        plugins: ['unicorn'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            ...(item.mode === 'typed' ? { project: ['./tsconfig.json'] } : {}),
          },
        },
        rules: { [ruleName]: ['error', ...(item.options ?? [])] },
      },
    ],
    fileContents: { [filename]: item.code },
  };
}

// The wire protocol uses UTF-16 offsets, as JavaScript's slice does.
function applyFixes(
  code: string,
  fixes: NonNullable<Diagnostic['fixes']>,
): string {
  const parts: string[] = [];
  let end = 0;
  for (const fix of [...fixes].sort(
    (a, b) => a.startPos - b.startPos || a.endPos - b.endPos,
  )) {
    parts.push(code.slice(end, fix.startPos), fix.text);
    end = fix.endPos;
  }
  parts.push(code.slice(end));
  return parts.join('');
}

describe(ruleName, () => {
  for (const group of groups) {
    test(group.name, async () => {
      for (const value of [...(group.valid ?? []), ...(group.invalid ?? [])]) {
        const item: Case = typeof value === 'string' ? { code: value } : value;
        const result = await service.lint(request(item));
        const expected = item.errors ?? [];
        expect(result.fileCount, item.code).toBe(1);
        expect(result.ruleCount, item.code).toBe(1);
        expect(result.diagnostics, item.code).toHaveLength(expected.length);
        for (const [index, diagnostic] of result.diagnostics.entries()) {
          const error = expected[index];
          const [line, column, endLine, endColumn] = error.range;
          const message =
            error.id === 'get-last-function'
              ? 'Prefer `.at(-1)` over `' +
                error.description +
                '(…)` to get the last element.'
              : messages[error.id];
          expect(diagnostic.ruleName, item.code).toBe(ruleName);
          expect(diagnostic.messageId, item.code).toBe(error.id);
          expect(diagnostic.message, item.code).toBe(message);
          expect(diagnostic.range, item.code).toEqual({
            start: { line, column },
            end: { line: endLine, column: endColumn },
          });
          const fixed = diagnostic.fixes?.length
            ? applyFixes(item.code, diagnostic.fixes)
            : undefined;
          expect(fixed, item.code).toBe(error.fixed);
          const suggestions =
            diagnostic.suggestions?.map((suggestion) => ({
              id: suggestion.messageId,
              message: suggestion.message,
              output: applyFixes(item.code, suggestion.fixes ?? []),
            })) ?? [];
          expect(suggestions, item.code).toEqual(
            error.suggested === undefined
              ? []
              : [
                  {
                    id: 'use-at',
                    message: 'Use `.at(…)`.',
                    output: error.suggested,
                  },
                ],
          );
        }
      }
    });
  }

  test('runtime behavior of constructor, reference, and retained-comment edits', async () => {
    const runtime = groups.find(
      ({ name }) => name === 'Upstream runtime edit checks',
    )!;
    for (const item of [...runtime.invalid!.slice(1), ...runtime.valid!]) {
      const input = typeof item === 'string' ? { code: item } : item;
      const result = await service.lint({ ...request(input), fix: true });
      expect(result.diagnostics).toEqual([]);
      const fixed = Object.values(result.output ?? {})[0] ?? input.code;
      expect(runInNewContext(fixed), input.code).toBe(
        runInNewContext(input.code),
      );
    }
    const item = runtime.invalid![0];
    const result = await service.lint({ ...request(item), fix: true });
    expect(result.diagnostics).toEqual([]);
    const [fixed] = Object.values(result.output ?? {});
    expect(fixed).toBe(item.errors![0].fixed);
    for (const items of [[], [1], [1, 2, 3]]) {
      for (const source of [item.code, fixed]) {
        let calls = 0;
        expect(
          runInNewContext(source, {
            _: { last: (array: number[]) => array[array.length - 1] },
            getItems: () => {
              calls++;
              return items;
            },
          }),
        ).toBe(items.at(-1));
        expect(calls).toBe(1);
      }
    }
  });
});
