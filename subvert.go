// Package subvert provides functions to subvert go's type & memory protections,
// and expose unexported values & functions.
//
// This is not a power to be taken lightly! It's expected that you're fully
// versed in how the go type system works, and why there are protections and
// restrictions in the first place. Using this package incorrectly will quickly
// lead to undefined behavior and bizarre crashes, even segfaults or nuclear
// missile launches.

// YOU HAVE BEEN WARNED!
package subvert

import (
	"bytes"
	"debug/gosym"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"reflect"
	"unsafe"
)

// MakeWritable clears a value's RO flags. The RO flags are generally used to
// determine whether a value is exported (and thus accessible) or not.
func MakeWritable(v *reflect.Value) error {
	if !flagsFound {
		return flagsError
	}
	*getFlagPtr(v) &= ^flagRO
	return nil
}

// MakeAddressable adds the addressable flag to a value, allowing you to take
// its address. The most common reason for making an object non-addressable is
// because it's allocated on the stack. Making a pointer to a stack value will
// cause undefined behavior if you attempt to access it outside of the
// stack-allocated object's scope.
func MakeAddressable(v *reflect.Value) error {
	if !flagsFound {
		return flagsError
	}
	*getFlagPtr(v) |= flagAddr
	return nil
}

// SliceAtAddress turns a memory range into a slice that can be read in goland.
//
// Warning: This function makes no warranty as to whether the memory is
// accessible or writable!
func SliceAtAddress(address uintptr, length int) []byte {
	return (*[math.MaxInt32]byte)(unsafe.Pointer(address))[:length:length]
}

// PatchMemory applies a patch to the specified memory location. If that memory
// is protected, it will be made temporarily writable.
func PatchMemory(address uintptr, patch []byte) (oldMemory []byte, err error) {
	memory := SliceAtAddress(address, len(patch))
	oldMemory = make([]byte, len(memory))
	copy(oldMemory, memory)
	err = applyToProtectedMemory(address, len(patch), func() {
		copy(memory, patch)
	})
	return
}

// ExposeFunction exposes a function or method, allowing you to bypass export
// restrictions. It looks for the symbol specified by funcSymName and returns a
// function with its implementation, or nil if the symbol wasn't found.
//
// funcSymName must be the exact symbol name from the binary. Use AllFunctions()
// to find it. If your program doesn't have any references to a function, it
// will be omitted from the binary during compilation. You can prevent this by
// saving a reference to it somewhere, or calling a function that indirectly
// references it.
//
// templateFunc MUST have the correct function type, or else undefined behavior
// will result!
//
// Example:
//   exposed := ExposeFunction("reflect.methodName", (func() string)(nil))
//   if exposed != nil {
//       f := exposed.(func() string)
//       fmt.Printf("Result of reflect.methodName: %v\n", f())
//   }
func ExposeFunction(funcSymName string, templateFunc interface{}) (function interface{}, err error) {
	if err = ensureSymbolTableIsLoaded(); err != nil {
		return
	}

	fn := symbolTable.LookupFunc(funcSymName)
	if fn == nil {
		err = fmt.Errorf("Could not find function symbol %v", funcSymName)
		return
	}
	rFunc := reflect.MakeFunc(reflect.TypeOf(templateFunc), func([]reflect.Value) []reflect.Value {
		return []reflect.Value{}
	})
	oldFlag := *getFlagPtr(&rFunc)
	if err = MakeAddressable(&rFunc); err != nil {
		return
	}
	fPtr := (*unsafe.Pointer)(unsafe.Pointer(rFunc.UnsafeAddr()))
	*fPtr = unsafe.Pointer(uintptr(fn.Entry))
	*getFlagPtr(&rFunc) = oldFlag
	function = rFunc.Interface()
	return
}

// AllFunctions returns the name of every function that has been compiled
// into the current binary. Use it as a debug helper to see if a function
// has been compiled in or not.
func AllFunctions() (functions map[string]bool, err error) {
	if err = ensureSymbolTableIsLoaded(); err != nil {
		return
	}

	functions = make(map[string]bool)
	for _, function := range symbolTable.Funcs {
		functions[function.Name] = true
	}
	return
}

var (
	flagAddr uintptr
	flagRO   uintptr

	flagOffset uintptr
	flagsFound bool
	flagsError = fmt.Errorf("This function is disabled because the internal " +
		"flags structure has changed with this go release. Please open " +
		"an issue at https://github.com/kstenerud/go-subvert/issues/new")

	symbolTable          *gosym.Table
	symbolTableLoadError error
)

func init() {
	initReflectValueFlags()
}

type flagTester struct {
	A   int // reflect/value.go: flagAddr
	a   int // reflect/value.go: flagStickyRO
	int     // reflect/value.go: flagEmbedRO
	// Note: flagRO = flagStickyRO | flagEmbedRO as of go 1.5
}

func initReflectValueFlags() {
	fail := func(reason string) {
		flagsFound = false
		log.Println(fmt.Sprintf("reflect.Value flags could not be determined because %v."+
			"Please open an issue at https://github.com/kstenerud/go-subvert/issues", reason))
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
	flagsFound = true
}

func ensureSymbolTableIsLoaded() (err error) {
	if symbolTableLoadError != nil || symbolTable != nil {
		return symbolTableLoadError
	}

	var reader io.ReaderAt

	if canLoadSymbolsFromMemory {
		reader = bytes.NewReader(SliceAtAddress(processStartAddress, 0x10000000))
		if symbolTable, err = readSymbols(reader); err == nil {
			// Successfully loaded from memory
			return
		}
	}

	// If memory load fails, read from disk

	var exePath string
	if exePath, err = os.Executable(); err != nil {
		symbolTableLoadError = err
		return
	}

	if reader, err = os.Open(exePath); err != nil {
		symbolTableLoadError = err
		return
	}

	symbolTable, err = readSymbols(reader)
	symbolTableLoadError = err
	return
}

func getFlagPtr(v *reflect.Value) *uintptr {
	return (*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + flagOffset))
}
