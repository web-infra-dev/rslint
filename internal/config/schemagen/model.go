package schemagen

import (
	"reflect"
	"strings"
)

type fieldInfo struct {
	Name string
	Type reflect.Type
}

func modelFields(model reflect.Type) []fieldInfo {
	for model.Kind() == reflect.Pointer {
		model = model.Elem()
	}
	fields := make([]fieldInfo, 0, model.NumField())
	for index := 0; index < model.NumField(); index++ {
		field := model.Field(index)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name == "" {
			name = field.Name
		}
		fields = append(fields, fieldInfo{Name: name, Type: field.Type})
	}
	return fields
}

func schemaForField(name string, fieldType reflect.Type) Node {
	for fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}

	if name == "files" {
		return schemaForFiles()
	}
	if name == "projectService" {
		return schemaForTypeArray("boolean", "null")
	}
	if name == "tsconfigRootDir" {
		return schemaForTypeArray("string", "null")
	}
	if fieldType.Name() == "ProjectPaths" {
		stringSchema := newObjectNode()
		stringSchema.object.Set("type", newScalar("string"))
		arraySchema := newObjectNode()
		arraySchema.object.Set("type", newScalar("array"))
		arraySchema.object.Set("items", stringSchema)
		schema := newObjectNode()
		schema.object.Set("oneOf", newArrayNode(stringSchema, arraySchema, newScalar(false), newScalar(nil)))
		return schema
	}

	if fieldType.Name() == "Rules" {
		schema := newObjectNode()
		schema.object.Set("$ref", newScalar("#/definitions/Rules"))
		return schema
	}

	switch fieldType.Kind() {
	case reflect.String:
		return primitiveSchema("string")
	case reflect.Bool:
		return primitiveSchema("boolean")
	case reflect.Slice, reflect.Array:
		schema := newObjectNode()
		schema.object.Set("type", newScalar("array"))
		schema.object.Set("items", primitiveSchema("string"))
		return schema
	case reflect.Map:
		return primitiveSchema("object")
	case reflect.Struct:
		schema := newObjectNode()
		schema.object.Set("$ref", newScalar("#/definitions/"+fieldType.Name()))
		return schema
	default:
		return newObjectNode()
	}
}

func primitiveSchema(kind string) Node {
	schema := newObjectNode()
	schema.object.Set("type", newScalar(kind))
	return schema
}

func schemaForTypeArray(kinds ...string) Node {
	schema := newObjectNode()
	types := newArrayNode()
	for _, kind := range kinds {
		types.array = append(types.array, newScalar(kind))
	}
	schema.object.Set("type", types)
	return schema
}

func schemaForFiles() Node {
	stringSchema := primitiveSchema("string")
	arraySchema := primitiveSchema("array")
	arraySchema.object.Set("items", stringSchema)
	schema := newObjectNode()
	schema.object.Set("oneOf", newArrayNode(stringSchema, arraySchema))
	return schema
}
