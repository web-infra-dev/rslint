package code_path_analysis

type CodePathState struct {
	idGenerator   *IdGenerator // idGenerator An id generator to generate id for code
	forkContext   *ForkContext
	notifyLooped  func(fromSegment *CodePathSegment, toSegment *CodePathSegment)
	choiceContext *ChoiceContext
	chainContext  *ChainContext
	breakContext  *BreakContext
	switchContext *SwitchContext
	tryContext    *TryContext
	loopContext   *LoopContext

	currentSegments  []*CodePathSegment
	initialSegment   *CodePathSegment
	finalSegments    []*CodePathSegment
	returnedSegments []*CodePathSegment
	thrownSegments   []*CodePathSegment
}

func NewCodePathState(idGenerator *IdGenerator, onLooped func(fromSegment *CodePathSegment, toSegment *CodePathSegment)) *CodePathState {
	forkContext := NewRootForkContext(idGenerator)
	return &CodePathState{
		idGenerator:      idGenerator,
		notifyLooped:     onLooped,
		forkContext:      forkContext,
		currentSegments:  make([]*CodePathSegment, 0),
		initialSegment:   forkContext.Head()[0],
		finalSegments:    make([]*CodePathSegment, 0),
		returnedSegments: make([]*CodePathSegment, 0),
		thrownSegments:   make([]*CodePathSegment, 0),
	}
}

// The head segments.
func (s *CodePathState) HeadSegments() []*CodePathSegment {
	return s.forkContext.Head()
}

// The parent forking context. This is used for the root of new forks.
func (s *CodePathState) ParentForkContext() *ForkContext {
	current := s.forkContext
	if current == nil {
		return nil
	}
	return current.upper
}

// Creates and stacks new forking context.
func (s *CodePathState) PushForkContext(forkLeavingPath *ForkContext) *ForkContext {
	s.forkContext = NewEmptyForkContext(s.forkContext, forkLeavingPath)
	return s.forkContext
}

// Pops and merges the last forking context.
func (s *CodePathState) PopForkContext() *ForkContext {
	lastContext := s.forkContext

	s.forkContext = lastContext.upper
	s.forkContext.ReplaceHead(lastContext.MakeNext(0, -1))

	return lastContext
}

// Creates a new path.
func (s *CodePathState) ForkPath() {
	s.forkContext.Add(s.ParentForkContext().MakeNext(-1, -1))
}

// Creates a bypass path.
func (s *CodePathState) ForkBypassPath() {
	s.forkContext.Add(s.ParentForkContext().Head())
}

// Creates looping path.
func (s *CodePathState) MakeLooped(unflattenedFromSegments []*CodePathSegment, unflattenedToSegments []*CodePathSegment) {
	fromSegments := flattenUnusedSegments(
		unflattenedFromSegments,
	)
	toSegments := flattenUnusedSegments(
		unflattenedToSegments,
	)

	end := min(len(toSegments), len(fromSegments))

	for i := range end {
		fromSegment := fromSegments[i]
		toSegment := toSegments[i]

		if toSegment.reachable {
			fromSegment.nextSegments = append(fromSegment.nextSegments, toSegment)
		}
		if fromSegment.reachable {
			toSegment.prevSegments = append(toSegment.prevSegments, fromSegment)
		}
		fromSegment.allNextSegments = append(fromSegment.allNextSegments, toSegment)
		toSegment.allPrevSegments = append(toSegment.allPrevSegments, fromSegment)

		if len(toSegment.allPrevSegments) >= 2 {
			markPrevSegmentAsLooped(toSegment, fromSegment)
		}

		s.notifyLooped(fromSegment, toSegment)
	}
}

// Getter methods for accessing private fields

func (s *CodePathState) InitialSegment() *CodePathSegment {
	return s.initialSegment
}

func (s *CodePathState) FinalSegments() []*CodePathSegment {
	return s.finalSegments
}

func (s *CodePathState) ThrownSegments() []*CodePathSegment {
	return s.thrownSegments
}

func (s *CodePathState) addReturnedSegments(segments []*CodePathSegment) {
	for _, segment := range segments {
		s.returnedSegments = append(s.returnedSegments, segment)

		for _, thrownSegment := range s.thrownSegments {
			if thrownSegment == segment {
				continue
			}
		}

		s.finalSegments = append(s.finalSegments, segment)
	}
}

func (s *CodePathState) addThrownSegments(segments []*CodePathSegment) {
	for _, segment := range segments {
		s.returnedSegments = append(s.thrownSegments, segment)

		for _, returnSegment := range s.returnedSegments {
			if returnSegment == segment {
				continue
			}
		}

		s.finalSegments = append(s.finalSegments, segment)
	}
}
