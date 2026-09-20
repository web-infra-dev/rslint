// Package jsonorder retains object property order alongside ordinary decoded
// JSON values. It uses hujson for parsing; values remain encoding/json maps and
// slices so schema validation and existing rule options keep their usual types.
package jsonorder

import (
	"encoding/json"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/tailscale/hujson"
)

// Order is immutable after construction. Children use property names for
// objects and decimal indices for arrays. Scalar children need no metadata.
type Order struct {
	Keys     []string
	Children map[string]*Order
}

func Parse(data []byte) (*Order, error) {
	value, err := hujson.Parse(data)
	if err != nil {
		return nil, err
	}
	return collect(value), nil
}

func collect(value hujson.Value) *Order {
	var order *Order
	switch value := value.Value.(type) {
	case *hujson.Object:
		order = &Order{Children: make(map[string]*Order, len(value.Members))}
		for _, member := range value.Members {
			literal, _ := member.Name.Value.(hujson.Literal)
			name := literal.String()
			if _, exists := order.Children[name]; !exists {
				order.Keys = append(order.Keys, name)
			}
			order.Children[name] = collect(member.Value)
		}
	case *hujson.Array:
		order = &Order{Children: map[string]*Order{}}
		for index, element := range value.Elements {
			if child := collect(element); child != nil {
				order.Children[strconv.Itoa(index)] = child
			}
		}
	default:
		return nil
	}
	return order
}

// At selects a nested value without constructing or escaping a JSON pointer.
func (order *Order) At(path ...string) *Order {
	for _, key := range path {
		if order == nil {
			return nil
		}
		order = order.Children[key]
	}
	return order
}

// PropertyKeys returns present keys in JavaScript enumeration order. Keys
// added after decoding (or from a Go map with no authored order) are sorted.
func (order *Order) PropertyKeys(object map[string]any) []string {
	keys := slices.Collect(maps.Keys(object))
	order.sort(keys)
	return keys
}

func (order *Order) sort(keys []string) {
	if len(keys) < 2 {
		return
	}
	positions := map[string]int{}
	if order != nil {
		for index, key := range order.Keys {
			positions[key] = index + 1
		}
	}
	slices.SortFunc(keys, func(a, b string) int {
		// ECMAScript enumerates array-index keys before other strings, even
		// when they were inserted later. 2^32-1 is not an array index.
		ai, aIndex := arrayIndex(a)
		bi, bIndex := arrayIndex(b)
		if aIndex || bIndex {
			if !aIndex || aIndex == bIndex && ai > bi {
				return 1
			}
			if !bIndex || ai < bi {
				return -1
			}
			return 0
		}
		ap, bp := positions[a], positions[b]
		if ap != bp {
			if ap == 0 || bp != 0 && ap > bp {
				return 1
			}
			return -1
		}
		return strings.Compare(a, b)
	})
}

func arrayIndex(key string) (uint64, bool) {
	if key == "" || key[0] < '0' || key[0] > '9' {
		return 0, false
	}
	index, err := strconv.ParseUint(key, 10, 32)
	return index, err == nil && index < 1<<32-1 && strconv.FormatUint(index, 10) == key
}

// Marshal preserves authored key order while encoding the current values,
// including schema defaults and subsequent configuration edits.
func Marshal(value any, order *Order) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil || order == nil {
		return encoded, err
	}
	tree, err := hujson.Parse(encoded)
	if err != nil {
		return nil, err
	}
	order.restore(&tree)
	return tree.Pack(), nil
}

func (order *Order) restore(value *hujson.Value) {
	if order == nil {
		return
	}
	switch object := value.Value.(type) {
	case *hujson.Object:
		members := make(map[string]hujson.ObjectMember, len(object.Members))
		keys := make([]string, 0, len(object.Members))
		for _, member := range object.Members {
			literal, _ := member.Name.Value.(hujson.Literal)
			key := literal.String()
			order.At(key).restore(&member.Value)
			members[key] = member
			keys = append(keys, key)
		}
		order.sort(keys)
		for index, key := range keys {
			object.Members[index] = members[key]
		}
	case *hujson.Array:
		for index := range object.Elements {
			order.At(strconv.Itoa(index)).restore(&object.Elements[index])
		}
	}
}
