package output

import "github.com/fatih/color"

type colorScheme struct {
	RuleName    func(format string, a ...interface{}) string
	FileName    func(format string, a ...interface{}) string
	ErrorText   func(format string, a ...interface{}) string
	SuccessText func(format string, a ...interface{}) string
	DimText     func(format string, a ...interface{}) string
	BorderText  func(format string, a ...interface{}) string
	WarnText    func(format string, a ...interface{}) string
}

func newPinnedColor(enabled bool, attrs ...color.Attribute) *color.Color {
	c := color.New(attrs...)
	if enabled {
		c.EnableColor()
	} else {
		c.DisableColor()
	}
	return c
}

func newColorScheme(enabled bool) colorScheme {
	return colorScheme{
		RuleName:    newPinnedColor(enabled, color.FgHiGreen).SprintfFunc(),
		FileName:    newPinnedColor(enabled, color.FgCyan, color.Italic).SprintfFunc(),
		ErrorText:   newPinnedColor(enabled, color.FgRed, color.Bold).SprintfFunc(),
		SuccessText: newPinnedColor(enabled, color.FgGreen, color.Bold).SprintfFunc(),
		DimText:     newPinnedColor(enabled, color.Faint).SprintfFunc(),
		BorderText:  newPinnedColor(enabled, color.Faint).SprintfFunc(),
		WarnText:    newPinnedColor(enabled, color.FgYellow).SprintfFunc(),
	}
}
