package describe

import (
	"fmt"
	"strings"
	"testing"
)

type InnerStruct struct {
	number int
}

type OuterStruct struct {
	AnInt          int
	PInt           *int
	Bytes          []byte
	AStruct        InnerStruct
	PStruct        *InnerStruct
	AnotherPStruct *InnerStruct
	AMap           map[interface{}]interface{}
}

func TestDemonstration(t *testing.T) {
	intVal := 1
	structVal := InnerStruct{100}
	v := OuterStruct{
		4,
		&intVal,
		[]byte{0xff, 0x80, 0x44, 0x01},
		InnerStruct{200},
		&structVal,
		nil,
		map[interface{}]interface{}{
			"flt":   1.5,
			"str":   "blah",
			"inner": InnerStruct{99},
		},
	}
	fmt.Printf("%v\n", v)
	fmt.Printf("%v\n", Describe(v))

	// Outputs:
	// {4 0xc0000a0088 [255 128 68 1] {200} 0xc0000a0090 <nil> map[flt:1.5 inner:{99} str:blah]}
	// OuterStruct(AnInt:4 PInt:&1 Bytes:[0xff 0x80 0x44 0x01] AStruct:InnerStruct(number:200) PStruct:&InnerStruct(number:100) AnotherPStruct:nil AMap:{@"str":@"blah" @"inner":@InnerStruct(number:99) @"flt":@1.5})

	// Can't test the entire string because the map entries are in unpredictable order
	expectedPrefix := `OuterStruct(AnInt:4 PInt:&1 Bytes:[0xff 0x80 0x44 0x01] AStruct:InnerStruct(number:200) PStruct:&InnerStruct(number:100) AnotherPStruct:nil AMap:`
	actual := Describe(v)
	if !strings.HasPrefix(actual, expectedPrefix) {
		t.Errorf("Expected %v to start with %v", actual, expectedPrefix)
	}
}
