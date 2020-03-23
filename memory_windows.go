package subvert

import (
	"unsafe"
)

func osSetMemoryProtection(address uintptr, length uintptr, protection memProtect) (old memProtect, err error) {
	newProtection := protToOS[protection&0xff] | uintptr(protection&^0xff)

	// https://docs.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-virtualprotect
	// https://docs.microsoft.com/en-us/windows/win32/memory/memory-protection-constants
	var oldProtection uint32
	result, _, err := virtualProtect.Call(address,
		uintptr(length),
		newProtection,
		uintptr(unsafe.Pointer(&oldProtection)))
	if result != 0 {
		err = nil
		old = osToProt[int(oldProtection&0xff)] | memProtect(oldProtection&^0xff)
	}
	return
}

var protToOS = []uintptr{
	memProtectNone: 0,
	memProtectR:    0x02,
	memProtectW:    0,
	memProtectX:    0x10,
	memProtectRW:   0x04,
	memProtectRX:   0x20,
	memProtectWX:   0,
	memProtectRWX:  0x40,
}

var osToProt = map[int]memProtect{
	0:    memProtectNone,
	0x02: memProtectR,
	0x04: memProtectRW,
	0x10: memProtectX,
	0x20: memProtectRX,
	0x40: memProtectRWX,
}
