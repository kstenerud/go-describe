Describe
========

A go package for describing objects as a single line or multiline. It provides
MUCH more information than the `%v` formatter does, allowing you to see more
about complex objects at a glance.

It handles recursive data, and can describe structures that would cause `%v`
to stack overflow.

The description is structured as follows:

 * Basic types are printed the same as by `%v`
 * Strings are enclosed in quotes `""`
 * Non-nil pointers are prefixed with `*`
 * Nil pointers are printed as `nil`
 * Interfaces are prefixed with `@`
 * The empty interface type is printed as `interface` (not `interface{}`)
 * Slices and arrays are preceded by a type, and enclosed in `[]`
 * Slices and arrays of unsigned int types are printed as hex
 * Maps begin with `key_type:value_type`, with elements enclosed in `{}`.
   Key-value pairs separated by `=`
 * Structs are preceded by a type, with elements enclosed in `<>`.
   Field-value pairs are separated by `=`
 * Functions begin with `func`, with in and out params enclosed in `()`.
   Example: `func(int, bool)(string, bool)`
 * Nil functions begin with `nilfunc`. Example: `nilfunc(int)(string)`
 * Unidirectional channels are printed as `<-chan type` and `chan<- type`
 * Bidirectional channels are printed as `chan<sometype>`
 * Uintptr and UnsafePointer are printed as hex, in the width of the host system
 * Invalid values are printed as `invalid`
 * Custom describers by convention print a type name, then a description within
   `<>`. Example: `url.URL<http://xyz.com>`
 * Duplicate and cyclic data will be marked as follows:
   - The first instance is prefixed by a unique numeric reference ID, then `~`
   - Further instances are replaced by `$`, then the referenced ID

**Note:** Only data is printed; type-specific things such as methods are not.

**Note:** Describe uses the `unsafe` package to expose unexported
          `reflect.Value` and `reflect.Type` objects. This functionality can
          be disabled by compiling with `-tags safe`, or by setting
          `describe.EnableUnsafeOperations` to `false`. It will be
          automatically disabled if compiling for GopherJS or AppEngine.


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
	fmt.Printf("Printf:\n%v\n\n", v)
	fmt.Printf("Describe (indent 0):\n%v\n\n", Describe(v, 0))
	fmt.Printf("Describe (indent 4):\n%v\n\n", Describe(v, 4))
}
```

Outputs:

```
Printf:
{4 0xc0000a00e8 [255 128 68 1] http://example.com 2020-01-01 01:01:01 +0000 UTC {200} 0xc0000a00f0 <nil> map[flt:1.5 inner:{99} str:blah]}

Describe (indent 0):
describe.OuterStruct<AnInt=4 PInt=*1 Bytes=uint8[0xff 0x80 0x44 0x01] URL=*url.URL<http://example.com> Time=time.Time<2020-01-01 01:01:01 +0000 UTC> AStruct=describe.InnerStruct<number=200> PStruct=*describe.InnerStruct<number=100> AnotherPStruct=nil AMap=interface:interface{@"flt"=@1.5 @"str"=@"blah" @"inner"=@describe.InnerStruct<number=99>}>

Describe (indent 4):
describe.OuterStruct<
    AnInt = 4
    PInt = *1
    Bytes = uint8[
        0xff
        0x80
        0x44
        0x01
    ]
    URL = *url.URL<http://example.com>
    Time = time.Time<2020-01-01 01:01:01 +0000 UTC>
    AStruct = describe.InnerStruct<
        number = 200
    >
    PStruct = *describe.InnerStruct<
        number = 100
    >
    AnotherPStruct = nil
    AMap = interface:interface{
        @"inner" = @describe.InnerStruct<
            number = 99
        >
        @"flt" = @1.5
        @"str" = @"blah"
    }
>
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
	v := []interface{}{&v1, &v2, &v1}

	fmt.Printf("Printf:\n%v\n\n", v)
	fmt.Printf("Describe (indent 0):\n%v\n\n", Describe(v, 0))
	fmt.Printf("Describe (indent 4):\n%v\n\n", Describe(v, 4))
}
```

Outputs:

```
Printf:
[0xc0000ac540 0xc0000ac560 0xc0000ac540]

Describe (indent 0):
interface[@*1~describe.RecursiveStruct<IntVal=100 RecursivePtr=*$1 data=@2~string:interface{"mykey"=@$2}> @*describe.RecursiveStruct<IntVal=5 RecursivePtr=*$1 data=@$2> @*$1]

Describe (indent 4):
interface[
    @*1~describe.RecursiveStruct<
        IntVal = 100
        RecursivePtr = *$1
        data = @2~string:interface{
            "mykey" = @$2
        }
    >
    @*describe.RecursiveStruct<
        IntVal = 5
        RecursivePtr = *$1
        data = @$2
    >
    @*$1
]
```

#### Recursive structure that would cause `%v` to stack overflow

```golang
type RecursiveStruct struct {
	IntVal       int
	RecursivePtr *RecursiveStruct
	data         interface{}
}

func DemonstrateRecursiveMapInStruct() {
	someMap := make(map[string]interface{})
	someMap["mykey"] = someMap
	v := RecursiveStruct{}
	v.data = someMap

	// Uncommenting the next line would result in a stack overflow
	// fmt.Printf("Printf:\n%v\n\n", v)
	fmt.Printf("Describe (indent 0):\n%v\n\n", Describe(v, 0))
	fmt.Printf("Describe (indent 4):\n%v\n\n", Describe(v, 4))
}
```

Outputs:

```
Describe (indent 0):
describe.RecursiveStruct<IntVal=0 RecursivePtr=nil data=@1~string:interface{"mykey"=@$1}>

Describe (indent 4):
describe.RecursiveStruct<
    IntVal = 0
    RecursivePtr = nil
    data = @1~string:interface{
        "mykey" = @$1
    }
>
```


License
-------

MIT License:

Copyright 2020 Karl Stenerud

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
