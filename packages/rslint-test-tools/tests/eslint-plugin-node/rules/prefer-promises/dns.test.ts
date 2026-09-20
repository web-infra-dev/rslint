// Every upstream test and documentation example from eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-promises/dns.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-promises/dns.md
import { RuleTester } from '../../rule-tester';

new RuleTester({
  languageOptions: {
    sourceType: 'module',
    globals: { require: 'readonly', process: 'readonly' },
  },
  fixtureFiles: { 'package.json': '{}' },
}).run(
  'prefer-promises/dns',
  {},
  {
    valid: [
      {
        name: 'Upstream valid 1',
        code: "const dns = require('dns'); dns.lookupSync()",
      },
      {
        name: 'Upstream valid 2',
        code: "const dns = require('dns'); dns.promises.lookup()",
      },
      {
        name: 'Upstream valid 3',
        code: "const dns = require('node:dns'); dns.promises.lookup()",
      },
      {
        name: 'Upstream valid 4',
        code: "const {promises} = require('dns'); promises.lookup()",
      },
      {
        name: 'Upstream valid 5',
        code: "const {promises: dns} = require('dns'); dns.lookup()",
      },
      {
        name: 'Upstream valid 6',
        code: "const {promises: {lookup}} = require('dns'); lookup()",
      },
      {
        name: 'Upstream valid 7',
        code: "import dns from 'dns'; dns.promises.lookup()",
      },
      {
        name: 'Upstream valid 8',
        code: "import dns from 'node:dns'; dns.promises.lookup()",
      },
      {
        name: 'Upstream valid 9',
        code: "import * as dns from 'dns'; dns.promises.lookup()",
      },
      {
        name: 'Upstream valid 10',
        code: "import {promises} from 'dns'; promises.lookup()",
      },
      {
        name: 'Upstream valid 11',
        code: "import {promises as dns} from 'dns'; dns.lookup()",
      },
      {
        name: 'Upstream valid 12',
        code: "const dns = process.getBuiltinModule('dns'); dns.promises.lookup()",
      },
      {
        name: 'Upstream valid 13',
        code: "const dns = process.getBuiltinModule('node:dns'); dns.promises.lookup()",
      },
      {
        name: 'Upstream valid 14',
        code: "const {promises} = process.getBuiltinModule('dns'); promises.lookup()",
      },
      {
        name: 'Upstream valid 15',
        code: "const {promises: dns} = process.getBuiltinModule('dns'); dns.lookup()",
      },
      {
        name: 'Documentation example 3',
        code: 'const { promises: dns } = require("dns")\n\nasync function lookup(hostname) {\n    const { address, family } = await dns.lookup(hostname)\n    //...\n}',
      },
      {
        name: 'Documentation example 4',
        code: 'import { promises as dns } from "dns"\n\nasync function lookup(hostname) {\n    const { address, family } = await dns.lookup(hostname)\n    //...\n}',
      },
    ],
    invalid: [
      {
        name: 'Upstream invalid 1',
        code: "const dns = require('dns'); dns.lookup()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 41,
          },
        ],
      },
      {
        name: 'Upstream invalid 2',
        code: "const dns = require('node:dns'); dns.lookup()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 1,
            column: 34,
            endLine: 1,
            endColumn: 46,
          },
        ],
      },
      {
        name: 'Upstream invalid 3',
        code: "const {lookup} = require('dns'); lookup()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 1,
            column: 34,
            endLine: 1,
            endColumn: 42,
          },
        ],
      },
      {
        name: 'Upstream invalid 4',
        code: "import dns from 'dns'; dns.lookup()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 1,
            column: 24,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'Upstream invalid 5',
        code: "import dns from 'node:dns'; dns.lookup()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 41,
          },
        ],
      },
      {
        name: 'Upstream invalid 6',
        code: "import * as dns from 'dns'; dns.lookup()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 41,
          },
        ],
      },
      {
        name: 'Upstream invalid 7',
        code: "import {lookup} from 'dns'; lookup()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'Upstream invalid 8',
        code: "const dns = process.getBuiltinModule('dns'); dns.lookup()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 1,
            column: 46,
            endLine: 1,
            endColumn: 58,
          },
        ],
      },
      {
        name: 'Upstream invalid 9',
        code: "const dns = process.getBuiltinModule('node:dns'); dns.lookup()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 1,
            column: 51,
            endLine: 1,
            endColumn: 63,
          },
        ],
      },
      {
        name: 'Upstream invalid 10',
        code: "const {lookup} = process.getBuiltinModule('dns'); lookup()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 1,
            column: 51,
            endLine: 1,
            endColumn: 59,
          },
        ],
      },
      {
        name: 'Upstream invalid 11',
        code: "const dns = require('dns'); dns.lookupService()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookupService()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 48,
          },
        ],
      },
      {
        name: 'Upstream invalid 12',
        code: "const dns = require('dns'); new dns.Resolver()",
        errors: [
          {
            messageId: 'preferPromisesNew',
            message: "Use 'new dns.promises.Resolver()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 47,
          },
        ],
      },
      {
        name: 'Upstream invalid 13',
        code: "const dns = require('dns'); dns.getServers()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.getServers()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'Upstream invalid 14',
        code: "const dns = require('dns'); dns.resolve()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolve()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 42,
          },
        ],
      },
      {
        name: 'Upstream invalid 15',
        code: "const dns = require('dns'); dns.resolve4()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolve4()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 43,
          },
        ],
      },
      {
        name: 'Upstream invalid 16',
        code: "const dns = require('dns'); dns.resolve6()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolve6()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 43,
          },
        ],
      },
      {
        name: 'Upstream invalid 17',
        code: "const dns = require('dns'); dns.resolveAny()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolveAny()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'Upstream invalid 18',
        code: "const dns = require('dns'); dns.resolveCname()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolveCname()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 47,
          },
        ],
      },
      {
        name: 'Upstream invalid 19',
        code: "const dns = require('dns'); dns.resolveMx()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolveMx()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'Upstream invalid 20',
        code: "const dns = require('dns'); dns.resolveNaptr()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolveNaptr()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 47,
          },
        ],
      },
      {
        name: 'Upstream invalid 21',
        code: "const dns = require('dns'); dns.resolveNs()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolveNs()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'Upstream invalid 22',
        code: "const dns = require('dns'); dns.resolvePtr()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolvePtr()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'Upstream invalid 23',
        code: "const dns = require('dns'); dns.resolveSoa()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolveSoa()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'Upstream invalid 24',
        code: "const dns = require('dns'); dns.resolveSrv()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolveSrv()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'Upstream invalid 25',
        code: "const dns = require('dns'); dns.resolveTxt()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.resolveTxt()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'Upstream invalid 26',
        code: "const dns = require('dns'); dns.reverse()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.reverse()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 42,
          },
        ],
      },
      {
        name: 'Upstream invalid 27',
        code: "const dns = require('dns'); dns.setServers()",
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.setServers()' instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'Documentation example 1',
        code: 'const dns = require("dns")\n\nfunction lookup(hostname) {\n    dns.lookup(hostname, (error, address, family) => {\n        //...\n    })\n}',
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 4,
            column: 5,
            endLine: 6,
            endColumn: 7,
          },
        ],
      },
      {
        name: 'Documentation example 2',
        code: 'import dns from "dns"\n\nfunction lookup(hostname) {\n    dns.lookup(hostname, (error, address, family) => {\n        //...\n    })\n}',
        errors: [
          {
            messageId: 'preferPromises',
            message: "Use 'dns.promises.lookup()' instead.",
            line: 4,
            column: 5,
            endLine: 6,
            endColumn: 7,
          },
        ],
      },
    ],
  },
);
