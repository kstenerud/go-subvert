Subvert
=======

Package subvert provides functions for subverting go's type system.

As this package modifies internal type data, there's no guarantee that it
will continue to work in future versions of go (although an incompatible
change has yet to happen, so it seems stable enough). If, in future, an
incompatible change were to be introduced, `IsEnabled()` would return false
when this package is built using that particular version of go. It's on you
to check `IsEnabled()` as part of your CI process (or just run the unit tests
in this package).

This is not a power to be taken lightly! It's expected that you're fully
versed in how the go type system works, and why there are protections and
restrictions in the first place. Using this package incorrectly will quickly
lead to undefined behavior and bizarre crashes, perhaps even segfaults or
nuclear missile launches.

YOU HAVE BEEN WARNED!


Example
-------

```golang
import (
	"fmt"
	"reflect"
)

type SubvertTester struct {
	A int
	a int
	int
}

func Demonstrate() {
	v := SubvertTester{1, 2, 3}

	rv := reflect.ValueOf(v)
	rv_A := rv.FieldByName("A")
	rv_a := rv.FieldByName("a")
	rv_int := rv.FieldByName("int")

	fmt.Printf("Interface of A: %v\n", rv_A.Interface())

	// rv_a.Interface() // This would panic
	MakeWritable(&rv_a)
	fmt.Printf("Interface of a: %v\n", rv_a.Interface())

	// rv_int.Interface() // This would panic
	MakeWritable(&rv_int)
	fmt.Printf("Interface of int: %v\n", rv_int.Interface())

	// rv.Addr() // This would panic
	MakeAddressable(&rv)
	fmt.Printf("Pointer to v: %v\n", rv.Addr())
}
```

Prints:

```
Interface of A: 1
Interface of a: 2
Interface of int: 3
Pointer to v: &{1 2 3}
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
