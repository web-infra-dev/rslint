package config

// bundledPluginDeclaration maps accepted string declarations to the namespace
// used by that bundled plugin's rules. It contains no rule implementations;
// callers supply those through a rule catalog.
type bundledPluginDeclaration struct {
	ruleNamespace    string
	declarationNames []string
}

var bundledPluginDeclarations = []bundledPluginDeclaration{
	{ruleNamespace: "@typescript-eslint", declarationNames: []string{"@typescript-eslint"}},
	{ruleNamespace: "import", declarationNames: []string{"eslint-plugin-import", "import"}},
	{ruleNamespace: "jest", declarationNames: []string{"eslint-plugin-jest", "jest"}},
	{ruleNamespace: "jsx-a11y", declarationNames: []string{"eslint-plugin-jsx-a11y", "jsx-a11y"}},
	{ruleNamespace: "promise", declarationNames: []string{"eslint-plugin-promise", "promise"}},
	{ruleNamespace: "react", declarationNames: []string{"react"}},
	{ruleNamespace: "react-hooks", declarationNames: []string{"eslint-plugin-react-hooks", "react-hooks"}},
	{ruleNamespace: "rstest", declarationNames: []string{"rstest"}},
	{ruleNamespace: "unicorn", declarationNames: []string{"eslint-plugin-unicorn", "unicorn"}},
}

var bundledPluginByDeclarationName = func() map[string]bundledPluginDeclaration {
	byName := make(map[string]bundledPluginDeclaration)
	for _, plugin := range bundledPluginDeclarations {
		for _, name := range plugin.declarationNames {
			byName[name] = plugin
		}
	}
	return byName
}()

// NormalizePluginName converts a plugin declaration name to its rule prefix form.
// Unknown declaration names are returned unchanged.
func NormalizePluginName(pluginName string) string {
	if plugin, ok := bundledPluginByDeclarationName[pluginName]; ok {
		return plugin.ruleNamespace
	}
	return pluginName
}
