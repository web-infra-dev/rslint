package main

// This file ports the static known-global classifications used by Oxc's
// side-effect analysis: oxc_ecmascript/src/side_effects/known_globals.rs.

type knownGlobalSymbol struct {
	Name     string
	LoadPure bool
	CallPure bool
}

type knownGlobalMemberGroup struct {
	Global  string
	Members []string
}

var pureGlobalFunctions = []string{
	"decodeURI", "decodeURIComponent", "encodeURI", "encodeURIComponent",
	"escape", "isFinite", "isNaN", "parseFloat", "parseInt",
}

// Constructors that are side-effect-free when called as functions. Constructors
// whose behavior depends on argument values are intentionally excluded.
var pureCallableConstructors = []string{"Date", "Boolean", "Object", "String"}

var typedArrayConstructors = []string{
	"Int8Array", "Uint8Array", "Uint8ClampedArray",
	"Int16Array", "Uint16Array", "Int32Array", "Uint32Array",
	"Float32Array", "Float64Array", "BigInt64Array", "BigUint64Array",
}

// NaN, Infinity, and undefined are handled separately by Oxc and are included
// here because the semantic transport has no separate literal representation
// for those identifier symbols.
var knownGlobalIdentifiers = []string{
	"NaN", "Infinity", "undefined",
	// Core JS globals.
	"Array", "Boolean", "Function", "Math", "Number", "Object", "RegExp", "String",
	// Globals present in both browsers and Node.
	"AbortController", "AbortSignal", "AggregateError", "ArrayBuffer", "BigInt",
	"DataView", "Date", "Error", "EvalError", "Event", "EventTarget",
	"Float32Array", "Float64Array", "Int16Array", "Int32Array", "Int8Array", "Intl",
	"JSON", "Map", "MessageChannel", "MessageEvent", "MessagePort", "Promise",
	"Proxy", "RangeError", "ReferenceError", "Reflect", "Set", "Symbol",
	"SyntaxError", "TextDecoder", "TextEncoder", "TypeError", "URIError", "URL",
	"URLSearchParams", "Uint16Array", "Uint32Array", "Uint8Array",
	"Uint8ClampedArray", "WeakMap", "WeakSet", "WebAssembly",
	"clearInterval", "clearTimeout", "console", "decodeURI", "decodeURIComponent",
	"encodeURI", "encodeURIComponent", "escape", "globalThis", "isFinite", "isNaN",
	"parseFloat", "parseInt", "queueMicrotask", "setInterval", "setTimeout",
	"unescape",
	// CSSOM APIs.
	"CSSAnimation", "CSSFontFaceRule", "CSSImportRule", "CSSKeyframeRule",
	"CSSKeyframesRule", "CSSMediaRule", "CSSNamespaceRule", "CSSPageRule", "CSSRule",
	"CSSRuleList", "CSSStyleDeclaration", "CSSStyleRule", "CSSStyleSheet",
	"CSSSupportsRule", "CSSTransition",
	// SVG DOM.
	"SVGAElement", "SVGAngle", "SVGAnimateElement", "SVGAnimateMotionElement",
	"SVGAnimateTransformElement", "SVGAnimatedAngle", "SVGAnimatedBoolean",
	"SVGAnimatedEnumeration", "SVGAnimatedInteger", "SVGAnimatedLength",
	"SVGAnimatedLengthList", "SVGAnimatedNumber", "SVGAnimatedNumberList",
	"SVGAnimatedPreserveAspectRatio", "SVGAnimatedRect", "SVGAnimatedString",
	"SVGAnimatedTransformList", "SVGAnimationElement", "SVGCircleElement",
	"SVGClipPathElement", "SVGComponentTransferFunctionElement", "SVGDefsElement",
	"SVGDescElement", "SVGElement", "SVGEllipseElement", "SVGFEBlendElement",
	"SVGFEColorMatrixElement", "SVGFEComponentTransferElement",
	"SVGFECompositeElement", "SVGFEConvolveMatrixElement",
	"SVGFEDiffuseLightingElement", "SVGFEDisplacementMapElement",
	"SVGFEDistantLightElement", "SVGFEDropShadowElement", "SVGFEFloodElement",
	"SVGFEFuncAElement", "SVGFEFuncBElement", "SVGFEFuncGElement",
	"SVGFEFuncRElement", "SVGFEGaussianBlurElement", "SVGFEImageElement",
	"SVGFEMergeElement", "SVGFEMergeNodeElement", "SVGFEMorphologyElement",
	"SVGFEOffsetElement", "SVGFEPointLightElement", "SVGFESpecularLightingElement",
	"SVGFESpotLightElement", "SVGFETileElement", "SVGFETurbulenceElement",
	"SVGFilterElement", "SVGForeignObjectElement", "SVGGElement",
	"SVGGeometryElement", "SVGGradientElement", "SVGGraphicsElement",
	"SVGImageElement", "SVGLength", "SVGLengthList", "SVGLineElement",
	"SVGLinearGradientElement", "SVGMPathElement", "SVGMarkerElement",
	"SVGMaskElement", "SVGMatrix", "SVGMetadataElement", "SVGNumber",
	"SVGNumberList", "SVGPathElement", "SVGPatternElement", "SVGPoint",
	"SVGPointList", "SVGPolygonElement", "SVGPolylineElement",
	"SVGPreserveAspectRatio", "SVGRadialGradientElement", "SVGRect",
	"SVGRectElement", "SVGSVGElement", "SVGScriptElement", "SVGSetElement",
	"SVGStopElement", "SVGStringList", "SVGStyleElement", "SVGSwitchElement",
	"SVGSymbolElement", "SVGTSpanElement", "SVGTextContentElement",
	"SVGTextElement", "SVGTextPathElement", "SVGTextPositioningElement",
	"SVGTitleElement", "SVGTransform", "SVGTransformList", "SVGUnitTypes",
	"SVGUseElement", "SVGViewElement",
	// Other browser APIs.
	"AnalyserNode", "Animation", "AnimationEffect", "AnimationEvent",
	"AnimationPlaybackEvent", "AnimationTimeline", "Attr", "Audio", "AudioBuffer",
	"AudioBufferSourceNode", "AudioDestinationNode", "AudioListener", "AudioNode",
	"AudioParam", "AudioProcessingEvent", "AudioScheduledSourceNode", "BarProp",
	"BeforeUnloadEvent", "BiquadFilterNode", "Blob", "BlobEvent",
	"ByteLengthQueuingStrategy", "CDATASection", "CSS", "CanvasGradient",
	"CanvasPattern", "CanvasRenderingContext2D", "ChannelMergerNode",
	"ChannelSplitterNode", "CharacterData", "ClipboardEvent", "CloseEvent",
	"Comment", "CompositionEvent", "ConvolverNode", "CountQueuingStrategy",
	"Crypto", "CustomElementRegistry", "CustomEvent", "DOMException",
	"DOMImplementation", "DOMMatrix", "DOMMatrixReadOnly", "DOMParser", "DOMPoint",
	"DOMPointReadOnly", "DOMQuad", "DOMRect", "DOMRectList", "DOMRectReadOnly",
	"DOMStringList", "DOMStringMap", "DOMTokenList", "DataTransfer",
	"DataTransferItem", "DataTransferItemList", "DelayNode", "Document",
	"DocumentFragment", "DocumentTimeline", "DocumentType", "DragEvent",
	"DynamicsCompressorNode", "Element", "ErrorEvent", "EventSource", "File",
	"FileList", "FileReader", "FocusEvent", "FontFace", "FormData", "GainNode",
	"Gamepad", "GamepadButton", "GamepadEvent", "Geolocation",
	"GeolocationPositionError", "HTMLAllCollection", "HTMLAnchorElement",
	"HTMLAreaElement", "HTMLAudioElement", "HTMLBRElement", "HTMLBaseElement",
	"HTMLBodyElement", "HTMLButtonElement", "HTMLCanvasElement", "HTMLCollection",
	"HTMLDListElement", "HTMLDataElement", "HTMLDataListElement",
	"HTMLDetailsElement", "HTMLDirectoryElement", "HTMLDivElement", "HTMLDocument",
	"HTMLElement", "HTMLEmbedElement", "HTMLFieldSetElement", "HTMLFontElement",
	"HTMLFormControlsCollection", "HTMLFormElement", "HTMLFrameElement",
	"HTMLFrameSetElement", "HTMLHRElement", "HTMLHeadElement",
	"HTMLHeadingElement", "HTMLHtmlElement", "HTMLIFrameElement",
	"HTMLImageElement", "HTMLInputElement", "HTMLLIElement", "HTMLLabelElement",
	"HTMLLegendElement", "HTMLLinkElement", "HTMLMapElement", "HTMLMarqueeElement",
	"HTMLMediaElement", "HTMLMenuElement", "HTMLMetaElement", "HTMLMeterElement",
	"HTMLModElement", "HTMLOListElement", "HTMLObjectElement",
	"HTMLOptGroupElement", "HTMLOptionElement", "HTMLOptionsCollection",
	"HTMLOutputElement", "HTMLParagraphElement", "HTMLParamElement",
	"HTMLPictureElement", "HTMLPreElement", "HTMLProgressElement",
	"HTMLQuoteElement", "HTMLScriptElement", "HTMLSelectElement",
	"HTMLSlotElement", "HTMLSourceElement", "HTMLSpanElement", "HTMLStyleElement",
	"HTMLTableCaptionElement", "HTMLTableCellElement", "HTMLTableColElement",
	"HTMLTableElement", "HTMLTableRowElement", "HTMLTableSectionElement",
	"HTMLTemplateElement", "HTMLTextAreaElement", "HTMLTimeElement",
	"HTMLTitleElement", "HTMLTrackElement", "HTMLUListElement",
	"HTMLUnknownElement", "HTMLVideoElement", "HashChangeEvent", "Headers",
	"History", "IDBCursor", "IDBCursorWithValue", "IDBDatabase", "IDBFactory",
	"IDBIndex", "IDBKeyRange", "IDBObjectStore", "IDBOpenDBRequest", "IDBRequest",
	"IDBTransaction", "IDBVersionChangeEvent", "Image", "ImageData", "InputEvent",
	"IntersectionObserver", "IntersectionObserverEntry", "KeyboardEvent",
	"KeyframeEffect", "Location", "MediaCapabilities",
	"MediaElementAudioSourceNode", "MediaEncryptedEvent", "MediaError",
	"MediaList", "MediaQueryList", "MediaQueryListEvent", "MediaRecorder",
	"MediaSource", "MediaStream", "MediaStreamAudioDestinationNode",
	"MediaStreamAudioSourceNode", "MediaStreamTrack", "MediaStreamTrackEvent",
	"MimeType", "MimeTypeArray", "MouseEvent", "MutationEvent",
	"MutationObserver", "MutationRecord", "NamedNodeMap", "Navigator", "Node",
	"NodeFilter", "NodeIterator", "NodeList", "Notification",
	"OfflineAudioCompletionEvent", "Option", "OscillatorNode",
	"PageTransitionEvent", "Path2D", "Performance", "PerformanceEntry",
	"PerformanceMark", "PerformanceMeasure", "PerformanceNavigation",
	"PerformanceObserver", "PerformanceObserverEntryList",
	"PerformanceResourceTiming", "PerformanceTiming", "PeriodicWave", "Plugin",
	"PluginArray", "PointerEvent", "PopStateEvent", "ProcessingInstruction",
	"ProgressEvent", "PromiseRejectionEvent", "RTCCertificate", "RTCDTMFSender",
	"RTCDTMFToneChangeEvent", "RTCDataChannel", "RTCDataChannelEvent",
	"RTCIceCandidate", "RTCPeerConnection", "RTCPeerConnectionIceEvent",
	"RTCRtpReceiver", "RTCRtpSender", "RTCRtpTransceiver",
	"RTCSessionDescription", "RTCStatsReport", "RTCTrackEvent", "RadioNodeList",
	"Range", "ReadableStream", "Request", "ResizeObserver",
	"ResizeObserverEntry", "Response", "Screen", "ScriptProcessorNode",
	"SecurityPolicyViolationEvent", "Selection", "ShadowRoot", "SourceBuffer",
	"SourceBufferList", "SpeechSynthesisEvent", "SpeechSynthesisUtterance",
	"StaticRange", "Storage", "StorageEvent", "StyleSheet", "StyleSheetList",
	"Text", "TextMetrics", "TextTrack", "TextTrackCue", "TextTrackCueList",
	"TextTrackList", "TimeRanges", "TrackEvent", "TransitionEvent", "TreeWalker",
	"UIEvent", "VTTCue", "ValidityState", "VisualViewport", "WaveShaperNode",
	"WebGLActiveInfo", "WebGLBuffer", "WebGLContextEvent", "WebGLFramebuffer",
	"WebGLProgram", "WebGLQuery", "WebGLRenderbuffer", "WebGLRenderingContext",
	"WebGLSampler", "WebGLShader", "WebGLShaderPrecisionFormat", "WebGLSync",
	"WebGLTexture", "WebGLUniformLocation", "WebKitCSSMatrix", "WebSocket",
	"WheelEvent", "Window", "Worker", "XMLDocument", "XMLHttpRequest",
	"XMLHttpRequestEventTarget", "XMLHttpRequestUpload", "XMLSerializer",
	"XPathEvaluator", "XPathExpression", "XPathResult", "XSLTProcessor",
	"alert", "atob", "blur", "btoa", "cancelAnimationFrame", "captureEvents",
	"close", "closed", "confirm", "customElements", "devicePixelRatio",
	"document", "event", "fetch", "find", "focus", "frameElement", "frames",
	"getComputedStyle", "getSelection", "history", "indexedDB", "isSecureContext",
	"length", "location", "locationbar", "matchMedia", "menubar", "moveBy",
	"moveTo", "name", "navigator",
	"onabort", "onafterprint", "onanimationend", "onanimationiteration",
	"onanimationstart", "onbeforeprint", "onbeforeunload", "onblur", "oncanplay",
	"oncanplaythrough", "onchange", "onclick", "oncontextmenu", "oncuechange",
	"ondblclick", "ondrag", "ondragend", "ondragenter", "ondragleave",
	"ondragover", "ondragstart", "ondrop", "ondurationchange", "onemptied",
	"onended", "onerror", "onfocus", "ongotpointercapture", "onhashchange",
	"oninput", "oninvalid", "onkeydown", "onkeypress", "onkeyup",
	"onlanguagechange", "onload", "onloadeddata", "onloadedmetadata",
	"onloadstart", "onlostpointercapture", "onmessage", "onmousedown",
	"onmouseenter", "onmouseleave", "onmousemove", "onmouseout", "onmouseover",
	"onmouseup", "onoffline", "ononline", "onpagehide", "onpageshow", "onpause",
	"onplay", "onplaying", "onpointercancel", "onpointerdown", "onpointerenter",
	"onpointerleave", "onpointermove", "onpointerout", "onpointerover",
	"onpointerup", "onpopstate", "onprogress", "onratechange",
	"onrejectionhandled", "onreset", "onresize", "onscroll", "onseeked",
	"onseeking", "onselect", "onstalled", "onstorage", "onsubmit", "onsuspend",
	"ontimeupdate", "ontoggle", "ontransitioncancel", "ontransitionend",
	"ontransitionrun", "ontransitionstart", "onunhandledrejection", "onunload",
	"onvolumechange", "onwaiting", "onwebkitanimationend",
	"onwebkitanimationiteration", "onwebkitanimationstart",
	"onwebkittransitionend", "onwheel",
	"open", "opener", "origin", "outerHeight", "outerWidth", "parent",
	"performance", "personalbar", "postMessage", "print", "prompt",
	"releaseEvents", "requestAnimationFrame", "resizeBy", "resizeTo", "screen",
	"screenLeft", "screenTop", "screenX", "screenY", "scroll", "scrollBy",
	"scrollTo", "scrollbars", "self", "speechSynthesis", "status", "statusbar",
	"stop", "toolbar", "top", "webkitURL", "window",
}

