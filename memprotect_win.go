// +build windows

package subvert

import (
	"fmt"
	"syscall"
	"unsafe"
)

var kernel32 *syscall.LazyDLL
var virtualProtect *syscall.LazyProc

func applyToProtectedMemory(address uintptr, length int, operation func()) (err error) {
	if kernel32 == nil {
		kernel32 = syscall.NewLazyDLL("kernel32.dll")
		virtualProtect = kernel32.NewProc("VirtualProtect")
	}

	// https://docs.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-virtualprotect
	// https://docs.microsoft.com/en-us/windows/win32/memory/memory-protection-constants
	const rwProtection = uintptr(0x40)
	var oldProtection uint32
	result, _, _ := virtualProtect.Call(address, uintptr(length), rwProtection, uintptr(unsafe.Pointer(&oldProtection)))
	if result == 0 {
		return fmt.Errorf("Call to VirtualProtect failed (enabling RW)")
	}

	operation()

	var dummy uint32
	result, _, _ = virtualProtect.Call(address, uintptr(length), uintptr(oldProtection), uintptr(unsafe.Pointer(&dummy)))
	if result == 0 {
		return fmt.Errorf("Call to VirtualProtect failed (restoring old protection)")
	}

	return
}
