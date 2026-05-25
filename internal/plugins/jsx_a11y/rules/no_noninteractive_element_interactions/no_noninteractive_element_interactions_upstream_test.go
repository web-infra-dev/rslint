package no_noninteractive_element_interactions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError mirrors upstream's `expectedError` constant — the single
// shape every invalid case in the rule produces. Centralized so a future
// error-text tweak touches one place.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "noNoninteractiveElementInteractions",
	Message:   errorMessage,
	Line:      1,
	Column:    1,
}

// recommendedConfig mirrors `configs.recommended.rules['jsx-a11y/no-noninteractive-element-interactions'][1]`.
// Source: eslint-plugin-jsx-a11y/src/index.js.
//
// The `alert` and `dialog` entries are dead in upstream — `hasOwn(config,
// type)` consults the resolved ELEMENT name (via getElementType), not the
// role; `<alert>` is not a real DOM element, and `<dialog>` is a real
// element but its role-driven exemption isn't what upstream's allow-list is
// designed for. They're preserved here for byte-for-byte parity.
var recommendedConfig = map[string]interface{}{
	"handlers": []interface{}{
		"onClick", "onError", "onLoad",
		"onMouseDown", "onMouseUp",
		"onKeyPress", "onKeyDown", "onKeyUp",
	},
	"alert":  []interface{}{"onKeyUp", "onKeyDown", "onKeyPress"},
	"body":   []interface{}{"onError", "onLoad"},
	"dialog": []interface{}{"onKeyUp", "onKeyDown", "onKeyPress"},
	"iframe": []interface{}{"onError", "onLoad"},
	"img":    []interface{}{"onError", "onLoad"},
}

// strictConfig mirrors `configs.strict.rules['jsx-a11y/no-noninteractive-element-interactions'][1]`.
// No `handlers` override — falls back to defaultInteractiveProps (focus +
// image + keyboard + mouse). Source: eslint-plugin-jsx-a11y/src/index.js.
var strictConfig = map[string]interface{}{
	"body":   []interface{}{"onError", "onLoad"},
	"iframe": []interface{}{"onError", "onLoad"},
	"img":    []interface{}{"onError", "onLoad"},
}

// imageComponentsSettings mirrors upstream's
//
//	settings: { 'jsx-a11y': { components: { Image: 'img' } } }
//
// so `<Image>` resolves to `img` and the rule treats it as a non-interactive
// DOM element (`img` is in `nonInteractiveElementAXSchemas`).
var imageComponentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Image": "img",
		},
	},
}

// buttonComponentsSettings mirrors upstream's
//
//	settings: { 'jsx-a11y': { components: { Button: 'button' } } }
//
// — `<Button>` resolves to `button` and the rule treats it as an interactive
// DOM element, so the `<Button onClick={...} />` valid case lands on the
// inherent-interactive bail-out.
var buttonComponentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Button": "button",
		},
	},
}