var pureMathMethods = []string{
	"abs", "acos", "acosh", "asin", "asinh", "atan", "atan2", "atanh",
	"cbrt", "ceil", "clz32", "cos", "cosh", "exp", "expm1", "floor",
	"fround", "hypot", "imul", "log", "log10", "log1p", "log2", "max",
	"min", "pow", "random", "round", "sign", "sin", "sinh", "sqrt",
	"tan", "tanh", "trunc",
}

// Calls to these members are side-effect-free. Argument evaluation remains
// represented separately in RSLIM's IR.
var pureGlobalMethodGroups = []knownGlobalMemberGroup{
	{Global: "Array", Members: []string{"isArray", "of"}},
	{Global: "ArrayBuffer", Members: []string{"isView"}},
	{Global: "Date", Members: []string{"now", "parse", "UTC"}},
	{Global: "Math", Members: pureMathMethods},
	{Global: "Number", Members: []string{"isFinite", "isInteger", "isNaN", "isSafeInteger", "parseFloat", "parseInt"}},
	{Global: "Object", Members: []string{"is"}},
	{Global: "String", Members: []string{"fromCharCode", "fromCodePoint"}},
	{Global: "Symbol", Members: []string{"for"}},
	{Global: "URL", Members: []string{"canParse"}},
}

