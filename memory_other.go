// +build !windows

package subvert

import (
	"syscall"
)

var protToOS = []int{
	memProtectNone: 0,
	memProtectR:    syscall.PROT_READ,
	memProtectW:    syscall.PROT_WRITE,
	memProtectX:    syscall.PROT_EXEC,
	memProtectRW:   syscall.PROT_READ | syscall.PROT_WRITE,
	memProtectRX:   syscall.PROT_READ | syscall.PROT_EXEC,
	memProtectWX:   syscall.PROT_WRITE | syscall.PROT_EXEC,
	memProtectRWX:  syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC,
}

var osToProt = []memProtect{
	0:                                      memProtectNone,
	syscall.PROT_READ:                      memProtectR,
	syscall.PROT_WRITE:                     memProtectW,
	syscall.PROT_EXEC:                      memProtectX,
	syscall.PROT_READ | syscall.PROT_WRITE: memProtectRW,
	syscall.PROT_READ | syscall.PROT_EXEC:  memProtectRX,
	syscall.PROT_WRITE | syscall.PROT_EXEC: memProtectWX,
	syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC: memProtectRWX,
}

func osSetMemoryProtection(address uintptr, length uintptr, protection memProtect) (old memProtect, err error) {
	old = osGetMemoryProtection(address)
	end := address + uintptr(length)
	for pageStart := address & pageBeginMask; pageStart < end; pageStart += uintptr(pageSize) {
		page := SliceAtAddress(address&pageBeginMask, pageSize)
		if err = syscall.Mprotect(page, protToOS[protection]); err != nil {
			return
		}
	}
	return
}
