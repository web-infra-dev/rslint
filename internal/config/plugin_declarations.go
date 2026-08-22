package config

// nativePluginDeclaration maps the names accepted in JSON configuration to
// the namespace used by that native plugin's rules. It deliberately contains
// no rule implementations; callers supply those through a rule catalog.
type nativePluginDeclaration struct {
	rulePrefix       string
	declarationNames []string
}

var nativePluginDeclarations = []nativePluginDeclaration{
	{rulePrefix: "@typescript-eslint", declarationNames: []string{"@typescript-eslint"}},
	{rulePrefix: "import", declarationNames: []string{"eslint-plugin-import", "import"}},
	{rulePrefix: "jest", declarationNames: []string{"eslint-plugin-jest", "jest"}},
	{rulePrefix: "jsx-a11y", declarationNames: []string{"eslint-plugin-jsx-a11y", "jsx-a11y"}},
	{rulePrefix: "promise", declarationNames: []string{"eslint-plugin-promise", "promise"}},
	{rulePrefix: "react", declarationNames: []string{"react"}},
	{rulePrefix: "react-hooks", declarationNames: []string{"eslint-plugin-react-hooks", "react-hooks"}},
	{rulePrefix: "rstest", declarationNames: []string{"rstest"}},
	{rulePrefix: "unicorn", declarationNames: []string{"eslint-plugin-unicorn", "unicorn"}},
}

var nativePluginByDeclarationName = func() map[string]nativePluginDeclaration {
	byName := make(map[string]nativePluginDeclaration)
	for _, plugin := range nativePluginDeclarations {
		for _, name := range plugin.declarationNames {
			byName[name] = plugin
		}
	}
	return byName
}()

// NormalizePluginName converts a plugin declaration name to its rule prefix form.
// Unknown declaration names are returned unchanged.
func NormalizePluginName(pluginName string) string {
	if plugin, ok := nativePluginByDeclarationName[pluginName]; ok {
		return plugin.rulePrefix
	}
	return pluginName
}
