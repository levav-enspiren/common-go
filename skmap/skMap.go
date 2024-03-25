/*
String-key-map Package

It is to help managing unstructured map, esp. unmarshalled from json / yaml library.
It support dot-notation string as patht to access content of the map.

It is recommended to use "encoding/json" and "github.com/goccy/go-yaml", which
would return map[string]any for unmarshal. The data type for map[any]any is not supported yet.

See main type "Map" for the usage
*/
package skmap

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Hash = map[string]any

/*
String-key-map main type

Usage:

	// Unmarshal json / yaml
	// Requirement: The json root must be object
	var myMap skMap.Map
	json.Unmarshal(jsonBytes, &myMap)
	// Convert from map
	var m map[string]any
	myMap := skMap.Map(m)
	// Get value from map
	anyValue, err := myMap.Get("path")
	stringValue, err := myMap.GetString("path.for.the.value")
	stringValue2 := myMap.GetStringDefault("path.for.another.value", "default value")
	// Get value with array
	// Requirement: array must be []any. Other type of array is not supported
	// Requirement: the array name must not contain square branket
	stringValue, err := myMap.GetString("arr[1].value")
	arrValue, err := myMap.GetArray("path.for.the.arr")
	strArrValue, err := myMap.GetStringArray("path.for.the.strArr")
	// Set value from map
	// Requirement: the parent map has to be already in the map ("path.for.the")
	myMap.Set("path.for.the.newField", "new value to create or replace it")
	// the following would replace "path.for" with a new structure
	myMap.Set("path.for", skMap.Map{
		"newParent": skMap.Map {
			"newField": "new value",
		},
	})
	// Requirement: array index must not excess array length
	myMap.Set("arr[0]", "new value to replace it")
	// replace the whole array to change array length
	myMap.Set("arr", newArray)
	// Helper for casting
	myMap, err := skMap.CastToMap(anyPtr)
	m := myMap.ToMap()
	// Advance path resolving
	fields, lastField := skMap.ParsePath("path.for.the.value")
	ptr, val, err := myMap.ResolvePath(fields, lastField)
*/
type Map Hash

var ErrInvalidPath = errors.New("invalid path")
var ErrInvalidField = errors.New("invalid field")
var ErrInvalidType = errors.New("invalid type")

func ParsePath(path string) (fields []string, lastField string) {
	fields = strings.Split(path, ".")
	lastField = fields[len(fields)-1]
	fields = fields[:len(fields)-1]
	return
}

func (m Map) ToMap() Hash {
	return Hash(m)
}

func parseField(field string) (isArray bool, field2 string, arrayIndex int) {
	indexedArrayRe := regexp.MustCompile(`^([^[]+)\[(\d+)\]$`)
	matches := indexedArrayRe.FindSubmatch([]byte(field))
	if len(matches) == 3 {
		// parse for array
		field2 = string(matches[1])
		var err error
		arrayIndex, err = strconv.Atoi(string(matches[2]))
		if err == nil && arrayIndex >= 0 {
			isArray = true
		}
	}
	return
}

func resolveField(ptr Map, field string) (val any, err error) {
	isArray, field2, arrayIndex := parseField(field)
	if !isArray {
		field2 = field
	}
	val, ok := ptr[field2]
	if !ok {
		err = ErrInvalidField
		return
	}
	if isArray {
		var arr []any
		arr, err = CastToArrayAndVerify(val, field2, arrayIndex)
		if err != nil {
			return
		}
		val = arr[arrayIndex]
	}
	return
}

func CastToMap(val any) (umap Map, err error) {
	switch val := val.(type) {
	case Map:
		umap = val
	case Hash:
		umap = Map(val)
	default:
		err = ErrInvalidPath
	}
	return
}

func CastToArrayAndVerify(val any, field string, arrayIndex int) (arr []any, err error) {
	if val == nil {
		err = fmt.Errorf("%s: the field %s is empty", ErrInvalidPath, field)
		return
	}
	arr, ok := val.([]any)
	if !ok {
		err = fmt.Errorf("%s: the field %s is not an []any", ErrInvalidPath, field)
		return
	}
	if len(arr) <= arrayIndex {
		err = fmt.Errorf("%s: the index %d excess size of %s", ErrInvalidPath, arrayIndex, field)
		return
	}
	return
}

func (m Map) ResolvePath(fields []string, lastField string) (ptr Map, val any, err error) {
	if lastField == "" {
		val = m
		return
	}
	ptr = m
	var v any
	for i, field := range fields {
		v, err = resolveField(ptr, field)
		if err != nil {
			return
		}
		ptr, err = CastToMap(v)
		if err != nil {
			path := strings.Join(fields[:i+1], ".")
			err = fmt.Errorf("%s: path %s is not a map", err, path)
			return
		}
	}
	// retrive last field
	val, err = resolveField(ptr, lastField)
	return
}

func (m Map) Get(path string) (val any, err error) {
	fields, lastField := ParsePath(path)
	_, val, err = m.ResolvePath(fields, lastField)
	return
}

func (m Map) GetDefault(path string, defaultVal any) (val any) {
	val, err := m.Get(path)
	if err != nil || val == nil {
		val = defaultVal
	}
	return
}

func (m Map) GetString(path string) (str string, err error) {
	val, err := m.Get(path)
	if err != nil {
		return
	}
	str, ok := val.(string)
	if !ok {
		err = ErrInvalidType
	}
	return
}

func (m Map) GetStringDefault(path string, defaultStr string) (str string) {
	str, _ = m.GetString(path)
	if str == "" {
		str = defaultStr
	}
	return
}

