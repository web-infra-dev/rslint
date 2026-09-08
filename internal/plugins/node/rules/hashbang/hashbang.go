package hashbang

import (
	_ "embed"
	"path"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed hashbang.schema.json
var schemaJSON []byte

const envShebang = "#!/usr/bin/env"

var (
	envFlags = esregexp.MustCompile(`^\s*-(-.*?\b|[ivS]+|[Pu](\s+|=)\S+)(?=\s|$)`, "")
	envVars  = esregexp.MustCompile(`^\s*\w+=(?:"(?:[^"\\]|\\.)*"|\w+)`, "")
)

func isNodeShebang(shebang, executable string) bool {
	if shebang == "" {
		return false
	}
	if index := strings.Index(shebang, envShebang); index >= 0 {
		// The end of this ASCII match is also the JavaScript slice boundary.
		shebang = shebang[index+len(envShebang):]
	} else {
		// Preserve upstream's UTF-16 slice even for a non-env interpreter.
		units := ecmascript.StringCodeUnits(shebang)
		shebang = ecmascript.StringFromCodeUnits(units[min(len(envShebang)-1, len(units)):])
	}
	for {
		previous := shebang
		for _, expression := range []*esregexp.RegExp{envFlags, envVars} {
			match, _ := expression.Unwrap().FindStringMatch(shebang)
			if match != nil {
				shebang = shebang[len(match.String()):]
			}
		}
		if previous == shebang {
			break
		}
	}
	command, _, _ := strings.Cut(ecmascript.StringTrim(shebang), " ")
	return command == executable
}

func fileExtension(fileName string) string {
	base := path.Base(fileName)
	if strings.HasPrefix(base, ".") && !strings.Contains(base[1:], ".") {
		return ""
	}
	return path.Ext(fileName)
}

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/hashbang.js
var HashbangRule = rule.Rule{
	Name:   "node/hashbang",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		program := ctx.Program()
		if program == nil || ctx.SourceFile.FileName() == "<input>" {
			return nil
		}
		fileName := ctx.SourceFile.FileName()
		pkg := nodeutil.FindPackage(program, fileName)
		if pkg == nil {
			return nil
		}
		opts := map[string]any{}
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		relative := tspath.GetRelativePathFromDirectory(pkg.Directory(), fileName, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
		converted, ok := nodeutil.ConvertPath(relative, opts, ctx.Settings)
		if !ok {
			return nil
		}
		absolute := tspath.ResolvePath(pkg.Directory(), converted)
		additional := nodeutil.MatchIgnorePatterns(program, utils.ToStringSlice(opts["additionalExecutables"]), converted)
		if ignore, _ := opts["ignoreUnpublished"].(bool); ignore && !additional && nodeutil.IsUnpublished(program, absolute, converted) {
			return nil
		}
		needsShebang := additional || pkg.IsBinFile(absolute)
		executable := "node"
		if mapping, ok := opts["executableMap"].(map[string]any); ok {
			if value, ok := mapping[fileExtension(fileName)].(string); ok {
				executable = value
			}
		}

		text := ctx.SourceFile.Text()
		shebang := scanner.GetShebang(text)
		lineEnd := len(shebang)
		if shebang == "" {
			lineEnd = strings.IndexAny(text, "\r\n\u2028\u2029")
			if lineEnd < 0 {
				lineEnd = len(text)
			}
		}
		length, cr := 0, false
		ending := ecmascript.LineTerminatorSequenceAt(text, lineEnd)
		if shebang != "" {
			// Replacing an unterminated or empty hashbang must consume the
			// existing header, rather than insert a second, invalid hashbang.
			length = lineEnd + len(ending)
		}
		// Upstream's /^(#!.+?)?(\r)?\n/u also consumes an empty first
		// line, but requires content after #! and an LF or CRLF ending.
		if (len(shebang) > 2 || lineEnd == 0) && (ending == "\n" || ending == "\r\n") {
			length = lineEnd + len(ending)
			cr = ending == "\r\n"
			if shebang != "" {
				shebang += "\n"
			}
		} else {
			shebang = ""
		}
		location := core.NewTextRange(0, lineEnd)
		report := func(id, description string, fixRange core.TextRange, replacement string) {
			message := rule.RuleMessage{Id: id, Description: description}
			if id == "expectedHashbangNode" {
				message.Data = map[string]string{"executableName": executable}
			}
			ctx.ReportRangeWithDeferredFixes(location, message, func() []rule.RuleFix {
				return []rule.RuleFix{rule.RuleFixReplaceRange(fixRange, replacement)}
			})
		}
		switch {
		case needsShebang && isNodeShebang(shebang, executable):
			if ctx.HasBOM() {
				report("unexpectedBOM", "This file must not have Unicode BOM.", core.NewTextRange(-1, 0), "")
			}
			if cr {
				index := strings.IndexByte(text, '\r')
				report("expectedLF", "This file must have Unix linebreaks (LF).", core.NewTextRange(index, index+1), "")
			}
		case needsShebang:
			report("expectedHashbangNode", `This file needs shebang "#!/usr/bin/env `+executable+`".`, core.NewTextRange(-1, length), envShebang+" "+executable+"\n")
		case shebang != "":
			report("expectedHashbang", "This file needs no shebang.", core.NewTextRange(0, length), "")
		}
		return nil
	},
}
