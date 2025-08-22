package code_path_analyzer

type CodePath struct {
	id             string         // An identifier
	origin         string         // The type of code path origin
	upper          *CodePath      // The code path of the upper function scope
	onLooped       func()         // A callback funciton to notify looping
	childCodePaths []*CodePath    // The code paths of nested function scopes
	state          *CodePathState // The state of the code path
}
