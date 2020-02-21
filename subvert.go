// Package subvert provides functions for subverting go's type system.
//
// As this package modifies internal type data, there's no guarantee that it
// will continue to work in future versions of go (although an incompatible
// change has yet to happen, so it seems stable enough). If, in future, an
// incompatible change were to be introduced, `IsEnabled()` would return false
// when this package is built using that particular version of go. It's on you
// to check `IsEnabled()` as part of your CI process.
//
// This is not a power to be taken lightly! It's expected that you're fully
// versed in how the go type system works, and why there are protections and
// restrictions in the first place. Using this package incorrectly will quickly
// lead to undefined behavior and bizarre crashes, perhaps even segfaults or
// nuclear missile launches.
//
// YOU HAVE BEEN WARNED!
package subvert

import (
	"fmt"
	"log"
	"reflect"
	"unsafe"
)

const failureFmt = "go-subvert is disabled because %v. Please open an issue " +
	"at https://github.com/kstenerud/go-subvert/issues"

type flagTester struct {
	A   int // reflect/value.go: flagAddr
	a   int // reflect/value.go: flagStickyRO
	int     // reflect/value.go: flagEmbedRO
	// Note: flagRO = flagStickyRO | flagEmbedRO as of go 1.5
}

var (
	flagAddr uintptr
	flagRO   uintptr
)

var flagOffset uintptr
var failureReason string

func init() {
	fail := func(reason string) {
		failureReason = reason
		log.Println(fmt.Sprintf(failureFmt, failureReason))
	}
	getFlag := func(v reflect.Value) uintptr {
		return uintptr(reflect.ValueOf(v).FieldByName("flag").Uint())
	}
	getFldFlag := func(v reflect.Value, fieldName string) uintptr {
		return getFlag(v.FieldByName(fieldName))
	}

	if field, ok := reflect.TypeOf(reflect.Value{}).FieldByName("flag"); ok {
		flagOffset = field.Offset
	} else {
		fail("reflect.Value no longer has a flag field")
		return
	}

	v := flagTester{}
	rv := reflect.ValueOf(&v).Elem()
	flagRO = (getFldFlag(rv, "a") | getFldFlag(rv, "int")) ^ getFldFlag(rv, "A")
	if flagRO == 0 {
		fail("reflect.Value.flag no longer has flagEmbedRO or flagStickyRO bit")
		return
	}

	flagAddr = getFlag(reflect.ValueOf(int(1))) ^ getFldFlag(rv, "A")
	if flagAddr == 0 {
		fail("reflect.Value.flag no longer has a flagAddr bit")
		return
	}
}

func assertIsEnabled() {
	if !IsEnabled() {
		panic(fmt.Errorf(failureFmt, failureReason))
	}
}

func getFlagPtr(v *reflect.Value) *uintptr {
	return (*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + flagOffset))
}

// ----------
// Public API
// ----------

// Check if initialization succeeded. If this function returns false, Calling
// other functions in this package will panic.
//
// Initialization will only fail if the assertions this package makes about
// certain internal golang type structures turn out to be false in the newer
// version of go you are compiling with.
//
// Warning:
// While it's almost certain that IsEnabled() will alert you to a problem,
// there's no 100% guarantee that it will! We're messing with internal data,
// which always leaves the possibility, no matter how slight, of an undetected
// problem.
func IsEnabled() bool {
	return failureReason == ""
}

func MakeWritable(v *reflect.Value) {
	assertIsEnabled()
	*getFlagPtr(v) &= ^flagRO
}

func MakeAddressable(v *reflect.Value) {
	assertIsEnabled()
	*getFlagPtr(v) |= flagAddr
}
