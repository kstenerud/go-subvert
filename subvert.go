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
	"bytes"
	"debug/elf"
	"debug/gosym"
	"fmt"
	"log"
	"os"
	"reflect"
	"unsafe"
)

// IsEnabled checks if initialization succeeded. If this function returns false,
// Calling other functions in this package will panic.
//
// Initialization will only fail if the assertions this package makes about
// certain internal golang type structures turn out to be false in the newer
// version of go you are compiling with.
func IsEnabled() bool {
	return failureReason == ""
}

// MakeWritable clears a value's RO flags. The RO flags are generally used to
// determine whether a value is exported (and thus accessible) or not.
func MakeWritable(v *reflect.Value) {
	assertIsEnabled()
	*getFlagPtr(v) &= ^flagRO
}

// MakeAddressable adds the addressable flag to a value, allowing you to take
// its address. The most common reason for making an object non-addressable is
// because it's allocated on the stack. Making a pointer to a stack value will
// cause undefined behavior if you attempt to access it outside of the
// stack-allocated object's scope.
func MakeAddressable(v *reflect.Value) {
	assertIsEnabled()
	*getFlagPtr(v) |= flagAddr
}

// ExposeFunction exposes a function or method, allowing you to bypass export
// restrictions. It looks for the symbol specified by funcSymName and returns a
// function with its implementation, or nil if the symbol wasn't found.
// templateFunc MUST be the correct function type, or else undefined behavior
// will result!
//
// funcSymName must be the exact symbol name from the binary. Use your developer
// tools to find it (for example, on Linux: `readelf -s myexe`)
//
// Example:
//   f := subvert.ExposeFunction("reflect.methodName", (func() string)(nil)).(func() string)
//   if f != nil {
// 	     fmt.Println(f())
//   }
//
// Note: If your program doesn't have any references to the function you're
// interested in, either directly or indirectly, it will be omitted from the
// binary during compilation. You can prevent this by saving a reference to it
// somewhere, or calling a function that indirectly references it.
//
// Currently, this function only works in Linux.
func ExposeFunction(funcSymName string, templateFunc interface{}) interface{} {
	// TODO: Support other systems
	assertIsEnabled()
	if symTableLoadFailed {
		return nil
	}
	loadSymbolTableLinux()

	fn := symTable.LookupFunc(funcSymName)
	if fn == nil {
		return nil
	}
	rf := reflect.MakeFunc(reflect.TypeOf(templateFunc), func([]reflect.Value) []reflect.Value {
		return []reflect.Value{}
	})
	oldFlag := *getFlagPtr(&rf)
	MakeAddressable(&rf)
	fPtr := (*unsafe.Pointer)(unsafe.Pointer(rf.UnsafeAddr()))
	*fPtr = unsafe.Pointer(uintptr(fn.Entry))
	*getFlagPtr(&rf) = oldFlag
	return rf.Interface()
}

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
var symTable *gosym.Table
var symTableLoadFailed bool

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

func loadSymbolTableLinux() {
	if symTable != nil || symTableLoadFailed {
		return
	}

	// Try to load from memory first
	const elfStartAddress = uintptr(0x400000)
	const maxSize = 0x10000000
	processMemory := (*[maxSize]byte)(unsafe.Pointer(elfStartAddress))[:maxSize:maxSize]
	exe, err := elf.NewFile(bytes.NewReader(processMemory))
	if err != nil {
		// Failing that, load a copy from disk
		exePath, err := os.Executable()
		if err != nil {
			symTableLoadFailed = true
			log.Printf("subvert: Error finding executable: %v", err)
			return
		}

		exe, err = elf.Open(exePath)
		if err != nil {
			symTableLoadFailed = true
			log.Printf("subvert: Error opening %v: %v", exePath, err)
			return
		}
	}
	defer exe.Close()

	lineTableData, err := exe.Section(".gopclntab").Data()
	if err != nil {
		symTableLoadFailed = true
		log.Printf("subvert: Error reading .gopclntab data: %v", err)
		return
	}

	addr := exe.Section(".text").Addr
	lineTable := gosym.NewLineTable(lineTableData, addr)
	symTable, err = gosym.NewTable([]byte{}, lineTable)
	if err != nil {
		symTableLoadFailed = true
		log.Printf("subvert: Error creating symbol table: %v", err)
		symTable = nil
	}
}
