package order

import (
	"reflect"
	"slices"
	"testing"
)

func TestReverseOutOfOrderMatchesNegatedCopy(t *testing.T) {
	t.Parallel()

	for length := 1; length <= 7; length++ {
		ranks := make([]float64, length)
		var visit func(int)
		visit = func(index int) {
			if index != length {
				for rank := range 3 {
					ranks[index] = float64(rank)
					visit(index + 1)
				}
				return
			}

			imports := make([]*importEntry, length)
			positions := make(map[*importEntry]int, length)
			for position, rank := range ranks {
				imports[position] = &importEntry{rank: rank}
				positions[imports[position]] = position
			}

			want := referenceReversePairs(ranks)
			count := countReverseOutOfOrder(imports)
			out := findReverseOutOfOrder(imports, count)
			got := make([][2]int, 0, len(out))
			for _, entry := range out {
				for position := len(imports) - 1; position >= 0; position-- {
					if imports[position].rank < entry.rank {
						got = append(got, [2]int{position, positions[entry]})
						break
					}
				}
			}

			if count != len(want) || !reflect.DeepEqual(got, want) {
				t.Fatalf("ranks %v: count/pairs = %d/%v, want %d/%v", ranks, count, got, len(want), want)
			}
		}
		visit(0)
	}
}

func referenceReversePairs(ranks []float64) [][2]int {
	type entry struct {
		position int
		rank     float64
	}

	reversed := make([]entry, len(ranks))
	for position, rank := range ranks {
		reversed[len(ranks)-1-position] = entry{position: position, rank: -rank}
	}
	maxSeen := reversed[0].rank
	var out []entry
	for _, current := range reversed {
		if current.rank < maxSeen {
			out = append(out, current)
		}
		if current.rank > maxSeen {
			maxSeen = current.rank
		}
	}
	slices.SortFunc(out, func(left, right entry) int {
		return left.position - right.position
	})

	pairs := make([][2]int, 0, len(out))
	for _, current := range out {
		for _, candidate := range reversed {
			if candidate.rank > current.rank {
				pairs = append(pairs, [2]int{candidate.position, current.position})
				break
			}
		}
	}
	return pairs
}
