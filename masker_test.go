package masker

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsMaskedPath(t *testing.T) {

	testTable := []struct {
		name      string
		path      string
		maskPaths map[string]bool
		expected  bool
	}{
		{
			name: "mask by path",
			path: "someField.subField",
			maskPaths: map[string]bool{
				"someField.subField": true,
			},
			expected: true,
		},
		{
			name: "mask by path with index",
			path: "someField[2].subField",
			maskPaths: map[string]bool{
				"someField[].subField": true,
			},
			expected: true,
		},
		{
			name: "not matching",
			path: "someField.subField",
			maskPaths: map[string]bool{
				"test": true,
			},
			expected: false,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ok := isMaskedPath(tt.path, tt.maskPaths)
			assert.Equal(t, tt.expected, ok)
		})
	}
}

func TestMask_genericFields(t *testing.T) {
	testTime, _ := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	objectToJson := func(obj interface{}) string {
		bytes, _ := json.Marshal(obj)
		return string(bytes)
	}
	type subStruct struct {
		Field1 string
		Field2 []int
		Field3 map[string]int
	}
	type testStruct struct {
		Field1 string
		Field2 int
		Field3 []string
		Field4 map[string]string
		Field5 subStruct
	}

	testTable := []struct {
		name        string
		input       string
		maskPaths   []string
		expected    string
		expectedErr error
	}{
		{
			name:        "Test with invalid json",
			input:       "invalid",
			maskPaths:   []string{},
			expected:    "",
			expectedErr: fmt.Errorf("failed to unmarshal input: invalid character 'i' looking for beginning of value"),
		},
		{
			name:      "Test with mask on path",
			input:     objectToJson(struct{ Time time.Time }{Time: time.Now()}),
			maskPaths: []string{"$.Time"},
			expected:  `{"Time":"[REDACTED]"}`,
		},
		{
			name:      "Test with mask on path not matching",
			input:     objectToJson(struct{ TestValue int }{TestValue: 1}),
			maskPaths: []string{"$.TestValueTest"},
			expected:  `{"TestValue":1}`,
		},
		{
			name:     "Test with time and no mask",
			input:    objectToJson(struct{ Time time.Time }{Time: testTime}),
			expected: `{"Time":"2021-01-01T00:00:00Z"}`,
		},
		{
			name:  "Test with nested struct",
			input: objectToJson(testStruct{Field1: "test", Field2: 1, Field3: []string{"a", "b"}, Field4: map[string]string{"a": "1", "b": "2"}, Field5: subStruct{Field1: "sub", Field2: []int{1, 2}, Field3: map[string]int{"a": 1, "b": 2}}}),
			maskPaths: []string{
				"$.Field5.Field3.b",
				"$.Field5.Field2[]",
			},
			expected: `{"Field1":"test","Field2":1,"Field3":["a","b"],"Field4":{"a":"1","b":"2"},"Field5":{"Field1":"sub","Field2":["[REDACTED]","[REDACTED]"],"Field3":{"a":1,"b":"[REDACTED]"}}}`,
		},
		{
			name:      "Test with slices",
			input:     objectToJson([]int{1, 2, 3}),
			maskPaths: []string{},
			expected:  "[1,2,3]",
		},
		{
			name:      "Test with slices and mask",
			input:     objectToJson([]int{1, 2, 3}),
			maskPaths: []string{"$[]"},
			expected:  "[\"[REDACTED]\",\"[REDACTED]\",\"[REDACTED]\"]",
		},
		{
			name: "Test with map",
			input: objectToJson(map[string]int{
				"a": 1,
				"b": 2,
			}),
			maskPaths: []string{"$.b"},
			expected:  `{"a":1,"b":"[REDACTED]"}`,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			masker := NewMasker(tt.maskPaths, withFixedMaskString("[REDACTED]"), withDebugMode())
			output, err := masker.Mask(tt.input, tt.maskPaths)
			assert.Equal(t, tt.expected, output)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			}
		})
	}
}
