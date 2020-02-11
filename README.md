Describe
========

A go module for describing objects. It provides MUCH more information than the
`%v` formatter does, allowing you to see more about complex objects at a glance.

It's designed primarily for debugging and testing.

The description is built as a single line, and is structured as follows:

 * Basic types are printed as-is
 * Strings are enclosed in quotes `""`
 * Pointers are preceded by `&`
 * Interfaces are preceded by `@`
 * Slices and arrays are enclosed in `[]`
 * Slices with unsigned int types are printed as hex
 * Maps are enclosed in `{}`
 * Structs are preceded by the struct type name, with fields enclosed in `()`


Example
-------

```golang
type InnerStruct struct {
	number int
}

type OuterStruct struct {
	AnInt          int
	PInt           *int
	AStruct        InnerStruct
	PStruct        *InnerStruct
	AnotherPStruct *InnerStruct
	AMap           map[interface{}]interface{}
}

func Demonstration() {
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

}
```

Outputs:

```
{4 0xc0000a0088 [255 128 68 1] {200} 0xc0000a0090 <nil> map[flt:1.5 inner:{99} str:blah]}
OuterStruct(AnInt:4 PInt:&1 Bytes:[0xff 0x80 0x44 0x01] AStruct:InnerStruct(number:200) PStruct:&InnerStruct(number:100) AnotherPStruct:nil AMap:{@"str":@"blah" @"inner":@InnerStruct(number:99) @"flt":@1.5})
```


License
-------

MIT License:

Copyright 2020 Karl Stenerud

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.