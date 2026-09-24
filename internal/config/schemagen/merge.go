package schemagen

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/web-infra-dev/rslint/internal/config"
)

type MergeResult struct {
	Added   []string
	Removed []string
	Updated []string
}

type modelSpec struct {
	Definition string
	Model      reflect.Type
	Preserve   map[string]bool
}

func Merge(existing []byte) ([]byte, MergeResult, error) {
	root, err := Parse(existing)
	if err != nil {
		return nil, MergeResult{}, fmt.Errorf("parse schema: %w", err)
	}
	rootObject, ok := root.AsObject()
	if !ok {
		return nil, MergeResult{}, fmt.Errorf("schema root must be an object")
	}

	result := MergeResult{}
	for _, spec := range modelSpecs() {
		properties, err := objectAt(
			rootObject,
			"definitions",
			spec.Definition,
			"properties",
		)
		if err != nil {
			return nil, MergeResult{}, err
		}

		expected := make(map[string]bool)
		for _, field := range modelFields(spec.Model) {
			expected[field.Name] = true
			desired := schemaForField(field.Name, field.Type)
			if existing, exists := properties.Get(field.Name); exists {
				if reconcileFieldSchema(field.Name, existing, desired) {
					properties.Set(field.Name, existing)
					result.Updated = append(
						result.Updated,
						fmt.Sprintf("/definitions/%s/properties/%s", spec.Definition, field.Name),
					)
				}
				continue
			}
			properties.Set(field.Name, desired)
			result.Added = append(
				result.Added,
				fmt.Sprintf("/definitions/%s/properties/%s", spec.Definition, field.Name),
			)
		}

		for _, name := range properties.Keys() {
			if expected[name] || spec.Preserve[name] {
				continue
			}
			properties.Delete(name)
			result.Removed = append(
				result.Removed,
				fmt.Sprintf("/definitions/%s/properties/%s", spec.Definition, name),
			)
		}
	}

	if err := updatePluginEnum(rootObject, &result); err != nil {
		return nil, MergeResult{}, err
	}

	generated, err := root.MarshalIndent()
	if err != nil {
		return nil, MergeResult{}, fmt.Errorf("encode schema: %w", err)
	}
	return generated, result, nil
}

func reconcileFieldSchema(name string, existing, desired Node) bool {
	if name == "files" || name == "project" {
		return false
	}
	existingObject, existingOK := existing.AsObject()
	desiredObject, desiredOK := desired.AsObject()
	if !existingOK || !desiredOK {
		return false
	}
	changed := false
	for _, key := range []string{"type", "$ref"} {
		desiredValue, exists := desiredObject.Get(key)
		if !exists {
			continue
		}
		currentValue, currentExists := existingObject.Get(key)
		if !currentExists || !reflect.DeepEqual(currentValue.toAny(), desiredValue.toAny()) {
			existingObject.Set(key, desiredValue.clone())
			changed = true
		}
	}
	desiredItems, desiredItemsExists := desiredObject.Get("items")
	currentItems, currentItemsExists := existingObject.Get("items")
	if desiredItemsExists && currentItemsExists {
		desiredItemsObject, desiredItemsOK := desiredItems.AsObject()
		currentItemsObject, currentItemsOK := currentItems.AsObject()
		if desiredItemsOK && currentItemsOK && reconcileObjectStructure(currentItemsObject, desiredItemsObject) {
			changed = true
		}
	}
	return changed
}

func reconcileObjectStructure(existing, desired *Object) bool {
	changed := false
	for _, key := range []string{"type", "$ref"} {
		desiredValue, exists := desired.Get(key)
		if !exists {
			continue
		}
		currentValue, currentExists := existing.Get(key)
		if !currentExists || !reflect.DeepEqual(currentValue.toAny(), desiredValue.toAny()) {
			existing.Set(key, desiredValue.clone())
			changed = true
		}
	}
	return changed
}

func modelSpecs() []modelSpec {
	return []modelSpec{
		{
			Definition: "ConfigEntry",
			Model:      reflect.TypeOf(config.ConfigEntry{}),
		},
		{
			Definition: "LanguageOptions",
			Model:      reflect.TypeOf(config.LanguageOptions{}),
			Preserve: map[string]bool{
				"ecmaVersion":  true,
				"sourceType":   true,
				"globals":      true,
				"ecmaFeatures": true,
			},
		},
		{
			Definition: "ParserOptions",
			Model:      reflect.TypeOf(config.ParserOptions{}),
			Preserve: map[string]bool{
				"ecmaFeatures": true,
			},
		},
	}
}

func updatePluginEnum(root *Object, result *MergeResult) error {
	items, err := objectAt(
		root,
		"definitions",
		"ConfigEntry",
		"properties",
		"plugins",
		"items",
	)
	if err != nil {
		return err
	}
	desired := newArrayNode()
	for _, name := range config.BundledPluginDeclarationNames() {
		desired.array = append(desired.array, newScalar(name))
	}
	current, exists := items.Get("enum")
	if exists && reflect.DeepEqual(current.toAny(), desired.toAny()) {
		return nil
	}
	items.Set("enum", desired)
	result.Updated = append(result.Updated, "/definitions/ConfigEntry/properties/plugins/items/enum")
	return nil
}

func objectAt(root *Object, path ...string) (*Object, error) {
	current := root
	for index, key := range path {
		node, exists := current.Get(key)
		if !exists {
			return nil, fmt.Errorf("schema path /%s is missing", joinPath(path[:index+1]))
		}
		object, ok := node.AsObject()
		if !ok {
			return nil, fmt.Errorf("schema path /%s must be an object", joinPath(path[:index+1]))
		}
		current = object
	}
	return current, nil
}

func joinPath(path []string) string {
	result := ""
	for index, part := range path {
		if index > 0 {
			result += "/"
		}
		result += part
	}
	return result
}

func CanonicalEqual(left, right []byte) (bool, error) {
	leftCanonical, err := Canonical(left)
	if err != nil {
		return false, err
	}
	rightCanonical, err := Canonical(right)
	if err != nil {
		return false, err
	}
	return bytes.Equal(leftCanonical, rightCanonical), nil
}
