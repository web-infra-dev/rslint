package no_late_event_control

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
	if binding == parameter.Name() {
		annotation, _ := unicornutil.ValueTypeAnnotation(ctx, parameter)
		if annotation != nil {
			class = unicornutil.ClassifyTypeAnnotation(ctx, annotation, eventTypes)
		}
	}
	if class == unicornutil.TypeUnknown && ctx.TypeChecker != nil {
		class = unicornutil.ClassifyType(ctx, ctx.TypeChecker.GetTypeAtLocation(binding), eventTypes)
	}
	return class == unicornutil.TypeTarget ||
		(class == unicornutil.TypeUnknown && eventParameterName.MatchString(binding.Text()))
}
