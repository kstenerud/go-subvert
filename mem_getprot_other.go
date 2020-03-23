// +build !windows,!linux,!darwin

package subvert

func osGetMemoryProtection(address uintptr) memProtect {
	// TODO
	return memProtectRX
}
