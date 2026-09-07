package regexp

import "fmt"

// captureLayout maps JavaScript's left-to-right capture numbers to regexp2's
// unnamed-first numbering. It is immutable and shared by matching and string
// substitution. Patterns with no named groups allocate no mapping.
type captureLayout struct {
	count   int
	named   bool
	numbers []int
	names   map[string]int
}

func parseCaptureLayout(source string) (captureLayout, error) {
	count, named := countGroups(source)
	layout := captureLayout{count: count, named: named}
	if !named {
		return layout, nil
	}
	layout.names = make(map[string]int)
	index, firstNamed := 0, count+1
	var err error
	scanGroups(source, func(name string) {
		index++
		if name != "" {
			firstNamed = min(firstNamed, index)
			if _, exists := layout.names[name]; exists {
				// regexp2 merges duplicate names into one capture. JavaScript
				// permits some duplicates across alternatives but still gives
				// them separate numeric slots, which that engine cannot expose.
				err = fmt.Errorf("%w: duplicate capture name %q", ErrUnsupportedSyntax, name)
			}
			layout.names[name] = index
		}
	})
	unnamed := count - len(layout.names)
	if err != nil || firstNamed > unnamed {
		// All named groups already follow the unnamed groups (including a
		// pattern whose captures are all named), so the numbering is identical.
		return layout, err
	}
	layout.numbers = make([]int, count+1)
	for _, index := range layout.names {
		layout.numbers[index] = -1
	}
	nextUnnamed := 0
	for index := 1; index <= count; index++ {
		if layout.numbers[index] == -1 {
			unnamed++
			layout.numbers[index] = unnamed
		} else {
			nextUnnamed++
			layout.numbers[index] = nextUnnamed
		}
	}
	return layout, nil
}

func (layout captureLayout) number(index int) int {
	if layout.numbers == nil {
		return index
	}
	return layout.numbers[index]
}
