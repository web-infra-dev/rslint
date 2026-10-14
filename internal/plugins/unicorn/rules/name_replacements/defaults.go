package name_replacements

type defaultReplacementPair struct {
	name, replacement string
}

func newDefaultReplacements() map[string]map[string]bool {
	result := make(map[string]map[string]bool, 92)
	for _, pair := range defaultReplacementPairs {
		replacements := result[pair.name]
		if replacements == nil {
			replacements = make(map[string]bool)
			result[pair.name] = replacements
		}
		replacements[pair.replacement] = true
	}
	return result
}
