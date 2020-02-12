Describe
========

A go module for describing objects as a single line of text. It provides MUCH
more information than the `%v` formatter does, allowing you to see more about
complex objects at a glance.

It handles recursive data, and can describe structures that would cause `%v`
to stack overflow.

The description is built as a single line, and is structured as follows:

 * Basic types are printed as-is (as if printed via `%v`)
 * Strings are enclosed in quotes `""`
 * Nil pointers are simply printed as `nil`
 * Pointers are preceded by `&`
 * Interfaces are preceded by `@`
 * Slices and arrays are enclosed in `[]`
 * Slices and arrays of unsigned int types are printed as hex
 * Maps are enclosed in `{}`, listing keys and values separated by `:`
 * Structs are preceded by the type name, with fields enclosed in `()`, listing
   field names and values separated by `:`
 * Custom describers by convention print a type name, then a description
   within `<>` (example: `url<http://example.com>`)
 * Duplicate and cyclic references will be marked as follows:
   - The first instance will be prepended by an ID and `=`
   - Further instances will be replaced by a reference: `$` and the ID


Examples
--------

#### Basic types, pointers, and interfaces

```golang
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

func Demonstration() {
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
	fmt.Printf("%v\n", v)
	fmt.Printf("%v\n", Describe(v))
}
```

Outputs:

```
{4 0xc000018258 [255 128 68 1] http://example.com 2020-01-01 01:01:01 +0000 UTC {200} 0xc000018260 <nil> map[flt:1.5 inner:{99} str:blah]}
OuterStruct(AnInt:4 PInt:&1 Bytes:[0xff 0x80 0x44 0x01] URL:&url<http://example.com> Time:time<2020-01-01 01:01:01 +0000 UTC> AStruct:InnerStruct(number:200) PStruct:&InnerStruct(number:100) AnotherPStruct:nil AMap:{@"flt":@1.5 @"str":@"blah" @"inner":@InnerStruct(number:99)})
```

#### Recursive or repetitive data structures

```golang
type RecursiveStruct struct {
	IntVal       int
	RecursivePtr *RecursiveStruct
	data         interface{}
}

func DemonstrateRecursive() {
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
	slice := []interface{}{&v1, &v2, &v1}

	fmt.Printf("%v\n", slice)
	fmt.Printf("%v\n", Describe(slice))
}
```

Outputs:

```
[0xc00000c340 0xc00000c360 0xc00000c340]
[@&1=RecursiveStruct(IntVal:100 RecursivePtr:&$1 data:@2={"mykey":@$2}) @&RecursiveStruct(IntVal:5 RecursivePtr:&$1 data:@$2) @&$1]
```

#### Recursive structure that would cause `%v` to stack overflow

```golang
func DemonstrateRecursiveMapInStruct() {
	someMap := make(map[string]interface{})
	someMap["mykey"] = someMap
	v := RecursiveStruct{}
	v.data = someMap

	// Uncommenting this would result in a stack overflow
	// fmt.Printf("%v\n", v)

	fmt.Printf("%v\n", Describe(v))
}
```

Outputs:

```
RecursiveStruct(IntVal:0 RecursivePtr:nil data:@1={"mykey":@$1})
```


License
-------

MIT License:

Copyright 2020 Karl Stenerud

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.