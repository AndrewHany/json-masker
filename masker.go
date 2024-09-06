package masker

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
)

var removeIndexRegex = regexp.MustCompile(`\[\d+\]`)

func Mask(input string, maskPaths []string, maskStr string) (string, error) {
	maskPathsMap := make(map[string]bool)
	for _, path := range maskPaths {
		maskPathsMap[path] = true
	}
	var inputValue interface{}
	if err := json.Unmarshal([]byte(input), &inputValue); err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %w", err)
	}
	maskedObject, err := maskWithPaths(reflect.ValueOf(inputValue), maskPathsMap, maskStr, "$")
	if err != nil {
		return "", fmt.Errorf("failed to mask object: %w", err)
	}
	maskedBytes, err := json.Marshal(maskedObject)
	if err != nil {
		return "", fmt.Errorf("failed to marshal masked object: %w", err)
	}
	return string(maskedBytes), nil
}

func maskWithPaths(
	input reflect.Value,
	maskPaths map[string]bool,
	maskStr string,
	path string,
) (any, error) {
	// Dereference pointers
	for input.Kind() == reflect.Ptr {
		input = input.Elem()
	}

	// handle nil pointers
	if !input.IsValid() {
		return reflect.ValueOf(nil), nil
	}

	// check if custom assertion is defined for the path
	if isMaskedPath(path, maskPaths) {
		return maskStr, nil
	}
	// print input kind and value
	fmt.Printf("Path:%s, kind: %v, value: %v\n", path, input.Kind(), input.Interface())

	switch input.Kind() {
	case reflect.Struct:
		for i := 0; i < input.NumField(); i++ {
			field := input.Type().Field(i)
			fieldPath := path + "." + field.Name
			if maskedValue, err := maskWithPaths(input.Field(i), maskPaths, maskStr, fieldPath); err != nil {
				return nil, err
			} else {
				input.Field(i).Set(reflect.ValueOf(maskedValue))
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < input.Len(); i++ {
			if maskedValue, err := maskWithPaths(input.Index(i), maskPaths, maskStr, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return nil, err
			} else {
				input.Index(i).Set(reflect.ValueOf(maskedValue))
			}
		}
	case reflect.Map:
		for _, key := range input.MapKeys() {
			if maskedValue, err := maskWithPaths(input.MapIndex(key), maskPaths, maskStr, fmt.Sprintf("%s.%v", path, key.Interface())); err != nil {
				return nil, err
			} else {
				input.SetMapIndex(key, reflect.ValueOf(maskedValue))
			}
		}
	case reflect.Interface:
		if input.IsNil() {
			return nil, nil
		}
		if maskedValue, err := maskWithPaths(input.Elem(), maskPaths, maskStr, path); err != nil {
			return nil, err
		} else if input.CanSet() {
			input.Set(reflect.ValueOf(maskedValue))
		}
	default:
		// do nothing
	}
	return input.Interface(), nil
}

func isMaskedPath(path string, maskPaths map[string]bool) bool {
	_, ok := maskPaths[removeIndexRegex.ReplaceAllString(path, "[]")]
	return ok
}
