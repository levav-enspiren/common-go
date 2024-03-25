package skmap_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/levav-enspiren/common-go/skmap"
)

// ------------ helpers

func ExpectEqual[CT any](t *testing.T, expression string, expectedValue CT, actualValue CT) {
	ExpectWithDesF(t, func() string {
		expectedValueStr := fmt.Sprint(expectedValue)
		actualValueStr := fmt.Sprint(actualValue)
		return fmt.Sprintf("%s should be equals to %s, but got %s", expression, expectedValueStr, actualValueStr)
	}, reflect.DeepEqual(expectedValue, actualValue))
}

func ExpectNotEqual[CT any](t *testing.T, expression string, expectedValue CT, actualValue CT) {
	ExpectWithDesF(t, func() string {
		expectedValueStr := fmt.Sprint(expectedValue)
		return fmt.Sprintf("%s should not be equals to %s", expression, expectedValueStr)
	}, !reflect.DeepEqual(expectedValue, actualValue))
}

func Expect(t *testing.T, description string, condition bool) {
	ExpectWithDesF(t, func() string { return description }, condition)
}

func ExpectWithDesF(t *testing.T, descriptionFactory func() string, condition bool) {
	if !condition {
		t.Fatal(descriptionFactory())
	}
}

// ------------ test config

func getTestObject0() skmap.Map {
	return skmap.Map{
		"str":   "abc",
		"int":   123,
		"float": 12.3,
		"obj": skmap.Map{
			"int": 456,
		},
		"arr": []any{1, "a", skmap.Map{
			"int": 789,
		}},
		"strArr": []any{"a", "b"},
	}
}

func performTest0(t *testing.T, myMap skmap.Map) {
	// ----- get
	// primitive
	ExpectEqual(t, "str", "abc", myMap.GetStringDefault("str", ""))
	ExpectEqual(t, "int", 123, myMap.GetIntDefault("int", -1))
	ExpectEqual(t, "float", 12.3, myMap.GetFloat64Default("float", -1))
	// invalid key
	ExpectEqual(t, "foo", "", myMap.GetStringDefault("foo", ""))
	// map
	ExpectEqual(t, "obj.int", 456, myMap.GetIntDefault("obj.int", -1))
	// invalid key in map
	ExpectEqual(t, "obj.foo", nil, myMap.GetDefault("obj.foo", nil))
	// array
	arr, err := myMap.GetArray("arr")
	Expect(t, "getting array arr should not throw err", err == nil)
	ExpectNotEqual(t, "arr", nil, arr)
	ExpectEqual(t, "arr[0]", 1, myMap.GetIntDefault("arr[0]", -1))
	ExpectEqual(t, "arr[1]", "a", myMap.GetStringDefault("arr[1]", ""))
	ExpectNotEqual(t, "arr[2]", nil, myMap.GetMapDefault("arr[2]", nil))
	ExpectEqual(t, "arr[2].int", 789, myMap.GetIntDefault("arr[2].int", -1))
	ExpectEqual(t, "strArr", "a", myMap.GetStringArrayDefault("strArr", []string{"error"})[0])
	ExpectEqual(t, "strArr length (safe)", 2, len(myMap.GetStringArraySafe("strArr")))
	ExpectEqual(t, "arr length (safe)", 1, len(myMap.GetStringArraySafe("arr")))
	// invalid array index
	ExpectEqual(t, "arr[4]", nil, myMap.GetDefault("arr[4]", nil))
	// ----- set
	myMap.Set("str", 321)
	ExpectEqual(t, "str", "", myMap.GetStringDefault("str", ""))
	ExpectEqual(t, "str", 321, myMap.GetIntDefault("str", -1))
	// map
	myMap.Set("obj.int", 654)
	ExpectEqual(t, "obj.int", 654, myMap.GetIntDefault("obj.int", -1))
	// new key of map
	myMap.Set("obj.str", "test1")
	ExpectEqual(t, "obj.str", "test1", myMap.GetStringDefault("obj.str", ""))
	// array
	myMap.Set("arr[2].int", 987)
	ExpectEqual(t, "arr[2].int", 987, myMap.GetIntDefault("arr[2].int", -1))
	// index excess size of array
	err = myMap.Set("arr[100]", "test2")
	Expect(t, "setting arr[100] should throw err", err != nil)
	// replace array
	arr2 := append(arr, "test3")
	err = myMap.Set("arr", arr2)
	Expect(t, "setting arr should not throw err", err == nil)
	ExpectEqual(t, "arr[3]", "test3", myMap.GetStringDefault("arr[3]", ""))
}

func TestCustomMap(t *testing.T) {
	myMap := getTestObject0()
	performTest0(t, myMap)
}

func TestUnmarshalMap(t *testing.T) {
	myMap := getTestObject0()
	bytes, _ := json.Marshal(myMap)
	json.Unmarshal(bytes, &myMap)
	performTest0(t, myMap)
}
