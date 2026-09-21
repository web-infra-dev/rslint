// cspell:ignore Chrs Cpmn Diak Dogr Elym Hmnp Hrkt Maka Medf Nagm Nand Ougr Rohg Sogd Sogo Tnsa Vith Wcho Yezi dsuvy
package es_syntax

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (c *syntaxChecker) regexpCall(node *ast.Node) {
	args := node.Arguments()
	pattern, flags := "", ""
	if len(args) > 0 {
		pattern, _ = c.static().EvalToString(args[0])
	}
	if len(args) > 1 {
		flags, _ = c.static().EvalToString(args[1])
	}
	c.regexp(node, pattern, flags)
}

func (c *syntaxChecker) regexp(node *ast.Node, pattern, flags string) {
	found := map[string]bool{}
	for _, flag := range "dsuvy" {
		if strings.ContainsRune(flags, flag) {
			found["regexp-"+string(flag)+"-flag"] = true
		}
	}
	report := func() {
		for _, feature := range features {
			if found[feature.name] {
				c.report(feature.name, node)
			}
		}
	}

	// es-x 7.8.0 validates patterns with only its Unicode (u) switch, even for v.
	parsedFlags := utils.RegexFlags{Unicode: strings.Contains(flags, "u")}
	if !utils.IsValidRegexPattern(pattern, parsedFlags) {
		report()
		return
	}
	inClass := false
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '\\':
			if parsedFlags.Unicode && i+2 < len(pattern) && (pattern[i+1] == 'p' || pattern[i+1] == 'P') && pattern[i+2] == '{' {
				if end := strings.IndexByte(pattern[i+3:], '}'); end >= 0 {
					property := pattern[i+3 : i+3+end]
					found["regexp-unicode-property-escapes"] = true
					key, value, pair := strings.Cut(property, "=")
					for _, entry := range unicodeProperties {
						if pair && strings.Contains(" Script Script_Extensions sc scx ", " "+key+" ") && strings.Contains(" "+entry.scripts+" ", " "+value+" ") || !pair && strings.Contains(" "+entry.binary+" ", " "+key+" ") {
							found["regexp-unicode-property-escapes-"+entry.year] = true
						}
					}
				}
			}
			step, ok := utils.SkipPatternEscape(pattern, i, parsedFlags)
			if !ok {
				return
			}
			i += step
		case '[':
			inClass = true
			i++
		case ']':
			inClass = false
			i++
		default:
			if !inClass && strings.HasPrefix(pattern[i:], "(?<") && i+3 < len(pattern) {
				if pattern[i+3] == '=' || pattern[i+3] == '!' {
					found["regexp-lookbehind-assertions"] = true
				} else {
					found["regexp-named-capture-groups"] = true
				}
			}
			i++
		}
	}
	report()
}

// New property names in the editions tracked by es-x 7.8.0. Validation and
// escape handling stay in the existing ECMAScript regular-expression helpers.
var unicodeProperties = []struct{ year, scripts, binary string }{
	{"2019", "Dogr Dogra Gong Gunjala_Gondi Hanifi_Rohingya Maka Makasar Medefaidrin Medf Old_Sogdian Rohg Sogd Sogdian Sogo", "Extended_Pictographic"},
	{"2020", "Elym Elymaic Hmnp Nand Nandinagari Nyiakeng_Puachue_Hmong Wancho Wcho", ""},
	{"2021", "Chorasmian Chrs Diak Dives_Akuru Khitan_Small_Script Kits Yezi Yezidi", "EBase EComp EMod EPres ExtPict"},
	{"2022", "Cpmn Cypro_Minoan Old_Uyghur Ougr Tangsa Tnsa Toto Vith Vithkuqi", ""},
	{"2023", "Hrkt Katakana_Or_Hiragana Kawi Nag_Mundari Nagm Unknown Zzzz", ""},
}
