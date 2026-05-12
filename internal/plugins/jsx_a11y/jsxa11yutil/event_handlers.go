package jsxa11yutil

// Event handler name groups mirror jsx-ast-utils'
// `eventHandlersByType` map. Each group corresponds to a single key in that
// upstream map; rules that need a multi-group default (e.g. the
// no-noninteractive-element-interactions / no-static-element-interactions
// defaults of `focus + image + keyboard + mouse`) compose these slices
// themselves rather than depending on a fixed shape.
//
// Source: https://github.com/jsx-eslint/jsx-ast-utils/blob/main/src/eventHandlers.js
// Order within each group matches the upstream array — semantically
// irrelevant (every consumer uses set membership) but kept for auditability.
var (
	// EventHandlersMouse mirrors `eventHandlersByType.mouse`.
	EventHandlersMouse = []string{
		"onClick",
		"onContextMenu",
		"onDblClick",
		"onDoubleClick",
		"onDrag",
		"onDragEnd",
		"onDragEnter",
		"onDragExit",
		"onDragLeave",
		"onDragOver",
		"onDragStart",
		"onDrop",
		"onMouseDown",
		"onMouseEnter",
		"onMouseLeave",
		"onMouseMove",
		"onMouseOut",
		"onMouseOver",
		"onMouseUp",
	}

	// EventHandlersKeyboard mirrors `eventHandlersByType.keyboard`.
	EventHandlersKeyboard = []string{
		"onKeyDown",
		"onKeyPress",
		"onKeyUp",
	}

	// EventHandlersFocus mirrors `eventHandlersByType.focus`.
	EventHandlersFocus = []string{
		"onFocus",
		"onBlur",
	}

	// EventHandlersImage mirrors `eventHandlersByType.image`.
	EventHandlersImage = []string{
		"onLoad",
		"onError",
	}
)

// InteractiveEventHandlerNames mirrors the
// `[].concat(eventHandlersByType.mouse, eventHandlersByType.keyboard)`
// constant that `interactive-supports-focus` (and any future rule that
// imports `interactiveProps` from jsx-ast-utils) hands to `hasAnyProp`.
//
// Order matches the upstream concat (mouse first, then keyboard) for
// auditability — semantically irrelevant because every consumer uses set
// membership rather than ordering.
var InteractiveEventHandlerNames = func() []string {
	out := make([]string, 0, len(EventHandlersMouse)+len(EventHandlersKeyboard))
	out = append(out, EventHandlersMouse...)
	out = append(out, EventHandlersKeyboard...)
	return out
}()

// DefaultStaticInteractionHandlers mirrors upstream's
// `defaultInteractiveProps` constant in `no-static-element-interactions`:
//
//	[].concat(
//	  eventHandlersByType.focus,
//	  eventHandlersByType.keyboard,
//	  eventHandlersByType.mouse,
//	);
//
// Composed from the per-group slices above to keep the source-of-truth in one
// place. Order matches the upstream concat (focus → keyboard → mouse);
// consumers iterate to check `hasProp + non-null value`, so order is
// semantically irrelevant.
var DefaultStaticInteractionHandlers = func() []string {
	out := make([]string, 0, len(EventHandlersFocus)+len(EventHandlersKeyboard)+len(EventHandlersMouse))
	out = append(out, EventHandlersFocus...)
	out = append(out, EventHandlersKeyboard...)
	out = append(out, EventHandlersMouse...)
	return out
}()
