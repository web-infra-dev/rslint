package no_noninteractive_element_to_interactive_role

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError mirrors upstream's `expectedError` constant — the single
// shape every invalid case in the rule produces. Centralized so a future
// error-text tweak touches one place.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "noNoninteractiveElementToInteractiveRole",
	Message:   errorMessage,
}

// componentsSettings mirrors upstream's
//
//	settings: { 'jsx-a11y': { components: { Article: 'article', Input: 'input' } } }
//
// so `<Article>` resolves to `article` (non-interactive) and `<Input>`
// resolves to `input` (interactive).
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Article": "article",
			"Input":   "input",
		},
	},
}

// recommendedConfig mirrors `configs.recommended.rules['jsx-a11y/no-noninteractive-element-to-interactive-role'][1]`
// from eslint-plugin-jsx-a11y/src/index.js.
var recommendedConfig = map[string]interface{}{
	"ul":       []interface{}{"listbox", "menu", "menubar", "radiogroup", "tablist", "tree", "treegrid"},
	"ol":       []interface{}{"listbox", "menu", "menubar", "radiogroup", "tablist", "tree", "treegrid"},
	"li":       []interface{}{"menuitem", "menuitemradio", "menuitemcheckbox", "option", "row", "tab", "treeitem"},
	"table":    []interface{}{"grid"},
	"td":       []interface{}{"gridcell"},
	"fieldset": []interface{}{"radiogroup", "presentation"},
}

