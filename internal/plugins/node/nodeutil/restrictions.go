package nodeutil

import (
	"strings"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type restrictionPattern struct {
	matcher           *GlobMatcher
	absolute, negated bool
}

type restriction struct {
	patterns []restrictionPattern
	message  string
}

// Restrictions shares the ordered module restrictions used by import and require.
type Restrictions []restriction

// ParseRestrictions compiles the validated no-restricted-import/require options.
func ParseRestrictions(options []any) Restrictions {
	if len(options) == 0 {
		return nil
	}
	definitions, _ := options[0].([]any)
	restrictions := make(Restrictions, 0, len(definitions))
	for _, definition := range definitions {
		var names []string
		var message string
		switch value := definition.(type) {
		case string:
			names = []string{value}
		case map[string]any:
			if name, ok := value["name"].(string); ok {
				names = []string{name}
			} else {
				names = utils.ToStringSlice(value["name"])
			}
			message, _ = value["message"].(string)
		}
		if message != "" {
			message = " " + message
		}
		r := restriction{message: message}
		for _, name := range names {
			negated := strings.HasPrefix(name, "!") && !strings.HasPrefix(name, "!(")
			if negated {
				name = name[1:]
			}
			// Preserve the pattern: changing separators changes glob semantics.
			r.patterns = append(r.patterns, restrictionPattern{CompileGlob(name), IsAbsolutePath(name), negated})
		}
		restrictions = append(restrictions, r)
	}
	return restrictions
}

// Match returns the first matching restriction. Paths are resolved at most once,
// and only when an absolute pattern needs them. Callers select import/require resolution.
func (restrictions Restrictions) Match(name string, resolvePath func() string) *rule.RuleMessage {
	var filePath string
	resolved := false
	for _, restriction := range restrictions {
		matched := false
		for _, pattern := range restriction.patterns {
			// Positive patterns add matches, negatives remove them in order.
			if pattern.negated != matched {
				continue
			}
			target := name
			if pattern.absolute {
				if !resolved {
					filePath = resolvePath()
					resolved = true
				}
				if filePath == "" {
					continue
				}
				target = filePath
			}
			if pattern.matcher.Match(target) {
				matched = !pattern.negated
			}
		}
		if matched {
			return &rule.RuleMessage{
				Id: "restricted", Description: "'" + name + "' module is restricted from being used." + restriction.message,
				Data: map[string]string{"name": name, "customMessage": restriction.message},
			}
		}
	}
	return nil
}
