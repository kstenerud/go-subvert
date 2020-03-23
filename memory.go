package subvert

func setMemoryProtection(address uintptr, length uintptr, protection memProtect) (oldProtection memProtect, err error) {
	return osSetMemoryProtection(address, length, protection)
}

// Unprotect a memory region, perform an operation, and then restore the old
// protection.
func applyToProtectedMemory(address uintptr, length uintptr, operation func()) (err error) {
	oldProtection, err := setMemoryProtection(address, length, memProtectRWX)
	if err != nil {
		return
	}

	operation()

	_, err = setMemoryProtection(address, length, oldProtection)
	return
}

type memProtect int

const (
	memProtectNone memProtect = 0
	memProtectR               = 1
	memProtectW               = 2
	memProtectX               = 4
	memProtectRW              = memProtectR | memProtectW
	memProtectRX              = memProtectR | memProtectX
	memProtectWX              = memProtectW | memProtectX
	memProtectRWX             = memProtectR | memProtectW | memProtectX
)
