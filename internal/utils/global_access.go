package utils

// GlobalAccess is the access level of a name declared through
// `languageOptions.globals` or a `/* global */` comment, mirroring ESLint's
// three settings. The zero value is GlobalAccessUnset, so reading a name that
// neither source mentions yields "no opinion" rather than a real setting.
type GlobalAccess uint8

const (
	// GlobalAccessUnset means no config entry or comment mentions the name.
	// Callers fall back to whatever they already know about it, e.g.
	// IsECMAScriptGlobal.
	GlobalAccessUnset GlobalAccess = iota
	// GlobalAccessOff removes the name from the global scope, including
	// built-ins that would otherwise be declared.
	GlobalAccessOff
	// GlobalAccessReadonly declares the name and forbids assigning to it.
	GlobalAccessReadonly
	// GlobalAccessWritable declares the name and allows assigning to it.
	GlobalAccessWritable
)

// IsDeclared reports whether the name resolves to a global variable.
func (access GlobalAccess) IsDeclared() bool {
	return access == GlobalAccessReadonly || access == GlobalAccessWritable
}

// IsWritable reports whether assigning to the name is allowed.
func (access GlobalAccess) IsWritable() bool {
	return access == GlobalAccessWritable
}

func (access GlobalAccess) String() string {
	switch access {
	case GlobalAccessOff:
		return "off"
	case GlobalAccessReadonly:
		return "readonly"
	case GlobalAccessWritable:
		return "writable"
	default:
		return "unset"
	}
}

// NormalizeGlobalAccess maps an authored `languageOptions.globals` value to its
// access level, accepting every alias ESLint does. The `globals` npm package
// spells the two declared levels as booleans: true is writable, false is
// readonly. ok is false for values ESLint rejects.
func NormalizeGlobalAccess(value any) (access GlobalAccess, ok bool) {
	switch value := value.(type) {
	case nil:
		return GlobalAccessReadonly, true
	case bool:
		if value {
			return GlobalAccessWritable, true
		}
		return GlobalAccessReadonly, true
	case string:
		switch value {
		case "true", "writable", "writeable":
			return GlobalAccessWritable, true
		case "false", "readonly", "readable":
			return GlobalAccessReadonly, true
		case "off":
			return GlobalAccessOff, true
		}
	}
	return GlobalAccessUnset, false
}

// NormalizeInlineGlobalAccess maps the setting written after a name in a
// `/* global name:setting */` comment. A name with no setting at all is
// readonly; every other value goes through the shared alias table, where a
// bare `name:` (empty setting) is as invalid as `name:nonsense`.
func NormalizeInlineGlobalAccess(setting string, hasSetting bool) (access GlobalAccess, ok bool) {
	if !hasSetting {
		return GlobalAccessReadonly, true
	}
	return NormalizeGlobalAccess(setting)
}
