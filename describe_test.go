package describe

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

type InnerStruct struct {
	number int
}

type OuterStruct struct {
	AnInt          int
	PInt           *int
	Bytes          []byte
	URL            *url.URL
	Time           time.Time
	AStruct        InnerStruct
	PStruct        *InnerStruct
	AnotherPStruct *InnerStruct
	AMap           map[interface{}]interface{}
}

func TestSafe(t *testing.T) {
	if canExposeInterface() {
		t.Log("Testing with unsafe code...")
	} else {
		t.Log("Testing with safe code...")
	}
}

func TestDemonstration(t *testing.T) {
	urlVal, _ := url.Parse("http://example.com")
	intVal := 1
	structVal := InnerStruct{100}
	v := OuterStruct{
		AnInt:          4,
		PInt:           &intVal,
		Bytes:          []byte{0xff, 0x80, 0x44, 0x01},
		URL:            urlVal,
		Time:           time.Date(2020, time.Month(1), 1, 1, 1, 1, 0, time.UTC),
		AStruct:        InnerStruct{number: 200},
		PStruct:        &structVal,
		AnotherPStruct: nil,
		AMap: map[interface{}]interface{}{
			"flt":   1.5,
			"str":   "blah",
			"inner": InnerStruct{number: 99},
		},
	}
	fmt.Printf("Printf:\n%v\n\n", v)
	fmt.Printf("Describe (indent 0):\n%v\n\n", Describe(v, 0))
	fmt.Printf("Describe (indent 4):\n%v\n\n", Describe(v, 4))

	// Can't compare the entire string because the map entries are in unpredictable order
	expectedPrefix := `describe.OuterStruct<AnInt=4 PInt=*1 Bytes=uint8[0xff 0x80 0x44 0x01] URL=*url.URL<http://example.com> Time=time.Time<2020-01-01 01:01:01 +0000 UTC> AStruct=describe.InnerStruct<number=200> PStruct=*describe.InnerStruct<number=100> AnotherPStruct=nil AMap=interface:interface{`
	actual := Describe(v, 0)
	if !strings.HasPrefix(actual, expectedPrefix) {
		t.Errorf("Expected %v to start with %v", actual, expectedPrefix)
	}
}

type RecursiveStruct struct {
	IntVal       int
	RecursivePtr *RecursiveStruct
	data         interface{}
}

func TestDemonstrateRecursive(t *testing.T) {
	someMap := make(map[string]interface{})
	someMap["mykey"] = someMap
	v1 := RecursiveStruct{}
	v1.IntVal = 100
	v1.RecursivePtr = &v1
	v1.data = someMap
	v2 := RecursiveStruct{}
	v2.IntVal = 5
	v2.RecursivePtr = &v1
	v2.data = someMap
	v := []interface{}{&v1, &v2, &v1}

	fmt.Printf("Printf:\n%v\n\n", v)
	fmt.Printf("Describe (indent 0):\n%v\n\n", Describe(v, 0))
	fmt.Printf("Describe (indent 4):\n%v\n\n", Describe(v, 4))

	expected1 := `interface[@*1~describe.RecursiveStruct<IntVal=100 RecursivePtr=*$1 data=@2~string:interface{"mykey"=@$2}> @*describe.RecursiveStruct<IntVal=5 RecursivePtr=*$1 data=@$2> @*$1]`
	expected2 := `interface[@*2~describe.RecursiveStruct<IntVal=100 RecursivePtr=*$2 data=@1~string:interface{"mykey"=@$1}> @*describe.RecursiveStruct<IntVal=5 RecursivePtr=*$2 data=@$1> @*$2]`
	actual := Describe(v, 0)
	if actual != expected1 && actual != expected2 {
		t.Errorf("Expected %v but got %v", expected1, actual)
	}
}

