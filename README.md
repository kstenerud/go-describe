Describe
========

A go module for describing objects as a single line of text. It provides MUCH
more information than the `%v` formatter does, allowing you to see more about
complex objects at a glance.

It handles recursive data, and can describe structures that would cause `%v`
to stack overflow.

The description is built as a single line, and is structured as follows:

 * Basic types are printed as if using `%v`
 * Strings are enclosed in quotes `""`
 * Nil pointers are printed as `nil`
 * The empty interface is printed as `interface` rather than `interface {}`
 * Non-nil pointers are prefixed with `*`
 * Interfaces are prefixed with `@`
 * Slices and arrays are preceded by a type, and enclosed in `[]`
 * Slices and arrays of unsigned int types are printed as hex
 * Maps are preceded by key and value types separated by `:`, and enclosed in `{}`. Keys-value pairs are separated by `=`
 * Structs are preceded by a type, with fields enclosed in `()`. Name-value pairs are separated by `=`
 * Custom describers by convention print a type name, then a description within `()`. Example: `url.URL(http://example.com)`
 * Duplicate and cyclic references will be marked as follows:
   - The first instance will be prepended by a unique numeric reference ID, then `->`
   - Further instances will be replaced by `$` and the referenced ID

#### Note:

Describe will use the unsafe package to expose unexported reflect.Value and
reflect.Type objects unless compiled with `-tags safe`, or if
`EnableUnsafeOperations` is set to false. It will also be disabled if compiling
for GopherJS or AppEngine.


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
{4 0xc000016288 [255 128 68 1] http://example.com 2020-01-01 01:01:01 +0000 UTC {200} 0xc000016290 <nil> map[flt:1.5 inner:{99} str:blah]}
describe.OuterStruct(AnInt=4 PInt=*1 Bytes=uint8[0xff 0x80 0x44 0x01] URL=*url.URL(http://example.com) Time=time.Time(2020-01-01 01:01:01 +0000 UTC) AStruct=describe.InnerStruct(number=200) PStruct=*describe.InnerStruct(number=100) AnotherPStruct=nil AMap=interface:interface{@"inner"=@describe.InnerStruct(number=99) @"flt"=@1.5 @"str"=@"blah"})
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
[0xc00000c3e0 0xc00000c400 0xc00000c3e0]
interface[@*1->describe.RecursiveStruct(IntVal=100 RecursivePtr=*$1 data=@2->string:interface{"mykey"=@$2}) @*describe.RecursiveStruct(IntVal=5 RecursivePtr=*$1 data=@$2) @*$1]
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
describe.RecursiveStruct(IntVal=0 RecursivePtr=nil data=@1->string:interface{"mykey"=@$1})
```


License
-------

MIT License:

Copyright 2020 Karl Stenerud

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.