package order

// The sorting algorithm is derived from V8's TimSort implementation, itself
// based on CPython's list sort:
// Copyright 2018 the V8 project authors. All rights reserved.
// Use of the V8 source is governed by a BSD-style license:
// https://chromium.googlesource.com/v8/v8/+/main/LICENSE.v8
// Copyright (c) 2001-2018 Python Software Foundation; All Rights Reserved.
// https://docs.python.org/3/license.html
//
// eslint-plugin-import sorts alphabetized entries with Array.prototype.sort.
// Its path comparator can be non-transitive when parent and sibling imports
// share a group, so the comparator call order becomes observable. Node 22 and
// 24 both use V8's TimSort here. Keep this implementation rule-local: the
// ECMAScript specification does not promise a particular sorting algorithm.
//
// Algorithm reference:
// https://chromium.googlesource.com/v8/v8.git/+/refs/tags/13.6.233/third_party/v8/builtins/array-sort.tq

const v8MinGallopWins = 7

type v8SortRun struct {
	base   int
	length int
}

type v8AlphabetizeSort struct {
	data      []alphabetizeEntry
	compare   func(a, b alphabetizeEntry) int
	minGallop int
	runs      []v8SortRun
	temp      []alphabetizeEntry
}

func v8StableSortAlphabetized(data []alphabetizeEntry, compare func(a, b alphabetizeEntry) int) {
	if len(data) < 2 {
		return
	}
	state := v8AlphabetizeSort{
		data:      data,
		compare:   compare,
		minGallop: v8MinGallopWins,
	}
	minRun := v8MinRunLength(len(data))
	for low := 0; low < len(data); {
		runLength := state.countAndMakeRun(low, len(data))
		if runLength < minRun {
			forcedLength := min(minRun, len(data)-low)
			state.binaryInsertionSort(low, low+runLength, low+forcedLength)
			runLength = forcedLength
		}
		state.runs = append(state.runs, v8SortRun{base: low, length: runLength})
		state.mergeCollapse()
		low += runLength
	}
	state.mergeForceCollapse()
}

func v8MinRunLength(length int) int {
	remainder := 0
	for length >= 64 {
		remainder |= length & 1
		length >>= 1
	}
	return length + remainder
}

func (state *v8AlphabetizeSort) binaryInsertionSort(low, start, high int) {
	if start == low {
		start++
	}
	for ; start < high; start++ {
		pivot := state.data[start]
		left, right := low, start
		for left < right {
			middle := left + (right-left)/2
			if state.compare(pivot, state.data[middle]) < 0 {
				right = middle
			} else {
				left = middle + 1
			}
		}
		copy(state.data[left+1:start+1], state.data[left:start])
		state.data[left] = pivot
	}
}

func (state *v8AlphabetizeSort) countAndMakeRun(low, high int) int {
	if low+1 == high {
		return 1
	}
	runLength := 2
	descending := state.compare(state.data[low+1], state.data[low]) < 0
	for low+runLength < high {
		order := state.compare(state.data[low+runLength], state.data[low+runLength-1])
		if descending {
			if order >= 0 {
				break
			}
		} else if order < 0 {
			break
		}
		runLength++
	}
	if descending {
		for left, right := low, low+runLength-1; left < right; left, right = left+1, right-1 {
			state.data[left], state.data[right] = state.data[right], state.data[left]
		}
	}
	return runLength
}

func (state *v8AlphabetizeSort) mergeCollapse() {
	for len(state.runs) > 1 {
		index := len(state.runs) - 2
		if !state.runInvariantEstablished(index+1) || !state.runInvariantEstablished(index) {
			if state.runs[index-1].length < state.runs[index+1].length {
				index--
			}
			state.mergeAt(index)
		} else if state.runs[index].length <= state.runs[index+1].length {
			state.mergeAt(index)
		} else {
			return
		}
	}
}

func (state *v8AlphabetizeSort) runInvariantEstablished(index int) bool {
	if index < 2 {
		return true
	}
	return state.runs[index-2].length > state.runs[index-1].length+state.runs[index].length
}

func (state *v8AlphabetizeSort) mergeForceCollapse() {
	for len(state.runs) > 1 {
		index := len(state.runs) - 2
		if index > 0 && state.runs[index-1].length < state.runs[index+1].length {
			index--
		}
		state.mergeAt(index)
	}
}

func (state *v8AlphabetizeSort) mergeAt(index int) {
	left, right := state.runs[index], state.runs[index+1]
	state.runs[index].length = left.length + right.length
	copy(state.runs[index+1:], state.runs[index+2:])
	state.runs = state.runs[:len(state.runs)-1]

	skipLeft := state.gallopRight(state.data[right.base], state.data, left.base, left.length, 0)
	left.base += skipLeft
	left.length -= skipLeft
	if left.length == 0 {
		return
	}
	right.length = state.gallopLeft(
		state.data[left.base+left.length-1],
		state.data,
		right.base,
		right.length,
		right.length-1,
	)
	if right.length == 0 {
		return
	}
	if left.length <= right.length {
		state.mergeLow(left.base, left.length, right.base, right.length)
	} else {
		state.mergeHigh(left.base, left.length, right.base, right.length)
	}
}