func (m Map) GetInt(path string) (intVal int, err error) {
	val, err := m.Get(path)
	if err != nil {
		return
	}
	switch val := val.(type) {
	case int:
		intVal = val
	case float64:
		// float64 is default number type from json.Unmarshal
		intVal = int(val)
	case float32:
		intVal = int(val)
	case int64:
		intVal = int(val)
	case int32:
		intVal = int(val)
	case int16:
		intVal = int(val)
	case int8:
		intVal = int(val)
	case uint:
		intVal = int(val)
	case uint64:
		intVal = int(val)
	case uint32:
		intVal = int(val)
	case uint16:
		intVal = int(val)
	case uint8:
		intVal = int(val)
	default:
		err = ErrInvalidType
	}
	return
}

func (m Map) GetIntDefault(path string, defaultNum int) (num int) {
	num, err := m.GetInt(path)
	if err != nil {
		num = defaultNum
	}
	return
}

func (m Map) GetFloat64(path string) (float64Val float64, err error) {
	val, err := m.Get(path)
	if err != nil {
		return
	}
	switch val := val.(type) {
	case float64:
		float64Val = val
	case float32:
		float64Val = float64(val)
	case int:
		float64Val = float64(val)
	case int64:
		float64Val = float64(val)
	case int32:
		float64Val = float64(val)
	case int16:
		float64Val = float64(val)
	case int8:
		float64Val = float64(val)
	case uint:
		float64Val = float64(val)
	case uint64:
		float64Val = float64(val)
	case uint32:
		float64Val = float64(val)
	case uint16:
		float64Val = float64(val)
	case uint8:
		float64Val = float64(val)
	default:
		err = ErrInvalidType
	}
	return
}

func (m Map) GetFloat64Default(path string, defaultNum float64) (num float64) {
	num, err := m.GetFloat64(path)
	if err != nil {
		num = defaultNum
	}
	return
}

func (m Map) GetBool(path string) (b bool, err error) {
	val, err := m.Get(path)
	if err != nil {
		return
	}
	b, ok := val.(bool)
	if !ok {
		err = ErrInvalidType
	}
	return
}

func (m Map) GetBoolDefault(path string, defaultBool bool) (b bool) {
	b, err := m.GetBool(path)
	if err != nil {
		b = defaultBool
	}
	return
}

func (m Map) GetMap(path string) (umap Map, err error) {
	val, err := m.Get(path)
	if err != nil {
		return
	}
	return CastToMap(val)
}

func (m Map) GetMapDefault(path string, defaultMap Map) (umap Map) {
	umap, _ = m.GetMap(path)
	if umap == nil {
		umap = defaultMap
	}
	return
}

func (m Map) GetArray(path string) (arr []any, err error) {
	val, err := m.Get(path)
	if err != nil {
		return
	}
	arr, ok := val.([]any)
	if !ok {
		err = ErrInvalidType
		return
	}
	return
}

func (m Map) GetArrayDefault(path string, defaultArr []any) (arr []any) {
	arr, err := m.GetArray(path)
	if err != nil {
		arr = defaultArr
	}
	return
}

func (m Map) GetStringArray(path string) (arr []string, err error) {
	valArr, err := m.GetArray(path)
	if err != nil {
		return
	}
	for _, val := range valArr {
		str, ok := val.(string)
		if !ok {
			err = ErrInvalidType
			return
		}
		arr = append(arr, str)
	}
	return
}

func (m Map) GetStringArrayDefault(path string, defaultArr []string) (arr []string) {
	arr, err := m.GetStringArray(path)
	if err != nil {
		return defaultArr
	}
	return
}

func (m Map) GetStringArraySafe(path string) (arr []string) {
	valArr, err := m.GetArray(path)
	if err != nil {
		return
	}
	for _, val := range valArr {
		str, ok := val.(string)
		if ok {
			arr = append(arr, str)
		}
	}
	return
}

func (m Map) GetMapArray(path string) (mapArr []Map, err error) {
	arr, err := m.GetArray(path)
	if err != nil {
		return
	}
	for _, ele := range arr {
		var mapV Map
		mapV, err = CastToMap(ele)
		if err != nil {
			return
		}
		mapArr = append(mapArr, mapV)
	}
	return
}

func (m Map) GetMapArrayDefault(path string, defaultArr []Map) (mapArr []Map) {
	mapArr, err := m.GetMapArray(path)
	if err != nil {
		mapArr = defaultArr
	}
	return
}

func (m Map) Set(path string, val any) (err error) {
	if path == "" {
		err = ErrInvalidField
		return
	}
	fields, lastField := ParsePath(path)
	// retrieve parent map
	parent := m
	if len(fields) > 0 {
		parentField := fields[len(fields)-1]
		fields = fields[:len(fields)-1]
		_, parentPtr, err2 := m.ResolvePath(fields, parentField)
		if err2 != nil {
			err = err2
			return
		}
		parent, err = CastToMap(parentPtr)
		if err != nil {
			return
		}
	}
	// set field for parent map / array
	isArray, field2, arrayIndex := parseField(lastField)
	if isArray {
		var arr []any
		arr, err = CastToArrayAndVerify(parent[field2], field2, arrayIndex)
		if err != nil {
			return
		}
		arr[arrayIndex] = val
	} else {
		parent[lastField] = val
	}
	return
}

// to support gORM

func (m *Map) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
		}
		bytes = []byte(str)
	}
	var result Map
	err := json.Unmarshal(bytes, &result)
	*m = result
	return err
}

func (m Map) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}
