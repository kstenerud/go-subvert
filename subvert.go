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
	"debug/gosym"
	"math"
	"reflect"
	"unsafe"
)

// MakeWritable clears a value's RO flags. The RO flags are generally used to
// determine whether a value is exported (and thus accessible) or not.
func MakeWritable(v *reflect.Value) error {
	if !rvFlagsFound {
		return rvFlagsError
	}
	*getRVFlagPtr(v) &= ^rvFlagRO
	return nil
}

// MakeAddressable adds the addressable flag to a value, allowing you to take
// its address. The most common reason for making an object non-addressable is
// because it's allocated on the stack or in read-only memory.
//
// Making a pointer to a stack value will cause undefined behavior if you
// attempt to access it outside of the stack-allocated object's scope.
//
// Do not write to an object in read-only memory. It would be bad.
func MakeAddressable(v *reflect.Value) error {
	if !rvFlagsFound {
		return rvFlagsError
	}
	*getRVFlagPtr(v) |= rvFlagAddr
	return nil
}

// SliceAtAddress turns a memory range into a go slice.
//
// No checks are made as to whether the memory is writable or even readable.
//
// Do not append to the slice.
func SliceAtAddress(address uintptr, length int) []byte {
	return (*[math.MaxInt32]byte)(unsafe.Pointer(address))[:length:length]
}

// GetSliceAddr gets the address of a slice
func GetSliceAddr(slice []byte) uintptr {
	pSlice := (*unsafe.Pointer)((unsafe.Pointer)(&slice))
	return uintptr(*pSlice)
}

// PatchMemory applies a patch to the specified memory location. If the memory
// is read-only, it will be made temporarily writable while the patch is applied.
func PatchMemory(address uintptr, patch []byte) (oldMemory []byte, err error) {
	memory := SliceAtAddress(address, len(patch))
	oldMemory = make([]byte, len(memory))
	copy(oldMemory, memory)
	err = applyToProtectedMemory(address, uintptr(len(patch)), func() {
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
	fn, err := getFunctionSymbolByName(funcSymName)
	if err != nil {
		return
	}
	return newFunctionWithImplementation(templateFunc, uintptr(fn.Entry))
}

// AliasFunction returns a new function object that calls the same underlying
// code as the original function.
func AliasFunction(function interface{}) (aliasedFunction interface{}, err error) {
	rFunc := reflect.ValueOf(function)
	if err = MakeAddressable(&rFunc); err != nil {
		return
	}
	fAddr := *(*unsafe.Pointer)(unsafe.Pointer(rFunc.UnsafeAddr()))
	return newFunctionWithImplementation(function, uintptr(fAddr))
}

// GetSymbolTable loads (if necessary) and returns the symbol table for this process
func GetSymbolTable() (*gosym.Table, error) {
	if symTable == nil && symTableLoadError == nil {
		symTable, symTableLoadError = loadSymbolTable()
	}

	return symTable, symTableLoadError
}

// AllFunctions returns the name of every function that has been compiled
// into the current binary. Use it as a debug helper to see if a function
// has been compiled in or not.
func AllFunctions() (functions map[string]bool, err error) {
	var table *gosym.Table
	if table, err = GetSymbolTable(); err != nil {
		return
	}

	functions = make(map[string]bool)
	for _, function := range table.Funcs {
		functions[function.Name] = true
	}
	return
}

func init() {
	initReflectValue()
	initProcess()
}

const is64BitUintptr = uint64(^uintptr(0)) == ^uint64(0)
