package ecmascript

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	jsxtx "github.com/microsoft/typescript-go/shim/transformers/jsxtransforms"
)

// DecodeJSXEntities decodes the character references that JSX parsers expose
// as decoded text in direct attribute values. Invalid or unterminated
// references remain in the result, matching the JSX token reader.
func DecodeJSXEntities(raw string) string {
	units := make([]jsxDecodedUnit, 0, len(raw))
	for pos := 0; pos < len(raw); {
		if raw[pos] == '&' {
			first, second, count, next, ok := decodeJSXTextEntity(raw, pos)
			if ok {
				units = append(units, jsxDecodedUnit{value: first})
				if count == 2 {
					units = append(units, jsxDecodedUnit{value: second})
				}
				pos = next
				continue
			}
			units = append(units, jsxDecodedUnit{value: '&'})
			pos++
			continue
		}

		r, size := decodeStringRune(raw[pos:])
		if r > 0xFFFF {
			high, low := utf16.EncodeRune(r)
			units = append(units,
				jsxDecodedUnit{value: uint16(high)},
				jsxDecodedUnit{value: uint16(low)},
			)
		} else {
			units = append(units, jsxDecodedUnit{value: uint16(r)})
		}
		pos += size
	}

	var decoded strings.Builder
	for pos := 0; pos < len(units); {
		unit := units[pos]
		r := rune(unit.value)
		if pos+1 < len(units) && r >= 0xD800 && r <= 0xDBFF &&
			units[pos+1].value >= 0xDC00 && units[pos+1].value <= 0xDFFF {
			r = utf16.DecodeRune(r, rune(units[pos+1].value))
			pos++
		}
		if r >= 0xD800 && r <= 0xDFFF {
			r = utf8.RuneError
		}
		decoded.WriteRune(r)
		pos++
	}
	return decoded.String()
}

type jsxDecodedUnit struct {
	value uint16
}

// JSXTextTokenValuesEqual reports whether two raw JSX text spans produce the
// same token value in Espree. acorn-jsx decodes XHTML entities and folds source
// CRLF pairs to LF while reading JSX text. Numeric entities use
// String.fromCharCode semantics, so comparison has to happen in UTF-16 code
// units rather than Go runes.
func JSXTextTokenValuesEqual(left, right string) bool {
	if left == right {
		return true
	}

	leftUnits := jsxTextCodeUnits{raw: left}
	rightUnits := jsxTextCodeUnits{raw: right}
	for {
		leftUnit, leftOK := leftUnits.next()
		rightUnit, rightOK := rightUnits.next()
		if leftOK != rightOK {
			return false
		}
		if !leftOK {
			return true
		}
		if leftUnit != rightUnit {
			return false
		}
	}
}

type jsxTextCodeUnits struct {
	raw        string
	pos        int
	pending    uint16
	hasPending bool
}

func (it *jsxTextCodeUnits) next() (uint16, bool) {
	if it.hasPending {
		it.hasPending = false
		return it.pending, true
	}
	if it.pos >= len(it.raw) {
		return 0, false
	}

	switch it.raw[it.pos] {
	case '&':
		first, second, count, next, ok := decodeJSXTextEntity(it.raw, it.pos)
		if ok {
			it.pos = next
			if count == 2 {
				it.pending = second
				it.hasPending = true
			}
			return first, true
		}
		// acorn-jsx restores its cursor to immediately after an invalid or
		// unterminated entity's ampersand, then reads the remaining text again.
		it.pos++
		return '&', true
	case '\r':
		if it.pos+1 < len(it.raw) && it.raw[it.pos+1] == '\n' {
			it.pos += 2
			return '\n', true
		}
	}

	r, size := decodeStringRune(it.raw[it.pos:])
	it.pos += size
	if r > 0xFFFF {
		high, low := utf16.EncodeRune(r)
		it.pending = uint16(low)
		it.hasPending = true
		return uint16(high), true
	}
	return uint16(r), true
}

// decodeJSXTextEntity mirrors acorn-jsx's jsx_readEntity. The semicolon must
// occur within ten UTF-16 code units after the ampersand, and is itself one of
// those ten units. Invalid entities leave the caller positioned after '&'.
func decodeJSXTextEntity(raw string, ampersand int) (first, second uint16, count, next int, ok bool) {
	entityStart := ampersand + 1
	pos := entityStart
	unitsRead := 0
	for pos < len(raw) && unitsRead < 10 {
		r, size := decodeStringRune(raw[pos:])
		width := 1
		if r > 0xFFFF {
			width = 2
		}
		if unitsRead+width > 10 {
			break
		}
		unitsRead += width
		pos += size
		if r != ';' {
			continue
		}

		first, second, count, ok = decodeJSXEntityValue(raw[entityStart : pos-size])
		if !ok {
			return 0, 0, 0, 0, false
		}
		return first, second, count, pos, true
	}
	return 0, 0, 0, 0, false
}

func decodeJSXEntityValue(entity string) (first, second uint16, count int, ok bool) {
	if len(entity) == 0 {
		return 0, 0, 0, false
	}
	decoded, ok := jsxtx.DecodeEntity(entity)
	if !ok {
		return 0, 0, 0, false
	}
	if entity[0] == '#' {
		// acorn-jsx turns numeric entities into exactly one UTF-16 code unit
		// with String.fromCharCode, including values above the BMP.
		return uint16(decoded), 0, 1, true
	}
	if decoded <= 0xFFFF {
		return uint16(decoded), 0, 1, true
	}
	high, low := utf16.EncodeRune(decoded)
	return uint16(high), uint16(low), 2, true
}