func TestDemonstrateRecursiveMapInStruct(t *testing.T) {
	someMap := make(map[string]interface{})
	someMap["mykey"] = someMap
	v := RecursiveStruct{}
	v.data = someMap

	// Uncommenting the next line would result in a stack overflow
	// fmt.Printf("Printf:\n%v\n\n", v)
	fmt.Printf("Describe (indent 0):\n%v\n\n", Describe(v, 0))
	fmt.Printf("Describe (indent 4):\n%v\n\n", Describe(v, 4))

	expected := `describe.RecursiveStruct<IntVal=0 RecursivePtr=nil data=@1~string:interface{"mykey"=@$1}>`
	actual := Describe(v, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestRecursiveNotAddressable(t *testing.T) {
	v := RecursiveStruct{}
	v.RecursivePtr = &v

	expected := `describe.RecursiveStruct<IntVal=0 RecursivePtr=*1~describe.RecursiveStruct<IntVal=0 RecursivePtr=*$1 data=nil> data=nil>`
	actual := Describe(v, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestRecursiveMap(t *testing.T) {
	v := make(map[string]interface{})
	v["mykey"] = v
	expected := `1~string:interface{"mykey"=@$1}`
	actual := Describe(v, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestNil(t *testing.T) {
	expected := `invalid`
	actual := Describe(nil, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestReflectValue(t *testing.T) {
	v := 1
	rv := reflect.ValueOf(v)

	expected := `reflect.Value<1>`
	actual := Describe(rv, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestReflectZeroValue(t *testing.T) {
	var rv reflect.Value

	expected := `reflect.Value<invalid>`
	actual := Describe(rv, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestReflectType(t *testing.T) {
	rv := reflect.TypeOf(1)

	expected := `reflect.Type<int>`
	actual := Describe(rv, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestReflectTypeZeroValue(t *testing.T) {
	var rv reflect.Type

	expected := `invalid`
	actual := Describe(rv, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

type MyReflect struct {
	rv reflect.Value
}

func TestStructReflectValue(t *testing.T) {
	v := MyReflect{rv: reflect.ValueOf(1)}

	actual := Describe(v, 0)
	if canExposeInterface() {
		expected := `describe.MyReflect<rv=reflect.Value<1>>`
		if actual != expected {
			t.Errorf("Expected %v but got %v", expected, actual)
		}
	} else {
		expectedPrefix := "describe.MyReflect<rv=reflect.Value<{0x"
		if !strings.HasPrefix(actual, expectedPrefix) {
			t.Errorf("Expected %v to start with %v", actual, expectedPrefix)
		}
	}
}

type MyType struct {
	T reflect.Type
}

func TestStructReflectType(t *testing.T) {
	rv := MyType{reflect.TypeOf(1)}

	expected := `describe.MyType<T=reflect.Type<int>>`
	actual := Describe(rv, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestStructReflectTypeZeroValue(t *testing.T) {
	rv := MyType{}

	expected := `describe.MyType<T=nil>`
	actual := Describe(rv, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

type MyTypeUnexported struct {
	T reflect.Type
	t reflect.Type
}

func TestStructUnexportedReflectType(t *testing.T) {
	rv := MyTypeUnexported{reflect.TypeOf(1), reflect.TypeOf(1)}

	actual := Describe(rv, 0)
	if canExposeInterface() {
		expected := `describe.MyTypeUnexported<T=reflect.Type<int> t=reflect.Type<int>>`
		if actual != expected {
			t.Errorf("Expected %v but got %v", expected, actual)
		}
	} else {
		expectedPrefix := "describe.MyTypeUnexported<T=reflect.Type<int> t=reflect.Type<0x"
		if !strings.HasPrefix(actual, expectedPrefix) {
			t.Errorf("Expected %v to start with %v", actual, expectedPrefix)
		}
	}
}

func TestChan(t *testing.T) {
	ch := make(chan int)
	expected := "chan<int>"
	actual := Describe(ch, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestChanRecv(t *testing.T) {
	var f func(chan<- int)

	expected := "nilfunc(chan<- int)()"
	actual := Describe(f, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestChanSend(t *testing.T) {
	var f func(<-chan int)

	expected := "nilfunc(<-chan int)()"
	actual := Describe(f, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestFunc(t *testing.T) {
	f := func(paramA string, paramB int) (retA int, retB string) {
		retB = paramA
		retA = paramB
		return
	}
	expected := "func(string, int)(int, string)"
	actual := Describe(f, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestFuncZeroValue(t *testing.T) {
	var f func(paramA string, paramB int) (retA int, retB string)
	expected := "nilfunc(string, int)(int, string)"
	actual := Describe(f, 0)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestD(t *testing.T) {
	v := 1
	expected := "1"
	actual := D(v)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

type StructWithArray struct {
	Arr [4]byte
}

func TestStructWithArray(t *testing.T) {
	v := StructWithArray{}
	expected := "describe.StructWithArray<Arr=uint8[0x00 0x00 0x00 0x00]>"
	actual := D(v)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

type StructWithArrayRecursive struct {
	Arr [4]*StructWithArrayRecursive
}

func TestStructWithArrayRecursive(t *testing.T) {
	v := StructWithArrayRecursive{}
	v.Arr[0] = &v
	expected := "describe.StructWithArrayRecursive<Arr=*describe.StructWithArrayRecursive[*1~describe.StructWithArrayRecursive<Arr=$1> nil nil nil]>"
	actual := D(v)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}