var knownGlobalPropertyGroups = []knownGlobalMemberGroup{
	{
		Global:  "Math",
		Members: append([]string{"E", "LN10", "LN2", "LOG10E", "LOG2E", "PI", "SQRT1_2", "SQRT2"}, pureMathMethods...),
	},
	{
		Global: "console",
		Members: []string{
			"assert", "clear", "count", "countReset", "debug", "dir", "dirxml",
			"error", "group", "groupCollapsed", "groupEnd", "info", "log",
			"table", "time", "timeEnd", "timeLog", "trace", "warn",
		},
	},
	{
		Global: "Object",
		Members: []string{
			"assign", "create", "defineProperties", "defineProperty", "entries",
			"freeze", "fromEntries", "getOwnPropertyDescriptor",
			"getOwnPropertyDescriptors", "getOwnPropertyNames",
			"getOwnPropertySymbols", "getPrototypeOf", "is", "isExtensible",
			"isFrozen", "isSealed", "keys", "preventExtensions", "prototype",
			"seal", "setPrototypeOf", "values",
		},
	},
	{
		Global: "Reflect",
		Members: []string{
			"apply", "construct", "defineProperty", "deleteProperty", "get",
			"getOwnPropertyDescriptor", "getPrototypeOf", "has", "isExtensible",
			"ownKeys", "preventExtensions", "set", "setPrototypeOf",
		},
	},
	{
		Global: "Symbol",
		Members: []string{
			"asyncDispose", "asyncIterator", "dispose", "hasInstance",
			"isConcatSpreadable", "iterator", "match", "matchAll", "replace",
			"search", "species", "split", "toPrimitive", "toStringTag",
			"unscopables",
		},
	},
	{Global: "JSON", Members: []string{"parse", "stringify"}},
}

