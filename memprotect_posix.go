// +build !windows

package subvert

import (
	"syscall"
)

func changeMemoryProtection(address uintptr, length int, protection int) (oldProtection int, err error) {
	oldProtection = getCurrentMemoryProtection(address)
	end := address + uintptr(length)
	for pageStart := address & pageBeginMask; pageStart < end; pageStart += uintptr(sysPageSize) {
		page := SliceAtAddress(address&pageBeginMask, sysPageSize)
		if err = syscall.Mprotect(page, protection); err != nil {
			return
		}
	}
	return
}

func applyToProtectedMemory(address uintptr, length int, operation func()) (err error) {
	var oldProtection int
	rwxProtection := syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC

	if oldProtection, err = changeMemoryProtection(address, length, rwxProtection); err != nil {
		return
	}
	operation()
	_, err = changeMemoryProtection(address, length, oldProtection)
	return
}

var sysPageSize = syscall.Getpagesize()
var pageBeginMask = ^uintptr(sysPageSize - 1)

func getCurrentMemoryProtection(address uintptr) int {
	// TODO: Parse /proc/self/maps:
	// 559576822000-559576827000 r-xp 00002000 00:1a 4586   /usr/bin/cat

	// TODO: We don't have this for macos.
	// https://stackoverflow.com/questions/1627998/retrieving-the-memory-map-of-its-own-process-in-os-x-10-5-10-6
	// https://stackoverflow.com/questions/9198385/on-os-x-how-do-you-find-out-the-current-memory-protection-level

	return syscall.PROT_READ | syscall.PROT_EXEC
}