func (state *v8AlphabetizeSort) gallopLeft(
	key alphabetizeEntry,
	array []alphabetizeEntry,
	base, length, hint int,
) int {
	lastOffset, offset := 0, 1
	if state.compare(array[base+hint], key) < 0 {
		maxOffset := length - hint
		for offset < maxOffset && state.compare(array[base+hint+offset], key) < 0 {
			lastOffset = offset
			offset = nextGallopOffset(offset, maxOffset)
		}
		offset = min(offset, maxOffset)
		lastOffset += hint
		offset += hint
	} else {
		maxOffset := hint + 1
		for offset < maxOffset && state.compare(array[base+hint-offset], key) >= 0 {
			lastOffset = offset
			offset = nextGallopOffset(offset, maxOffset)
		}
		offset = min(offset, maxOffset)
		lastOffset, offset = hint-offset, hint-lastOffset
	}

	lastOffset++
	for lastOffset < offset {
		middle := lastOffset + (offset-lastOffset)/2
		if state.compare(array[base+middle], key) < 0 {
			lastOffset = middle + 1
		} else {
			offset = middle
		}
	}
	return offset
}

func (state *v8AlphabetizeSort) gallopRight(
	key alphabetizeEntry,
	array []alphabetizeEntry,
	base, length, hint int,
) int {
	lastOffset, offset := 0, 1
	if state.compare(key, array[base+hint]) < 0 {
		maxOffset := hint + 1
		for offset < maxOffset && state.compare(key, array[base+hint-offset]) < 0 {
			lastOffset = offset
			offset = nextGallopOffset(offset, maxOffset)
		}
		offset = min(offset, maxOffset)
		lastOffset, offset = hint-offset, hint-lastOffset
	} else {
		maxOffset := length - hint
		for offset < maxOffset && state.compare(key, array[base+hint+offset]) >= 0 {
			lastOffset = offset
			offset = nextGallopOffset(offset, maxOffset)
		}
		offset = min(offset, maxOffset)
		lastOffset += hint
		offset += hint
	}

	lastOffset++
	for lastOffset < offset {
		middle := lastOffset + (offset-lastOffset)/2
		if state.compare(key, array[base+middle]) < 0 {
			offset = middle
		} else {
			lastOffset = middle + 1
		}
	}
	return offset
}

func nextGallopOffset(offset, maximum int) int {
	if offset > (maximum-1)/2 {
		return maximum
	}
	return offset*2 + 1
}

func (state *v8AlphabetizeSort) temporary(length int) []alphabetizeEntry {
	if cap(state.temp) < length {
		state.temp = make([]alphabetizeEntry, length)
	} else {
		state.temp = state.temp[:length]
	}
	return state.temp
}

func (state *v8AlphabetizeSort) mergeLow(baseA, lengthA, baseB, lengthB int) {
	temp := state.temporary(lengthA)
	copy(temp, state.data[baseA:baseA+lengthA])
	dest, cursorA, cursorB := baseA, 0, baseB

	state.data[dest] = state.data[cursorB]
	dest++
	cursorB++
	lengthB--
	if lengthB == 0 {
		copy(state.data[dest:dest+lengthA], temp[cursorA:cursorA+lengthA])
		return
	}
	if lengthA == 1 {
		state.finishMergeLowSingleA(temp, cursorA, cursorB, dest, lengthB)
		return
	}

	minGallop := state.minGallop
	for {
		winsA, winsB := 0, 0
		for {
			if state.compare(state.data[cursorB], temp[cursorA]) < 0 {
				state.data[dest] = state.data[cursorB]
				dest++
				cursorB++
				lengthB--
				winsB++
				winsA = 0
				if lengthB == 0 {
					copy(state.data[dest:dest+lengthA], temp[cursorA:cursorA+lengthA])
					return
				}
				if winsB >= minGallop {
					break
				}
			} else {
				state.data[dest] = temp[cursorA]
				dest++
				cursorA++
				lengthA--
				winsA++
				winsB = 0
				if lengthA == 1 {
					state.finishMergeLowSingleA(temp, cursorA, cursorB, dest, lengthB)
					return
				}
				if winsA >= minGallop {
					break
				}
			}
		}

		minGallop++
		for first := true; first || winsA >= v8MinGallopWins || winsB >= v8MinGallopWins; first = false {
			minGallop = max(1, minGallop-1)
			state.minGallop = minGallop

			winsA = state.gallopRight(state.data[cursorB], temp, cursorA, lengthA, 0)
			if winsA > 0 {
				copy(state.data[dest:dest+winsA], temp[cursorA:cursorA+winsA])
				dest += winsA
				cursorA += winsA
				lengthA -= winsA
				if lengthA == 1 {
					state.finishMergeLowSingleA(temp, cursorA, cursorB, dest, lengthB)
					return
				}
				if lengthA == 0 {
					return
				}
			}
			state.data[dest] = state.data[cursorB]
			dest++
			cursorB++
			lengthB--
			if lengthB == 0 {
				copy(state.data[dest:dest+lengthA], temp[cursorA:cursorA+lengthA])
				return
			}

			winsB = state.gallopLeft(temp[cursorA], state.data, cursorB, lengthB, 0)
			if winsB > 0 {
				copy(state.data[dest:dest+winsB], state.data[cursorB:cursorB+winsB])
				dest += winsB
				cursorB += winsB
				lengthB -= winsB
				if lengthB == 0 {
					copy(state.data[dest:dest+lengthA], temp[cursorA:cursorA+lengthA])
					return
				}
			}
			state.data[dest] = temp[cursorA]
			dest++
			cursorA++
			lengthA--
			if lengthA == 1 {
				state.finishMergeLowSingleA(temp, cursorA, cursorB, dest, lengthB)
				return
			}
		}
		minGallop++
		state.minGallop = minGallop
	}
}