// alwaysValid mirrors upstream's `alwaysValid` array verbatim — valid under
// EVERY option combination. Order matches the upstream file so a future
// audit can grep across both side-by-side.
func alwaysValid() []rule_tester.ValidTestCase {
	return []rule_tester.ValidTestCase{
		// ---- Custom components — element type isn't in aria-query's
		//      `dom` map → upstream short-circuits before checking
		//      interactivity. ----
		{Code: `<TestComponent onClick={doFoo} />`, Tsx: true},
		{Code: `<Button onClick={doFoo} />`, Tsx: true},
		{Code: `<Image onClick={() => void 0} />;`, Tsx: true},
		{Code: `<Button onClick={() => void 0} />;`, Tsx: true, Settings: buttonComponentsSettings},

		// ---- All flavors of input — `<input>` resolves to an interactive
		//      element via aria-query's elementAXObjects schema (the bare
		//      `{Name: "input"}` entry matches every type). ----
		{Code: `<input onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="button" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="checkbox" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="color" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="date" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="datetime" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="datetime-local" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="email" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="file" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="image" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="month" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="number" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="password" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="radio" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="range" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="reset" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="search" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="submit" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="tel" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="text" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="time" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="url" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="week" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="hidden" onClick={() => void 0} />`, Tsx: true},

		// ---- Interactive elements (anchor / button / option / select /
		//      textarea / area / body / menuitem / tr). ----
		{Code: `<a onClick={() => void 0} />`, Tsx: true},
		{Code: `<a onClick={() => {}} />;`, Tsx: true},
		{Code: `<a tabIndex="0" onClick={() => void 0} />`, Tsx: true},
		{Code: `<a onClick={() => void 0} href="http://x.y.z" />`, Tsx: true},
		{Code: `<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />`, Tsx: true},
		{Code: `<area onClick={() => {}} />;`, Tsx: true},
		{Code: `<body onClick={() => {}} />;`, Tsx: true},
		{Code: `<button onClick={() => void 0} className="foo" />`, Tsx: true},
		{Code: `<menuitem onClick={() => {}} />;`, Tsx: true},
		{Code: `<option onClick={() => void 0} className="foo" />`, Tsx: true},
		{Code: `<select onClick={() => void 0} className="foo" />`, Tsx: true},
		{Code: `<textarea onClick={() => void 0} className="foo" />`, Tsx: true},
		{Code: `<tr onClick={() => {}} />;`, Tsx: true},

		// ---- HTML elements with neither an interactive nor non-interactive
		//      valence (static / "no opinion"). aria-query has no entry for
		//      these in either schema, so the rule lands on the
		//      `(!isNonInteractiveElement && !isNonInteractiveRole)` arm and
		//      bails. ----
		{Code: `<acronym onClick={() => {}} />;`, Tsx: true},
		{Code: `<applet onClick={() => {}} />;`, Tsx: true},
		{Code: `<audio onClick={() => {}} />;`, Tsx: true},
		{Code: `<b onClick={() => {}} />;`, Tsx: true},
		{Code: `<base onClick={() => {}} />;`, Tsx: true},
		{Code: `<bdi onClick={() => {}} />;`, Tsx: true},
		{Code: `<bdo onClick={() => {}} />;`, Tsx: true},
		{Code: `<big onClick={() => {}} />;`, Tsx: true},
		{Code: `<blink onClick={() => {}} />;`, Tsx: true},
		{Code: `<body onLoad={() => {}} />;`, Tsx: true},
		{Code: `<canvas onClick={() => {}} />;`, Tsx: true},
		{Code: `<center onClick={() => {}} />;`, Tsx: true},
		{Code: `<cite onClick={() => {}} />;`, Tsx: true},
		{Code: `<col onClick={() => {}} />;`, Tsx: true},
		{Code: `<colgroup onClick={() => {}} />;`, Tsx: true},
		{Code: `<content onClick={() => {}} />;`, Tsx: true},
		{Code: `<data onClick={() => {}} />;`, Tsx: true},
		{Code: `<datalist onClick={() => {}} />;`, Tsx: true},
		{Code: `<div />;`, Tsx: true},
		{Code: `<div className="foo" />;`, Tsx: true},
		{Code: `<div className="foo" {...props} />;`, Tsx: true},
		{Code: `<div onClick={() => void 0} aria-hidden />;`, Tsx: true},
		{Code: `<div onClick={() => void 0} aria-hidden={true} />;`, Tsx: true},
		{Code: `<div onClick={() => void 0} />;`, Tsx: true},
		{Code: `<div onClick={() => void 0} role={undefined} />;`, Tsx: true},
		{Code: `<div onClick={() => void 0} {...props} />;`, Tsx: true},
		{Code: `<div onClick={null} />;`, Tsx: true},
		{Code: `<div onKeyUp={() => void 0} aria-hidden={false} />;`, Tsx: true},
		{Code: `<embed onClick={() => {}} />;`, Tsx: true},
		{Code: `<font onClick={() => {}} />;`, Tsx: true},
		{Code: `<font onSubmit={() => {}} />;`, Tsx: true},
		{Code: `<form onSubmit={() => {}} />;`, Tsx: true},
		{Code: `
				<form onSubmit={this.handleSubmit.bind(this)} method="POST">
					<button type="submit">
							Save
					</button>
				</form>
			`, Tsx: true},
		{Code: `<frame onClick={() => {}} />;`, Tsx: true},
		{Code: `<frameset onClick={() => {}} />;`, Tsx: true},
		{Code: `<head onClick={() => {}} />;`, Tsx: true},
		{Code: `<header onClick={() => {}} />;`, Tsx: true},
		{Code: `<hgroup onClick={() => {}} />;`, Tsx: true},
		{Code: `<i onClick={() => {}} />;`, Tsx: true},
		{Code: `<iframe onLoad={() => {}} />;`, Tsx: true},
		{Code: `
				<iframe
					name="embeddedExternalPayment"
					ref="embeddedExternalPayment"
					style={iframeStyle}
					onLoad={this.handleLoadIframe}
				/>
			`, Tsx: true},
		{Code: `<img {...props} onError={() => {}} />;`, Tsx: true},
		{Code: `<img onLoad={() => {}} />;`, Tsx: true},
		{Code: `<img src={currentPhoto.imageUrl} onLoad={this.handleImageLoad} alt="for review" />`, Tsx: true},
		{Code: `
					<img
					ref={this.ref}
					className="c-responsive-image-placeholder__image"
					src={src}
					alt={alt}
					data-test-id="test-id"
					onLoad={this.fetchCompleteImage}
				/>
			`, Tsx: true},
		{Code: `<kbd onClick={() => {}} />;`, Tsx: true},
		{Code: `<keygen onClick={() => {}} />;`, Tsx: true},
		{Code: `<link onClick={() => {}} />;`, Tsx: true},
		{Code: `<main onClick={null} />;`, Tsx: true},
		{Code: `<map onClick={() => {}} />;`, Tsx: true},
		{Code: `<meta onClick={() => {}} />;`, Tsx: true},
		{Code: `<noembed onClick={() => {}} />;`, Tsx: true},
		{Code: `<noscript onClick={() => {}} />;`, Tsx: true},
		{Code: `<object onClick={() => {}} />;`, Tsx: true},
		{Code: `<param onClick={() => {}} />;`, Tsx: true},
		{Code: `<picture onClick={() => {}} />;`, Tsx: true},
		{Code: `<q onClick={() => {}} />;`, Tsx: true},
		{Code: `<rp onClick={() => {}} />;`, Tsx: true},
		{Code: `<rt onClick={() => {}} />;`, Tsx: true},
		{Code: `<rtc onClick={() => {}} />;`, Tsx: true},
		{Code: `<s onClick={() => {}} />;`, Tsx: true},
		{Code: `<samp onClick={() => {}} />;`, Tsx: true},
		{Code: `<script onClick={() => {}} />;`, Tsx: true},
		{Code: `<section onClick={() => {}} />;`, Tsx: true},
		{Code: `<small onClick={() => {}} />;`, Tsx: true},
		{Code: `<source onClick={() => {}} />;`, Tsx: true},
		{Code: `<spacer onClick={() => {}} />;`, Tsx: true},
		{Code: `<span onClick={() => {}} />;`, Tsx: true},
		{Code: `<strike onClick={() => {}} />;`, Tsx: true},
		{Code: `<style onClick={() => {}} />;`, Tsx: true},
		{Code: `<summary onClick={() => {}} />;`, Tsx: true},
		{Code: `<th onClick={() => {}} />;`, Tsx: true},
		{Code: `<title onClick={() => {}} />;`, Tsx: true},
		{Code: `<track onClick={() => {}} />;`, Tsx: true},
		{Code: `<td onClick={() => {}} />;`, Tsx: true},
		{Code: `<tt onClick={() => {}} />;`, Tsx: true},
		{Code: `<u onClick={() => {}} />;`, Tsx: true},
		{Code: `<var onClick={() => {}} />;`, Tsx: true},
		{Code: `<video onClick={() => {}} />;`, Tsx: true},
		{Code: `<wbr onClick={() => {}} />;`, Tsx: true},
		{Code: `<xmp onClick={() => {}} />;`, Tsx: true},

		// ---- HTML elements attributed with an interactive role. ----
		{Code: `<div role="button" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="checkbox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="columnheader" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="combobox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="grid" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="gridcell" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="link" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="listbox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="menu" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="menubar" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="menuitem" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="menuitemcheckbox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="menuitemradio" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="option" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="progressbar" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="radio" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="radiogroup" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="row" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="rowheader" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="scrollbar" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="searchbox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="slider" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="spinbutton" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="switch" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="tab" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="textbox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="treeitem" onClick={() => {}} />;`, Tsx: true},

		// ---- presentation / none roles → intentional static semantics. ----
		{Code: `<div role="presentation" onClick={() => {}} />;`, Tsx: true},

		// ---- Abstract roles (the rule has no opinion). ----
		{Code: `<div role="command" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="composite" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="input" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="landmark" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="range" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="roletype" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="sectionhead" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="select" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="structure" onClick={() => {}} />;`, Tsx: true},
		// `tablist` / `toolbar` / `tree` / `treegrid` are concrete INTERACTIVE
		// roles (interactiveRolesSet), so they land on the inherent-interactive
		// arm rather than the abstract-role arm — observable result identical.
		{Code: `<div role="tablist" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="toolbar" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="tree" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="treegrid" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="widget" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="window" onClick={() => {}} />;`, Tsx: true},

		// ---- Non-triggering handlers on a non-interactive role —
		//      handler not in the recommended/strict handlers list, so the
		//      rule short-circuits at hasInteractiveProps. ----
		{Code: `<div role="article" onCopy={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onCut={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onPaste={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onCompositionEnd={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onCompositionStart={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onCompositionUpdate={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onChange={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onInput={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onSubmit={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onSelect={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onTouchCancel={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onTouchEnd={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onTouchMove={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onTouchStart={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onScroll={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onWheel={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onAbort={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onCanPlay={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onCanPlayThrough={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onDurationChange={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onEmptied={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onEncrypted={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onEnded={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onLoadedData={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onLoadedMetadata={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onLoadStart={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onPause={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onPlay={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onPlaying={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onProgress={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onRateChange={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onSeeked={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onSeeking={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onStalled={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onSuspend={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onTimeUpdate={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onVolumeChange={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onWaiting={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onAnimationStart={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onAnimationEnd={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onAnimationIteration={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onTransitionEnd={() => {}} />;`, Tsx: true},
	}
}

