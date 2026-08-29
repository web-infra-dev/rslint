package typescriptutil

import "github.com/web-infra-dev/rslint/internal/utils/ecmascript"

// JSXFactoryRoot returns the local binding name consumed implicitly by a JSX
// factory option such as "React.createElement" or "h".
func JSXFactoryRoot(factory string) string {
	for index, char := range factory {
		if char == '.' {
			return ecmascript.StringTrim(factory[:index])
		}
	}
	return ecmascript.StringTrim(factory)
}