func (state *v8AlphabetizeSort) finishMergeLowSingleA(
	temp []alphabetizeEntry,
	cursorA, cursorB, dest, lengthB int,
) {
	copy(state.data[dest:dest+lengthB], state.data[cursorB:cursorB+lengthB])
	state.data[dest+lengthB] = temp[cursorA]
}

func (state *v8AlphabetizeSort) mergeHigh(baseA, lengthA, baseB, lengthB int) {
	temp := state.temporary(lengthB)
	copy(temp, state.data[baseB:baseB+lengthB])
	dest := baseB + lengthB - 1
	cursorA, cursorB := baseA+lengthA-1, lengthB-1

	state.data[dest] = state.data[cursorA]
	dest--
	cursorA--
	lengthA--
	if lengthA == 0 {
		copy(state.data[dest-(lengthB-1):dest+1], temp[:lengthB])
		return
	}
	if lengthB == 1 {
		state.finishMergeHighSingleB(temp, cursorB, cursorA, dest, lengthA)
		return
	}

	minGallop := state.minGallop
	for {
		winsA, winsB := 0, 0
		for {
			if state.compare(temp[cursorB], state.data[cursorA]) < 0 {
				state.data[dest] = state.data[cursorA]
				dest--
				cursorA--
				lengthA--
				winsA++
				winsB = 0
				if lengthA == 0 {
					copy(state.data[dest-(lengthB-1):dest+1], temp[:lengthB])
					return
				}
				if winsA >= minGallop {
					break
				}
			} else {
				state.data[dest] = temp[cursorB]
				dest--
				cursorB--
				lengthB--
				winsB++
				winsA = 0
				if lengthB == 1 {
					state.finishMergeHighSingleB(temp, cursorB, cursorA, dest, lengthA)
					return
				}
				if winsB >= minGallop {
					break
				}
			}
		}

		minGallop++
		for first := true; first || winsA >= v8MinGallopWins || winsB >= v8MinGallopWins; first = false {
			minGallop = max(1, minGallop-1)
			state.minGallop = minGallop

			offset := state.gallopRight(temp[cursorB], state.data, baseA, lengthA, lengthA-1)
			winsA = lengthA - offset
			if winsA > 0 {
				dest -= winsA
				cursorA -= winsA
				copy(state.data[dest+1:dest+1+winsA], state.data[cursorA+1:cursorA+1+winsA])
				lengthA -= winsA
				if lengthA == 0 {
					copy(state.data[dest-(lengthB-1):dest+1], temp[:lengthB])
					return
				}
			}
			state.data[dest] = temp[cursorB]
			dest--
			cursorB--
			lengthB--
			if lengthB == 1 {
				state.finishMergeHighSingleB(temp, cursorB, cursorA, dest, lengthA)
				return
			}

			offset = state.gallopLeft(state.data[cursorA], temp, 0, lengthB, lengthB-1)
			winsB = lengthB - offset
			if winsB > 0 {
				dest -= winsB
				cursorB -= winsB
				copy(state.data[dest+1:dest+1+winsB], temp[cursorB+1:cursorB+1+winsB])
				lengthB -= winsB
				if lengthB == 1 {
					state.finishMergeHighSingleB(temp, cursorB, cursorA, dest, lengthA)
					return
				}
				if lengthB == 0 {
					return
				}
			}
			state.data[dest] = state.data[cursorA]
			dest--
			cursorA--
			lengthA--
			if lengthA == 0 {
				copy(state.data[dest-(lengthB-1):dest+1], temp[:lengthB])
				return
			}
		}
		minGallop++
		state.minGallop = minGallop
	}
}

func (state *v8AlphabetizeSort) finishMergeHighSingleB(
	temp []alphabetizeEntry,
	cursorB, cursorA, dest, lengthA int,
) {
	dest -= lengthA
	cursorA -= lengthA
	copy(state.data[dest+1:dest+1+lengthA], state.data[cursorA+1:cursorA+1+lengthA])
	state.data[dest] = temp[cursorB]
}