// alwaysValid mirrors upstream's `alwaysValid` array verbatim — valid under
// EVERY option combination (recommended AND strict). Order matches the
// upstream file so a future audit can grep across both side-by-side.
func alwaysValid() []rule_tester.ValidTestCase {
	return []rule_tester.ValidTestCase{
		// ---- Custom components — element type isn't in aria-query's
		//      `dom` map → upstream short-circuits before the report. ----
		{Code: `<TestComponent onClick={doFoo} />`, Tsx: true},
		{Code: `<Button onClick={doFoo} />`, Tsx: true},

		// ---- Interactive elements with an interactive role. ----
		{Code: `<a tabIndex="0" role="button" />`, Tsx: true},
		{Code: `<a href="http://x.y.z" role="button" />`, Tsx: true},
		{Code: `<a href="http://x.y.z" tabIndex="0" role="button" />`, Tsx: true},
		{Code: `<area role="button" />;`, Tsx: true},
		{Code: `<area role="menuitem" />;`, Tsx: true},
		{Code: `<button className="foo" role="button" />`, Tsx: true},
		{Code: `<body role="button" />;`, Tsx: true},
		{Code: `<frame role="button" />;`, Tsx: true},
		{Code: `<td role="button" />;`, Tsx: true},
		{Code: `<frame role="menuitem" />;`, Tsx: true},
		{Code: `<td role="menuitem" />;`, Tsx: true},

		// ---- All flavors of input with role="button". ----
		{Code: `<input role="button" />`, Tsx: true},
		{Code: `<input type="button" role="button" />`, Tsx: true},
		{Code: `<input type="checkbox" role="button" />`, Tsx: true},
		{Code: `<input type="color" role="button" />`, Tsx: true},
		{Code: `<input type="date" role="button" />`, Tsx: true},
		{Code: `<input type="datetime" role="button" />`, Tsx: true},
		{Code: `<input type="datetime-local" role="button" />`, Tsx: true},
		{Code: `<input type="email" role="button" />`, Tsx: true},
		{Code: `<input type="file" role="button" />`, Tsx: true},
		{Code: `<input type="hidden" role="button" />`, Tsx: true},
		{Code: `<input type="image" role="button" />`, Tsx: true},
		{Code: `<input type="month" role="button" />`, Tsx: true},
		{Code: `<input type="number" role="button" />`, Tsx: true},
		{Code: `<input type="password" role="button" />`, Tsx: true},
		{Code: `<input type="radio" role="button" />`, Tsx: true},
		{Code: `<input type="range" role="button" />`, Tsx: true},
		{Code: `<input type="reset" role="button" />`, Tsx: true},
		{Code: `<input type="search" role="button" />`, Tsx: true},
		{Code: `<input type="submit" role="button" />`, Tsx: true},
		{Code: `<input type="tel" role="button" />`, Tsx: true},
		{Code: `<input type="text" role="button" />`, Tsx: true},
		{Code: `<input type="time" role="button" />`, Tsx: true},
		{Code: `<input type="url" role="button" />`, Tsx: true},
		{Code: `<input type="week" role="button" />`, Tsx: true},
		{Code: `<input type="hidden" role="img" />`, Tsx: true},

		// ---- End all flavors of input. ----
		{Code: `<menuitem role="button" />;`, Tsx: true},
		{Code: `<option className="foo" role="button" />`, Tsx: true},
		{Code: `<select className="foo" role="button" />`, Tsx: true},
		{Code: `<textarea className="foo" role="button" />`, Tsx: true},
		{Code: `<tr role="button" />;`, Tsx: true},
		{Code: `<tr role="presentation" />;`, Tsx: true},

		// ---- Interactive elements with role="img". ----
		{Code: `<a tabIndex="0" role="img" />`, Tsx: true},
		{Code: `<a href="http://x.y.z" role="img" />`, Tsx: true},
		{Code: `<a href="http://x.y.z" tabIndex="0" role="img" />`, Tsx: true},

		// ---- All flavors of input with role="img". ----
		{Code: `<input role="img" />`, Tsx: true},
		{Code: `<input type="img" role="img" />`, Tsx: true},
		{Code: `<input type="checkbox" role="img" />`, Tsx: true},
		{Code: `<input type="color" role="img" />`, Tsx: true},
		{Code: `<input type="date" role="img" />`, Tsx: true},
		{Code: `<input type="datetime" role="img" />`, Tsx: true},
		{Code: `<input type="datetime-local" role="img" />`, Tsx: true},
		{Code: `<input type="email" role="img" />`, Tsx: true},
		{Code: `<input type="file" role="img" />`, Tsx: true},
		{Code: `<input type="hidden" role="button" />`, Tsx: true},
		{Code: `<input type="image" role="img" />`, Tsx: true},
		{Code: `<input type="month" role="img" />`, Tsx: true},
		{Code: `<input type="number" role="img" />`, Tsx: true},
		{Code: `<input type="password" role="img" />`, Tsx: true},
		{Code: `<input type="radio" role="img" />`, Tsx: true},
		{Code: `<input type="range" role="img" />`, Tsx: true},
		{Code: `<input type="reset" role="img" />`, Tsx: true},
		{Code: `<input type="search" role="img" />`, Tsx: true},
		{Code: `<input type="submit" role="img" />`, Tsx: true},
		{Code: `<input type="tel" role="img" />`, Tsx: true},
		{Code: `<input type="text" role="img" />`, Tsx: true},
		{Code: `<input type="time" role="img" />`, Tsx: true},
		{Code: `<input type="url" role="img" />`, Tsx: true},
		{Code: `<input type="week" role="img" />`, Tsx: true},

		// ---- End all flavors of input. ----
		{Code: `<menuitem role="img" />;`, Tsx: true},
		{Code: `<option className="foo" role="img" />`, Tsx: true},
		{Code: `<select className="foo" role="img" />`, Tsx: true},
		{Code: `<textarea className="foo" role="img" />`, Tsx: true},
		{Code: `<tr role="img" />;`, Tsx: true},

		// ---- Interactive elements + role="listitem" (listitem is non-interactive). ----
		{Code: `<a tabIndex="0" role="listitem" />`, Tsx: true},
		{Code: `<a href="http://x.y.z" role="listitem" />`, Tsx: true},
		{Code: `<a href="http://x.y.z" tabIndex="0" role="listitem" />`, Tsx: true},

		// ---- All flavors of input with role="listitem". ----
		{Code: `<input role="listitem" />`, Tsx: true},
		{Code: `<input type="listitem" role="listitem" />`, Tsx: true},
		{Code: `<input type="checkbox" role="listitem" />`, Tsx: true},
		{Code: `<input type="color" role="listitem" />`, Tsx: true},
		{Code: `<input type="date" role="listitem" />`, Tsx: true},
		{Code: `<input type="datetime" role="listitem" />`, Tsx: true},
		{Code: `<input type="datetime-local" role="listitem" />`, Tsx: true},
		{Code: `<input type="email" role="listitem" />`, Tsx: true},
		{Code: `<input type="file" role="listitem" />`, Tsx: true},
		{Code: `<input type="image" role="listitem" />`, Tsx: true},
		{Code: `<input type="month" role="listitem" />`, Tsx: true},
		{Code: `<input type="number" role="listitem" />`, Tsx: true},
		{Code: `<input type="password" role="listitem" />`, Tsx: true},
		{Code: `<input type="radio" role="listitem" />`, Tsx: true},
		{Code: `<input type="range" role="listitem" />`, Tsx: true},
		{Code: `<input type="reset" role="listitem" />`, Tsx: true},
		{Code: `<input type="search" role="listitem" />`, Tsx: true},
		{Code: `<input type="submit" role="listitem" />`, Tsx: true},
		{Code: `<input type="tel" role="listitem" />`, Tsx: true},
		{Code: `<input type="text" role="listitem" />`, Tsx: true},
		{Code: `<input type="time" role="listitem" />`, Tsx: true},
		{Code: `<input type="url" role="listitem" />`, Tsx: true},
		{Code: `<input type="week" role="listitem" />`, Tsx: true},

		// ---- End all flavors of input. ----
		{Code: `<menuitem role="listitem" />;`, Tsx: true},
		{Code: `<option className="foo" role="listitem" />`, Tsx: true},
		{Code: `<select className="foo" role="listitem" />`, Tsx: true},
		{Code: `<textarea className="foo" role="listitem" />`, Tsx: true},
		{Code: `<tr role="listitem" />;`, Tsx: true},

		// ---- HTML elements with neither an interactive nor non-interactive
		//      valence (static). Element gate fails — not non-interactive. ----
		{Code: `<acronym role="button" />;`, Tsx: true},
		{Code: `<applet role="button" />;`, Tsx: true},
		{Code: `<audio role="button" />;`, Tsx: true},
		{Code: `<b role="button" />;`, Tsx: true},
		{Code: `<base role="button" />;`, Tsx: true},
		{Code: `<bdi role="button" />;`, Tsx: true},
		{Code: `<bdo role="button" />;`, Tsx: true},
		{Code: `<big role="button" />;`, Tsx: true},
		{Code: `<blink role="button" />;`, Tsx: true},
		{Code: `<canvas role="button" />;`, Tsx: true},
		{Code: `<center role="button" />;`, Tsx: true},
		{Code: `<cite role="button" />;`, Tsx: true},
		{Code: `<col role="button" />;`, Tsx: true},
		{Code: `<colgroup role="button" />;`, Tsx: true},
		{Code: `<content role="button" />;`, Tsx: true},
		{Code: `<data role="button" />;`, Tsx: true},
		{Code: `<datalist role="button" />;`, Tsx: true},
		{Code: `<div role="button" />;`, Tsx: true},
		{Code: `<div className="foo" role="button" />;`, Tsx: true},
		{Code: `<div className="foo" {...props} role="button" />;`, Tsx: true},
		{Code: `<div aria-hidden role="button" />;`, Tsx: true},
		{Code: `<div aria-hidden={true} role="button" />;`, Tsx: true},
		{Code: `<div role="button" />;`, Tsx: true},
		{Code: `<div role={undefined} role="button" />;`, Tsx: true},
		{Code: `<div {...props} role="button" />;`, Tsx: true},
		{Code: `<div onKeyUp={() => void 0} aria-hidden={false} role="button" />;`, Tsx: true},
		{Code: `<embed role="button" />;`, Tsx: true},
		{Code: `<font role="button" />;`, Tsx: true},
		{Code: `<frameset role="button" />;`, Tsx: true},
		{Code: `<head role="button" />;`, Tsx: true},
		// `<header>` is hard-coded to return false from isNonInteractiveElement
		// upstream (parent-context dependent — banner landmark only inside
		// <body>); the role check therefore never fires.
		{Code: `<header role="button" />;`, Tsx: true},
		{Code: `<hgroup role="button" />;`, Tsx: true},
		{Code: `<i role="button" />;`, Tsx: true},
		{Code: `<kbd role="button" />;`, Tsx: true},
		{Code: `<keygen role="button" />;`, Tsx: true},
		{Code: `<link role="button" />;`, Tsx: true},
		{Code: `<map role="button" />;`, Tsx: true},
		{Code: `<meta role="button" />;`, Tsx: true},
		{Code: `<noembed role="button" />;`, Tsx: true},
		{Code: `<noscript role="button" />;`, Tsx: true},
		{Code: `<object role="button" />;`, Tsx: true},
		{Code: `<param role="button" />;`, Tsx: true},
		{Code: `<picture role="button" />;`, Tsx: true},
		{Code: `<q role="button" />;`, Tsx: true},
		{Code: `<rp role="button" />;`, Tsx: true},
		{Code: `<rt role="button" />;`, Tsx: true},
		{Code: `<rtc role="button" />;`, Tsx: true},
		{Code: `<s role="button" />;`, Tsx: true},
		{Code: `<samp role="button" />;`, Tsx: true},
		{Code: `<script role="button" />;`, Tsx: true},
		{Code: `<small role="button" />;`, Tsx: true},
		{Code: `<source role="button" />;`, Tsx: true},
		{Code: `<spacer role="button" />;`, Tsx: true},
		{Code: `<span role="button" />;`, Tsx: true},
		{Code: `<strike role="button" />;`, Tsx: true},
		{Code: `<style role="button" />;`, Tsx: true},
		{Code: `<summary role="button" />;`, Tsx: true},
		{Code: `<th role="button" />;`, Tsx: true},
		{Code: `<title role="button" />;`, Tsx: true},
		{Code: `<track role="button" />;`, Tsx: true},
		{Code: `<tt role="button" />;`, Tsx: true},
		{Code: `<u role="button" />;`, Tsx: true},
		{Code: `<var role="button" />;`, Tsx: true},
		{Code: `<video role="button" />;`, Tsx: true},
		{Code: `<wbr role="button" />;`, Tsx: true},
		{Code: `<xmp role="button" />;`, Tsx: true},

		// ---- HTML elements attributed with an interactive role
		//      (div is generic, not non-interactive → no report). ----
		{Code: `<div role="button" />;`, Tsx: true},
		{Code: `<div role="checkbox" />;`, Tsx: true},
		{Code: `<div role="columnheader" />;`, Tsx: true},
		{Code: `<div role="combobox" />;`, Tsx: true},
		{Code: `<div role="grid" />;`, Tsx: true},
		{Code: `<div role="gridcell" />;`, Tsx: true},
		{Code: `<div role="link" />;`, Tsx: true},
		{Code: `<div role="listbox" />;`, Tsx: true},
		{Code: `<div role="menu" />;`, Tsx: true},
		{Code: `<div role="menubar" />;`, Tsx: true},
		{Code: `<div role="menuitem" />;`, Tsx: true},
		{Code: `<div role="menuitemcheckbox" />;`, Tsx: true},
		{Code: `<div role="menuitemradio" />;`, Tsx: true},
		{Code: `<div role="option" />;`, Tsx: true},
		{Code: `<div role="progressbar" />;`, Tsx: true},
		{Code: `<div role="radio" />;`, Tsx: true},
		{Code: `<div role="radiogroup" />;`, Tsx: true},
		{Code: `<div role="row" />;`, Tsx: true},
		{Code: `<div role="rowheader" />;`, Tsx: true},
		{Code: `<div role="searchbox" />;`, Tsx: true},
		{Code: `<div role="slider" />;`, Tsx: true},
		{Code: `<div role="spinbutton" />;`, Tsx: true},
		{Code: `<div role="switch" />;`, Tsx: true},
		{Code: `<div role="tab" />;`, Tsx: true},
		{Code: `<div role="textbox" />;`, Tsx: true},
		{Code: `<div role="treeitem" />;`, Tsx: true},

		// ---- Presentation role on a div (div is not non-interactive →
		//      no report regardless). ----
		{Code: `<div role="presentation" />;`, Tsx: true},

		// ---- HTML elements attributed with an abstract role
		//      (IsInteractiveRole returns false for abstract → no report). ----
		{Code: `<div role="command" />;`, Tsx: true},
		{Code: `<div role="composite" />;`, Tsx: true},
		{Code: `<div role="input" />;`, Tsx: true},
		{Code: `<div role="landmark" />;`, Tsx: true},
		{Code: `<div role="range" />;`, Tsx: true},
		{Code: `<div role="roletype" />;`, Tsx: true},
		{Code: `<div role="section" />;`, Tsx: true},
		{Code: `<div role="sectionhead" />;`, Tsx: true},
		{Code: `<div role="select" />;`, Tsx: true},
		{Code: `<div role="structure" />;`, Tsx: true},
		{Code: `<div role="tablist" />;`, Tsx: true},
		{Code: `<div role="toolbar" />;`, Tsx: true},
		{Code: `<div role="tree" />;`, Tsx: true},
		{Code: `<div role="treegrid" />;`, Tsx: true},
		{Code: `<div role="widget" />;`, Tsx: true},
		{Code: `<div role="window" />;`, Tsx: true},

		// ---- HTML elements with an inherent non-interactive role,
		//      assigned a non-interactive (listitem) role. Element gate
		//      passes, but role gate (interactive?) fails → no report. ----
		{Code: `<main role="listitem" />;`, Tsx: true},
		{Code: `<a role="listitem" />`, Tsx: true},
		{Code: `<a role="listitem" />;`, Tsx: true},
		{Code: `<a role="button" />`, Tsx: true},
		{Code: `<a role="button" />;`, Tsx: true},
		{Code: `<a role="menuitem" />`, Tsx: true},
		{Code: `<a role="menuitem" />;`, Tsx: true},
		{Code: `<area role="listitem" />;`, Tsx: true},
		{Code: `<article role="listitem" />;`, Tsx: true},
		{Code: `<article role="listitem" />;`, Tsx: true},
		{Code: `<dd role="listitem" />;`, Tsx: true},
		{Code: `<dfn role="listitem" />;`, Tsx: true},
		{Code: `<dt role="listitem" />;`, Tsx: true},
		{Code: `<fieldset role="listitem" />;`, Tsx: true},
		{Code: `<figure role="listitem" />;`, Tsx: true},
		{Code: `<form role="listitem" />;`, Tsx: true},
		{Code: `<frame role="listitem" />;`, Tsx: true},
		{Code: `<h1 role="listitem" />;`, Tsx: true},
		{Code: `<h2 role="listitem" />;`, Tsx: true},
		{Code: `<h3 role="listitem" />;`, Tsx: true},
		{Code: `<h4 role="listitem" />;`, Tsx: true},
		{Code: `<h5 role="listitem" />;`, Tsx: true},
		{Code: `<h6 role="listitem" />;`, Tsx: true},
		{Code: `<hr role="listitem" />;`, Tsx: true},
		{Code: `<img role="listitem" />;`, Tsx: true},
		{Code: `<li role="listitem" />;`, Tsx: true},
		{Code: `<li role="presentation" />;`, Tsx: true},
		{Code: `<nav role="listitem" />;`, Tsx: true},
		{Code: `<ol role="listitem" />;`, Tsx: true},
		{Code: `<table role="listitem" />;`, Tsx: true},
		{Code: `<tbody role="listitem" />;`, Tsx: true},
		{Code: `<td role="listitem" />;`, Tsx: true},
		{Code: `<tfoot role="listitem" />;`, Tsx: true},
		{Code: `<thead role="listitem" />;`, Tsx: true},
		{Code: `<ul role="listitem" />;`, Tsx: true},

		// ---- HTML elements attributed with a non-interactive role
		//      (div is generic → element gate fails). ----
		{Code: `<div role="alert" />;`, Tsx: true},
		{Code: `<div role="alertdialog" />;`, Tsx: true},
		{Code: `<div role="application" />;`, Tsx: true},
		{Code: `<div role="article" />;`, Tsx: true},
		{Code: `<div role="banner" />;`, Tsx: true},
		{Code: `<div role="cell" />;`, Tsx: true},
		{Code: `<div role="complementary" />;`, Tsx: true},
		{Code: `<div role="contentinfo" />;`, Tsx: true},
		{Code: `<div role="definition" />;`, Tsx: true},
		{Code: `<div role="dialog" />;`, Tsx: true},
		{Code: `<div role="directory" />;`, Tsx: true},
		{Code: `<div role="document" />;`, Tsx: true},
		{Code: `<div role="feed" />;`, Tsx: true},
		{Code: `<div role="figure" />;`, Tsx: true},
		{Code: `<div role="form" />;`, Tsx: true},
		{Code: `<div role="group" />;`, Tsx: true},
		{Code: `<div role="heading" />;`, Tsx: true},
		{Code: `<div role="img" />;`, Tsx: true},
		{Code: `<div role="list" />;`, Tsx: true},
		{Code: `<div role="listitem" />;`, Tsx: true},
		{Code: `<div role="log" />;`, Tsx: true},
		{Code: `<div role="main" />;`, Tsx: true},
		{Code: `<div role="marquee" />;`, Tsx: true},
		{Code: `<div role="math" />;`, Tsx: true},
		{Code: `<div role="navigation" />;`, Tsx: true},
		{Code: `<div role="note" />;`, Tsx: true},
		{Code: `<div role="region" />;`, Tsx: true},
		{Code: `<div role="rowgroup" />;`, Tsx: true},
		{Code: `<div role="search" />;`, Tsx: true},
		{Code: `<div role="separator" />;`, Tsx: true},
		{Code: `<div role="scrollbar" />;`, Tsx: true},
		{Code: `<div role="status" />;`, Tsx: true},
		{Code: `<div role="table" />;`, Tsx: true},
		{Code: `<div role="tabpanel" />;`, Tsx: true},
		{Code: `<div role="term" />;`, Tsx: true},
		{Code: `<div role="timer" />;`, Tsx: true},
		{Code: `<div role="tooltip" />;`, Tsx: true},
		{Code: `<ul role="list" />;`, Tsx: true},

		// ---- Custom components — non-DOM tags exempt regardless of role. ----
		{Code: `<Article role="button" />`, Tsx: true},
		{Code: `<Input role="button" />`, Tsx: true, Settings: componentsSettings},
	}
}

