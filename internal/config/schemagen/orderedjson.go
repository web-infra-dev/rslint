package schemagen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type nodeKind uint8

const (
	scalarNode nodeKind = iota
	objectNode
	arrayNode
)

type Node struct {
	kind   nodeKind
	scalar any
	object *Object
	array  []Node
}

type Object struct {
	keys   []string
	values map[string]Node
}

func newObject() *Object {
	return &Object{values: make(map[string]Node)}
}

func newScalar(value any) Node {
	return Node{kind: scalarNode, scalar: value}
}

func newObjectNode() Node {
	return Node{kind: objectNode, object: newObject()}
}

func newArrayNode(values ...Node) Node {
	return Node{kind: arrayNode, array: values}
}

func (o *Object) Get(key string) (Node, bool) {
	value, ok := o.values[key]
	return value, ok
}

func (o *Object) Keys() []string {
	return append([]string(nil), o.keys...)
}

func (o *Object) Set(key string, value Node) {
	if _, exists := o.values[key]; !exists {
		o.keys = append(o.keys, key)
	}
	o.values[key] = value
}

func (o *Object) Delete(key string) {
	if _, exists := o.values[key]; !exists {
		return
	}
	delete(o.values, key)
	for index, existing := range o.keys {
		if existing == key {
			o.keys = append(o.keys[:index], o.keys[index+1:]...)
			break
		}
	}
}

func (n Node) AsObject() (*Object, bool) {
	if n.kind != objectNode {
		return nil, false
	}
	return n.object, true
}

func (n Node) AsArray() ([]Node, bool) {
	if n.kind != arrayNode {
		return nil, false
	}
	return n.array, true
}

func Parse(data []byte) (Node, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := parseValue(decoder)
	if err != nil {
		return Node{}, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return Node{}, fmt.Errorf("unexpected trailing JSON value")
		}
		return Node{}, err
	}
	return value, nil
}

func parseValue(decoder *json.Decoder) (Node, error) {
	token, err := decoder.Token()
	if err != nil {
		return Node{}, err
	}
	switch value := token.(type) {
	case json.Delim:
		switch value {
		case '{':
			return parseObject(decoder)
		case '[':
			return parseArray(decoder)
		default:
			return Node{}, fmt.Errorf("unexpected JSON delimiter %q", value)
		}
	case string, json.Number, bool, nil:
		return newScalar(value), nil
	default:
		return Node{}, fmt.Errorf("unsupported JSON token %T", token)
	}
}

func parseObject(decoder *json.Decoder) (Node, error) {
	object := newObject()
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return Node{}, err
		}
		key, ok := token.(string)
		if !ok {
			return Node{}, fmt.Errorf("object key is %T, want string", token)
		}
		value, err := parseValue(decoder)
		if err != nil {
			return Node{}, err
		}
		object.Set(key, value)
	}
	end, err := decoder.Token()
	if err != nil {
		return Node{}, err
	}
	if end != json.Delim('}') {
		return Node{}, fmt.Errorf("object ended with %v", end)
	}
	return Node{kind: objectNode, object: object}, nil
}

func parseArray(decoder *json.Decoder) (Node, error) {
	values := make([]Node, 0)
	for decoder.More() {
		value, err := parseValue(decoder)
		if err != nil {
			return Node{}, err
		}
		values = append(values, value)
	}
	end, err := decoder.Token()
	if err != nil {
		return Node{}, err
	}
	if end != json.Delim(']') {
		return Node{}, fmt.Errorf("array ended with %v", end)
	}
	return Node{kind: arrayNode, array: values}, nil
}

func (n Node) MarshalIndent() ([]byte, error) {
	var buffer bytes.Buffer
	if err := n.write(&buffer, 0); err != nil {
		return nil, err
	}
	buffer.WriteByte('\n')
	return buffer.Bytes(), nil
}

func (n Node) write(buffer *bytes.Buffer, depth int) error {
	switch n.kind {
	case scalarNode:
		encoded, err := json.Marshal(n.scalar)
		if err != nil {
			return err
		}
		buffer.Write(encoded)
		return nil
	case objectNode:
		buffer.WriteByte('{')
		if len(n.object.keys) == 0 {
			buffer.WriteByte('}')
			return nil
		}
		buffer.WriteByte('\n')
		for index, key := range n.object.keys {
			writeIndent(buffer, depth+1)
			encodedKey, err := json.Marshal(key)
			if err != nil {
				return err
			}
			buffer.Write(encodedKey)
			buffer.WriteString(": ")
			if err := n.object.values[key].write(buffer, depth+1); err != nil {
				return err
			}
			if index < len(n.object.keys)-1 {
				buffer.WriteByte(',')
			}
			buffer.WriteByte('\n')
		}
		writeIndent(buffer, depth)
		buffer.WriteByte('}')
		return nil
	case arrayNode:
		buffer.WriteByte('[')
		if len(n.array) == 0 {
			buffer.WriteByte(']')
			return nil
		}
		buffer.WriteByte('\n')
		for index, value := range n.array {
			writeIndent(buffer, depth+1)
			if err := value.write(buffer, depth+1); err != nil {
				return err
			}
			if index < len(n.array)-1 {
				buffer.WriteByte(',')
			}
			buffer.WriteByte('\n')
		}
		writeIndent(buffer, depth)
		buffer.WriteByte(']')
		return nil
	default:
		return fmt.Errorf("unknown JSON node kind %d", n.kind)
	}
}

func writeIndent(buffer *bytes.Buffer, depth int) {
	for index := 0; index < depth; index++ {
		buffer.WriteString("  ")
	}
}

func Canonical(data []byte) ([]byte, error) {
	value, err := Parse(data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(value.toAny())
}

func (n Node) clone() Node {
	switch n.kind {
	case scalarNode:
		return newScalar(n.scalar)
	case objectNode:
		clone := newObjectNode()
		for _, key := range n.object.keys {
			clone.object.Set(key, n.object.values[key].clone())
		}
		return clone
	case arrayNode:
		clone := newArrayNode()
		for _, value := range n.array {
			clone.array = append(clone.array, value.clone())
		}
		return clone
	default:
		return Node{}
	}
}

func (n Node) toAny() any {
	switch n.kind {
	case scalarNode:
		return n.scalar
	case objectNode:
		value := make(map[string]any, len(n.object.keys))
		for _, key := range n.object.keys {
			value[key] = n.object.values[key].toAny()
		}
		return value
	case arrayNode:
		value := make([]any, len(n.array))
		for index, item := range n.array {
			value[index] = item.toAny()
		}
		return value
	default:
		return nil
	}
}
