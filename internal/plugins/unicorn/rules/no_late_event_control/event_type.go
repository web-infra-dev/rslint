package no_late_event_control

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var eventParameterName = regexp.MustCompile("^(e|event|evt|[a-z][0-9A-Za-z]*Event)$")

// Event names recognized by Unicorn v79.0.0, including DOM, worker and React types.
var eventTypeNames = utils.NewSetFromItems(
	"AnimationEvent",
	"AnimationPlaybackEvent",
	"AudioProcessingEvent",
	"BaseSyntheticEvent",
	"BeforeUnloadEvent",
	"BlobEvent",
	"ChangeEvent",
	"ClipboardEvent",
	"CloseEvent",
	"CommandEvent",
	"CompositionEvent",
	"ContentVisibilityAutoStateChangeEvent",
	"CookieChangeEvent",
	"CustomEvent",
	"DeviceMotionEvent",
	"DeviceOrientationEvent",
	"DragEvent",
	"ErrorEvent",
	"Event",
	"ExtendableCookieChangeEvent",
	"ExtendableEvent",
	"ExtendableMessageEvent",
	"FetchEvent",
	"FocusEvent",
	"FontFaceSetLoadEvent",
	"FormDataEvent",
	"FormEvent",
	"GPUUncapturedErrorEvent",
	"GamepadEvent",
	"HashChangeEvent",
	"IDBVersionChangeEvent",
	"InputEvent",
	"InvalidEvent",
	"KeyboardEvent",
	"MIDIConnectionEvent",
	"MIDIMessageEvent",
	"MediaEncryptedEvent",
	"MediaKeyMessageEvent",
	"MediaQueryListEvent",
	"MediaStreamTrackEvent",
	"MessageEvent",
	"MouseEvent",
	"NavigateEvent",
	"NavigationCurrentEntryChangeEvent",
	"NotificationEvent",
	"OfflineAudioCompletionEvent",
	"PageRevealEvent",
	"PageSwapEvent",
	"PageTransitionEvent",
	"PaymentMethodChangeEvent",
	"PaymentRequestUpdateEvent",
	"PictureInPictureEvent",
	"PointerEvent",
	"PopStateEvent",
	"ProgressEvent",
	"PromiseRejectionEvent",
	"PushEvent",
	"PushSubscriptionChangeEvent",
	"RTCDTMFToneChangeEvent",
	"RTCDataChannelEvent",
	"RTCErrorEvent",
	"RTCPeerConnectionIceErrorEvent",
	"RTCPeerConnectionIceEvent",
	"RTCTrackEvent",
	"RTCTransformEvent",
	"SecurityPolicyViolationEvent",
	"SpeechRecognitionErrorEvent",
	"SpeechRecognitionEvent",
	"SpeechSynthesisErrorEvent",
	"SpeechSynthesisEvent",
	"StorageEvent",
	"SubmitEvent",
	"SyntheticEvent",
	"TaskPriorityChangeEvent",
	"TextEvent",
	"ToggleEvent",
	"TouchEvent",
	"TrackEvent",
	"TransitionEvent",
	"UIEvent",
	"WebGLContextEvent",
	"WheelEvent",
)

var reactEventTypeNames = utils.NewSetFromItems(
	"AnimationEvent",
	"BaseSyntheticEvent",
	"ChangeEvent",
	"ClipboardEvent",
	"CompositionEvent",
	"DragEvent",
	"FocusEvent",
	"FormEvent",
	"InputEvent",
	"InvalidEvent",
	"KeyboardEvent",
	"MouseEvent",
	"PointerEvent",
	"SubmitEvent",
	"SyntheticEvent",
	"TouchEvent",
	"ToggleEvent",
	"TransitionEvent",
	"UIEvent",
	"WheelEvent",
)

var eventTypes = unicornutil.TypeClassifierOptions{
	TargetTypeNames:     eventTypeNames,
	ClassifyTypeImport:  classifyEventTypeImport,
	HeritageSymbolFlags: ast.SymbolFlagsClass | ast.SymbolFlagsInterface,
}

// Count event types as known non-targets to detect unresolved members that
// would otherwise be masked by an event in an intersection.
var knownTypeMembers = unicornutil.TypeClassifierOptions{
	NonTargetTypeNames:  eventTypeNames,
	HeritageSymbolFlags: eventTypes.HeritageSymbolFlags,
	ClassifyTypeImport: func(moduleName, exportName string) unicornutil.TypeClass {
		if classifyEventTypeImport(moduleName, exportName) == unicornutil.TypeUnknown {
			return unicornutil.TypeUnknown
		}
		return unicornutil.TypeNonTarget
	},
}

func classifyEventTypeImport(moduleName, exportName string) unicornutil.TypeClass {
	if moduleName == "react" {
		if reactEventTypeNames.Has(exportName) {
			return unicornutil.TypeTarget
		}
		return unicornutil.TypeNonTarget
	}
	if eventTypeNames.Has(exportName) {
		return unicornutil.TypeNonTarget
	}
	return unicornutil.TypeUnknown
}

func isEventParameter(ctx rule.RuleContext, parameter, binding *ast.Node) bool {
	class := unicornutil.TypeUnknown
	var annotation *ast.Node
	if binding == parameter.Name() {
		annotation, _ = unicornutil.ValueTypeAnnotation(ctx, parameter)
		if annotation != nil {
			class = classifyEventAnnotation(ctx, annotation)
		}
	}
	if ctx.TypeChecker != nil {
		switch {
		case class == unicornutil.TypeUnknown:
			class = classifyEventType(ctx, ctx.TypeChecker.GetTypeAtLocation(binding))
		case class == unicornutil.TypeTarget && !utils.IsJSDocSyntaxNode(annotation):
			// An alias can hide an any type that only TypeScript can resolve.
			resolved := ctx.TypeChecker.GetTypeAtLocation(annotation)
			if utils.IsTypeAnyType(resolved) && !utils.IsIntrinsicErrorType(resolved) {
				class = unicornutil.TypeUnknown
			}
		}
	}
	return class == unicornutil.TypeTarget ||
		(class == unicornutil.TypeUnknown && eventParameterName.MatchString(binding.Text()))
}

func classifyEventAnnotation(ctx rule.RuleContext, annotation *ast.Node) unicornutil.TypeClass {
	class := unicornutil.ClassifyTypeAnnotation(ctx, annotation, eventTypes)
	if class == unicornutil.TypeTarget &&
		unicornutil.ClassifyTypeAnnotation(ctx, annotation, knownTypeMembers) == unicornutil.TypeUnknown {
		return unicornutil.TypeUnknown
	}
	return class
}

func classifyEventType(ctx rule.RuleContext, typ *checker.Type) unicornutil.TypeClass {
	class := unicornutil.ClassifyType(ctx, typ, eventTypes)
	if class == unicornutil.TypeTarget &&
		unicornutil.ClassifyType(ctx, typ, unicornutil.TypeClassifierOptions{}) == unicornutil.TypeUnknown {
		return unicornutil.TypeUnknown
	}
	return class
}
