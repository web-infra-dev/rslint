// Package typeservice exposes a closeable, caller-owned view of one bound
// Program's checker to JavaScript. It does not select projects or execute rules.
package typeservice

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// EncodePacket produces independently owned, versioned binary facts. Objects
// contain only values and handles, never Go pointers. The injected transport
// owns these bytes until its consumer has copied or decoded them.
func EncodePacket(value any) ([]byte, error) {
	data := []byte{'R', 'T', 'S', 1}
	var encode func(reflect.Value, int) error
	word := func(n uint32) { data = binary.LittleEndian.AppendUint32(data, n) }
	encode = func(v reflect.Value, depth int) error {
		if depth > 128 {
			return errors.New("type packet nesting exceeds 128")
		}
		if !v.IsValid() || ((v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer) && v.IsNil()) {
			data = append(data, 0)
			return nil
		}
		if v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
			return encode(v.Elem(), depth)
		}
		switch v.Kind() {
		case reflect.Bool:
			if v.Bool() {
				data = append(data, 2)
			} else {
				data = append(data, 1)
			}
		case reflect.String:
			s := ecmascript.StringCodeUnits(v.String())
			if uint64(len(s)) > math.MaxUint32 {
				return errors.New("type packet string too large")
			}
			data = append(data, 3)
			word(uint32(len(s)))
			for _, unit := range s {
				data = binary.LittleEndian.AppendUint16(data, unit)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
			n := float64(0)
			if v.Kind() >= reflect.Uint && v.Kind() <= reflect.Uint64 {
				n = float64(v.Uint())
			} else if v.Kind() >= reflect.Int && v.Kind() <= reflect.Int64 {
				n = float64(v.Int())
			} else {
				n = v.Float()
			}
			data = append(data, 4)
			data = binary.LittleEndian.AppendUint64(data, math.Float64bits(n))
		case reflect.Slice, reflect.Array:
			if v.Type() == reflect.TypeFor[[]byte]() {
				data = append(data, 7)
				word(uint32(v.Len()))
				data = append(data, v.Bytes()...)
				return nil
			}
			data = append(data, 5)
			word(uint32(v.Len()))
			for i := range v.Len() {
				if err := encode(v.Index(i), depth+1); err != nil {
					return err
				}
			}
		case reflect.Map:
			if v.Type().Key().Kind() != reflect.String {
				return errors.New("type packet requires string keys")
			}
			keys := v.MapKeys()
			sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
			data = append(data, 6)
			word(uint32(len(keys)))
			for _, key := range keys {
				if err := encode(key, depth+1); err != nil {
					return err
				}
				if err := encode(v.MapIndex(key), depth+1); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unsupported type packet value %v", v.Type())
		}
		return nil
	}
	if err := encode(reflect.ValueOf(value), 0); err != nil {
		return nil, err
	}
	return data, nil
}
