package ecmascript

import "unicode"

// The general categories a rule asks a character about are the ones a
// JavaScript regexp names with \p{...}, and they are read from the standard
// library's tables. Node 26 carries ICU 78 and Go 1.27 carries Unicode 17.0, so
// the two agree; the questions are asked through here rather than of the
// unicode package directly so that a rule keeps reading characters through one
// door, and so an edition the toolchain has not caught up with is one package
// to correct rather than every caller.

// IsUpper reports whether r is an uppercase letter — the general category Lu,
// which a JavaScript regexp writes \p{Lu}.
func IsUpper(r rune) bool { return unicode.IsUpper(r) }

// IsLower says for the category Ll what [IsUpper] says for Lu.
func IsLower(r rune) bool { return unicode.IsLower(r) }

// IsLetter says for the category L — a letter of any case — what [IsUpper] says
// for Lu.
func IsLetter(r rune) bool { return unicode.IsLetter(r) }

// IsMark says for the category M — the combining marks — what [IsUpper] says
// for Lu.
func IsMark(r rune) bool { return unicode.Is(unicode.M, r) }
