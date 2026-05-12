package jsxa11yutil

// InteractiveEventHandlerNames mirrors the
// `[].concat(eventHandlersByType.mouse, eventHandlersByType.keyboard)`
// constant that `interactive-supports-focus` (and any future rule that
// imports `interactiveProps` from jsx-ast-utils) hands to `hasAnyProp`.
//
// Sourced from jsx-ast-utils' `eventHandlers.js`:
//
//	mouse: [
//	  'onClick', 'onContextMenu', 'onDblClick', 'onDoubleClick',
//	  'onDrag', 'onDragEnd', 'onDragEnter', 'onDragExit',
//	  'onDragLeave', 'onDragOver', 'onDragStart', 'onDrop',
//	  'onMouseDown', 'onMouseEnter', 'onMouseLeave', 'onMouseMove',
//	  'onMouseOut', 'onMouseOver', 'onMouseUp',
//	],
//	keyboard: ['onKeyDown', 'onKeyPress', 'onKeyUp'],
//
// Order matches the upstream concat (mouse first, then keyboard) for
// auditability — semantically irrelevant because every consumer uses set
// membership rather than ordering.
var InteractiveEventHandlerNames = []string{
	// mouse
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
	// keyboard
	"onKeyDown",
	"onKeyPress",
	"onKeyUp",
}