// neverValid mirrors upstream's `neverValid` array verbatim — invalid under
// EVERY option combination. Order matches the upstream file.
func neverValid() []rule_tester.InvalidTestCase {
	cases := []struct {
		Code     string
		Settings map[string]interface{}
	}{
		// ---- HTML elements with an inherent, non-interactive role. ----
		{Code: `<main onClick={() => void 0} />;`},
		{Code: `<address onClick={() => {}} />;`},
		{Code: `<article onClick={() => {}} />;`},
		{Code: `<aside onClick={() => {}} />;`},
		{Code: `<blockquote onClick={() => {}} />;`},
		{Code: `<br onClick={() => {}} />;`},
		{Code: `<caption onClick={() => {}} />;`},
		{Code: `<code onClick={() => {}} />;`},
		{Code: `<dd onClick={() => {}} />;`},
		{Code: `<del onClick={() => {}} />;`},
		{Code: `<details onClick={() => {}} />;`},
		{Code: `<dfn onClick={() => {}} />;`},
		{Code: `<dl onClick={() => {}} />;`},
		{Code: `<dir onClick={() => {}} />;`},
		{Code: `<dt onClick={() => {}} />;`},
		{Code: `<em onClick={() => {}} />;`},
		{Code: `<fieldset onClick={() => {}} />;`},
		{Code: `<figcaption onClick={() => {}} />;`},
		{Code: `<figure onClick={() => {}} />;`},
		{Code: `<footer onClick={() => {}} />;`},
		{Code: `<form onClick={() => {}} />;`},
		{Code: `<h1 onClick={() => {}} />;`},
		{Code: `<h2 onClick={() => {}} />;`},
		{Code: `<h3 onClick={() => {}} />;`},
		{Code: `<h4 onClick={() => {}} />;`},
		{Code: `<h5 onClick={() => {}} />;`},
		{Code: `<h6 onClick={() => {}} />;`},
		{Code: `<hr onClick={() => {}} />;`},
		{Code: `<html onClick={() => {}} />;`},
		{Code: `<iframe onClick={() => {}} />;`},
		{Code: `<img onClick={() => {}} />;`},
		{Code: `<ins onClick={() => {}} />;`},
		{Code: `<label onClick={() => {}} />;`},
		{Code: `<legend onClick={() => {}} />;`},
		{Code: `<li onClick={() => {}} />;`},
		{Code: `<mark onClick={() => {}} />;`},
		{Code: `<marquee onClick={() => {}} />;`},
		{Code: `<menu onClick={() => {}} />;`},
		{Code: `<meter onClick={() => {}} />;`},
		{Code: `<nav onClick={() => {}} />;`},
		{Code: `<ol onClick={() => {}} />;`},
		{Code: `<optgroup onClick={() => {}} />;`},
		{Code: `<output onClick={() => {}} />;`},
		{Code: `<p onClick={() => {}} />;`},
		{Code: `<pre onClick={() => {}} />;`},
		{Code: `<progress onClick={() => {}} />;`},
		{Code: `<ruby onClick={() => {}} />;`},
		{Code: `<section onClick={() => {}} aria-label="Aardvark" />;`},
		{Code: `<section onClick={() => {}} aria-labelledby="js_1" />;`},
		{Code: `<strong onClick={() => {}} />;`},
		{Code: `<sub onClick={() => {}} />;`},
		{Code: `<sup onClick={() => {}} />;`},
		{Code: `<table onClick={() => {}} />;`},
		{Code: `<tbody onClick={() => {}} />;`},
		{Code: `<tfoot onClick={() => {}} />;`},
		{Code: `<thead onClick={() => {}} />;`},
		{Code: `<time onClick={() => {}} />;`},
		{Code: `<ul onClick={() => {}} />;`},
		// `contentEditable="false"` — the raw value isn't `"true"`, so the
		// IsContentEditable bail-out does NOT fire and the rule reports.
		{Code: `<ul contentEditable="false" onClick={() => {}} />;`},
		// `contentEditable` boolean form — no `.value.raw`, so the
		// IsContentEditable check is also false and the rule reports.
		{Code: `<article contentEditable onClick={() => {}} />;`},
		// `contentEditable` boolean form on a `role="article"` div — same
		// reasoning as above.
		{Code: `<div contentEditable role="article" onKeyDown={() => {}} />;`},

		// ---- HTML elements attributed with a non-interactive role. ----
		{Code: `<div role="alert" onClick={() => {}} />;`},
		{Code: `<div role="alertdialog" onClick={() => {}} />;`},
		{Code: `<div role="application" onClick={() => {}} />;`},
		{Code: `<div role="banner" onClick={() => {}} />;`},
		{Code: `<div role="cell" onClick={() => {}} />;`},
		{Code: `<div role="complementary" onClick={() => {}} />;`},
		{Code: `<div role="contentinfo" onClick={() => {}} />;`},
		{Code: `<div role="definition" onClick={() => {}} />;`},
		{Code: `<div role="dialog" onClick={() => {}} />;`},
		{Code: `<div role="directory" onClick={() => {}} />;`},
		{Code: `<div role="document" onClick={() => {}} />;`},
		{Code: `<div role="feed" onClick={() => {}} />;`},
		{Code: `<div role="figure" onClick={() => {}} />;`},
		{Code: `<div role="form" onClick={() => {}} />;`},
		{Code: `<div role="group" onClick={() => {}} />;`},
		{Code: `<div role="heading" onClick={() => {}} />;`},
		{Code: `<div role="img" onClick={() => {}} />;`},
		{Code: `<div role="list" onClick={() => {}} />;`},
		{Code: `<div role="listitem" onClick={() => {}} />;`},
		{Code: `<div role="log" onClick={() => {}} />;`},
		{Code: `<div role="main" onClick={() => {}} />;`},
		{Code: `<div role="marquee" onClick={() => {}} />;`},
		{Code: `<div role="math" onClick={() => {}} />;`},
		{Code: `<div role="navigation" onClick={() => {}} />;`},
		{Code: `<div role="note" onClick={() => {}} />;`},
		{Code: `<div role="region" onClick={() => {}} />;`},
		{Code: `<div role="rowgroup" onClick={() => {}} />;`},
		{Code: `<div role="search" onClick={() => {}} />;`},
		{Code: `<div role="separator" onClick={() => {}} />;`},
		{Code: `<div role="status" onClick={() => {}} />;`},
		{Code: `<div role="table" onClick={() => {}} />;`},
		{Code: `<div role="tabpanel" onClick={() => {}} />;`},
		{Code: `<div role="term" onClick={() => {}} />;`},
		{Code: `<div role="timer" onClick={() => {}} />;`},
		{Code: `<div role="tooltip" onClick={() => {}} />;`},

		// ---- Triggering handlers on a non-interactive role. ----
		{Code: `<div role="article" onKeyDown={() => {}} />;`},
		{Code: `<div role="article" onKeyPress={() => {}} />;`},
		{Code: `<div role="article" onKeyUp={() => {}} />;`},
		{Code: `<div role="article" onClick={() => {}} />;`},
		{Code: `<div role="article" onLoad={() => {}} />;`},
		{Code: `<div role="article" onError={() => {}} />;`},
		{Code: `<div role="article" onMouseDown={() => {}} />;`},
		{Code: `<div role="article" onMouseUp={() => {}} />;`},

		// ---- Custom component resolved to a non-interactive DOM via
		//      `settings['jsx-a11y'].components`. ----
		{Code: `<Image onClick={() => void 0} />;`, Settings: imageComponentsSettings},
	}

	out := make([]rule_tester.InvalidTestCase, len(cases))
	for i, c := range cases {
		out[i] = rule_tester.InvalidTestCase{
			Code:     c.Code,
			Tsx:      true,
			Settings: c.Settings,
			Errors:   []rule_tester.InvalidTestCaseError{expectedError},
		}
	}
	return out
}

