import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('capitalized-comments', {
  valid: [
    // No options: capitalization required
    '//Uppercase',
    '// Uppercase',
    '/*Uppercase */',
    '/* Uppercase */',
    '/*\nUppercase */',
    '/** Uppercase */',
    '/**\nUppercase */',
    '//Über',
    '//Π',
    '/* Uppercase\nsecond line need not be uppercase */',

    // No options: skips comments that only contain whitespace
    '// ',
    '//\t',
    '/* */',
    '/*\t*/',
    '/*\n*/',
    '/*\r*/',
    '/*\r\n*/',
    '/* */',
    '/* */',

    // No options: non-alphabetical is okay
    '//123',
    '// 123',
    '/*123*/',
    '/* 123 */',
    '/**123 */',
    '/** 123 */',
    '/**\n123 */',
    '/*\n123 */',
    '/*123\nsecond line need not be uppercase */',
    '/**\n * @fileoverview This is a file */',

    // No options: eslint/istanbul/jshint/jscs/globals?/exported are okay
    '// jscs: enable',
    '// jscs:disable',
    '// eslint-disable-line',
    '// eslint-disable-next-line',
    '/* eslint semi:off */',
    '/* eslint-enable */',
    '/* istanbul ignore next */',
    '/* jshint asi:true */',
    '/* jscs: enable */',
    '/* global var1, var2 */',
    '/* global var1:true, var2 */',
    '/* globals var1, var2 */',
    '/* globals var1:true, var2 */',
    '/* exported myVar */',

    // Ignores shebangs
    '#!foo',
    { code: '#!foo', options: ['always'] as any },
    { code: '#!Foo', options: ['never'] as any },
    '#!/usr/bin/env node',
    { code: '#!/usr/bin/env node', options: ['always'] as any },
    { code: '#!/usr/bin/env node', options: ['never'] as any },

    // Using "always" string option
    { code: '//Uppercase', options: ['always'] as any },
    { code: '// Uppercase', options: ['always'] as any },
    { code: '/*Uppercase */', options: ['always'] as any },
    { code: '/* Uppercase */', options: ['always'] as any },
    { code: '/*\nUppercase */', options: ['always'] as any },
    { code: '/** Uppercase */', options: ['always'] as any },
    { code: '/**\nUppercase */', options: ['always'] as any },
    { code: '//Über', options: ['always'] as any },
    { code: '//Π', options: ['always'] as any },
    {
      code: '/* Uppercase\nsecond line need not be uppercase */',
      options: ['always'] as any,
    },

    // Using "always" string option: non-alphabetical is okay
    { code: '//123', options: ['always'] as any },
    { code: '// 123', options: ['always'] as any },
    { code: '/*123*/', options: ['always'] as any },
    { code: '/**123*/', options: ['always'] as any },
    { code: '/* 123 */', options: ['always'] as any },
    { code: '/** 123*/', options: ['always'] as any },
    { code: '/**\n123*/', options: ['always'] as any },
    { code: '/*\n123 */', options: ['always'] as any },
    {
      code: '/*123\nsecond line need not be uppercase */',
      options: ['always'] as any,
    },
    {
      code: '/**\n @todo: foobar\n */',
      options: ['always'] as any,
    },
    {
      code: '/**\n * @fileoverview This is a file */',
      options: ['always'] as any,
    },

    // Using "always" string option: eslint/istanbul/jshint/jscs/globals?/exported are okay
    { code: '// jscs: enable', options: ['always'] as any },
    { code: '// jscs:disable', options: ['always'] as any },
    { code: '// eslint-disable-line', options: ['always'] as any },
    { code: '// eslint-disable-next-line', options: ['always'] as any },
    { code: '/* eslint semi:off */', options: ['always'] as any },
    { code: '/* eslint-enable */', options: ['always'] as any },
    { code: '/* istanbul ignore next */', options: ['always'] as any },
    { code: '/* jshint asi:true */', options: ['always'] as any },
    { code: '/* jscs: enable */', options: ['always'] as any },
    { code: '/* global var1, var2 */', options: ['always'] as any },
    { code: '/* global var1:true, var2 */', options: ['always'] as any },
    { code: '/* globals var1, var2 */', options: ['always'] as any },
    { code: '/* globals var1:true, var2 */', options: ['always'] as any },
    { code: '/* exported myVar */', options: ['always'] as any },

    // Using "never" string option
    { code: '//lowercase', options: ['never'] as any },
    { code: '// lowercase', options: ['never'] as any },
    { code: '/*lowercase */', options: ['never'] as any },
    { code: '/* lowercase */', options: ['never'] as any },
    { code: '/*\nlowercase */', options: ['never'] as any },
    { code: '//über', options: ['never'] as any },
    { code: '//π', options: ['never'] as any },
    {
      code: '/* lowercase\nSecond line need not be lowercase */',
      options: ['never'] as any,
    },

    // Using "never" string option: non-alphabetical is okay
    { code: '//123', options: ['never'] as any },
    { code: '// 123', options: ['never'] as any },
    { code: '/*123*/', options: ['never'] as any },
    { code: '/* 123 */', options: ['never'] as any },
    { code: '/*\n123 */', options: ['never'] as any },
    {
      code: '/*123\nsecond line need not be uppercase */',
      options: ['never'] as any,
    },
    {
      code: '/**\n @TODO: foobar\n */',
      options: ['never'] as any,
    },
    {
      code: '/**\n * @Fileoverview This is a file */',
      options: ['never'] as any,
    },

    // If first word in comment matches ignorePattern, don't warn
    {
      code: '// matching',
      options: ['always', { ignorePattern: 'match' }] as any,
    },
    {
      code: '// Matching',
      options: ['never', { ignorePattern: 'Match' }] as any,
    },
    {
      code: '// bar',
      options: ['always', { ignorePattern: 'foo|bar' }] as any,
    },
    {
      code: '// Bar',
      options: ['never', { ignorePattern: 'Foo|Bar' }] as any,
    },

    // Inline comments are not warned if ignoreInlineComments: true
    {
      code: 'foo(/* ignored */ a);',
      options: ['always', { ignoreInlineComments: true }] as any,
    },
    {
      code: 'foo(/* Ignored */ a);',
      options: ['never', { ignoreInlineComments: true }] as any,
    },

    // Inline comments can span multiple lines
    {
      code: 'foo(/*\nignored */ a);',
      options: ['always', { ignoreInlineComments: true }] as any,
    },
    {
      code: 'foo(/*\nIgnored */ a);',
      options: ['never', { ignoreInlineComments: true }] as any,
    },

    // Tolerating consecutive comments
    {
      code: [
        '// This comment is valid since it is capitalized,',
        '// and this one is valid since it follows a valid one,',
        '// and same with this one.',
      ].join('\n'),
      options: ['always', { ignoreConsecutiveComments: true }] as any,
    },
    {
      code: [
        '/* This comment is valid since it is capitalized, */',
        '/* and this one is valid since it follows a valid one, */',
        '/* and same with this one. */',
      ].join('\n'),
      options: ['always', { ignoreConsecutiveComments: true }] as any,
    },
    {
      code: [
        '/*',
        ' * This comment is valid since it is capitalized,',
        ' */',
        '/* and this one is valid since it follows a valid one, */',
        '/*',
        ' * and same with this one.',
        ' */',
      ].join('\n'),
      options: ['always', { ignoreConsecutiveComments: true }] as any,
    },
    {
      code: [
        '// This comment is valid since it is capitalized,',
        '// and this one is valid since it follows a valid one,',
        'foo();',
        '// This comment now has to be capitalized.',
      ].join('\n'),
      options: ['always', { ignoreConsecutiveComments: true }] as any,
    },

    // Comments which start with URLs should always be valid
    { code: '// https://github.com', options: ['always'] as any },
    { code: '// HTTPS://GITHUB.COM', options: ['never'] as any },

    // Using different options for line/block comments
    {
      code: [
        '// Valid capitalized line comment',
        '/* Valid capitalized block comment */',
        '// lineCommentIgnorePattern',
        '/* blockCommentIgnorePattern */',
      ].join('\n'),
      options: [
        'always',
        {
          line: { ignorePattern: 'lineCommentIgnorePattern' },
          block: { ignorePattern: 'blockCommentIgnorePattern' },
        },
      ] as any,
    },
  ],

  invalid: [
    // No options: capitalization required
    {
      code: '//lowercase',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '// lowercase',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*lowercase */',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/* lowercase */',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/** lowercase */',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\nlowercase */',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/**\nlowercase */',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '//über',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '//π',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/* lowercase\nSecond line need not be lowercase */',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '// ꮳꮃꭹ',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/* 𐳡𐳡𐳡 */',
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },

    // Using "always" string option
    {
      code: '//lowercase',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '// lowercase',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*lowercase */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/* lowercase */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/** lowercase */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/**\nlowercase */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '//über',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '//π',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/* lowercase\nsecond line need not be uppercase */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },

    // Using "never" string option
    {
      code: '//Uppercase',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },
    {
      code: '// Uppercase',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*Uppercase */',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/* Uppercase */',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\nUppercase */',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },
    {
      code: '//Über',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },
    {
      code: '//Π',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/* Uppercase\nsecond line need not be uppercase */',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },
    {
      code: '// Გ',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },
    {
      code: '// 𑢢',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },

    // Default ignore words should be warned if there are non-whitespace characters in the way
    {
      code: '//* jscs: enable',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '//* jscs:disable',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '//* eslint-disable-line',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '//* eslint-disable-next-line',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\n * eslint semi:off */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\n * eslint-env node */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\n *  istanbul ignore next */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\n *  jshint asi:true */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\n *  jscs: enable */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\n *  global var1, var2 */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\n *  global var1:true, var2 */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\n *  globals var1, var2 */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\n *  globals var1:true, var2 */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '/*\n *  exported myVar */',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },

    // Inline comments should be warned if ignoreInlineComments is omitted or false
    {
      code: 'foo(/* invalid */a);',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 5 }],
    },
    {
      code: 'foo(/* invalid */a);',
      options: ['always', { ignoreInlineComments: false }] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 5 }],
    },

    // ignoreInlineComments should only allow inline comments to pass
    {
      code: 'foo(a, // not an inline comment\nb);',
      options: ['always', { ignoreInlineComments: true }] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 8 }],
    },
    {
      code: 'foo(a, /* not an inline comment */\nb);',
      options: ['always', { ignoreInlineComments: true }] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 8 }],
    },
    {
      code: 'foo(a,\n/* not an inline comment */b);',
      options: ['always', { ignoreInlineComments: true }] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 2, column: 1 }],
    },
    {
      code: 'foo(a,\n/* not an inline comment */\nb);',
      options: ['always', { ignoreInlineComments: true }] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 2, column: 1 }],
    },
    {
      code: 'foo(a, // Not an inline comment\nb);',
      options: ['never', { ignoreInlineComments: true }] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 8 }],
    },
    {
      code: 'foo(a, /* Not an inline comment */\nb);',
      options: ['never', { ignoreInlineComments: true }] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 8 }],
    },
    {
      code: 'foo(a,\n/* Not an inline comment */b);',
      options: ['never', { ignoreInlineComments: true }] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 2, column: 1 }],
    },
    {
      code: 'foo(a,\n/* Not an inline comment */\nb);',
      options: ['never', { ignoreInlineComments: true }] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 2, column: 1 }],
    },

    // Comments which do not match ignorePattern are still warned
    {
      code: '// not matching',
      options: ['always', { ignorePattern: 'ignored?' }] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '// Not matching',
      options: ['never', { ignorePattern: 'ignored?' }] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },

    // ignoreConsecutiveComments only applies to comments with no tokens between them
    {
      code: [
        '// This comment is valid since it is capitalized,',
        '// and this one is valid since it follows a valid one,',
        'foo();',
        '// this comment is now invalid.',
      ].join('\n'),
      options: ['always', { ignoreConsecutiveComments: true }] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 4, column: 1 }],
    },

    // Only the initial comment should warn if ignoreConsecutiveComments:true
    {
      code: [
        '// this comment is invalid since it is not capitalized,',
        '// but this one is ignored since it is consecutive.',
      ].join('\n'),
      options: ['always', { ignoreConsecutiveComments: true }] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: [
        '// This comment is invalid since it is not capitalized,',
        '// But this one is ignored since it is consecutive.',
      ].join('\n'),
      options: ['never', { ignoreConsecutiveComments: true }] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },

    // Consecutive comments should warn if ignoreConsecutiveComments:false
    {
      code: [
        '// This comment is valid since it is capitalized,',
        '// but this one is invalid even if it follows a valid one.',
      ].join('\n'),
      options: ['always', { ignoreConsecutiveComments: false }] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 2, column: 1 }],
    },

    // Comments are warned if URL is not at the start of the comment
    {
      code: '// should fail. https://github.com',
      options: ['always'] as any,
      errors: [{ messageId: 'unexpectedLowercaseComment', line: 1, column: 1 }],
    },
    {
      code: '// Should fail. https://github.com',
      options: ['never'] as any,
      errors: [{ messageId: 'unexpectedUppercaseComment', line: 1, column: 1 }],
    },
  ],
});