// neverValid mirrors upstream's `neverValid` array verbatim — invalid under
// every option combination (recommended OR strict). Order matches the
// upstream file.
func neverValid() []rule_tester.InvalidTestCase {
	return []rule_tester.InvalidTestCase{
		// ---- HTML elements with an inherent non-interactive role, assigned
		//      an interactive role. ----
		{Code: `<address role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<article role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<aside role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<blockquote role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<br role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<caption role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<code role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dd role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<del role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<details role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dfn role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dir role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dl role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dt role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<em role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<fieldset role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<figcaption role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<figure role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<footer role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<form role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h1 role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h2 role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h3 role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h4 role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h5 role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h6 role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<hr role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<html role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<iframe role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ins role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<label role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<legend role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<li role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<main role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<mark role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<marquee role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<menu role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<meter role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<nav role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ol role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<optgroup role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<output role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<pre role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<progress role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ruby role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<strong role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<sub role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<sup role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<table role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<tbody role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<tfoot role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<thead role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<time role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ul role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Non-interactive elements + role="menuitem". ----
		{Code: `<main role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<article role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dd role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dfn role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<dt role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<fieldset role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<figure role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<form role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h1 role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h2 role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h3 role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h4 role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h5 role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<h6 role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<hr role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<nav role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ol role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<p role="button" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<section role="button" aria-label="Aardvark" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<table role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<tbody role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<tfoot role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<thead role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Custom component resolved to `article` via settings. ----
		{Code: `<Article role="button" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}, Settings: componentsSettings},
	}
}

// recommendedStrictDelta mirrors the cases that upstream's `:strict`
// suite appends on top of `neverValid` but the `:recommended` suite
// keeps in the valid set — `<ul role="menu" />` etc. fall under the
// recommendedConfig allow-list.
func recommendedStrictDelta() []rule_tester.InvalidTestCase {
	return []rule_tester.InvalidTestCase{
		{Code: `<ul role="menu" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ul role="menubar" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ul role="radiogroup" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ul role="tablist" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ul role="tree" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ul role="treegrid" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ol role="menu" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ol role="menubar" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ol role="radiogroup" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ol role="tablist" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ol role="tree" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<ol role="treegrid" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<li role="tab" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<li role="menuitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<li role="row" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<li role="treeitem" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	}
}

// recommendedExtraValid mirrors the additional valid cases that upstream's
// `:recommended` suite layers on top of `alwaysValid` — list / li / fieldset
// / `<Component>` combinations the recommendedConfig allow-list exempts.
// `<li role="menuitemcheckbox">` / `<li role="menuitemradio">` appear here
// (allowed by the current upstream preset) even though upstream's test file
// hard-codes the older preset without those entries; we mirror the
// preset-source-of-truth so a future preset change auto-fixes the test.
func recommendedExtraValid() []rule_tester.ValidTestCase {
	return []rule_tester.ValidTestCase{
		{Code: `<ul role="menu" />;`, Tsx: true},
		{Code: `<ul role="menubar" />;`, Tsx: true},
		{Code: `<ul role="radiogroup" />;`, Tsx: true},
		{Code: `<ul role="tablist" />;`, Tsx: true},
		{Code: `<ul role="tree" />;`, Tsx: true},
		{Code: `<ul role="treegrid" />;`, Tsx: true},
		{Code: `<ol role="menu" />;`, Tsx: true},
		{Code: `<ol role="menubar" />;`, Tsx: true},
		{Code: `<ol role="radiogroup" />;`, Tsx: true},
		{Code: `<ol role="tablist" />;`, Tsx: true},
		{Code: `<ol role="tree" />;`, Tsx: true},
		{Code: `<ol role="treegrid" />;`, Tsx: true},
		{Code: `<li role="tab" />;`, Tsx: true},
		{Code: `<li role="menuitem" />;`, Tsx: true},
		{Code: `<li role="menuitemcheckbox" />;`, Tsx: true},
		{Code: `<li role="menuitemradio" />;`, Tsx: true},
		{Code: `<li role="row" />;`, Tsx: true},
		{Code: `<li role="treeitem" />;`, Tsx: true},
		{Code: `<Component role="treeitem" />;`, Tsx: true},
		{Code: `<fieldset role="radiogroup" />;`, Tsx: true},
		{Code: `<fieldset role="presentation" />;`, Tsx: true},
	}
}

// applyOptionsValid sets `Options` on every case — mirrors upstream's
// `ruleOptionsMapperFactory(options)`.
func applyOptionsValid(cases []rule_tester.ValidTestCase, opts map[string]interface{}) []rule_tester.ValidTestCase {
	out := make([]rule_tester.ValidTestCase, len(cases))
	for i, c := range cases {
		c.Options = opts
		out[i] = c
	}
	return out
}

func applyOptionsInvalid(cases []rule_tester.InvalidTestCase, opts map[string]interface{}) []rule_tester.InvalidTestCase {
	out := make([]rule_tester.InvalidTestCase, len(cases))
	for i, c := range cases {
		c.Options = opts
		out[i] = c
	}
	return out
}

// TestNoNoninteractiveElementToInteractiveRoleUpstreamRecommended mirrors
// upstream's `no-noninteractive-element-to-interactive-role:recommended`
// suite — alwaysValid + extras under recommendedConfig allow-list,
// neverValid as-is.
func TestNoNoninteractiveElementToInteractiveRoleUpstreamRecommended(t *testing.T) {
	valid := append([]rule_tester.ValidTestCase{}, alwaysValid()...)
	valid = append(valid, recommendedExtraValid()...)
	valid = applyOptionsValid(valid, recommendedConfig)

	invalid := applyOptionsInvalid(neverValid(), recommendedConfig)

	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t,
		&NoNoninteractiveElementToInteractiveRoleRule, valid, invalid)
}

// TestNoNoninteractiveElementToInteractiveRoleUpstreamStrict mirrors
// upstream's `no-noninteractive-element-to-interactive-role:strict`
// suite — no options, alwaysValid as-is, neverValid + delta cases that
// recommendedConfig would otherwise exempt.
func TestNoNoninteractiveElementToInteractiveRoleUpstreamStrict(t *testing.T) {
	valid := alwaysValid()

	invalid := append([]rule_tester.InvalidTestCase{}, neverValid()...)
	invalid = append(invalid, recommendedStrictDelta()...)

	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t,
		&NoNoninteractiveElementToInteractiveRoleRule, valid, invalid)
}