// recommendedExtraValid mirrors the extra `valid` cases the upstream
// recommended suite appends after `alwaysValid` — every handler not in
// recommended's `handlers` list short-circuits at `hasInteractiveProps`.
//
// Under recommended `handlers = ['onClick', 'onError', 'onLoad',
// 'onMouseDown', 'onMouseUp', 'onKeyPress', 'onKeyDown', 'onKeyUp']`, all of
// these handler names are OUTSIDE the list, so each case is valid.
func recommendedExtraValid() []rule_tester.ValidTestCase {
	handlers := []string{
		"onCopy", "onCut", "onPaste",
		"onCompositionEnd", "onCompositionStart", "onCompositionUpdate",
		"onFocus", "onBlur",
		"onChange", "onInput", "onSubmit",
		"onContextMenu", "onDblClick", "onDoubleClick",
		"onDrag", "onDragEnd", "onDragEnter", "onDragExit",
		"onDragLeave", "onDragOver", "onDragStart", "onDrop",
		"onMouseEnter", "onMouseLeave", "onMouseMove", "onMouseOut", "onMouseOver",
		"onSelect",
		"onTouchCancel", "onTouchEnd", "onTouchMove", "onTouchStart",
		"onScroll", "onWheel",
		"onAbort", "onCanPlay", "onCanPlayThrough", "onDurationChange",
		"onEmptied", "onEncrypted", "onEnded", "onLoadedData",
		"onLoadedMetadata", "onLoadStart", "onPause", "onPlay", "onPlaying",
		"onProgress", "onRateChange", "onSeeked", "onSeeking", "onStalled",
		"onSuspend", "onTimeUpdate", "onVolumeChange", "onWaiting",
		"onAnimationStart", "onAnimationEnd", "onAnimationIteration",
		"onTransitionEnd",
	}
	out := make([]rule_tester.ValidTestCase, len(handlers))
	for i, h := range handlers {
		out[i] = rule_tester.ValidTestCase{
			Code: "<div role=\"article\" " + h + "={() => {}} />;",
			Tsx:  true,
		}
	}
	return out
}

