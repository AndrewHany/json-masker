package masker

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
)

var removeIndexRegex = regexp.MustCompile(`\[\d+\]`)

type Masker interface {
	Mask(data string, maskPaths []string) (string, error)
	log(data string)
}

type masker struct {
	maskFunc    func(field any) string
	isDebugMode bool
}

type option func(*masker)

func WithMaskFunc(maskFunc func(field any) string) option {
	return func(m *masker) {
		m.maskFunc = maskFunc
	}
}

func WithFixedMaskString(maskStr string) option {
	return WithMaskFunc(func(field any) string {
		return maskStr
	})
}

func WithDebugMode() option {
	return func(m *masker) {
		m.isDebugMode = true
	}
}

func NewMasker(maskPaths []string, opts ...option) Masker {
	m := &masker{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Mask masks the input JSON string based on the provided maskPaths.
// maskPaths is a list of JSON paths that should be masked.
// maskStr is the string that will replace the masked values.
// The function returns the masked JSON string.
func (m *masker) Mask(input string, maskPaths []string) (string, error) {
	maskPathsMap := make(map[string]bool)
	for _, path := range maskPaths {
		maskPathsMap[path] = true
	}
	var inputValue interface{}
	if err := json.Unmarshal([]byte(input), &inputValue); err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %w", err)
	}
	maskedObject, err := m.maskWithPaths(reflect.ValueOf(inputValue), maskPathsMap, "$")
	if err != nil {
		return "", fmt.Errorf("failed to mask object: %w", err)
	}
	maskedBytes, err := json.Marshal(maskedObject)
	if err != nil {
		return "", fmt.Errorf("failed to marshal masked object: %w", err)
	}
	return string(maskedBytes), nil
}

// maskWithPaths recursively masks the input object based on the provided maskPaths.
// maskPaths is a map of JSON paths that should be masked.
// maskStr is the string that will replace the masked values.
// path is the current path of the object in the JSON.
// The function returns the masked object.
func (m *masker) maskWithPaths(
	input reflect.Value,
	maskPaths map[string]bool,
	path string,
) (any, error) {

	m.log(fmt.Sprintf("Processing path: %s", path))
	// Dereference pointers
	for input.Kind() == reflect.Ptr {
		input = input.Elem()
	}

	// handle nil pointers
	if !input.IsValid() {
		return reflect.ValueOf(nil), nil
	}

	// check if the path should be masked
	if isMaskedPath(path, maskPaths) {
		m.log(fmt.Sprintf("Masking path: %s", path))
		return m.maskFunc(input.Interface()), nil
	}

	switch input.Kind() {
	case reflect.Struct:
		for i := 0; i < input.NumField(); i++ {
			m.log(fmt.Sprintf("Processing field: %s", input.Type().Field(i).Name))
			field := input.Type().Field(i)
			fieldPath := path + "." + field.Name
			if maskedValue, err := m.maskWithPaths(input.Field(i), maskPaths, fieldPath); err != nil {
				return nil, err
			} else {
				input.Field(i).Set(reflect.ValueOf(maskedValue))
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < input.Len(); i++ {
			m.log(fmt.Sprintf("Processing index: %d", i))
			if maskedValue, err := m.maskWithPaths(input.Index(i), maskPaths, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return nil, err
			} else {
				input.Index(i).Set(reflect.ValueOf(maskedValue))
			}
		}
	case reflect.Map:
		for _, key := range input.MapKeys() {
			m.log(fmt.Sprintf("Processing key: %v", key.Interface()))
			if maskedValue, err := m.maskWithPaths(input.MapIndex(key), maskPaths, fmt.Sprintf("%s.%v", path, key.Interface())); err != nil {
				return nil, err
			} else {
				input.SetMapIndex(key, reflect.ValueOf(maskedValue))
			}
		}
	case reflect.Interface:
		m.log(fmt.Sprintf("Processing interface: %v", input.Interface()))
		if input.IsNil() {
			return nil, nil
		}
		if maskedValue, err := m.maskWithPaths(input.Elem(), maskPaths, path); err != nil {
			return nil, err
		} else if input.CanSet() {
			input.Set(reflect.ValueOf(maskedValue))
		}
	default:
		m.log(fmt.Sprintf("No action needed for: %v", input.Interface()))
		// do nothing
	}
	return input.Interface(), nil
}

func (m *masker) log(data string) {
	if m.isDebugMode {
		fmt.Println(data)
	}
}

// isMaskedPath checks if the path is in the maskPaths map.
// removeIndexRegex is used to remove array indexes from the path.
func isMaskedPath(path string, maskPaths map[string]bool) bool {
	_, ok := maskPaths[removeIndexRegex.ReplaceAllString(path, "[]")]
	return ok
}
