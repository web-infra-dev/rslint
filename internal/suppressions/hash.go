package suppressions

import (
	"math/bits"
	"strconv"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// ESLint names directory-based files with imurmurhash(cwd).result().toString(36).
// imurmurhash packs four UTF-16 code units at a time, including the upper
// byte of the fourth unit in the low byte of the word.
func cwdHash(cwd string) string {
	units := ecmascript.StringCodeUnits(cwd)
	mix := func(word uint32) uint32 {
		return bits.RotateLeft32(word*0xcc9e2d51, 15) * 0x1b873593
	}
	var hash uint32
	remaining := units
	for len(remaining) >= 4 {
		last := uint32(remaining[3])
		word := uint32(remaining[0]) ^ uint32(remaining[1])<<8 ^ uint32(remaining[2])<<16 ^ last<<24 ^ last>>8
		hash = bits.RotateLeft32(hash^mix(word), 13)*5 + 0xe6546b64
		remaining = remaining[4:]
	}
	var tail uint32
	for i, unit := range remaining {
		tail ^= uint32(unit) << (8 * i)
	}
	// The upstream tail accumulator is a signed JavaScript bitwise result.
	if int32(tail) > 0 {
		hash ^= mix(tail)
	}
	hash ^= uint32(len(units))
	hash ^= hash >> 16
	hash *= 0x85ebca6b
	hash ^= hash >> 13
	hash *= 0xc2b2ae35
	hash ^= hash >> 16
	return strconv.FormatUint(uint64(hash), 36)
}
