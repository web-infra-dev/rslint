package code_path_analysis

type CodePath struct {
	id             string                                                         // An identifier
	origin         string                                                         // The type of code path origin
	upper          *CodePath                                                      // The code path of the upper function scope
	onLooped       func(fromSegment *CodePathSegment, toSegment *CodePathSegment) // A callback funciton to notify looping
	childCodePaths []*CodePath                                                    // The code paths of nested function scopes
	state          *CodePathState                                                 // The state of the code path
}

func NewCodePath(id string, origin string, upper *CodePath, onLooped func(fromSegment *CodePathSegment, toSegment *CodePathSegment)) *CodePath {
	codePath := &CodePath{
		id:             id,
		origin:         origin,
		upper:          upper,
		onLooped:       onLooped,
		childCodePaths: make([]*CodePath, 0),
		state:          NewCodePathState(NewIdGenerator(id+"_"), onLooped),
	}
	// Adds this into `childCodePaths` of `upper`.
	if upper != nil {
		upper.childCodePaths = append(upper.childCodePaths, codePath)
	}
	return codePath
}

// Getter methods for accessing private fields

func (cp *CodePath) ID() string {
	return cp.id
}

func (cp *CodePath) Origin() string {
	return cp.origin
}

func (cp *CodePath) Upper() *CodePath {
	return cp.upper
}

func (cp *CodePath) ChildCodePaths() []*CodePath {
	return cp.childCodePaths
}

func (cp *CodePath) State() *CodePathState {
	return cp.state
}

func (cp *CodePath) InitialSegment() *CodePathSegment {
	if cp.state != nil {
		return cp.state.InitialSegment()
	}
	return nil
}

func (cp *CodePath) FinalSegments() []*CodePathSegment {
	if cp.state != nil {
		return cp.state.FinalSegments()
	}
	return nil
}

func (cp *CodePath) ThrownSegments() []*CodePathSegment {
	if cp.state != nil {
		return cp.state.ThrownSegments()
	}
	return nil
}

// Helper function to check if a segment is in thrown segments
func (cp *CodePath) HasThrownSegment(segment *CodePathSegment) bool {
	thrownSegments := cp.ThrownSegments()
	if thrownSegments == nil {
		return false
	}

	for _, thrownSegment := range thrownSegments {
		if thrownSegment.ID() == segment.ID() {
			return true
		}
	}
	return false
}