// strictExtraInvalid mirrors the extra `invalid` cases the upstream strict
// suite appends after `neverValid`. Strict has no `handlers` override, so
// the rule falls back to defaultInteractiveProps (focus + image + keyboard
// + mouse). These handlers are in that default but NOT in recommended's
// `handlers` list, so they fail under strict only.
func strictExtraInvalid() []rule_tester.InvalidTestCase {
	handlers := []string{
		"onFocus", "onBlur",
		"onContextMenu", "onDblClick", "onDoubleClick",
		"onDrag", "onDragEnd", "onDragEnter", "onDragExit",
		"onDragLeave", "onDragOver", "onDragStart", "onDrop",
		"onMouseEnter", "onMouseLeave", "onMouseMove",
		"onMouseOut", "onMouseOver",
	}
	out := make([]rule_tester.InvalidTestCase, len(handlers))
	for i, h := range handlers {
		out[i] = rule_tester.InvalidTestCase{
			Code:   "<div role=\"article\" " + h + "={() => {}} />;",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		}
	}
	return out
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

// TestNoNoninteractiveElementInteractionsUpstreamRecommended mirrors
// upstream's `no-noninteractive-element-interactions:recommended` suite —
// the same valid/invalid arrays the upstream `ruleTester.run` block feeds.
func TestNoNoninteractiveElementInteractionsUpstreamRecommended(t *testing.T) {
	valid := append([]rule_tester.ValidTestCase{}, alwaysValid()...)
	valid = append(valid, recommendedExtraValid()...)
	valid = applyOptionsValid(valid, recommendedConfig)

	invalid := applyOptionsInvalid(neverValid(), recommendedConfig)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&NoNoninteractiveElementInteractionsRule, valid, invalid)
}

// TestNoNoninteractiveElementInteractionsUpstreamStrict mirrors upstream's
// `no-noninteractive-element-interactions:strict` suite — strict has no
// `handlers` override, so it adds focus / contextMenu / etc. invalid cases.
func TestNoNoninteractiveElementInteractionsUpstreamStrict(t *testing.T) {
	valid := applyOptionsValid(alwaysValid(), strictConfig)

	invalid := append([]rule_tester.InvalidTestCase{}, neverValid()...)
	invalid = append(invalid, strictExtraInvalid()...)
	invalid = applyOptionsInvalid(invalid, strictConfig)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&NoNoninteractiveElementInteractionsRule, valid, invalid)
}