var knownObjectPrototypeProperties = []string{
	"__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__",
	"hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable", "toLocaleString",
	"toString", "unwatch", "valueOf", "watch",
}

var knownGlobalSymbols = buildKnownGlobalSymbols()

func buildKnownGlobalSymbols() []knownGlobalSymbol {
	symbols := make([]knownGlobalSymbol, 0, len(knownGlobalIdentifiers))
	indexes := make(map[string]int, len(knownGlobalIdentifiers))
	add := func(name string, loadPure bool, callPure bool) {
		if index, ok := indexes[name]; ok {
			symbols[index].LoadPure = symbols[index].LoadPure || loadPure
			symbols[index].CallPure = symbols[index].CallPure || callPure
			return
		}
		indexes[name] = len(symbols)
		symbols = append(symbols, knownGlobalSymbol{
			Name:     name,
			LoadPure: loadPure,
			CallPure: callPure,
		})
	}

	for _, name := range knownGlobalIdentifiers {
		add(name, true, false)
	}
	for _, name := range pureGlobalFunctions {
		add(name, false, true)
	}
	for _, name := range pureCallableConstructors {
		add(name, false, true)
	}
	for _, group := range knownGlobalPropertyGroups {
		for _, member := range group.Members {
			add(group.Global+"."+member, true, false)
		}
	}
	for _, group := range pureGlobalMethodGroups {
		for _, member := range group.Members {
			add(group.Global+"."+member, false, true)
		}
	}
	for _, global := range typedArrayConstructors {
		add(global+".of", false, true)
	}
	for _, property := range knownObjectPrototypeProperties {
		add("Object.prototype."+property, true, false)
	}

	return symbols
}
